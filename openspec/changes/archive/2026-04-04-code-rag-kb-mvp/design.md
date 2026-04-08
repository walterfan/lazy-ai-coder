## Context

The repository already contains reusable building blocks for an MVP:

- `internal/rag/code_parser.go` for Tree-sitter-based parsing
- `internal/rag/embedding_service.go` for embedding generation
- `internal/llm/openai.go` for LLM-based text generation
- Gin + GORM integration patterns already used by the application
- an existing web app where a verification page can be added

The MVP intentionally reuses those pieces and adds **sqlite-vec** for vector storage and KNN search, keeping the stack SQLite-only with zero external dependencies.

## Goals / Non-Goals

**Goals**

- prove the user workflow end to end
- index a user-specified repository with minimal setup
- expose enough search and browsing capability to validate usefulness
- generate PKB-style repository knowledge docs for verification
- keep the architecture simple enough to ship quickly
- use a single SQLite database for both relational data and vector search

**Non-Goals**

- relationship extraction between code entities
- Memgraph integration
- pgvector integration
- hybrid retrieval and RRF
- incremental Git-diff sync
- MCP tools and resources
- rich graph visualization

## MVP Architecture

```mermaid
flowchart LR
    User[User] --> UI[CodeKGView.vue]
    UI --> API[/api/v1/codekg/*]
    API --> Service[codekg.Service]
    Service --> Parser[rag.CodeParser]
    Service --> Embedder[rag.EmbeddingService]
    Service --> LLM[internal/llm]
    Service --> DB[(SQLite + sqlite-vec)]
```

## Key Decisions

### D1: Reuse existing parser, embedder, and LLM wrappers

**Decision**: Compose the MVP out of `internal/rag` and `internal/llm` instead of introducing new foundational services.

**Rationale**: This keeps MVP scope small and reduces integration risk.

**Caveats**:
- `llm.AskLLM()` does not check HTTP status codes or validate response shape; a non-200 reply or empty `Choices` will produce confusing errors. Acceptable for MVP, but should be hardened before production use.
- `rag.CodeParser` does not populate `DocString` for extracted functions/classes (comments are collected globally but not associated). This degrades embedding quality. Acceptable for MVP demonstration.
- TypeScript files (`.ts`/`.tsx`) are parsed using the JavaScript grammar, which is best-effort. Documented as a known limitation.
- `rag.EmbeddingService` has no retry/backoff for rate-limiting (429) or transient errors. A single batch failure skips that batch but does not abort the sync.

### D2: SQLite + sqlite-vec for all storage

**Decision**: Use GORM tables for relational data (repositories, entities, knowledge docs) and a **sqlite-vec `vec0` virtual table** for embedding vectors. Both live in the same SQLite database file.

**Rationale**: This keeps the deployment as a single binary + single file database with zero external services, while providing real KNN vector search instead of brute-force in-memory scanning. sqlite-vec is loaded via `github.com/asg017/sqlite-vec-go-bindings/cgo` which statically links with the existing `mattn/go-sqlite3` driver used by GORM.

**Schema**:
```sql
CREATE VIRTUAL TABLE IF NOT EXISTS codekg_entity_vec USING vec0(
    entity_id TEXT PRIMARY KEY,
    embedding float[1536] distance_metric=cosine
);
```

### D3: Use full-rebuild sync instead of incremental sync

**Decision**: Re-index all code files on sync for the MVP.

**Rationale**: Simpler correctness model; adequate for basic validation.

### D4: KNN search via sqlite-vec, keyword fallback

**Decision**: When embeddings are available, search uses sqlite-vec `MATCH` operator for KNN retrieval. When the embedding service is not configured, the system falls back to keyword matching.

**Search query**:
```sql
SELECT entity_id, distance
FROM codekg_entity_vec
WHERE embedding MATCH ? AND k = ?
```

**Rationale**: sqlite-vec provides efficient, indexed vector search without loading all entities into memory. This eliminates the previous OOM risk on large repositories and produces better-ranked results via true cosine distance.

### D5: Make PKB docs first-class MVP outputs

**Decision**: Generate `repo-map` mechanically and `overview` via LLM after sync.

**Rationale**: PKB documents are valuable even without graph retrieval and make the MVP easier to validate by both humans and AI agents.

## Knowledge Docs

### Repo Map

Generated without LLM from indexed files/entities. Includes:

- directory structure summary
- language breakdown
- entity type counts
- key entry points
- indexing stats

### Overview

Generated via LLM from repository metadata and sampled indexed entities. Includes:

- purpose
- technology stack
- high-level architecture
- key components
- entry points

## Risks / Trade-offs

**Sync cost risk**  
Full rebuilds are inefficient, but acceptable until incremental sync is justified.

**Entity depth risk**  
Flat entity extraction will not answer relationship-heavy questions well; that limitation is explicit and deferred.

**CGO dependency**  
sqlite-vec CGO bindings require a C compiler at build time. This is already the case for `mattn/go-sqlite3`. Cross-compilation requires CGO_ENABLED=1 and the appropriate C toolchain.

**vec0 virtual table constraints**  
sqlite-vec `vec0` tables do not support conflict resolution clauses (`INSERT OR REPLACE`, `ON CONFLICT`). The code uses explicit `DELETE` + `INSERT` when upserting vectors. Entity IDs include `startLine` in the hash to prevent collisions from same-named entities in a file.

**Sync status volatility**  
Sync status is stored in-memory only (`Service.syncJobs`). A server restart during sync loses progress. Acceptable for MVP; a persistent job table is deferred.
