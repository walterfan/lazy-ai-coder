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

// Code Knowledge Graph Types
export interface CodeKGRepo {
  id: string;
  name: string;
  url: string;
  local_path: string;
  branch: string;
  last_commit: string;
  last_sync: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CodeKGEntity {
  id: string;
  repo_id: string;
  entity_type: string;
  name: string;
  file_path: string;
  start_line: number;
  end_line: number;
  signature: string;
  doc_string: string;
  body?: string;
  summary: string;
  language: string;
  created_at?: string;
}

export interface CodeKGSearchResult {
  entities: CodeKGEntity[];
  answer: string;
}

export interface CodeKGSyncStatus {
  job_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'idle';
  total_files: number;
  processed_files: number;
  entities_created: number;
  entities_updated: number;
  entities_deleted: number;
  error?: string;
}

export interface CodeKGEntityPage {
  data: CodeKGEntity[];
  total: number;
  page: number;
}

export interface CodeKGKnowledgeDoc {
  id: string;
  repo_id: string;
  doc_type: string;
  title: string;
  content: string;
  created_at: string;
  updated_at: string;
}

// Export cursor rule types
export * from './cursorRule';
export * from './cursorCommand';
export * from './chatRecord';