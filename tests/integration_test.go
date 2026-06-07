// Package tests contains end-to-end integration tests for the edu-platform service.
// Each test uses a real PostgreSQL database (via testcontainers or TEST_DATABASE_URL),
// an in-memory S3 mock, and a local JWKS server backed by a generated RSA key pair.
// Docker must be running on the host (or TEST_DATABASE_URL set) for these to execute.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"edu-platform/internal/handlers"
	"edu-platform/internal/jwks"
	"edu-platform/internal/server"
	"edu-platform/internal/store"
	"edu-platform/internal/testutil"
)

const (
	groupA = "BSE-2024" // student1 + admin
	groupB = "BSE-2023" // student2
)

var (
	apiURL  string
	jwksSrv *testutil.JWKSServer
	s3mock  *testutil.S3Mock

	tokenStudent1 string // group A
	tokenStudent2 string // group B
	tokenAdmin    string // group_admin of group A
	tokenSuper    string // super_admin, no group

	skipReason string
)

func TestMain(m *testing.M) {
	os.Exit(setupAndRun(context.Background(), m))
}

func setupAndRun(ctx context.Context, m *testing.M) int {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		c, d, err := startPostgres(ctx)
		if err != nil {
			skipReason = fmt.Sprintf("no Docker and TEST_DATABASE_URL unset (%v) — skipping", err)
			fmt.Fprintln(os.Stderr, skipReason)
			return m.Run()
		}
		defer c.Terminate(ctx) //nolint:errcheck
		dsn = d
	}

	pool, err := connectAndMigrate(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		return 1
	}
	defer pool.Close()

	jwksSrv, err = testutil.NewJWKSServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "jwks: %v\n", err)
		return 1
	}
	defer jwksSrv.Server.Close()

	tokenStudent1 = jwksSrv.Sign("kc-student1", "student1@edu.local", "student", groupA)
	tokenStudent2 = jwksSrv.Sign("kc-student2", "student2@edu.local", "student", groupB)
	tokenAdmin = jwksSrv.Sign("kc-admin1", "admin1@edu.local", "group_admin", groupA)
	tokenSuper = jwksSrv.Sign("kc-super", "super@edu.local", "super_admin")

	jw, err := jwks.New(ctx, jwksSrv.Server.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jwks client: %v\n", err)
		return 1
	}

	s3mock = &testutil.S3Mock{}
	h := handlers.New(store.New(pool), s3mock, jw)
	srv := httptest.NewServer(server.NewRouter(h))
	apiURL = srv.URL
	defer srv.Close()

	return m.Run()
}

// ── infra helpers ─────────────────────────────────────────────────────────────

func startPostgres(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres", "POSTGRES_DB": "testdb"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	if err != nil {
		return nil, "", err
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	return c, fmt.Sprintf("postgresql://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port()), nil
}

func connectAndMigrate(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir("../migrations")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join("../migrations", e.Name()))
		}
	}
	sort.Strings(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return nil, fmt.Errorf("exec %s: %w", f, err)
		}
	}
	return pool, nil
}

// ── request helpers ────────────────────────────────────────────────────────────

func requireDB(t *testing.T) {
	t.Helper()
	if skipReason != "" {
		t.Skip(skipReason)
	}
}

func api(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var br io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		br = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, apiURL+path, br)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decode(t *testing.T, r *http.Response, out any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(r.Body).Decode(out))
	r.Body.Close()
}

