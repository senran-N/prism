package main

import (
	"log"
	"net/http"
	"os"

	"github.com/senran-N/prism/internal/account"
	"github.com/senran-N/prism/internal/api"
	"github.com/senran-N/prism/internal/config"
	"github.com/senran-N/prism/internal/db"
	"github.com/senran-N/prism/internal/scheduler"
)

func main() {
	cfg := config.Load()

	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Printf("[warn] database not available: %v (running without persistence)", err)
	} else {
		migrations := []string{
			"migrations/001_init.sql",
			"migrations/002_user_ban.sql",
			"migrations/003_credits_and_codes.sql",
		}
		for _, f := range migrations {
			schema, err := os.ReadFile(f)
			if err != nil {
				log.Printf("[warn] migration %s not found: %v", f, err)
				continue
			}
			if err := db.Migrate(string(schema)); err != nil {
				log.Printf("[warn] migration %s error: %v", f, err)
			}
		}
	}

	// Set rotation cost from config
	db.RotationCost = cfg.RotationCost

	pool := account.NewPool()
	sched := scheduler.New(pool, scheduler.Config{
		YYDSAPIKey: cfg.YYDSAPIKey,
		GitHubUser: cfg.GitHubUser,
		GitHubPass: cfg.GitHubPass,
		GitHubTOTP: cfg.GitHubTOTP,
		RepoID:     cfg.RepoID,
	})

	apiServer := api.New(sched, pool, cfg)

	mux := http.NewServeMux()
	mux.Handle("/api/", apiServer)

	handler := api.AntiBotMiddleware(
		api.SecurityHeaders(
			api.CORSMiddleware(cfg.BaseURL)(mux),
		),
	)

	log.Printf("Prism starting on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, handler))
}
