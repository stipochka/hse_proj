package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"edu-platform/internal/s3"
	"edu-platform/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

const maxUploadSize = 20 << 20 // 20 MB

var allowedExts = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".png": true, ".jpg": true, ".jpeg": true,
	".zip": true, ".txt": true,
}

type Handler struct {
	st         *store.Store
	s3c        *s3.Client
	jwtSecret  []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

func New(st *store.Store, s3c *s3.Client) *Handler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	ae := parseDur(os.Getenv("ACCESS_TOKEN_EXP"), 15*time.Minute)
	re := parseDur(os.Getenv("REFRESH_TOKEN_EXP"), 7*24*time.Hour)
	return &Handler{st: st, s3c: s3c, jwtSecret: []byte(secret), accessExp: ae, refreshExp: re}
}

func parseDur(s string, d time.Duration) time.Duration {
	if s == "" {
		return d
	}
	if t, err := time.ParseDuration(s); err == nil {
		return t
	}
	return d
}

func (h *Handler) makeAccessToken(userID int64, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  strconv.FormatInt(userID, 10),
		"role": role,
		"exp":  now.Add(h.accessExp).Unix(),
		"iat":  now.Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(h.jwtSecret)
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
