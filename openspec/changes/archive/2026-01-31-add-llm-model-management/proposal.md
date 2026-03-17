# Change: LLM Model Management

## Why

Currently, users can only configure a single LLM provider via hardcoded settings (API key, model name, base URL, temperature). This limits users to one model at a time and requires manually changing settings to switch between different LLM providers (e.g., GPT, Claude, Gemini, Qwen, DeepSeek). Users need the ability to configure multiple LLM models, mark one as default, enable/disable models, and easily switch between them without losing configuration.

## What Changes

- Add `LLMModel` database model to store multiple LLM configurations per user/realm
- Create CRUD API endpoints for LLM model management (`GET`, `POST`, `PUT`, `DELETE`, set default)
- Add a **new dedicated page** for managing LLM models (list, add, edit, delete, enable/disable, set default)
- Keep the existing Settings page **unchanged** for backward compatibility (legacy single-model localStorage config)
- Add navigation entry points to open the new LLM model management page:
  - under **Tools** menu
  - optional link in **Settings** page (allowed; does not change legacy settings behavior)
- **Keep API key separate** in existing settings (security concern - not stored in database)
- Add LLM model selector in Lazy AI Coder page to choose which model to use for each request
- Pre-populate database with 7 example models (GPT, Claude, Gemini, Qwen, DeepSeek variants)
- Update Settings Store to handle selected model configuration
- **NON-BREAKING**: Existing single-model settings still work as fallback

## Impact

- **Affected specs**: New capability `llm-config` (LLM Configuration Management)
- **Affected code**:
  - Backend: `pkg/models/llm_model.go` (new), `internal/handlers/llm_handlers.go` (new)
  - Frontend: `web/src/stores/llmModelStore.ts` (new), `web/src/views/LLMModelsView.vue` (new), `web/src/views/SettingsView.vue` (kept for back-compat)
  - Frontend: `web/src/views/AssistantView.vue` (add model selector)
  - Database: Migration to create `llm_models` table
- **Database schema**: New `llm_models` table with fields: id, name, llm_type, base_url, model, extra_params, temperature, max_tokens, is_default, is_enabled, description, user_id, realm_id, timestamps
- **Breaking**: None (backward compatible with existing settings)

