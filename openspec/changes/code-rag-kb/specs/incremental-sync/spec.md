## ADDED Requirements

### Requirement: Git diff-based incremental knowledge update
The system SHALL support incremental knowledge graph updates by computing the Git diff between the last synced commit and the current HEAD. Only changed files SHALL be re-processed — added files are fully parsed and indexed, modified files are re-parsed and diff-applied, deleted files have their entities and relationships removed.

#### Scenario: Sync after adding new files
- **WHEN** `SyncFromDiff` is called and the Git diff shows 3 new files added since last sync
- **THEN** all 3 files are parsed, entities and relationships are extracted, embeddings are generated, and all data is upserted to both pgvector and Memgraph

#### Scenario: Sync after modifying a file
- **WHEN** a file has been modified since last sync (new functions added, existing function body changed)
- **THEN** the system parses the new version, computes an entity-level diff against stored entities, updates changed entities' embeddings and summaries, adds new entities, and removes entities that no longer exist in the file

#### Scenario: Sync after deleting files
- **WHEN** the Git diff shows 2 files were deleted since last sync
- **THEN** all entities from those files and their relationships are removed from both pgvector and Memgraph

### Requirement: Commit tracking per repository
The system SHALL track the last synced commit SHA per repository. After a successful sync, the last_commit and last_sync fields in the `repositories` table SHALL be updated.

#### Scenario: Track sync progress
- **WHEN** an incremental sync completes successfully
- **THEN** the repository's last_commit is updated to the current HEAD SHA and last_sync is updated to the current timestamp

#### Scenario: First-time sync (full build)
- **WHEN** a repository has no last_commit recorded (first-time indexing)
- **THEN** the system performs a full parse of all files in the repository, treating every file as "added"

### Requirement: Async background sync with progress reporting
The system SHALL support running sync operations as background async jobs with progress reporting (files processed / total files, entities created/updated/deleted counts).

#### Scenario: Background sync with progress
- **WHEN** a sync is triggered via API with async=true
- **THEN** the system returns a job ID immediately and processes the sync in the background, reporting progress via a status endpoint

#### Scenario: Resume on failure
- **WHEN** a sync operation fails partway through (e.g., OpenAI API error on file 50 of 100)
- **THEN** the system records which files were successfully processed and allows resuming from the point of failure
