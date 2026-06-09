package domain

import "time"

// Activity statuses (minimal lifecycle from the design doc).
const (
	StatusPending   = "PENDING"   // created, awaiting file upload
	StatusSubmitted = "SUBMITTED" // file uploaded & confirmed, ready for review
	StatusEvaluated = "EVALUATED" // points awarded
	StatusRejected  = "REJECTED"  // rejected with a comment
)

// Activity is a student-submitted achievement. Author and group come from the
// Keycloak token; the group is snapshotted at submission time for reporting.
type Activity struct {
	ID           int64       `json:"id"`
	StudentID    string      `json:"student_id"`    // Keycloak sub
	StudentName  string      `json:"student_name"`  // preferred_username snapshot
	StudentGroup string      `json:"student_group"` // snapshot
	Title        string      `json:"title"`
	Category     string      `json:"category"`
	Description  string      `json:"description"`
	PDFKey       string      `json:"-"` // internal S3 key, never exposed
	Status       string      `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	Evaluation   *Evaluation `json:"evaluation,omitempty"`
}

// Evaluation is an admin's assessment of an activity.
type Evaluation struct {
	ID          int64     `json:"id"`
	ActivityID  int64     `json:"activity_id"`
	AdminID     string    `json:"admin_id"` // Keycloak sub of the evaluator
	Points      int       `json:"points"`
	Credits     float64   `json:"credits"`
	Comment     string    `json:"comment"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// StudentStats is one aggregate row of the admin summary, keyed by Keycloak id.
type StudentStats struct {
	StudentID      string  `json:"student_id"`
	StudentName    string  `json:"student_name"`
	StudentGroup   string  `json:"student_group"`
	TotalPoints    int64   `json:"total_points"`
	TotalCredits   float64 `json:"total_credits"`
	ActivityCount  int     `json:"activity_count"`
	EvaluatedCount int     `json:"evaluated_count"`
}

// DashboardMe is the personal aggregate for a student dashboard.
type DashboardMe struct {
	TotalPoints   int64          `json:"total_points"`
	TotalCredits  float64        `json:"total_credits"`
	ActivityCount int            `json:"activity_count"`
	ByStatus      map[string]int `json:"by_status"`
	ByCategory    map[string]int `json:"by_category"`
}
