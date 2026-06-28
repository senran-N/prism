// Package mail wraps the YYDS Mail API for temporary email creation.
package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const base = "https://maliapi.215.im/v1"

// Domains to try, in order. Wildcard-capable domains get unique child
// domains each time, making blocklisting much harder.
var Domains = []string{
	"a0.engineer",
	"3jsfuye.tech",
	"acgfun.eu.org",
}

type createReq struct {
	LocalPart string `json:"localPart"`
	Domain    string `json:"domain"`
}

type createResp struct {
	Success bool `json:"success"`
	Data    struct {
		Address string `json:"address"`
		Token   string `json:"token"`
		ID      string `json:"id"`
	} `json:"data"`
}

// CreateTempEmail tries multiple domains until one succeeds.
func CreateTempEmail(apiKey, prefix string) (string, string, error) {
	var lastErr error
	for _, domain := range Domains {
		addr, token, err := createWithDomain(apiKey, prefix, domain)
		if err == nil {
			return addr, token, nil
		}
		lastErr = err
	}
	return "", "", fmt.Errorf("all domains failed, last: %v", lastErr)
}

func createWithDomain(apiKey, prefix, domain string) (string, string, error) {
	payload, _ := json.Marshal(createReq{LocalPart: prefix, Domain: domain})
	req, _ := http.NewRequest("POST", base+"/accounts", bytes.NewReader(payload))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result createResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", err
	}
	if !result.Success {
		return "", "", fmt.Errorf("domain %s: %s", domain, string(body))
	}
	return result.Data.Address, result.Data.Token, nil
}
