// Package antibot implements mechanisms to confuse AI-powered scanners
// and automated analysis tools while remaining transparent to real users.
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ── Honeypot endpoints ──────────────────────────
// Fake juicy-looking endpoints that real users never hit.
// Any access = automated scanner → log + ban IP.

var honeypotPaths = []string{
	"/api/v1/auth/token",
	"/api/v2/users/admin",
	"/api/internal/debug",
	"/api/graphql",
	"/.env",
	"/api/config.json",
	"/api/keys",
	"/api/v1/admin/credentials",
	"/api/swagger.json",
	"/api/v1/oauth/token",
	"/wp-admin/",
	"/actuator/env",
	"/debug/pprof/",
	"/api/.git/config",
	"/api/v1/secret",
}

type honeypotTracker struct {
	mu      sync.Mutex
	flagged map[string]time.Time // IP → first trigger time
}

var honeypots = &honeypotTracker{
	flagged: make(map[string]time.Time),
}

func (h *honeypotTracker) trigger(ip, path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.flagged[ip]; !ok {
		log.Printf("[HONEYPOT] triggered: ip=%s path=%s", ip, path)
	}
	h.flagged[ip] = time.Now()
}

func (h *honeypotTracker) isFlagged(ip string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	t, ok := h.flagged[ip]
	if !ok {
		return false
	}
	// Flag expires after 1 hour
	if time.Since(t) > time.Hour {
		delete(h.flagged, ip)
		return false
	}
	return true
}

func registerHoneypots(mux *http.ServeMux) {
	for _, p := range honeypotPaths {
		path := p
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			honeypots.trigger(ip, path)
			// Return convincing but fake response to waste scanner time
			time.Sleep(time.Duration(500+rand.Intn(2000)) * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(generateFakeResponse(path))
		})
	}
}

func generateFakeResponse(path string) map[string]any {
	// Return different fake data based on path to look realistic
	switch {
	case strings.Contains(path, "token"):
		return map[string]any{
			"access_token": generateFakeToken(),
			"token_type":   "bearer",
			"expires_in":   3600,
			"scope":        "read write",
		}
	case strings.Contains(path, "config") || strings.Contains(path, "env"):
		return map[string]any{
			"database_url": "postgres://readonly:readonly@internal-db:5432/app",
			"redis_url":    "redis://cache:6379/0",
			"api_key":      generateFakeToken(),
			"debug":        false,
		}
	case strings.Contains(path, "admin") || strings.Contains(path, "user"):
		return map[string]any{
			"users": []map[string]any{
				{"id": 1, "role": "admin", "email": "admin@internal.local"},
				{"id": 2, "role": "user", "email": "user@internal.local"},
			},
		}
	case strings.Contains(path, "secret") || strings.Contains(path, "key") || strings.Contains(path, "credential"):
		return map[string]any{
			"aws_access_key":     "AKIA" + generateFakeToken()[:16],
			"aws_secret_key":     generateFakeToken(),
			"stripe_secret_key":  "sk_live_" + generateFakeToken()[:24],
			"github_token":       "ghp_" + generateFakeToken()[:36],
		}
	default:
		return map[string]any{"status": "ok", "version": "2.1.0"}
	}
}

func generateFakeToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ── Behavioral analysis ─────────────────────────
// Track request patterns to detect AI scanners:
// - Too-regular timing intervals
// - Systematic path enumeration
// - Missing normal browser signals

type behaviorTracker struct {
	mu       sync.Mutex
	sessions map[string]*clientBehavior
}

type clientBehavior struct {
	requests    []requestLog
	score       float64 // higher = more suspicious
	lastChecked time.Time
}

type requestLog struct {
	path      string
	timestamp time.Time
	hasReferer bool
	hasCookie  bool
}

var behavior = &behaviorTracker{
	sessions: make(map[string]*clientBehavior),
}

