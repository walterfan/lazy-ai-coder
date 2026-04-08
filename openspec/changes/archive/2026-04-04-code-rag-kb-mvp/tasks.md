## 0. MVP — End-to-End Pipeline

Thin vertical slice: register repo → parse (Tree-sitter) → embed (OpenAI) → store (SQLite + sqlite-vec) → search (KNN via sqlite-vec) → LLM answer. Reuses existing `internal/rag` and `internal/llm` packages. **No Memgraph, no pgvector, no RRF, no incremental sync.**

- [x] 0.1 Create `internal/codekg/model.go` with GORM models + sqlite-vec virtual table: `Repository`, `Entity`, `KnowledgeDoc`, `codekg_entity_vec`
- [x] 0.2 Create `internal/codekg/service.go` composing existing parser, embedder, and LLM services
- [x] 0.3 Create Gin HTTP handlers for repos, sync, search, entities, and knowledge docs
- [x] 0.4 Register CodeKG routes in `cmd/web.go`
- [x] 0.5 Add frontend `CodeKGView.vue` for repository indexing and search
- [x] 0.6 Generate PKB-inspired knowledge docs after sync: `repo-map` and `overview`
- [x] 0.7 Add frontend "Knowledge Docs" tab for browsing generated docs
- [x] 0.8 Integrate sqlite-vec: store embeddings in `vec0` virtual table, replace brute-force with KNN `MATCH` query
- [x] 0.9 Fix vec0 INSERT: replace `INSERT OR REPLACE` (unsupported by vec0) with DELETE+INSERT; add startLine to entity hash to prevent duplicate IDs
- [x] 0.10 Write basic unit test `internal/codekg/service_test.go` covering parse + keyword search (12 tests, all pass)
- [x] 0.11 E2E test with live LLM gateway: register repo → sync (252 files, 1405 entities, 0 embedding errors) → verify knowledge docs (repo-map 13K chars, overview 3.3K chars) → vector search (KNN returns semantically relevant entities) → LLM answer (accurate, cites files/lines) → frontend smoke test (all 5 actions verified in browser)
