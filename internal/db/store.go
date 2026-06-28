package db

import (
	"database/sql"
	"strings"
	"time"
)

// ── User ────────────────────────────────────────

type User struct {
	ID              int64     `json:"id"`
	GitHubID        int64     `json:"github_id"`
	GitHubLogin     string    `json:"github_login"`
	AvatarURL       string    `json:"avatar_url"`
	GitHubToken     string    `json:"-"`
	SelectedRepo    string    `json:"selected_repo"`
	LinuxDoID       int64     `json:"linuxdo_id"`
	LinuxDoUsername  string    `json:"linuxdo_username"`
	LinuxDoName     string    `json:"linuxdo_name"`
	TrustLevel      int       `json:"trust_level"`
	IsBanned        bool      `json:"is_banned"`
	BanReason       string    `json:"ban_reason"`
	IsAdmin         bool      `json:"is_admin"`
	Balance         float64   `json:"balance"`
	TotalRotations  int       `json:"total_rotations"`
	CreatedAt       time.Time `json:"created_at"`
}

func BanUser(id int64, reason string) error {
	_, err := DB.Exec(`UPDATE users SET is_banned = true, ban_reason = $1, updated_at = now() WHERE id = $2`, reason, id)
	return err
}

func UnbanUser(id int64) error {
	_, err := DB.Exec(`UPDATE users SET is_banned = false, ban_reason = '', updated_at = now() WHERE id = $1`, id)
	return err
}

func UpsertUser(githubID int64, login, avatarURL, token string) (*User, error) {
	u := &User{}
	err := DB.QueryRow(`
		INSERT INTO users (github_id, github_login, avatar_url, github_token)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (github_id) DO UPDATE SET
			github_login = EXCLUDED.github_login,
			avatar_url = EXCLUDED.avatar_url,
			github_token = EXCLUDED.github_token,
			updated_at = now()
		RETURNING id, COALESCE(github_id, 0), COALESCE(github_login, ''), COALESCE(avatar_url, ''),
		          COALESCE(selected_repo, ''), created_at
	`, githubID, login, avatarURL, token).Scan(
		&u.ID, &u.GitHubID, &u.GitHubLogin, &u.AvatarURL, &u.SelectedRepo, &u.CreatedAt,
	)
	u.GitHubToken = token
	return u, err
}

