## ADDED Requirements

### Requirement: Embedding vector generation for code entities
The system SHALL generate embedding vectors for code entities using the OpenAI Embeddings API. The embedding input SHALL combine entity language, type, signature, doc string, and truncated body (max 2000 characters). The default model SHALL be `text-embedding-3-large` (3072 dimensions) and SHALL be configurable.

#### Scenario: Generate embedding for a function entity
- **WHEN** a Function entity with signature, doc string, and body is submitted for embedding
- **THEN** the system constructs a combined text input and returns a 3072-dimensional float64 vector

#### Scenario: Configurable embedding model
- **WHEN** the configuration specifies `text-embedding-3-small` as the embedding model
- **THEN** the system uses that model and produces vectors of the corresponding dimension (1536)

### Requirement: Batch embedding generation with rate limiting
The system SHALL support batch embedding generation with configurable batch size (default 100, max 2048 per API call) and built-in rate limiting to respect OpenAI API limits.

#### Scenario: Batch process 500 entities
- **WHEN** 500 code entities are submitted for embedding generation
- **THEN** the system processes them in batches (e.g., 5 batches of 100), respecting rate limits between batches, and returns all 500 embedding vectors

#### Scenario: Rate limit handling
- **WHEN** the OpenAI API returns a 429 rate limit error during batch processing
- **THEN** the system retries with exponential backoff and continues processing remaining batches

### Requirement: Code summary generation
The system SHALL generate natural language summaries for code entities using an LLM chat model (default `gpt-4o-mini`). Summaries SHALL include a one-sentence functional overview, input/output description, and key logic steps.

#### Scenario: Generate summary for a complex function
- **WHEN** a Function entity with a body longer than 10 lines is submitted for summarization
- **THEN** the system returns a structured summary with functional overview, I/O description, and key logic steps

#### Scenario: Skip summary for trivial entities
- **WHEN** a Function entity with a body of 3 lines or fewer is submitted
- **THEN** the system MAY skip LLM summarization and use the doc string as the summary to save API costs
