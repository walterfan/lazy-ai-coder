# Migration 010: learning_record → code_mate_artifacts

Renames the learning_record table to code_mate_artifacts and maps legacy `input_type` values to Code Mate types.

## input_type mapping

| Old (learning_record) | New (code_mate_artifacts) |
|-----------------------|----------------------------|
| word, sentence, question | learn_tech |
| idea                  | tech_design |
| topic                 | learn_tech |

## SQLite

Use `010_rename_learning_records_to_code_mate_artifacts.sql` as-is.

- Creates `code_mate_artifacts` table and indexes.
- Copies rows from `learning_record` with mapped `input_type`.
- Drops `learning_record`.

**Fresh install (no learning_record table):** Run only the `CREATE TABLE` and `CREATE INDEX` statements from the file; omit the `INSERT...SELECT` and `DROP TABLE` to avoid errors.

## PostgreSQL

Equivalent steps (run in order):

```sql
CREATE TABLE IF NOT EXISTS code_mate_artifacts (
    id TEXT PRIMARY KEY,
    input_type TEXT NOT NULL,
    user_input TEXT NOT NULL,
    response_payload TEXT,
    user_id TEXT NOT NULL,
    realm_id TEXT,
    created_by TEXT,
    created_time TIMESTAMP DEFAULT NOW(),
    updated_by TEXT,
    updated_time TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_input_type ON code_mate_artifacts(input_type);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_user_id ON code_mate_artifacts(user_id);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_realm_id ON code_mate_artifacts(realm_id);
CREATE INDEX IF NOT EXISTS idx_code_mate_artifacts_deleted_at ON code_mate_artifacts(deleted_at);

-- Only if learning_record exists:
INSERT INTO code_mate_artifacts (id, input_type, user_input, response_payload, user_id, realm_id, created_by, created_time, updated_by, updated_time, deleted_at)
SELECT id,
  CASE input_type WHEN 'idea' THEN 'tech_design' WHEN 'word' THEN 'learn_tech' WHEN 'sentence' THEN 'learn_tech' WHEN 'question' THEN 'learn_tech' WHEN 'topic' THEN 'learn_tech' ELSE 'learn_tech' END,
  user_input, response_payload, user_id, realm_id, created_by, created_time, updated_by, updated_time, deleted_at
FROM learning_record;

DROP TABLE IF EXISTS learning_record;
```

## MySQL

Same structure as PostgreSQL; use `TIMESTAMP` and `DEFAULT CURRENT_TIMESTAMP` / `ON UPDATE CURRENT_TIMESTAMP` as in other MySQL migrations in this project. Table and index names unchanged; copy and drop logic identical to PostgreSQL.
