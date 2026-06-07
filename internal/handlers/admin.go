package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"

	"edu-platform/internal/domain"
	"edu-platform/internal/store"
)

// ListAdminActivities serves the activity feed for group_admin and super_admin.
// group_admin is restricted to their own group (from the JWT claim); super_admin may pass group_id as a query param.
func (h *Handler) ListAdminActivities(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(ctxRoleKey{}).(string)
	tokenGroupID := r.Context().Value(ctxGroupIDKey{}).(int64)

	q := r.URL.Query()

	f := store.ActivityFilter{
		Status:   q.Get("status"),
		Category: q.Get("category"),
	}

	if v := q.Get("student_id"); v != "" {
		f.StudentID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := q.Get("limit"); v != "" {
		f.Limit, _ = strconv.Atoi(v)
	}
	if v := q.Get("offset"); v != "" {
		f.Offset, _ = strconv.Atoi(v)
	}

	switch role {
	case "group_admin":
		// Hard-scoped to the group from the token; query param is ignored.
		f.GroupID = tokenGroupID
	case "super_admin":
		if v := q.Get("group_id"); v != "" {
			f.GroupID, _ = strconv.ParseInt(v, 10, 64)
		}
	}

	list, err := h.st.ListActivitiesAdmin(r.Context(), f)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []domain.Activity{}
	}
	json.NewEncoder(w).Encode(list)
}

// AdminReports returns aggregate statistics per student.
// group_admin sees only their own group; super_admin can filter freely.
func (h *Handler) AdminReports(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(ctxRoleKey{}).(string)
	tokenGroupID := r.Context().Value(ctxGroupIDKey{}).(int64)

	q := r.URL.Query()
	f := store.ReportFilter{}

	switch role {
	case "group_admin":
		f.GroupID = tokenGroupID
	case "super_admin":
		if v := q.Get("group_id"); v != "" {
			f.GroupID, _ = strconv.ParseInt(v, 10, 64)
		}
		if v := q.Get("user_id"); v != "" {
			f.UserID, _ = strconv.ParseInt(v, 10, 64)
		}
		if v := q.Get("course_id"); v != "" {
			f.CourseID, _ = strconv.ParseInt(v, 10, 64)
		}
		f.Stream = q.Get("stream")
	}

	stats, err := h.st.AdminReport(r.Context(), f)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if stats == nil {
		stats = []domain.StudentStats{}
	}

	if q.Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="report.csv"`)
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"user_id", "email", "group", "currency", "credits", "activities"})
		for _, s := range stats {
			_ = cw.Write([]string{
				strconv.FormatInt(s.UserID, 10),
				s.Email,
				s.GroupName,
				strconv.FormatInt(s.TotalCurrency, 10),
				strconv.FormatFloat(s.TotalCredits, 'f', 2, 64),
				strconv.Itoa(s.ActivityCount),
			})
		}
		cw.Flush()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name       string `json:"name"`
		Stream     string `json:"stream"`
		CourseYear int    `json:"course_year"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	id, err := h.st.CreateGroup(r.Context(), in.Name, in.Stream, in.CourseYear)
	if err != nil {
		http.Error(w, "create group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	list, err := h.st.ListGroups(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []domain.Group{}
	}
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) AssignUserToGroup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID  int64 `json:"user_id"`
		GroupID int64 `json:"group_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.UserID == 0 || in.GroupID == 0 {
		http.Error(w, "user_id and group_id required", http.StatusBadRequest)
		return
	}
	if err := h.st.AssignUserToGroup(r.Context(), in.UserID, in.GroupID); err != nil {
		http.Error(w, "assign: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID  int64 `json:"user_id"`
		GroupID int64 `json:"group_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.st.RemoveUserFromGroup(r.Context(), in.UserID, in.GroupID); err != nil {
		http.Error(w, "remove: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	id, err := h.st.CreateCourse(r.Context(), in.Name)
	if err != nil {
		http.Error(w, "create course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	list, err := h.st.ListCourses(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []domain.Course{}
	}
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) AssignUserToCourse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID   int64 `json:"user_id"`
		CourseID int64 `json:"course_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.UserID == 0 || in.CourseID == 0 {
		http.Error(w, "user_id and course_id required", http.StatusBadRequest)
		return
	}
	if err := h.st.AssignUserToCourse(r.Context(), in.UserID, in.CourseID); err != nil {
		http.Error(w, "assign: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
