package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

const RotationCost = 20.0 // LDC credits per rotation

// ── User balance ────────────────────────────────

func GetUserBalance(userID int64) (float64, error) {
	var balance float64
	err := DB.QueryRow("SELECT COALESCE(balance, 0) FROM users WHERE id = $1", userID).Scan(&balance)
	return balance, err
}

func DeductBalance(userID int64, amount float64) error {
	result, err := DB.Exec(
		"UPDATE users SET balance = balance - $1, total_rotations = total_rotations + 1, updated_at = now() WHERE id = $2 AND balance >= $1",
		amount, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient balance")
	}
	return nil
}

func AddBalance(userID int64, amount float64) error {
	_, err := DB.Exec("UPDATE users SET balance = balance + $1, updated_at = now() WHERE id = $2", amount, userID)
	return err
}

// ── Redemption codes ────────────────────────────

type RedemptionCode struct {
	ID        int64      `json:"id"`
	Code      string     `json:"code"`
	Rotations int        `json:"rotations"`
	UsedCount int        `json:"used_count"`
	MaxUses   int        `json:"max_uses"`
	CreatedBy int64      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func GenerateCode() string {
	b := make([]byte, 12)
	rand.Read(b)
	return "PRISM-" + hex.EncodeToString(b)[:16]
}

func CreateRedemptionCode(rotations, maxUses int, createdBy int64, expiresAt *time.Time) (*RedemptionCode, error) {
	code := GenerateCode()
	rc := &RedemptionCode{}
	err := DB.QueryRow(`
		INSERT INTO redemption_codes (code, rotations, max_uses, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, rotations, used_count, max_uses, created_by, created_at, expires_at
	`, code, rotations, maxUses, createdBy, expiresAt).Scan(
		&rc.ID, &rc.Code, &rc.Rotations, &rc.UsedCount, &rc.MaxUses, &rc.CreatedBy, &rc.CreatedAt, &rc.ExpiresAt,
	)
	return rc, err
}

func RedeemCode(code string, userID int64) (int, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var rc RedemptionCode
	err = tx.QueryRow(`
		SELECT id, code, rotations, used_count, max_uses, expires_at
		FROM redemption_codes WHERE code = $1 FOR UPDATE
	`, code).Scan(&rc.ID, &rc.Code, &rc.Rotations, &rc.UsedCount, &rc.MaxUses, &rc.ExpiresAt)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid code")
	}
	if err != nil {
		return 0, err
	}

	if rc.UsedCount >= rc.MaxUses {
		return 0, fmt.Errorf("code already fully used")
	}
	if rc.ExpiresAt != nil && time.Now().After(*rc.ExpiresAt) {
		return 0, fmt.Errorf("code expired")
	}

	// Check if user already used this code
	var exists int
	tx.QueryRow("SELECT 1 FROM redemption_log WHERE code_id = $1 AND user_id = $2", rc.ID, userID).Scan(&exists)
	if exists == 1 {
		return 0, fmt.Errorf("already redeemed")
	}

	// Add rotations as balance (each rotation = 20 credits)
	addAmount := float64(rc.Rotations) * RotationCost
	if _, err := tx.Exec("UPDATE users SET balance = balance + $1, updated_at = now() WHERE id = $2", addAmount, userID); err != nil {
		return 0, err
	}

	if _, err := tx.Exec("UPDATE redemption_codes SET used_count = used_count + 1 WHERE id = $1", rc.ID); err != nil {
		return 0, err
	}

	if _, err := tx.Exec("INSERT INTO redemption_log (code_id, user_id, rotations) VALUES ($1, $2, $3)", rc.ID, userID, rc.Rotations); err != nil {
		return 0, err
	}

	return rc.Rotations, tx.Commit()
}

func ListRedemptionCodes() ([]RedemptionCode, error) {
	rows, err := DB.Query(`
		SELECT id, code, rotations, used_count, max_uses, created_by, created_at, expires_at
		FROM redemption_codes ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []RedemptionCode
	for rows.Next() {
		var c RedemptionCode
		rows.Scan(&c.ID, &c.Code, &c.Rotations, &c.UsedCount, &c.MaxUses, &c.CreatedBy, &c.CreatedAt, &c.ExpiresAt)
		codes = append(codes, c)
	}
	return codes, nil
}

// ── Fingerprints ────────────────────────────────

type Fingerprint struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Language  string `json:"language"`
	Timezone  string `json:"timezone"`
	Screen    string `json:"screen"`
	Platform  string `json:"platform"`
}

func SaveFingerprint(userID int64, fp Fingerprint) error {
	_, err := DB.Exec(`
		INSERT INTO user_fingerprints (user_id, ip, user_agent, language, timezone, screen, platform)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, fp.IP, fp.UserAgent, fp.Language, fp.Timezone, fp.Screen, fp.Platform)
	return err
}

func GetLatestFingerprint(userID int64) (*Fingerprint, error) {
	fp := &Fingerprint{}
	err := DB.QueryRow(`
		SELECT ip, user_agent, language, timezone, screen, platform
		FROM user_fingerprints WHERE user_id = $1 ORDER BY collected_at DESC LIMIT 1
	`, userID).Scan(&fp.IP, &fp.UserAgent, &fp.Language, &fp.Timezone, &fp.Screen, &fp.Platform)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return fp, err
}
