-- Migration: Add Role-Based Access Control and User Approval Workflow
-- =====================================================================

-- Create role table
CREATE TABLE IF NOT EXISTS role (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    UNIQUE(realm_id, name)
);

-- Create user_role join table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_role (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    UNIQUE(user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);

-- Create indexes for role lookups
CREATE INDEX IF NOT EXISTS idx_role_realm_id ON role(realm_id);
CREATE INDEX IF NOT EXISTS idx_role_name ON role(name);
CREATE INDEX IF NOT EXISTS idx_user_role_user_id ON user_role(user_id);
CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON user_role(role_id);

-- Modify app_user table: is_active now defaults to false (requires admin approval)
-- Note: SQLite doesn't support ALTER COLUMN DEFAULT, so this affects new users only
-- Existing users will remain active. Application layer handles the default for new users.

-- Create index on is_active for filtering pending users
CREATE INDEX IF NOT EXISTS idx_user_is_active ON app_user(is_active);
CREATE INDEX IF NOT EXISTS idx_user_realm_id ON app_user(realm_id);

-- Seed system-level realm for super admins (special realm that spans all realms)
INSERT INTO realm (id, name, description, created_by, created_time, updated_by, updated_time)
VALUES ('system', 'System', 'System realm for super administrators', 'system', CURRENT_TIMESTAMP, 'system', CURRENT_TIMESTAMP)
ON CONFLICT(id) DO NOTHING;

-- Seed default roles in system realm
INSERT INTO role (id, realm_id, name, description, created_by, created_time, updated_by, updated_time)
VALUES
    ('role_super_admin', 'system', 'super_admin', 'Super administrator with full system access across all realms', 'system', CURRENT_TIMESTAMP, 'system', CURRENT_TIMESTAMP),
    ('role_admin', 'system', 'admin', 'Administrator with full access within their own realm', 'system', CURRENT_TIMESTAMP, 'system', CURRENT_TIMESTAMP),
    ('role_user', 'system', 'user', 'Regular user with standard permissions', 'system', CURRENT_TIMESTAMP, 'system', CURRENT_TIMESTAMP)
ON CONFLICT(id) DO NOTHING;

-- Comments (documenting fields since SQLite doesn't support column comments)
-- role.id: Unique role identifier
-- role.realm_id: Realm this role belongs to ('system' for global roles)
-- role.name: Role name (super_admin, admin, user)
-- role.description: Human-readable description
-- user_role.id: Unique join record identifier
-- user_role.user_id: Reference to user
-- user_role.role_id: Reference to role
