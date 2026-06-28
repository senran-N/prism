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
		schema, err := os.ReadFile("migrations/001_init.sql")
		if err != nil {
			log.Printf("[warn] migration file not found: %v", err)
		} else if err := db.Migrate(string(schema)); err != nil {
			log.Printf("[warn] migration error: %v", err)
		}
	}

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
