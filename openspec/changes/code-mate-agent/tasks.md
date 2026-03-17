# Tasks: Code Mate Agent (Refactor from Learning Record)

**Methodology**: Incremental refactor with TDD verification at each step.

**Approach**: Database migration first (rename table, migrate input_type), then backend (model → repository → service → handlers/routes → agent), then frontend. Verify each phase with tests.

**Note**: Diagrams (`.puml`, `.png`) have been removed from this change folder; architecture and sequence flows are described inline in `design.md`.

---

## Phase 1: Database Migration

### 1.1 Rename Table and Migrate input_type

- [x] 1.1.1 Add migration (e.g. `deploy/db/migrations/010_rename_learning_records_to_code_mate_artifacts.sql`):
  - Rename table `learning_records` → `code_mate_artifacts` (or create new table, copy data, drop old per DB engine).
  - Update `input_type` values: map `word`/`sentence`/`question` → `learn_tech` (or `research_solution` per product rule); `idea` → `tech_design`; `topic` → `learn_tech`.
- [x] 1.1.2 Document migration for SQLite, PostgreSQL, MySQL if applicable.

### 1.2 Verify Migration

- [ ] 1.2.1 Run migration against dev DB; verify table name and column types.
- [ ] 1.2.2 Verify existing rows have valid new `input_type` values.

**Checkpoint**: Table `code_mate_artifacts` exists; legacy data migrated.

---

## Phase 2: Backend – Model and Constants

### 2.1 Rename Model and Input Types

- [x] 2.1.1 Rename `pkg/models/learning_record.go` → `pkg/models/code_mate_artifact.go` (or add new file and remove old after migration).
- [x] 2.1.2 Rename struct `LearningRecord` → `CodeMateArtifact`; keep fields compatible with table (id, input_type, user_input, response_payload, user_id, realm_id, timestamps).
- [x] 2.1.3 Replace InputType constants: `InputTypeWord`, `InputTypeSentence`, `InputTypeQuestion`, `InputTypeIdea`, `InputTypeTopic` → `InputTypeResearchSolution`, `InputTypeLearnTech`, `InputTypeTechDesign`.
- [x] 2.1.4 Update `ResponsePayloadData` (or equivalent) to match new response shapes (research summary, learning path, design) per `design.md`.
- [x] 2.1.5 Register model in `pkg/models/models.go` (update `GetAllModels()` if needed).

### 2.2 Verify Model

- [x] 2.2.1 Unit test: create CodeMateArtifact, save, read back; verify input_type values.
- [ ] 2.2.2 Run app and confirm table name used by GORM matches migrated table.

**Checkpoint**: Model and constants align with Code Mate types.

---

## Phase 3: Backend – Repository Layer

### 3.1 Rename Repository

- [ ] 3.1.1 In `internal/learningrecord/` (or new `internal/codemate/`): rename interface `LearningRecordRepository` → `CodeMateArtifactRepository` (or `Repository` in codemate package).
- [ ] 3.1.2 Rename implementation struct and all method signatures (Create, FindByID, FindByUserWithFilters, SoftDelete, CountByType, CountStreak, FindSimilar) to use `CodeMateArtifact`.
- [ ] 3.1.3 Update `ListFilters` to use new type filter values (research_solution, learn_tech, tech_design).
- [ ] 3.1.4 Update any raw SQL or GORM table names to `code_mate_artifacts`.

### 3.2 Verify Repository Tests

- [ ] 3.2.1 Update and run repository unit tests for new types and table name.
- [ ] 3.2.2 All CRUD and filter tests pass.

**Checkpoint**: Repository works with `code_mate_artifacts` and new input types.

---

## Phase 4: Backend – Service Layer

### 4.1 Rename Service and Methods

- [ ] 4.1.1 Rename `LearningRecordService` → `CodeMateService` (or keep package as learningrecord and rename type; prefer consistency with route name).
- [ ] 4.1.2 Rename methods: `CreateRecord` → `ConfirmArtifact` (or equivalent), `ListRecords` → `ListArtifacts`, `GetRecord` → `GetArtifact`, `DeleteRecord` → `DeleteArtifact`; keep `SubmitInput`, `GetStats`, `FindSimilar` (signatures may keep same names, internal types change).
- [ ] 4.1.3 Update all references to LearningRecord / learning_record to CodeMateArtifact / code_mate_artifact.
- [ ] 4.1.4 Ensure session memory and similarity retriever use new types and artifact wording where exposed.

### 4.2 Verify Service Tests

- [ ] 4.2.1 Update service unit tests for new types and method names.
- [ ] 4.2.2 All service tests pass.

