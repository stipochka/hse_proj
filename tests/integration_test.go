// Package tests contains end-to-end integration tests for the edu-platform service.
// Each test uses a real PostgreSQL database (spun up via testcontainers), an
// in-memory S3 mock, and a local JWKS server backed by a generated RSA key pair.
// Docker must be running on the host for these tests to execute.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// ── shared test state ───────────────────────────────────────────────────────

var (
	apiURL string // base URL of the httptest server

	jwksSrv *testutil.JWKSServer

	// tokens for the four personas used across tests
	tokenStudent1   string // student in group1
	tokenStudent2   string // student in group2
	tokenGroupAdmin string // group_admin for group1
	tokenSuper      string // super_admin (no group)

	// DB IDs assigned at setup time
	group1ID int64
	group2ID int64
	// keycloak sub values — used to identify users across token / DB
	subStudent1   = "kc-student1"
	subStudent2   = "kc-student2"
	subGroupAdmin = "kc-admin1"
	subSuper      = "kc-super"

	// skipReason is non-empty when the DB is unavailable; all tests call requireDB.
	skipReason string
)

// ── TestMain: infrastructure setup / teardown ───────────────────────────────

func TestMain(m *testing.M) {
	ctx := context.Background()
	code := setupAndRun(ctx, m)
	os.Exit(code)
}

func setupAndRun(ctx context.Context, m *testing.M) int {
	// Resolve the PostgreSQL DSN.
	// Priority: TEST_DATABASE_URL env var → Docker container via testcontainers.
	dsn := os.Getenv("TEST_DATABASE_URL")
	var pgContainer testcontainers.Container
	if dsn == "" {
		var err error
		pgContainer, dsn, err = startPostgres(ctx)
		if err != nil {
			skipReason = fmt.Sprintf("no Docker available and TEST_DATABASE_URL is not set (%v) — skipping integration tests", err)
			fmt.Fprintln(os.Stderr, skipReason)
			return m.Run() // each test will call requireDB and t.Skip
		}
		defer pgContainer.Terminate(ctx) //nolint:errcheck
	}

	pool, err := connectAndMigrate(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		return 1
	}
	defer pool.Close()

	jwksSrv, err = testutil.NewJWKSServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "jwks server: %v\n", err)
		return 1
	}
	defer jwksSrv.Server.Close()

	group1ID, group2ID, err = seedGroups(ctx, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed groups: %v\n", err)
		return 1
	}
	if err = seedUsers(ctx, pool, group1ID, group2ID); err != nil {
		fmt.Fprintf(os.Stderr, "seed users: %v\n", err)
		return 1
	}

	// Build tokens; group_id is a string in JWT matching middleware parsing.
	g1Str := strconv.FormatInt(group1ID, 10)
	tokenStudent1 = jwksSrv.Sign(subStudent1, "student1@test.com", "student", "")
	tokenStudent2 = jwksSrv.Sign(subStudent2, "student2@test.com", "student", "")
	tokenGroupAdmin = jwksSrv.Sign(subGroupAdmin, "admin1@test.com", "group_admin", g1Str)
	tokenSuper = jwksSrv.Sign(subSuper, "super@test.com", "super_admin", "")

	jw, err := jwks.New(ctx, jwksSrv.Server.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jwks client: %v\n", err)
		return 1
	}

	s3mock := &testutil.S3Mock{}
	st := store.New(pool)
	h := handlers.New(st, s3mock, jw)

	srv := httptest.NewServer(server.NewRouter(h))
	apiURL = srv.URL
	defer srv.Close()

	return m.Run()
}

// ── helpers ─────────────────────────────────────────────────────────────────

func startPostgres(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}
	host, err := c.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	port, err := c.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", err
	}
	dsn := fmt.Sprintf("postgresql://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())
	return c, dsn, nil
}

func connectAndMigrate(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	// Find migrations relative to this test file's module root.
	migrationsDir := filepath.Join("..", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(sqlFiles)

	for _, f := range sqlFiles {
		sql, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return nil, fmt.Errorf("exec %s: %w", f, err)
		}
	}
	return pool, nil
}

