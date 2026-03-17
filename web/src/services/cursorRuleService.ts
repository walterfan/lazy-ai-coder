import type {
  CursorRule,
  GenerateRuleRequest,
  RefineRuleRequest,
  CursorRulesResponse,
} from '@/types/cursorRule';
import { apiService } from './apiService';

class CursorRuleService {
  // List cursor rules with filters
  async listCursorRules(params?: {
    scope?: string;
    q?: string;
    tags?: string;
    language?: string;
    framework?: string;
    sort?: string;
    limit?: number;
    offset?: number;
  }): Promise<CursorRulesResponse> {
    try {
      const response = await apiService.get<CursorRulesResponse>('/cursor-rules', { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch cursor rules:', error);
      throw error;
    }
  }

  // Get cursor rule by ID
  async getCursorRule(id: string): Promise<CursorRule> {
    const response = await apiService.get<CursorRule>(`/cursor-rules/${id}`);
    return response.data;
  }

  // Create cursor rule
  async createCursorRule(rule: Partial<CursorRule>): Promise<CursorRule> {
    const response = await apiService.post<CursorRule>('/cursor-rules', rule);
    return response.data;
  }

  // Update cursor rule
  async updateCursorRule(id: string, rule: Partial<CursorRule>): Promise<CursorRule> {
    const response = await apiService.put<CursorRule>(`/cursor-rules/${id}`, rule);
    return response.data;
  }

  // Delete cursor rule
  async deleteCursorRule(id: string): Promise<void> {
    await apiService.delete(`/cursor-rules/${id}`);
  }

  // Generate cursor rule
  async generateCursorRule(request: GenerateRuleRequest): Promise<{ content: string }> {
    const response = await apiService.post<{ content: string }>('/cursor-rules/generate', request);
    return response.data;
  }

  // Refine cursor rule
  async refineCursorRule(id: string, request: RefineRuleRequest): Promise<{ content: string }> {
    const response = await apiService.post<{ content: string }>(
      `/cursor-rules/${id}/refine`,
      request
    );
    return response.data;
  }

  // Export cursor rule
  async exportCursorRule(id: string): Promise<string> {
    const response = await apiService.get(`/cursor-rules/${id}/export`, {
      responseType: 'text',
    });
    return response.data;
  }

  // Import cursor rule
  async importCursorRule(rule: Partial<CursorRule>): Promise<CursorRule> {
    const response = await apiService.post<CursorRule>('/cursor-rules/import', rule);
    return response.data;
  }

  // Validate cursor rule
  async validateCursorRule(content: string): Promise<{ valid: boolean; errors: string[] }> {
    const response = await apiService.post<{ valid: boolean; errors: string[] }>(
      '/cursor-rules/validate',
      { content }
    );
    return response.data;
  }
}

export const cursorRuleService = new CursorRuleService();

