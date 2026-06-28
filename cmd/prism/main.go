package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/senran-N/prism/internal/account"
	"github.com/senran-N/prism/internal/api"
	"github.com/senran-N/prism/internal/config"
	"github.com/senran-N/prism/internal/db"
	"github.com/senran-N/prism/internal/proxy"
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

	// Configure and start pool warmer
	if cfg.PoolMinReady > 0 {
		scheduler.MinReadyAccounts = cfg.PoolMinReady
	}
	if cfg.PoolMaxWarm > 0 {
		scheduler.MaxConcurrentWarm = cfg.PoolMaxWarm
	}
	sched.StartPoolWarmer()

	apiServer := api.New(sched, pool, cfg)

	proxyHandler := proxy.New(pool)

	mux := http.NewServeMux()
	mux.Handle("/api/", apiServer)
	mux.Handle("/proxy/", proxyHandler)

	handler := api.AntiBotMiddleware(
		api.SecurityHeaders(
			api.CORSMiddleware(cfg.BaseURL)(mux),
		),
	)

	server := &http.Server{
		Addr:           cfg.Addr,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   120 * time.Second, // long for SSE + SC operations
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}
	log.Printf("Prism starting on %s", cfg.Addr)
	log.Fatal(server.ListenAndServe())
}
