// Package api provides the HTTP API for the Prism frontend.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/senran-N/prism/internal/account"
	"github.com/senran-N/prism/internal/config"
	"github.com/senran-N/prism/internal/db"
	"github.com/senran-N/prism/internal/scheduler"
)

type Server struct {
	mux       *http.ServeMux
	scheduler *scheduler.Scheduler
	pool      *account.Pool
	cfg       config.Config

	mu           sync.Mutex
	ghToken      string     // user's GitHub OAuth token
	ghUserName   string     // user's GitHub username
	ghRepos      []RepoInfo // user's repos
	selectedRepo string     // "owner/repo"
}

func New(sched *scheduler.Scheduler, pool *account.Pool, cfg config.Config) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		scheduler: sched,
		pool:      pool,
		cfg:       cfg,
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/models", s.handleModels)
	s.mux.HandleFunc("GET /api/pool/stats", s.handlePoolStats)
	s.mux.HandleFunc("GET /api/pool/accounts", s.handlePoolAccounts)
	s.mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	s.mux.HandleFunc("GET /api/tasks/{id}/status", s.handleTaskStatus)
	// User
	s.mux.HandleFunc("GET /api/me", s.handleMe)

	// SSE
	s.mux.HandleFunc("GET /api/events", s.handleSSE)

	// Admin
	s.mux.HandleFunc("GET /api/admin/stats", s.handleAdminStats)
	s.mux.HandleFunc("GET /api/admin/accounts", s.handleAdminAccounts)
	s.mux.HandleFunc("GET /api/admin/tasks", s.handleAdminTasks)
	s.mux.HandleFunc("GET /api/admin/users", s.handleAdminUsers)
	s.mux.HandleFunc("GET /api/admin/config", s.handleAdminGetConfig)
	s.mux.HandleFunc("POST /api/admin/config", s.handleAdminUpdateConfig)

	// Tasks
	s.mux.HandleFunc("GET /api/tasks/history", s.handleTaskHistory)
	s.mux.HandleFunc("GET /api/github/login", s.handleGitHubLogin)
	s.mux.HandleFunc("GET /api/github/callback", s.handleGitHubCallback)
	s.mux.HandleFunc("POST /api/github/select-repo", s.handleSelectRepo)
	s.mux.HandleFunc("GET /api/github/status", s.handleGitHubOAuthStatus)
	s.mux.HandleFunc("POST /api/github/disconnect", s.handleGitHubOAuthDisconnect)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// GET /api/models
func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	type model struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Agent string `json:"agent"`
	}
	models := []model{
		{"codex_gpt_5_5_medium", "GPT-5.5 (Medium)", "Codex"},
		{"codex_gpt_5_5_high", "GPT-5.5 (High)", "Codex"},
		{"codex_gpt_5_5_xhigh", "GPT-5.5 (Xhigh)", "Codex"},
		{"claude_code_claude_opus_4_8", "Opus 4.8", "Claude Code"},
		{"claude_code_claude_opus_4_7", "Opus 4.7", "Claude Code"},
		{"claude_code_claude_opus_4_6", "Opus 4.6", "Claude Code"},
		{"claude_code_claude_opus_4_5", "Opus 4.5", "Claude Code"},
		{"claude_code_claude_sonnet_4_6", "Sonnet 4.6", "Claude Code"},
		{"opencode_opus_4_8", "Opus 4.8", "OpenCode"},
		{"opencode_sonnet_4_6", "Sonnet 4.6", "OpenCode"},
		{"opencode_gemini_3_1_pro", "Gemini 3.1 Pro", "OpenCode"},
		{"opencode_gemini_3_flash", "Gemini 3 Flash", "OpenCode"},
		{"opencode_gpt_5_5", "GPT-5.5", "OpenCode"},
		{"opencode_gpt_5_5_pro", "GPT-5.5 Pro", "OpenCode"},
		{"pi_deepseek_v4_pro", "DeepSeek V4 Pro", "Pi"},
		{"pi_deepseek_v4_flash", "DeepSeek V4 Flash", "Pi"},
	}
	writeJSON(w, models)
}

// GET /api/pool/stats
func (s *Server) handlePoolStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.pool.Stats())
}

// GET /api/pool/accounts
func (s *Server) handlePoolAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := s.pool.ListAll()
	type acctView struct {
		ID          string  `json:"id"`
		Email       string  `json:"email"`
		Status      string  `json:"status"`
		Credits     float64 `json:"credits"`
		GitHubBound bool    `json:"github_bound"`
		ProjectID   string  `json:"project_id"`
	}
	result := make([]acctView, len(accounts))
	for i, a := range accounts {
		result[i] = acctView{
			ID:          a.ID,
			Email:       a.Email,
			Status:      string(a.Status),
			Credits:     a.Credits,
			GitHubBound: a.GitHubBound,
			ProjectID:   a.ProjectID,
		}
	}
	writeJSON(w, result)
}

// POST /api/tasks
type createTaskReq struct {
	Description string `json:"description"`
	Model       string `json:"model"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.Description == "" {
		writeError(w, 400, "description is required")
		return
	}
	if req.Model == "" {
		req.Model = "claude_code_claude_opus_4_8"
	}

	// Acquire an account (auto-rotates if needed)
	acct, err := s.scheduler.AcquireAccount()
	if err != nil {
		log.Printf("[api] acquire account error: %v", err)
		writeError(w, 503, "no accounts available: "+err.Error())
		return
	}

	// Create ticket
	ticketID, err := acct.Client.CreateTicket(acct.ProjectID, req.Description, req.Model)
	if err != nil {
		s.scheduler.ReleaseAccount(acct.ID)
		writeError(w, 500, "create ticket failed: "+err.Error())
		return
	}

	s.scheduler.ReleaseAccount(acct.ID)

	writeJSON(w, map[string]string{
		"task_id":    ticketID,
		"account_id": acct.ID,
		"model":      req.Model,
		"status":     "created",
		"view_url":   "/proxy/tickets/" + ticketID,
	})
}

// GET /api/tasks/{id}/status
func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	ticketID := r.PathValue("id")

	// Find which account owns this ticket
	for _, acct := range s.pool.ListAll() {
		if acct.Client == nil {
			continue
		}
		status, err := acct.Client.GetTicketStatus(ticketID)
		if err == nil && status.Status != "" {
			writeJSON(w, status)
			return
		}
	}
	writeError(w, 404, "task not found")
}

// GET /api/me
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeJSON(w, map[string]any{"logged_in": false})
		return
	}
	writeJSON(w, map[string]any{
		"logged_in":     true,
		"id":            user.ID,
		"github_login":  user.GitHubLogin,
		"avatar_url":    user.AvatarURL,
		"selected_repo": user.SelectedRepo,
	})
}

// GET /api/tasks/history
func (s *Server) handleTaskHistory(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeError(w, 401, "not logged in")
		return
	}
	if db.DB == nil {
		writeJSON(w, []any{})
		return
	}
	tasks, err := db.ListTasksByUser(user.ID, 50)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, tasks)
}

// CORS middleware for development
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
