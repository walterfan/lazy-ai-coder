# Change: Code Mate Agent (Rename/Refactor from Learning Record)

## Why

Users need a dedicated AI companion to research tech solutions, learn new technologies, and create tech designs for programming—without mixing this with general chat or one-off prompts. The existing **Learning Record** feature (word/sentence/question/idea/topic) is being **renamed and refactored** into the **Code Mate Agent** with three intents: research solution, learn tech, and tech design. This change repurposes the existing flow, table, and UI rather than building net-new.

## Development Approach

- **TDD (Test-Driven Development)**: Write tests FIRST (red), then implement (green), then refactor. See `tasks.md` for test-first task order.
- **MDD (Metrics-Driven Development)**: Define success metrics upfront, instrument during implementation, validate before release. See `design.md` for metrics definitions and thresholds.

## What Changes

- **Rename/refactor** the existing Learning Record agent, data model, API, and UI into **Code Mate Agent**.
- **Classification and response types** change from word / sentence / question / idea / topic to **research_solution**, **learn_tech**, **tech_design**:
  - **Research solution**: Summarize options, pros/cons, trade-offs, recommendations; offer to save on confirm.
  - **Learn tech**: Introduction, key concepts, learning path, resources, prerequisites, time estimate; offer to save on confirm.
  - **Tech design**: Problem statement, approach options, chosen approach, components/APIs, risks; offer to save on confirm.
- **Database**: Rename table `learning_records` → `code_mate_artifacts`; update `input_type` values via migration. Existing rows are migrated (e.g. map old types to the closest new type or a default).
- **Backend**: Rename model (e.g. `LearningRecord` → `CodeMateArtifact`), repository, service, handlers, and routes; change API path from `/api/v1/chat-record/` to `/api/v1/code-mate/`; update Eino agent classification and prompts.
- **Frontend**: Rename store, views, types, and routes (e.g. `learningRecordStore` → `codeMateStore`, `LearningRecordView` → `CodeMateView`, `/learning` → `/code-mate`).
- **Metrics**: Rename from `learning_record_*` to `code_mate_*`.
- **Assets integration**: The Code Mate Agent SHALL be able to use multiple **skills**, **commands**, and **rules** from the **assets** folder (configurable path, default `assets/`) when generating responses. Relevant assets (e.g. by language, topic, or user selection) are loaded and injected into the agent context so research, learning, and design answers are grounded in project conventions and skill content.
- **BREAKING**: API path and request/response shapes for this feature change; clients using `/api/v1/chat-record/*` must switch to `/api/v1/code-mate/*`. Frontend routes change.

## Impact

- **Affected specs**: Capability `chat-record` is **REMOVED** (replaced); new capability `code-mate-agent` is **ADDED**. See `specs/chat-record/spec.md` (REMOVED) and `specs/code-mate-agent/spec.md` (ADDED).
- **Affected code**:
  - Backend: Rename/refactor `pkg/models/learning_record.go` → `code_mate_artifact.go` (or equivalent); `internal/learningrecord/` → `internal/codemate/` (or keep package name and rename types); `internal/handlers/learning_record_handlers.go` → `code_mate_handlers.go`; routes and handler wiring in `cmd/web.go`; Eino agent classification and prompts; **asset loader** (read `assets/rules`, `assets/commands`, `assets/skills`) and inject selected assets into agent context on submit.
  - Frontend: Rename store, views, types, router paths, and nav links (Learning Record → Code Mate).
  - Database: Migration to rename table and migrate `input_type` values.
- **Database schema**: Table `code_mate_artifacts` with id, input_type (research_solution | learn_tech | tech_design), user_input, response_payload, user_id, realm_id, timestamps.
- **Breaking**: Yes — API path and response types change; old chat-record types (word/sentence/question/idea/topic) are removed.
