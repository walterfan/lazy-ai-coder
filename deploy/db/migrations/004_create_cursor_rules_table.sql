-- Migration: Create cursor_rules table with user isolation
-- =====================

CREATE TABLE IF NOT EXISTS cursor_rule (
    id TEXT PRIMARY KEY,
    user_id TEXT, -- NULL for global templates
    realm_id TEXT, -- NULL for global templates, NOT NULL for shared rules
    name TEXT NOT NULL,
    description TEXT,
    content TEXT, -- Full .cursorrules content
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
CREATE INDEX IF NOT EXISTS idx_cursor_rule_user_id ON cursor_rule(user_id);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_realm_id ON cursor_rule(realm_id);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_name ON cursor_rule(name);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_language ON cursor_rule(language);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_framework ON cursor_rule(framework);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_tags ON cursor_rule(tags);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_is_template ON cursor_rule(is_template);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_deleted_at ON cursor_rule(deleted_at);
CREATE INDEX IF NOT EXISTS idx_cursor_rule_created_time ON cursor_rule(created_time);

-- Composite index for user + realm queries
CREATE INDEX IF NOT EXISTS idx_cursor_rule_user_realm ON cursor_rule(user_id, realm_id);

-- Comments (documenting the isolation model):
-- user_id IS NULL AND realm_id IS NULL: Global templates (visible to all)
-- user_id IS NOT NULL AND realm_id IS NOT NULL: Personal rules (visible only to owner)
-- user_id IS NULL AND realm_id IS NOT NULL: Shared team rules (visible to all in realm)

