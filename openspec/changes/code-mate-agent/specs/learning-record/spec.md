# chat-record Specification (Delta)

## REMOVED Requirements

**Reason**: The Learning Record capability is replaced by the Code Mate Agent. The previous behavior (word, sentence, question, idea, topic classification and chat-record persistence) is removed in favor of research_solution, learn_tech, and tech_design with code_mate_artifacts persistence.

**Migration**: API path changes from `/api/v1/chat-record/*` to `/api/v1/code-mate/*`. Table renames from `learning_records` to `code_mate_artifacts`. Existing data is migrated via migration script (input_type values mapped to new types). Clients must use the new Code Mate API and types.

### Requirement: Learning record agent classifies input and generates response

**Reason**: Replaced by Code Mate Agent (research_solution, learn_tech, tech_design).

### Requirement: Word handling – Chinese explanation, pronunciation, and usage example

**Reason**: Replaced by Code Mate types; no word-specific flow.

### Requirement: Sentence handling – Chinese explanation and usage example

**Reason**: Replaced by Code Mate types; no sentence-specific flow.

### Requirement: Question handling – answer and record on confirm

**Reason**: Replaced by Code Mate research_solution / learn_tech flows.

### Requirement: Idea handling – executable plan and record on confirm

**Reason**: Replaced by Code Mate tech_design flow.

### Requirement: Topic handling – learning plan, concepts, and resources

**Reason**: Replaced by Code Mate learn_tech flow.

### Requirement: Confirm-to-record persistence

**Reason**: Replaced by Code Mate confirm-to-save persistence to code_mate_artifacts.
