package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // "student", "teacher", "admin"
	CreatedAt time.Time `json:"created_at"`
}

type Activity struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Category     string     `json:"category"`
	Status       string     `json:"status"` // draft, submitted, under_review, approved, rejected
	ActivityDate *time.Time `json:"activity_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	Files        []ActivityFile `json:"files,omitempty"`
}

type ActivityFile struct {
	ID         int64     `json:"id"`
	ActivityID int64     `json:"activity_id"`
	Filename   string    `json:"filename"`
	S3Key      string    `json:"-"`
	SizeBytes  int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
}

type Evaluation struct {
	ID          int64     `json:"id"`
	ActivityID  int64     `json:"activity_id"`
	EvaluatorID int64     `json:"evaluator_id"`
	Score       int       `json:"score"`
	Currency    int64     `json:"currency"`
	Credits     float64   `json:"credits"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

type Transaction struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    int64     `json:"amount"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type Course struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Group struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Stream     string `json:"stream"`
	CourseYear int    `json:"course_year"`
}

// StudentStats — агрегат для отчётов администратора
type StudentStats struct {
	UserID        int64   `json:"user_id"`
	Email         string  `json:"email"`
	GroupName     string  `json:"group_name"`
	TotalCurrency int64   `json:"total_currency"`
	TotalCredits  float64 `json:"total_credits"`
	ActivityCount int     `json:"activity_count"`
}
