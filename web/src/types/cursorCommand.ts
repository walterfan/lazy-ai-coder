export interface CursorCommand {
  id: string;
  user_id?: string | null;
  realm_id?: string | null;
  name: string;
  description: string;
  command: string; // The actual command/prompt text
  category: string; // e.g., "refactor", "debug", "generate", "review"
  language: string;
  framework: string;
  tags: string;
  is_template: boolean;
  usage_count: number;
  created_by: string;
  created_time: string;
  updated_by: string;
  updated_time: string;
}

export interface GenerateCommandRequest {
  category?: string;
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

export interface RefineCommandRequest {
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

export interface ImportCommandRequest {
  name: string;
  description?: string;
  command: string;
  category?: string;
  language?: string;
  framework?: string;
  tags?: string;
  scope?: string;
}

