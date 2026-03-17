-- Migration: Add user_id to project table for user isolation
-- =====================

-- Add user_id column to project table
ALTER TABLE project ADD COLUMN user_id TEXT;

-- Add foreign key constraint (SQLite handles this in application layer)
-- FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE

-- Create index for user_id lookups
CREATE INDEX IF NOT EXISTS idx_project_user_id ON project(user_id);

-- Composite index for user + realm queries
CREATE INDEX IF NOT EXISTS idx_project_user_realm ON project(user_id, realm_id);

-- Create index for soft deletes if deleted_at column exists
CREATE INDEX IF NOT EXISTS idx_project_deleted_at ON project(deleted_at);

-- Comments (documenting the isolation model):
-- user_id IS NULL AND realm_id IS NOT NULL: Shared team projects (visible to all in realm)
-- user_id IS NOT NULL: Personal projects (visible only to owner)
-- Realm-based filtering is already supported via existing realm_id column
