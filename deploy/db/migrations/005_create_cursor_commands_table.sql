-- Migration: Create cursor_commands table with user isolation
-- =====================

CREATE TABLE IF NOT EXISTS cursor_command (
    id TEXT PRIMARY KEY,
    user_id TEXT, -- NULL for global templates
    realm_id TEXT, -- NULL for global templates, NOT NULL for shared commands
    name TEXT NOT NULL,
    description TEXT,
    command TEXT, -- The actual command/prompt text
    category TEXT, -- e.g., "refactor", "debug", "generate", "review"
    language TEXT, -- e.g., "go", "typescript", "general"
    framework TEXT, -- e.g., "gin", "vue", "general"
    tags TEXT, -- Comma-separated tags
    is_template INTEGER DEFAULT 0, -- Template for generation (0=false, 1=true)
    usage_count INTEGER DEFAULT 0, -- Track popularity
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME, -- For soft deletes
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_cursor_command_user_id ON cursor_command(user_id);
CREATE INDEX IF NOT EXISTS idx_cursor_command_realm_id ON cursor_command(realm_id);
CREATE INDEX IF NOT EXISTS idx_cursor_command_name ON cursor_command(name);
CREATE INDEX IF NOT EXISTS idx_cursor_command_category ON cursor_command(category);
CREATE INDEX IF NOT EXISTS idx_cursor_command_language ON cursor_command(language);
CREATE INDEX IF NOT EXISTS idx_cursor_command_framework ON cursor_command(framework);
CREATE INDEX IF NOT EXISTS idx_cursor_command_tags ON cursor_command(tags);
CREATE INDEX IF NOT EXISTS idx_cursor_command_is_template ON cursor_command(is_template);
CREATE INDEX IF NOT EXISTS idx_cursor_command_deleted_at ON cursor_command(deleted_at);
CREATE INDEX IF NOT EXISTS idx_cursor_command_created_time ON cursor_command(created_time);

-- Composite index for user + realm queries
CREATE INDEX IF NOT EXISTS idx_cursor_command_user_realm ON cursor_command(user_id, realm_id);

-- Comments (documenting the isolation model):
-- user_id IS NULL AND realm_id IS NULL: Global templates (visible to all)
-- user_id IS NOT NULL AND realm_id IS NOT NULL: Personal commands (visible only to owner)
-- user_id IS NULL AND realm_id IS NOT NULL: Shared team commands (visible to all in realm)

