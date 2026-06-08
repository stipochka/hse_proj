package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"edu-platform/internal/jwks"
	"edu-platform/internal/s3"
	"edu-platform/internal/store"
)

// presignTTL is the lifetime of the presigned upload/download URLs handed to the client.
const presignTTL = 15 * time.Minute

type Handler struct {
	st   *store.Store
	s3c  s3.Storage
	jwks jwks.Parser
}

func New(st *store.Store, s3c s3.Storage, jw jwks.Parser) *Handler {
	return &Handler{st: st, s3c: s3c, jwks: jw}
}

// errorResponse is the shape of every error body (documented for swagger).
type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
