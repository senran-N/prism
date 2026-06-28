// Package scheduler orchestrates account rotation and task dispatch.
// One account at a time. Use it until credits run out, then rotate.
package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/senran-N/prism/internal/account"
	"github.com/senran-N/prism/internal/db"
	"github.com/senran-N/prism/internal/github"
	"github.com/senran-N/prism/internal/mail"
	"github.com/senran-N/prism/internal/scproto"
)

type Config struct {
	YYDSAPIKey     string
	GitHubUser     string
	GitHubPass     string
	GitHubTOTP     string
	RepoName       string
	RepoID         string
	InitialCredits float64
}

type Scheduler struct {
	mu   sync.Mutex
	pool *account.Pool
	cfg  Config
}

func New(pool *account.Pool, cfg Config) *Scheduler {
	if cfg.InitialCredits == 0 {
		cfg.InitialCredits = 20.0
	}
	return &Scheduler{pool: pool, cfg: cfg}
}

// AcquireAccount returns the current account, or rotates if exhausted.
func (s *Scheduler) AcquireAccount(userID int64) (*account.Account, error) {
	if a := s.pool.Acquire(); a != nil {
		return a, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if a := s.pool.Acquire(); a != nil {
		return a, nil
	}

	// Rotation needed
	if userID > 0 && db.DB != nil {
		balance, err := db.GetUserBalance(userID)
		if err != nil {
			log.Printf("[scheduler] balance check error: %v", err)
		} else if balance < db.RotationCost {
			return nil, fmt.Errorf("insufficient balance (%.2f < %.2f), please recharge", balance, db.RotationCost)
		}
		if err := db.DeductBalance(userID, db.RotationCost); err != nil {
			return nil, fmt.Errorf("deduct balance: %w", err)
		}
		log.Printf("[scheduler] deducted %.2f from user %d for rotation", db.RotationCost, userID)
	}

	log.Println("[scheduler] rotating to new account...")
	return s.rotate(userID)
}

func (s *Scheduler) rotate(userID int64) (*account.Account, error) {
	old := s.pool.FindExhaustedWithGitHub()
	if old != nil && old.Client != nil {
		log.Printf("[scheduler] unbinding GitHub from exhausted %s", old.Email)
		if err := old.Client.DisconnectGitHub(); err != nil {
			log.Printf("[scheduler] unbind warning: %v", err)
		}
		old.GitHubBound = false
	}

	ghClient, err := github.Login(s.cfg.GitHubUser, s.cfg.GitHubPass, s.cfg.GitHubTOTP)
	if err != nil {
		return nil, fmt.Errorf("github login: %w", err)
	}

	prefix := fmt.Sprintf("prism-%d", time.Now().Unix())
	emailAddr, _, err := mail.CreateTempEmail(s.cfg.YYDSAPIKey, prefix)
	if err != nil {
		return nil, fmt.Errorf("create email: %w", err)
	}

	password := fmt.Sprintf("Prism#%06dxPass!", time.Now().Unix()%1000000)
	sc := scproto.NewClient()

	if userID > 0 {
		fp, err := db.GetLatestFingerprint(userID)
		if err == nil && fp != nil && fp.UserAgent != "" {
			sc.SetFingerprint(scproto.FingerprintFromUser(fp.UserAgent, fp.Language, fp.Platform))
		}
	}

	if err := sc.Register(emailAddr, password, "Prism User"); err != nil {
		return nil, fmt.Errorf("sc register: %w", err)
	}
	if err := sc.ConnectGitHub(ghClient); err != nil {
		return nil, fmt.Errorf("sc oauth: %w", err)
	}

	projectID, err := sc.CreateProject(s.cfg.RepoID)
	if err != nil {
		return nil, fmt.Errorf("sc project: %w", err)
	}

	if err := sc.CompleteEnvironmentSetup(projectID, s.cfg.RepoName); err != nil {
		log.Printf("[scheduler] env setup warning: %v", err)
	}
	if err := sc.WaitForEnvironment(projectID, 60*time.Second); err != nil {
		log.Printf("[scheduler] env wait warning: %v", err)
	}

	// Try to get actual credits from SC
	credits := s.cfg.InitialCredits
	if actual, err := sc.GetCredits(); err == nil && actual > 0 {
		credits = actual
		log.Printf("[scheduler] SC reports $%.2f credits", credits)
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
		Status:      account.StatusActive,
		GitHubBound: true,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		Client:      sc,
	}
	s.pool.Add(newAcct)

	log.Printf("[scheduler] new account ready: %s ($%.2f)", emailAddr, credits)
	return newAcct, nil
}

// ReleaseAccount returns the account and syncs credits from SC.
func (s *Scheduler) ReleaseAccount(acct *account.Account) {
	if acct.Client != nil {
		if actual, err := acct.Client.GetCredits(); err == nil && actual >= 0 {
			log.Printf("[scheduler] SC credits for %s: $%.2f", acct.Email, actual)
			if actual < 0.50 {
				s.pool.MarkExhausted(acct.ID)
				log.Printf("[scheduler] account %s exhausted, will rotate next task", acct.Email)
			} else {
				acct.Credits = actual
			}
		}
	}
	s.pool.Release(acct.ID)
}
