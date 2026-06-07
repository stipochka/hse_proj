package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"edu-platform/internal/domain"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)

	activityIDStr := r.FormValue("activity_id")
	if activityIDStr == "" {
		http.Error(w, "activity_id required", http.StatusBadRequest)
		return
	}
	activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid activity_id", http.StatusBadRequest)
		return
	}

	act, err := h.st.GetActivity(r.Context(), activityID)
	if err != nil || act.UserID != uid {
		http.Error(w, "activity not found or forbidden", http.StatusForbidden)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large (max 20 MB)", http.StatusRequestEntityTooLarge)
		return
	}

	f, fh, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExts[ext] {
		http.Error(w, "only PDF files are allowed", http.StatusBadRequest)
		return
	}

	s3Key := fmt.Sprintf("activities/%d/%d_%s%s", activityID, time.Now().UnixNano(), randomHex(8), ext)
	if err := h.s3c.Upload(r.Context(), s3Key, "application/pdf", f, fh.Size); err != nil {
		http.Error(w, "upload error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileRec := domain.ActivityFile{
		ActivityID: activityID,
		Filename:   filepath.Base(fh.Filename),
		S3Key:      s3Key,
		SizeBytes:  fh.Size,
	}
	fileID, err := h.st.CreateActivityFile(r.Context(), fileRec)
	if err != nil {
		_ = h.s3c.Delete(r.Context(), s3Key)
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"file_id": fileID, "filename": fh.Filename})
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(ctxUserKey{}).(int64)
	role := r.Context().Value(ctxRoleKey{}).(string)

	fileID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	f, err := h.st.GetActivityFile(r.Context(), fileID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if role == "student" {
		act, err := h.st.GetActivity(r.Context(), f.ActivityID)
		if err != nil || act.UserID != uid {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	obj, meta, err := h.s3c.Download(r.Context(), f.S3Key)
	if err != nil {
		http.Error(w, "download error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+f.Filename+`"`)
	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	io.Copy(w, obj)
}
