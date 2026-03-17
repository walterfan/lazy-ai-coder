import { defineStore } from 'pinia';
import type { CursorRule } from '@/types/cursorRule';
import { cursorRuleService } from '@/services/cursorRuleService';
import { showToast } from '@/utils/toast';

interface CursorRuleState {
  cursorRules: CursorRule[];
  filteredRules: CursorRule[];
  selectedRule: CursorRule | null;
  loading: boolean;
  error: string | null;
  searchQuery: string;
  languageFilter: string;
  frameworkFilter: string;
  tagsFilter: string;
  scope: 'all' | 'personal' | 'shared' | 'templates';
  sortBy: 'name' | 'updated_at' | 'usage_count' | 'created_at';
  sortOrder: 'asc' | 'desc';
}

export const useCursorRuleStore = defineStore('cursorRule', {
  state: (): CursorRuleState => ({
    cursorRules: [],
    filteredRules: [],
    selectedRule: null,
    loading: true, // Start true to prevent blank flash on first load
    error: null,
    searchQuery: '',
    languageFilter: '',
    frameworkFilter: '',
    tagsFilter: '',
    scope: 'all',
    sortBy: 'created_at',
    sortOrder: 'desc',
  }),

  getters: {
    getRuleById: (state) => (id: string) => {
      return state.cursorRules.find((r) => r.id === id);
    },

    getRuleByName: (state) => (name: string) => {
      return state.cursorRules.find((r) => r.name === name);
    },

    templates: (state) => {
      return state.cursorRules.filter((r) => r.is_template);
    },

    byLanguage: (state) => (language: string) => {
      return state.cursorRules.filter((r) => r.language === language);
    },

    byFramework: (state) => (framework: string) => {
      return state.cursorRules.filter((r) => r.framework === framework);
    },

    uniqueLanguages: (state) => {
      const languages = new Set<string>();
      state.cursorRules.forEach((r) => {
        if (r.language) {
          languages.add(r.language);
        }
      });
      return Array.from(languages).sort();
    },

    uniqueFrameworks: (state) => {
      const frameworks = new Set<string>();
      state.cursorRules.forEach((r) => {
        if (r.framework) {
          frameworks.add(r.framework);
        }
      });
      return Array.from(frameworks).sort();
    },

    uniqueTags: (state) => {
      const tags = new Set<string>();
      state.cursorRules.forEach((r) => {
        if (r.tags) {
          r.tags.split(',').forEach((tag) => {
            const trimmed = tag.trim();
            if (trimmed) {
              tags.add(trimmed);
            }
          });
        }
      });
      return Array.from(tags).sort();
    },

    searchAndSortedRules: (state) => {
      let results = [...state.cursorRules];

      // Apply search filter
      if (state.searchQuery) {
        const query = state.searchQuery.toLowerCase();
        results = results.filter(
          (rule) =>
            rule.name.toLowerCase().includes(query) ||
            rule.description?.toLowerCase().includes(query) ||
            rule.content?.toLowerCase().includes(query) ||
            rule.tags?.toLowerCase().includes(query)
        );
      }

      // Apply language filter
      if (state.languageFilter) {
        results = results.filter((rule) => rule.language === state.languageFilter);
      }

      // Apply framework filter
      if (state.frameworkFilter) {
        results = results.filter((rule) => rule.framework === state.frameworkFilter);
      }

      // Apply tags filter
      if (state.tagsFilter) {
        results = results.filter((rule) => rule.tags?.includes(state.tagsFilter));
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
    async fetchCursorRules() {
      this.loading = true;
      this.error = null;
      try {
        const response = await cursorRuleService.listCursorRules({
          scope: this.scope,
          q: this.searchQuery || undefined,
          tags: this.tagsFilter || undefined,
          language: this.languageFilter || undefined,
          framework: this.frameworkFilter || undefined,
          sort: this.sortBy,
        });
        this.cursorRules = response.data;
        this.filteredRules = this.searchAndSortedRules;
      } catch (error: any) {
        this.error = error.message || 'Failed to fetch cursor rules';
        showToast(this.error || 'Failed to fetch cursor rules', 'danger');
      } finally {
        this.loading = false;
      }
    },

    async fetchCursorRule(id: string) {
      this.loading = true;
      this.error = null;
      try {
        const rule = await cursorRuleService.getCursorRule(id);
        this.selectedRule = rule;
        return rule;
      } catch (error: any) {
        this.error = error.message || 'Failed to fetch cursor rule';
        showToast(this.error || 'Failed to fetch cursor rule', 'danger');
        return null;
      } finally {
        this.loading = false;
      }
    },

    async createCursorRule(rule: Partial<CursorRule>) {
      this.loading = true;
      this.error = null;
      try {
        const newRule = await cursorRuleService.createCursorRule(rule);
        this.cursorRules.push(newRule);
        this.filteredRules = this.searchAndSortedRules;
        showToast('Cursor rule created successfully', 'success');
        return newRule;
      } catch (error: any) {
        this.error = error.message || 'Failed to create cursor rule';
        showToast(this.error || 'Failed to create cursor rule', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async updateCursorRule(id: string, rule: Partial<CursorRule>) {
      this.loading = true;
      this.error = null;
      try {
        const updatedRule = await cursorRuleService.updateCursorRule(id, rule);
        const index = this.cursorRules.findIndex((r) => r.id === id);
        if (index !== -1) {
          this.cursorRules[index] = updatedRule;
        }
        if (this.selectedRule?.id === id) {
          this.selectedRule = updatedRule;
        }
        this.filteredRules = this.searchAndSortedRules;
        showToast('Cursor rule updated successfully', 'success');
        return updatedRule;
      } catch (error: any) {
        this.error = error.message || 'Failed to update cursor rule';
        showToast(this.error || 'Failed to update cursor rule', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deleteCursorRule(id: string) {
      this.loading = true;
      this.error = null;
      try {
        await cursorRuleService.deleteCursorRule(id);
        this.cursorRules = this.cursorRules.filter((r) => r.id !== id);
        if (this.selectedRule?.id === id) {
          this.selectedRule = null;
        }
        this.filteredRules = this.searchAndSortedRules;
        showToast('Cursor rule deleted successfully', 'success');
      } catch (error: any) {
        this.error = error.message || 'Failed to delete cursor rule';
        showToast(this.error || 'Failed to delete cursor rule', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async generateCursorRule(request: any) {
      this.loading = true;
      this.error = null;
      try {
        const response = await cursorRuleService.generateCursorRule(request);
        return response.content;
      } catch (error: any) {
        this.error = error.message || 'Failed to generate cursor rule';
        showToast(this.error || 'Failed to generate cursor rule', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async refineCursorRule(id: string, request: any) {
      this.loading = true;
      this.error = null;
      try {
        const response = await cursorRuleService.refineCursorRule(id, request);
        return response.content;
      } catch (error: any) {
        this.error = error.message || 'Failed to refine cursor rule';
        showToast(this.error || 'Failed to refine cursor rule', 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    setSearchQuery(query: string) {
      this.searchQuery = query;
      this.filteredRules = this.searchAndSortedRules;
    },

    setLanguageFilter(language: string) {
      this.languageFilter = language;
      this.filteredRules = this.searchAndSortedRules;
    },

    setFrameworkFilter(framework: string) {
      this.frameworkFilter = framework;
      this.filteredRules = this.searchAndSortedRules;
    },

    setTagsFilter(tags: string) {
      this.tagsFilter = tags;
      this.filteredRules = this.searchAndSortedRules;
    },

    setScope(scope: 'all' | 'personal' | 'shared' | 'templates') {
      this.scope = scope;
      this.fetchCursorRules();
    },

    setSortBy(sortBy: 'name' | 'updated_at' | 'usage_count' | 'created_at') {
      this.sortBy = sortBy;
      this.filteredRules = this.searchAndSortedRules;
    },

    setSortOrder(order: 'asc' | 'desc') {
      this.sortOrder = order;
      this.filteredRules = this.searchAndSortedRules;
    },
  },
});

