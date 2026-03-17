import { defineStore } from 'pinia';
import { llmModelService, type LLMModel, type LLMModelWithKey, type CreateLLMModelRequest, type UpdateLLMModelRequest } from '@/services/llmModelService';
import { showToast } from '@/utils/toast';
import { storage } from '@/utils/storage';

interface LLMModelState {
  models: LLMModel[];
  selectedModel: LLMModel | null;
  defaultModel: LLMModel | null;
  useLegacy: boolean;
  loading: boolean;
  error: string | null;
  total: number;
  pageNumber: number;
  pageSize: number;
  totalPages: number;
  searchQuery: string;
  scope: 'all' | 'personal' | 'shared' | 'templates';
  enabledOnly: boolean;
}

// Keys for localStorage
const SELECTED_MODEL_KEY = 'selected_llm_model_id';
const MODEL_API_KEYS_KEY = 'llm_model_api_keys'; // Stores { modelId: apiKey } mapping

/**
 * Get all stored API keys from localStorage
 */
function getAllApiKeys(): Record<string, string> {
  try {
    const stored = storage.getItem(MODEL_API_KEYS_KEY);
    return stored ? JSON.parse(stored) : {};
  } catch {
    return {};
  }
}

/**
 * Get API key for a specific model from localStorage
 */
function getModelApiKey(modelId: string): string | null {
  const keys = getAllApiKeys();
  return keys[modelId] || null;
}

/**
 * Set API key for a specific model in localStorage
 */
function setModelApiKey(modelId: string, apiKey: string | null): void {
  const keys = getAllApiKeys();
  if (apiKey) {
    keys[modelId] = apiKey;
  } else {
    delete keys[modelId];
  }
  storage.setItem(MODEL_API_KEYS_KEY, JSON.stringify(keys));
}

/**
 * Remove API key for a model from localStorage
 */
function removeModelApiKey(modelId: string): void {
  const keys = getAllApiKeys();
  delete keys[modelId];
  storage.setItem(MODEL_API_KEYS_KEY, JSON.stringify(keys));
}

