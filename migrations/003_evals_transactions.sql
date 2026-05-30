CREATE TABLE IF NOT EXISTS evaluations (
  id bigserial PRIMARY KEY,
  activity_id bigint NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
  evaluator_id bigint REFERENCES users(id) ON DELETE SET NULL,
  score integer,
  currency bigint,
  credits double precision,
  comment text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount bigint NOT NULL,
  reason text,
  created_at timestamptz NOT NULL DEFAULT now()
);
