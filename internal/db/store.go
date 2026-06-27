package db

import (
	"database/sql"
	"time"
)

// ── User ────────────────────────────────────────

type User struct {
	ID           int64  `json:"id"`
	GitHubID     int64  `json:"github_id"`
	GitHubLogin  string `json:"github_login"`
	AvatarURL    string `json:"avatar_url"`
	GitHubToken  string `json:"-"`
	SelectedRepo string `json:"selected_repo"`
	CreatedAt    time.Time `json:"created_at"`
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
		RETURNING id, github_id, github_login, avatar_url, selected_repo, created_at
	`, githubID, login, avatarURL, token).Scan(
		&u.ID, &u.GitHubID, &u.GitHubLogin, &u.AvatarURL, &u.SelectedRepo, &u.CreatedAt,
	)
	u.GitHubToken = token
	return u, err
}

func GetUser(id int64) (*User, error) {
	u := &User{}
	err := DB.QueryRow(`
		SELECT id, github_id, github_login, avatar_url, github_token, selected_repo, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.GitHubID, &u.GitHubLogin, &u.AvatarURL, &u.GitHubToken, &u.SelectedRepo, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
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
