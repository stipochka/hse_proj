package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"edu-platform/internal/domain"
	"edu-platform/internal/store"
)

// adminScope resolves the group filter: group_admin is hard-scoped to their own
// group; super_admin may optionally pass ?group=.
func adminScope(r *http.Request) string {
	if role(r) == "super_admin" {
		return r.URL.Query().Get("group")
	}
	return group(r)
}

// ListActivities godoc
// @Summary  Admin activity feed (group-scoped)
// @Description group_admin sees only their own group; super_admin may filter by ?group.
// @Tags     admin
// @Security BearerAuth
// @Produce  json
// @Param    group      query string false "Group (super_admin only)"
// @Param    status     query string false "Filter by status"
// @Param    category   query string false "Filter by category"
// @Param    student_id query string false "Filter by Keycloak student id"
// @Success  200 {array} domain.Activity
// @Router   /activities [get]
func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.st.ListActivities(r.Context(), store.ActivityFilter{
		Group:     adminScope(r),
		StudentID: q.Get("student_id"),
		Status:    q.Get("status"),
		Category:  q.Get("category"),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Summary godoc
// @Summary  Aggregate stats per student (group-scoped)
// @Tags     admin
// @Security BearerAuth
// @Produce  json
// @Param    group      query string false "Group (super_admin only)"
// @Param    category   query string false "Filter by category"
// @Param    student_id query string false "Filter by Keycloak student id"
// @Success  200 {array} domain.StudentStats
// @Router   /dashboard/summary [get]
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	stats, err := h.summary(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// ExportSummary godoc
// @Summary  Export the admin summary as CSV
// @Tags     admin
// @Security BearerAuth
// @Produce  text/csv
// @Param    group    query string false "Group (super_admin only)"
// @Param    category query string false "Filter by category"
// @Success  200 {string} string "CSV"
// @Router   /export/summary [get]
func (h *Handler) ExportSummary(w http.ResponseWriter, r *http.Request) {
	stats, err := h.summary(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="summary.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"student_id", "group", "points", "credits", "activities", "evaluated"})
	for _, s := range stats {
		_ = cw.Write([]string{
			s.StudentID,
			s.StudentGroup,
			strconv.FormatInt(s.TotalPoints, 10),
			strconv.FormatFloat(s.TotalCredits, 'f', 2, 64),
			strconv.Itoa(s.ActivityCount),
			strconv.Itoa(s.EvaluatedCount),
		})
	}
	cw.Flush()
}

func (h *Handler) summary(r *http.Request) ([]domain.StudentStats, error) {
	q := r.URL.Query()
	return h.st.Summary(r.Context(), store.SummaryFilter{
		Group:     adminScope(r),
		StudentID: q.Get("student_id"),
		Category:  q.Get("category"),
	})
}
