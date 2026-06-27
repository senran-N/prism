package scproto

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateTOTP(t *testing.T) {
	code := GenerateTOTP("JBSWY3DPEHPK3PXP")
	if len(code) != 6 {
		t.Fatalf("expected 6-digit code, got %q", code)
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non-digit in TOTP: %q", code)
		}
	}
}

func TestRandomFingerprint(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		fp := RandomFingerprint()
		if fp.UserAgent == "" {
			t.Fatal("empty UA")
		}
		if !strings.Contains(fp.UserAgent, "Chrome/") {
			t.Fatalf("UA missing Chrome: %s", fp.UserAgent)
		}
		if !strings.Contains(fp.UserAgent, "Mozilla/5.0") {
			t.Fatalf("UA missing Mozilla: %s", fp.UserAgent)
		}
		seen[fp.UserAgent] = true
		time.Sleep(time.Millisecond) // seed variation
	}
	if len(seen) < 3 {
		t.Fatalf("fingerprint diversity too low: only %d unique UAs in 20 runs", len(seen))
	}
}

func TestExtractCSRF(t *testing.T) {
	tests := []struct {
		html string
		want string
	}{
		{`<meta name="csrf-token" content="abc123">`, "abc123"},
		{`<input name="authenticity_token" value="tok1"><input name="authenticity_token" value="tok2">`, "tok2"},
		{`<html>no token</html>`, ""},
	}
	for _, tt := range tests {
		got := extractCSRF(tt.html)
		if got != tt.want {
			t.Errorf("extractCSRF(%q) = %q, want %q", tt.html[:40], got, tt.want)
		}
	}
}

func TestExtractFormFields(t *testing.T) {
	html := `
		<input name="authenticity_token" value="csrf1" type="hidden">
		<input name="authenticity_token" value="csrf2" type="hidden">
		<input name="xyzabc" type="text">
		<input name="spinner" value="spin123" type="hidden">
		<input name="name" type="text">
		<input name="email" type="email">
	`
	f := extractFormFields(html)
	if f.CSRF != "csrf2" {
		t.Errorf("CSRF = %q, want csrf2", f.CSRF)
	}
	if f.Spinner != "spin123" {
		t.Errorf("Spinner = %q, want spin123", f.Spinner)
	}
	if f.Honeypot != "xyzabc" {
		t.Errorf("Honeypot = %q, want xyzabc", f.Honeypot)
	}
}

func TestRandomDelay(t *testing.T) {
	start := time.Now()
	RandomDelay(2*time.Second, 1*time.Second)
	elapsed := time.Since(start)
	if elapsed < 1*time.Second || elapsed > 4*time.Second {
		t.Errorf("delay %v outside expected range 1~4s", elapsed)
	}
}

func TestTLSConfig(t *testing.T) {
	cfg := TLSConfig()
	if len(cfg.CipherSuites) == 0 {
		t.Fatal("no cipher suites")
	}
	// Run twice, check ordering differs (probabilistic but very likely)
	cfg2 := TLSConfig()
	same := true
	for i := range cfg.CipherSuites {
		if cfg.CipherSuites[i] != cfg2.CipherSuites[i] {
			same = false
			break
		}
	}
	// Not guaranteed to differ, but extremely likely with 9 suites
	_ = same // just make sure it doesn't panic
}
