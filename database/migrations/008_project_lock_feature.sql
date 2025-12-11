-- Migration: 008_project_lock_feature.sql
-- Purpose: Add project locking and balance transfer capability

-- Add lock-related columns to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT false;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS lock_reason TEXT;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS locked_by UUID REFERENCES users(id);

-- Add index for quick lookup of locked projects
CREATE INDEX IF NOT EXISTS idx_projects_is_locked ON projects(is_locked);

-- Comment on columns
COMMENT ON COLUMN projects.is_locked IS 'Whether the project is locked for new transactions';
COMMENT ON COLUMN projects.locked_at IS 'Timestamp when project was locked';
COMMENT ON COLUMN projects.lock_reason IS 'Reason for locking the project';
COMMENT ON COLUMN projects.locked_by IS 'User who locked the project';
