-- Replace the stale role constraint with the actual values used by Keycloak integration.
ALTER TABLE users DROP CONSTRAINT IF EXISTS valid_role;
ALTER TABLE users ADD CONSTRAINT valid_role
    CHECK (role IN ('student', 'group_admin', 'super_admin'));

-- Migrate any legacy values left over from the old schema.
UPDATE users SET role = 'group_admin' WHERE role = 'teacher';
UPDATE users SET role = 'super_admin'  WHERE role = 'admin';
