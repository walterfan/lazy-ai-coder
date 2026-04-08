## ADDED Requirements

### Requirement: CLI sub-commands for code knowledge graph
The system SHALL provide cobra CLI sub-commands under `codekg`: `parse` (parse a repo/directory), `search` (search the knowledge base), and `sync` (incremental sync from Git). Each command SHALL accept relevant flags and output results to stdout.

#### Scenario: Parse a local repository via CLI
- **WHEN** `lazy-ai-coder codekg parse --repo /path/to/repo --languages go,python` is executed
- **THEN** the system parses the repository, extracts entities and relationships, generates embeddings, stores them in pgvector and Memgraph, and prints a summary (entity/relationship counts, time elapsed)

#### Scenario: Search via CLI
- **WHEN** `lazy-ai-coder codekg search "How does authentication work?"` is executed
- **THEN** the system runs the hybrid retrieval pipeline, generates an LLM answer, and prints the answer with code references to stdout

#### Scenario: Incremental sync via CLI
- **WHEN** `lazy-ai-coder codekg sync --repo-id repo_001` is executed
- **THEN** the system computes the Git diff since last sync, processes changed files, and prints a sync summary (added/modified/deleted entity counts)

### Requirement: REST API endpoints for code knowledge base
The system SHALL expose REST endpoints under `/api/v1/codekg/`:
- `POST /api/v1/codekg/search` — search the knowledge base with a natural language query
- `GET /api/v1/codekg/entities` — list/browse code entities with pagination and filters
- `GET /api/v1/codekg/entities/:id` — get a single entity with its relationships
- `POST /api/v1/codekg/repos` — register a repository for indexing
- `POST /api/v1/codekg/repos/:id/sync` — trigger incremental sync
- `GET /api/v1/codekg/repos/:id/status` — get sync status

#### Scenario: Search via REST API
- **WHEN** a POST request to `/api/v1/codekg/search` with body `{"query": "retry logic", "top_k": 10}` is received
- **THEN** the system returns a JSON response with matched entities, their sub-graph relationships, and an LLM-generated answer

#### Scenario: Browse entities via REST API
- **WHEN** a GET request to `/api/v1/codekg/entities?type=function&page=1&per_page=20` is received
- **THEN** the system returns a paginated JSON list of function entities with their metadata

#### Scenario: Trigger sync via REST API
- **WHEN** a POST request to `/api/v1/codekg/repos/repo_001/sync` is received
- **THEN** the system triggers an async incremental sync and returns a job ID for status polling

### Requirement: MCP tools for code knowledge base
The system SHALL register two new MCP tools: `codekg_search` (search the code knowledge base with a natural language query) and `codekg_analyze` (analyze a specific code entity — its callers, callees, implementors, dependencies).

#### Scenario: Search via MCP tool
- **WHEN** the `codekg_search` tool is called with argument `{"query": "How does JWT validation work?"}`
- **THEN** the tool returns a text content item with the LLM-generated answer including code references

#### Scenario: Analyze via MCP tool
- **WHEN** the `codekg_analyze` tool is called with `{"entity_name": "ValidateToken", "analysis": "callers"}`
- **THEN** the tool returns a text content item listing all functions that call ValidateToken with file paths and call context

### Requirement: Docker Compose deployment extension
The system SHALL add Memgraph and Memgraph Lab services to the existing Docker Compose deployment configuration.

#### Scenario: Deploy with Memgraph
- **WHEN** `docker-compose up -d` is executed with the updated compose file
- **THEN** Memgraph (port 7687 for Bolt, port 7444 for HTTP) and Memgraph Lab (port 3000) are running alongside the existing services, and the codekg service connects to Memgraph via Bolt
