// Package account manages a pool of Superconductor accounts with
// automatic rotation when credits are exhausted.
package account

import (
	"fmt"
	"sync"
	"time"

	"github.com/senran-N/prism/internal/scproto"
)

type Status string

const (
	StatusReady      Status = "ready"       // logged in, GitHub connected, has credits
	StatusActive     Status = "active"      // currently serving a task
	StatusExhausted  Status = "exhausted"   // credits depleted
	StatusError      Status = "error"       // login or setup failure
	StatusRegistered Status = "registered"  // registered but not yet set up
)

// Account represents one Superconductor account in the pool.
type Account struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	ProjectID   string    `json:"project_id"`
	RepoID      string    `json:"repo_id"`
	Credits     float64   `json:"credits"` // estimated remaining credits in USD
	Status      Status    `json:"status"`
	GitHubBound bool      `json:"github_bound"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	Client      *scproto.Client `json:"-"`
}

// Pool manages a set of SC accounts and handles automatic rotation.
// Uses channel-based ready queue for lock-free acquisition.
type Pool struct {
	mu       sync.RWMutex
	accounts map[string]*Account
	active   string
	ready    chan *Account // buffered channel for fast acquire
}

func NewPool() *Pool {
	return &Pool{
		accounts: make(map[string]*Account),
		ready:    make(chan *Account, 50),
	}
}

// Add registers an account in the pool. If ready, also enqueue.
func (p *Pool) Add(a *Account) {
	p.mu.Lock()
	p.accounts[a.ID] = a
	p.mu.Unlock()

	if a.Status == StatusReady && a.Credits > 0.50 {
		select {
		case p.ready <- a:
		default: // channel full, that's ok
		}
	}
}

// Get returns an account by ID.
func (p *Pool) Get(id string) (*Account, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	a, ok := p.accounts[id]
	return a, ok
}

// Acquire selects a ready account. Fast path via channel (no lock).
func (p *Pool) Acquire() *Account {
	// Fast path: try channel (lock-free)
	select {
	case a := <-p.ready:
		if a.Status == StatusReady && a.Credits > 0.50 {
			p.mu.Lock()
			a.Status = StatusActive
			a.LastUsedAt = time.Now()
			p.active = a.ID
			p.mu.Unlock()
			return a
		}
		// Account no longer valid, fall through
	default:
	}

	// Slow path: scan map
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, a := range p.accounts {
		if a.Status == StatusReady && a.Credits > 0.50 && a.GitHubBound {
			a.Status = StatusActive
			a.LastUsedAt = time.Now()
			p.active = a.ID
			return a
		}
	}
	return nil
}

// Release marks an account as ready again after task completion.
func (p *Pool) Release(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if a, ok := p.accounts[id]; ok {
		if a.Status == StatusActive {
			a.Status = StatusReady
		}
	}
}

// MarkExhausted flags an account as having no remaining credits.
func (p *Pool) MarkExhausted(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if a, ok := p.accounts[id]; ok {
		a.Status = StatusExhausted
		a.Credits = 0
	}
}

// DeductCredits subtracts estimated cost from an account.
func (p *Pool) DeductCredits(id string, amount float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if a, ok := p.accounts[id]; ok {
		a.Credits -= amount
		if a.Credits <= 0 {
			a.Credits = 0
			a.Status = StatusExhausted
		}
	}
}

// ListAll returns a snapshot of all accounts.
func (p *Pool) ListAll() []*Account {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*Account, 0, len(p.accounts))
	for _, a := range p.accounts {
		result = append(result, a)
	}
	return result
}

// Stats returns pool-level statistics.
type PoolStats struct {
	Total      int     `json:"total"`
	Ready      int     `json:"ready"`
	Active     int     `json:"active"`
	Exhausted  int     `json:"exhausted"`
	TotalCredits float64 `json:"total_credits"`
}

func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var s PoolStats
	s.Total = len(p.accounts)
	for _, a := range p.accounts {
		switch a.Status {
		case StatusReady:
			s.Ready++
		case StatusActive:
			s.Active++
		case StatusExhausted:
			s.Exhausted++
		}
		s.TotalCredits += a.Credits
	}
	return s
}

// NeedsRotation returns true if no ready accounts are available.
func (p *Pool) NeedsRotation() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, a := range p.accounts {
		if a.Status == StatusReady && a.Credits > 0.50 {
			return false
		}
	}
	return true
}

// FindExhaustedWithGitHub returns an exhausted account that still has
// GitHub bound, suitable for unbinding during rotation.
func (p *Pool) FindExhaustedWithGitHub() *Account {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, a := range p.accounts {
		if a.GitHubBound {
			return a
		}
	}
	return nil
}

func GenerateAccountID() string {
	return fmt.Sprintf("sc-%d", time.Now().UnixNano())
}
