// Smart Prompt Generator TypeScript types

export interface SmartPromptRequest {
  input: string;
  gitlab_project: string;
  gitlab_branch: string;
  gitlab_code_path: string;
  analyze_context: boolean;
  preset_id: string;
  settings: {
    LLM_API_KEY: string;
    LLM_MODEL: string;
    LLM_BASE_URL: string;
    LLM_TEMPERATURE: string;
    GITLAB_BASE_URL: string;
    GITLAB_TOKEN: string;
  };
}

export interface ProjectContext {
  language: string;
  framework: string;
  framework_version: string;
  build_tool: string;
  database: string;
  has_tests: boolean;
  test_framework: string;
  dependencies: string[];
}

export interface CodeExample {
  title: string;
  language: string;
  code: string;
  description: string;
}

export interface QualityScore {
  score: number;
  max_score: number;
  feedback: string[];
  suggestions: string[];
}

export interface SmartPromptResponse {
  context: string;
  action: string;
  result: string;
  full_prompt: string;
  detected_context: ProjectContext;
  examples: CodeExample[];
  quality_score: QualityScore;
}

export interface Preset {
  id: string;
  name: string;
  language: string;
  framework: string;
  context_hints: string;
  result_hints: string;
}

// Framework-related types
export interface FrameworkField {
  id: string;
  label: string;
  description: string;
  placeholder: string;
  required: boolean;
  type: string; // 'text' | 'textarea' | 'select'
  options?: string[];
}

export interface FrameworkExample {
  use_case: string;
  input: string;
  generated_prompt: string;
}

export interface Framework {
  id: string;
  name: string;
  description: string;
  fields: FrameworkField[];
  template: string;
  best_for: string;
  example: FrameworkExample;
}

// Template-related types
export interface TemplateCategory {
  id: string;
  name: string;
  description: string;
  icon: string;
}

export interface PromptTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  framework: string; // Framework ID this template uses
  fields: Record<string, string>; // Pre-filled field values
  tags: string[];
  use_count: number;
}

// Refinement-related types
export interface RefinementSuggestion {
  title: string;
  before: string;
  after: string;
  impact: 'high' | 'medium' | 'low';
  description: string;
}

export interface RefinementResponse {
  original_prompt: string;
  refined_prompt: string;
  assessment: string;
  suggestions: RefinementSuggestion[];
  quality_before: number;
  quality_after: number;
  improvement_tips: string[];
}

// Request types
export interface GenerateFromFrameworkRequest {
  framework_id: string;
  fields: Record<string, string>;
}

export interface RefinePromptRequest {
  prompt: string;
  settings: {
    LLM_API_KEY: string;
    LLM_MODEL: string;
    LLM_BASE_URL: string;
    LLM_TEMPERATURE?: string;
    GITLAB_BASE_URL?: string;
    GITLAB_TOKEN?: string;
  };
}

export interface QuickRefineRequest {
  prompt: string;
  framework_id: string;
}

// Response types
export interface GenerateFromFrameworkResponse {
  system_prompt: string;
  user_prompt: string;
  full_prompt: string;
  framework_id: string;
  framework_name: string;
  quality_score: number;
  max_score: number;
}

export interface QuickRefineResponse {
  suggestions: string[];
  quality_score: number;
  max_score: number;
}

// Auto-fill Fields API
export interface AutoFillFieldsRequest {
  framework_id: string;
  user_input: string;
  settings: {
    LLM_API_KEY: string;
    LLM_MODEL: string;
    LLM_BASE_URL: string;
    LLM_TEMPERATURE: string;
    GITLAB_BASE_URL: string;
    GITLAB_TOKEN: string;
  };
}

export interface AutoFillFieldsResponse {
  fields: Record<string, string>;
}
