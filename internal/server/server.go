package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"edu-platform/internal/handlers"
	"edu-platform/internal/jwks"
	"edu-platform/internal/s3"
	"edu-platform/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	st := store.New(pool)
	return NewRouter(handlers.New(st, s3c, jw)), nil
}

// NewRouter wires all routes to the given handler and returns the chi router.
// Exported so integration tests can build a server without env-var dependencies.
func NewRouter(h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	// Student: own profile & stats
	r.Get("/me", h.Auth(h.Me))
	r.Get("/me/balance", h.Auth(h.MyBalance))
	r.Get("/me/transactions", h.Auth(h.MyTransactions))
	r.Get("/me/evaluations", h.Auth(h.MyEvaluations))

	// Student: activities
	r.Post("/activities", h.Auth(h.CreateActivity))
	r.Get("/activities", h.Auth(h.ListMyActivities))
	r.Get("/activities/{id}", h.Auth(h.GetActivity))
	r.Delete("/activities/{id}", h.Auth(h.DeleteActivity))

	// Files (PDF only)
	r.Post("/files", h.Auth(h.UploadFile))
	r.Get("/files/{id}", h.Auth(h.DownloadFile))

	// group_admin + super_admin: evaluate & activity feed
	r.Get("/admin/activities", h.AuthGroupAdmin(h.ListAdminActivities))
	r.Post("/evaluate", h.AuthGroupAdmin(h.Evaluate))

	// group_admin + super_admin: reports
	r.Get("/admin/reports", h.AuthGroupAdmin(h.AdminReports))

	// super_admin only: groups & courses management
	r.Get("/admin/groups", h.AuthSuperAdmin(h.ListGroups))
	r.Post("/admin/groups", h.AuthSuperAdmin(h.CreateGroup))
	r.Post("/admin/groups/assign", h.AuthSuperAdmin(h.AssignUserToGroup))
	r.Post("/admin/groups/remove", h.AuthSuperAdmin(h.RemoveUserFromGroup))
	r.Get("/admin/courses", h.AuthSuperAdmin(h.ListCourses))
	r.Post("/admin/courses", h.AuthSuperAdmin(h.CreateCourse))
	r.Post("/admin/courses/assign", h.AuthSuperAdmin(h.AssignUserToCourse))

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
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "minio:9000"
	}
	accessKey := os.Getenv("S3_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("S3_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "edu-files"
	}
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
