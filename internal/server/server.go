package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"edu-platform/internal/handlers"
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

	st := store.New(pool)
	h := handlers.New(st, s3c)
	r := chi.NewRouter()

	// Public
	r.Post("/signup", h.SignUp)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Auth(h.Logout))

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

	// Files
	r.Post("/files", h.Auth(h.UploadFile))
	r.Get("/files/{id}", h.Auth(h.DownloadFile))

	// Teacher / admin: evaluate
	r.Post("/evaluate", h.AuthTeacher(h.Evaluate))

	// Admin: reports
	r.Get("/admin/reports", h.AuthAdmin(h.AdminReports))

	// Admin: groups
	r.Get("/admin/groups", h.AuthAdmin(h.ListGroups))
	r.Post("/admin/groups", h.AuthAdmin(h.CreateGroup))
	r.Post("/admin/groups/assign", h.AuthAdmin(h.AssignUserToGroup))
	r.Post("/admin/groups/remove", h.AuthAdmin(h.RemoveUserFromGroup))

	// Admin: courses
	r.Get("/admin/courses", h.AuthAdmin(h.ListCourses))
	r.Post("/admin/courses", h.AuthAdmin(h.CreateCourse))
	r.Post("/admin/courses/assign", h.AuthAdmin(h.AssignUserToCourse))

	return r, nil
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