**Checkpoint**: Service layer uses Code Mate types and naming.

---

## Phase 5: Backend – Handlers and Routes

### 5.1 Rename Handlers and API Path

- [ ] 5.1.1 Rename `internal/handlers/learning_record_handlers.go` → `internal/handlers/code_mate_handlers.go`; rename handler struct to `CodeMateHandlers` (or equivalent).
- [ ] 5.1.2 Change route prefix from `/api/v1/chat-record/` to `/api/v1/code-mate/` in `cmd/web.go`.
- [ ] 5.1.3 Update request/response struct names and JSON tags (e.g. `SubmitLearningRecordResponse` → `SubmitCodeMateResponse`); keep payload structure aligned with new response types.
- [ ] 5.1.4 Add/update auth middleware on all code-mate routes.

### 5.2 Verify API Tests

- [ ] 5.2.1 Update handler/API tests to use `/api/v1/code-mate/*` and new input_type values (research_solution, learn_tech, tech_design).
- [ ] 5.2.2 Acceptance tests: submit (each type), confirm, list, get, delete, stats.
- [ ] 5.2.3 All API tests pass.

**Checkpoint**: API is exposed under `/api/v1/code-mate/` and uses new types.

---

## Phase 6: Backend – Agent (Eino)

### 6.1 Update Classification and Prompts

- [x] 6.1.1 In `internal/learningrecord/agent_eino.go` (or `internal/codemate/agent_eino.go`): update classification prompt to output one of `research_solution`, `learn_tech`, `tech_design`.
- [x] 6.1.2 Update type-specific response generation prompts for:
  - research_solution: summary, options (pros/cons), trade-offs, recommendation.
  - learn_tech: intro, key concepts, learning path, resources, prerequisites, time estimate.
  - tech_design: problem statement, options, chosen approach, components/APIs, risks.
- [x] 6.1.3 Ensure `Process` returns new input types and response payload shape; update `ResponsePayloadData` (or equivalent) parsing.
- [x] 6.1.4 Update similarity/context wording from "learning record" to "artifact" or "saved item" in prompts if present.

### 6.2 Verify Agent Tests

- [ ] 6.2.1 Update agent unit tests for new classification labels and response shapes.
- [ ] 6.2.2 Mock LLM tests for research_solution, learn_tech, tech_design.
- [ ] 6.2.3 All agent tests pass.

**Checkpoint**: Agent classifies and generates Code Mate response types.

---

## Phase 7: Assets Integration (Rules, Commands, Skills)

### 7.1 Asset Loader

- [ ] 7.1.1 Add config for assets path (e.g. `config.yaml` or env `ASSETS_PATH`; default `assets/` relative to app root).
- [ ] 7.1.2 Implement asset loader that lists and reads content from:
  - `assets/rules/` (recursive: .mdc, .md; optional grouping by language e.g. common/, golang/, java/).
  - `assets/commands/` (.md files).
  - `assets/skills/` (each subdir: read SKILL.md; optional config.json for name/description).
- [ ] 7.1.3 Expose a function to load selected assets by key/path and return concatenated content (with optional size limit for context).

### 7.2 List Assets API (optional)

- [ ] 7.2.1 Add GET `/api/v1/code-mate/assets` (or `/api/v1/code-mate/assets/list`): returns list of available rules, commands, skills (id/key, name, path, optional language/tags). Auth required.
- [ ] 7.2.2 Document response shape; add tests.

### 7.3 Agent Integration

- [ ] 7.3.1 Extend submit request model to accept optional `asset_rules[]`, `asset_commands[]`, `asset_skills[]` (or single `assets` object with rule/command/skill keys).
- [ ] 7.3.2 In service/handler: when assets are requested, call asset loader and inject returned content into agent context (e.g. append to system prompt or a dedicated context message).
- [ ] 7.3.3 Enforce context size limit (e.g. truncate or skip assets if total exceeds token/char limit).
- [ ] 7.3.4 Update agent to accept optional context string and include it in the prompt passed to the LLM.

### 7.4 Verify

- [ ] 7.4.1 Unit test: asset loader lists and loads from fixtures; respects size limit.
- [ ] 7.4.2 Integration test: submit with asset keys; response reflects content from selected rule/skill (e.g. mention project convention or skill guideline).
- [ ] 7.4.3 Submit without assets: behavior unchanged (no asset context).

**Checkpoint**: Code Mate Agent can use selected rules, commands, and skills from assets folder.

---

## Phase 8: Frontend – Store, Views, Types, Routes

### 8.1 Rename Store and Types

