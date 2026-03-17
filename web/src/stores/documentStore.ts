import { defineStore } from 'pinia';
import type { Document } from '@/types';
import documentService from '@/services/documentService';
import type { LoadURLRequest, CreateFromTextRequest, DocumentStatsResponse } from '@/services/documentService';
import { showToast } from '@/utils/toast';

interface DocumentState {
  documents: Document[];
  selectedDocument: Document | null;
  loading: boolean;
  uploadProgress: number;
  error: string | null;
  searchQuery: string;
  sortBy: 'name' | 'created_at' | 'updated_at';
  sortOrder: 'asc' | 'desc';
  currentProjectId: string | null;
  stats: DocumentStatsResponse | null;
}

export const useDocumentStore = defineStore('document', {
  state: (): DocumentState => ({
    documents: [],
    selectedDocument: null,
    loading: true, // Start true to prevent blank flash on first load
    uploadProgress: 0,
    error: null,
    searchQuery: '',
    sortBy: 'created_at',
    sortOrder: 'desc',
    currentProjectId: null,
    stats: null,
  }),

  getters: {
    getDocumentById: (state) => (id: string) => {
      return state.documents.find((d) => d.id === id);
    },

    documentsByProject: (state) => (projectId: string) => {
      return state.documents.filter((d) => d.project_id === projectId);
    },

    searchAndSortedDocuments: (state) => {
      let results = [...state.documents];

      // Apply search filter
      if (state.searchQuery) {
        const query = state.searchQuery.toLowerCase();
        results = results.filter(doc =>
          doc.name.toLowerCase().includes(query) ||
          doc.path.toLowerCase().includes(query) ||
          doc.content.toLowerCase().includes(query)
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
          case 'created_at':
            aVal = a.created_time || '';
            bVal = b.created_time || '';
            break;
          case 'updated_at':
            aVal = a.updated_time || '';
            bVal = b.updated_time || '';
            break;
        }

        const comparison = aVal.localeCompare(bVal);
        return state.sortOrder === 'asc' ? comparison : -comparison;
      });

      return results;
    },
  },

  actions: {
    async fetchDocuments(params?: {
      scope?: 'all' | 'personal' | 'shared';
      project_id?: string;
      limit?: number;
      offset?: number;
    }) {
      this.loading = true;
      this.error = null;
      try {
        const response = await documentService.listDocuments({
          ...params,
          q: this.searchQuery || undefined,
          sort: this.sortBy,
        });
        this.documents = response.data;
        if (params?.project_id) {
          this.currentProjectId = params.project_id;
        }
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch documents';
        showToast(this.error, 'danger');
      } finally {
        this.loading = false;
      }
    },

    async fetchDocument(id: string) {
      this.loading = true;
      try {
        const document = await documentService.getDocument(id);
        this.selectedDocument = document;
        return document;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to fetch document';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async fetchDocumentChunks(projectId: string, path: string) {
      this.loading = true;
      try {
        const response = await documentService.getDocumentChunks(projectId, path);
        return response.chunks;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to fetch document chunks';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async loadFromURL(request: LoadURLRequest) {
      this.loading = true;
      this.error = null;
      try {
        const response = await documentService.loadFromURL(request);
        showToast(`Successfully loaded document from URL: ${response.chunks_created} chunks created`, 'success');

        // Refresh documents list
        await this.fetchDocuments({ project_id: request.project_id });

        return response;
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to load document from URL';
        showToast(this.error, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async uploadFiles(files: File[], projectId: string, chunkSize: number = 1000, chunkOverlap: number = 200) {
      this.loading = true;
      this.uploadProgress = 0;
      this.error = null;

      try {
        const response = await documentService.uploadFiles(
          files,
          projectId,
          chunkSize,
          chunkOverlap,
          (progress) => {
            this.uploadProgress = progress;
          }
        );

        const successMsg = `Successfully uploaded ${response.files_uploaded} file(s): ${response.chunks_created} chunks created`;
        showToast(successMsg, 'success');

        if (response.errors && response.errors.length > 0) {
          showToast(`Errors occurred: ${response.errors.join(', ')}`, 'warning');
        }

        // Refresh documents list
        await this.fetchDocuments({ project_id: projectId });

        return response;
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to upload files';
        showToast(this.error, 'danger');
        throw error;
      } finally {
        this.loading = false;
        this.uploadProgress = 0;
      }
    },

    async createFromText(request: CreateFromTextRequest) {
      this.loading = true;
      this.error = null;
      try {
        const response = await documentService.createFromText(request);
        showToast(`Successfully created document: ${response.chunks_created} chunks created`, 'success');

        // Refresh documents list
        await this.fetchDocuments({ project_id: request.project_id });

        return response;
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to create document from text';
        showToast(this.error, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deleteDocument(id: string) {
      this.loading = true;
      try {
        await documentService.deleteDocument(id);
        this.documents = this.documents.filter((d) => d.id !== id);
        if (this.selectedDocument?.id === id) {
          this.selectedDocument = null;
        }
        showToast('Document deleted successfully', 'success');
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to delete document';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deleteDocumentByPath(projectId: string, path: string) {
      this.loading = true;
      try {
        const result = await documentService.deleteDocumentByPath(projectId, path);
        showToast(`Successfully deleted ${result.chunks_deleted} chunk(s)`, 'success');

        // Remove from local state
        this.documents = this.documents.filter((d) => !(d.project_id === projectId && d.path === path));

        return result;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to delete document';
        showToast(message, 'danger');
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async fetchStats(projectId?: string) {
      try {
        this.stats = await documentService.getDocumentStats(projectId);
        return this.stats;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to fetch document stats';
        showToast(message, 'danger');
        throw error;
      }
    },

    selectDocument(document: Document) {
      this.selectedDocument = document;
    },

    setSearchQuery(query: string) {
      this.searchQuery = query;
    },

    setSorting(sortBy: 'name' | 'created_at' | 'updated_at', sortOrder: 'asc' | 'desc') {
      this.sortBy = sortBy;
      this.sortOrder = sortOrder;
    },

    toggleSortOrder() {
      this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
    },

    setCurrentProject(projectId: string | null) {
      this.currentProjectId = projectId;
    },

    resetUploadProgress() {
      this.uploadProgress = 0;
    },
  },
});
