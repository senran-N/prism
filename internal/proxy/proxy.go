// Package proxy reverse-proxies Superconductor implementation pages,
// injecting the correct session cookies so users see agent work directly.
package proxy

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

const scBase = "https://www.superconductor.com"

// Handler serves as a reverse proxy for SC implementation pages.
// It injects cookies from the provided jar so the user sees the
// authenticated agent view without knowing about SC accounts.
type Handler struct {
	jar *cookiejar.Jar
}

func New(jar *cookiejar.Jar) *Handler {
	return &Handler{jar: jar}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Path: /proxy/tickets/{tid}/implementations/{iid}
	targetPath := strings.TrimPrefix(r.URL.Path, "/proxy")
	targetURL := scBase + targetPath

	client := &http.Client{Jar: h.jar}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "upstream error", 502)
		return
	}
	defer resp.Body.Close()

	// Rewrite SC URLs in the response to go through our proxy
	for k, vv := range resp.Header {
		if strings.EqualFold(k, "Set-Cookie") {
			continue // strip SC cookies from client
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
