package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	return pool, nil
}

// Migrate runs schema migrations idempotently (all statements use IF NOT EXISTS).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS activities (
    id            BIGSERIAL PRIMARY KEY,
    student_id    TEXT        NOT NULL,
    student_group TEXT        NOT NULL DEFAULT '',
    title         TEXT        NOT NULL,
    category      TEXT        NOT NULL DEFAULT '',
    description   TEXT        NOT NULL DEFAULT '',
    pdf_key       TEXT        NOT NULL DEFAULT '',
    status        TEXT        NOT NULL DEFAULT 'PENDING',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_activities_student ON activities (student_id);
CREATE INDEX IF NOT EXISTS idx_activities_group   ON activities (student_group);
CREATE INDEX IF NOT EXISTS idx_activities_status  ON activities (status);

CREATE TABLE IF NOT EXISTS evaluations (
    id           BIGSERIAL PRIMARY KEY,
    activity_id  BIGINT       NOT NULL REFERENCES activities (id) ON DELETE CASCADE,
    admin_id     TEXT         NOT NULL,
    points       INT          NOT NULL DEFAULT 0,
    credits      NUMERIC(6,2) NOT NULL DEFAULT 0,
    comment      TEXT         NOT NULL DEFAULT '',
    evaluated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (activity_id)
);

CREATE INDEX IF NOT EXISTS idx_evaluations_activity ON evaluations (activity_id);
`