export const useLLMModelStore = defineStore('llmModel', {
  state: (): LLMModelState => ({
    models: [],
    selectedModel: null,
    defaultModel: null,
    useLegacy: true, // Default to legacy settings
    loading: true, // Start true to prevent blank flash on first load
    error: null,
    total: 0,
    pageNumber: 1,
    pageSize: 50,
    totalPages: 0,
    searchQuery: '',
    scope: 'all',
    enabledOnly: false,
  }),

  getters: {
    enabledModels: (state): LLMModel[] => {
      return state.models.filter(m => m.is_enabled);
    },

    hasModels: (state): boolean => {
      return state.models.length > 0;
    },

    getModelById: (state) => (id: string): LLMModel | undefined => {
      return state.models.find(m => m.id === id);
    },

    /**
     * Get the effective LLM configuration for API requests.
     * Returns the selected model config with API key, or null if legacy settings should be used.
     */
    effectiveConfig: (state): { baseUrl: string; model: string; temperature: number; maxTokens: number; apiKey: string | null } | null => {
      if (state.useLegacy || !state.selectedModel) {
        return null; // Use legacy settings from settingsStore
      }
      // Get API key from localStorage for this model
      const apiKey = getModelApiKey(state.selectedModel.id);
      return {
        baseUrl: state.selectedModel.base_url,
        model: state.selectedModel.model,
        temperature: state.selectedModel.temperature,
        maxTokens: state.selectedModel.max_tokens,
        apiKey,
      };
    },

    /**
     * Get model with its locally stored API key
     */
    getModelWithKey: (state) => (id: string): LLMModelWithKey | undefined => {
      const model = state.models.find(m => m.id === id);
      if (!model) return undefined;
      return {
        ...model,
        api_key: getModelApiKey(id) || undefined,
      };
    },
  },

  actions: {
    /**
     * Fetch LLM models from the API
     */
    async fetchModels() {
      this.loading = true;
      this.error = null;
      try {
        const response = await llmModelService.listLLMModels(
          this.scope,
          this.searchQuery || undefined,
          this.enabledOnly,
          this.pageNumber,
          this.pageSize
        );
        this.models = response.data;
        this.total = response.total;
        this.pageNumber = response.pageNumber;
        this.pageSize = response.pageSize;
        this.totalPages = response.totalPages;
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch LLM models';
        showToast(this.error, 'danger');
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch the default LLM model
     */
    async fetchDefaultModel() {
      try {
        const response = await llmModelService.getDefaultLLMModel();
        this.defaultModel = response.model;
        this.useLegacy = response.use_legacy;

        // If no model is selected yet, use the default
        if (!this.selectedModel && this.defaultModel) {
          this.selectedModel = this.defaultModel;
          storage.setItem(SELECTED_MODEL_KEY, this.defaultModel.id);
        }
      } catch (error) {
        console.warn('Failed to fetch default LLM model:', error);
        // Fall back to legacy settings
        this.useLegacy = true;
      }
    },

    /**
     * Initialize the store: load models and restore selected model from localStorage
     */
    async initialize() {
      await this.fetchModels();
      await this.fetchDefaultModel();

      // Try to restore selected model from localStorage
      const savedModelId = storage.getItem(SELECTED_MODEL_KEY);
      if (savedModelId) {
        const model = this.models.find(m => m.id === savedModelId);
        if (model && model.is_enabled) {
          this.selectedModel = model;
          this.useLegacy = false;
        }
      }
    },

    /**
     * Select a model to use for requests
     */
    selectModel(modelId: string | null) {
      if (modelId === null) {
        // Switch to legacy settings
        this.selectedModel = null;
        this.useLegacy = true;
        storage.removeItem(SELECTED_MODEL_KEY);
        showToast('Using legacy settings', 'info');
        return;
      }

      const model = this.models.find(m => m.id === modelId);
      if (model) {
        this.selectedModel = model;
        this.useLegacy = false;
        storage.setItem(SELECTED_MODEL_KEY, modelId);
        showToast(`Selected model: ${model.name}`, 'success');
      }
    },

    /**
     * Create a new LLM model
     */
    async createModel(data: CreateLLMModelRequest) {
      this.loading = true;
      try {
        const newModel = await llmModelService.createLLMModel(data);
        this.models.unshift(newModel);
        this.total++;
        showToast('LLM model created successfully', 'success');
        return newModel;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to create LLM model';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing LLM model
     */
    async updateModel(id: string, data: UpdateLLMModelRequest) {
      this.loading = true;
      try {
        const updated = await llmModelService.updateLLMModel(id, data);
        const index = this.models.findIndex(m => m.id === id);
        if (index !== -1) {
          this.models[index] = updated;
        }
        // Update selected model if it was updated
        if (this.selectedModel?.id === id) {
          this.selectedModel = updated;
        }
        showToast('LLM model updated successfully', 'success');
        return updated;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to update LLM model';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete an LLM model
     */
    async deleteModel(id: string) {
      this.loading = true;
      try {
        await llmModelService.deleteLLMModel(id);
        this.models = this.models.filter(m => m.id !== id);
        this.total--;
        // Clear selection if deleted model was selected
        if (this.selectedModel?.id === id) {
          this.selectedModel = null;
          this.useLegacy = true;
          storage.removeItem(SELECTED_MODEL_KEY);
        }
        // Also remove the stored API key
        removeModelApiKey(id);
        showToast('LLM model deleted successfully', 'success');
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to delete LLM model';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Set API key for a model (stored locally in localStorage, not sent to server)
     */
    setApiKey(modelId: string, apiKey: string | null) {
      setModelApiKey(modelId, apiKey);
      showToast(apiKey ? 'API key saved locally' : 'API key removed', 'success');
    },

    /**
     * Get API key for a model from localStorage
     */
    getApiKey(modelId: string): string | null {
      return getModelApiKey(modelId);
    },

    /**
     * Check if a model has an API key configured
     */
    hasApiKey(modelId: string): boolean {
      return !!getModelApiKey(modelId);
    },

    /**
     * Set a model as the default
     */
    async setDefault(id: string) {
      this.loading = true;
      try {
        const updated = await llmModelService.setDefaultLLMModel(id);
        // Update all models to reflect the new default
        this.models = this.models.map(m => ({
          ...m,
          is_default: m.id === id,
        }));
        this.defaultModel = updated;
        showToast(`${updated.name} set as default`, 'success');
        return updated;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to set default LLM model';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Toggle the enabled status of a model
     */
    async toggleEnabled(id: string, enabled: boolean) {
      this.loading = true;
      try {
        const updated = await llmModelService.toggleLLMModelEnabled(id, enabled);
        const index = this.models.findIndex(m => m.id === id);
        if (index !== -1) {
          this.models[index] = updated;
        }
        // If disabled model was selected, switch to legacy
        if (!enabled && this.selectedModel?.id === id) {
          this.selectedModel = null;
          this.useLegacy = true;
          storage.removeItem(SELECTED_MODEL_KEY);
        }
        showToast(`${updated.name} ${enabled ? 'enabled' : 'disabled'}`, 'success');
        return updated;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to toggle LLM model';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Set search query and refetch
     */
    setSearchQuery(query: string) {
      this.searchQuery = query;
      this.pageNumber = 1;
      this.fetchModels();
    },

    /**
     * Set scope filter and refetch
     */
    setScope(scope: 'all' | 'personal' | 'shared' | 'templates') {
      this.scope = scope;
      this.pageNumber = 1;
      this.fetchModels();
    },

    /**
     * Set enabled-only filter and refetch
     */
    setEnabledOnly(enabledOnly: boolean) {
      this.enabledOnly = enabledOnly;
      this.pageNumber = 1;
      this.fetchModels();
    },

    /**
     * Go to a specific page
     */
    goToPage(page: number) {
      if (page >= 1 && page <= this.totalPages) {
        this.pageNumber = page;
        this.fetchModels();
      }
    },
  },
});

