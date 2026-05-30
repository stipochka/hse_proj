ALTER TABLE activities
  ADD COLUMN IF NOT EXISTS category text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS status   text NOT NULL DEFAULT 'submitted'
    CHECK (status IN ('draft','submitted','under_review','approved','rejected')),
  ADD COLUMN IF NOT EXISTS activity_date date;

CREATE TABLE IF NOT EXISTS activity_files (
  id          bigserial PRIMARY KEY,
  activity_id bigint NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
  filename    text NOT NULL,
  stored_name text NOT NULL,
  size_bytes  bigint NOT NULL DEFAULT 0,
  created_at  timestamptz NOT NULL DEFAULT now()
);
