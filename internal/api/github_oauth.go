package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/senran-N/prism/internal/db"
)

// GET /api/github/login — redirect user to GitHub OAuth
func (s *Server) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := generateOAuthState()
	params := url.Values{
		"client_id":    {s.cfg.GitHubClientID},
		"redirect_uri": {s.cfg.BaseURL + "/api/github/callback"},
		"scope":        {"repo,read:user"},
		"state":        {state},
	}
	http.Redirect(w, r, "https://github.com/login/oauth/authorize?"+params.Encode(), http.StatusFound)
}

// GET /api/github/callback — GitHub redirects here with ?code=
func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		http.Error(w, "missing code", 400)
		return
	}
	if !validateOAuthState(state) {
		http.Error(w, "invalid state parameter", 400)
		return
	}

	// Exchange code for access token
	token, err := exchangeGitHubCode(s.cfg.GitHubClientID, s.cfg.GitHubClientSecret, code)
	if err != nil {
		log.Printf("[oauth] token exchange error: %v", err)
		http.Error(w, "token exchange failed", 500)
		return
	}

	// Get user info
	user, err := getGitHubUser(token)
	if err != nil {
		log.Printf("[oauth] get user error: %v", err)
		http.Error(w, "failed to get user info", 500)
		return
	}

	// Get user's repos
	repos, err := getGitHubRepos(token)
	if err != nil {
		log.Printf("[oauth] get repos error: %v", err)
	}

	// Link GitHub to existing logged-in user (don't create new user)
	if db.DB != nil {
		existingUser := s.getSessionUser(r)
		if existingUser != nil {
			// User is already logged in (via LinuxDo) — link GitHub to their account
			err := db.LinkGitHub(existingUser.ID, user.ID, user.Login, user.AvatarURL, token)
			if err != nil {
				log.Printf("[oauth] link github error: %v", err)
				http.Redirect(w, r, s.cfg.BaseURL+"/?error=github_link_failed", http.StatusFound)
				return
			}
			log.Printf("[oauth] linked GitHub %s to user %d (%s)", user.Login, existingUser.ID, existingUser.LinuxDoUsername)
		} else {
			// No existing session — create/update user from GitHub (standalone login)
			dbUser, err := db.UpsertUser(user.ID, user.Login, user.AvatarURL, token)
			if err != nil {
				log.Printf("[oauth] db upsert error: %v", err)
				http.Redirect(w, r, s.cfg.BaseURL+"/?error=github_save_failed", http.StatusFound)
				return
			}
			s.setSession(w, dbUser.ID)
			log.Printf("[oauth] github login: id=%d login=%s", dbUser.ID, dbUser.GitHubLogin)
		}
	}

	// Store in server state
	s.mu.Lock()
	s.ghToken = token
	s.ghUserName = user.Login
	s.ghRepos = repos
	s.mu.Unlock()

	log.Printf("[oauth] GitHub connected: %s (%d repos)", user.Login, len(repos))
	http.Redirect(w, r, s.cfg.BaseURL+"/?github=connected", http.StatusFound)
}

// POST /api/github/select-repo — user picks a repo, we add service account as collaborator
func (s *Server) handleSelectRepo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Repo string `json:"repo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Repo == "" {
		writeError(w, 400, "repo is required (owner/name format)")
		return
	}

	// Get token from current user's DB record
	user := s.getSessionUser(r)
	token := ""
	if user != nil && user.GitHubToken != "" {
		token = user.GitHubToken
	}
	if token == "" {
		s.mu.Lock()
		token = s.ghToken
		s.mu.Unlock()
	}
	if token == "" {
		writeError(w, 401, "GitHub not connected. Please connect GitHub first.")
		return
	}

	serviceAccount := s.cfg.GitHubUser
	log.Printf("[oauth] adding %s as collaborator to %s", serviceAccount, req.Repo)
	err := addCollaborator(token, req.Repo, serviceAccount)
	if err != nil {
		log.Printf("[oauth] add collaborator error: %v", err)
		writeError(w, 500, fmt.Sprintf("failed to add collaborator: %v", err))
		return
	}

	// Store in DB
	if user != nil && db.DB != nil {
		db.UpdateUserRepo(user.ID, req.Repo)
	}

	s.mu.Lock()
	s.selectedRepo = req.Repo
	s.mu.Unlock()

	log.Printf("[oauth] repo selected: %s, collaborator added: %s", req.Repo, serviceAccount)
	writeJSON(w, map[string]any{"repo": req.Repo, "collaborator": serviceAccount, "status": "ready"})
}

// GET /api/github/status — read from DB, not just memory
func (s *Server) handleGitHubOAuthStatus(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)

	// DB-backed state
	if user != nil && user.GitHubLogin != "" {
		repos := s.ghRepos
		if repos == nil && user.GitHubToken != "" {
			repos, _ = getGitHubRepos(user.GitHubToken)
			if repos != nil {
				s.mu.Lock()
				s.ghRepos = repos
				s.mu.Unlock()
			}
		}
		writeJSON(w, map[string]any{
			"connected":    true,
			"user":         user.GitHubLogin,
			"repos":        repos,
			"selectedRepo": user.SelectedRepo,
		})
		return
	}

	// Fallback to memory state
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, map[string]any{
		"connected":    s.ghToken != "",
		"user":         s.ghUserName,
		"repos":        s.ghRepos,
		"selectedRepo": s.selectedRepo,
	})
}

// POST /api/github/disconnect
func (s *Server) handleGitHubOAuthDisconnect(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user != nil && db.DB != nil {
		db.DB.Exec("UPDATE users SET github_id = NULL, github_login = '', github_token = '', selected_repo = '', updated_at = now() WHERE id = $1", user.ID)
	}

	s.mu.Lock()
	s.ghToken = ""
	s.ghUserName = ""
	s.ghRepos = nil
	s.selectedRepo = ""
	s.mu.Unlock()

	writeJSON(w, map[string]bool{"disconnected": true})
}

// ── GitHub API helpers ──────────────────────────────────────

func exchangeGitHubCode(clientID, clientSecret, code string) (string, error) {
	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		return "", fmt.Errorf("oauth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func getGitHubUser(token string) (*GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	json.NewDecoder(resp.Body).Decode(&user)
	return &user, nil
}

type RepoInfo struct {
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

func getGitHubRepos(token string) ([]RepoInfo, error) {
	var all []RepoInfo
	page := 1
	for {
		reqURL := fmt.Sprintf("https://api.github.com/user/repos?per_page=100&page=%d&sort=updated", page)
		req, _ := http.NewRequest("GET", reqURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return all, err
		}
		defer resp.Body.Close()

		var repos []RepoInfo
		json.NewDecoder(resp.Body).Decode(&repos)
		if len(repos) == 0 {
			break
		}
		all = append(all, repos...)
		if len(repos) < 100 {
			break
		}
		page++
	}
	return all, nil
}

func addCollaborator(token, repo, username string) error {
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", repo, username)
	body := strings.NewReader(`{"permission":"push"}`)
	req, _ := http.NewRequest("PUT", reqURL, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 || resp.StatusCode == 204 {
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("GitHub API %d: %s", resp.StatusCode, string(respBody))
}
