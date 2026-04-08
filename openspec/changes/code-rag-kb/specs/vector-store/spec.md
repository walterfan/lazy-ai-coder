## ADDED Requirements

### Requirement: Store code entity embeddings in pgvector
The system SHALL store code entity embeddings in a PostgreSQL table `code_entities` with a `vector(3072)` column and an HNSW index using cosine distance. The table SHALL include columns: id, repo_id, entity_type, name, file_path, start_line, end_line, signature, doc_string, body, summary, embedding, metadata (JSONB), created_at, updated_at.

#### Scenario: Insert a new code entity with embedding
- **WHEN** a code entity with its embedding vector is stored
- **THEN** a row is created in `code_entities` with all fields populated and the embedding stored as a pgvector vector type

#### Scenario: Upsert an existing code entity
- **WHEN** an entity with the same ID already exists in the table
- **THEN** the system updates the existing row's fields including the embedding vector and updated_at timestamp

### Requirement: Vector similarity search
The system SHALL support approximate nearest neighbor (ANN) search on code entity embeddings using pgvector's HNSW index with cosine distance operator (`<=>`). The search SHALL return Top-K results with similarity scores.

#### Scenario: Search for semantically similar code entities
- **WHEN** a query embedding and topK=20 are provided
- **THEN** the system returns the 20 most similar code entities ordered by cosine similarity, each with a similarity score

#### Scenario: Filter by entity type during vector search
- **WHEN** a vector search is performed with entity_type filter "function"
- **THEN** only Function entities are considered in the similarity search

### Requirement: Full-text search on code content
The system SHALL support full-text search on code entity names, signatures, and bodies using PostgreSQL GIN indexes with `pg_trgm` for trigram matching and `tsvector` for English full-text search.

#### Scenario: Search by function name substring
- **WHEN** a text query "HandleRequest" is submitted for full-text search
- **THEN** the system returns entities whose name or signature contains "HandleRequest", ranked by relevance

#### Scenario: Search by code content keyword
- **WHEN** a text query "retry backoff" is submitted
- **THEN** the system returns entities whose body or doc_string matches "retry" and "backoff" using full-text search ranking

### Requirement: Code chunk storage for fine-grained retrieval
The system SHALL store code chunks (sub-entity segments) in a `code_chunks` table with entity_id reference, chunk_index, content, embedding, and metadata. Chunks SHALL have their own HNSW vector index.

#### Scenario: Store chunks from a large function
- **WHEN** a function with 200 lines of code is chunked into 4 segments
- **THEN** 4 rows are created in `code_chunks`, each referencing the parent entity ID, with sequential chunk_index values and individual embeddings

### Requirement: Repository tracking
The system SHALL maintain a `repositories` table tracking indexed repositories with fields: id, name, url, branch, last_commit, last_sync timestamp, and config (JSONB).

#### Scenario: Register a new repository
- **WHEN** a repository is registered for indexing
- **THEN** a row is created in `repositories` with the repo name, URL, default branch, and empty last_commit/last_sync

#### Scenario: Update sync status after indexing
- **WHEN** repository indexing completes successfully
- **THEN** the system updates last_commit to the current HEAD and last_sync to the current timestamp

### Requirement: Delete by file path
The system SHALL support deleting all code entities and their associated chunks for a given file path.

#### Scenario: Delete entities for a removed file
- **WHEN** deletion is requested for file path "internal/auth/jwt.go"
- **THEN** all rows in `code_entities` with that file_path and their associated `code_chunks` rows are deleted
