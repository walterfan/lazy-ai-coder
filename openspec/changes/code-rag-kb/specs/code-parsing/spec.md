## ADDED Requirements

### Requirement: Multi-language AST parsing with entity extraction
The system SHALL parse source code files using Tree-sitter and extract structured code entities including Package, File, Function, Struct/Class, Interface, and Variable/Constant types. Each entity SHALL have a unique ID, name, file path, line range, signature, doc string, body content, and metadata map.

#### Scenario: Parse a Go source file with functions and structs
- **WHEN** a Go file containing functions and struct declarations is parsed
- **THEN** the system extracts each function as a `Function` entity with name, signature, parameter list, return types, and doc comment; and each struct as a `Struct` entity with name, fields list, and associated methods

#### Scenario: Parse a Python file with classes and methods
- **WHEN** a Python file containing class definitions with methods is parsed
- **THEN** the system extracts each class as a `Class` entity and each method as a `Function` entity linked to its parent class

#### Scenario: Unsupported language file
- **WHEN** a file with an unsupported language extension is encountered during directory parsing
- **THEN** the system SHALL skip the file and log a warning without failing the overall parse operation

### Requirement: Relationship extraction from AST
The system SHALL extract typed relationships between code entities including CALLS, IMPLEMENTS, IMPORTS, CONTAINS, INHERITS/EMBEDS, DEPENDS_ON, RETURNS, and ACCEPTS. Each relationship SHALL have source ID, target ID, type, weight, and context string.

#### Scenario: Extract function call relationships in Go
- **WHEN** function A contains a call expression to function B within its body
- **THEN** the system creates a CALLS relationship from entity A to entity B with weight 1.0

#### Scenario: Extract interface implementation in Go
- **WHEN** struct S implements all methods declared in interface I
- **THEN** the system creates an IMPLEMENTS relationship from S to I

#### Scenario: Extract import relationships
- **WHEN** file F imports package P
- **THEN** the system creates an IMPORTS relationship from the File entity to the Package entity

#### Scenario: Extract containment relationships
- **WHEN** a package directory contains Go source files with function declarations
- **THEN** the system creates CONTAINS relationships from Package to File entities and from File to Function/Struct entities

### Requirement: Directory-level batch parsing
The system SHALL support parsing an entire directory tree with configurable options including file extension filters, exclude patterns (e.g., vendor, node_modules, test files), and concurrency level.

#### Scenario: Parse a Go project directory
- **WHEN** `ParseDirectory` is called with a Go project root path and default options
- **THEN** all `.go` files (excluding vendor/ and *_test.go if configured) are parsed, and the combined list of entities and relationships is returned as a `ParseResult`

#### Scenario: Parse with exclude patterns
- **WHEN** `ParseDirectory` is called with exclude patterns `["vendor/**", "**/*_test.go"]`
- **THEN** files matching those patterns are skipped during parsing

### Requirement: Supported languages
The system SHALL support parsing Go, Python, TypeScript, Java, JavaScript, and C++ using Tree-sitter grammars.

#### Scenario: Language detection from file extension
- **WHEN** a file with extension `.go`, `.py`, `.ts`, `.java`, `.js`, or `.cpp`/`.cc`/`.h` is encountered
- **THEN** the system correctly identifies the language and uses the corresponding Tree-sitter grammar
