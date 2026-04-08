# semantic-enrichment Specification

## Purpose
TBD - created by archiving change code-rag-kb-mvp. Update Purpose after archive.
## Requirements
### Requirement: Embedding generation and vector storage
The system SHALL generate embeddings for indexed entities using the configured embedding API when credentials are available. The embedding input SHALL combine language, entity type, name, signature, and doc string. Embeddings SHALL be stored in a **sqlite-vec `vec0` virtual table** (`codekg_entity_vec`) with cosine distance metric, keyed by entity ID.

#### Scenario: Generate and store embeddings
- **WHEN** sync produces one or more entities and embedding credentials are configured
- **THEN** the system generates embeddings via the configured API and inserts them into the `codekg_entity_vec` virtual table using `sqlite_vec.SerializeFloat32` for binary serialization

#### Scenario: Continue without embeddings
- **WHEN** embedding credentials are not configured
- **THEN** the system completes sync without embeddings and remains searchable via keyword fallback

### Requirement: KNN search via sqlite-vec
The system SHALL use the sqlite-vec `MATCH` operator for vector similarity search when embeddings are available.

#### Scenario: Vector-based search
- **WHEN** a search request is received and the embedding service is configured
- **THEN** the system embeds the query, runs `SELECT entity_id, distance FROM codekg_entity_vec WHERE embedding MATCH ? AND k = ?`, and joins results with the entity table

#### Scenario: Keyword fallback
- **WHEN** the embedding service is not configured or the vec query fails
- **THEN** the system falls back to keyword-based ranking over entity names and bodies

