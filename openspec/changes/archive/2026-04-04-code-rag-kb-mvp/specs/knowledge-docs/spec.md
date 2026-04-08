## ADDED Requirements

### Requirement: PKB-style repository map generation
The system SHALL generate a `repo-map` knowledge document after sync without requiring an LLM. The document SHALL summarize repository structure using indexed files and entities.

#### Scenario: Generate repo map after sync
- **WHEN** sync completes successfully
- **THEN** the system stores a `repo-map` document including directory structure summary, language breakdown, entity counts, and key entry points

### Requirement: Knowledge docs are repository-scoped artifacts
The system SHALL store generated knowledge docs per repository and make them retrievable via the API.

#### Scenario: Retrieve docs for a repository
- **WHEN** knowledge docs are requested for a repository
- **THEN** the system returns all generated docs for that repository, including `repo-map` and `overview` when available
