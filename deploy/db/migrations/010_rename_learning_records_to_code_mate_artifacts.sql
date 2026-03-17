-- Migration 010: Rename learning_record to code_mate_artifacts and migrate input_type
-- Code Mate Agent refactor: table rename + map old types to research_solution | learn_tech | tech_design
--
-- input_type mapping:
--   word, sentence, question -> learn_tech
--   idea -> tech_design
--   topic -> learn_tech
--
-- SQLite: create new table, copy with mapping, drop old (SQLite has no ALTER COLUMN for bulk update)

-- Create new table (same structure as learning_record, name code_mate_artifacts)
CREATE TABLE IF NOT EXISTS code_mate_artifacts (
    id TEXT PRIMARY KEY,
    input_type TEXT NOT NULL,
    user_input TEXT NOT NULL,
    response_payload TEXT,
    user_id TEXT NOT NULL,
    realm_id TEXT,
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_input_type ON code_mate_artifacts(input_type);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_user_id ON code_mate_artifacts(user_id);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_realm_id ON code_mate_artifacts(realm_id);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_deleted_at ON code_mate_artifacts(deleted_at);

-- Copy data from learning_record with input_type mapping.
-- word/sentence/question -> learn_tech; idea -> tech_design; topic -> learn_tech
-- If learning_record does not exist (fresh install), skip this block and run only CREATE above; app will use code_mate_artifacts via GORM.
INSERT INTO code_mate_artifacts (
    id, input_type, user_input, response_payload, user_id, realm_id,
    created_by, created_time, updated_by, updated_time, deleted_at
)
SELECT
    id,
    CASE input_type
        WHEN 'idea' THEN 'tech_design'
        WHEN 'word' THEN 'learn_tech'
        WHEN 'sentence' THEN 'learn_tech'
        WHEN 'question' THEN 'learn_tech'
        WHEN 'topic' THEN 'learn_tech'
        ELSE 'learn_tech'
    END,
    user_input,
    response_payload,
    user_id,
    realm_id,
    created_by,
    created_time,
    updated_by,
    updated_time,
    deleted_at
FROM learning_record;

DROP TABLE IF EXISTS learning_record;
