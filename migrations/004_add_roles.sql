-- Add role column to users (default: student)
ALTER TABLE users ADD COLUMN role text NOT NULL DEFAULT 'student';

-- Add constraint to ensure valid roles
ALTER TABLE users ADD CONSTRAINT valid_role CHECK (role IN ('student', 'teacher', 'admin'));
