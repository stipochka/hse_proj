package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"edu-platform/internal/domain"
)

// DashboardMe godoc
// @Summary  Personal dashboard aggregates
// @Tags     dashboard
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} domain.DashboardMe
// @Router   /dashboard/me [get]
func (h *Handler) DashboardMe(w http.ResponseWriter, r *http.Request) {
	var d domain.DashboardMe
	d, err := h.st.DashboardMe(r.Context(), sub(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// ExportMe godoc
// @Summary  Export my activities as CSV
// @Tags     dashboard
// @Security BearerAuth
// @Produce  text/csv
// @Success  200 {string} string "CSV"
// @Router   /export/me [get]
func (h *Handler) ExportMe(w http.ResponseWriter, r *http.Request) {
	list, err := h.st.ListMyActivities(r.Context(), sub(r), "", "")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="my_activities.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "title", "category", "status", "points", "credits", "comment", "created_at"})
	for _, a := range list {
		points, credits, comment := "", "", ""
		if a.Evaluation != nil {
			points = strconv.Itoa(a.Evaluation.Points)
			credits = strconv.FormatFloat(a.Evaluation.Credits, 'f', 2, 64)
			comment = a.Evaluation.Comment
		}
		_ = cw.Write([]string{
			strconv.FormatInt(a.ID, 10),
			a.Title, a.Category, a.Status, points, credits, comment,
			a.CreatedAt.Format("2006-01-02"),
		})
	}
	cw.Flush()
}
