-- MySQL Database Schema
-- Optimized for MySQL 8.0+ with JSON support
-- =====================

-- Set SQL mode for strict standards compliance
SET sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO';

-- Use UTF8MB4 character set for full Unicode support including emojis
SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- =====================
-- 🔐 Authentication & Authorization
-- =====================

CREATE TABLE realm (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_realm_name (name)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE app_user (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    hashed_password TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    INDEX idx_user_realm_id (realm_id),
    INDEX idx_user_email (email),
    INDEX idx_user_username (username),
    INDEX idx_user_active (is_active)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE role (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    UNIQUE KEY uk_role_realm_name (realm_id, name),
    INDEX idx_role_realm_id (realm_id)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE policy (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(20) DEFAULT '2012-10-17', -- AWS policy version format
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    UNIQUE KEY uk_policy_realm_name (realm_id, name),
    INDEX idx_policy_realm_id (realm_id)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- AWS-style policy statements
CREATE TABLE statement (
    id CHAR(36) PRIMARY KEY,
    policy_id CHAR(36) NOT NULL,
    sid VARCHAR(100), -- Statement ID (optional)
    effect ENUM('Allow', 'Deny') NOT NULL,
    actions JSON NOT NULL, -- JSON array: ["read:*", "write:project"]
    resources JSON NOT NULL, -- JSON array: ["project:*", "user:123"]
    conditions JSON, -- JSON object: {"StringEquals": {"project:owner": "${user:id}"}}
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE,
    INDEX idx_statement_policy_id (policy_id),
    INDEX idx_statement_effect (effect),
    -- MySQL 8.0+ functional indexes for JSON
    INDEX idx_statement_actions ((CAST(actions AS JSON))),
    INDEX idx_statement_resources ((CAST(resources AS JSON))),
    INDEX idx_statement_conditions ((CAST(conditions AS JSON))),
    -- JSON validation constraints
    CONSTRAINT chk_actions_valid CHECK (JSON_VALID(actions)),
    CONSTRAINT chk_resources_valid CHECK (JSON_VALID(resources)),
    CONSTRAINT chk_conditions_valid CHECK (conditions IS NULL OR JSON_VALID(conditions))
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE user_role (
    user_id CHAR(36),
    role_id CHAR(36),
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE,
    INDEX idx_user_role_user_id (user_id),
    INDEX idx_user_role_role_id (role_id)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE role_policy (
    role_id CHAR(36),
    policy_id CHAR(36),
    PRIMARY KEY (role_id, policy_id),
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE,
    INDEX idx_role_policy_role_id (role_id),
    INDEX idx_role_policy_policy_id (policy_id)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Direct user policies (like AWS managed policies attached directly to users)
CREATE TABLE user_policy (
    user_id CHAR(36),
    policy_id CHAR(36),
    PRIMARY KEY (user_id, policy_id),
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE,
    INDEX idx_user_policy_user_id (user_id),
    INDEX idx_user_policy_policy_id (policy_id)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Resource-based policies (like S3 bucket policies)
CREATE TABLE resource_policy (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    resource_type VARCHAR(50) NOT NULL, -- 'project', 'document', 'code', etc.
    resource_id CHAR(36) NOT NULL, -- Reference to the actual resource
    policy_id CHAR(36) NOT NULL,
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES policy(id) ON DELETE CASCADE,
    UNIQUE KEY uk_resource_policy (resource_type, resource_id, policy_id),
    INDEX idx_resource_policy_type_id (resource_type, resource_id),
    INDEX idx_resource_policy_realm_id (realm_id)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================
-- 📁 Project & Content Management
-- =====================

CREATE TABLE project (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    git_url TEXT,
    git_repo VARCHAR(500),
    git_branch VARCHAR(100),
    language VARCHAR(50),
    entry_point TEXT,
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    INDEX idx_project_realm_id (realm_id),
    INDEX idx_project_name (name),
    INDEX idx_project_language (language)
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE code (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    project_id CHAR(36) NOT NULL,
    path TEXT NOT NULL,
    code LONGTEXT NOT NULL,
    vector_embedding JSON, -- Store vector as JSON array: [0.1, 0.2, 0.3, ...]
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    INDEX idx_code_realm_id (realm_id),
    INDEX idx_code_project_id (project_id),
    -- Full-text search index on code content
    FULLTEXT KEY ft_code_content (code),
    -- JSON validation for vector embeddings
    CONSTRAINT chk_vector_valid CHECK (vector_embedding IS NULL OR JSON_VALID(vector_embedding))
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE document (
    id CHAR(36) PRIMARY KEY,
    realm_id CHAR(36) NOT NULL,
    project_id CHAR(36) NOT NULL,
    name VARCHAR(200) NOT NULL,
    path TEXT NOT NULL,
    content LONGTEXT NOT NULL,
    vector_embedding JSON, -- Store vector as JSON array: [0.1, 0.2, 0.3, ...]
    created_by CHAR(36),
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by CHAR(36),
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (realm_id) REFERENCES realm(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    INDEX idx_document_realm_id (realm_id),
    INDEX idx_document_project_id (project_id),
    INDEX idx_document_name (name),
    -- Full-text search index on document content
    FULLTEXT KEY ft_document_content (content),
    FULLTEXT KEY ft_document_name_content (name, content),
    -- JSON validation for vector embeddings
    CONSTRAINT chk_vector_valid CHECK (vector_embedding IS NULL OR JSON_VALID(vector_embedding))
) ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================
-- 🎯 MySQL-Specific Views and Stored Procedures
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
WHERE u.is_active = TRUE;

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

-- View for user permissions (flattened)
CREATE VIEW user_permissions AS
SELECT DISTINCT
    u.id as user_id,
    u.username,
    u.realm_id,
    'role' as permission_source,
    r.name as source_name,
    s.effect,
    s.actions,
    s.resources,
    s.conditions
FROM app_user u
JOIN user_role ur ON u.id = ur.user_id
JOIN role r ON ur.role_id = r.id
JOIN role_policy rp ON r.id = rp.role_id
JOIN policy p ON rp.policy_id = p.id
JOIN statement s ON p.id = s.policy_id
WHERE u.is_active = TRUE

UNION ALL

SELECT DISTINCT
    u.id as user_id,
    u.username,
    u.realm_id,
    'direct' as permission_source,
    p.name as source_name,
    s.effect,
    s.actions,
    s.resources,
    s.conditions
FROM app_user u
JOIN user_policy up ON u.id = up.user_id
JOIN policy p ON up.policy_id = p.id
JOIN statement s ON p.id = s.policy_id
WHERE u.is_active = TRUE;

-- =====================
-- 🚀 MySQL-Specific Functions and Procedures
-- =====================

DELIMITER //

-- Function to check if an action matches a pattern
CREATE FUNCTION action_matches(action VARCHAR(255), pattern VARCHAR(255))
RETURNS BOOLEAN
READS SQL DATA
DETERMINISTIC
BEGIN
    -- Simple wildcard matching: 'read:*' matches 'read:project'
    IF pattern LIKE '%*' THEN
        RETURN action LIKE REPLACE(pattern, '*', '%');
    ELSE
        RETURN action = pattern;
    END IF;
END //

-- Function to check if a resource matches a pattern
CREATE FUNCTION resource_matches(resource VARCHAR(255), pattern VARCHAR(255))
RETURNS BOOLEAN
READS SQL DATA
DETERMINISTIC
BEGIN
    -- Simple wildcard matching with variable substitution placeholder
    IF pattern LIKE '%*' THEN
        RETURN resource LIKE REPLACE(pattern, '*', '%');
    ELSE
        RETURN resource = pattern;
    END IF;
END //

-- Procedure to check user permissions
CREATE PROCEDURE check_user_permission(
    IN p_user_id CHAR(36),
    IN p_action VARCHAR(255),
    IN p_resource VARCHAR(255),
    OUT p_allowed BOOLEAN
)
READS SQL DATA
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE stmt_effect ENUM('Allow', 'Deny');
    DECLARE stmt_actions JSON;
    DECLARE stmt_resources JSON;
    DECLARE action_count INT;
    DECLARE resource_count INT;
    DECLARE i INT;
    DECLARE temp_action VARCHAR(255);
    DECLARE temp_resource VARCHAR(255);
    DECLARE has_deny BOOLEAN DEFAULT FALSE;
    DECLARE has_allow BOOLEAN DEFAULT FALSE;
    
    -- Cursor for user permissions
    DECLARE permission_cursor CURSOR FOR
        SELECT effect, actions, resources
        FROM user_permissions
        WHERE user_id = p_user_id;
    
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    SET p_allowed = FALSE;
    
    OPEN permission_cursor;
    
    permission_loop: LOOP
        FETCH permission_cursor INTO stmt_effect, stmt_actions, stmt_resources;
        IF done THEN
            LEAVE permission_loop;
        END IF;
        
        -- Check if action matches any in the statement
        SET action_count = JSON_LENGTH(stmt_actions);
        SET i = 0;
        action_loop: WHILE i < action_count DO
            SET temp_action = JSON_UNQUOTE(JSON_EXTRACT(stmt_actions, CONCAT('$[', i, ']')));
            IF action_matches(p_action, temp_action) THEN
                -- Check if resource matches any in the statement
                SET resource_count = JSON_LENGTH(stmt_resources);
                SET i = 0;
                resource_loop: WHILE i < resource_count DO
                    SET temp_resource = JSON_UNQUOTE(JSON_EXTRACT(stmt_resources, CONCAT('$[', i, ']')));
                    IF resource_matches(p_resource, temp_resource) THEN
                        IF stmt_effect = 'Deny' THEN
                            SET has_deny = TRUE;
                        ELSEIF stmt_effect = 'Allow' THEN
                            SET has_allow = TRUE;
                        END IF;
                        LEAVE resource_loop;
                    END IF;
                    SET i = i + 1;
                END WHILE;
                LEAVE action_loop;
            END IF;
            SET i = i + 1;
        END WHILE;
    END LOOP;
    
    CLOSE permission_cursor;
    
    -- AWS-style evaluation: Explicit deny always wins, explicit allow required
    IF has_deny THEN
        SET p_allowed = FALSE;
    ELSEIF has_allow THEN
        SET p_allowed = TRUE;
    ELSE
        SET p_allowed = FALSE; -- Default deny
    END IF;
    
END //

DELIMITER ;

-- =====================
-- 🔑 Example Data and Usage
-- =====================

-- Example data (commented out)
/*
-- Create a default realm
INSERT INTO realm (id, name, description) VALUES 
('default-realm', 'Default Realm', 'Default organizational realm');

-- Create a read-only policy
INSERT INTO policy (id, realm_id, name, description) VALUES 
('readonly-policy', 'default-realm', 'ReadOnlyAccess', 'Provides read-only access to all resources');

INSERT INTO statement (id, policy_id, sid, effect, actions, resources) VALUES 
('readonly-stmt', 'readonly-policy', 'ReadOnlyStatement', 'Allow', 
 JSON_ARRAY('read:*', 'list:*'), JSON_ARRAY('*'));

-- Create project admin policy with conditions
INSERT INTO statement (id, policy_id, sid, effect, actions, resources, conditions) VALUES 
('project-admin-stmt', 'readonly-policy', 'ProjectAdminStatement', 'Allow', 
 JSON_ARRAY('project:*'), JSON_ARRAY('project:${user:project_id}'),
 JSON_OBJECT('StringEquals', JSON_OBJECT('project:owner', '${user:id}')));

-- Create an admin user
INSERT INTO app_user (id, realm_id, username, email, hashed_password) VALUES 
('admin-user', 'default-realm', 'admin', 'admin@example.com', '$2a$10$...');

-- Create admin role
INSERT INTO role (id, realm_id, name, description) VALUES 
('admin-role', 'default-realm', 'Administrator', 'Full access administrator role');

-- Assign admin role to admin user
INSERT INTO user_role (user_id, role_id) VALUES ('admin-user', 'admin-role');

-- Test permission check
CALL check_user_permission('admin-user', 'read:project', 'project:123', @allowed);
SELECT @allowed as permission_granted;
*/

-- =====================
-- 📖 MySQL-Specific Documentation
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

-- MySQL JSON Query Examples:
-- Find statements with specific conditions:
--   SELECT * FROM statement WHERE JSON_EXTRACT(conditions, '$.StringEquals.realm_id') = '123';
--
-- Find statements that allow specific actions:
--   SELECT * FROM statement WHERE JSON_CONTAINS(actions, '"read:project"');
--
-- Check if action array contains pattern:
--   SELECT * FROM statement WHERE JSON_SEARCH(actions, 'one', 'read:*') IS NOT NULL;
--
-- Complex JSON path queries:
--   SELECT * FROM statement WHERE JSON_EXTRACT(conditions, '$.StringEquals') IS NOT NULL;
--
-- Full-text search in code:
--   SELECT * FROM code WHERE MATCH(code) AGAINST('function calculateTotal' IN NATURAL LANGUAGE MODE);
--
-- Permission check using stored procedure:
--   CALL check_user_permission('user-id', 'read:project', 'project:123', @result);

-- End of MySQL schema 