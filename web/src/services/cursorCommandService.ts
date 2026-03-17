import type {
  CursorCommand,
  GenerateCommandRequest,
  RefineCommandRequest,
  ImportCommandRequest,
} from '@/types/cursorCommand';
import { apiService } from './apiService';

interface CursorCommandsResponse {
  data: CursorCommand[];
  total: number;
  limit: number;
  offset: number;
  auth_type?: string;
  username?: string;
}

class CursorCommandService {
  // List cursor commands with filters
  async listCursorCommands(params?: {
    scope?: string;
    q?: string;
    tags?: string;
    category?: string;
    language?: string;
    framework?: string;
    sort?: string;
    limit?: number;
    offset?: number;
  }): Promise<CursorCommandsResponse> {
    try {
      const response = await apiService.get<CursorCommandsResponse>('/cursor-commands', { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch cursor commands:', error);
      throw error;
    }
  }

  // Get cursor command by ID
  async getCursorCommand(id: string): Promise<CursorCommand> {
    const response = await apiService.get<CursorCommand>(`/cursor-commands/${id}`);
    return response.data;
  }

  // Create cursor command
  async createCursorCommand(command: Partial<CursorCommand>): Promise<CursorCommand> {
    const response = await apiService.post<CursorCommand>('/cursor-commands', command);
    return response.data;
  }

  // Update cursor command
  async updateCursorCommand(id: string, command: Partial<CursorCommand>): Promise<CursorCommand> {
    const response = await apiService.put<CursorCommand>(`/cursor-commands/${id}`, command);
    return response.data;
  }

  // Delete cursor command
  async deleteCursorCommand(id: string): Promise<void> {
    await apiService.delete(`/cursor-commands/${id}`);
  }

  // Generate cursor command
  async generateCursorCommand(request: GenerateCommandRequest): Promise<{ command: string }> {
    const response = await apiService.post<{ command: string }>('/cursor-commands/generate', request);
    return response.data;
  }

  // Refine cursor command
  async refineCursorCommand(id: string, request: RefineCommandRequest): Promise<{ command: string }> {
    const response = await apiService.post<{ command: string }>(
      `/cursor-commands/${id}/refine`,
      request
    );
    return response.data;
  }

  // Export cursor command
  async exportCursorCommand(id: string): Promise<string> {
    const response = await apiService.get(`/cursor-commands/${id}/export`, {
      responseType: 'text',
    });
    return response.data;
  }

  // Import cursor command
  async importCursorCommand(command: ImportCommandRequest): Promise<CursorCommand> {
    const response = await apiService.post<CursorCommand>('/cursor-commands/import', command);
    return response.data;
  }
}

export const cursorCommandService = new CursorCommandService();