// submitActivity runs the full presigned flow (upload-url → simulate upload → confirm)
// and returns the new activity id.
func submitActivity(t *testing.T, token, title string) int64 {
	t.Helper()
	resp := api(t, http.MethodPost, "/activities/upload-url", token, map[string]any{"title": title, "category": "sport"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var up struct {
		ActivityID int64  `json:"activity_id"`
		UploadURL  string `json:"upload_url"`
		PDFKey     string `json:"pdf_key"`
	}
	decode(t, resp, &up)
	require.NotEmpty(t, up.UploadURL)
	require.NotEmpty(t, up.PDFKey)

	// Simulate the browser's direct PUT to S3.
	s3mock.SeedObject(up.PDFKey, 1024)

	conf := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/confirm", up.ActivityID), token, nil)
	require.Equal(t, http.StatusOK, conf.StatusCode)
	conf.Body.Close()
	return up.ActivityID
}

// ── tests ──────────────────────────────────────────────────────────────────────

func TestUnauthenticated(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodGet, "/activities/my", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestStudentForbiddenOnAdminFeed(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodGet, "/activities", tokenStudent1, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestConfirmRequiresUploadedFile(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodPost, "/activities/upload-url", tokenStudent1, map[string]any{"title": "no file yet"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var up struct {
		ActivityID int64 `json:"activity_id"`
	}
	decode(t, resp, &up)

	// confirm without seeding the object → 400
	c := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/confirm", up.ActivityID), tokenStudent1, nil)
	assert.Equal(t, http.StatusBadRequest, c.StatusCode)
	c.Body.Close()
}

func TestSubmitListAndGet(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "My activity")

	// appears in /activities/my as SUBMITTED
	resp := api(t, http.MethodGet, "/activities/my", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var mine []map[string]any
	decode(t, resp, &mine)
	found := false
	for _, a := range mine {
		if int64(a["id"].(float64)) == id {
			found = true
			assert.Equal(t, "SUBMITTED", a["status"])
		}
	}
	assert.True(t, found)

	// owner can GET details
	g := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", id), tokenStudent1, nil)
	assert.Equal(t, http.StatusOK, g.StatusCode)
	g.Body.Close()
}

func TestStudentCannotSeeOthersActivity(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "private")
	resp := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", id), tokenStudent2, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestFileURL(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "with file")
	resp := api(t, http.MethodGet, fmt.Sprintf("/activities/%d/file", id), tokenStudent1, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		FileURL string `json:"file_url"`
	}
	decode(t, resp, &out)
	assert.NotEmpty(t, out.FileURL)
}

func TestAdminFeedScopedToGroup(t *testing.T) {
	requireDB(t)
	a1 := submitActivity(t, tokenStudent1, "groupA activity")
	a2 := submitActivity(t, tokenStudent2, "groupB activity")

	resp := api(t, http.MethodGet, "/activities", tokenAdmin, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var feed []map[string]any
	decode(t, resp, &feed)

	var sawA1, sawA2 bool
	for _, a := range feed {
		switch int64(a["id"].(float64)) {
		case a1:
			sawA1 = true
		case a2:
			sawA2 = true
		}
	}
	assert.True(t, sawA1, "admin should see own-group activity")
	assert.False(t, sawA2, "admin must not see other-group activity")
}

func TestEvaluateAndDashboard(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "to evaluate")

	ev := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/evaluation", id), tokenAdmin,
		map[string]any{"points": 8, "comment": "great"})
	require.Equal(t, http.StatusCreated, ev.StatusCode)
	ev.Body.Close()

	// activity is now EVALUATED with embedded evaluation
	g := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", id), tokenStudent1, nil)
	var act map[string]any
	decode(t, g, &act)
	assert.Equal(t, "EVALUATED", act["status"])
	evObj := act["evaluation"].(map[string]any)
	assert.Equal(t, float64(8), evObj["points"])
	assert.InDelta(t, 3.2, evObj["credits"], 0.01) // 8 / 2.5

	// student dashboard reflects the points
	d := api(t, http.MethodGet, "/dashboard/me", tokenStudent1, nil)
	var dash map[string]any
	decode(t, d, &dash)
	assert.GreaterOrEqual(t, dash["total_points"].(float64), float64(8))
}

func TestEvaluateOtherGroupForbidden(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent2, "groupB to evaluate")
	resp := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/evaluation", id), tokenAdmin,
		map[string]any{"points": 5})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestReject(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "to reject")
	resp := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/evaluation", id), tokenAdmin,
		map[string]any{"reject": true, "comment": "insufficient"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	g := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", id), tokenStudent1, nil)
	var act map[string]any
	decode(t, g, &act)
	assert.Equal(t, "REJECTED", act["status"])
}

func TestScoreValidation(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "bad score")
	resp := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/evaluation", id), tokenAdmin,
		map[string]any{"points": 999})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestSummaryAndExport(t *testing.T) {
	requireDB(t)
	id := submitActivity(t, tokenStudent1, "for summary")
	ev := api(t, http.MethodPost, fmt.Sprintf("/activities/%d/evaluation", id), tokenAdmin, map[string]any{"points": 6})
	ev.Body.Close()

	// JSON summary (super_admin filters by group)
	resp := api(t, http.MethodGet, "/dashboard/summary?group="+groupA, tokenSuper, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var stats []map[string]any
	decode(t, resp, &stats)
	assert.NotEmpty(t, stats)

	// CSV export (group_admin, own group)
	csv := api(t, http.MethodGet, "/export/summary", tokenAdmin, nil)
	require.Equal(t, http.StatusOK, csv.StatusCode)
	assert.Contains(t, csv.Header.Get("Content-Type"), "text/csv")
	body, _ := io.ReadAll(csv.Body)
	csv.Body.Close()
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	assert.Equal(t, "student_id,group,points,credits,activities,evaluated", strings.TrimSpace(lines[0]))
}
