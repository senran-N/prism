// Pool warmer: background goroutine that pre-registers SC accounts
// so users never wait for registration. Keeps N ready accounts warm.
package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/senran-N/prism/internal/account"
	"github.com/senran-N/prism/internal/github"
	"github.com/senran-N/prism/internal/mail"
	"github.com/senran-N/prism/internal/scproto"
)

var (
	MinReadyAccounts  = 1                // with one GitHub account, only 1 can be active
	WarmCheckInterval = 30 * time.Second
	MaxConcurrentWarm = 1                // serialize — one GitHub account means one at a time
)

// StartPoolWarmer runs a background loop that keeps the pool warm.
func (s *Scheduler) StartPoolWarmer() {
	go func() {
		time.Sleep(5 * time.Second)
		s.warmPool()

		ticker := time.NewTicker(WarmCheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			s.warmPool()
		}
	}()
	log.Printf("[warmer] started: min_ready=%d check_interval=%s", MinReadyAccounts, WarmCheckInterval)
}

func (s *Scheduler) warmPool() {
	stats := s.pool.Stats()
	needed := MinReadyAccounts - stats.Ready
	if needed <= 0 {
		return
	}

	log.Printf("[warmer] pool needs %d accounts (ready=%d, active=%d, exhausted=%d)",
		needed, stats.Ready, stats.Active, stats.Exhausted)

	for i := 0; i < needed && i < MaxConcurrentWarm; i++ {
		if err := s.registerOneAccount(); err != nil {
			log.Printf("[warmer] account creation failed: %v", err)
		}
	}

	newStats := s.pool.Stats()
	log.Printf("[warmer] done: ready=%d total=%d credits=$%.2f",
		newStats.Ready, newStats.Total, newStats.TotalCredits)
}

func (s *Scheduler) registerOneAccount() error {
	// Only take GitHub from exhausted accounts
	old := s.pool.FindExhaustedWithGitHub()
	if old != nil && old.Client != nil {
		log.Printf("[warmer] unbinding GitHub from exhausted %s", old.Email)
		if err := old.Client.DisconnectGitHub(); err != nil {
			log.Printf("[warmer] unbind warning: %v", err)
		}
		old.GitHubBound = false
	}

	ghClient, err := github.Login(s.cfg.GitHubUser, s.cfg.GitHubPass, s.cfg.GitHubTOTP)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("prism-w-%d", time.Now().UnixNano()%1000000000)
	emailAddr, _, err := mail.CreateTempEmail(s.cfg.YYDSAPIKey, prefix)
	if err != nil {
		return err
	}

	password := fmt.Sprintf("PrismW#%06dxPass!", time.Now().Unix()%1000000)
	sc := scproto.NewClient()
	if err := sc.Register(emailAddr, password, "Prism User"); err != nil {
		return err
	}

	if err := sc.ConnectGitHub(ghClient); err != nil {
		return err
	}

	projectID, err := sc.CreateProject(s.cfg.RepoID)
	if err != nil {
		return err
	}

	if err := sc.CompleteEnvironmentSetup(projectID); err != nil {
		log.Printf("[warmer] env setup warning: %v", err)
	}
	if err := sc.WaitForEnvironment(projectID, 60*time.Second); err != nil {
		log.Printf("[warmer] env wait warning: %v", err)
	}

	newAcct := &account.Account{
		ID:          account.GenerateAccountID(),
		Email:       emailAddr,
		Password:    password,
		WorkspaceID: sc.WorkspaceID,
		UserID:      sc.UserID,
		ProjectID:   projectID,
		RepoID:      s.cfg.RepoID,
		Credits:     s.cfg.InitialCredits,
		Status:      account.StatusReady,
		GitHubBound: true,
		CreatedAt:   time.Now(),
		Client:      sc,
	}
	s.pool.Add(newAcct)
	log.Printf("[warmer] account ready: %s ($%.2f)", emailAddr, newAcct.Credits)
	return nil
}
