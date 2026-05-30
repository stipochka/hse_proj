CREATE TABLE IF NOT EXISTS groups (
  id   bigserial PRIMARY KEY,
  name text NOT NULL UNIQUE,
  stream text,
  course_year integer
);

CREATE TABLE IF NOT EXISTS courses (
  id   bigserial PRIMARY KEY,
  name text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_groups (
  user_id  bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id bigint NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS user_courses (
  user_id   bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  course_id bigint NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, course_id)
);
