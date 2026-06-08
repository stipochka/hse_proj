package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	_ "edu-platform/docs" // generated swagger docs
	"edu-platform/internal/handlers"
	"edu-platform/internal/jwks"
	"edu-platform/internal/s3"
	"edu-platform/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

func New(ctx context.Context, pool *pgxpool.Pool) (http.Handler, error) {
	s3c, err := newS3Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("s3 init: %w", err)
	}
	jw, err := newJWKSClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("jwks init: %w", err)
	}
	return NewRouter(handlers.New(store.New(pool), s3c, jw)), nil
}

// NewRouter wires all routes. Exported so integration tests can build a server
// without env-var dependencies.
func NewRouter(h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// ── student ──────────────────────────────────────────────────────────────
	r.Post("/activities/upload-url", h.Auth(h.UploadURL))
	r.Post("/activities/{id}/confirm", h.Auth(h.Confirm))
	r.Get("/activities/my", h.Auth(h.ListMyActivities))
	r.Get("/dashboard/me", h.Auth(h.DashboardMe))
	r.Get("/export/me", h.Auth(h.ExportMe))

	// ── shared (student owner OR admin of the group) ───────────────────────────
	r.Get("/activities/{id}", h.Auth(h.GetActivity))
	r.Get("/activities/{id}/file", h.Auth(h.GetActivityFile))

	// ── admin (group_admin + super_admin) ──────────────────────────────────────
	r.Get("/activities", h.AuthAdmin(h.ListActivities))
	r.Post("/activities/{id}/evaluation", h.AuthAdmin(h.Evaluate))
	r.Get("/dashboard/summary", h.AuthAdmin(h.Summary))
	r.Get("/export/summary", h.AuthAdmin(h.ExportSummary))

	return r
}

func newJWKSClient(ctx context.Context) (*jwks.Client, error) {
	url := os.Getenv("KEYCLOAK_JWKS_URL")
	if url == "" {
		url = "http://keycloak:8080/realms/edu/protocol/openid-connect/certs"
	}
	return jwks.New(ctx, url)
}

func newS3Client(ctx context.Context) (*s3.Client, error) {
	endpoint := envOr("S3_ENDPOINT", "minio:9000")
	accessKey := envOr("S3_ACCESS_KEY", "minioadmin")
	secretKey := envOr("S3_SECRET_KEY", "minioadmin")
	bucket := envOr("S3_BUCKET", "edu-files")
	useSSL := false
	if v := os.Getenv("S3_USE_SSL"); v != "" {
		useSSL, _ = strconv.ParseBool(v)
	}
	c, err := s3.New(endpoint, accessKey, secretKey, bucket, useSSL)
	if err != nil {
		return nil, err
	}
	return c, c.EnsureBucket(ctx)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
