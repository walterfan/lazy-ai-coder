import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios';
import type {
  Prompt,
  GitlabProjects,
  AgentRequest,
  AgentResponse,
  Settings,
  ListAssetsResponse,
} from '@/types';
import { showToast } from '@/utils/toast';
import { useAuthStore } from '@/stores/authStore';
import { storage } from '@/utils/storage';

class ApiService {
  private api: AxiosInstance;
  private baseURL: string;
  private isRefreshing = false;
  private refreshSubscribers: ((token: string) => void)[] = [];

  constructor() {
    const env = import.meta.env.VITE_API_BASE_URL || '';
    this.baseURL = env
      ? env.replace(/\/api\/v1\/?$/, '') + '/api/v1'
      : '/api/v1';
    this.api = axios.create({
      baseURL: this.baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 60000,
    });

    this.setupInterceptors();
  }

  // Queue requests while token is being refreshed
  private subscribeTokenRefresh(cb: (token: string) => void): void {
    this.refreshSubscribers.push(cb);
  }

  // Notify all queued requests with new token
  private onTokenRefreshed(token: string): void {
    this.refreshSubscribers.forEach(cb => cb(token));
    this.refreshSubscribers = [];
  }

  private setupInterceptors(): void {
    // Request interceptor: Add JWT token to authenticated requests
    this.api.interceptors.request.use(
      (config) => {
        // Get auth token from localStorage
        const token = storage.getItem('auth_token');
        const authMode = storage.getItem('auth_mode');

        // Add Authorization header for authenticated users (OAuth or password)
        if ((authMode === 'oauth' || authMode === 'password') && token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        // Guest mode: No Authorization header (uses credentials from request body)

        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor: Handle errors globally
    this.api.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const authMode = storage.getItem('auth_mode');
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

        // Handle 401 Unauthorized - token expired or invalid
        if (error.response?.status === 401 && originalRequest) {
          // Only handle for authenticated users
          if (authMode === 'oauth' || authMode === 'password') {
            // Avoid infinite retry loop
            if (originalRequest._retry) {
              // Show auth warning dialog
              const authStore = useAuthStore();
              authStore.showAuthWarning('expired', 'Your session has expired. Please login again to continue.');

              // Clear auth data and return to guest mode
              storage.removeItem('auth_token');
              storage.removeItem('user');
              storage.setItem('auth_mode', 'guest');

              return Promise.reject(error);
            }

            // Don't retry the refresh endpoint itself
            if (originalRequest.url?.includes('/auth/refresh')) {
              // Show auth warning dialog
              const authStore = useAuthStore();
              authStore.showAuthWarning('expired', 'Your session has expired. Please login again to continue.');

              storage.removeItem('auth_token');
              storage.removeItem('user');
              storage.setItem('auth_mode', 'guest');

              return Promise.reject(error);
            }

            originalRequest._retry = true;

            // If already refreshing, queue this request
            if (this.isRefreshing) {
              return new Promise((resolve) => {
                this.subscribeTokenRefresh((token: string) => {
                  if (originalRequest.headers) {
                    originalRequest.headers.Authorization = `Bearer ${token}`;
                  }
                  resolve(this.api(originalRequest));
                });
              });
            }

            // Start token refresh
            this.isRefreshing = true;

            try {
              const token = storage.getItem('auth_token');
              if (!token) {
                throw new Error('No token available');
              }

              // Attempt to refresh token
              const response = await fetch('/api/v1/auth/refresh', {
                method: 'POST',
                headers: {
                  'Authorization': `Bearer ${token}`,
                  'Content-Type': 'application/json',
                },
              });

              if (!response.ok) {
                throw new Error('Token refresh failed');
              }

              const data = await response.json();
              const newToken = data.token;

              // Update stored token
              storage.setItem('auth_token', newToken);
              if (data.user) {
                storage.setItem('user', JSON.stringify(data.user));
              }

              // Update the failed request with new token
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${newToken}`;
              }

              // Notify all queued requests
              this.onTokenRefreshed(newToken);

              // Retry the original request
              return this.api(originalRequest);
            } catch (refreshError) {
              // Refresh failed, show auth warning dialog
              const authStore = useAuthStore();
              authStore.showAuthWarning('expired', 'Your session has expired. Please login again to continue.');

              storage.removeItem('auth_token');
              storage.removeItem('user');
              storage.setItem('auth_mode', 'guest');

              return Promise.reject(refreshError);
            } finally {
              this.isRefreshing = false;
            }
          }
        }

        // Handle 403 Forbidden - read-only guest trying to modify or insufficient permissions
        if (error.response?.status === 403) {
          const authStore = useAuthStore();

          if (authMode === 'guest') {
            // Guest user trying to modify
            const errorData = error.response.data as any;
            const message = errorData?.error || 'This feature is only available for registered users.';
            authStore.showAuthWarning('guest', message);
            return Promise.reject(error);
          } else {
            // Authenticated user with insufficient permissions
            const errorData = error.response.data as any;
            const message = errorData?.error || 'You don\'t have permission to perform this action.';
            authStore.showAuthWarning('unauthorized', message);
            return Promise.reject(error);
          }
        }

        const message = error.response?.data
          ? JSON.stringify(error.response.data)
          : error.message;
        showToast(`API Error: ${message}`, 'danger');
        return Promise.reject(error);
      }
    );
  }

  // Generic HTTP methods for external use
  async get<T = any>(url: string, config?: any): Promise<import('axios').AxiosResponse<T>> {
    return this.api.get<T>(url, config);
  }

  async post<T = any>(url: string, data?: any, config?: any): Promise<import('axios').AxiosResponse<T>> {
    return this.api.post<T>(url, data, config);
  }

  async put<T = any>(url: string, data?: any, config?: any): Promise<import('axios').AxiosResponse<T>> {
    return this.api.put<T>(url, data, config);
  }

  async patch<T = any>(url: string, data?: any, config?: any): Promise<import('axios').AxiosResponse<T>> {
    return this.api.patch<T>(url, data, config);
  }

  async delete<T = any>(url: string, config?: any): Promise<import('axios').AxiosResponse<T>> {
    return this.api.delete<T>(url, config);
  }

  // Assets API (commands, rules, skills from assets folder)
  async listAssets(params?: { type?: string; q?: string; category?: string }): Promise<ListAssetsResponse> {
    const response = await this.api.get<ListAssetsResponse>('/assets', { params });
    return response.data;
  }

  getAssetDownloadUrl(path: string, asAttachment = true): string {
    const p = new URLSearchParams({ path });
    if (asAttachment) p.set('download', '1');
    return `${this.api.defaults.baseURL}/assets/download?${p.toString()}`;
  }

  async downloadAsset(path: string, filename?: string): Promise<void> {
    const response = await this.api.get('/assets/download', {
      params: { path, download: '1' },
      responseType: 'blob',
    });
    const name = filename || path.split('/').pop() || 'download';
    this.triggerBlobDownload(response.data, name);
  }

  async downloadSkillZip(path: string): Promise<void> {
    const response = await this.api.get('/assets/download-skill', {
      params: { path },
      responseType: 'blob',
    });
    const disposition = response.headers['content-disposition'] || '';
    const match = disposition.match(/filename="?([^"]+)"?/);
    const name = match?.[1] || path.split('/').slice(-2, -1)[0] + '.zip' || 'skill.zip';
    this.triggerBlobDownload(response.data, name);
  }

  private triggerBlobDownload(data: Blob | ArrayBuffer, name: string): void {
    const url = window.URL.createObjectURL(new Blob([data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', name);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  }

  // Prompts API
  async getPrompts(tags?: string, searchQuery?: string): Promise<Prompt[]> {
    try {
      const params: any = {};
      if (tags) params.tags = tags;
      if (searchQuery) params.q = searchQuery;

      const response = await this.api.get<any>('/prompts', { params });
      // API returns {data: [...], total: N} format
      const prompts = response.data.data || response.data;
      // Parse arguments from JSON string to array if needed
      return prompts.map((prompt: any) => ({
        ...prompt,
        arguments: typeof prompt.arguments === 'string' && prompt.arguments
          ? JSON.parse(prompt.arguments)
          : prompt.arguments || []
      }));
    } catch (error) {
      console.error('Failed to fetch prompts:', error);
      throw error;
    }
  }

  async createPrompt(prompt: Prompt): Promise<Prompt> {
    const response = await this.api.post<Prompt>('/prompts', prompt);
    return response.data;
  }

  async updatePrompt(name: string, prompt: Prompt): Promise<Prompt> {
    const response = await this.api.put<Prompt>(`/prompts/${name}`, prompt);
    return response.data;
  }

  async deletePrompt(name: string): Promise<void> {
    await this.api.delete(`/prompts/${name}`);
  }

  // Projects API
  async getProjects(): Promise<any> {
    try {
      const response = await this.api.get<any>('/projects');
      // API returns {data: [...], total: N} format
      return response.data.data || response.data;
    } catch (error) {
      console.error('Failed to fetch projects:', error);
      throw error;
    }
  }

  async getProject(id: string): Promise<any> {
    const response = await this.api.get<any>(`/projects/${id}`);
    return response.data;
  }

  async createProject(project: any): Promise<any> {
    const response = await this.api.post<any>('/projects', project);
    return response.data;
  }

  async updateProject(id: string, project: any): Promise<any> {
    const response = await this.api.put<any>(`/projects/${id}`, project);
    return response.data;
  }

  async deleteProject(id: string): Promise<void> {
    await this.api.delete(`/projects/${id}`);
  }

  async exportProjects(scope: string = 'all'): Promise<Blob> {
    const response = await this.api.get('/projects/export', {
      params: { scope },
      responseType: 'blob',
    });
    return response.data;
  }

  async importProjects(projects: any[], updateExisting: boolean = false, scope: string = 'personal'): Promise<any> {
    const response = await this.api.post('/projects/import', {
      projects,
      update_existing: updateExisting,
      scope,
    });
    return response.data;
  }

  // GitLab Configuration API (Legacy)
  async getGitlabConfig(): Promise<GitlabProjects> {
    try {
      const response = await this.api.get<GitlabProjects>('/gitlab_config');
      return response.data;
    } catch (error) {
      console.error('Failed to fetch GitLab config:', error);
      return {};
    }
  }

  // Agent API (non-streaming)
  async executeAgent(request: AgentRequest): Promise<AgentResponse> {
    const response = await this.api.post<AgentResponse>('/agent', request);
    return response.data;
  }

  // WebSocket for streaming
  createWebSocketConnection(
    formData: Partial<AgentRequest>,
    settings: Settings
  ): Promise<WebSocket> {
    return new Promise((resolve, reject) => {
      try {
        const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const ws = new WebSocket(`${protocol}://${window.location.host}/api/v1/stream`);

        const requestData: AgentRequest = {
          ...formData as AgentRequest,
          settings,
        };

        ws.onopen = () => {
          console.log('WebSocket opened, sending data:', requestData);
          showToast('Connected! Streaming response...', 'info');
          ws.send(JSON.stringify(requestData));
          resolve(ws);
        };

        ws.onerror = (err) => {
          console.error('WebSocket Error:', err);
          showToast('Failed to connect to streaming service', 'danger');
          reject(err);
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  // Smart Prompt Generator - Frameworks API
  async getFrameworks(): Promise<import('@/types/smart-prompt').Framework[]> {
    const response = await this.api.get('/smart-prompt/frameworks');
    return response.data;
  }

  async getFramework(id: string): Promise<import('@/types/smart-prompt').Framework> {
    const response = await this.api.get(`/smart-prompt/frameworks/${id}`);
    return response.data;
  }

  async createFramework(framework: import('@/types/smart-prompt').Framework): Promise<import('@/types/smart-prompt').Framework> {
    const response = await this.api.post('/smart-prompt/frameworks', framework);
    return response.data;
  }

  async updateFramework(id: string, framework: import('@/types/smart-prompt').Framework): Promise<import('@/types/smart-prompt').Framework> {
    const response = await this.api.put(`/smart-prompt/frameworks/${id}`, framework);
    return response.data;
  }

  async deleteFramework(id: string): Promise<void> {
    await this.api.delete(`/smart-prompt/frameworks/${id}`);
  }

  // Smart Prompt Generator - Templates API
  async getTemplateCategories(): Promise<import('@/types/smart-prompt').TemplateCategory[]> {
    const response = await this.api.get('/smart-prompt/templates/categories');
    return response.data;
  }

  async getTemplates(category?: string): Promise<import('@/types/smart-prompt').PromptTemplate[]> {
    const params = category ? { category } : {};
    const response = await this.api.get('/smart-prompt/templates', { params });
    return response.data;
  }

  async getTemplate(id: string): Promise<import('@/types/smart-prompt').PromptTemplate> {
    const response = await this.api.get(`/smart-prompt/templates/${id}`);
    return response.data;
  }

  async useTemplate(id: string): Promise<import('@/types/smart-prompt').PromptTemplate> {
    const response = await this.api.post(`/smart-prompt/templates/${id}/use`);
    return response.data;
  }

  async createTemplate(template: import('@/types/smart-prompt').PromptTemplate): Promise<import('@/types/smart-prompt').PromptTemplate> {
    const response = await this.api.post('/smart-prompt/templates', template);
    return response.data;
  }

  async updateTemplate(id: string, template: import('@/types/smart-prompt').PromptTemplate): Promise<import('@/types/smart-prompt').PromptTemplate> {
    const response = await this.api.put(`/smart-prompt/templates/${id}`, template);
    return response.data;
  }

  async deleteTemplate(id: string): Promise<void> {
    await this.api.delete(`/smart-prompt/templates/${id}`);
  }

  // Smart Prompt Generator - Refinement API
  async refinePrompt(
    prompt: string,
    settings: Settings
  ): Promise<import('@/types/smart-prompt').RefinementResponse> {
    const request: import('@/types/smart-prompt').RefinePromptRequest = {
      prompt,
      settings: {
        LLM_API_KEY: settings.LLM_API_KEY,
        LLM_MODEL: settings.LLM_MODEL,
        LLM_BASE_URL: settings.LLM_BASE_URL,
        LLM_TEMPERATURE: settings.LLM_TEMPERATURE,
        GITLAB_BASE_URL: settings.GITLAB_BASE_URL,
        GITLAB_TOKEN: settings.GITLAB_TOKEN,
      },
    };
    const response = await this.api.post('/smart-prompt/refine', request);
    return response.data;
  }

  async quickRefine(
    prompt: string,
    frameworkId: string
  ): Promise<import('@/types/smart-prompt').QuickRefineResponse> {
    const request: import('@/types/smart-prompt').QuickRefineRequest = {
      prompt,
      framework_id: frameworkId,
    };
    const response = await this.api.post('/smart-prompt/quick-refine', request);
    return response.data;
  }

  async refinePromptWithRequirements(request: {
    system_prompt: string;
    user_prompt: string;
    requirements: string;
    settings: {
      LLM_API_KEY: string;
      LLM_MODEL: string;
      LLM_BASE_URL: string;
      LLM_TEMPERATURE: string;
    };
  }): Promise<{ system_prompt: string; user_prompt: string }> {
    const response = await this.api.post('/smart-prompt/refine-with-requirements', request);
    return response.data;
  }

  // Smart Prompt Generator - Generation API
  async generateFromFramework(
    frameworkId: string,
    fields: Record<string, string>
  ): Promise<import('@/types/smart-prompt').GenerateFromFrameworkResponse> {
    const request: import('@/types/smart-prompt').GenerateFromFrameworkRequest = {
      framework_id: frameworkId,
      fields,
    };
    const response = await this.api.post('/smart-prompt/generate-from-framework', request);
    return response.data;
  }

  // Smart Prompt Generator - Auto-fill Fields API
  async autoFillFields(
    frameworkId: string,
    userInput: string,
    settings: Settings
  ): Promise<import('@/types/smart-prompt').AutoFillFieldsResponse> {
    const request: import('@/types/smart-prompt').AutoFillFieldsRequest = {
      framework_id: frameworkId,
      user_input: userInput,
      settings: {
        LLM_API_KEY: settings.LLM_API_KEY,
        LLM_MODEL: settings.LLM_MODEL,
        LLM_BASE_URL: settings.LLM_BASE_URL,
        LLM_TEMPERATURE: settings.LLM_TEMPERATURE,
        GITLAB_BASE_URL: settings.GITLAB_BASE_URL,
        GITLAB_TOKEN: settings.GITLAB_TOKEN,
      },
    };
    const response = await this.api.post('/smart-prompt/auto-fill-fields', request);
    return response.data;
  }

  /**
   * Call an MCP tool with arguments and optional settings
   * Settings for MCP tool calls (credentials passed from frontend):
   * - gitlab_token: GitLab personal access token
   * - gitlab_url: GitLab base URL
   * - llm_api_key: LLM API key
   * - llm_base_url: LLM base URL
   * - llm_model: LLM model name
   *
   * @param toolName - Name of the MCP tool to call
   * @param args - Arguments to pass to the tool
   * @param settings - Optional settings (GitLab token, LLM API key, etc.)
   * @returns Tool call result
   */
  async callMCPTool(
    toolName: string,
    args: Record<string, unknown>,
    settings?: {
      gitlab_token?: string;
      gitlab_url?: string;
      llm_api_key?: string;
      llm_base_url?: string;
      llm_model?: string;
    }
  ): Promise<{
    isError?: boolean;
    content: Array<{ type: string; text: string }>;
  }> {
    const response = await fetch('/mcp/v1/call-tool', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: toolName,
        arguments: args,
        settings: settings,
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }

    return response.json();
  }
}

export const apiService = new ApiService();
