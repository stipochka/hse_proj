package handlers

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMakeAccessTokenWithRole(t *testing.T) {
	h := &Handler{jwtSecret: []byte("test-secret"), accessExp: 15 * time.Minute}
	tok, err := h.makeAccessToken(12345, "teacher")
	if err != nil {
		t.Fatalf("make token: %v", err)
	}

	parsed, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if !parsed.Valid {
		t.Fatalf("token not valid")
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["role"] != "teacher" {
		t.Fatalf("role claim: got %v", claims["role"])
	}
	if claims["sub"] != "12345" {
		t.Fatalf("sub claim: got %v", claims["sub"])
	}
}
