package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"edu-platform/internal/store"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.Email == "" || in.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	id, err := h.st.CreateUser(r.Context(), in.Email, string(hash), "student")
	if err != nil {
		http.Error(w, "create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "role": "student"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	_ = json.NewDecoder(r.Body).Decode(&in)
	u, err := h.st.GetUserByEmail(r.Context(), in.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	tok, err := h.makeAccessToken(u.ID, u.Role)
	if err != nil {
		http.Error(w, "token err", http.StatusInternalServerError)
		return
	}
	refresh := randomHex(32)
	exp := time.Now().Add(h.refreshExp)
	_ = h.st.SaveRefreshToken(r.Context(), store.RefreshToken{Token: refresh, UserID: u.ID, ExpiresAt: exp})
	json.NewEncoder(w).Encode(map[string]any{
		"access_token":  tok,
		"refresh_token": refresh,
		"expires_in":    int(h.accessExp.Seconds()),
		"role":          u.Role,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in struct{ RefreshToken string }
	_ = json.NewDecoder(r.Body).Decode(&in)
	t, err := h.st.GetRefreshToken(r.Context(), in.RefreshToken)
	if err != nil || t.ExpiresAt.Before(time.Now()) {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	tok, err := h.makeAccessToken(t.UserID, t.Role)
	if err != nil {
		http.Error(w, "token err", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"access_token": tok, "expires_in": int(h.accessExp.Seconds())})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var in struct{ RefreshToken string }
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.RefreshToken != "" {
		_ = h.st.DeleteRefreshToken(r.Context(), in.RefreshToken)
	}
	w.WriteHeader(http.StatusNoContent)
}
