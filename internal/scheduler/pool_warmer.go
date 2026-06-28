// Pool warmer: background goroutine that pre-registers SC accounts
// so users never wait for registration. Keeps N ready accounts warm.
package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/senran-N/prism/internal/account"
	"github.com/senran-N/prism/internal/github"
	"github.com/senran-N/prism/internal/mail"
	"github.com/senran-N/prism/internal/scproto"
)

var (
	MinReadyAccounts  = 5                // keep at least N accounts ready
	WarmCheckInterval = 30 * time.Second // check every 30s
	MaxConcurrentWarm = 3                // register up to N accounts in parallel
	EnvSetupWait      = 25 * time.Second
)

// StartPoolWarmer runs a background loop that keeps the pool warm.
func (s *Scheduler) StartPoolWarmer() {
	go func() {
		// Initial warm-up after 5s startup delay
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
	if needed > MaxConcurrentWarm {
		needed = MaxConcurrentWarm
	}

	log.Printf("[warmer] pool needs %d accounts (ready=%d)", needed, stats.Ready)

	var wg sync.WaitGroup
	for i := 0; i < needed; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := s.registerOneAccount(); err != nil {
				log.Printf("[warmer] account %d failed: %v", idx, err)
			}
		}(i)
	}
	wg.Wait()

	newStats := s.pool.Stats()
	log.Printf("[warmer] done: ready=%d total=%d", newStats.Ready, newStats.Total)
}

func (s *Scheduler) registerOneAccount() error {
	// Find account with GitHub to unbind if needed
	old := s.pool.FindExhaustedWithGitHub()
	if old != nil && old.Client != nil {
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
	log.Printf("[warmer] account ready: %s", emailAddr)
	return nil
}
