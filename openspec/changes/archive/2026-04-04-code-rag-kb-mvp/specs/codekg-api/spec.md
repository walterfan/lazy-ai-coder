## ADDED Requirements

### Requirement: REST API for MVP code knowledge base
The system SHALL expose REST endpoints under `/api/v1/codekg/` for the MVP workflow:

- `POST /api/v1/codekg/repos`
- `GET /api/v1/codekg/repos`
- `POST /api/v1/codekg/repos/:id/sync`
- `GET /api/v1/codekg/repos/:id/status`
- `POST /api/v1/codekg/search`
- `GET /api/v1/codekg/entities`
- `GET /api/v1/codekg/repos/:id/knowledge`

#### Scenario: Register a repository
- **WHEN** a POST request to `/api/v1/codekg/repos` with repository metadata is received
- **THEN** the system stores the repository and returns the created repository record

#### Scenario: Trigger async sync
- **WHEN** a POST request to `/api/v1/codekg/repos/:id/sync` is received
- **THEN** the system starts a background sync job and returns a job ID

#### Scenario: Query sync status
- **WHEN** a GET request to `/api/v1/codekg/repos/:id/status` is received
- **THEN** the system returns sync status, processed file counts, and entity counts

#### Scenario: Search indexed knowledge
- **WHEN** a POST request to `/api/v1/codekg/search` with a natural language query is received
- **THEN** the system returns matched entities and an LLM-generated answer

#### Scenario: Browse knowledge docs
- **WHEN** a GET request to `/api/v1/codekg/repos/:id/knowledge` is received
- **THEN** the system returns generated PKB-style knowledge docs for that repository
