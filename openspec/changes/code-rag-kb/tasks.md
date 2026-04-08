## 0. Dependency

This change starts **after** `code-rag-kb-mvp` is accepted and merged. Reuse its user flow, routes, PKB docs, and baseline `internal/codekg/` package instead of re-implementing them.

## 1. Project Setup and Core Models

- [ ] 1.1 Extend `internal/codekg/` with sub-packages: `parser/`, `enricher/`, `store/pgvector/`, `store/memgraph/`, `retriever/`, `generator/`, `syncer/`
- [ ] 1.2 Add full graph-aware models: `CodeEntity`, `CodeRelation`, `ParseResult`, `SubGraph`, and supporting enums/types
- [ ] 1.3 Add dependencies for Memgraph and pgvector stores
- [ ] 1.4 Add advanced `codekg` configuration for embedding, retrieval, graph, and sync settings

## 2. Relationship-Aware Parsing

- [ ] 2.1 Extend the parser to extract relationships (`CALLS`, `IMPLEMENTS`, `IMPORTS`, `CONTAINS`, `EMBEDS`, `DEPENDS_ON`, `RETURNS`, `ACCEPTS`)
- [ ] 2.2 Implement Go-specific relationship extraction
- [ ] 2.3 Implement Python-specific relationship extraction
- [ ] 2.4 Implement TypeScript/JavaScript relationship extraction
- [ ] 2.5 Add directory-level parsing with exclude patterns and concurrency controls
- [ ] 2.6 Add unit tests across supported languages

## 3. Semantic Enrichment

- [ ] 3.1 Add a dedicated semantic enricher on top of the MVP embedding flow
- [ ] 3.2 Add batching, retry, and rate limiting
- [ ] 3.3 Add code summary generation for non-trivial entities
- [ ] 3.4 Add tests with mocked LLM and embedding responses

## 4. Vector Store

- [ ] 4.1 Add PostgreSQL migrations for `code_entities`, `code_chunks`, and indexes
- [ ] 4.2 Implement pgvector-backed upsert and retrieval operations
- [ ] 4.3 Implement ANN similarity search and full-text search
- [ ] 4.4 Implement chunk storage for fine-grained retrieval
- [ ] 4.5 Add integration tests against PostgreSQL with `vector` and `pg_trgm`

## 5. Graph Store

- [ ] 5.1 Implement Memgraph-backed node and edge upserts
- [ ] 5.2 Implement Cypher-based graph search and BFS sub-graph expansion
- [ ] 5.3 Implement delete-by-file and graph refresh helpers for incremental sync
- [ ] 5.4 Add vector-index-based graph filtering where supported
- [ ] 5.5 Add integration tests against Memgraph

## 6. Hybrid Retrieval

- [ ] 6.1 Implement intent analysis for natural language queries
- [ ] 6.2 Run vector, graph, and full-text retrieval in parallel
- [ ] 6.3 Merge results with RRF
- [ ] 6.4 Add graph expansion and embedding-based pruning
- [ ] 6.5 Add partial-failure handling and retrieval tests

## 7. LLM Generation

- [ ] 7.1 Add graph-aware prompt assembly from entities, edges, and code snippets
- [ ] 7.2 Add non-streaming and streaming answer generation
- [ ] 7.3 Add tests with mocked LLM output

## 8. Incremental Sync

- [ ] 8.1 Add Git-diff-based sync orchestration
- [ ] 8.2 Handle added, modified, and deleted files correctly across both stores
- [ ] 8.3 Add resume-on-failure and progress reporting
- [ ] 8.4 Add sync tests for diff classification and entity-level updates

## 9. Interfaces

- [ ] 9.1 Add CLI commands: `codekg parse`, `codekg search`, `codekg sync`
- [ ] 9.2 Extend REST API responses with graph-aware search results and entity relationship detail
- [ ] 9.3 Add MCP tools: `codekg_search` and `codekg_analyze`
- [ ] 9.4 Add MCP resources for indexed repositories and graph context

## 10. Frontend Extension

- [ ] 10.1 Extend `CodeKGView.vue` to visualize relationship graphs and richer analysis results
- [ ] 10.2 Add entity detail views with relationship navigation
- [ ] 10.3 Add state management if the richer UI needs shared store logic

## 11. Deployment and Docs

- [ ] 11.1 Finalize pgvector and Memgraph deployment requirements
- [ ] 11.2 Add full environment variable and setup documentation
- [ ] 11.3 Update README and API docs for CLI, REST, and MCP usage
- [ ] 11.4 Add manual end-to-end validation for hybrid retrieval and incremental sync
