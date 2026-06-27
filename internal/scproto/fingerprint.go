package scproto

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Fingerprint holds a browser identity used for a single session.
type Fingerprint struct {
	UserAgent      string
	AcceptLanguage string
	Platform       string
	SecCHUA        string
	Viewport       [2]int // width, height
}

var chromeVersions = []struct {
	version int
	build   string
}{
	{149, "149.0.7827.196"},
	{148, "148.0.7816.100"},
	{147, "147.0.7793.132"},
	{146, "146.0.7775.116"},
	{145, "145.0.7764.85"},
}

var platforms = []struct {
	os        string
	uaPart    string
	secCHPlat string
}{
	{"linux", "X11; Linux x86_64", "\"Linux\""},
	{"mac", "Macintosh; Intel Mac OS X 14_7_6", "\"macOS\""},
	{"mac", "Macintosh; Intel Mac OS X 15_5", "\"macOS\""},
	{"win", "Windows NT 10.0; Win64; x64", "\"Windows\""},
	{"win", "Windows NT 11.0; Win64; x64", "\"Windows\""},
}

var languages = []string{
	"en-US,en;q=0.9",
	"en-US,en;q=0.9,zh-CN;q=0.8",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.9,ja;q=0.8",
	"zh-CN,zh;q=0.9,en;q=0.8",
}

var viewports = [][2]int{
	{1920, 1080},
	{2560, 1440},
	{1440, 900},
	{1536, 864},
	{1366, 768},
	{1680, 1050},
}

// RandomFingerprint generates a realistic browser fingerprint.
func RandomFingerprint() Fingerprint {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	chrome := chromeVersions[r.Intn(len(chromeVersions))]
	plat := platforms[r.Intn(len(platforms))]
	lang := languages[r.Intn(len(languages))]
	vp := viewports[r.Intn(len(viewports))]

	ua := fmt.Sprintf(
		"Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		plat.uaPart, chrome.build,
	)

	secCHUA := fmt.Sprintf(
		`"Chromium";v="%d", "Google Chrome";v="%d", "Not-A.Brand";v="99"`,
		chrome.version, chrome.version,
	)

	return Fingerprint{
		UserAgent:      ua,
		AcceptLanguage: lang,
		Platform:       plat.secCHPlat,
		SecCHUA:        secCHUA,
		Viewport:       vp,
	}
}

// ApplyHeaders sets all fingerprint headers on an http.Request.
func (fp Fingerprint) ApplyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", fp.UserAgent)
	req.Header.Set("Accept-Language", fp.AcceptLanguage)
	req.Header.Set("Sec-CH-UA", fp.SecCHUA)
	req.Header.Set("Sec-CH-UA-Platform", fp.Platform)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")
}

// RandomDelay sleeps for a human-like random interval.
// base ± jitter, e.g. RandomDelay(3*time.Second, 2*time.Second) → 1~5s
func RandomDelay(base, jitter time.Duration) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := base + time.Duration(r.Int63n(int64(2*jitter))) - jitter
	if d < time.Second {
		d = time.Second
	}
	time.Sleep(d)
}

// TLSConfig returns a randomized TLS config to vary the JA3 fingerprint.
func TLSConfig() *tls.Config {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Shuffle cipher suite order slightly
	suites := []uint16{
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	}
	r.Shuffle(len(suites), func(i, j int) { suites[i], suites[j] = suites[j], suites[i] })

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		CipherSuites: suites,
	}
}
