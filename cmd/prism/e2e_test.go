package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/senran-N/prism/internal/github"
	"github.com/senran-N/prism/internal/mail"
	"github.com/senran-N/prism/internal/scproto"
)

// Run with:
//   YYDS_API_KEY=AC-... GITHUB_USER=... GITHUB_PASS=... GITHUB_TOTP=... REPO_ID=... \
//     go test ./cmd/prism/ -run TestE2E -v -count=1 -timeout 300s

func skipIfNoEnv(t *testing.T) {
	for _, key := range []string{"YYDS_API_KEY", "GITHUB_USER", "GITHUB_PASS", "GITHUB_TOTP", "REPO_ID"} {
		if os.Getenv(key) == "" {
			t.Skipf("skipping e2e: %s not set", key)
		}
	}
}

func TestE2EFullPipeline(t *testing.T) {
	skipIfNoEnv(t)

	apiKey := os.Getenv("YYDS_API_KEY")
	ghUser := os.Getenv("GITHUB_USER")
	ghPass := os.Getenv("GITHUB_PASS")
	ghTOTP := os.Getenv("GITHUB_TOTP")
	repoID := os.Getenv("REPO_ID")

	// 1. Create temp email
	t.Log("Creating temp email...")
	prefix := fmt.Sprintf("prism-e2e-%d", time.Now().Unix())
	addr, _, err := mail.CreateTempEmail(apiKey, prefix)
	if err != nil {
		t.Fatalf("create email: %v", err)
	}
	t.Logf("Email: %s", addr)

	// 2. Login GitHub
	t.Log("Logging into GitHub...")
	ghClient, err := github.Login(ghUser, ghPass, ghTOTP)
	if err != nil {
		t.Fatalf("github login: %v", err)
	}
	t.Log("GitHub login OK")

	// 3. Register SC
	t.Log("Registering SC account...")
	sc := scproto.NewClient()
	t.Logf("Fingerprint: %s", sc.Email) // just to verify it's not empty struct
	err = sc.Register(addr, "PrismE2E#2026xTest!", "E2E Tester")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Logf("Registered: workspace=%s", sc.WorkspaceID)

	// 4. Connect GitHub
	t.Log("Connecting GitHub OAuth...")
	err = sc.ConnectGitHub(ghClient)
	if err != nil {
		t.Fatalf("connect github: %v", err)
	}
	t.Log("GitHub connected")

	// 5. Create project
	t.Log("Creating project...")
	projectID, err := sc.CreateProject(repoID)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	t.Logf("Project: %s", projectID)

	// 6. Create ticket
	t.Log("Creating ticket...")
	ticketID, err := sc.CreateTicket(projectID, "E2E test: say hello", "codex_gpt_5_5_medium")
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	t.Logf("Ticket: https://www.superconductor.com/tickets/%s", ticketID)

	// 7. Check status
	t.Log("Checking status...")
	status, err := sc.GetTicketStatus(ticketID)
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	t.Logf("Status: %s, Cost: %s", status.Status, status.Cost)

	// 8. Disconnect GitHub (cleanup)
	t.Log("Disconnecting GitHub...")
	err = sc.DisconnectGitHub()
	if err != nil {
		t.Logf("disconnect warning: %v", err)
	}
	t.Log("Done!")
}
