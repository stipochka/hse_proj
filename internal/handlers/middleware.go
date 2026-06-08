package handlers

import (
	"context"
	"net/http"
	"strings"
)

type ctxSubKey struct{}
type ctxRoleKey struct{}
type ctxGroupKey struct{}
type ctxEmailKey struct{}
type ctxUsernameKey struct{}

var rolePriority = map[string]int{"student": 0, "group_admin": 1, "super_admin": 2}

type identity struct {
	sub      string
	email    string
	username string // preferred_username
	role     string
	group    string
}

func (h *Handler) authenticate(r *http.Request) (identity, error) {
	var id identity
	ah := r.Header.Get("Authorization")
	if len(ah) <= 7 || !strings.EqualFold(ah[:7], "Bearer ") {
		return id, errUnauthorized("missing or malformed Authorization header")
	}
	claims, err := h.jwks.Parse(ah[7:])
	if err != nil {
		return id, errUnauthorized(err.Error())
	}

	id.sub, _ = claims["sub"].(string)
	if id.sub == "" {
		return id, errUnauthorized("missing sub claim")
	}
	id.email, _ = claims["email"].(string)
	id.username, _ = claims["preferred_username"].(string)

	id.role = "student"
	if ra, ok := claims["realm_access"].(map[string]any); ok {
		if roles, ok := ra["roles"].([]any); ok {
			for _, rv := range roles {
				if rs, ok := rv.(string); ok {
					if rolePriority[rs] > rolePriority[id.role] {
						id.role = rs
					}
				}
			}
		}
	}

	if gs, ok := claims["groups"].([]any); ok {
		for _, gv := range gs {
			if g, ok := gv.(string); ok && g != "" {
				id.group = strings.TrimPrefix(g, "/")
				break
			}
		}
	}
	return id, nil
}

type authError struct{ msg string }

func (e authError) Error() string      { return e.msg }
func errUnauthorized(msg string) error { return authError{msg} }

func (h *Handler) withIdentity(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	id, err := h.authenticate(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	ctx := context.WithValue(r.Context(), ctxSubKey{}, id.sub)
	ctx = context.WithValue(ctx, ctxRoleKey{}, id.role)
	ctx = context.WithValue(ctx, ctxGroupKey{}, id.group)
	ctx = context.WithValue(ctx, ctxEmailKey{}, id.email)
	ctx = context.WithValue(ctx, ctxUsernameKey{}, id.username)
	next(w, r.WithContext(ctx))
}

func (h *Handler) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { h.withIdentity(w, r, next) }
}

func (h *Handler) AuthAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.withIdentity(w, r, func(w http.ResponseWriter, r *http.Request) {
			if role(r) != "group_admin" && role(r) != "super_admin" {
				writeErr(w, http.StatusForbidden, "admin role required")
				return
			}
			next(w, r)
		})
	}
}

func sub(r *http.Request) string      { v, _ := r.Context().Value(ctxSubKey{}).(string); return v }
func role(r *http.Request) string     { v, _ := r.Context().Value(ctxRoleKey{}).(string); return v }
func group(r *http.Request) string    { v, _ := r.Context().Value(ctxGroupKey{}).(string); return v }
func username(r *http.Request) string { v, _ := r.Context().Value(ctxUsernameKey{}).(string); return v }
