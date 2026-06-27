// Package scproto implements the Superconductor protocol: registration,
// login, OAuth, project/ticket management — all via standard HTTP.
package scproto

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	scBase      = "https://www.superconductor.com"
	userAgent   = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36"
	antiBotWait = 3 * time.Second
)

// Client holds an authenticated Superconductor session.
type Client struct {
	http        *http.Client
	jar         *cookiejar.Jar
	fp          Fingerprint
	Email       string
	Password    string
	WorkspaceID string
	UserID      string
}

// NewClient creates an unauthenticated client with a random fingerprint.
func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	fp := RandomFingerprint()
	transport := &http.Transport{
		TLSClientConfig: TLSConfig(),
	}
	return &Client{
		http: &http.Client{
			Jar:       jar,
			Transport: transport,
		},
		jar: jar,
		fp:  fp,
	}
}

// newNoRedirectClient returns a client that does NOT follow redirects.
func (c *Client) newNoRedirectClient() *http.Client {
	return &http.Client{
		Jar:       c.jar,
		Transport: c.http.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// ── HTTP helpers ────────────────────────────────────────────

func (c *Client) get(rawURL string) (string, error) {
	req, _ := http.NewRequest("GET", rawURL, nil)
	c.fp.ApplyHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return readBody(resp)
}

func (c *Client) post(rawURL string, data url.Values, extraHeaders map[string]string) (body string, finalURL string, status int, err error) {
	req, _ := http.NewRequest("POST", rawURL, strings.NewReader(data.Encode()))
	c.fp.ApplyHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	u, _ := url.Parse(rawURL)
	req.Header.Set("Origin", u.Scheme+"://"+u.Host)
	req.Header.Set("Referer", rawURL)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	body, err = readBody(resp)
	return body, resp.Request.URL.String(), resp.StatusCode, err
}

func readBody(resp *http.Response) (string, error) {
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gz.Close()
		reader = gz
	}
	b, err := io.ReadAll(reader)
	return string(b), err
}

// ── Form field extraction ───────────────────────────────────

var (
	reCSRFMeta  = regexp.MustCompile(`csrf-token"\s+content="([^"]+)"`)
	reCSRFInput = regexp.MustCompile(`authenticity_token.*?value="([^"]+)"`)
	reSpinner   = regexp.MustCompile(`name="spinner".*?value="([^"]+)"`)
	reHoneypot  = regexp.MustCompile(`<input[^>]*(?:type="text"[^>]*name="([^"]+)"|name="([^"]+)"[^>]*type="text")`)
	reWorkspace = regexp.MustCompile(`/workspaces/([A-Za-z0-9]+)`)
	reUserID    = regexp.MustCompile(`data-current-user-id="(\d+)"`)
	reProject   = regexp.MustCompile(`/projects/([A-Za-z0-9]+)`)
	reTicket    = regexp.MustCompile(`/tickets/([A-Za-z0-9]+)`)
	reIdentity  = regexp.MustCompile(`/identities/([A-Za-z0-9]+)`)
	reModels    = regexp.MustCompile(`implementations\[([^\]]+)\]`)
	reBranch    = regexp.MustCompile(`ticket\[base_branches\]\[([^\]]+)\]`)
)

func extractCSRF(html string) string {
	if m := reCSRFMeta.FindStringSubmatch(html); m != nil {
		return m[1]
	}
	all := reCSRFInput.FindAllStringSubmatch(html, -1)
	if len(all) > 0 {
		return all[len(all)-1][1]
	}
	return ""
}

type formFields struct {
	CSRF     string
	Spinner  string
	Honeypot string // field name, value should be ""
}

func extractFormFields(html string) formFields {
	f := formFields{CSRF: extractCSRF(html)}
	if m := reSpinner.FindStringSubmatch(html); m != nil {
		f.Spinner = m[1]
	}
	for _, m := range reHoneypot.FindAllStringSubmatch(html, -1) {
		name := m[1]
		if name == "" {
			name = m[2]
		}
		if name != "" && name != "name" && name != "email" {
			f.Honeypot = name
			break
		}
	}
	return f
}

func buildFormData(f formFields, extra [][2]string) url.Values {
	v := url.Values{}
	v.Set("authenticity_token", f.CSRF)
	if f.Honeypot != "" {
		v.Set(f.Honeypot, "")
	}
	if f.Spinner != "" {
		v.Set("spinner", f.Spinner)
	}
	for _, kv := range extra {
		v.Set(kv[0], kv[1])
	}
	return v
}

// ── Register ────────────────────────────────────────────────

func (c *Client) Register(email, password, name string) error {
	html, err := c.get(scBase + "/sign_up")
	if err != nil {
		return fmt.Errorf("get signup: %w", err)
	}
	fields := extractFormFields(html)
	if fields.CSRF == "" {
		return fmt.Errorf("no CSRF on signup page")
	}

	log.Printf("[scproto] register: csrf=%s... spinner=%s... honeypot=%s",
		fields.CSRF[:20], fields.Spinner[:min(20, len(fields.Spinner))], fields.Honeypot)

	RandomDelay(3*time.Second, 2*time.Second) // 1~5s random

	data := buildFormData(fields, [][2]string{
		{"name", name},
		{"email", email},
		{"password", password},
		{"commit", "Sign Up"},
	})

	log.Printf("[scproto] POST /sign_up data keys: %v", keysOf(data))

	body, finalURL, status, err := c.post(scBase+"/sign_up", data, nil)
	if err != nil {
		return fmt.Errorf("post signup: %w", err)
	}

	log.Printf("[scproto] register response: status=%d url=%s bodyLen=%d", status, finalURL, len(body))

	m := reWorkspace.FindStringSubmatch(finalURL)
	if m == nil {
		return fmt.Errorf("registration failed, landed on: %s", finalURL)
	}

	c.Email = email
	c.Password = password
	c.WorkspaceID = m[1]
	return nil
}

// ── Login ───────────────────────────────────────────────────

func (c *Client) Login(email, password string) error {
	html, err := c.get(scBase + "/log_in")
	if err != nil {
		return fmt.Errorf("get login: %w", err)
	}
	fields := extractFormFields(html)

	RandomDelay(3*time.Second, 2*time.Second) // 1~5s random

	data := buildFormData(fields, [][2]string{
		{"email", email},
		{"password", password},
		{"commit", "Log In"},
	})

	body, finalURL, _, err := c.post(scBase+"/log_in", data, nil)
	if err != nil {
		return fmt.Errorf("post login: %w", err)
	}
	if strings.Contains(finalURL, "/log_in") {
		return fmt.Errorf("login failed")
	}

	c.Email = email
	c.Password = password
	if m := reWorkspace.FindStringSubmatch(finalURL); m != nil {
		c.WorkspaceID = m[1]
	}
	if m := reUserID.FindStringSubmatch(body); m != nil {
		c.UserID = m[1]
	}
	return nil
}

// ── GitHub OAuth ────────────────────────────────────────────

// ConnectGitHub performs the 3-step OAuth dance using cookies from a
// logged-in GitHub http.Client.
func (c *Client) ConnectGitHub(ghClient *http.Client) error {
	html, err := c.get(scBase + "/workspaces/" + c.WorkspaceID)
	if err != nil {
		return err
	}
	csrf := extractCSRF(html)

	scNR := c.newNoRedirectClient()

	// Step A: POST /auth/github → 302 → github oauth
	data := url.Values{"authenticity_token": {csrf}}
	req, _ := http.NewRequest("POST", scBase+"/auth/github", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", scBase)
	resp, err := scNR.Do(req)
	if err != nil {
		return fmt.Errorf("oauth step A: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 302 {
		return fmt.Errorf("oauth step A: expected 302, got %d", resp.StatusCode)
	}
	oauthURL := resp.Header.Get("Location")

	// Step B: GitHub auto-authorize → 302 → callback
	ghNR := &http.Client{
		Jar:       ghClient.Jar,
		Transport: ghClient.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	ghReq, _ := http.NewRequest("GET", oauthURL, nil)
	ghReq.Header.Set("User-Agent", userAgent)
	resp2, err := ghNR.Do(ghReq)
	if err != nil {
		return fmt.Errorf("oauth step B: %w", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 302 {
		return fmt.Errorf("oauth step B: expected 302, got %d", resp2.StatusCode)
	}
	callbackURL := resp2.Header.Get("Location")
	if !strings.Contains(callbackURL, "superconductor.com") {
		return fmt.Errorf("oauth callback not SC: %s", callbackURL)
	}

	// Step C: callback → 302 → workspace
	resp3, err := scNR.Get(callbackURL)
	if err != nil {
		return fmt.Errorf("oauth step C: %w", err)
	}
	resp3.Body.Close()
	if resp3.StatusCode == 302 {
		loc := resp3.Header.Get("Location")
		c.get(loc)
	}
	return nil
}

// DisconnectGitHub unbinds the GitHub account from this SC user.
func (c *Client) DisconnectGitHub() error {
	html, err := c.get(scBase + "/identities")
	if err != nil {
		return err
	}
	csrf := extractCSRF(html)
	m := reIdentity.FindStringSubmatch(html)
	if m == nil {
		return fmt.Errorf("no connected GitHub identity found")
	}
	data := url.Values{
		"_method":            {"delete"},
		"authenticity_token": {csrf},
	}
	_, _, _, err = c.post(scBase+"/identities/"+m[1], data, nil)
	return err
}

// ── Project ─────────────────────────────────────────────────

func (c *Client) CreateProject(repoID string) (projectID string, err error) {
	html, err := c.get(scBase + "/workspaces/" + c.WorkspaceID)
	if err != nil {
		return "", err
	}
	csrf := extractCSRF(html)

	data := url.Values{
		"authenticity_token":       {csrf},
		"project[repository_ids][]": {repoID},
		"commit":                   {"Create project"},
	}
	_, finalURL, _, err := c.post(
		scBase+"/workspaces/"+c.WorkspaceID+"/projects",
		data,
		map[string]string{
			"X-CSRF-Token": csrf,
			"Accept":       "text/vnd.turbo-stream.html, text/html, application/xhtml+xml",
		},
	)
	if err != nil {
		return "", err
	}
	m := reProject.FindStringSubmatch(finalURL)
	if m == nil {
		return "", fmt.Errorf("project creation failed: %s", finalURL)
	}
	return m[1], nil
}

// ── Ticket ──────────────────────────────────────────────────

func (c *Client) CreateTicket(projectID, description, model string) (ticketID string, err error) {
	html, err := c.get(scBase + "/projects/" + projectID)
	if err != nil {
		return "", err
	}
	csrf := extractCSRF(html)

	allModels := reModels.FindAllStringSubmatch(html, -1)
	seen := map[string]bool{}
	var models []string
	for _, m := range allModels {
		if !seen[m[1]] {
			seen[m[1]] = true
			models = append(models, m[1])
		}
	}

	data := url.Values{
		"authenticity_token":  {csrf},
		"ticket[description]": {description},
		"button":              {""},
	}
	for _, mid := range models {
		val := "0"
		if mid == model {
			val = "1"
		}
		data.Set("implementations["+mid+"]", val)
	}

	if m := reBranch.FindStringSubmatch(html); m != nil {
		data.Set("ticket[base_branches]["+m[1]+"]", "main")
	}

	body, _, _, err := c.post(
		scBase+"/projects/"+projectID+"/tickets",
		data,
		map[string]string{
			"X-CSRF-Token": csrf,
			"Accept":       "text/vnd.turbo-stream.html, text/html, application/xhtml+xml",
		},
	)
	if err != nil {
		return "", err
	}

	tickets := reTicket.FindAllStringSubmatch(body, -1)
	if len(tickets) == 0 {
		return "", fmt.Errorf("ticket creation failed")
	}
	return tickets[0][1], nil
}

// ── Status ──────────────────────────────────────────────────

type TicketStatus struct {
	TicketID string
	Status   string // Running, Waiting, Completed, Failed
	Cost     string
	URL      string
}

func (c *Client) GetTicketStatus(ticketID string) (*TicketStatus, error) {
	html, err := c.get(scBase + "/tickets/" + ticketID)
	if err != nil {
		return nil, err
	}

	ts := &TicketStatus{
		TicketID: ticketID,
		URL:      scBase + "/tickets/" + ticketID,
	}

	statusRe := regexp.MustCompile(`(Running|Waiting|Completed|Failed)`)
	if m := statusRe.FindStringSubmatch(html); m != nil {
		ts.Status = m[1]
	}
	costRe := regexp.MustCompile(`\$(\d+\.\d+)`)
	if m := costRe.FindStringSubmatch(html); m != nil {
		ts.Cost = m[0]
	}
	return ts, nil
}

// ── Send follow-up message ──────────────────────────────────

func (c *Client) SendMessage(conversationID, content string) error {
	html, err := c.get(scBase + "/")
	if err != nil {
		return err
	}
	csrf := extractCSRF(html)

	data := url.Values{
		"authenticity_token":                      {csrf},
		"message[messageable_type]":               {"ChatMessage"},
		"message[shell_mode]":                     {"false"},
		"message[messageable_attributes][content]": {content},
		"button": {""},
	}
	_, _, _, err = c.post(
		scBase+"/conversations/"+conversationID+"/messages",
		data,
		map[string]string{
			"X-CSRF-Token": csrf,
			"Accept":       "text/vnd.turbo-stream.html, text/html, application/xhtml+xml",
		},
	)
	return err
}

// ── Proxy URL builder ───────────────────────────────────────

// ImplementationURL returns the full SC URL for embedding in the proxy.
func ImplementationURL(ticketID, implID string) string {
	return scBase + "/tickets/" + ticketID + "/implementations/" + implID
}

func keysOf(v url.Values) []string {
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}

// ── TOTP ────────────────────────────────────────────────────

func GenerateTOTP(secret string) string {
	key, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	counter := uint64(time.Now().Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)
	offset := h[len(h)-1] & 0x0F
	code := (binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7FFFFFFF) % 1000000
	return fmt.Sprintf("%06d", code)
}
