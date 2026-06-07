package handlers

import (
	"encoding/json"
	"net/http"

	"edu-platform/internal/domain"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	role := r.Context().Value(ctxRoleKey{}).(string)
	groupID := r.Context().Value(ctxGroupIDKey{}).(int64)
	email := r.Context().Value(ctxEmailKey{}).(string)

	bal, _ := h.st.GetBalance(r.Context(), uid)

	json.NewEncoder(w).Encode(map[string]any{
		"user_id":  uid,
		"email":    email,
		"role":     role,
		"group_id": groupID,
		"balance":  bal,
	})
}

func (h *Handler) MyBalance(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	bal, err := h.st.GetBalance(r.Context(), uid)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"balance": bal})
}

func (h *Handler) MyTransactions(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	txns, err := h.st.ListTransactions(r.Context(), uid)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if txns == nil {
		txns = []domain.Transaction{}
	}
	json.NewEncoder(w).Encode(txns)
}

func (h *Handler) MyEvaluations(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	evals, err := h.st.ListEvaluationsByUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if evals == nil {
		evals = []domain.Evaluation{}
	}
	json.NewEncoder(w).Encode(evals)
}
