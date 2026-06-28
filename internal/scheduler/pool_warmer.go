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
	MinReadyAccounts  = 1
	WarmCheckInterval = 60 * time.Second
	MaxConcurrentWarm = 1
)

// StartPoolWarmer ensures one account is always ready.
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
	log.Printf("[warmer] started: min_ready=%d", MinReadyAccounts)
}

func (s *Scheduler) warmPool() {
	stats := s.pool.Stats()
	if stats.Ready+stats.Active >= MinReadyAccounts {
		return
	}

	log.Printf("[warmer] no ready accounts (ready=%d active=%d exhausted=%d), creating one...",
		stats.Ready, stats.Active, stats.Exhausted)

	if err := s.registerOneAccount(); err != nil {
		log.Printf("[warmer] failed: %v", err)
	}
}

func (s *Scheduler) registerOneAccount() error {
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

	if err := sc.CompleteEnvironmentSetup(projectID, s.cfg.RepoName); err != nil {
		log.Printf("[warmer] env setup warning: %v", err)
	}
	if err := sc.WaitForEnvironment(projectID, 60*time.Second); err != nil {
		log.Printf("[warmer] env wait warning: %v", err)
	}

	credits := s.cfg.InitialCredits
	if actual, err := sc.GetCredits(); err == nil && actual > 0 {
		credits = actual
	}

	newAcct := &account.Account{
		ID:          account.GenerateAccountID(),
		Email:       emailAddr,
		Password:    password,
		WorkspaceID: sc.WorkspaceID,
		UserID:      sc.UserID,
		ProjectID:   projectID,
		RepoID:      s.cfg.RepoID,
		Credits:     credits,
		Status:      account.StatusReady,
		GitHubBound: true,
		CreatedAt:   time.Now(),
		Client:      sc,
	}
	s.pool.Add(newAcct)
	log.Printf("[warmer] account ready: %s ($%.2f)", emailAddr, credits)
	return nil
}