func GetUser(id int64) (*User, error) {
	u := &User{}
	err := DB.QueryRow(`
		SELECT id, COALESCE(github_id, 0), COALESCE(github_login, ''), COALESCE(avatar_url, ''),
		       COALESCE(github_token, ''), COALESCE(selected_repo, ''),
		       COALESCE(linuxdo_id, 0), COALESCE(linuxdo_username, ''), COALESCE(linuxdo_name, ''),
		       COALESCE(trust_level, 0), COALESCE(is_banned, false), COALESCE(ban_reason, ''),
		       COALESCE(is_admin, false), COALESCE(balance, 0), COALESCE(total_rotations, 0), created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.GitHubID, &u.GitHubLogin, &u.AvatarURL, &u.GitHubToken, &u.SelectedRepo,
		&u.LinuxDoID, &u.LinuxDoUsername, &u.LinuxDoName, &u.TrustLevel,
		&u.IsBanned, &u.BanReason, &u.IsAdmin,
		&u.Balance, &u.TotalRotations, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func UpsertLinuxDoUser(linuxdoID int64, username, name, avatarTemplate string, trustLevel int) (*User, error) {
	avatarURL := avatarTemplate
	if strings.Contains(avatarURL, "{size}") {
		avatarURL = strings.Replace(avatarURL, "{size}", "120", 1)
	}
	if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
		avatarURL = "https://linux.do" + avatarURL
	}

	u := &User{}
	err := DB.QueryRow(`
		INSERT INTO users (linuxdo_id, linuxdo_username, linuxdo_name, avatar_url, trust_level)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (linuxdo_id) DO UPDATE SET
			linuxdo_username = EXCLUDED.linuxdo_username,
			linuxdo_name = EXCLUDED.linuxdo_name,
			avatar_url = EXCLUDED.avatar_url,
			trust_level = EXCLUDED.trust_level,
			updated_at = now()
		RETURNING id, COALESCE(github_id, 0), COALESCE(github_login, ''), COALESCE(avatar_url, ''),
		          COALESCE(selected_repo, ''), COALESCE(linuxdo_id, 0), COALESCE(linuxdo_username, ''),
		          COALESCE(linuxdo_name, ''), COALESCE(trust_level, 0), created_at
	`, linuxdoID, username, name, avatarURL, trustLevel).Scan(
		&u.ID, &u.GitHubID, &u.GitHubLogin, &u.AvatarURL,
		&u.SelectedRepo, &u.LinuxDoID, &u.LinuxDoUsername,
		&u.LinuxDoName, &u.TrustLevel, &u.CreatedAt,
	)
	return u, err
}

func LinkGitHub(userID, githubID int64, login, avatarURL, token string) error {
	_, err := DB.Exec(`
		UPDATE users SET github_id = $1, github_login = $2, avatar_url = $3, github_token = $4, updated_at = now()
		WHERE id = $5
	`, githubID, login, avatarURL, token, userID)
	return err
}

func UpdateUserRepo(id int64, repo string) error {
	_, err := DB.Exec(`UPDATE users SET selected_repo = $1, updated_at = now() WHERE id = $2`, repo, id)
	return err
}

// ── Task ────────────────────────────────────────

type Task struct {
	ID          string    `json:"id"`
	UserID      int64     `json:"user_id"`
	AccountID   string    `json:"account_id"`
	TicketID    string    `json:"ticket_id"`
	ProjectID   string    `json:"project_id"`
	Description string    `json:"description"`
	Model       string    `json:"model"`
	Status      string    `json:"status"`
	Cost        float64   `json:"cost"`
	ViewURL     string    `json:"view_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func InsertTask(t *Task) error {
	_, err := DB.Exec(`
		INSERT INTO tasks (id, user_id, account_id, ticket_id, project_id, description, model, status, view_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, t.ID, t.UserID, t.AccountID, t.TicketID, t.ProjectID, t.Description, t.Model, t.Status, t.ViewURL)
	return err
}

func UpdateTaskStatus(id, status string, cost float64) error {
	_, err := DB.Exec(`
		UPDATE tasks SET status = $1, cost = $2, updated_at = now() WHERE id = $3
	`, status, cost, id)
	return err
}

func ListTasksByUser(userID int64, limit int) ([]Task, error) {
	rows, err := DB.Query(`
		SELECT id, user_id, account_id, ticket_id, project_id, description, model, status, cost, view_url, created_at
		FROM tasks WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccountID, &t.TicketID, &t.ProjectID, &t.Description, &t.Model, &t.Status, &t.Cost, &t.ViewURL, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// ── SC Account ──────────────────────────────────

type SCAccount struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Password    string  `json:"-"`
	WorkspaceID string  `json:"workspace_id"`
	ProjectID   string  `json:"project_id"`
	Credits     float64 `json:"credits"`
	Status      string  `json:"status"`
	GitHubBound bool    `json:"github_bound"`
}

func InsertSCAccount(a *SCAccount) error {
	_, err := DB.Exec(`
		INSERT INTO sc_accounts (id, email, password, workspace_id, user_id, project_id, repo_id, credits, status, github_bound)
		VALUES ($1, $2, $3, $4, '', $5, '', $6, $7, $8)
	`, a.ID, a.Email, a.Password, a.WorkspaceID, a.ProjectID, a.Credits, a.Status, a.GitHubBound)
	return err
}

func UpdateSCAccountCredits(id string, credits float64, status string) error {
	_, err := DB.Exec(`
		UPDATE sc_accounts SET credits = $1, status = $2, last_used_at = now() WHERE id = $3
	`, credits, status, id)
	return err
}

func ListSCAccounts() ([]SCAccount, error) {
	rows, err := DB.Query(`SELECT id, email, workspace_id, project_id, credits, status, github_bound FROM sc_accounts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []SCAccount
	for rows.Next() {
		var a SCAccount
		if err := rows.Scan(&a.ID, &a.Email, &a.WorkspaceID, &a.ProjectID, &a.Credits, &a.Status, &a.GitHubBound); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}
