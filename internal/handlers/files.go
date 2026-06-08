package handlers

import "net/http"

type fileURLResponse struct {
	FileURL string `json:"file_url"`
}

// GetActivityFile godoc
// @Summary  Get a presigned URL to view/download the activity PDF
// @Description Returns a short-lived presigned GET URL. The browser fetches the PDF directly from S3.
// @Tags     activities
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "Activity id"
// @Success  200 {object} fileURLResponse
// @Failure  403 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Router   /activities/{id}/file [get]
func (h *Handler) GetActivityFile(w http.ResponseWriter, r *http.Request) {
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
	if a.PDFKey == "" {
		writeErr(w, http.StatusNotFound, "no file for this activity")
		return
	}
	url, err := h.s3c.PresignGet(r.Context(), a.PDFKey, presignTTL)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "presign: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fileURLResponse{FileURL: url})
}
