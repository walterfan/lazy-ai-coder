## ADDED Requirements

### Requirement: Store code entities as graph nodes in Memgraph
The system SHALL store code entities as labeled nodes in Memgraph with labels matching entity type (Package, File, Function, Struct, Interface). Each node SHALL have properties: id, name, file_path, signature, doc, summary, and an embedding vector property for graph-level semantic filtering.

#### Scenario: Upsert a function entity as a graph node
- **WHEN** a Function entity is upserted to the graph store
- **THEN** a node with label `:Function` is created or updated in Memgraph with all entity properties, using the entity ID as the merge key

#### Scenario: Upsert with existing node
- **WHEN** an entity with the same ID already exists in the graph
- **THEN** the system updates the existing node's properties without creating a duplicate

### Requirement: Store code relationships as graph edges
The system SHALL store code relationships as typed edges in Memgraph. Supported edge types: CALLS, IMPLEMENTS, IMPORTS, CONTAINS, EMBEDS, DEPENDS_ON, RETURNS, ACCEPTS. Each edge SHALL have weight and context properties.

#### Scenario: Create a CALLS edge between two functions
- **WHEN** a CALLS relationship between function A and function B is upserted
- **THEN** an edge `(A)-[:CALLS {weight: 1.0, context: "..."}]->(B)` is created in Memgraph

### Requirement: Cypher-based graph search
The system SHALL support Cypher queries for graph-based code search, including pattern matching on entity names/types, relationship traversal, and path queries.

#### Scenario: Find all callers of a function
- **WHEN** a Cypher query searches for all nodes connected via incoming CALLS edges to a function named "HandleRequest"
- **THEN** the system returns all Function nodes that call HandleRequest, along with the relationship metadata

#### Scenario: Find implementation chain
- **WHEN** a query asks for all structs implementing interface "Repository"
- **THEN** the system returns all Struct nodes connected via IMPLEMENTS edges to the Repository interface node

### Requirement: Sub-graph expansion via BFS
The system SHALL support expanding a set of seed entities into a sub-graph by performing BFS traversal up to a configurable depth (default 2 hops). The expansion SHALL follow all relationship types.

#### Scenario: Expand from a seed function with 2 hops
- **WHEN** sub-graph expansion is requested from function F with maxHops=2
- **THEN** the system returns F, all entities directly related to F (1-hop), and all entities related to those (2-hop), along with all connecting edges

### Requirement: Delete entities by file path
The system SHALL support deleting all graph nodes and their relationships for a given file path, to support incremental updates when files are deleted or fully replaced.

#### Scenario: Delete all entities from a removed file
- **WHEN** deletion is requested for file path "internal/auth/jwt.go"
- **THEN** all nodes with file_path="internal/auth/jwt.go" and their incoming/outgoing edges are removed from the graph

### Requirement: Memgraph vector index for semantic filtering
The system SHALL create vector indexes on entity node labels in Memgraph to support in-graph semantic similarity filtering using cosine distance.

#### Scenario: Vector similarity search within graph nodes
- **WHEN** a query embedding is provided with a similarity threshold of 0.8
- **THEN** the system returns graph nodes whose embedding has cosine similarity >= 0.8 to the query embedding
