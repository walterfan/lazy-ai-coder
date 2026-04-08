## Why

The MVP change provides a thin vertical slice for repository registration, AST-based entity extraction, embedding-backed search, PKB-style repo documents, and a verification UI. That MVP proves the user workflow, but it does not yet provide the core differentiators needed for a production-grade **codebase RAG + knowledge base platform**:

- no explicit relationship graph between code entities
- no durable vector-native storage for scalable retrieval
- no hybrid retrieval across vector, graph, and text channels
- no incremental sync from Git history
- no MCP-native structured access for AI agents

This follow-up change turns the MVP into the full system originally envisioned in `doc/code-rag-kb.md`: a hybrid GraphRAG platform for both human developers and AI agents.

## What Changes

This change **builds on** the separate MVP change `code-rag-kb-mvp`.

- extend code parsing from flat entity extraction to relationship extraction (`CALLS`, `IMPLEMENTS`, `IMPORTS`, `CONTAINS`, `EMBEDS`, `DEPENDS_ON`, `RETURNS`, `ACCEPTS`)
- add a dual-store architecture:
  - `pgvector` for persistent vector + full-text retrieval
  - `Memgraph` for graph traversal and sub-graph expansion
- add chunk storage for fine-grained retrieval from large entities
- replace brute-force search with hybrid retrieval:
  - vector similarity
  - graph traversal
  - PostgreSQL full-text search
  - Reciprocal Rank Fusion (RRF)
- add graph-guided expansion and pruning for richer LLM context assembly
- add Git diff-based incremental sync instead of full rebuild-only sync
- add CLI commands for parse/search/sync workflows
- add MCP tools and resources so AI agents can query the code knowledge base directly
- extend the REST API with richer entity detail, sub-graph, and analysis responses
- integrate the full feature set with deployment and configuration for pgvector + Memgraph

## Capabilities

### New Capabilities
- `graph-store`: Memgraph graph storage for code entities/relationships with Cypher queries and vector indexes
- `vector-store`: pgvector storage for code embeddings with HNSW indexing, full-text search, and code chunk management
- `hybrid-retrieval`: three-channel hybrid search engine with RRF fusion, sub-graph expansion, and relevance pruning
- `incremental-sync`: Git diff-based incremental knowledge graph updates for added/modified/deleted files

### Extended Capabilities
- `code-parsing`: extend MVP parsing from entities-only to entities + relationships
- `semantic-enrichment`: extend MVP embedding flow with code summarization, batching, and retry/rate-limit control
- `llm-generation`: extend MVP answer generation with graph-aware prompt construction and streaming
- `codekg-api`: extend MVP REST API with CLI commands, MCP tools, MCP resources, and graph-aware responses

## Impact

- **Depends on**: `code-rag-kb-mvp`
- **New Go dependencies**: `github.com/neo4j/neo4j-go-driver/v5`, `github.com/pgvector/pgvector-go`
- **Database**: PostgreSQL with `vector` and `pg_trgm` extensions; new `code_entities`, `code_chunks`, and related indexes
- **Infrastructure**: Memgraph + Memgraph Lab become required for full hybrid retrieval
- **API surface**: richer `/api/v1/codekg/*` responses plus MCP tools/resources
- **CLI**: add `lazy-ai-coder codekg parse|search|sync`
- **Config**: add advanced retrieval, storage, and sync configuration
- **Cost/ops**: higher operational complexity than MVP due to external stores and async sync pipeline
