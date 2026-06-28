package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/senran-N/prism/internal/db"
)

// GET /api/balance — get user balance and rotation info
func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeError(w, 401, "login required")
		return
	}
	balance, _ := db.GetUserBalance(user.ID)
	writeJSON(w, map[string]any{
		"balance":          balance,
		"rotation_cost":    db.RotationCost,
		"can_rotate":       balance >= db.RotationCost,
		"total_rotations":  user.TotalRotations,
	})
}

// POST /api/redeem — redeem a code for rotation credits
func (s *Server) handleRedeem(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeError(w, 401, "login required")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeError(w, 400, "code is required")
		return
	}

	rotations, err := db.RedeemCode(req.Code, user.ID)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}

	balance, _ := db.GetUserBalance(user.ID)
	log.Printf("[credits] redeemed: user=%d code=%s rotations=%d new_balance=%.2f", user.ID, req.Code, rotations, balance)

	writeJSON(w, map[string]any{
		"rotations": rotations,
		"credits":   float64(rotations) * db.RotationCost,
		"balance":   balance,
	})
}

// POST /api/fingerprint — collect user browser fingerprint
func (s *Server) handleFingerprint(w http.ResponseWriter, r *http.Request) {
	user := s.getSessionUser(r)
	if user == nil {
		writeError(w, 401, "login required")
		return
	}

	var fp db.Fingerprint
	if err := json.NewDecoder(r.Body).Decode(&fp); err != nil {
		writeError(w, 400, "invalid fingerprint data")
		return
	}

	// Override IP with actual client IP
	fp.IP = clientIP(r)

	if err := db.SaveFingerprint(user.ID, fp); err != nil {
		log.Printf("[fingerprint] save error: %v", err)
	}

	writeJSON(w, map[string]bool{"saved": true})
}

// ── Admin: redemption code management ───────────

// POST /api/admin/codes — generate redemption codes
func (s *Server) handleAdminCreateCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Rotations int    `json:"rotations"` // how many rotations this code grants
		MaxUses   int    `json:"max_uses"`  // how many users can use it
		Count     int    `json:"count"`     // how many codes to generate
		ExpiresIn string `json:"expires_in"` // "24h", "7d", "" for no expiry
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request")
		return
	}
	if req.Rotations < 1 {
		req.Rotations = 1
	}
	if req.MaxUses < 1 {
		req.MaxUses = 1
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.Count > 100 {
		req.Count = 100
	}

	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			// Try "7d" format
			if len(req.ExpiresIn) > 1 && req.ExpiresIn[len(req.ExpiresIn)-1] == 'd' {
				days := 0
				for _, c := range req.ExpiresIn[:len(req.ExpiresIn)-1] {
					days = days*10 + int(c-'0')
				}
				d = time.Duration(days) * 24 * time.Hour
			} else {
				writeError(w, 400, "invalid expires_in format")
				return
			}
		}
		t := time.Now().Add(d)
		expiresAt = &t
	}

	admin := s.getSessionUser(r)
	adminID := int64(0)
	if admin != nil {
		adminID = admin.ID
	}

	var codes []db.RedemptionCode
	for i := 0; i < req.Count; i++ {
		code, err := db.CreateRedemptionCode(req.Rotations, req.MaxUses, adminID, expiresAt)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		codes = append(codes, *code)
	}

	log.Printf("[admin] generated %d codes: rotations=%d max_uses=%d", req.Count, req.Rotations, req.MaxUses)
	writeJSON(w, codes)
}

// GET /api/admin/codes — list all codes
func (s *Server) handleAdminListCodes(w http.ResponseWriter, r *http.Request) {
	codes, err := db.ListRedemptionCodes()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if codes == nil {
		codes = []db.RedemptionCode{}
	}
	writeJSON(w, codes)
}
