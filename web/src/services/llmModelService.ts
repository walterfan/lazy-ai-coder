import { apiService } from './apiService';

export interface LLMModel {
  id: string;
  name: string;
  llm_type: string;
  base_url: string;
  model: string;
  extra_params?: string;
  temperature: number;
  max_tokens: number;
  is_default: boolean;
  is_enabled: boolean;
  description?: string;
  user_id?: string;
  realm_id?: string;
  created_by?: string;
  created_time?: string;
  updated_by?: string;
  updated_time?: string;
}

/**
 * Extended LLM model with local API key (stored in localStorage, not sent to server)
 */
export interface LLMModelWithKey extends LLMModel {
  api_key?: string; // Stored locally, never sent to server
}

export interface LLMModelListResponse {
  data: LLMModel[];
  total: number;
  pageNumber: number;
  pageSize: number;
  totalPages: number;
}

export interface DefaultLLMModelResponse {
  model: LLMModel | null;
  use_legacy: boolean;
  description?: string;
}

export interface CreateLLMModelRequest {
  name: string;
  llm_type: string;
  base_url: string;
  model: string;
  extra_params?: string;
  temperature?: number;
  max_tokens?: number;
  is_default?: boolean;
  is_enabled?: boolean;
  description?: string;
  scope?: 'personal' | 'shared' | 'templates';
}

export interface UpdateLLMModelRequest {
  name: string;
  llm_type: string;
  base_url: string;
  model: string;
  extra_params?: string;
  temperature?: number;
  max_tokens?: number;
  is_default?: boolean;
  is_enabled?: boolean;
  description?: string;
}

class LLMModelService {
  /**
   * List LLM models with optional filtering and pagination
   */
  async listLLMModels(
    scope: 'all' | 'personal' | 'shared' | 'templates' = 'all',
    query?: string,
    enabledOnly: boolean = false,
    pageNumber: number = 1,
    pageSize: number = 50
  ): Promise<LLMModelListResponse> {
    const params: Record<string, any> = {
      scope,
      pageNumber,
      pageSize,
    };
    if (query) params.q = query;
    if (enabledOnly) params.enabled_only = 'true';

    const response = await apiService.get<LLMModelListResponse>('/llm-models', { params });
    return response.data;
  }

  /**
   * Get a single LLM model by ID
   */
  async getLLMModel(id: string): Promise<LLMModel> {
    const response = await apiService.get<LLMModel>(`/llm-models/${id}`);
    return response.data;
  }

  /**
   * Get the default LLM model for the current user/realm
   * Returns null if no default is set (legacy settings should be used)
   */
  async getDefaultLLMModel(): Promise<DefaultLLMModelResponse> {
    const response = await apiService.get<DefaultLLMModelResponse>('/llm-models/default');
    return response.data;
  }

  /**
   * Create a new LLM model
   */
  async createLLMModel(data: CreateLLMModelRequest): Promise<LLMModel> {
    const response = await apiService.post<LLMModel>('/llm-models', data);
    return response.data;
  }

  /**
   * Update an existing LLM model
   */
  async updateLLMModel(id: string, data: UpdateLLMModelRequest): Promise<LLMModel> {
    const response = await apiService.put<LLMModel>(`/llm-models/${id}`, data);
    return response.data;
  }

  /**
   * Delete an LLM model
   */
  async deleteLLMModel(id: string): Promise<void> {
    await apiService.delete(`/llm-models/${id}`);
  }

  /**
   * Set an LLM model as the default
   */
  async setDefaultLLMModel(id: string): Promise<LLMModel> {
    const response = await apiService.post<LLMModel>(`/llm-models/${id}/default`);
    return response.data;
  }

  /**
   * Toggle the enabled status of an LLM model
   */
  async toggleLLMModelEnabled(id: string, enabled: boolean): Promise<LLMModel> {
    const response = await apiService.post<LLMModel>(`/llm-models/${id}/toggle`, { enabled });
    return response.data;
  }
}

export const llmModelService = new LLMModelService();

