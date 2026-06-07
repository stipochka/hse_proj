package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"edu-platform/internal/jwks"
	"edu-platform/internal/s3"
	"edu-platform/internal/store"
)

const maxUploadSize = 20 << 20 // 20 MB

var allowedExts = map[string]bool{
	".pdf": true,
}

type Handler struct {
	st   *store.Store
	s3c  s3.Storage
	jwks jwks.Parser
}

func New(st *store.Store, s3c s3.Storage, jw jwks.Parser) *Handler {
	return &Handler{st: st, s3c: s3c, jwks: jw}
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
