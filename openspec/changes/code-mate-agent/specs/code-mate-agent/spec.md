# code-mate-agent Specification (Delta)

## ADDED Requirements

### Requirement: Code Mate Agent classifies input and generates response

The system SHALL provide an AI agent (Code Mate Agent) that accepts free-form user input, classifies it as one of research_solution, learn_tech, or tech_design, and generates a type-appropriate response using the existing LLM integration without persisting until the user confirms.

#### Scenario: Submit input and receive classification and response

- **WHEN** the user submits input (e.g. a tech question, a topic to learn, or a design problem)
- **THEN** the system classifies the input type and returns a generated response appropriate to that type (research_solution: summary, options, pros/cons, recommendation; learn_tech: intro, concepts, learning path, resources; tech_design: problem, options, chosen approach, components, risks)

#### Scenario: No persistence until user confirms

- **WHEN** the user submits input and receives a response
- **THEN** the system does NOT write to the code-mate artifacts table until the user explicitly confirms to save

### Requirement: Research solution handling – summary, options, recommendation

The system SHALL treat input classified as research_solution by returning a structured research note: summary, options with pros/cons, trade-offs, and a recommendation; and SHALL offer to save only after the user confirms.

#### Scenario: Research input returns summary and recommendation

- **WHEN** the user input is classified as research_solution
- **THEN** the system returns a summary, options (with pros/cons), trade-offs, and a recommendation

#### Scenario: Save research note only on confirm

- **WHEN** the user confirms after viewing the research response
- **THEN** the system persists the input and full response into the code_mate_artifacts table

### Requirement: Learn tech handling – learning path and resources

The system SHALL treat input classified as learn_tech by providing an introduction, key concepts, a learning path, recommended resources (docs, tutorials, courses), prerequisites, and time estimate; and SHALL offer to save only after the user confirms.

#### Scenario: Learn-tech input returns learning plan

- **WHEN** the user input is classified as learn_tech
- **THEN** the system returns introduction, key concepts, learning path, resources, prerequisites, and time estimate

#### Scenario: Save learning plan only on confirm

- **WHEN** the user confirms after viewing the learning plan
- **THEN** the system persists the input and full response into the code_mate_artifacts table

### Requirement: Tech design handling – problem, options, approach, risks

The system SHALL treat input classified as tech_design by producing a structured design: problem statement, approach options, chosen approach, components/APIs, and risks; and SHALL offer to save only after the user confirms.

#### Scenario: Tech-design input returns structured design

- **WHEN** the user input is classified as tech_design
- **THEN** the system returns problem statement, approach options, chosen approach, components/APIs, and risks or mitigations

#### Scenario: Save design artifact only on confirm

- **WHEN** the user confirms after viewing the design
- **THEN** the system persists the input and full response into the code_mate_artifacts table

### Requirement: Confirm-to-save persistence

The system SHALL persist a code-mate artifact (input, type, and full response payload) to the code_mate_artifacts table only when the user explicitly confirms; artifacts SHALL be associated with the authenticated user and realm.

#### Scenario: Confirm creates one artifact

- **WHEN** the user triggers confirm to save with valid input and response payload
- **THEN** the system creates exactly one row in the code_mate_artifacts table with the given input, classification type, response payload, user_id, and realm_id

#### Scenario: Unauthenticated confirm is rejected

- **WHEN** a confirm request is made without valid authentication
- **THEN** the system does not persist and returns an error

### Requirement: Code Mate History – list, filter, search, delete, stats

The system SHALL provide API and UI for users to list their saved artifacts with pagination, filter by type (research_solution, learn_tech, tech_design), search by input or response content, view a single artifact, soft-delete an artifact, and view summary statistics.

#### Scenario: List artifacts for current user

- **WHEN** the user requests a paginated list of artifacts with optional type and search filters
- **THEN** the system returns only that user's artifacts matching the filters

#### Scenario: Get artifact detail and stats

- **WHEN** the user requests a single artifact by id (with valid ownership) or requests stats
- **THEN** the system returns the artifact detail or aggregate counts by type and streak

#### Scenario: Soft-delete artifact

- **WHEN** the user requests delete for an artifact they own
- **THEN** the system soft-deletes the artifact and returns success

### Requirement: Code Mate Agent may use assets (rules, commands, skills)

The system SHALL allow the Code Mate Agent to use content from a configurable **assets** folder (default `assets/`), comprising **rules** (e.g. `assets/rules/`), **commands** (e.g. `assets/commands/`), and **skills** (e.g. `assets/skills/` with `SKILL.md` per skill). When generating a response, the agent MAY include selected or auto-selected assets in its context so that research, learning, and design answers are grounded in project conventions and skill content.

#### Scenario: Submit with selected assets

- **WHEN** the user (or client) submits input and optionally specifies which rules, commands, or skills to include (e.g. by key or path)
- **THEN** the system loads the content of those assets and injects them into the agent context, and the generated response MAY reflect those rules or skills (e.g. cite project conventions or skill guidelines)

#### Scenario: List available assets

- **WHEN** the user or client requests a list of available assets (e.g. via GET /api/v1/code-mate/assets)
- **THEN** the system returns a list of discoverable rules, commands, and skills (e.g. id/key, name, path, optional language or tags) so the caller can choose which to include on submit

#### Scenario: Submit without assets

- **WHEN** the user submits input without selecting any assets
- **THEN** the system generates a response using only the default agent prompt (no asset context); behavior is unchanged from the case where no assets exist
