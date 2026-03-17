-- Migration: Create prompts table with user isolation
-- =====================

CREATE TABLE IF NOT EXISTS prompt (
    id TEXT PRIMARY KEY,
    user_id TEXT, -- NULL for global templates
    realm_id TEXT, -- NULL for global templates, NOT NULL for shared prompts
    name TEXT NOT NULL,
    description TEXT,
    system_prompt TEXT,
    user_prompt TEXT,
    tags TEXT, -- Comma-separated tags
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME, -- For soft deletes
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_prompt_user_id ON prompt(user_id);
CREATE INDEX IF NOT EXISTS idx_prompt_realm_id ON prompt(realm_id);
CREATE INDEX IF NOT EXISTS idx_prompt_name ON prompt(name);
CREATE INDEX IF NOT EXISTS idx_prompt_tags ON prompt(tags);
CREATE INDEX IF NOT EXISTS idx_prompt_deleted_at ON prompt(deleted_at);
CREATE INDEX IF NOT EXISTS idx_prompt_created_time ON prompt(created_time);

-- Composite index for user + realm queries
CREATE INDEX IF NOT EXISTS idx_prompt_user_realm ON prompt(user_id, realm_id);

-- Comments (documenting the isolation model):
-- user_id IS NULL AND realm_id IS NULL: Global templates (visible to all)
-- user_id IS NOT NULL AND realm_id IS NOT NULL: Personal prompts (visible only to owner)
-- user_id IS NULL AND realm_id IS NOT NULL: Shared team prompts (visible to all in realm)
