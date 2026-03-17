import { defineStore } from 'pinia';
import type { Prompt } from '@/types';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';

export interface TagCount {
  tag: string;
  count: number;
}

interface PromptState {
  prompts: Prompt[];
  filteredPrompts: Prompt[];
  selectedPrompt: Prompt | null;
  loading: boolean;
  error: string | null;
  searchQuery: string;
  sortBy: 'name' | 'description' | 'created_at';
  sortOrder: 'asc' | 'desc';
  availableTags: TagCount[];
  selectedTags: string[];
}

export const usePromptStore = defineStore('prompt', {
  state: (): PromptState => ({
    prompts: [],
    filteredPrompts: [],
    selectedPrompt: null,
    loading: true, // Start true to prevent blank flash on first load
    error: null,
    searchQuery: '',
    sortBy: 'name',
    sortOrder: 'asc',
    availableTags: [],
    selectedTags: [],
  }),

  getters: {
    getPromptByName: (state) => (name: string) => {
      return state.prompts.find((p) => p.name === name);
    },

    searchAndSortedPrompts: (state) => {
      let results = [...state.prompts];

      // Apply search filter
      if (state.searchQuery) {
        const query = state.searchQuery.toLowerCase();
        results = results.filter(prompt =>
          prompt.name.toLowerCase().includes(query) ||
          prompt.description?.toLowerCase().includes(query) ||
          prompt.tags?.toLowerCase().includes(query) ||
          prompt.system_prompt?.toLowerCase().includes(query) ||
          prompt.user_prompt?.toLowerCase().includes(query)
        );
      }

      // Apply sorting
      results.sort((a, b) => {
        let aVal: string = '';
        let bVal: string = '';

        switch (state.sortBy) {
          case 'name':
            aVal = a.name || '';
            bVal = b.name || '';
            break;
          case 'description':
            aVal = a.description || '';
            bVal = b.description || '';
            break;
          case 'created_at':
            // If you add created_at field in future
            aVal = '';
            bVal = '';
            break;
        }

        const comparison = aVal.localeCompare(bVal);
        return state.sortOrder === 'asc' ? comparison : -comparison;
      });

      return results;
    },
  },

  actions: {
    async fetchPrompts(tags?: string, searchQuery?: string) {
      this.loading = true;
      this.error = null;
      try {
        this.prompts = await apiService.getPrompts(tags, searchQuery);
        this.filteredPrompts = this.prompts;
        // Only extract tags if no tags filter is applied (to get accurate counts from all prompts)
        if (!tags && !searchQuery) {
          this.extractTags();
        }
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch prompts';
        showToast(this.error, 'danger');
      } finally {
        this.loading = false;
      }
    },

    filterPrompts(module?: string, tags?: string[]) {
      if (!module && (!tags || tags.length === 0)) {
        this.filteredPrompts = this.prompts;
        return;
      }

      this.filteredPrompts = this.prompts.filter((prompt) => {
        if (!prompt.tags) return false;

        const promptTags = prompt.tags.toLowerCase().split(',').map((tag) => tag.trim());

        // Check module
        if (module && !promptTags.includes(module.toLowerCase())) {
          return false;
        }

        // Check tags
        if (tags && tags.length > 0) {
          const hasMatchingTag = tags.some((tag) =>
            promptTags.includes(tag.toLowerCase())
          );
          if (!hasMatchingTag) {
            return false;
          }
        }

        return true;
      });

      // If no prompts match, show all
      if (this.filteredPrompts.length === 0 && this.prompts.length > 0) {
        console.warn(`No prompts found for module: ${module}, tags: ${tags?.join(',')}`);
        this.filteredPrompts = this.prompts;
      }
    },

    selectPrompt(name: string) {
      this.selectedPrompt = this.getPromptByName(name) || null;
    },

    async createPrompt(prompt: Prompt) {
      this.loading = true;
      try {
        const newPrompt = await apiService.createPrompt(prompt);
        this.prompts.push(newPrompt);
        this.filteredPrompts = this.prompts;
        showToast('Prompt created successfully', 'success');
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to create prompt';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async updatePrompt(name: string, prompt: Prompt) {
      this.loading = true;
      try {
        const updated = await apiService.updatePrompt(name, prompt);
        const index = this.prompts.findIndex((p) => p.name === name);
        if (index !== -1) {
          this.prompts[index] = updated;
          this.filteredPrompts = this.prompts;
        }
        showToast('Prompt updated successfully', 'success');
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to update prompt';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deletePrompt(name: string) {
      this.loading = true;
      try {
        await apiService.deletePrompt(name);
        this.prompts = this.prompts.filter((p) => p.name !== name);
        this.filteredPrompts = this.prompts;
        showToast('Prompt deleted successfully', 'success');
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to delete prompt';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    setSearchQuery(query: string) {
      this.searchQuery = query;
    },

    setSorting(sortBy: 'name' | 'description' | 'created_at', sortOrder: 'asc' | 'desc') {
      this.sortBy = sortBy;
      this.sortOrder = sortOrder;
    },

    toggleSortOrder() {
      this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
    },

    // Extract unique tags from all prompts
    extractTags() {
      const tagMap = new Map<string, number>();

      this.prompts.forEach(prompt => {
        if (prompt.tags) {
          const tags = prompt.tags.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
          tags.forEach(tag => {
            const lowerTag = tag.toLowerCase();
            tagMap.set(lowerTag, (tagMap.get(lowerTag) || 0) + 1);
          });
        }
      });

      this.availableTags = Array.from(tagMap.entries())
        .map(([tag, count]) => ({ tag, count }))
        .sort((a, b) => b.count - a.count); // Sort by count descending
    },

    // Toggle tag selection
    toggleTag(tag: string) {
      const index = this.selectedTags.indexOf(tag);
      if (index > -1) {
        this.selectedTags.splice(index, 1);
      } else {
        this.selectedTags.push(tag);
      }
    },

    // Clear all selected tags
    clearTags() {
      this.selectedTags = [];
    },

    // Fetch prompts filtered by selected tags
    async fetchPromptsByTags() {
      if (this.selectedTags.length === 0) {
        // No tags selected, fetch all prompts
        await this.fetchPrompts();
        return;
      }

      this.loading = true;
      this.error = null;
      try {
        // Join tags with comma for API
        const tagsParam = this.selectedTags.join(',');
        this.prompts = await apiService.getPrompts(tagsParam);
        this.filteredPrompts = this.prompts;
        // Don't re-extract tags here - we want to keep the original tag counts from all prompts
        // The counts represent total prompts with each tag, not filtered results
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch prompts by tags';
        showToast(this.error, 'danger');
      } finally {
        this.loading = false;
      }
    },

    // Fetch all prompts for accurate tag counts (no limit)
    async fetchAllPromptsForTags() {
      try {
        // Fetch without tag filter or search to get all prompts
        const allPrompts = await apiService.getPrompts(undefined, undefined);
        // Extract tags from all prompts for accurate counts
        const tagMap = new Map<string, number>();
        allPrompts.forEach(prompt => {
          if (prompt.tags) {
            const tags = prompt.tags.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
            tags.forEach(tag => {
              const lowerTag = tag.toLowerCase();
              tagMap.set(lowerTag, (tagMap.get(lowerTag) || 0) + 1);
            });
          }
        });
        this.availableTags = Array.from(tagMap.entries())
          .map(([tag, count]) => ({ tag, count }))
          .sort((a, b) => b.count - a.count);
      } catch (error) {
        console.error('Failed to fetch all prompts for tag counts:', error);
      }
    },
  },
});