func (bt *behaviorTracker) record(ip string, r *http.Request) float64 {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	cb, ok := bt.sessions[ip]
	if !ok {
		cb = &clientBehavior{}
		bt.sessions[ip] = cb
	}

	cb.requests = append(cb.requests, requestLog{
		path:       r.URL.Path,
		timestamp:  time.Now(),
		hasReferer: r.Referer() != "",
		hasCookie:  r.Header.Get("Cookie") != "",
	})

	// Keep last 100 requests
	if len(cb.requests) > 100 {
		cb.requests = cb.requests[len(cb.requests)-100:]
	}

	// Analyze every 10 requests
	if len(cb.requests)%10 == 0 {
		cb.score = analyzeRequests(cb.requests)
		cb.lastChecked = time.Now()
	}

	// Clean old sessions
	if len(bt.sessions) > 10000 {
		for k, v := range bt.sessions {
			if time.Since(v.lastChecked) > 30*time.Minute {
				delete(bt.sessions, k)
			}
		}
	}

	return cb.score
}

func analyzeRequests(reqs []requestLog) float64 {
	if len(reqs) < 5 {
		return 0
	}

	score := 0.0

	// Check 1: Too-regular timing (AI agents tend to have consistent intervals)
	var intervals []time.Duration
	for i := 1; i < len(reqs); i++ {
		intervals = append(intervals, reqs[i].timestamp.Sub(reqs[i-1].timestamp))
	}
	if len(intervals) > 3 {
		avg := time.Duration(0)
		for _, d := range intervals {
			avg += d
		}
		avg /= time.Duration(len(intervals))

		variance := 0.0
		for _, d := range intervals {
			diff := float64(d-avg) / float64(time.Millisecond)
			variance += diff * diff
		}
		variance /= float64(len(intervals))

		// Very low variance = bot-like regularity
		if variance < 1000 && avg < 2*time.Second {
			score += 30
		}
	}

	// Check 2: No referer on non-entry pages (bots skip navigation)
	noReferer := 0
	for _, r := range reqs {
		if !r.hasReferer && !isEntryPath(r.path) {
			noReferer++
		}
	}
	if float64(noReferer)/float64(len(reqs)) > 0.8 {
		score += 20
	}

	// Check 3: Systematic path enumeration
	uniquePaths := map[string]bool{}
	for _, r := range reqs {
		uniquePaths[r.path] = true
	}
	if float64(len(uniquePaths))/float64(len(reqs)) > 0.7 {
		score += 25
	}

	// Check 4: High request rate
	if len(reqs) > 2 {
		duration := reqs[len(reqs)-1].timestamp.Sub(reqs[0].timestamp)
		if duration > 0 {
			rps := float64(len(reqs)) / duration.Seconds()
			if rps > 5 {
				score += 25
			}
		}
	}

	return score
}

func isEntryPath(p string) bool {
	return p == "/" || p == "/admin" || p == "/terms" || p == "/privacy" ||
		strings.HasPrefix(p, "/api/linuxdo/") || strings.HasPrefix(p, "/api/github/")
}

// ── Response noise injection ────────────────────
// Add random fields to API error responses to confuse automated parsers.

func noisyError(w http.ResponseWriter, code int, msg string) {
	resp := map[string]any{
		"error":   msg,
		"code":    code,
		"request": generateRequestID(),
	}

	// Add random noise fields that change per request
	noiseFields := []string{"trace_id", "region", "node", "shard", "ref", "ctx"}
	n := 1 + rand.Intn(3)
	for i := 0; i < n; i++ {
		field := noiseFields[rand.Intn(len(noiseFields))]
		resp[field] = generateRequestID()[:8]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func generateRequestID() string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())))
	return hex.EncodeToString(h[:8])
}

// ── Anti-analysis middleware ────────────────────

func AntiBotMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		// Block honeypot-flagged IPs
		if honeypots.isFlagged(ip) {
			// Random delay to slow down scanners
			time.Sleep(time.Duration(1000+rand.Intn(3000)) * time.Millisecond)
			noisyError(w, 403, "access denied")
			return
		}

		// Record behavior and check suspicion score
		score := behavior.record(ip, r)
		if score >= 70 {
			log.Printf("[ANTIBOT] suspicious activity: ip=%s score=%.0f path=%s", ip, score, r.URL.Path)
			// Don't block immediately — add delays to slow down
			time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
		}

		// Add jitter to response timing (makes timing analysis harder)
		if rand.Intn(10) < 3 {
			time.Sleep(time.Duration(10+rand.Intn(50)) * time.Millisecond)
		}

		next.ServeHTTP(w, r)
	})
}
