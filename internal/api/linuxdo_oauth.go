package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/senran-N/prism/internal/db"
)

const (
	linuxdoAuthorize = "https://connect.linux.do/oauth2/authorize"
	linuxdoToken     = "https://connect.linux.do/oauth2/token"
	linuxdoUserAPI   = "https://connect.linux.do/api/user"
)

// GET /api/linuxdo/login
func (s *Server) handleLinuxDoLogin(w http.ResponseWriter, r *http.Request) {
	state := generateOAuthState()
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {s.cfg.LinuxDoClientID},
		"state":         {state},
	}
	// Only include redirect_uri if explicitly configured
	if s.cfg.LinuxDoRedirectURI != "" {
		params.Set("redirect_uri", s.cfg.LinuxDoRedirectURI)
	}
	http.Redirect(w, r, linuxdoAuthorize+"?"+params.Encode(), http.StatusFound)
}

// GET /api/linuxdo/callback
func (s *Server) handleLinuxDoCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		http.Error(w, "missing code", 400)
		return
	}
	if !validateOAuthState(state) {
		log.Printf("[linuxdo] invalid state: %s (may have expired or server restarted)", state)
		// Redirect to login page instead of showing error
		http.Redirect(w, r, s.cfg.BaseURL+"/?error=state_expired", http.StatusFound)
		return
	}

	// Exchange code for token
	redirectURI := s.cfg.LinuxDoRedirectURI
	token, err := exchangeLinuxDoCode(s.cfg.LinuxDoClientID, s.cfg.LinuxDoClientSecret, code, redirectURI)
	if err != nil {
		log.Printf("[linuxdo] token exchange error: %v", err)
		http.Redirect(w, r, s.cfg.BaseURL+"/?error=token_exchange", http.StatusFound)
		return
	}

	// Get user info
	user, err := getLinuxDoUser(token)
	if err != nil {
		log.Printf("[linuxdo] get user error: %v", err)
		http.Redirect(w, r, s.cfg.BaseURL+"/?error=user_fetch", http.StatusFound)
		return
	}

	log.Printf("[linuxdo] login: id=%d username=%s trust_level=%d", user.ID, user.Username, user.TrustLevel)

	// Persist to database
	if db.DB != nil {
		dbUser, err := db.UpsertLinuxDoUser(user.ID, user.Username, user.Name, user.AvatarTemplate, user.TrustLevel)
		if err != nil {
			log.Printf("[linuxdo] db upsert error: %v", err)
		} else {
			s.setSession(w, dbUser.ID)
			log.Printf("[linuxdo] user saved: db_id=%d", dbUser.ID)
		}
	}

	http.Redirect(w, r, s.cfg.BaseURL+"/?login=linuxdo", http.StatusFound)
}

// GET /api/linuxdo/status
func (s *Server) handleLinuxDoStatus(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeJSON(w, map[string]any{"logged_in": false})
		return
	}
	writeJSON(w, map[string]any{
		"logged_in":        true,
		"id":               user.ID,
		"linuxdo_username": user.LinuxDoUsername,
		"linuxdo_name":     user.LinuxDoName,
		"trust_level":      user.TrustLevel,
		"avatar_url":       user.AvatarURL,
		"github_login":     user.GitHubLogin,
		"selected_repo":    user.SelectedRepo,
	})
}

// ── LinuxDo API helpers ─────────────────────────

func exchangeLinuxDoCode(clientID, clientSecret, code, redirectURI string) (string, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	}
	req, _ := http.NewRequest("POST", linuxdoToken, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		return "", fmt.Errorf("linuxdo oauth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

type LinuxDoUser struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	Name            string `json:"name"`
	AvatarTemplate  string `json:"avatar_template"`
	Active          bool   `json:"active"`
	TrustLevel      int    `json:"trust_level"`
	Silenced        bool   `json:"silenced"`
}

func getLinuxDoUser(token string) (*LinuxDoUser, error) {
	req, _ := http.NewRequest("GET", linuxdoUserAPI, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user LinuxDoUser
	json.NewDecoder(resp.Body).Decode(&user)
	return &user, nil
}
