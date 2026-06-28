package proxy

import (
	"io"
	"net/http"
	"path"
	"regexp"
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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetPath := strings.TrimPrefix(r.URL.Path, "/proxy")
	targetPath = path.Clean(targetPath)
	if strings.Contains(targetPath, "..") {
		http.Error(w, "forbidden", 403)
		return
	}
	if !strings.HasPrefix(targetPath, "/tickets/") {
		http.Error(w, "forbidden", 403)
		return
	}

	// Find an account that has a valid SC client
	var client *http.Client
	for _, acct := range h.pool.ListAll() {
		if acct.Client != nil && acct.Client.HTTPClient() != nil {
			client = acct.Client.HTTPClient()
			break
		}
	}
	if client == nil {
		http.Error(w, "no available account", 503)
		return
	}

	// If path is just /tickets/{id}, find the implementation first
	targetURL := scBase + targetPath
	if matched, _ := regexp.MatchString(`^/tickets/[A-Za-z0-9]+$`, targetPath); matched {
		// Fetch ticket page to find implementation ID
		implReq, _ := http.NewRequest("GET", targetURL, nil)
		implReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
		implResp, err := client.Do(implReq)
		if err == nil {
			body, _ := io.ReadAll(implResp.Body)
			implResp.Body.Close()
			implRe := regexp.MustCompile(targetPath + `/implementations/([A-Za-z0-9]+)`)
			if m := implRe.FindStringSubmatch(string(body)); m != nil {
				targetURL = scBase + targetPath + "/implementations/" + m[1]
			}
		}
	}

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

	for k, vv := range resp.Header {
		lower := strings.ToLower(k)
		if lower == "set-cookie" || lower == "x-frame-options" || lower == "content-security-policy" {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
