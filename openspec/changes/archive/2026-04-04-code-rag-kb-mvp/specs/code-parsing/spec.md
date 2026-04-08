## ADDED Requirements

### Requirement: MVP entity extraction from source files
The system SHALL parse supported source files using the existing Tree-sitter-based parser and extract basic code entities needed for the MVP search flow. Each entity SHALL include ID, name, type, file path, line range, signature, doc string, body, and language.

#### Scenario: Parse a Go file
- **WHEN** a Go source file containing functions or structs/classes is parsed
- **THEN** the system extracts supported entities and stores their metadata for search and browsing

#### Scenario: Parse a Python or JavaScript-family file
- **WHEN** a supported non-Go source file is parsed by the existing parser
- **THEN** the system extracts supported entities without requiring relationship extraction

#### Scenario: Skip unsupported files
- **WHEN** a file with an unsupported extension is encountered during sync
- **THEN** the system skips that file without failing the overall sync
