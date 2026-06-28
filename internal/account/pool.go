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
	StatusReady      Status = "ready"
	StatusActive     Status = "active"
	StatusExhausted  Status = "exhausted"
	StatusError      Status = "error"
	StatusRegistered Status = "registered"
)

// Account represents one Superconductor account in the pool.
type Account struct {
	ID          string          `json:"id"`
	Email       string          `json:"email"`
	Password    string          `json:"password"`
	WorkspaceID string          `json:"workspace_id"`
	UserID      string          `json:"user_id"`
	ProjectID   string          `json:"project_id"`
	RepoID      string          `json:"repo_id"`
	Credits     float64         `json:"credits"`
	Status      Status          `json:"status"`
	GitHubBound bool            `json:"github_bound"`
	CreatedAt   time.Time       `json:"created_at"`
	LastUsedAt  time.Time       `json:"last_used_at"`
	Client      *scproto.Client `json:"-"`
}

// Pool manages a set of SC accounts.
type Pool struct {
	mu       sync.RWMutex
	accounts map[string]*Account
	tickets  map[string]string // ticketID → accountID
}

func NewPool() *Pool {
	return &Pool{
		accounts: make(map[string]*Account),
		tickets:  make(map[string]string),
	}
}

func (p *Pool) MapTicket(ticketID, accountID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickets[ticketID] = accountID
}

func (p *Pool) GetTicketAccount(ticketID string) *Account {
	p.mu.RLock()
	defer p.mu.RUnlock()
	acctID, ok := p.tickets[ticketID]
	if !ok {
		return nil
	}
	return p.accounts[acctID]
}

func (p *Pool) GetProjectAccount(projectID string) *Account {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, a := range p.accounts {
		if a.ProjectID == projectID {
			return a
		}
	}
	return nil
}

func (p *Pool) Add(a *Account) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.accounts[a.ID] = a
}

func (p *Pool) Get(id string) (*Account, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	a, ok := p.accounts[id]
	return a, ok
}

// Acquire returns a ready account with remaining credits.
// Prefers accounts with GitHub bound, falls back to any ready account.
func (p *Pool) Acquire() *Account {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Prefer accounts with GitHub (can create PRs)
	var fallback *Account
	for _, a := range p.accounts {
		if a.Status == StatusReady && a.Credits > 0.50 {
			if a.GitHubBound {
				a.Status = StatusActive
				a.LastUsedAt = time.Now()
				return a
			}
			if fallback == nil {
				fallback = a
			}
		}
	}
	if fallback != nil {
		fallback.Status = StatusActive
		fallback.LastUsedAt = time.Now()
		return fallback
	}
	return nil
}

// Release marks an account as ready again.
func (p *Pool) Release(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if a, ok := p.accounts[id]; ok && a.Status == StatusActive {
		a.Status = StatusReady
	}
}

func (p *Pool) MarkExhausted(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if a, ok := p.accounts[id]; ok {
		a.Status = StatusExhausted
		a.Credits = 0
	}
}

// DeductCredits subtracts estimated cost. Marks exhausted when empty.
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

func (p *Pool) ListAll() []*Account {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*Account, 0, len(p.accounts))
	for _, a := range p.accounts {
		result = append(result, a)
	}
	return result
}

type PoolStats struct {
	Total        int     `json:"total"`
	Ready        int     `json:"ready"`
	Active       int     `json:"active"`
	Exhausted    int     `json:"exhausted"`
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
		if a.Status == StatusExhausted && a.GitHubBound {
			return a
		}
	}
	return nil
}

func GenerateAccountID() string {
	return fmt.Sprintf("sc-%d", time.Now().UnixNano())
}
