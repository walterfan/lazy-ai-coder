## Why

Large codebases are hard to navigate, and AI agents do not understand project-specific structure out of the box. Before investing in a full GraphRAG architecture, we need a working end-to-end path that proves the user flow:

- register a repository
- index code entities
- search the indexed knowledge
- generate basic project knowledge documents
- verify the results in a frontend page

This MVP establishes that thin slice with minimal infrastructure and minimal operational cost.

## What Changes

- add a new `internal/codekg/` module built on top of existing `internal/rag` and `internal/llm`
- support repository registration and async sync for local paths / repository sources
- parse source files with the existing Tree-sitter-based parser and extract basic code entities
- optionally generate embeddings via the existing embedding service
- persist repositories, entities, and generated knowledge docs with GORM (SQLite)
- store embeddings in a sqlite-vec `vec0` virtual table for KNN vector search, with keyword fallback
- generate PKB-inspired documents after sync:
  - `repo-map`
  - `overview`
- expose REST endpoints for repos, sync status, entity browsing, search, and knowledge docs
- add a frontend page to verify the MVP end to end

## Capabilities

### New Capabilities

- `code-parsing`: entity extraction using existing Tree-sitter parsing
- `semantic-enrichment`: embedding generation for indexed entities using the existing embedding API
- `llm-generation`: LLM-backed answer generation and project overview generation
- `codekg-api`: REST API for repo registration, sync, search, entity browsing, and knowledge docs
- `knowledge-docs`: PKB-style repo map and overview generation for indexed repositories

## Impact

- **Storage**: single SQLite file — GORM tables for relational data + sqlite-vec `vec0` virtual table for embeddings
- **Infra**: no Memgraph, no pgvector, no separate vector DB — zero external services
- **Search**: KNN via sqlite-vec `MATCH` when embeddings available, keyword fallback otherwise
- **Sync**: full rebuild-style sync only, not incremental Git-diff sync
- **Follow-up**: advanced graph, pgvector, hybrid retrieval, CLI, and MCP are deferred to `code-rag-kb`
