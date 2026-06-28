// Package credit integrates LinuxDo Credit payment system.
// Docs: https://credit.linux.do
// Flow: create order → redirect user to pay → notify callback → confirm
package api

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	creditBase     = "https://credit.linux.do"
	creditSubmit   = creditBase + "/pay/submit.php"
	creditQueryAPI = creditBase + "/api.php"
)

// ── Create payment order ────────────────────────

type CreatePaymentReq struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// POST /api/credit/pay — create payment order, return redirect URL
func (s *Server) handleCreditPay(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeError(w, 401, "login required")
		return
	}

	var req CreatePaymentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request")
		return
	}
	if req.Amount < 0.01 {
		writeError(w, 400, "minimum amount is 0.01")
		return
	}

	orderNo := fmt.Sprintf("PRISM-%d-%d", user.ID, time.Now().UnixMilli())
	if req.Description == "" {
		req.Description = "Prism Credits"
	}

	// Build payment URL
	params := url.Values{
		"pid":          {s.cfg.CreditClientID},
		"type":         {"epay"},
		"out_trade_no": {orderNo},
		"notify_url":   {s.cfg.BaseURL + "/api/credit/notify"},
		"return_url":   {s.cfg.BaseURL + "/api/credit/callback"},
		"name":         {req.Description},
		"money":        {fmt.Sprintf("%.2f", req.Amount)},
	}

	// Sign the request
	sign := signCreditParams(params, s.cfg.CreditClientSecret)
	params.Set("sign", sign)
	params.Set("sign_type", "MD5")

	payURL := creditSubmit + "?" + params.Encode()

	log.Printf("[credit] order created: %s amount=%.2f user=%d", orderNo, req.Amount, user.ID)

	writeJSON(w, map[string]string{
		"order_no": orderNo,
		"pay_url":  payURL,
	})
}

// ── Payment notify (server-to-server) ───────────

// POST /api/credit/notify — Credit system pushes payment result
func (s *Server) handleCreditNotify(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tradeNo := r.FormValue("trade_no")
	outTradeNo := r.FormValue("out_trade_no")
	tradeStatus := r.FormValue("trade_status")
	money := r.FormValue("money")
	sign := r.FormValue("sign")

	// Verify signature
	params := url.Values{}
	for k, v := range r.Form {
		if k != "sign" && k != "sign_type" {
			params[k] = v
		}
	}
	expectedSign := signCreditParams(params, s.cfg.CreditClientSecret)
	if !hmac.Equal([]byte(sign), []byte(expectedSign)) {
		log.Printf("[credit] notify: invalid signature for order %s", outTradeNo)
		w.Write([]byte("fail"))
		return
	}

	log.Printf("[credit] notify: order=%s trade=%s status=%s amount=%s",
		outTradeNo, tradeNo, tradeStatus, money)

	if tradeStatus == "TRADE_SUCCESS" {
		amount, _ := strconv.ParseFloat(money, 64)
		log.Printf("[credit] payment success: order=%s amount=%.2f", outTradeNo, amount)
		// TODO: credit user account based on outTradeNo prefix "PRISM-{userID}-..."
	}

	w.Write([]byte("success"))
}

// ── Payment return (browser redirect) ───────────

// GET /api/credit/callback — user returns after payment
func (s *Server) handleCreditCallback(w http.ResponseWriter, r *http.Request) {
	tradeStatus := r.URL.Query().Get("trade_status")
	outTradeNo := r.URL.Query().Get("out_trade_no")

	if tradeStatus == "TRADE_SUCCESS" {
		log.Printf("[credit] user returned: order=%s success", outTradeNo)
		http.Redirect(w, r, s.cfg.BaseURL+"/?payment=success", http.StatusFound)
	} else {
		http.Redirect(w, r, s.cfg.BaseURL+"/?payment=pending", http.StatusFound)
	}
}

// ── Query order status ──────────────────────────

// GET /api/credit/order?order_no=xxx
func (s *Server) handleCreditQuery(w http.ResponseWriter, r *http.Request) {
	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		writeError(w, 400, "order_no required")
		return
	}

	params := url.Values{
		"act":          {"order"},
		"pid":          {s.cfg.CreditClientID},
		"key":          {s.cfg.CreditClientSecret},
		"out_trade_no": {orderNo},
	}

	resp, err := http.Get(creditQueryAPI + "?" + params.Encode())
	if err != nil {
		writeError(w, 502, "query failed")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// ── Signature ───────────────────────────────────

func signCreditParams(params url.Values, secret string) string {
	// Sort keys alphabetically, concatenate k=v with &, append key
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" && params.Get(k) != "" {
			keys = append(keys, k)
		}
	}
	// Simple sort
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + params.Get(k)
	}
	str := strings.Join(parts, "&") + secret

	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
