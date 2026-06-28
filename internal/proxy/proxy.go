package proxy

import (
	"fmt"
	"net/http"
	"net/url"
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

// ServeHTTP generates a self-login page that logs into SC and redirects
// to the ticket/implementation page. User's browser gets SC session cookies
// directly so they can interact with the full SC interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetPath := strings.TrimPrefix(r.URL.Path, "/proxy")
	targetPath = path.Clean(targetPath)
	if strings.Contains(targetPath, "..") || !strings.HasPrefix(targetPath, "/tickets/") {
		http.Error(w, "forbidden", 403)
		return
	}

	parts := strings.SplitN(strings.TrimPrefix(targetPath, "/tickets/"), "/", 2)
	ticketID := parts[0]

	// Find the SC account that created this ticket
	acct := h.pool.GetTicketAccount(ticketID)
	if acct == nil {
		for _, a := range h.pool.ListAll() {
			if a.Email != "" && a.Password != "" {
				acct = a
				break
			}
		}
	}
	if acct == nil {
		http.Error(w, "no account available", 503)
		return
	}

	// Render an auto-login page: logs into SC, then redirects to ticket
	targetURL := scBase + "/tickets/" + ticketID
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8">
<title>Opening workspace...</title>
<style>body{font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#f6f9fc;color:#697386;flex-direction:column;gap:12px;}
.spinner{width:24px;height:24px;border:3px solid #e3e8ee;border-top-color:#635bff;border-radius:50%%;animation:spin 0.8s linear infinite;}
@keyframes spin{to{transform:rotate(360deg)}}
</style></head><body>
<div class="spinner"></div>
<p>正在打开工作台 / Opening workspace...</p>
<form id="f" method="POST" action="%s/log_in">
<input type="hidden" name="email" value="%s">
<input type="hidden" name="password" value="%s">
<input type="hidden" name="commit" value="Log In">
</form>
<script>
// Submit login form, then redirect to ticket
var f = document.getElementById("f");
var xhr = new XMLHttpRequest();
xhr.open("POST", "%s/log_in", true);
xhr.withCredentials = true;
xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
xhr.onload = function() { window.location.href = "%s"; };
xhr.onerror = function() { window.location.href = "%s"; };
var data = "email=%s&password=%s&commit=Log+In";
xhr.send(data);
</script></body></html>`,
		scBase, acct.Email, acct.Password,
		scBase, targetURL, targetURL,
		url.QueryEscape(acct.Email), url.QueryEscape(acct.Password))
}
