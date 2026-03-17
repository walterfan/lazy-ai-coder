import { defineStore } from 'pinia';
import type { CursorCommand } from '@/types/cursorCommand';
import { cursorCommandService } from '@/services/cursorCommandService';
import { showToast } from '@/utils/toast';

interface CursorCommandState {
  cursorCommands: CursorCommand[];
  filteredCommands: CursorCommand[];
  selectedCommand: CursorCommand | null;
  loading: boolean;
  error: string | null;
  searchQuery: string;
  categoryFilter: string;
  languageFilter: string;
  frameworkFilter: string;
  tagsFilter: string;
  scope: 'all' | 'personal' | 'shared' | 'templates';
  sortBy: 'name' | 'updated_at' | 'usage_count' | 'created_at';
  sortOrder: 'asc' | 'desc';
}

export const useCursorCommandStore = defineStore('cursorCommand', {
  state: (): CursorCommandState => ({
    cursorCommands: [],
    filteredCommands: [],
    selectedCommand: null,
    loading: true, // Start true to prevent blank flash on first load
    error: null,
    searchQuery: '',
    categoryFilter: '',
    languageFilter: '',
    frameworkFilter: '',
    tagsFilter: '',
    scope: 'all',
    sortBy: 'created_at',
    sortOrder: 'desc',
  }),

  getters: {
    getCommandById: (state) => (id: string) => {
      return state.cursorCommands.find((c) => c.id === id);
    },

    getCommandByName: (state) => (name: string) => {
      return state.cursorCommands.find((c) => c.name === name);
    },

    templates: (state) => {
      return state.cursorCommands.filter((c) => c.is_template);
    },

    byCategory: (state) => (category: string) => {
      return state.cursorCommands.filter((c) => c.category === category);
    },

    byLanguage: (state) => (language: string) => {
      return state.cursorCommands.filter((c) => c.language === language);
    },

    byFramework: (state) => (framework: string) => {
      return state.cursorCommands.filter((c) => c.framework === framework);
    },

    uniqueCategories: (state) => {
      const categories = new Set<string>();
      state.cursorCommands.forEach((c) => {
        if (c.category) {
          categories.add(c.category);
        }
      });
      return Array.from(categories).sort();
    },

    uniqueLanguages: (state) => {
      const languages = new Set<string>();
      state.cursorCommands.forEach((c) => {
        if (c.language) {
          languages.add(c.language);
        }
      });
      return Array.from(languages).sort();
    },

    uniqueFrameworks: (state) => {
      const frameworks = new Set<string>();
      state.cursorCommands.forEach((c) => {
        if (c.framework) {
          frameworks.add(c.framework);
        }
      });
      return Array.from(frameworks).sort();
    },

    uniqueTags: (state) => {
      const tags = new Set<string>();
      state.cursorCommands.forEach((c) => {
        if (c.tags) {
          c.tags.split(',').forEach((tag) => {
            const trimmed = tag.trim();
            if (trimmed) {
              tags.add(trimmed);
            }
          });
        }
      });
      return Array.from(tags).sort();
    },

    searchAndSortedCommands: (state) => {
      let results = [...state.cursorCommands];

      // Apply search filter
      if (state.searchQuery) {
        const query = state.searchQuery.toLowerCase();
        results = results.filter(
          (cmd) =>
            cmd.name.toLowerCase().includes(query) ||
            cmd.description?.toLowerCase().includes(query) ||
            cmd.command?.toLowerCase().includes(query) ||
            cmd.tags?.toLowerCase().includes(query)
        );
      }

      // Apply category filter
      if (state.categoryFilter) {
        results = results.filter((cmd) => cmd.category === state.categoryFilter);
      }

      // Apply language filter
      if (state.languageFilter) {
        results = results.filter((cmd) => cmd.language === state.languageFilter);
      }

      // Apply framework filter
      if (state.frameworkFilter) {
        results = results.filter((cmd) => cmd.framework === state.frameworkFilter);
      }

      // Apply tags filter
      if (state.tagsFilter) {
        results = results.filter((cmd) => cmd.tags?.includes(state.tagsFilter));
      }

      // Apply sorting
      results.sort((a, b) => {
        let comparison = 0;

        switch (state.sortBy) {
          case 'name':
            comparison = (a.name || '').localeCompare(b.name || '');
            break;
          case 'usage_count':
            comparison = a.usage_count - b.usage_count;
            break;
          case 'updated_at':
            comparison =
              new Date(a.updated_time || 0).getTime() -
              new Date(b.updated_time || 0).getTime();
            break;
          case 'created_at':
            comparison =
              new Date(a.created_time || 0).getTime() -
              new Date(b.created_time || 0).getTime();
            break;
        }

        return state.sortOrder === 'asc' ? comparison : -comparison;
      });

      return results;
    },
  },

  actions: {
    async fetchCursorCommands() {
      this.loading = true;
      this.error = null;
      try {
        const response = await cursorCommandService.listCursorCommands({
          scope: this.scope,
          q: this.searchQuery || undefined,
          tags: this.tagsFilter || undefined,
          category: this.categoryFilter || undefined,
          language: this.languageFilter || undefined,
          framework: this.frameworkFilter || undefined,
          sort: this.sortBy,
        });
        this.cursorCommands = response.data;
        this.filteredCommands = this.searchAndSortedCommands;
      } catch (error: any) {
        this.error = error.message || 'Failed to fetch cursor commands';
        showToast(this.error || 'Failed to fetch cursor commands', 'danger');
      } finally {
        this.loading = false;
      }
    },

    async fetchCursorCommand(id: string) {
      this.loading = true;
      this.error = null;
      try {
        const cmd = await cursorCommandService.getCursorCommand(id);
        this.selectedCommand = cmd;
        return cmd;
      } catch (error: any) {
        this.error = error.message || 'Failed to fetch cursor command';
        showToast(this.error || 'Failed to fetch cursor command', 'danger');
        return null;
      } finally {
        this.loading = false;
      }
    },

    async createCursorCommand(command: Partial<CursorCommand>) {
      this.loading = true;
      this.error = null;
      try {
        const newCommand = await cursorCommandService.createCursorCommand(command);
        this.cursorCommands.push(newCommand);
        this.filteredCommands = this.searchAndSortedCommands;
        showToast('Cursor command created successfully', 'success');
        return newCommand;
      } catch (error: any) {
        this.error = error.message || 'Failed to create cursor command';
        showToast(this.error || 'Failed to create cursor command', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async updateCursorCommand(id: string, command: Partial<CursorCommand>) {
      this.loading = true;
      this.error = null;
      try {
        const updatedCommand = await cursorCommandService.updateCursorCommand(id, command);
        const index = this.cursorCommands.findIndex((c) => c.id === id);
        if (index !== -1) {
          this.cursorCommands[index] = updatedCommand;
        }
        if (this.selectedCommand?.id === id) {
          this.selectedCommand = updatedCommand;
        }
        this.filteredCommands = this.searchAndSortedCommands;
        showToast('Cursor command updated successfully', 'success');
        return updatedCommand;
      } catch (error: any) {
        this.error = error.message || 'Failed to update cursor command';
        showToast(this.error || 'Failed to update cursor command', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deleteCursorCommand(id: string) {
      this.loading = true;
      this.error = null;
      try {
        await cursorCommandService.deleteCursorCommand(id);
        this.cursorCommands = this.cursorCommands.filter((c) => c.id !== id);
        if (this.selectedCommand?.id === id) {
          this.selectedCommand = null;
        }
        this.filteredCommands = this.searchAndSortedCommands;
        showToast('Cursor command deleted successfully', 'success');
      } catch (error: any) {
        this.error = error.message || 'Failed to delete cursor command';
        showToast(this.error || 'Failed to delete cursor command', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async generateCursorCommand(request: any) {
      this.loading = true;
      this.error = null;
      try {
        const response = await cursorCommandService.generateCursorCommand(request);
        return response.command;
      } catch (error: any) {
        this.error = error.message || 'Failed to generate cursor command';
        showToast(this.error || 'Failed to generate cursor command', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async refineCursorCommand(id: string, request: any) {
      this.loading = true;
      this.error = null;
      try {
        const response = await cursorCommandService.refineCursorCommand(id, request);
        return response.command;
      } catch (error: any) {
        this.error = error.message || 'Failed to refine cursor command';
        showToast(this.error || 'Failed to refine cursor command', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    setSearchQuery(query: string) {
      this.searchQuery = query;
      this.filteredCommands = this.searchAndSortedCommands;
    },

    setCategoryFilter(category: string) {
      this.categoryFilter = category;
      this.filteredCommands = this.searchAndSortedCommands;
    },

    setLanguageFilter(language: string) {
      this.languageFilter = language;
      this.filteredCommands = this.searchAndSortedCommands;
    },

    setFrameworkFilter(framework: string) {
      this.frameworkFilter = framework;
      this.filteredCommands = this.searchAndSortedCommands;
    },

    setTagsFilter(tags: string) {
      this.tagsFilter = tags;
      this.filteredCommands = this.searchAndSortedCommands;
    },

    setScope(scope: 'all' | 'personal' | 'shared' | 'templates') {
      this.scope = scope;
      this.fetchCursorCommands();
    },

    setSortBy(sortBy: 'name' | 'updated_at' | 'usage_count' | 'created_at') {
      this.sortBy = sortBy;
      this.filteredCommands = this.searchAndSortedCommands;
    },

    setSortOrder(order: 'asc' | 'desc') {
      this.sortOrder = order;
      this.filteredCommands = this.searchAndSortedCommands;
    },
  },
});

