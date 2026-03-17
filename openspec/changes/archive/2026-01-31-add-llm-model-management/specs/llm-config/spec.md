## ADDED Requirements

### Requirement: LLM model management page
The system SHALL provide a dedicated UI page for managing LLM model configurations, separate from the legacy Settings page.

#### Scenario: Navigate to model management
- **WHEN** the user opens the LLM model management page from the **Tools** menu (or via a link in the Settings page)
- **THEN** the system displays a list of configured LLM models for the current realm/user

### Requirement: CRUD + lifecycle management for LLM models
The system SHALL allow users to create, update, delete, enable/disable, and set a default LLM model configuration.

#### Scenario: Set default model
- **WHEN** the user marks a model as default
- **THEN** the system persists the default flag and ensures only one model is default within the applicable scope

### Requirement: Backward compatible legacy settings
The system SHALL keep the existing Settings page and legacy single-model localStorage settings working as a fallback.

#### Scenario: No model selected / models unavailable
- **WHEN** no managed LLM model is selected (or the model list is unavailable)
- **THEN** requests use the legacy Settings configuration (API key + model/base_url/temperature) without breaking existing behavior

