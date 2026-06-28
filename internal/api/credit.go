// Package credit integrates LinuxDo Credit payment system (EasyPay compatible).
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
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	creditBase = "https://credit.linux.do"
)

var (
	creditOrders   = make(map[string]map[string]string)
	creditOrdersMu sync.Mutex
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

	params := map[string]string{
		"pid":          s.cfg.CreditClientID,
		"type":         "epay",
		"out_trade_no": orderNo,
		"notify_url":   s.cfg.BaseURL + "/api/credit/notify",
		"return_url":   s.cfg.BaseURL + "/api/credit/callback",
		"name":         req.Description,
		"money":        fmt.Sprintf("%.2f", req.Amount),
	}

	params["sign"] = epaySign(params, s.cfg.CreditClientSecret)
	params["sign_type"] = "MD5"

	log.Printf("[credit] order created: %s amount=%.2f user=%d sign=%s", orderNo, req.Amount, user.ID, params["sign"])

	writeJSON(w, map[string]string{
		"order_no": orderNo,
		"pay_url":  "/api/credit/redirect?order=" + orderNo,
	})

	// Cache params for the redirect page
	creditOrdersMu.Lock()
	creditOrders[orderNo] = params
	creditOrdersMu.Unlock()
}

// GET /api/credit/redirect?order=xxx — self-hosted page that auto-POSTs to credit
func (s *Server) handleCreditRedirect(w http.ResponseWriter, r *http.Request) {
	orderNo := r.URL.Query().Get("order")

	creditOrdersMu.Lock()
	params, ok := creditOrders[orderNo]
	if ok {
		delete(creditOrders, orderNo) // one-time use
	}
	creditOrdersMu.Unlock()

	if !ok {
		http.Error(w, "order not found or expired", 404)
		return
	}

	// Render auto-submit form page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Redirecting to payment...</title>
<style>body{font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#f6f9fc;color:#697386;}</style>
</head><body><div style="text-align:center"><p>Redirecting to LinuxDo Credit...</p><noscript>Please enable JavaScript</noscript></div>
<form id="f" method="POST" action="%s">`, creditBase+"/pay/submit.php")

	for k, v := range params {
		fmt.Fprintf(w, `<input type="hidden" name="%s" value="%s">`, k, v)
	}

	fmt.Fprint(w, `</form><script>document.getElementById("f").submit();</script></body></html>`)
}

// ── Payment notify (server-to-server) ───────────

// POST /api/credit/notify — Credit system pushes payment result
func (s *Server) handleCreditNotify(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		w.Write([]byte("fail"))
		return
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		w.Write([]byte("fail"))
		return
	}

	params := make(map[string]string)
	for k := range values {
		params[k] = values.Get(k)
	}

	sign := params["sign"]
	if sign == "" {
		log.Printf("[credit] notify: missing sign")
		w.Write([]byte("fail"))
		return
	}

	if !epayVerifySign(params, s.cfg.CreditClientSecret, sign) {
		log.Printf("[credit] notify: invalid signature for order %s", params["out_trade_no"])
		w.Write([]byte("fail"))
		return
	}

	log.Printf("[credit] notify: order=%s trade=%s status=%s amount=%s",
		params["out_trade_no"], params["trade_no"], params["trade_status"], params["money"])

	if params["trade_status"] == "TRADE_SUCCESS" {
		log.Printf("[credit] payment success: order=%s amount=%s", params["out_trade_no"], params["money"])
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

	resp, err := http.Get(creditBase + "/api.php?" + params.Encode())
	if err != nil {
		writeError(w, 502, "query failed")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

// ── EasyPay signing (identical to sub2api) ──────

func epaySign(params map[string]string, pkey string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k + "=" + params[k])
	}
	buf.WriteString(pkey)
	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}

func epayVerifySign(params map[string]string, pkey string, sign string) bool {
	return hmac.Equal([]byte(epaySign(params, pkey)), []byte(sign))
}
