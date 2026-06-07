-- Переход на Keycloak: пароли хранятся там, а не у нас
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT NULL;
UPDATE users SET password_hash = NULL WHERE password_hash = '';

-- Keycloak UUID для сопоставления с нашим internal ID
ALTER TABLE users ADD COLUMN IF NOT EXISTS keycloak_id text UNIQUE;

-- Refresh-токены больше не нужны — Keycloak управляет сессиями
DROP TABLE IF EXISTS refresh_tokens;
