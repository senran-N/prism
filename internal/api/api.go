// Package api provides the HTTP API for the Prism frontend.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

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
	registerHoneypots(s.mux)
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Public
	s.mux.HandleFunc("GET /api/models", s.handleModels)
	s.mux.HandleFunc("GET /api/me", s.handleMe)
	s.mux.HandleFunc("POST /api/logout", s.handleUserLogout)

	// OAuth (rate limited)
	s.mux.HandleFunc("GET /api/linuxdo/login", s.rateLimit(10, time.Minute, s.handleLinuxDoLogin))
	s.mux.HandleFunc("GET /api/linuxdo/callback", s.rateLimit(10, time.Minute, s.handleLinuxDoCallback))
	s.mux.HandleFunc("GET /api/linuxdo/status", s.handleLinuxDoStatus)

	// User credits & redemption
	s.mux.HandleFunc("GET /api/balance", s.requireAuth(s.handleGetBalance))
	s.mux.HandleFunc("POST /api/redeem", s.requireAuth(limitBody(1<<10, s.handleRedeem)))
	s.mux.HandleFunc("POST /api/fingerprint", s.requireAuth(limitBody(1<<14, s.handleFingerprint)))

	// Admin: redemption codes
	s.mux.HandleFunc("POST /api/admin/codes", s.requireAdmin(limitBody(1<<14, s.handleAdminCreateCode)))
	s.mux.HandleFunc("GET /api/admin/codes", s.requireAdmin(s.handleAdminListCodes))

	// Credit payment
	s.mux.HandleFunc("POST /api/credit/pay", s.requireAuth(limitBody(1<<16, s.handleCreditPay)))
	s.mux.HandleFunc("GET /api/credit/redirect", s.handleCreditRedirect)
	s.mux.HandleFunc("POST /api/credit/notify", limitBody(1<<16, s.handleCreditNotify))
	s.mux.HandleFunc("GET /api/credit/callback", s.handleCreditCallback)
	s.mux.HandleFunc("GET /api/credit/order", s.requireAuth(s.handleCreditQuery))

	// Authenticated user endpoints
	s.mux.HandleFunc("POST /api/tasks", s.requireAuth(limitBody(1<<20, s.handleCreateTask)))
	s.mux.HandleFunc("GET /api/tasks/{id}/status", s.requireAuth(s.handleTaskStatus))
	s.mux.HandleFunc("GET /api/tasks/{id}/messages", s.requireAuth(s.handleTaskMessages))
	s.mux.HandleFunc("POST /api/tasks/{id}/message", s.requireAuth(limitBody(1<<16, s.handleTaskSendMessage)))
	s.mux.HandleFunc("GET /api/tasks/history", s.requireAuth(s.handleTaskHistory))
	s.mux.HandleFunc("GET /api/events", s.requireAuth(s.handleSSE))

	// Admin auth (public)
	s.mux.HandleFunc("POST /api/admin/login", s.rateLimit(5, time.Minute, limitBody(1<<10, s.handleAdminLogin)))
	s.mux.HandleFunc("POST /api/admin/logout", s.handleAdminLogout)
	s.mux.HandleFunc("GET /api/admin/check", s.handleAdminCheck)

	// Admin only
	s.mux.HandleFunc("GET /api/admin/stats", s.requireAdmin(s.handleAdminStats))
	s.mux.HandleFunc("GET /api/admin/accounts", s.requireAdmin(s.handleAdminAccounts))
	s.mux.HandleFunc("GET /api/admin/tasks", s.requireAdmin(s.handleAdminTasks))
	s.mux.HandleFunc("GET /api/admin/users", s.requireAdmin(s.handleAdminUsers))
	s.mux.HandleFunc("GET /api/admin/config", s.requireAdmin(s.handleAdminGetConfig))
	s.mux.HandleFunc("POST /api/admin/config", s.requireAdmin(limitBody(1<<16, s.handleAdminUpdateConfig)))
	s.mux.HandleFunc("POST /api/admin/users/{id}/credits", s.requireAdmin(limitBody(1<<14, s.handleAdminAddCredits)))
	s.mux.HandleFunc("POST /api/admin/users/{id}/ban", s.requireAdmin(limitBody(1<<16, s.handleAdminBanUser)))
	s.mux.HandleFunc("POST /api/admin/users/{id}/unban", s.requireAdmin(s.handleAdminUnbanUser))

	// Pool info (admin only)
	s.mux.HandleFunc("GET /api/pool/stats", s.requireAdmin(s.handlePoolStats))
	s.mux.HandleFunc("GET /api/pool/accounts", s.requireAdmin(s.handlePoolAccounts))
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

	// Get current user ID for billing
	var userID int64
	if user := s.getSessionUser(r); user != nil {
		userID = user.ID
	}

	// Acquire an account (auto-rotates if needed, bills user)
	acct, err := s.scheduler.AcquireAccount(userID)
	if err != nil {
		log.Printf("[api] acquire account error: %v", err)
		writeError(w, 503, "no accounts available: "+err.Error())
		return
	}

	// Create ticket
	ticketID, err := acct.Client.CreateTicket(acct.ProjectID, req.Description, req.Model)
	if err != nil {
		s.scheduler.ReleaseAccount(acct.ID, false)
		writeError(w, 500, "create ticket failed: "+err.Error())
		return
	}

	// Release account back to pool and deduct estimated credits
	s.scheduler.ReleaseAccount(acct.ID, true)

	// Map ticket to account so proxy can use the right session
	s.pool.MapTicket(ticketID, acct.ID)

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

// POST /api/logout
func (s *Server) handleUserLogout(w http.ResponseWriter, r *http.Request) {
	s.clearSession(w)
	writeJSON(w, map[string]bool{"logged_out": true})
}

// GET /api/me
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeJSON(w, map[string]any{"logged_in": false})
		return
	}
	writeJSON(w, map[string]any{
		"logged_in":         true,
		"id":                user.ID,
		"github_login":      user.GitHubLogin,
		"avatar_url":        user.AvatarURL,
		"selected_repo":     user.SelectedRepo,
		"linuxdo_username":  user.LinuxDoUsername,
		"linuxdo_name":      user.LinuxDoName,
		"trust_level":       user.TrustLevel,
		"is_banned":         user.IsBanned,
		"ban_reason":        user.BanReason,
		"is_admin":          user.IsAdmin,
		"balance":           user.Balance,
		"total_rotations":   user.TotalRotations,
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

