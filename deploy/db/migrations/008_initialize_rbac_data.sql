-- Migration: Initialize RBAC (Role-Based Access Control) Data
-- Date: 2025-11-22
-- Description:
--   1. Create policies and policy statements for resource access control
--   2. Assign super_admin role to walter (cross-realm access)
--   3. Assign admin roles to fiona and cynthia (realm-scoped)
--   4. Link policies to roles via role_policies
--
-- Authorization Model:
--   - super_admin: Full access to ALL resources across ALL realms
--   - admin: Full CRUD access to resources within their own realm
--   - user: Read-only access to resources within their own realm

-- =======================
-- STEP 1: Create Policies
-- =======================

-- Policy 1: Super Admin Policy (Cross-realm, all operations)
INSERT OR IGNORE INTO policies (id, realm_id, name, description, version, created_by, created_time, updated_by, updated_time)
VALUES (
    'policy_super_admin_full_access',
    'system',
    'SuperAdminFullAccess',
    'Full access to all resources across all realms for super administrators',
    '2012-10-17',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- Policy 2: Admin Policy (Realm-scoped, full CRUD)
INSERT OR IGNORE INTO policies (id, realm_id, name, description, version, created_by, created_time, updated_by, updated_time)
VALUES (
    'policy_admin_realm_full_access',
    'system',
    'AdminRealmFullAccess',
    'Full CRUD access to all resources within the administrator''s own realm',
    '2012-10-17',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- Policy 3: User Policy (Realm-scoped, read + own resources CRUD)
INSERT OR IGNORE INTO policies (id, realm_id, name, description, version, created_by, created_time, updated_by, updated_time)
VALUES (
    'policy_user_realm_access',
    'system',
    'UserRealmAccess',
    'Read access to realm resources and full access to own resources',
    '2012-10-17',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- ====================================
-- STEP 2: Create Policy Statements
-- ====================================

-- Super Admin Statements: Allow ALL actions on ALL resources
INSERT OR IGNORE INTO statements (id, policy_id, s_id, effect, actions, resources, created_by, created_time, updated_by, updated_time)
VALUES (
    'stmt_super_admin_all',
    'policy_super_admin_full_access',
    'SuperAdminAllResources',
    'Allow',
    '["*"]',
    '["*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- Admin Statements: Full CRUD on realm-scoped resources
INSERT OR IGNORE INTO statements (id, policy_id, s_id, effect, actions, resources, created_by, created_time, updated_by, updated_time)
VALUES (
    'stmt_admin_projects_full',
    'policy_admin_realm_full_access',
    'AdminProjectsFullAccess',
    'Allow',
    '["projects:read", "projects:create", "projects:update", "projects:delete"]',
    '["projects:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_admin_prompts_full',
    'policy_admin_realm_full_access',
    'AdminPromptsFullAccess',
    'Allow',
    '["prompts:read", "prompts:create", "prompts:update", "prompts:delete"]',
    '["prompts:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_admin_documents_full',
    'policy_admin_realm_full_access',
    'AdminDocumentsFullAccess',
    'Allow',
    '["documents:read", "documents:create", "documents:update", "documents:delete"]',
    '["documents:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_admin_cursor_rules_full',
    'policy_admin_realm_full_access',
    'AdminCursorRulesFullAccess',
    'Allow',
    '["cursor_rules:read", "cursor_rules:create", "cursor_rules:update", "cursor_rules:delete"]',
    '["cursor_rules:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_admin_cursor_commands_full',
    'policy_admin_realm_full_access',
    'AdminCursorCommandsFullAccess',
    'Allow',
    '["cursor_commands:read", "cursor_commands:create", "cursor_commands:update", "cursor_commands:delete"]',
    '["cursor_commands:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- User Statements: Read realm resources + CRUD own resources
INSERT OR IGNORE INTO statements (id, policy_id, s_id, effect, actions, resources, created_by, created_time, updated_by, updated_time)
VALUES (
    'stmt_user_projects_read',
    'policy_user_realm_access',
    'UserProjectsRead',
    'Allow',
    '["projects:read"]',
    '["projects:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_user_prompts_read',
    'policy_user_realm_access',
    'UserPromptsRead',
    'Allow',
    '["prompts:read"]',
    '["prompts:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_user_own_prompts_full',
    'policy_user_realm_access',
    'UserOwnPromptsFullAccess',
    'Allow',
    '["prompts:create", "prompts:update", "prompts:delete"]',
    '["prompts:own"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_user_cursor_rules_read',
    'policy_user_realm_access',
    'UserCursorRulesRead',
    'Allow',
    '["cursor_rules:read"]',
    '["cursor_rules:*"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
),
(
    'stmt_user_own_cursor_rules_full',
    'policy_user_realm_access',
    'UserOwnCursorRulesFullAccess',
    'Allow',
    '["cursor_rules:create", "cursor_rules:update", "cursor_rules:delete"]',
    '["cursor_rules:own"]',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- ====================================
-- STEP 3: Link Policies to Roles
-- ====================================

-- Super Admin Role gets SuperAdminFullAccess policy
INSERT OR IGNORE INTO role_policies (role_id, policy_id)
VALUES ('role_super_admin', 'policy_super_admin_full_access');

-- Admin Role gets AdminRealmFullAccess policy
INSERT OR IGNORE INTO role_policies (role_id, policy_id)
VALUES ('role_admin', 'policy_admin_realm_full_access');

-- User Role gets UserRealmAccess policy
INSERT OR IGNORE INTO role_policies (role_id, policy_id)
VALUES ('role_user', 'policy_user_realm_access');

-- ====================================
-- STEP 4: Assign Roles to Users
-- ====================================

-- Assign super_admin role to walter
-- walter's user_id: 58d13410-c775-41b1-82c3-1956f1b70807
INSERT OR IGNORE INTO user_roles (id, user_id, role_id, created_by, created_time, updated_by, updated_time)
VALUES (
    'ur_walter_super_admin',
    '58d13410-c775-41b1-82c3-1956f1b70807',
    'role_super_admin',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- Assign admin role to fiona (for fiona realm)
-- fiona's user_id: 37aa0519-3bdc-4731-b101-c1d85afeffa0
INSERT OR IGNORE INTO user_roles (id, user_id, role_id, created_by, created_time, updated_by, updated_time)
VALUES (
    'ur_fiona_admin',
    '37aa0519-3bdc-4731-b101-c1d85afeffa0',
    'role_admin',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- Assign admin role to cynthia (for cynthia realm)
-- cynthia's user_id: 389989a9-3510-401c-8040-a3c30b0823e5
INSERT OR IGNORE INTO user_roles (id, user_id, role_id, created_by, created_time, updated_by, updated_time)
VALUES (
    'ur_cynthia_admin',
    '389989a9-3510-401c-8040-a3c30b0823e5',
    'role_admin',
    'system',
    datetime('now'),
    'system',
    datetime('now')
);

-- ====================================
-- VERIFICATION QUERIES (for reference)
-- ====================================

-- Check user roles:
-- SELECT u.username, u.realm_id, r.name as role_name
-- FROM app_user u
-- INNER JOIN user_roles ur ON u.id = ur.user_id
-- INNER JOIN roles r ON ur.role_id = r.id
-- WHERE u.deleted_at IS NULL AND ur.deleted_at IS NULL;

-- Check role policies:
-- SELECT r.name as role_name, p.name as policy_name, p.description
-- FROM roles r
-- INNER JOIN role_policies rp ON r.id = rp.role_id
-- INNER JOIN policies p ON rp.policy_id = p.id;

-- Check policy statements:
-- SELECT p.name as policy_name, s.s_id, s.effect, s.actions, s.resources
-- FROM policies p
-- INNER JOIN statements s ON p.id = s.policy_id
-- ORDER BY p.name, s.s_id;
