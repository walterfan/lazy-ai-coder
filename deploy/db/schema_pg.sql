-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- =====================
-- 🔐 Authentication & Authorization
-- =====================

CREATE TABLE realm (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW()
);

CREATE TABLE app_user (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    hashed_password TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW()
);

CREATE TABLE role (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW(),
    UNIQUE (realm_id, name)
);

CREATE TABLE policy (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(20) DEFAULT '2012-10-17', -- AWS policy version format
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW(),
    UNIQUE (realm_id, name)
);

-- AWS-style policy statements
CREATE TABLE statement (
    id UUID PRIMARY KEY,
    policy_id UUID NOT NULL REFERENCES policy(id) ON DELETE CASCADE,
    sid VARCHAR(100), -- Statement ID (optional)
    effect VARCHAR(10) NOT NULL CHECK (effect IN ('Allow', 'Deny')),
    actions TEXT[] NOT NULL, -- Array of actions like ['read:*', 'write:project']
    resources TEXT[] NOT NULL, -- Array of resources like ['project:*', 'user:123']
    conditions JSONB, -- JSONB conditions like {"StringEquals": {"project:owner": "${user:id}"}}
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_role (
    user_id UUID REFERENCES app_user(id) ON DELETE CASCADE,
    role_id UUID REFERENCES role(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE role_policy (
    role_id UUID REFERENCES role(id) ON DELETE CASCADE,
    policy_id UUID REFERENCES policy(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, policy_id)
);

-- Direct user policies (like AWS managed policies attached directly to users)
CREATE TABLE user_policy (
    user_id UUID REFERENCES app_user(id) ON DELETE CASCADE,
    policy_id UUID REFERENCES policy(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, policy_id)
);

-- Resource-based policies (like S3 bucket policies)
CREATE TABLE resource_policy (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    resource_type VARCHAR(50) NOT NULL, -- 'project', 'document', 'code', etc.
    resource_id UUID NOT NULL, -- Reference to the actual resource
    policy_id UUID NOT NULL REFERENCES policy(id) ON DELETE CASCADE,
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW(),
    UNIQUE (resource_type, resource_id, policy_id)
);

-- Indexes for performance
CREATE INDEX idx_statement_policy_id ON statement(policy_id);
CREATE INDEX idx_statement_effect ON statement(effect);
CREATE INDEX idx_statement_actions ON statement USING GIN(actions);
CREATE INDEX idx_statement_resources ON statement USING GIN(resources);
CREATE INDEX idx_statement_conditions ON statement USING GIN(conditions);
CREATE INDEX idx_resource_policy_type_id ON resource_policy(resource_type, resource_id);
CREATE INDEX idx_user_policy_user_id ON user_policy(user_id);
CREATE INDEX idx_role_policy_role_id ON role_policy(role_id);

-- =====================
-- 📁 Project & Content Management
-- =====================

CREATE TABLE project (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    git_url TEXT,
    git_repo TEXT,
    git_branch VARCHAR(100),
    language VARCHAR(50),
    entry_point TEXT,
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW()
);

CREATE TABLE code (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    project_id UUID NOT NULL REFERENCES project(id),
    path TEXT NOT NULL,
    code TEXT NOT NULL,
    vector VECTOR(1536),
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW()
);

CREATE INDEX code_vector_hnsw_idx ON code USING hnsw (vector vector_l2_ops);

CREATE TABLE document (
    id UUID PRIMARY KEY,
    realm_id UUID NOT NULL REFERENCES realm(id),
    project_id UUID NOT NULL REFERENCES project(id),
    name VARCHAR(200) NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    vector VECTOR(1536),
    created_by UUID,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by UUID,
    updated_time TIMESTAMP DEFAULT NOW()
);

CREATE INDEX document_vector_hnsw_idx ON document USING hnsw (vector vector_cosine_ops);

-- Optional: Full-text search support for documents
ALTER TABLE document ADD COLUMN content_tsvector tsvector
  GENERATED ALWAYS AS (to_tsvector('english', content)) STORED;

CREATE INDEX document_fts_idx ON document USING GIN (content_tsvector);

-- =====================
-- 🔑 AWS-Style Permission Examples
-- =====================

-- Example: Read-only policy
-- INSERT INTO policy (id, realm_id, name, description) VALUES 
-- ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'ReadOnlyAccess', 'Provides read-only access to all resources');

-- INSERT INTO statement (id, policy_id, sid, effect, actions, resources) VALUES 
-- ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'ReadOnlyStatement', 'Allow', 
--  ARRAY['read:*', 'list:*'], ARRAY['*']);

-- Example: Project admin policy with conditions
-- INSERT INTO statement (id, policy_id, sid, effect, actions, resources, conditions) VALUES 
-- ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'ProjectAdminStatement', 'Allow', 
--  ARRAY['project:*'], ARRAY['project:${user:project_id}'],
--  '{"StringEquals": {"project:owner": "${user:id}"}}');

