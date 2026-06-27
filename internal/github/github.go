// Package github handles GitHub login with TOTP 2FA.
package github

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/senran-N/prism/internal/scproto"
)

func readBody(resp *http.Response) string {
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err == nil {
			defer gz.Close()
			reader = gz
		}
	}
	b, _ := io.ReadAll(reader)
	return string(b)
}

var reCSRF = regexp.MustCompile(`authenticity_token" value="([^"]+)`)

// Login authenticates with GitHub and returns the logged-in http.Client.
func Login(username, password, totpSecret string) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	fp := scproto.RandomFingerprint()
	client := &http.Client{
		Jar:       jar,
		Transport: &http.Transport{TLSClientConfig: scproto.TLSConfig()},
	}

	// GET /login
	getReq, _ := http.NewRequest("GET", "https://github.com/login", nil)
	fp.ApplyHeaders(getReq)
	resp, err := client.Do(getReq)
	if err != nil {
		return nil, err
	}
	html := readBody(resp)
	resp.Body.Close()

	m := reCSRF.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("no CSRF on github login")
	}

	// POST /session
	data := url.Values{
		"authenticity_token": {m[1]},
		"login":              {username},
		"password":           {password},
		"commit":             {"Sign in"},
	}
	req, _ := http.NewRequest("POST", "https://github.com/session", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fp.ApplyHeaders(req)
	req.Header.Set("Origin", "https://github.com")
	req.Header.Set("Referer", "https://github.com/login")
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	html = readBody(resp)
	resp.Body.Close()

	// Handle 2FA
	if strings.Contains(resp.Request.URL.String(), "two-factor") || strings.Contains(html, "two-factor") {
		if totpSecret == "" {
			return nil, fmt.Errorf("2FA required but no TOTP secret")
		}
		code := scproto.GenerateTOTP(totpSecret)

		m2 := reCSRF.FindStringSubmatch(html)
		if m2 == nil {
			return nil, fmt.Errorf("no CSRF on 2FA page")
		}

		data2 := url.Values{
			"authenticity_token": {m2[1]},
			"app_otp":            {code},
		}
		req2, _ := http.NewRequest("POST", "https://github.com/sessions/two-factor", strings.NewReader(data2.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		fp.ApplyHeaders(req2)
		resp, err = client.Do(req2)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
	}

	if strings.Contains(resp.Request.URL.String(), "/login") {
		return nil, fmt.Errorf("github login failed")
	}

	return client, nil
}
