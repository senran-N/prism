package config

import "os"

type Config struct {
	Addr           string
	YYDSAPIKey     string
	GitHubUser     string // service account for SC binding
	GitHubPass     string
	GitHubTOTP     string
	RepoID         string

	// GitHub OAuth App (for user-facing auth)
	GitHubClientID     string
	GitHubClientSecret string
	BaseURL            string // e.g. http://localhost:3001
}

func Load() Config {
	return Config{
		Addr:               envOr("PRISM_ADDR", ":8080"),
		YYDSAPIKey:         envOr("YYDS_API_KEY", ""),
		GitHubUser:         envOr("GITHUB_USER", ""),
		GitHubPass:         envOr("GITHUB_PASS", ""),
		GitHubTOTP:         envOr("GITHUB_TOTP", ""),
		RepoID:             envOr("REPO_ID", ""),
		GitHubClientID:     envOr("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: envOr("GITHUB_CLIENT_SECRET", ""),
		BaseURL:            envOr("BASE_URL", "http://localhost:3001"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
