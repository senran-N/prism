package api

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/senran-N/prism/internal/db"
)

var startedAt = time.Now()

// GET /api/admin/stats
func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	stats := s.pool.Stats()

	var taskCount, userCount int
	if db.DB != nil {
		db.DB.QueryRow("SELECT count(*) FROM tasks").Scan(&taskCount)
		db.DB.QueryRow("SELECT count(*) FROM users").Scan(&userCount)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	writeJSON(w, map[string]any{
		"pool":    stats,
		"tasks":   taskCount,
		"users":   userCount,
		"uptime":  time.Since(startedAt).String(),
		"go_routines": runtime.NumGoroutine(),
		"mem_mb":  mem.Alloc / 1024 / 1024,
	})
}

// GET /api/admin/accounts
func (s *Server) handleAdminAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := s.pool.ListAll()
	type view struct {
		ID          string  `json:"id"`
		Email       string  `json:"email"`
		WorkspaceID string  `json:"workspace_id"`
		ProjectID   string  `json:"project_id"`
		Credits     float64 `json:"credits"`
		Status      string  `json:"status"`
		GitHubBound bool    `json:"github_bound"`
		CreatedAt   string  `json:"created_at"`
		LastUsedAt  string  `json:"last_used_at"`
	}
	result := make([]view, 0, len(accounts))
	for _, a := range accounts {
		v := view{
			ID: a.ID, Email: a.Email, WorkspaceID: a.WorkspaceID,
			ProjectID: a.ProjectID, Credits: a.Credits,
			Status: string(a.Status), GitHubBound: a.GitHubBound,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
		}
		if !a.LastUsedAt.IsZero() {
			v.LastUsedAt = a.LastUsedAt.Format(time.RFC3339)
		}
		result = append(result, v)
	}
	writeJSON(w, result)
}

// GET /api/admin/tasks
func (s *Server) handleAdminTasks(w http.ResponseWriter, r *http.Request) {
	if db.DB == nil {
		writeJSON(w, []any{})
		return
	}
	rows, err := db.DB.Query(`
		SELECT t.id, t.ticket_id, t.description, t.model, t.status, t.cost,
		       t.created_at, COALESCE(u.github_login, '') as user_login
		FROM tasks t LEFT JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC LIMIT 100
	`)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()

	type taskView struct {
		ID          string    `json:"id"`
		TicketID    string    `json:"ticket_id"`
		Description string    `json:"description"`
		Model       string    `json:"model"`
		Status      string    `json:"status"`
		Cost        float64   `json:"cost"`
		CreatedAt   time.Time `json:"created_at"`
		UserLogin   string    `json:"user_login"`
	}
	var tasks []taskView
	for rows.Next() {
		var t taskView
		rows.Scan(&t.ID, &t.TicketID, &t.Description, &t.Model, &t.Status, &t.Cost, &t.CreatedAt, &t.UserLogin)
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []taskView{}
	}
	writeJSON(w, tasks)
}

// GET /api/admin/users
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if db.DB == nil {
		writeJSON(w, []any{})
		return
	}
	rows, err := db.DB.Query(`
		SELECT u.id, u.github_login, u.avatar_url, u.selected_repo, u.created_at,
		       count(t.id) as task_count
		FROM users u LEFT JOIN tasks t ON u.id = t.user_id
		GROUP BY u.id ORDER BY u.created_at DESC LIMIT 100
	`)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()

	type userView struct {
		ID           int64     `json:"id"`
		GitHubLogin  string    `json:"github_login"`
		AvatarURL    string    `json:"avatar_url"`
		SelectedRepo string    `json:"selected_repo"`
		CreatedAt    time.Time `json:"created_at"`
		TaskCount    int       `json:"task_count"`
	}
	var users []userView
	for rows.Next() {
		var u userView
		rows.Scan(&u.ID, &u.GitHubLogin, &u.AvatarURL, &u.SelectedRepo, &u.CreatedAt, &u.TaskCount)
		users = append(users, u)
	}
	if users == nil {
		users = []userView{}
	}
	writeJSON(w, users)
}

// POST /api/admin/config
func (s *Server) handleAdminUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GitHubUser string `json:"github_user"`
		GitHubPass string `json:"github_pass"`
		GitHubTOTP string `json:"github_totp"`
		YYDSAPIKey string `json:"yyds_api_key"`
		RepoID     string `json:"repo_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}

	s.mu.Lock()
	if req.GitHubUser != "" {
		s.cfg.GitHubUser = req.GitHubUser
	}
	if req.GitHubPass != "" {
		s.cfg.GitHubPass = req.GitHubPass
	}
	if req.GitHubTOTP != "" {
		s.cfg.GitHubTOTP = req.GitHubTOTP
	}
	if req.YYDSAPIKey != "" {
		s.cfg.YYDSAPIKey = req.YYDSAPIKey
	}
	if req.RepoID != "" {
		s.cfg.RepoID = req.RepoID
	}
	s.mu.Unlock()

	log.Println("[admin] config updated")
	writeJSON(w, map[string]bool{"updated": true})
}

// GET /api/admin/config
func (s *Server) handleAdminGetConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, map[string]any{
		"github_user":       s.cfg.GitHubUser,
		"github_totp":       maskSecret(s.cfg.GitHubTOTP),
		"github_pass":       maskSecret(s.cfg.GitHubPass),
		"yyds_api_key":      maskSecret(s.cfg.YYDSAPIKey),
		"repo_id":           s.cfg.RepoID,
		"github_client_id":  s.cfg.GitHubClientID,
		"base_url":          s.cfg.BaseURL,
	})
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}
