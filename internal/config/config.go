package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
	BaseURL     string

	// Session
	SessionSecret string

	// LinuxDo OAuth
	LinuxDoClientID     string
	LinuxDoClientSecret string

	// GitHub OAuth App (user-facing)
	GitHubClientID     string
	GitHubClientSecret string

	// GitHub Service Account (SC binding)
	GitHubUser string
	GitHubPass string
	GitHubTOTP string

	// YYDS Mail
	YYDSAPIKey string

	// Default repo (optional, overridden by user selection)
	RepoID string
}

func Load() Config {
	return Config{
		Addr:               envOr("PRISM_ADDR", ":8080"),
		DatabaseURL:        envOr("DATABASE_URL", "postgres://prism:prism@localhost:5432/prism?sslmode=disable"),
		BaseURL:            envOr("BASE_URL", "http://localhost:3001"),
		SessionSecret:      envOr("SESSION_SECRET", "prism-dev-secret-change-me"),
		LinuxDoClientID:     envOr("LINUXDO_CLIENT_ID", ""),
		LinuxDoClientSecret: envOr("LINUXDO_CLIENT_SECRET", ""),
		GitHubClientID:     envOr("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: envOr("GITHUB_CLIENT_SECRET", ""),
		GitHubUser:         envOr("GITHUB_USER", ""),
		GitHubPass:         envOr("GITHUB_PASS", ""),
		GitHubTOTP:         envOr("GITHUB_TOTP", ""),
		YYDSAPIKey:         envOr("YYDS_API_KEY", ""),
		RepoID:             envOr("REPO_ID", ""),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
