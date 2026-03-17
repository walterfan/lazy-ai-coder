-- Migration: Create llm_models table for multi-model LLM configuration management
-- This allows users to configure multiple LLM models and switch between them

-- Create llm_models table
CREATE TABLE IF NOT EXISTS llm_models (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    llm_type TEXT NOT NULL DEFAULT 'openai',  -- openai, anthropic, google, alibaba, deepseek, etc.
    base_url TEXT NOT NULL,
    model TEXT NOT NULL,                       -- Model identifier (e.g., gpt-4, claude-3-opus)
    extra_params TEXT,                         -- JSON for additional provider-specific parameters
    temperature REAL DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4096,
    is_default INTEGER DEFAULT 0,              -- Boolean: 1 = default model
    is_enabled INTEGER DEFAULT 1,              -- Boolean: 1 = enabled
    description TEXT,
    user_id TEXT,                              -- NULL for realm-shared models
    realm_id TEXT,                             -- NULL for global/template models
    created_by TEXT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,                       -- Soft delete
    FOREIGN KEY (user_id) REFERENCES app_user(id),
    FOREIGN KEY (realm_id) REFERENCES realm(id)
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_llm_models_user_id ON llm_models(user_id);
CREATE INDEX IF NOT EXISTS idx_llm_models_realm_id ON llm_models(realm_id);
CREATE INDEX IF NOT EXISTS idx_llm_models_is_default ON llm_models(is_default);
CREATE INDEX IF NOT EXISTS idx_llm_models_is_enabled ON llm_models(is_enabled);
CREATE INDEX IF NOT EXISTS idx_llm_models_deleted_at ON llm_models(deleted_at);

-- Insert pre-configured LLM models (realm-scoped, no user_id)
-- These are template models that users can use with their own API keys
INSERT OR IGNORE INTO llm_models (id, name, llm_type, base_url, model, temperature, max_tokens, is_default, is_enabled, description, created_by)
VALUES
    ('llm_openai_gpt4', 'GPT-4', 'openai', 'https://api.openai.com/v1', 'gpt-4', 0.7, 8192, 0, 1, 'OpenAI GPT-4 - Most capable model for complex tasks', 'system'),
    ('llm_openai_gpt4_turbo', 'GPT-4 Turbo', 'openai', 'https://api.openai.com/v1', 'gpt-4-turbo-preview', 0.7, 128000, 0, 1, 'OpenAI GPT-4 Turbo - Faster with larger context window', 'system'),
    ('llm_openai_gpt35', 'GPT-3.5 Turbo', 'openai', 'https://api.openai.com/v1', 'gpt-3.5-turbo', 0.7, 16384, 0, 1, 'OpenAI GPT-3.5 Turbo - Fast and cost-effective', 'system'),
    ('llm_anthropic_claude3_opus', 'Claude 3 Opus', 'anthropic', 'https://api.anthropic.com/v1', 'claude-3-opus-20240229', 0.7, 4096, 0, 1, 'Anthropic Claude 3 Opus - Most intelligent model', 'system'),
    ('llm_anthropic_claude3_sonnet', 'Claude 3.5 Sonnet', 'anthropic', 'https://api.anthropic.com/v1', 'claude-3-5-sonnet-20241022', 0.7, 8192, 0, 1, 'Anthropic Claude 3.5 Sonnet - Balanced performance', 'system'),
    ('llm_google_gemini_pro', 'Gemini Pro', 'google', 'https://generativelanguage.googleapis.com/v1beta', 'gemini-pro', 0.7, 32768, 0, 1, 'Google Gemini Pro - Multimodal capabilities', 'system'),
    ('llm_alibaba_qwen_max', 'Qwen Max', 'alibaba', 'https://dashscope.aliyuncs.com/compatible-mode/v1', 'qwen-max', 0.7, 8192, 0, 1, 'Alibaba Qwen Max - Excellent for Chinese and English', 'system'),
    ('llm_deepseek_chat', 'DeepSeek Chat', 'deepseek', 'https://api.deepseek.com/v1', 'deepseek-chat', 0.7, 32768, 0, 1, 'DeepSeek Chat - Strong coding and reasoning', 'system');

