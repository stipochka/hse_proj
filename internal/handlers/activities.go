package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"edu-platform/internal/domain"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	var in struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		Category     string `json:"category"`
		ActivityDate string `json:"activity_date"` // "2006-01-02"
		Draft        bool   `json:"draft"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	status := "submitted"
	if in.Draft {
		status = "draft"
	}

	a := domain.Activity{
		UserID:      uid,
		Title:       in.Title,
		Description: in.Description,
		Category:    in.Category,
		Status:      status,
	}
	if in.ActivityDate != "" {
		t, err := time.Parse("2006-01-02", in.ActivityDate)
		if err != nil {
			http.Error(w, "invalid activity_date, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		a.ActivityDate = &t
	}

	id, err := h.st.CreateActivity(r.Context(), a)
	if err != nil {
		http.Error(w, "create activity: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	role := r.Context().Value(ctxRoleKey{}).(string)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	a, err := h.st.GetActivity(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if role == "student" && a.UserID != uid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) ListMyActivities(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	list, err := h.st.ListActivities(r.Context(), uid)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []domain.Activity{}
	}
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.st.DeleteActivity(r.Context(), id, uid); err != nil {
		http.Error(w, "delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
