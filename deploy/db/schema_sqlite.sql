-- SQLite Database Schema
-- Optimized for SQLite 3.38+ with JSON1 extension
-- =====================

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Enable JSON1 extension (usually built-in for modern SQLite)
-- PRAGMA compile_options; -- Use this to check if JSON1 is available

-- =====================
-- 🔐 Authentication & Authorization
-- =====================

CREATE TABLE realm (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app_user (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    username TEXT NOT NULL,
    email TEXT UNIQUE,
    hashed_password TEXT NOT NULL,
    is_active INTEGER DEFAULT 1 CHECK (is_active IN (0, 1)), -- SQLite doesn't have BOOLEAN
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE
);

CREATE TABLE role (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    UNIQUE (realm_id, name)
);

CREATE TABLE policy (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    version TEXT DEFAULT '2012-10-17', -- AWS policy version format
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    UNIQUE (realm_id, name)
);

-- AWS-style policy statements
CREATE TABLE statement (
    id TEXT PRIMARY KEY,
    policy_id TEXT NOT NULL,
    sid TEXT, -- Statement ID (optional)
    effect TEXT NOT NULL CHECK (effect IN ('Allow', 'Deny')),
    actions TEXT NOT NULL, -- JSON array as string: '["read:*", "write:project"]'
    resources TEXT NOT NULL, -- JSON array as string: '["project:*", "user:123"]'
    conditions TEXT, -- JSON string: '{"StringEquals": {"project:owner": "${user:id}"}}'
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE
);

CREATE TABLE user_role (
    user_id TEXT,
    role_id TEXT,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);

CREATE TABLE role_policy (
    role_id TEXT,
    policy_id TEXT,
    PRIMARY KEY (role_id, policy_id),
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE
);

-- Direct user policies (like AWS managed policies attached directly to users)
CREATE TABLE user_policy (
    user_id TEXT,
    policy_id TEXT,
    PRIMARY KEY (user_id, policy_id),
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE
);

-- Resource-based policies (like S3 bucket policies)
CREATE TABLE resource_policy (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    resource_type TEXT NOT NULL, -- 'project', 'document', 'code', etc.
    resource_id TEXT NOT NULL, -- Reference to the actual resource
    policy_id TEXT NOT NULL,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE,
    UNIQUE (resource_type, resource_id, policy_id)
);

-- =====================
-- 📁 Project & Content Management
-- =====================

CREATE TABLE project (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    git_url TEXT,
    git_repo TEXT,
    git_branch TEXT,
    language TEXT,
    entry_point TEXT,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE
);

CREATE TABLE code (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    path TEXT NOT NULL,
    code TEXT NOT NULL,
    vector_embedding TEXT, -- Store vector as JSON string: '[0.1, 0.2, 0.3, ...]'
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
);

CREATE TABLE document (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    vector_embedding TEXT, -- Store vector as JSON string: '[0.1, 0.2, 0.3, ...]'
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
);

-- =====================
-- 📊 Indexes for Performance
-- =====================

-- Basic indexes (SQLite optimized)
CREATE INDEX idx_user_realm_id ON app_user(realm_id);
CREATE INDEX idx_user_email ON app_user(email);
CREATE INDEX idx_user_username ON app_user(username);
CREATE INDEX idx_user_active ON app_user(is_active);
CREATE INDEX idx_role_realm_id ON role(realm_id);
CREATE INDEX idx_policy_realm_id ON policy(realm_id);
CREATE INDEX idx_statement_policy_id ON statement(policy_id);
CREATE INDEX idx_statement_effect ON statement(effect);
CREATE INDEX idx_resource_policy_type_id ON resource_policy(resource_type, resource_id);
CREATE INDEX idx_user_policy_user_id ON user_policy(user_id);
CREATE INDEX idx_role_policy_role_id ON role_policy(role_id);
CREATE INDEX idx_project_realm_id ON project(realm_id);
CREATE INDEX idx_code_project_id ON code(project_id);
CREATE INDEX idx_document_project_id ON document(project_id);

-- =====================
-- 🔍 JSON Validation Triggers (SQLite-specific)
-- =====================

-- Trigger to validate JSON in statement.actions
CREATE TRIGGER validate_statement_actions
    BEFORE INSERT ON statement
    FOR EACH ROW
    WHEN json_valid(NEW.actions) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in actions field');
END;

CREATE TRIGGER validate_statement_actions_update
    BEFORE UPDATE ON statement
    FOR EACH ROW
    WHEN json_valid(NEW.actions) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in actions field');
END;

-- Trigger to validate JSON in statement.resources
CREATE TRIGGER validate_statement_resources
    BEFORE INSERT ON statement
    FOR EACH ROW
    WHEN json_valid(NEW.resources) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in resources field');
END;

CREATE TRIGGER validate_statement_resources_update
    BEFORE UPDATE ON statement
    FOR EACH ROW
    WHEN json_valid(NEW.resources) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in resources field');
END;

-- Trigger to validate JSON in statement.conditions (optional field)
CREATE TRIGGER validate_statement_conditions
    BEFORE INSERT ON statement
    FOR EACH ROW
    WHEN NEW.conditions IS NOT NULL AND json_valid(NEW.conditions) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in conditions field');
END;

CREATE TRIGGER validate_statement_conditions_update
    BEFORE UPDATE ON statement
    FOR EACH ROW
    WHEN NEW.conditions IS NOT NULL AND json_valid(NEW.conditions) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in conditions field');
END;

-- Trigger to validate JSON in vector_embedding fields
CREATE TRIGGER validate_code_vector
    BEFORE INSERT ON code
    FOR EACH ROW
    WHEN NEW.vector_embedding IS NOT NULL AND json_valid(NEW.vector_embedding) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in vector_embedding field');
END;

CREATE TRIGGER validate_document_vector
    BEFORE INSERT ON document
    FOR EACH ROW
    WHEN NEW.vector_embedding IS NOT NULL AND json_valid(NEW.vector_embedding) = 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid JSON in vector_embedding field');
END;

-- =====================
-- 🎯 SQLite-Specific Views and Functions
-- =====================

-- View for active users with their roles
CREATE VIEW active_users_with_roles AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    u.realm_id,
    r.id as role_id,
    r.name as role_name
FROM app_user u
LEFT JOIN user_role ur ON u.id = ur.user_id
LEFT JOIN role r ON ur.role_id = r.id
WHERE u.is_active = 1;

-- View for policy statements with policy info
CREATE VIEW policy_statements AS
SELECT 
    p.realm_id,
    p.name as policy_name,
    p.description as policy_description,
    s.id as statement_id,
    s.sid,
    s.effect,
    s.actions,
    s.resources,
    s.conditions
FROM policy p
JOIN statement s ON p.id = s.policy_id;

-- =====================
-- 🔑 Example Data and Queries
-- =====================

-- Example policy data (commented out)
/*
-- Create a default realm
INSERT INTO realm (id, name, description) VALUES 
('default', 'Default Realm', 'Default organizational realm');

-- Create a read-only policy
INSERT INTO policy (id, realm_id, name, description) VALUES 
('readonly-policy', 'default', 'ReadOnlyAccess', 'Provides read-only access to all resources');

INSERT INTO statement (id, policy_id, sid, effect, actions, resources) VALUES 
('readonly-stmt', 'readonly-policy', 'ReadOnlyStatement', 'Allow', 
 '["read:*", "list:*"]', '["*"]');

-- Create an admin user
INSERT INTO app_user (id, realm_id, username, email, hashed_password) VALUES 
('admin-user', 'default', 'admin', 'admin@example.com', '$2a$10$...');

-- Create admin role
INSERT INTO role (id, realm_id, name, description) VALUES 
('admin-role', 'default', 'Administrator', 'Full access administrator role');

-- Assign admin role to admin user
INSERT INTO user_role (user_id, role_id) VALUES ('admin-user', 'admin-role');
*/

-- =====================
-- 📖 SQLite-Specific Documentation
-- =====================

-- Permission Evaluation Logic (AWS-style):
-- 1. Explicit DENY always wins
-- 2. Explicit ALLOW is required (default is implicit deny)
-- 3. Evaluation order: Resource-based policies → User policies → Role policies
-- 4. Conditions are evaluated using variable substitution

-- Action Examples:
-- - read:project, write:project, delete:project
-- - list:documents, create:document
-- - admin:users, modify:roles

-- Resource Examples:
-- - project:* (all projects)
-- - project:123 (specific project)
-- - document:project:123/* (all documents in project 123)
-- - user:${user:id} (current user)

-- JSON Condition Examples:
-- {
--   "StringEquals": {
--     "project:owner": "${user:id}",
--     "realm:id": "${user:realm_id}"
--   },
--   "DateGreaterThan": {
--     "current:time": "2024-01-01T00:00:00Z"
--   },
--   "IpAddress": {
--     "source:ip": "192.168.1.0/24"
--   }
-- }

-- SQLite JSON Query Examples:
-- Find statements with specific conditions:
--   SELECT * FROM statement WHERE json_extract(conditions, '$.StringEquals.realm_id') = '123';
--
-- Find statements that allow specific actions:
--   SELECT * FROM statement WHERE json_extract(actions, '$[0]') LIKE 'read:%';
--
-- Check if action array contains specific action:
--   SELECT * FROM statement WHERE EXISTS (
--     SELECT 1 FROM json_each(actions) WHERE value = 'read:project'
--   );
--
-- Extract and query nested JSON conditions:
--   SELECT * FROM statement WHERE json_extract(conditions, '$.StringEquals') IS NOT NULL;

-- End of SQLite schema 