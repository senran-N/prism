package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/prism/internal/db"
)

// ── Auth middleware ─────────────────────────────

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := s.getSessionUser(r)
		if user == nil {
			writeError(w, 401, "authentication required")
			return
		}
		if user.IsBanned {
			writeError(w, 403, "account suspended")
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.isAdminSession(r) {
			writeError(w, 401, "admin authentication required")
			return
		}
		next(w, r)
	}
}

// ── Rate limiter ────────────────────────────────

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old entries
	reqs := rl.requests[key]
	clean := reqs[:0]
	for _, t := range reqs {
		if t.After(cutoff) {
			clean = append(clean, t)
		}
	}

	if len(clean) >= rl.limit {
		rl.requests[key] = clean
		return false
	}

	rl.requests[key] = append(clean, now)
	return true
}

func (s *Server) rateLimit(limit int, window time.Duration, next http.HandlerFunc) http.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.Allow(ip) {
			writeError(w, 429, "too many requests")
			return
		}
		next(w, r)
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}

// ── CSRF token ──────────────────────────────────

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ── OAuth state ─────────────────────────────────

var (
	oauthStates   = make(map[string]time.Time)
	oauthStatesMu sync.Mutex
)

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	state := hex.EncodeToString(b)

	oauthStatesMu.Lock()
	oauthStates[state] = time.Now()
	for k, t := range oauthStates {
		if time.Since(t) > 30*time.Minute {
			delete(oauthStates, k)
		}
	}
	oauthStatesMu.Unlock()
	return state
}

func validateOAuthState(state string) bool {
	oauthStatesMu.Lock()
	defer oauthStatesMu.Unlock()
	t, ok := oauthStates[state]
	if !ok {
		return false
	}
	delete(oauthStates, state)
	return time.Since(t) < 30*time.Minute
}

// State with user ID encoded: "randomhex.userID"
var oauthUserStates = make(map[string]int64) // state → userID

func generateOAuthStateWithUser(userID int64) string {
	state := generateOAuthState()
	oauthStatesMu.Lock()
	oauthUserStates[state] = userID
	oauthStatesMu.Unlock()
	return state
}

// Returns userID (0 if not found or expired)
func validateOAuthStateWithUser(state string) int64 {
	oauthStatesMu.Lock()
	defer oauthStatesMu.Unlock()

	userID := oauthUserStates[state]
	delete(oauthUserStates, state)

	// Also validate timestamp
	t, ok := oauthStates[state]
	if ok {
		delete(oauthStates, state)
		if time.Since(t) > 30*time.Minute {
			return 0
		}
	}
	return userID
}

// ── Security headers ────────────────────────────

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Allow iframe for /proxy/ paths (SC implementation pages)
		if !strings.HasPrefix(r.URL.Path, "/proxy/") {
			w.Header().Set("X-Frame-Options", "DENY")
		}
		next.ServeHTTP(w, r)
	})
}

// ── CORS (restricted) ───────────────────────────

func CORSMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == allowedOrigin || allowedOrigin == "*" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ── Request body limit ──────────────────────────

func limitBody(maxBytes int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next(w, r)
	}
}

// ── Secure session cookie helper ────────────────

func isHTTPS(baseURL string) bool {
	return strings.HasPrefix(baseURL, "https://")
}

// Per-user GitHub token (avoid global state race) ─

func getUserGitHubToken(r *http.Request, s *Server) string {
	user := s.getSessionUser(r)
	if user != nil && user.GitHubToken != "" {
		return user.GitHubToken
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ghToken
}

// Ensure db is imported
var _ = db.DB
