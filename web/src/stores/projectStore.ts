import { defineStore } from 'pinia';
import type { Project } from '@/types';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import { storage } from '@/utils/storage';

// Key for storing default project ID in localStorage
const DEFAULT_PROJECT_KEY = 'default_project_id';

interface ProjectState {
  projects: Project[];
  selectedProject: Project | null;
  defaultProjectId: string | null; // ID of the default project
  loading: boolean;
  error: string | null;
  searchQuery: string;
  sortBy: 'name' | 'description' | 'language' | 'created_time';
  sortOrder: 'asc' | 'desc';
}

export const useProjectStore = defineStore('project', {
  state: (): ProjectState => ({
    projects: [],
    selectedProject: null,
    defaultProjectId: null,
    loading: true, // Start true to prevent blank flash on first load
    error: null,
    searchQuery: '',
    sortBy: 'name',
    sortOrder: 'asc',
  }),

  getters: {
    projectNames(): string[] {
      return this.projects.map(p => p.name);
    },

    getProjectById: (state) => (id: string) => {
      return state.projects.find((p) => p.id === id || p.name === id);
    },

    /**
     * Get the default project (if set and exists)
     */
    defaultProject(): Project | null {
      if (!this.defaultProjectId) return null;
      return this.projects.find(p => p.id === this.defaultProjectId) || null;
    },

    /**
     * Check if a project is the default
     */
    isDefaultProject: (state) => (projectId: string): boolean => {
      return state.defaultProjectId === projectId;
    },

    searchAndSortedProjects: (state) => {
      let results = [...state.projects];

      // Apply search filter
      if (state.searchQuery) {
        const query = state.searchQuery.toLowerCase();
        results = results.filter(project =>
          project.name.toLowerCase().includes(query) ||
          project.description?.toLowerCase().includes(query) ||
          project.language?.toLowerCase().includes(query) ||
          project.git_repo?.toLowerCase().includes(query) ||
          project.git_branch?.toLowerCase().includes(query)
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
          case 'language':
            aVal = a.language || '';
            bVal = b.language || '';
            break;
          case 'created_time':
            aVal = a.created_time || '';
            bVal = b.created_time || '';
            break;
        }

        const comparison = aVal.localeCompare(bVal);
        return state.sortOrder === 'asc' ? comparison : -comparison;
      });

      return results;
    },
  },

  actions: {
    /**
     * Load default project ID from localStorage
     */
    loadDefaultFromStorage() {
      this.defaultProjectId = storage.getItem(DEFAULT_PROJECT_KEY) || null;
    },

    async fetchProjects() {
      this.loading = true;
      this.error = null;
      try {
        this.projects = await apiService.getProjects();

        // Load default project ID from storage
        this.loadDefaultFromStorage();

        // Set selected project: prefer default, then first available
        if (this.projects.length > 0 && !this.selectedProject) {
          const defaultProj = this.defaultProject;
          this.selectedProject = defaultProj || this.projects[0];
        }
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch projects';
        showToast(this.error, 'danger');
      } finally {
        this.loading = false;
      }
    },

    async createProject(project: Partial<Project>) {
      this.loading = true;
      try {
        const newProject = await apiService.createProject(project);
        this.projects.push(newProject);
        showToast('Project created successfully', 'success');
        return newProject;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to create project';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async updateProject(id: string, project: Partial<Project>) {
      this.loading = true;
      try {
        const updated = await apiService.updateProject(id, project);
        const index = this.projects.findIndex((p) => p.id === id);
        if (index !== -1) {
          this.projects[index] = updated;
        }
        showToast('Project updated successfully', 'success');
        return updated;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to update project';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deleteProject(id: string) {
      this.loading = true;
      try {
        await apiService.deleteProject(id);
        this.projects = this.projects.filter((p) => p.id !== id);
        if (this.selectedProject?.id === id) {
          this.selectedProject = this.projects.length > 0 ? this.projects[0] : null;
        }
        showToast('Project deleted successfully', 'success');
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to delete project';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    selectProject(project: Project) {
      this.selectedProject = project;
    },

    /**
     * Set a project as the default (persisted in localStorage)
     */
    setDefaultProject(projectId: string | null) {
      if (projectId) {
        const project = this.projects.find(p => p.id === projectId);
        if (project) {
          this.defaultProjectId = projectId;
          storage.setItem(DEFAULT_PROJECT_KEY, projectId);
          showToast(`"${project.name}" set as default project`, 'success');
        }
      } else {
        this.defaultProjectId = null;
        storage.removeItem(DEFAULT_PROJECT_KEY);
        showToast('Default project cleared', 'info');
      }
    },

    /**
     * Clear the default project
     */
    clearDefaultProject() {
      this.setDefaultProject(null);
    },

    getProject(nameOrId: string): Project | null {
      return this.projects.find(p => p.name === nameOrId || p.id === nameOrId) || null;
    },

    setSearchQuery(query: string) {
      this.searchQuery = query;
    },

    setSorting(sortBy: 'name' | 'description' | 'language' | 'created_time', sortOrder: 'asc' | 'desc') {
      this.sortBy = sortBy;
      this.sortOrder = sortOrder;
    },

    toggleSortOrder() {
      this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
    },

    async exportProjects(scope: string = 'all'): Promise<Blob> {
      try {
        return await apiService.exportProjects(scope);
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to export projects';
        showToast(message, 'danger');
        throw error;
      }
    },

    async importProjects(projects: any[], updateExisting: boolean = false, scope: string = 'personal'): Promise<any> {
      try {
        const result = await apiService.importProjects(projects, updateExisting, scope);
        return result;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to import projects';
        showToast(message, 'danger');
        throw error;
      }
    },
  },
});
