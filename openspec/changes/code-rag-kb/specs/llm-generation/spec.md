## ADDED Requirements

### Requirement: Context-aware answer generation from retrieved code knowledge
The system SHALL generate natural language answers to user queries using an LLM, with retrieved code knowledge (entities, sub-graphs, code chunks) assembled into a structured prompt. The prompt SHALL include: knowledge graph context (sub-graph relationships in natural language), relevant code snippets (signature + summary + truncated body), call chain paths from the graph, and the original user question.

#### Scenario: Generate answer for a module understanding query
- **WHEN** the user asks "How does user authentication work?" and the retrieval returns relevant function entities and their call graph
- **THEN** the system generates an answer describing the authentication flow with specific function/file references and call chain paths

#### Scenario: Generate answer for an impact analysis query
- **WHEN** the user asks "What would be affected if I change CreateUser's signature?" and the retrieval returns callers and downstream dependencies
- **THEN** the system generates an answer listing all directly and indirectly affected code locations with file paths and function names

### Requirement: Structured prompt template
The system SHALL use a structured prompt template with sections: system role (code knowledge expert), knowledge graph context, relevant code snippets, call chain paths, and user question. The template SHALL instruct the LLM to cite specific function/file locations and explain module relationships.

#### Scenario: Prompt includes graph context
- **WHEN** the retrieval result contains a sub-graph with 5 nodes and 7 edges
- **THEN** the prompt's knowledge graph section includes a natural language description of all nodes and their relationships

### Requirement: Streaming response support
The system SHALL support streaming LLM responses for real-time output in CLI and web interfaces.

#### Scenario: Stream answer to CLI
- **WHEN** a search query is executed via CLI with streaming enabled
- **THEN** the LLM response is streamed token-by-token to stdout as it is generated

### Requirement: Configurable LLM model for generation
The system SHALL allow configuring the LLM model used for answer generation (default `gpt-4o`) separately from the embedding model. Temperature SHALL default to 0.1 for factual code answers.

#### Scenario: Use a different generation model
- **WHEN** the configuration specifies `gpt-4o-mini` as the generation model
- **THEN** the system uses `gpt-4o-mini` for answer generation while using the configured embedding model for embeddings
