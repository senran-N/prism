// Package scheduler orchestrates account rotation and task dispatch.
// When credits run out, it transparently provisions a new account.
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

type Config struct {
	YYDSAPIKey      string
	GitHubUser      string
	GitHubPass      string
	GitHubTOTP      string
	RepoID          string
	InitialCredits  float64 // default 20.0
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

// AcquireAccount returns a ready account. If none available, performs
// automatic rotation: unbind GitHub from exhausted account → register
// new account → bind GitHub → create project.
func (s *Scheduler) AcquireAccount() (*account.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if a := s.pool.Acquire(); a != nil {
		return a, nil
	}

	log.Println("[scheduler] no ready accounts, starting rotation...")
	return s.rotate()
}

func (s *Scheduler) rotate() (*account.Account, error) {
	// 1. Find an account with GitHub bound and unbind it
	old := s.pool.FindExhaustedWithGitHub()
	if old != nil && old.Client != nil {
		log.Printf("[scheduler] unbinding GitHub from %s", old.Email)
		if err := old.Client.DisconnectGitHub(); err != nil {
			log.Printf("[scheduler] unbind warning: %v", err)
		}
		old.GitHubBound = false
	}

	// 2. Login GitHub
	log.Println("[scheduler] logging into GitHub...")
	ghClient, err := github.Login(s.cfg.GitHubUser, s.cfg.GitHubPass, s.cfg.GitHubTOTP)
	if err != nil {
		return nil, fmt.Errorf("github login: %w", err)
	}

	// 3. Create temp email
	log.Println("[scheduler] creating temp email...")
	prefix := fmt.Sprintf("prism-%d", time.Now().Unix())
	emailAddr, _, err := mail.CreateTempEmail(s.cfg.YYDSAPIKey, prefix)
	if err != nil {
		return nil, fmt.Errorf("create email: %w", err)
	}

	// 4. Register SC account
	password := fmt.Sprintf("Prism#%06dxPass!", time.Now().Unix()%1000000)
	sc := scproto.NewClient()
	log.Printf("[scheduler] registering SC: %s", emailAddr)
	if err := sc.Register(emailAddr, password, "Prism User"); err != nil {
		return nil, fmt.Errorf("sc register: %w", err)
	}

	// 5. Connect GitHub OAuth
	log.Println("[scheduler] connecting GitHub OAuth...")
	if err := sc.ConnectGitHub(ghClient); err != nil {
		return nil, fmt.Errorf("sc oauth: %w", err)
	}

	// 6. Create project
	log.Println("[scheduler] creating project...")
	projectID, err := sc.CreateProject(s.cfg.RepoID)
	if err != nil {
		return nil, fmt.Errorf("sc project: %w", err)
	}

	// 7. Wait for environment setup
	log.Println("[scheduler] waiting for environment setup (30s)...")
	time.Sleep(30 * time.Second)

	// 8. Add to pool
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

	log.Printf("[scheduler] new account ready: %s ($%.2f)", emailAddr, newAcct.Credits)

	newAcct.Status = account.StatusActive
	newAcct.LastUsedAt = time.Now()
	return newAcct, nil
}

// ReleaseAccount returns the account to the pool.
func (s *Scheduler) ReleaseAccount(id string) {
	s.pool.Release(id)
}

// RecordUsage deducts cost and checks if rotation is needed.
func (s *Scheduler) RecordUsage(id string, cost float64) {
	s.pool.DeductCredits(id, cost)
}
