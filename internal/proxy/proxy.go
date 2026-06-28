package proxy

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/senran-N/prism/internal/account"
)

const scBase = "https://www.superconductor.com"

type Handler struct {
	pool *account.Pool
}

func New(pool *account.Pool) *Handler {
	return &Handler{pool: pool}
}

// ServeHTTP reverse-proxies SC ticket pages using the account that created
// the ticket, eliminating the need for browser-side SC login.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetPath := strings.TrimPrefix(r.URL.Path, "/proxy")
	targetPath = path.Clean(targetPath)
	if strings.Contains(targetPath, "..") || !strings.HasPrefix(targetPath, "/tickets/") {
		http.Error(w, "forbidden", 403)
		return
	}

	parts := strings.SplitN(strings.TrimPrefix(targetPath, "/tickets/"), "/", 2)
	ticketID := parts[0]

	acct := h.pool.GetTicketAccount(ticketID)
	if acct == nil || acct.Client == nil {
		http.Error(w, "ticket session not found — please create a new task", 404)
		return
	}

	targetURL := scBase + targetPath
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		http.Error(w, "proxy error", 500)
		return
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")

	resp, err := acct.Client.HTTPClient().Do(req)
	if err != nil {
		log.Printf("[proxy] upstream error for ticket %s: %v", ticketID, err)
		http.Error(w, "upstream error", 502)
		return
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	if !strings.Contains(finalURL, "/tickets/") && !strings.Contains(finalURL, "/implementations/") {
		log.Printf("[proxy] ticket %s: SC redirected to %s (possible session expiry)", ticketID, finalURL)
		http.Error(w, "SC session expired — please create a new task", 502)
		return
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		if gz, err := gzip.NewReader(resp.Body); err == nil {
			reader = gz
			defer gz.Close()
		}
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, "read error", 502)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		html := string(body)
		html = strings.Replace(html, "<head>", `<head><base href="`+scBase+`/">`, 1)
		html = strings.ReplaceAll(html, `href="`+scBase+`/tickets/`, `href="/proxy/tickets/`)
		html = strings.ReplaceAll(html, `href="/tickets/`, `href="/proxy/tickets/`)
		body = []byte(html)
	}

	for k, vv := range resp.Header {
		switch strings.ToLower(k) {
		case "set-cookie", "content-length", "content-encoding",
			"content-security-policy", "x-frame-options",
			"strict-transport-security", "transfer-encoding":
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
