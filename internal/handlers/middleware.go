package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type ctxUserKey struct{}
type ctxRoleKey struct{}

// parseJWT validates the Bearer token and returns userID + role.
func (h *Handler) parseJWT(r *http.Request) (int64, string, error) {
	ah := r.Header.Get("Authorization")
	if len(ah) <= 7 || ah[:7] != "Bearer " {
		return 0, "", fmt.Errorf("missing or malformed Authorization header")
	}
	token, err := jwt.Parse(ah[7:], func(t *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, "", fmt.Errorf("invalid sub claim")
	}
	uid, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid user id in token")
	}
	role, _ := claims["role"].(string)
	return uid, role, nil
}

func (h *Handler) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, role, err := h.parseJWT(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey{}, uid)
		ctx = context.WithValue(ctx, ctxRoleKey{}, role)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) AuthTeacher(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, role, err := h.parseJWT(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if role != "teacher" && role != "admin" {
			http.Error(w, "insufficient permissions (teacher role required)", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey{}, uid)
		ctx = context.WithValue(ctx, ctxRoleKey{}, role)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) AuthAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, role, err := h.parseJWT(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if role != "admin" {
			http.Error(w, "admin only", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey{}, uid)
		ctx = context.WithValue(ctx, ctxRoleKey{}, role)
		next(w, r.WithContext(ctx))
	}
}
