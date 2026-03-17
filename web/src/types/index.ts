// Asset (command, rule, skill from assets folder)
export interface AssetItem {
  type: 'command' | 'rule' | 'skill';
  path: string;
  name: string;
  snippet: string;
  category: string;
}

export interface ListAssetsResponse {
  data: AssetItem[];
  total: number;
}

// Settings Types
export interface Settings {
  LLM_API_KEY: string;
  LLM_MODEL: string;
  LLM_BASE_URL: string;
  LLM_TEMPERATURE: string;
  GITLAB_BASE_URL: string;
  GITLAB_TOKEN: string;
}

// Prompt Types
export interface PromptArgument {
  name: string;
  description: string;
  required: boolean;
}

export interface Prompt {
  name: string;
  title?: string;
  description: string;
  system_prompt: string;
  user_prompt: string;
  assistant_prompt?: string;
  arguments?: PromptArgument[];
  tags?: string;
  user_id?: string;
  realm_id?: string;
  created_by?: string;
  created_time?: string;
  updated_by?: string;
  updated_time?: string;
}

// Project Types
export interface Project {
  id: string;
  user_id?: string;
  realm_id: string;
  name: string;
  description?: string;
  git_url?: string;
  git_repo?: string;
  git_branch?: string;
  language?: string;
  entry_point?: string;
  created_by?: string;
  created_time?: string;
  updated_by?: string;
  updated_time?: string;
  // Legacy field names (for backwards compatibility)
  gitlab_code_repo?: string;
  gitlab_code_path?: string;
  gitlab_code_branch?: string;
}

export interface GitlabProjects {
  [key: string]: Project;
}

// Form Data Types
export interface AgentFormData {
  prompt: {
    name: string;
    system_prompt: string;
    user_prompt: string;
    assistant_prompt?: string;
  };
  session_id: string;
  computer_language: string;
  output_languages: string[];
  codePath: string;
  stream: boolean;
  remember: boolean;
  project: string;
  gitlab_code_repo: string;
  gitlab_code_path: string;
  gitlab_code_branch: string;
  gitlab_mr_id: string;
}

// API Request/Response Types
export interface AgentRequest {
  prompt: {
    name: string;
    system_prompt: string;
    user_prompt: string;
    assistant_prompt?: string;
  };
  session_id: string;
  computer_language: string;
  output_languages: string[];
  codePath?: string;
  stream?: boolean;
  remember?: boolean;
  project?: string;
  gitlab_code_repo?: string;
  gitlab_code_path?: string;
  gitlab_code_branch?: string;
  gitlab_mr_id?: string;
  settings: Settings;
}

export interface AgentResponse {
  content: string;
  session_id?: string;
  error?: string;
}

// WebSocket Message Types
export interface WSMessage {
  type: 'chunk' | 'complete' | 'error';
  content: string;
  session_id?: string;
}

// Navigation Types
export interface NavItem {
  name: string;
  path: string;
  icon: string;
  children?: NavItem[];
}

// Toast Types
export type ToastType = 'success' | 'danger' | 'warning' | 'info';

// File Types
export interface ConversationFile {
  timestamp: string;
  title: string;
  module?: string;
  tags?: string;
  content: string;
}

// Document Types
export interface Document {
  id: string;
  realm_id: string;
  project_id: string;
  name: string;
  path: string;
  content: string;
  vector_embedding?: string;
  created_by?: string;
  created_time?: string;
  updated_by?: string;
  updated_time?: string;
  deleted_at?: string;
}

// Export cursor rule types
export * from './cursorRule';
export * from './cursorCommand';
export * from './chatRecord';