func seedGroups(ctx context.Context, pool *pgxpool.Pool) (int64, int64, error) {
	var g1, g2 int64
	err := pool.QueryRow(ctx,
		`INSERT INTO groups (name, stream, course_year) VALUES ('Alpha', 'CS', 2024) RETURNING id`,
	).Scan(&g1)
	if err != nil {
		return 0, 0, err
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO groups (name, stream, course_year) VALUES ('Beta', 'CS', 2024) RETURNING id`,
	).Scan(&g2)
	return g1, g2, err
}

func seedUsers(ctx context.Context, pool *pgxpool.Pool, group1, group2 int64) error {
	type userRow struct {
		sub     string
		email   string
		role    string
		groupID int64
	}
	users := []userRow{
		{subStudent1, "student1@test.com", "student", group1},
		{subStudent2, "student2@test.com", "student", group2},
		{subGroupAdmin, "admin1@test.com", "group_admin", group1},
		{subSuper, "super@test.com", "super_admin", 0},
	}
	for _, u := range users {
		var uid int64
		err := pool.QueryRow(ctx,
			`INSERT INTO users (email, keycloak_id, role) VALUES ($1, $2, $3) RETURNING id`,
			u.email, u.sub, u.role,
		).Scan(&uid)
		if err != nil {
			return fmt.Errorf("insert user %s: %w", u.sub, err)
		}
		if u.groupID != 0 {
			if _, err := pool.Exec(ctx,
				`INSERT INTO user_groups (user_id, group_id) VALUES ($1, $2)`,
				uid, u.groupID,
			); err != nil {
				return fmt.Errorf("assign user %s to group: %w", u.sub, err)
			}
		}
	}
	return nil
}

// requireDB skips the test when no database was provisioned.
func requireDB(t *testing.T) {
	t.Helper()
	if skipReason != "" {
		t.Skip(skipReason)
	}
}

// api performs a JSON request and returns the response.
func api(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var bodyR io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyR = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, apiURL+path, bodyR)
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

func decodeJSON(t *testing.T, r io.Reader, out any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(r).Decode(out))
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestUnauthenticated(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodGet, "/me", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestStudentAccessDeniedToAdminRoutes(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodGet, "/admin/activities", tokenStudent1, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestStudentMe(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodGet, "/me", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	decodeJSON(t, resp.Body, &body)
	assert.Equal(t, "student1@test.com", body["email"])
	assert.Equal(t, "student", body["role"])
}

func TestStudentCreateListDeleteActivity(t *testing.T) {
	requireDB(t)
	// create
	resp := api(t, http.MethodPost, "/activities", tokenStudent1, map[string]any{
		"title":         "My first activity",
		"description":   "Integration test activity",
		"category":      "sport",
		"activity_date": "2024-03-15",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	decodeJSON(t, resp.Body, &created)
	idF, ok := created["id"].(float64)
	require.True(t, ok, "expected numeric id in response")
	actID := int64(idF)

	// list — must appear
	listResp := api(t, http.MethodGet, "/activities", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var list []map[string]any
	decodeJSON(t, listResp.Body, &list)
	found := false
	for _, a := range list {
		if int64(a["id"].(float64)) == actID {
			found = true
		}
	}
	assert.True(t, found, "created activity should appear in list")

	// get by id
	getResp := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", actID), tokenStudent1, nil)
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	// delete
	delResp := api(t, http.MethodDelete, fmt.Sprintf("/activities/%d", actID), tokenStudent1, nil)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	// should be gone now
	getResp2 := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", actID), tokenStudent1, nil)
	assert.Equal(t, http.StatusNotFound, getResp2.StatusCode)
}

func TestStudentCannotSeeOtherStudentActivity(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodPost, "/activities", tokenStudent1, map[string]any{
		"title": "Private activity",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	decodeJSON(t, resp.Body, &created)
	actID := int64(created["id"].(float64))
	defer api(t, http.MethodDelete, fmt.Sprintf("/activities/%d", actID), tokenStudent1, nil)

	resp2 := api(t, http.MethodGet, fmt.Sprintf("/activities/%d", actID), tokenStudent2, nil)
	assert.Equal(t, http.StatusForbidden, resp2.StatusCode)
}

func TestFileUploadPDFOnly(t *testing.T) {
	requireDB(t)
	// create activity first
	resp := api(t, http.MethodPost, "/activities", tokenStudent1, map[string]any{
		"title": "Activity with file",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	decodeJSON(t, resp.Body, &created)
	actID := int64(created["id"].(float64))
	defer api(t, http.MethodDelete, fmt.Sprintf("/activities/%d", actID), tokenStudent1, nil)

	upload := func(filename string, content []byte) *http.Response {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		_ = mw.WriteField("activity_id", strconv.FormatInt(actID, 10))
		fw, _ := mw.CreateFormFile("file", filename)
		fw.Write(content)
		mw.Close()

		req, _ := http.NewRequest(http.MethodPost, apiURL+"/files", &buf)
		req.Header.Set("Authorization", "Bearer "+tokenStudent1)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		r, _ := http.DefaultClient.Do(req)
		return r
	}

	// PDF is allowed
	pdfResp := upload("doc.pdf", []byte("%PDF-1.4 test content"))
	assert.Equal(t, http.StatusCreated, pdfResp.StatusCode)

	// non-PDF is rejected
	exeResp := upload("evil.exe", []byte("MZ"))
	assert.Equal(t, http.StatusBadRequest, exeResp.StatusCode)
}

func TestGroupAdminEvaluateAndStudentBalance(t *testing.T) {
	requireDB(t)
	// student1 submits an activity
	resp := api(t, http.MethodPost, "/activities", tokenStudent1, map[string]any{
		"title": "Activity to evaluate",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	decodeJSON(t, resp.Body, &created)
	actID := int64(created["id"].(float64))

	// get student1's DB id from /me
	meResp := api(t, http.MethodGet, "/me", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, meResp.StatusCode)
	var me map[string]any
	decodeJSON(t, meResp.Body, &me)
	studentID := int64(me["user_id"].(float64))

	// balance before evaluation
	balBefore := api(t, http.MethodGet, "/me/balance", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, balBefore.StatusCode)
	var balBody map[string]any
	decodeJSON(t, balBefore.Body, &balBody)
	before := int64(balBody["balance"].(float64))

	// group_admin evaluates
	evalResp := api(t, http.MethodPost, "/evaluate", tokenGroupAdmin, map[string]any{
		"activity_id": actID,
		"student_id":  studentID,
		"score":       8,
		"comment":     "Great work",
	})
	require.Equal(t, http.StatusCreated, evalResp.StatusCode)

	// balance after: currency = score * 10 = 80
	balAfter := api(t, http.MethodGet, "/me/balance", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, balAfter.StatusCode)
	var balAfterBody map[string]any
	decodeJSON(t, balAfter.Body, &balAfterBody)
	after := int64(balAfterBody["balance"].(float64))
	assert.Equal(t, before+80, after, "balance should increase by score*10")

	// student can see their evaluation
	evalsResp := api(t, http.MethodGet, "/me/evaluations", tokenStudent1, nil)
	require.Equal(t, http.StatusOK, evalsResp.StatusCode)
	var evals []map[string]any
	decodeJSON(t, evalsResp.Body, &evals)
	found := false
	for _, ev := range evals {
		if int64(ev["activity_id"].(float64)) == actID {
			found = true
			assert.InDelta(t, 3.2, ev["credits"].(float64), 0.01, "credits = score/2.5")
		}
	}
	assert.True(t, found, "evaluation should appear in student's list")
}

func TestGroupAdminCannotEvaluateOtherGroup(t *testing.T) {
	requireDB(t)
	// student2 is in group2; group_admin is in group1 — should be forbidden
	resp := api(t, http.MethodPost, "/activities", tokenStudent2, map[string]any{
		"title": "Activity in group2",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	decodeJSON(t, resp.Body, &created)
	actID := int64(created["id"].(float64))
	defer api(t, http.MethodDelete, fmt.Sprintf("/activities/%d", actID), tokenStudent2, nil)

	// get student2's id
	meResp := api(t, http.MethodGet, "/me", tokenStudent2, nil)
	require.Equal(t, http.StatusOK, meResp.StatusCode)
	var me map[string]any
	decodeJSON(t, meResp.Body, &me)
	student2ID := int64(me["user_id"].(float64))

	evalResp := api(t, http.MethodPost, "/evaluate", tokenGroupAdmin, map[string]any{
		"activity_id": actID,
		"student_id":  student2ID,
		"score":       5,
		"comment":     "Should fail",
	})
	assert.Equal(t, http.StatusForbidden, evalResp.StatusCode)
}

func TestGroupAdminActivityFeedScopedToGroup(t *testing.T) {
	requireDB(t)
	// student1 (group1) submits
	resp1 := api(t, http.MethodPost, "/activities", tokenStudent1, map[string]any{
		"title": "Group1 activity",
	})
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	// student2 (group2) submits
	resp2 := api(t, http.MethodPost, "/activities", tokenStudent2, map[string]any{
		"title": "Group2 activity",
	})
	require.Equal(t, http.StatusCreated, resp2.StatusCode)
	var cr2 map[string]any
	decodeJSON(t, resp2.Body, &cr2)
	act2ID := int64(cr2["id"].(float64))

	// group_admin (group1) sees the feed
	feedResp := api(t, http.MethodGet, "/admin/activities", tokenGroupAdmin, nil)
	require.Equal(t, http.StatusOK, feedResp.StatusCode)
	var feed []map[string]any
	decodeJSON(t, feedResp.Body, &feed)

	// group2's activity must not appear
	for _, a := range feed {
		assert.NotEqual(t, act2ID, int64(a["id"].(float64)),
			"group_admin should not see activities from other groups")
	}
}

func TestAdminReportsJSONAndCSV(t *testing.T) {
	requireDB(t)
	// JSON format
	jsonResp := api(t, http.MethodGet,
		fmt.Sprintf("/admin/reports?group_id=%d", group1ID), tokenSuper, nil)
	require.Equal(t, http.StatusOK, jsonResp.StatusCode)
	assert.Contains(t, jsonResp.Header.Get("Content-Type"), "json")
	var stats []map[string]any
	decodeJSON(t, jsonResp.Body, &stats)
	// At least one student in group1 (student1 was evaluated above)
	assert.NotEmpty(t, stats)

	// CSV format
	csvResp := api(t, http.MethodGet,
		fmt.Sprintf("/admin/reports?group_id=%d&format=csv", group1ID), tokenSuper, nil)
	require.Equal(t, http.StatusOK, csvResp.StatusCode)
	assert.Contains(t, csvResp.Header.Get("Content-Type"), "text/csv")

	csvBody, err := io.ReadAll(csvResp.Body)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(csvBody)), "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "CSV should have a header row plus at least one data row")
	assert.Equal(t, "user_id,email,group,currency,credits,activities", strings.TrimSpace(lines[0]))
}

func TestSuperAdminGroupManagement(t *testing.T) {
	requireDB(t)
	// Create a new group
	createResp := api(t, http.MethodPost, "/admin/groups", tokenSuper, map[string]any{
		"name":        "Gamma",
		"stream":      "Math",
		"course_year": 2025,
	})
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	var created map[string]any
	decodeJSON(t, createResp.Body, &created)
	newGroupID := int64(created["id"].(float64))
	assert.Greater(t, newGroupID, int64(0))

	// It must appear in the list
	listResp := api(t, http.MethodGet, "/admin/groups", tokenSuper, nil)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var groups []map[string]any
	decodeJSON(t, listResp.Body, &groups)
	found := false
	for _, g := range groups {
		if int64(g["id"].(float64)) == newGroupID {
			found = true
		}
	}
	assert.True(t, found, "newly created group should appear in list")

	// student-level token must be denied
	denyResp := api(t, http.MethodPost, "/admin/groups", tokenStudent1, map[string]any{
		"name": "ShouldFail",
	})
	assert.Equal(t, http.StatusForbidden, denyResp.StatusCode)
}

func TestScoreValidation(t *testing.T) {
	requireDB(t)
	resp := api(t, http.MethodPost, "/evaluate", tokenGroupAdmin, map[string]any{
		"activity_id": 1,
		"student_id":  1,
		"score":       999,
		"comment":     "out of range",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