-- Example: Deny sensitive operations
-- INSERT INTO statement (id, policy_id, sid, effect, actions, resources) VALUES 
-- ('550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'DenySensitiveStatement', 'Deny', 
--  ARRAY['delete:*', 'modify:schema'], ARRAY['*']);

-- =====================
-- 🚀 PostgreSQL-Specific Optimizations
-- =====================

-- Additional indexes for better performance
CREATE INDEX idx_user_realm_id ON app_user(realm_id);
CREATE INDEX idx_user_email ON app_user(email);
CREATE INDEX idx_user_username ON app_user(username);
CREATE INDEX idx_role_realm_id ON role(realm_id);
CREATE INDEX idx_policy_realm_id ON policy(realm_id);
CREATE INDEX idx_project_realm_id ON project(realm_id);
CREATE INDEX idx_code_project_id ON code(project_id);
CREATE INDEX idx_document_project_id ON document(project_id);

-- JSONB query optimization indexes
CREATE INDEX idx_statement_conditions_string_equals ON statement USING GIN((conditions -> 'StringEquals'));
CREATE INDEX idx_statement_conditions_date_greater ON statement USING GIN((conditions -> 'DateGreaterThan'));
CREATE INDEX idx_statement_conditions_ip_address ON statement USING GIN((conditions -> 'IpAddress'));

-- Partial indexes for performance
CREATE INDEX idx_active_users ON app_user(id) WHERE is_active = true;
CREATE INDEX idx_allow_statements ON statement(policy_id) WHERE effect = 'Allow';
CREATE INDEX idx_deny_statements ON statement(policy_id) WHERE effect = 'Deny';

-- =====================
-- 🎯 Useful Functions for Permission Checking
-- =====================

-- Function to check if an action matches a pattern
CREATE OR REPLACE FUNCTION action_matches(action TEXT, pattern TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- Simple wildcard matching: 'read:*' matches 'read:project'
    IF pattern LIKE '%*' THEN
        RETURN action LIKE REPLACE(pattern, '*', '%');
    ELSE
        RETURN action = pattern;
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to check if a resource matches a pattern
CREATE OR REPLACE FUNCTION resource_matches(resource TEXT, pattern TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- Simple wildcard matching with variable substitution placeholder
    IF pattern LIKE '%*' THEN
        RETURN resource LIKE REPLACE(pattern, '*', '%');
    ELSE
        RETURN resource = pattern;
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- =====================
-- 📖 Permission System Documentation
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

-- JSONB Condition Examples:
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
--   },
--   "NumericLessThan": {
--     "user:age": 30
--   },
--   "Bool": {
--     "user:is_admin": true
--   }
-- }

-- JSONB Query Examples:
-- Find statements with specific string conditions:
--   SELECT * FROM statement WHERE conditions @> '{"StringEquals": {"realm:id": "123"}}';
-- 
-- Find statements that allow project actions:
--   SELECT * FROM statement WHERE 'project:read' = ANY(actions);
-- 
-- Complex condition checking:
--   SELECT * FROM statement WHERE conditions ? 'StringEquals' AND conditions->'StringEquals' ? 'project:owner';

-- End of schema
