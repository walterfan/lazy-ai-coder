-- Migration: Add Role-Based Access Control and User Approval Workflow (PostgreSQL)
-- ==================================================================================

-- Create role table
CREATE TABLE IF NOT EXISTS role (
    id VARCHAR(255) PRIMARY KEY,
    realm_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by VARCHAR(255),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(realm_id, name)
);

COMMENT ON TABLE role IS 'Roles for role-based access control';
COMMENT ON COLUMN role.id IS 'Unique role identifier';
COMMENT ON COLUMN role.realm_id IS 'Realm this role belongs to (system for global roles)';
COMMENT ON COLUMN role.name IS 'Role name (super_admin, admin, user)';

-- Create user_role join table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_role (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    created_by VARCHAR(255),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);

COMMENT ON TABLE user_role IS 'Many-to-many relationship between users and roles';

-- Create indexes for role lookups
CREATE INDEX IF NOT EXISTS idx_role_realm_id ON role(realm_id);
CREATE INDEX IF NOT EXISTS idx_role_name ON role(name);
CREATE INDEX IF NOT EXISTS idx_user_role_user_id ON user_role(user_id);
CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON user_role(role_id);

-- Modify app_user table: is_active now defaults to false (requires admin approval)
ALTER TABLE app_user ALTER COLUMN is_active SET DEFAULT false;

-- Create indexes on app_user for filtering
CREATE INDEX IF NOT EXISTS idx_user_is_active ON app_user(is_active);
CREATE INDEX IF NOT EXISTS idx_user_realm_id ON app_user(realm_id);

-- Seed system-level realm for super admins
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
