## Context

The current system only supports a single LLM configuration through localStorage-based settings. Users must manually edit API keys, model names, and base URLs each time they want to switch between different LLM providers (GPT-4, Claude, Gemini, Qwen, DeepSeek, etc.). This creates friction and limits experimentation with different models.

The LLM Model Management feature will allow users to:
1. Configure multiple LLM models upfront with their respective settings
2. Quickly switch between models via a dropdown selector
3. Enable/disable models without deleting configuration
4. Mark one model as default for new sessions

**Constraints:**
- API keys remain in localStorage for security (not persisted to database)
- Must maintain backward compatibility with existing single-model settings
- Pre-configured models should be realm-scoped (shared by organization)
- User-specific models are private

**Stakeholders:**
- End users (developers using the AI assistant)
- System administrators (managing organizational models)

## Goals / Non-Goals

**Goals:**
- Allow users to configure and manage multiple LLM models
- Provide quick model switching without reconfiguration
- Pre-populate common LLM configurations (Qwen, Claude, GPT, etc.)
- Keep API keys secure (localStorage only, not in DB)
- Maintain backward compatibility with existing settings
- Keep the existing Settings page stable for legacy single-model configuration

**Non-Goals:**
- API key management/storage in database (security concern)
- Automatic model selection based on task type
- Cost tracking or usage metering per model
- Model performance comparison features

## Decisions

### Decision 1: Database Storage for Model Configurations
**What**: Store LLM model configurations (base_url, model name, temperature, max_tokens) in database table, but keep API keys in localStorage.

**Why**: 
- Enables multi-model management without constant manual reconfiguration
- Allows pre-seeding organizational models
- Maintains security by keeping secrets out of database

**Alternatives considered**:
- Store everything in localStorage: Rejected - doesn't scale, no sharing, lost on clear
- Store API keys in DB encrypted: Rejected - security risk, key management complexity

### Decision 1.1: Per-Model API Keys (stored locally)
**What**: Each LLM model can have its own API key stored in localStorage, keyed by model ID. If no key is set for a model, the system falls back to the default API key from Settings.

**Why**:
- Different LLM providers require different API keys (OpenAI vs Anthropic vs Google)
- Same provider might have different keys for different accounts/projects
- Users might want to use different API keys for cost allocation or access control
- Keeps secrets secure (never sent to server, stored only in browser)

**Implementation**:
- localStorage key: `llm_model_api_keys` stores `{ modelId: apiKey }` mapping
- UI shows a key icon badge for models with custom API keys
- Form allows setting/clearing API key per model
- When making requests, use model-specific key if set, else use legacy Settings key

### Decision 1.5: Dedicated UI Page (Keep Settings for Backward Compatibility)
**What**: Implement LLM model CRUD on a new page/route (e.g. “LLM Models”) reachable from the **Tools** menu, with an optional link in the Settings page, and keep the existing Settings page unchanged as the legacy single-model configuration surface.

**Why**:
- Avoids breaking/rewiring existing Settings UX and localStorage flows
- Reduces risk of regressions on a widely-used page
- Allows gradual migration: users can keep using legacy settings while trying multi-model
- Keeps feature discoverability consistent by grouping it under Tools

**Alternatives considered**:
- Put model management inside Settings page: Rejected - higher regression risk, mixes legacy/advanced flows
- Replace Settings page entirely: Rejected - breaks backward compatibility

### Decision 2: Default Model Selection
**What**: Allow marking one model as "default" via boolean flag `is_default`. API uses default if no model specified in request.

**Why**:
- Simplifies UX - users don't need to select model every time
- Natural fallback behavior
- Simple to implement (database query with `WHERE is_default=true`)

**Alternatives considered**:
- User preference table: Rejected - over-engineered for single setting
- Always require explicit selection: Rejected - poor UX for common case

### Decision 3: Enable/Disable Models
**What**: Add `is_enabled` boolean flag. Disabled models don't appear in selector dropdown but configuration is preserved.

**Why**:
- Allows temporary deactivation without losing configuration
- Organization admins can disable deprecated models
- Users can hide models they don't use

### Decision 4: Pre-Seed Common Models
**What**: Database migration includes 7 pre-configured models (see proposal for list) with sensible defaults.

**Why**:
- Immediate value for new users
- Reduces configuration burden
- Showcases available model options
- Users still need to provide their own API keys

**Alternatives considered**:
- Empty database: Rejected - poor first-run experience
- Only seed on first user signup: Rejected - complicates migration logic

## Risks / Trade-offs

**Risk**: API key confusion - users might not realize they can set per-model keys
**Mitigation**: Clear UI messaging that each model can have its own API key (stored locally), with fallback to Settings default

**Trade-off**: Database size increases with model configurations
**Accept**: Minimal (~KB per model) and provides significant UX value

**Risk**: Model parameter drift (e.g., max_tokens changes for new model versions)
**Mitigation**: Allow users to edit all parameters; document parameter meanings

**Risk**: Users might lose API keys if they clear browser storage
**Mitigation**: UI warning when clearing storage; keys are easy to re-enter from provider dashboards

## Migration Plan

### Forward Migration
1. Create `llm_models` table with schema
2. Insert 7 pre-configured models (if table is empty)
3. Add new LLM model management page/route (Settings remains as-is)
4. Existing users continue using localStorage settings (no data migration required)

### Rollback Plan
1. Drop `llm_models` table
2. Remove new model management page/route
3. System reverts to localStorage-only settings

### Data Migration
**None required** - new feature doesn't depend on migrating existing settings. Old localStorage approach remains valid fallback.

## Open Questions

- ❓ Should we support per-request model override via API query parameter?
  - **Answer**: Yes, but can be implemented later
- ❓ How to handle base_url sharing (e.g., multiple models from same gateway)?
  - **Answer**: ~~User provides one API key for the base_url, works for all models on that endpoint~~
  - **Updated Answer**: Each model can have its own API key. If not set, falls back to Settings default. This handles both same-provider-different-accounts and different-providers scenarios.
- ❓ How to handle different providers requiring different API keys?
  - **Answer**: Per-model API keys stored in localStorage. UI shows key icon badge for models with custom keys.

