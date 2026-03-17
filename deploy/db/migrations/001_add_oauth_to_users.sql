-- Migration: Add GitLab OAuth fields to app_user table
-- =====================

-- Add GitLab OAuth specific columns to app_user table
ALTER TABLE app_user ADD COLUMN gitlab_user_id INTEGER UNIQUE;
ALTER TABLE app_user ADD COLUMN name TEXT;
ALTER TABLE app_user ADD COLUMN avatar_url TEXT;
ALTER TABLE app_user ADD COLUMN gitlab_access_token TEXT;
ALTER TABLE app_user ADD COLUMN gitlab_refresh_token TEXT;
ALTER TABLE app_user ADD COLUMN token_expires_at DATETIME;
ALTER TABLE app_user ADD COLUMN last_login_at DATETIME;

-- Make hashed_password nullable for OAuth users
-- Note: SQLite doesn't support ALTER COLUMN, so we need to handle this in the application layer
-- OAuth users will have NULL hashed_password, manual token users will have their password

-- Create indexes for OAuth lookups
CREATE INDEX IF NOT EXISTS idx_user_gitlab_id ON app_user(gitlab_user_id);
CREATE INDEX IF NOT EXISTS idx_user_last_login ON app_user(last_login_at);

-- Add comments (SQLite doesn't support comments on columns, so documenting here)
-- gitlab_user_id: Unique GitLab user ID from OAuth
-- name: Full name from GitLab profile
-- avatar_url: Avatar image URL from GitLab
-- gitlab_access_token: OAuth access token (encrypted in application)
-- gitlab_refresh_token: OAuth refresh token (encrypted in application)
-- token_expires_at: Expiration time for access token
-- last_login_at: Track user login activity
