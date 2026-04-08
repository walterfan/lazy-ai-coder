## ADDED Requirements

### Requirement: LLM answer generation from matched entities
The system SHALL generate a natural language answer for search queries using the configured LLM and the matched code entities as context.

#### Scenario: Answer a code understanding query
- **WHEN** a search query returns relevant entities
- **THEN** the system generates an answer that cites specific files and entity names from the retrieved context

### Requirement: LLM-generated repository overview
The system SHALL generate a project overview document from indexed repository metadata and sampled entities when the LLM is configured.

#### Scenario: Generate overview after sync
- **WHEN** sync completes successfully and LLM configuration is available
- **THEN** the system generates an `overview` knowledge document summarizing purpose, stack, structure, and entry points

#### Scenario: Fallback overview without LLM
- **WHEN** sync completes successfully but LLM configuration is unavailable
- **THEN** the system stores a fallback overview stating that AI-generated overview content is unavailable
