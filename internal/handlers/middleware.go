package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type ctxUserKey struct{}
type ctxRoleKey struct{}
type ctxGroupIDKey struct{}
type ctxEmailKey struct{}

// rolePriority defines which role wins when a user has multiple roles assigned.
var rolePriority = map[string]int{"student": 0, "group_admin": 1, "super_admin": 2}

// extractClaims parses the Bearer token, extracts user identity from Keycloak claims,
// and upserts the user record in the database (JIT provisioning).
func (h *Handler) extractClaims(r *http.Request) (userID int64, role string, groupID int64, email string, err error) {
	ah := r.Header.Get("Authorization")
	if len(ah) <= 7 || ah[:7] != "Bearer " {
		return 0, "", 0, "", fmt.Errorf("missing or malformed Authorization header")
	}

	claims, err := h.jwks.Parse(ah[7:])
	if err != nil {
		return 0, "", 0, "", err
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return 0, "", 0, "", fmt.Errorf("missing sub claim")
	}

	email, _ = claims["email"].(string)

	// Pick the highest-priority role from realm_access.roles.
	role = "student"
	if ra, ok := claims["realm_access"].(map[string]any); ok {
		if roles, ok := ra["roles"].([]any); ok {
			for _, rv := range roles {
				if rs, ok := rv.(string); ok {
					if p, known := rolePriority[rs]; known && p > rolePriority[role] {
						role = rs
					}
				}
			}
		}
	}

	// group_id is a custom Keycloak user attribute mapped into the JWT for group_admin users.
	if gv, ok := claims["group_id"].(string); ok && gv != "" {
		groupID, _ = strconv.ParseInt(gv, 10, 64)
	}

	userID, err = h.st.GetOrCreateUser(r.Context(), sub, email, role)
	if err != nil {
		return 0, "", 0, "", fmt.Errorf("user provision: %w", err)
	}

	return userID, role, groupID, email, nil
}

func (h *Handler) withClaims(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	uid, role, groupID, email, err := h.extractClaims(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx := context.WithValue(r.Context(), ctxUserKey{}, uid)
	ctx = context.WithValue(ctx, ctxRoleKey{}, role)
	ctx = context.WithValue(ctx, ctxGroupIDKey{}, groupID)
	ctx = context.WithValue(ctx, ctxEmailKey{}, email)
	next(w, r.WithContext(ctx))
}

func (h *Handler) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.withClaims(w, r, next)
	}
}

func (h *Handler) AuthGroupAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.withClaims(w, r, func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(ctxRoleKey{}).(string)
			if role != "group_admin" && role != "super_admin" {
				http.Error(w, "group_admin role required", http.StatusForbidden)
				return
			}
			next(w, r)
		})
	}
}

func (h *Handler) AuthSuperAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.withClaims(w, r, func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(ctxRoleKey{}).(string)
			if role != "super_admin" {
				http.Error(w, "super_admin role required", http.StatusForbidden)
				return
			}
			next(w, r)
		})
	}
}
