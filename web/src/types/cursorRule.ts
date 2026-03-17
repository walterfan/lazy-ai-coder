// Cursor Rule Types
export interface CursorRule {
  id: string;
  user_id?: string;
  realm_id?: string;
  name: string;
  description?: string;
  content: string;
  language?: string;
  framework?: string;
  tags?: string;
  is_template: boolean;
  usage_count: number;
  created_by?: string;
  created_time?: string;
  updated_by?: string;
  updated_time?: string;
}

export interface GenerateRuleRequest {
  project_context?: {
    language?: string;
    framework?: string;
    framework_version?: string;
    build_tool?: string;
    database?: string;
    has_tests?: boolean;
    test_framework?: string;
    dependencies?: string[];
  };
  language?: string;
  framework?: string;
  requirements?: string;
  template_id?: string;
  settings: {
    LLM_API_KEY: string;
    LLM_MODEL: string;
    LLM_BASE_URL: string;
    LLM_TEMPERATURE: string;
    GITLAB_BASE_URL: string;
    GITLAB_TOKEN: string;
  };
}

export interface RefineRuleRequest {
  improvements?: string;
  focus_areas?: string[];
  settings: {
    LLM_API_KEY: string;
    LLM_MODEL: string;
    LLM_BASE_URL: string;
    LLM_TEMPERATURE: string;
    GITLAB_BASE_URL: string;
    GITLAB_TOKEN: string;
  };
}

export interface CursorRulesResponse {
  data: CursorRule[];
  total: number;
  limit: number;
  offset: number;
  auth_type?: string;
  username?: string;
}

