package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"edu-platform/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("not found")

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// ── Activities ──────────────────────────────────────────────────────────────

// CreateActivity inserts a PENDING activity with a pre-generated pdf_key.
func (s *Store) CreateActivity(ctx context.Context, a domain.Activity) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO activities (student_id, student_name, student_group, title, category, description, pdf_key, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		a.StudentID, a.StudentName, a.StudentGroup, a.Title, a.Category, a.Description, a.PDFKey, domain.StatusPending,
	).Scan(&id)
	return id, err
}

// ConfirmActivity moves a PENDING activity owned by studentID to SUBMITTED.
func (s *Store) ConfirmActivity(ctx context.Context, id int64, studentID string) error {
	ct, err := s.db.Exec(ctx,
		`UPDATE activities SET status=$1
		 WHERE id=$2 AND student_id=$3 AND status=$4`,
		domain.StatusSubmitted, id, studentID, domain.StatusPending,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetActivity loads one activity and its evaluation (if any).
func (s *Store) GetActivity(ctx context.Context, id int64) (*domain.Activity, error) {
	var a domain.Activity
	err := s.db.QueryRow(ctx,
		`SELECT id, student_id, student_name, student_group, title, category, description, pdf_key, status, created_at
		 FROM activities WHERE id=$1`, id,
	).Scan(&a.ID, &a.StudentID, &a.StudentName, &a.StudentGroup, &a.Title, &a.Category, &a.Description, &a.PDFKey, &a.Status, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	ev, err := s.getEvaluation(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Evaluation = ev
	return &a, nil
}

// ListMyActivities returns a student's own activities (newest first).
func (s *Store) ListMyActivities(ctx context.Context, studentID, status, category string) ([]domain.Activity, error) {
	where := []string{"a.student_id=$1"}
	args := []any{studentID}
	idx := 2
	if status != "" {
		where = append(where, fmt.Sprintf("a.status=$%d", idx))
		args = append(args, status)
		idx++
	}
	if category != "" {
		where = append(where, fmt.Sprintf("a.category=$%d", idx))
		args = append(args, category)
		idx++
	}
	return s.queryActivities(ctx, strings.Join(where, " AND "), args)
}

// ActivityFilter holds optional filters for the admin activity feed.
// Group is enforced for group_admin (set in the handler) and optional for super_admin.
type ActivityFilter struct {
	Group     string
	StudentID string
	Status    string
	Category  string
}

// ListActivities returns activities for the admin feed, scoped by filter.
func (s *Store) ListActivities(ctx context.Context, f ActivityFilter) ([]domain.Activity, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1
	add := func(cond string, val any) {
		where = append(where, fmt.Sprintf(cond, idx))
		args = append(args, val)
		idx++
	}
	if f.Group != "" {
		add("a.student_group=$%d", f.Group)
	}
	if f.StudentID != "" {
		add("a.student_id=$%d", f.StudentID)
	}
	if f.Status != "" {
		add("a.status=$%d", f.Status)
	}
	if f.Category != "" {
		add("a.category=$%d", f.Category)
	}
	return s.queryActivities(ctx, strings.Join(where, " AND "), args)
}

func (s *Store) queryActivities(ctx context.Context, where string, args []any) ([]domain.Activity, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.student_id, a.student_name, a.student_group, a.title, a.category, a.description, a.pdf_key, a.status, a.created_at,
		       e.id, e.activity_id, e.admin_id, e.points, e.credits, e.comment, e.evaluated_at
		FROM activities a
		LEFT JOIN evaluations e ON e.activity_id = a.id
		WHERE `+where+`
		ORDER BY a.created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []domain.Activity{}
	for rows.Next() {
		var a domain.Activity
		// Evaluation columns are nullable (LEFT JOIN), so scan into pointers.
		var (
			evID, evActID      *int64
			evAdmin, evComment *string
			evPoints           *int
			evCredits          *float64
			evAt               *time.Time
		)
		if err := rows.Scan(
			&a.ID, &a.StudentID, &a.StudentName, &a.StudentGroup, &a.Title, &a.Category, &a.Description, &a.PDFKey, &a.Status, &a.CreatedAt,
			&evID, &evActID, &evAdmin, &evPoints, &evCredits, &evComment, &evAt,
		); err != nil {
			return nil, err
		}
		if evID != nil {
			a.Evaluation = &domain.Evaluation{
				ID: *evID, ActivityID: *evActID, AdminID: *evAdmin,
				Points: *evPoints, Credits: *evCredits, Comment: *evComment, EvaluatedAt: *evAt,
			}
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

// DeleteActivity removes a PENDING activity owned by the student (used for orphan cleanup / cancel).
func (s *Store) DeleteActivity(ctx context.Context, id int64, studentID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM activities WHERE id=$1 AND student_id=$2`, id, studentID)
	return err
}

// ── Evaluations ─────────────────────────────────────────────────────────────

func (s *Store) getEvaluation(ctx context.Context, activityID int64) (*domain.Evaluation, error) {
	var e domain.Evaluation
	err := s.db.QueryRow(ctx,
		`SELECT id, activity_id, admin_id, points, credits, comment, evaluated_at
		 FROM evaluations WHERE activity_id=$1`, activityID,
	).Scan(&e.ID, &e.ActivityID, &e.AdminID, &e.Points, &e.Credits, &e.Comment, &e.EvaluatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpsertEvaluation writes (or replaces) the evaluation and sets the activity status
// in a single transaction. status must be EVALUATED or REJECTED.
func (s *Store) UpsertEvaluation(ctx context.Context, e domain.Evaluation, status string) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var id int64
	err = tx.QueryRow(ctx,
		`INSERT INTO evaluations (activity_id, admin_id, points, credits, comment)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (activity_id) DO UPDATE
		   SET admin_id=EXCLUDED.admin_id, points=EXCLUDED.points,
		       credits=EXCLUDED.credits, comment=EXCLUDED.comment, evaluated_at=now()
		 RETURNING id`,
		e.ActivityID, e.AdminID, e.Points, e.Credits, e.Comment,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `UPDATE activities SET status=$1 WHERE id=$2`, status, e.ActivityID); err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

// ── Aggregates ──────────────────────────────────────────────────────────────

// DashboardMe aggregates one student's own activities and awarded points.
func (s *Store) DashboardMe(ctx context.Context, studentID string) (domain.DashboardMe, error) {
	d := domain.DashboardMe{ByStatus: map[string]int{}, ByCategory: map[string]int{}}

	err := s.db.QueryRow(ctx, `
		SELECT
		  COALESCE(SUM(e.points),0),
		  COALESCE(SUM(e.credits),0),
		  COUNT(DISTINCT a.id)
		FROM activities a
		LEFT JOIN evaluations e ON e.activity_id = a.id
		WHERE a.student_id=$1`, studentID,
	).Scan(&d.TotalPoints, &d.TotalCredits, &d.ActivityCount)
	if err != nil {
		return d, err
	}

	rows, err := s.db.Query(ctx, `SELECT status, COUNT(*) FROM activities WHERE student_id=$1 GROUP BY status`, studentID)
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			rows.Close()
			return d, err
		}
		d.ByStatus[k] = n
	}
	rows.Close()

	rows, err = s.db.Query(ctx, `SELECT COALESCE(NULLIF(category,''),'uncategorized'), COUNT(*) FROM activities WHERE student_id=$1 GROUP BY 1`, studentID)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			return d, err
		}
		d.ByCategory[k] = n
	}
	return d, rows.Err()
}

// SummaryFilter holds optional filters for the admin summary.
type SummaryFilter struct {
	Group     string
	StudentID string
	Category  string
}

// Summary aggregates points/credits/counts per student, scoped by filter.
func (s *Store) Summary(ctx context.Context, f SummaryFilter) ([]domain.StudentStats, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1
	add := func(cond string, val any) {
		where = append(where, fmt.Sprintf(cond, idx))
		args = append(args, val)
		idx++
	}
	if f.Group != "" {
		add("a.student_group=$%d", f.Group)
	}
	if f.StudentID != "" {
		add("a.student_id=$%d", f.StudentID)
	}
	if f.Category != "" {
		add("a.category=$%d", f.Category)
	}

	rows, err := s.db.Query(ctx, `
		SELECT a.student_id,
		       COALESCE(MAX(a.student_name),'')  AS student_name,
		       COALESCE(MAX(a.student_group),'') AS student_group,
		       COALESCE(SUM(e.points),0)  AS total_points,
		       COALESCE(SUM(e.credits),0) AS total_credits,
		       COUNT(DISTINCT a.id)       AS activity_count,
		       COUNT(e.id)                AS evaluated_count
		FROM activities a
		LEFT JOIN evaluations e ON e.activity_id = a.id
		WHERE `+strings.Join(where, " AND ")+`
		GROUP BY a.student_id
		ORDER BY total_points DESC, a.student_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []domain.StudentStats{}
	for rows.Next() {
		var ss domain.StudentStats
		if err := rows.Scan(&ss.StudentID, &ss.StudentName, &ss.StudentGroup, &ss.TotalPoints, &ss.TotalCredits, &ss.ActivityCount, &ss.EvaluatedCount); err != nil {
			return nil, err
		}
		list = append(list, ss)
	}
	return list, rows.Err()
}
