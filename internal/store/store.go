package store

import (
	"context"
	"fmt"
	"strings"

	"edu-platform/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// ── Users ─────────────────────────────────────────────────────────────────────

// GetOrCreateUser upserts the user record on every authenticated request (JIT provisioning).
// Email and role are kept in sync with Keycloak on each call.
func (s *Store) GetOrCreateUser(ctx context.Context, keycloakID, email, role string) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (keycloak_id, email, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (keycloak_id) DO UPDATE SET email = EXCLUDED.email, role = EXCLUDED.role
		 RETURNING id`,
		keycloakID, email, role,
	).Scan(&id)
	return id, err
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, role, created_at FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ── Activities ────────────────────────────────────────────────────────────────

func (s *Store) CreateActivity(ctx context.Context, a domain.Activity) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO activities (user_id, title, description, category, status, activity_date)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		a.UserID, a.Title, a.Description, a.Category, a.Status, a.ActivityDate,
	).Scan(&id)
	return id, err
}

func (s *Store) GetActivity(ctx context.Context, id int64) (*domain.Activity, error) {
	var a domain.Activity
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, title, description, category, status, activity_date, created_at
		 FROM activities WHERE id=$1`, id,
	).Scan(&a.ID, &a.UserID, &a.Title, &a.Description, &a.Category, &a.Status, &a.ActivityDate, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	files, _ := s.GetActivityFiles(ctx, id)
	a.Files = files
	return &a, nil
}

func (s *Store) ListActivities(ctx context.Context, userID int64) ([]domain.Activity, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, title, description, category, status, activity_date, created_at
		 FROM activities WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Activity
	for rows.Next() {
		var a domain.Activity
		if err := rows.Scan(&a.ID, &a.UserID, &a.Title, &a.Description, &a.Category, &a.Status, &a.ActivityDate, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) UpdateActivityStatus(ctx context.Context, id int64, status string) error {
	_, err := s.db.Exec(ctx, `UPDATE activities SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (s *Store) DeleteActivity(ctx context.Context, id, ownerID int64) error {
	_, err := s.db.Exec(ctx, `DELETE FROM activities WHERE id=$1 AND user_id=$2`, id, ownerID)
	return err
}

// ── Activity files ────────────────────────────────────────────────────────────

func (s *Store) CreateActivityFile(ctx context.Context, f domain.ActivityFile) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO activity_files (activity_id, filename, s3_key, size_bytes)
		 VALUES ($1,$2,$3,$4) RETURNING id`,
		f.ActivityID, f.Filename, f.S3Key, f.SizeBytes,
	).Scan(&id)
	return id, err
}

func (s *Store) GetActivityFile(ctx context.Context, id int64) (*domain.ActivityFile, error) {
	var f domain.ActivityFile
	err := s.db.QueryRow(ctx,
		`SELECT id, activity_id, filename, s3_key, size_bytes, created_at
		 FROM activity_files WHERE id=$1`, id,
	).Scan(&f.ID, &f.ActivityID, &f.Filename, &f.S3Key, &f.SizeBytes, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Store) GetActivityFiles(ctx context.Context, activityID int64) ([]domain.ActivityFile, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, activity_id, filename, s3_key, size_bytes, created_at
		 FROM activity_files WHERE activity_id=$1 ORDER BY created_at`, activityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.ActivityFile
	for rows.Next() {
		var f domain.ActivityFile
		if err := rows.Scan(&f.ID, &f.ActivityID, &f.Filename, &f.S3Key, &f.SizeBytes, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

// ── Evaluations ───────────────────────────────────────────────────────────────

func (s *Store) CreateEvaluation(ctx context.Context, e domain.Evaluation) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO evaluations (activity_id, evaluator_id, score, currency, credits, comment)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		e.ActivityID, e.EvaluatorID, e.Score, e.Currency, e.Credits, e.Comment,
	).Scan(&id)
	return id, err
}

func (s *Store) ListEvaluationsByUser(ctx context.Context, userID int64) ([]domain.Evaluation, error) {
	rows, err := s.db.Query(ctx,
		`SELECT e.id, e.activity_id, e.evaluator_id, e.score, e.currency, e.credits, e.comment, e.created_at
		 FROM evaluations e
		 JOIN activities a ON a.id = e.activity_id
		 WHERE a.user_id=$1
		 ORDER BY e.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Evaluation
	for rows.Next() {
		var ev domain.Evaluation
		if err := rows.Scan(&ev.ID, &ev.ActivityID, &ev.EvaluatorID, &ev.Score, &ev.Currency, &ev.Credits, &ev.Comment, &ev.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, ev)
	}
	return list, rows.Err()
}

// ── Transactions ──────────────────────────────────────────────────────────────

func (s *Store) AddTransaction(ctx context.Context, t domain.Transaction) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO transactions (user_id, amount, reason) VALUES ($1,$2,$3) RETURNING id`,
		t.UserID, t.Amount, t.Reason,
	).Scan(&id)
	return id, err
}

func (s *Store) GetBalance(ctx context.Context, userID int64) (int64, error) {
	var balance int64
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM transactions WHERE user_id=$1`, userID,
	).Scan(&balance)
	return balance, err
}

func (s *Store) ListTransactions(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, amount, reason, created_at FROM transactions
		 WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Reason, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// ── Groups ────────────────────────────────────────────────────────────────────

func (s *Store) CreateGroup(ctx context.Context, name, stream string, courseYear int) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO groups (name, stream, course_year) VALUES ($1,$2,$3) RETURNING id`,
		name, stream, courseYear,
	).Scan(&id)
	return id, err
}

func (s *Store) ListGroups(ctx context.Context) ([]domain.Group, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, COALESCE(stream,''), COALESCE(course_year,0) FROM groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Group
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Stream, &g.CourseYear); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

func (s *Store) AssignUserToGroup(ctx context.Context, userID, groupID int64) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO user_groups (user_id, group_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		userID, groupID,
	)
	return err
}

func (s *Store) RemoveUserFromGroup(ctx context.Context, userID, groupID int64) error {
	_, err := s.db.Exec(ctx, `DELETE FROM user_groups WHERE user_id=$1 AND group_id=$2`, userID, groupID)
	return err
}

// IsUserInGroup returns true when the given user belongs to the given group.
func (s *Store) IsUserInGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_groups WHERE user_id=$1 AND group_id=$2)`,
		userID, groupID,
	).Scan(&exists)
	return exists, err
}

// ── Courses ───────────────────────────────────────────────────────────────────

func (s *Store) CreateCourse(ctx context.Context, name string) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO courses (name) VALUES ($1) RETURNING id`, name,
	).Scan(&id)
	return id, err
}

func (s *Store) ListCourses(ctx context.Context) ([]domain.Course, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name FROM courses ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Course
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *Store) AssignUserToCourse(ctx context.Context, userID, courseID int64) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO user_courses (user_id, course_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		userID, courseID,
	)
	return err
}

// ── Admin activities feed ─────────────────────────────────────────────────────

// ActivityFilter holds optional filters for the admin activity feed.
// GroupID is mandatory for group_admin (enforced in the handler) and optional for super_admin.
type ActivityFilter struct {
	GroupID   int64
	StudentID int64
	Status    string
	Category  string
	Limit     int
	Offset    int
}

func (s *Store) ListActivitiesAdmin(ctx context.Context, f ActivityFilter) ([]domain.Activity, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1

	if f.GroupID != 0 {
		where = append(where, fmt.Sprintf(
			"a.user_id IN (SELECT user_id FROM user_groups WHERE group_id=$%d)", idx))
		args = append(args, f.GroupID)
		idx++
	}
	if f.StudentID != 0 {
		where = append(where, fmt.Sprintf("a.user_id=$%d", idx))
		args = append(args, f.StudentID)
		idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("a.status=$%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Category != "" {
		where = append(where, fmt.Sprintf("a.category=$%d", idx))
		args = append(args, f.Category)
		idx++
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	q := fmt.Sprintf(`
		SELECT a.id, a.user_id, a.title, a.description, a.category, a.status, a.activity_date, a.created_at
		FROM activities a
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT %d OFFSET $%d`,
		strings.Join(where, " AND "), limit, idx)
	args = append(args, f.Offset)

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Activity
	for rows.Next() {
		var a domain.Activity
		if err := rows.Scan(&a.ID, &a.UserID, &a.Title, &a.Description, &a.Category, &a.Status, &a.ActivityDate, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

// ── Admin reports ─────────────────────────────────────────────────────────────

// ReportFilter holds optional filters for the aggregate student report.
type ReportFilter struct {
	UserID   int64
	GroupID  int64
	CourseID int64
	Stream   string
}

func (s *Store) AdminReport(ctx context.Context, f ReportFilter) ([]domain.StudentStats, error) {
	where := []string{"u.role = 'student'"}
	args := []any{}
	idx := 1

	if f.UserID != 0 {
		where = append(where, fmt.Sprintf("u.id = $%d", idx))
		args = append(args, f.UserID)
		idx++
	}
	if f.GroupID != 0 {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM user_groups ug WHERE ug.user_id=u.id AND ug.group_id=$%d)", idx))
		args = append(args, f.GroupID)
		idx++
	}
	if f.CourseID != 0 {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM user_courses uc WHERE uc.user_id=u.id AND uc.course_id=$%d)", idx))
		args = append(args, f.CourseID)
		idx++
	}
	if f.Stream != "" {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM user_groups ug2 JOIN groups g ON g.id=ug2.group_id WHERE ug2.user_id=u.id AND g.stream=$%d)", idx))
		args = append(args, f.Stream)
		idx++
	}

	q := fmt.Sprintf(`
		SELECT
		  u.id,
		  u.email,
		  COALESCE((SELECT g.name FROM user_groups ug JOIN groups g ON g.id=ug.group_id WHERE ug.user_id=u.id LIMIT 1), '') AS group_name,
		  COALESCE(SUM(t.amount), 0) AS total_currency,
		  COALESCE((SELECT SUM(ev.credits) FROM evaluations ev JOIN activities a ON a.id=ev.activity_id WHERE a.user_id=u.id), 0) AS total_credits,
		  COUNT(DISTINCT act.id) AS activity_count
		FROM users u
		LEFT JOIN transactions t ON t.user_id = u.id
		LEFT JOIN activities act ON act.user_id = u.id
		WHERE %s
		GROUP BY u.id, u.email
		ORDER BY u.email`, strings.Join(where, " AND "))

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.StudentStats
	for rows.Next() {
		var ss domain.StudentStats
		if err := rows.Scan(&ss.UserID, &ss.Email, &ss.GroupName, &ss.TotalCurrency, &ss.TotalCredits, &ss.ActivityCount); err != nil {
			return nil, err
		}
		list = append(list, ss)
	}
	return list, rows.Err()
}
