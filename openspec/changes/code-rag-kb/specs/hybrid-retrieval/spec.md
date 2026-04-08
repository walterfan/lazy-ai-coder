## ADDED Requirements

### Requirement: Query intent analysis
The system SHALL analyze natural language queries using an LLM to extract: entity names mentioned, relationship types relevant to the query, search keywords, and query intent category (e.g., "find code", "trace call chain", "impact analysis", "explain module").

#### Scenario: Analyze a relationship query
- **WHEN** the query is "What functions does HandleRequest call?"
- **THEN** the system extracts entity name "HandleRequest", relationship type "CALLS", and intent "trace call chain"

#### Scenario: Analyze a semantic search query
- **WHEN** the query is "Is there any retry logic in the project?"
- **THEN** the system extracts keywords ["retry", "backoff", "retries"], no specific entity names, and intent "find code"

### Requirement: Three-channel parallel retrieval
The system SHALL execute three retrieval channels in parallel for each query: (1) vector similarity search via pgvector, (2) graph traversal via Memgraph Cypher queries, and (3) full-text search via PostgreSQL. Each channel SHALL return scored entity lists.

#### Scenario: Execute parallel retrieval
- **WHEN** a search query with query embedding and parsed intent is submitted with topK=20
- **THEN** all three channels execute concurrently and return their individual ranked result lists within a combined timeout

#### Scenario: Partial channel failure
- **WHEN** the Memgraph graph channel is unavailable
- **THEN** the system proceeds with results from vector and full-text channels and logs a warning

### Requirement: Reciprocal Rank Fusion (RRF) result merging
The system SHALL merge results from all retrieval channels using RRF with constant k=60. The formula SHALL be: `score(entity) = Σ 1/(k + rank + 1)` across all channel result lists. The final results SHALL be sorted by descending fused score.

#### Scenario: Merge results from three channels
- **WHEN** vector channel returns [A, B, C], graph channel returns [B, D, A], and full-text returns [C, A, E]
- **THEN** entity A appears in all three lists and receives the highest RRF score; the merged list is ordered by descending RRF score

### Requirement: Sub-graph expansion and pruning
The system SHALL expand merged Top-K seed entities by retrieving their sub-graph from Memgraph (1-2 hop BFS). Expanded nodes SHALL be pruned by computing embedding similarity to the query — only nodes above a configurable relevance threshold (default 0.5) are retained.

#### Scenario: Expand and prune from seed results
- **WHEN** the top 10 merged entities are used as seeds with maxHops=2 and relevanceThreshold=0.5
- **THEN** the system returns the seed entities plus all graph-expanded entities whose embedding similarity to the query exceeds 0.5

### Requirement: Configurable search options
The system SHALL accept search configuration including: topK (default 20), maxHops for graph expansion (default 2), relevance threshold (default 0.5), enabled channels (vector/graph/text, default all), and entity type filters.

#### Scenario: Search with custom options
- **WHEN** a search is executed with topK=10, maxHops=1, and entity type filter "function"
- **THEN** only Function entities are returned, graph expansion is limited to 1 hop, and at most 10 final results are returned
