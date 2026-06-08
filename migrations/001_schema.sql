-- Clean schema per design-doc: only two tables. Users/groups/courses live in Keycloak.
-- Identity (student_id, admin_id) is the Keycloak subject (sub); student_group is a
-- snapshot of the student's Keycloak group taken at submission time.

CREATE TABLE IF NOT EXISTS activities (
    id            BIGSERIAL PRIMARY KEY,
    student_id    TEXT        NOT NULL,                 -- Keycloak user id (sub)
    student_group TEXT        NOT NULL DEFAULT '',      -- snapshot of group for reports
    title         TEXT        NOT NULL,
    category      TEXT        NOT NULL DEFAULT '',
    description   TEXT        NOT NULL DEFAULT '',
    pdf_key       TEXT        NOT NULL DEFAULT '',      -- object key in S3/MinIO
    status        TEXT        NOT NULL DEFAULT 'PENDING', -- PENDING|SUBMITTED|EVALUATED|REJECTED
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_activities_student ON activities (student_id);
CREATE INDEX IF NOT EXISTS idx_activities_group   ON activities (student_group);
CREATE INDEX IF NOT EXISTS idx_activities_status  ON activities (status);

CREATE TABLE IF NOT EXISTS evaluations (
    id           BIGSERIAL PRIMARY KEY,
    activity_id  BIGINT       NOT NULL REFERENCES activities (id) ON DELETE CASCADE,
    admin_id     TEXT         NOT NULL,                 -- Keycloak user id (sub) of evaluator
    points       INT          NOT NULL DEFAULT 0,
    credits      NUMERIC(6,2) NOT NULL DEFAULT 0,
    comment      TEXT         NOT NULL DEFAULT '',
    evaluated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (activity_id)                                 -- one current evaluation per activity
);

CREATE INDEX IF NOT EXISTS idx_evaluations_activity ON evaluations (activity_id);
