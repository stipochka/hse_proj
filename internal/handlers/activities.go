package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"edu-platform/internal/domain"
	"edu-platform/internal/store"

	"github.com/go-chi/chi/v5"
)

type uploadURLRequest struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type uploadURLResponse struct {
	ActivityID int64  `json:"activity_id"`
	UploadURL  string `json:"upload_url"`
	PDFKey     string `json:"pdf_key"`
}

// UploadURL godoc
// @Summary  Create activity and get a presigned PDF upload URL
// @Description Creates a PENDING activity (group snapshotted from the token) and returns a presigned PUT URL. The browser uploads the PDF directly to S3, then calls confirm.
// @Tags     activities
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body uploadURLRequest true "Activity metadata"
// @Success  201 {object} uploadURLResponse
// @Failure  400 {object} errorResponse
// @Failure  401 {object} errorResponse
// @Router   /activities/upload-url [post]
func (h *Handler) UploadURL(w http.ResponseWriter, r *http.Request) {
	var in uploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if in.Title == "" {
		writeErr(w, http.StatusBadRequest, "title required")
		return
	}

	pdfKey := "activities/" + randomHex(16) + ".pdf"
	id, err := h.st.CreateActivity(r.Context(), domain.Activity{
		StudentID:    sub(r),
		StudentName:  username(r),
		StudentGroup: group(r),
		Title:        in.Title,
		Category:     in.Category,
		Description:  in.Description,
		PDFKey:       pdfKey,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create activity: "+err.Error())
		return
	}

	url, err := h.s3c.PresignPut(r.Context(), pdfKey, "application/pdf", presignTTL)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "presign: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, uploadURLResponse{ActivityID: id, UploadURL: url, PDFKey: pdfKey})
}

// Confirm godoc
// @Summary  Confirm a finished PDF upload
// @Description Verifies the object exists in S3 (HEAD) and moves the activity from PENDING to SUBMITTED.
// @Tags     activities
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "Activity id"
// @Success  200 {object} domain.Activity
// @Failure  400 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Router   /activities/{id}/confirm [post]
func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := h.st.GetActivity(r.Context(), id)
	if err != nil || a.StudentID != sub(r) {
		writeErr(w, http.StatusNotFound, "activity not found")
		return
	}
	if _, err := h.s3c.Stat(r.Context(), a.PDFKey); err != nil {
		writeErr(w, http.StatusBadRequest, "uploaded file not found in storage")
		return
	}
	if err := h.st.ConfirmActivity(r.Context(), id, sub(r)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusBadRequest, "activity is not in PENDING state")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	a, _ = h.st.GetActivity(r.Context(), id)
	writeJSON(w, http.StatusOK, a)
}

// ListMyActivities godoc
// @Summary  List my activities
// @Tags     activities
// @Security BearerAuth
// @Produce  json
// @Param    status   query string false "Filter by status"
// @Param    category query string false "Filter by category"
// @Success  200 {array} domain.Activity
// @Router   /activities/my [get]
func (h *Handler) ListMyActivities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.st.ListMyActivities(r.Context(), sub(r), q.Get("status"), q.Get("category"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GetActivity godoc
// @Summary  Get activity details
// @Tags     activities
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "Activity id"
// @Success  200 {object} domain.Activity
// @Failure  403 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Router   /activities/{id} [get]
func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := h.st.GetActivity(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if !canAccessActivity(r, a) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// canAccessActivity: students see their own; group_admin sees their group; super_admin sees all.
func canAccessActivity(r *http.Request, a *domain.Activity) bool {
	switch role(r) {
	case "super_admin":
		return true
	case "group_admin":
		return a.StudentGroup == group(r)
	default:
		return a.StudentID == sub(r)
	}
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
