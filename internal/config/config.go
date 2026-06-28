package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr        string
	DatabaseURL string
	BaseURL     string

	// Admin credentials
	AdminUser string
	AdminPass string

	// Session
	SessionSecret string

	// LinuxDo OAuth
	LinuxDoClientID     string
	LinuxDoClientSecret string
	LinuxDoRedirectURI  string

	// Rotation pricing (LDC credits per rotation, admin-configurable)
	RotationCost float64

	// GitHub OAuth App (user-facing)
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string

	// GitHub Service Account (SC binding)
	GitHubUser string
	GitHubPass string
	GitHubTOTP string

	// LinuxDo Credit
	CreditClientID     string
	CreditClientSecret string

	// YYDS Mail
	YYDSAPIKey string

	// Pool sizing
	PoolMinReady   int // minimum ready accounts (default 5)
	PoolMaxWarm    int // max concurrent registrations (default 3)

	// Default repo
	RepoID string
}

func Load() Config {
	return Config{
		Addr:               envOr("PRISM_ADDR", ":8080"),
		DatabaseURL:        envOr("DATABASE_URL", "postgres://prism:prism@localhost:5432/prism?sslmode=disable"),
		BaseURL:            envOr("BASE_URL", "http://localhost:3001"),
		AdminUser:          envOr("ADMIN_USER", "admin"),
		AdminPass:          envOr("ADMIN_PASS", "admin"),
		SessionSecret:      envOr("SESSION_SECRET", "prism-dev-secret-change-me"),
		LinuxDoClientID:     envOr("LINUXDO_CLIENT_ID", ""),
		LinuxDoClientSecret: envOr("LINUXDO_CLIENT_SECRET", ""),
		LinuxDoRedirectURI:  envOr("LINUXDO_REDIRECT_URI", ""),
		RotationCost:        envOrFloat("ROTATION_COST", 20.0),
		GitHubClientID:     envOr("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: envOr("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:  envOr("GITHUB_REDIRECT_URI", ""),
		CreditClientID:     envOr("CREDIT_CLIENT_ID", ""),
		CreditClientSecret: envOr("CREDIT_CLIENT_SECRET", ""),
		GitHubUser:         envOr("GITHUB_USER", ""),
		GitHubPass:         envOr("GITHUB_PASS", ""),
		GitHubTOTP:         envOr("GITHUB_TOTP", ""),
		YYDSAPIKey:         envOr("YYDS_API_KEY", ""),
		PoolMinReady:       int(envOrFloat("POOL_MIN_READY", 5)),
		PoolMaxWarm:        int(envOrFloat("POOL_MAX_WARM", 3)),
		RepoID:             envOr("REPO_ID", ""),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
