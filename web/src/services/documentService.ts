import { apiService } from './apiService';
import type { Document } from '@/types';

export interface LoadURLRequest {
  url: string;
  project_id: string;
  chunk_size?: number;
  chunk_overlap?: number;
}

export interface LoadURLResponse {
  message: string;
  url: string;
  project_id: string;
  chunks_created: number;
  embeddings_stored: number;
  stats: {
    FilesProcessed: number;
    FilesSkipped: number;
    CodeChunks: number;
    DocumentChunks: number;
    TotalChunks: number;
    EmbeddingsStored: number;
    Errors: string[];
  };
}

export interface UploadFilesResponse {
  message: string;
  files_uploaded: number;
  files_processed: number;
  chunks_created: number;
  code_chunks: number;
  document_chunks: number;
  errors: string[];
}

export interface CreateFromTextRequest {
  name: string;
  content: string;
  project_id?: string;      // Optional: existing project ID
  project_name?: string;    // Optional: new project name
  chunk_size?: number;
  chunk_overlap?: number;
}

export interface CreateFromTextResponse {
  message: string;
  name: string;
  project_id: string;
  project_name?: string;
  chunks_created: number;
  embeddings_stored: number;
  stats: {
    FilesProcessed: number;
    FilesSkipped: number;
    CodeChunks: number;
    DocumentChunks: number;
    TotalChunks: number;
    EmbeddingsStored: number;
    Errors: string[];
  };
}

export interface DocumentListResponse {
  data: Document[];
  total: number;
  limit: number;
  offset: number;
}

export interface DocumentChunksResponse {
  chunks: Document[];
  total: number;
}

export interface DocumentStatsResponse {
  total_chunks: number;
  unique_documents: number;
}

class DocumentService {
  /**
   * List documents with filtering and pagination
   */
  async listDocuments(params: {
    scope?: 'all' | 'personal' | 'shared';
    project_id?: string;
    q?: string;
    sort?: 'created_at' | 'updated_at' | 'name';
    limit?: number;
    offset?: number;
  } = {}): Promise<DocumentListResponse> {
    const response = await apiService.get<DocumentListResponse>('/documents', { params });
    return response.data;
  }

  /**
   * Get a single document by ID
   */
  async getDocument(id: string): Promise<Document> {
    const response = await apiService.get<Document>(`/documents/${id}`);
    return response.data;
  }

  /**
   * Get all chunks for a specific document
   */
  async getDocumentChunks(project_id: string, path: string): Promise<DocumentChunksResponse> {
    const response = await apiService.get<DocumentChunksResponse>('/documents/chunks', {
      params: { project_id, path },
    });
    return response.data;
  }

  /**
   * Load document from URL
   */
  async loadFromURL(request: LoadURLRequest): Promise<LoadURLResponse> {
    const response = await apiService.post<LoadURLResponse>('/documents/load-url', request);
    return response.data;
  }

  /**
   * Upload files (single or multiple)
   */
  async uploadFiles(
    files: File[],
    projectId: string,
    chunkSize: number = 1000,
    chunkOverlap: number = 200,
    onProgress?: (progress: number) => void
  ): Promise<UploadFilesResponse> {
    const formData = new FormData();

    // Append all files
    files.forEach((file) => {
      formData.append('files', file);
    });

    formData.append('project_id', projectId);
    formData.append('chunk_size', chunkSize.toString());
    formData.append('chunk_overlap', chunkOverlap.toString());

    const response = await apiService.post<UploadFilesResponse>(
      '/documents/upload',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent: any) => {
          if (onProgress && progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
            onProgress(progress);
          }
        },
      }
    );
    return response.data;
  }

  /**
   * Create document from text input
   */
  async createFromText(request: CreateFromTextRequest): Promise<CreateFromTextResponse> {
    const response = await apiService.post<CreateFromTextResponse>(
      '/documents/create-from-text',
      request
    );
    return response.data;
  }

  /**
   * Delete a document by ID
   */
  async deleteDocument(id: string): Promise<void> {
    await apiService.delete(`/documents/${id}`);
  }

  /**
   * Delete all chunks of a document by path
   */
  async deleteDocumentByPath(project_id: string, path: string): Promise<{ chunks_deleted: number }> {
    const response = await apiService.post<{ message: string; chunks_deleted: number }>(
      '/documents/delete-by-path',
      { project_id, path }
    );
    return response.data;
  }

  /**
   * Get document statistics
   */
  async getDocumentStats(project_id?: string): Promise<DocumentStatsResponse> {
    const response = await apiService.get<DocumentStatsResponse>('/documents/stats', {
      params: project_id ? { project_id } : {},
    });
    return response.data;
  }
}

export default new DocumentService();
