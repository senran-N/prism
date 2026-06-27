package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/prism/internal/db"
)

const sessionCookie = "prism_session"

func (s *Server) setSession(w http.ResponseWriter, userID int64) {
	value := signSession(s.cfg.SessionSecret, userID)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(s.cfg.BaseURL),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})
}

func (s *Server) getSessionUser(r *http.Request) *db.User {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	userID, ok := verifySession(s.cfg.SessionSecret, cookie.Value)
	if !ok {
		return nil
	}
	user, err := db.GetUser(userID)
	if err != nil || user == nil {
		return nil
	}
	return user
}

func (s *Server) clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func signSession(secret string, userID int64) string {
	data := fmt.Sprintf("%d.%d", userID, time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))
	return data + "." + sig
}

const maxSessionAge = 30 * 24 * time.Hour

func verifySession(secret, value string) (int64, bool) {
	parts := strings.SplitN(value, ".", 3)
	if len(parts) != 3 {
		return 0, false
	}
	data := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return 0, false
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, false
	}
	if time.Since(time.Unix(ts, 0)) > maxSessionAge {
		return 0, false
	}
	return userID, true
}