- [ ] 8.1.1 Rename `web/src/stores/learningRecordStore.ts` → `codeMateStore.ts`; rename store to `codeMateStore`; update state/actions/getters to use "artifact" and new input_type values.
- [ ] 8.1.2 Rename `web/src/types/learningRecord.ts` → `codeMate.ts` (or merge into existing types); update interfaces to research_solution | learn_tech | tech_design and new response payload shapes.
- [ ] 8.1.3 Update `web/src/types/index.ts` exports.
- [ ] 8.1.4 Point store to `/api/v1/code-mate/` endpoints.
- [ ] 8.1.5 (Optional) Add state/actions for available assets and selected assets; call GET `/api/v1/code-mate/assets`; pass selected asset keys on submit.

### 8.2 Rename Views and Router

- [ ] 8.2.1 Rename `web/src/views/LearningRecordView.vue` → `CodeMateView.vue`; update labels ("Confirm to record" → "Confirm to save"), type badges, and response sections for research/learn/design.
- [ ] 8.2.2 Rename `web/src/views/LearningHistoryView.vue` → `CodeMateHistoryView.vue`; update title "Learning History" → "Code Mate History"; update type filters to research_solution, learn_tech, tech_design.
- [ ] 8.2.3 Update router: e.g. `/learning` → `/code-mate`, `/learning/history` → `/code-mate/history`; component imports.
- [ ] 8.2.4 Update NavigationBar (and any other nav): "Learning Assistant" → "Code Mate" (or equivalent); link to `/code-mate`.
- [ ] 8.2.5 (Optional) In CodeMateView, add UI to list and select rules/commands/skills to include in the next submit (dropdown or multi-select).

### 8.3 Verify Frontend Build

- [ ] 8.3.1 TypeScript type-check passes.
- [ ] 8.3.2 Vite build succeeds.
- [ ] 8.3.3 Manual smoke test: submit, confirm, list, filter, delete; optional: submit with assets selected.

**Checkpoint**: UI and routes use Code Mate naming and new types.

---

## Phase 9: Metrics and Documentation

### 9.1 Rename Metrics

- [ ] 9.1.1 Rename Prometheus metrics from `learning_record_*` to `code_mate_*` (submit_total, submit_duration_seconds, confirm_total, confirm_duration_seconds, classification_total, error_total).
- [ ] 9.1.2 Update handler instrumentation and any dashboards/alerts references.

### 9.2 API Documentation

- [ ] 9.2.1 Update Swagger annotations for code-mate handlers and paths (including optional GET /assets and submit body asset fields).
- [ ] 9.2.2 Regenerate: `swag init -g main.go` (or equivalent).

### 9.3 Verify

- [ ] 9.3.1 `/metrics` exposes `code_mate_*` metrics.
- [ ] 9.3.2 Swagger UI shows `/api/v1/code-mate/*` endpoints.

**Checkpoint**: Observability and docs reflect Code Mate.

---

## Summary: Refactor Order

| Phase | Focus | Verification |
|-------|--------|---------------|
| 1 | DB: rename table, migrate input_type | Migration runs; data valid |
| 2 | Model + constants | Unit test; GORM table name |
| 3 | Repository | Unit tests CRUD + filters |
| 4 | Service | Unit tests |
| 5 | Handlers + routes | API tests; path /code-mate |
| 6 | Agent (Eino) | Unit tests; new types + prompts |
| 7 | Assets (loader, list API, agent context) | Unit + integration tests; submit with assets |
| 8 | Frontend (store, views, types, routes) | Build + smoke test |
| 9 | Metrics + Swagger | /metrics, Swagger UI |

---

## Acceptance Test Cases (Quick Reference)

| ID | Endpoint | Test |
|----|----------|------|
| AT-1 | POST /code-mate/confirm | Valid auth → 201, artifact created |
| AT-2 | POST /code-mate/confirm | No auth → 401 |
| AT-3 | GET /code-mate/list | Returns user's artifacts |
| AT-4 | GET /code-mate/list?type=research_solution | Filters by type |
| AT-5 | GET /code-mate/:id | Returns artifact detail |
| AT-6 | DELETE /code-mate/:id | Soft-deletes, returns 204 |
| AT-7 | GET /code-mate/stats | Returns counts by type |
| AT-8 | POST /code-mate/submit (research) | Returns input_type=research_solution |
| AT-9 | POST /code-mate/submit (learn) | Returns input_type=learn_tech |
| AT-10 | POST /code-mate/submit (design) | Returns input_type=tech_design |
| AT-11 | POST /code-mate/submit | Does NOT write to DB |
| AT-12 | Submit → Confirm | End-to-end flow; artifact in list |
