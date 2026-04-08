// Coding Mate / Chat Record Types

export type InputType =
  | 'research_solution'
  | 'learn_tech'
  | 'tech_design'
  | 'word'
  | 'sentence'
  | 'question'
  | 'idea'
  | 'topic';

export interface ConceptItem {
  name: string;
  description: string;
  importance?: string;
}

export interface ChatStep {
  order: number;
  title: string;
  description: string;
  duration?: string;
  objectives?: string[];
}

export interface ResourceItem {
  type: string; // tutorial, documentation, course, video, book
  title: string;
  url?: string;
  description?: string;
  difficulty?: string; // beginner, intermediate, advanced
}

export interface OptionItem {
  name?: string;
  pros?: string;
  cons?: string;
  summary?: string;
}

export interface ResponsePayload {
  // Common
  explanation?: string;
  example?: string;

  // research_solution
  summary?: string;
  options?: OptionItem[];
  trade_offs?: string;
  recommendation?: string;
  references?: string[];

  // learn_tech / topic
  introduction?: string;
  key_concepts?: ConceptItem[];
  chat_path?: ChatStep[];
  resources?: ResourceItem[];
  prerequisites?: string[];
  time_estimate?: string;

  // tech_design / idea
  problem_statement?: string;
  approach_options?: string[];
  chosen_approach?: string;
  components?: string[];
  risks?: string;
  plan?: string[];

  // Legacy
  pronunciation?: string;
  answer?: string;
}

export interface ChatRecordSummary {
  id: string;
  input_type: InputType;
  user_input: string;
  response_summary: string;
  created_time: string;
}

export interface ChatRecord {
  id: string;
  input_type: InputType;
  user_input: string;
  response_payload: ResponsePayload;
  user_id: string;
  realm_id?: string;
  created_by?: string;
  created_time: string;
  updated_time?: string;
}

// API Request Types
export interface SubmitChatRecordRequest {
  user_input: string;
  session_id?: string;
}

export interface ConfirmChatRecordRequest {
  user_input: string;
  input_type: InputType;
  response_payload: ResponsePayload;
}

// API Response Types
export interface SubmitChatRecordResponse {
  input_type: InputType;
  response_payload: ResponsePayload;
  similar_records?: ChatRecordSummary[];
  session_id?: string;
}

// One round of multi-turn conversation (user question + assistant response)
export interface ConversationTurn {
  userInput: string;
  response: SubmitChatRecordResponse | null; // null while loading
  /** Raw markdown accumulated during SSE streaming (present only in streaming mode) */
  streamContent?: string;
  /** True while SSE stream is actively receiving tokens */
  streaming?: boolean;
}

export interface ListChatRecordsResponse {
  records: ChatRecordSummary[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ChatRecordStats {
  total: number;
  by_type: Record<InputType, number>;
  streak: number;
  last_record_at?: string;
}

// Helper for input type display (Code Mate types first, then legacy)
export const InputTypeConfig: Record<string, { label: string; icon: string; color: string }> = {
  research_solution: { label: 'Research', icon: 'fa-search', color: 'primary' },
  learn_tech: { label: 'Learn Tech', icon: 'fa-book', color: 'info' },
  tech_design: { label: 'Tech Design', icon: 'fa-drafting-compass', color: 'success' },
  skill_response: { label: 'Skill', icon: 'fa-graduation-cap', color: 'warning' },
  word: { label: 'Word', icon: 'fa-font', color: 'primary' },
  sentence: { label: 'Sentence', icon: 'fa-align-left', color: 'info' },
  question: { label: 'Question', icon: 'fa-question-circle', color: 'warning' },
  idea: { label: 'Idea', icon: 'fa-lightbulb', color: 'success' },
  topic: { label: 'Topic', icon: 'fa-book', color: 'secondary' },
};
