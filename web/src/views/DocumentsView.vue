<template>
  <div class="container mt-4 mb-5">
    <!-- Header with Action Buttons -->
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-file-alt"></i> Notes & Documents
      </h2>
      <div class="btn-group">
        <button
          v-if="authStore.canModify"
          @click="showURLLoaderModal = true"
          class="btn btn-outline-primary"
          title="Load document from URL"
        >
          <i class="fas fa-link"></i> Load from URL
        </button>
        <button
          v-if="authStore.canModify"
          @click="showFileUploaderModal = true"
          class="btn btn-outline-primary"
          title="Upload files"
        >
          <i class="fas fa-upload"></i> Upload Files
        </button>
        <button
          v-if="authStore.canModify"
          @click="showTextInputModal = true"
          class="btn btn-primary"
          title="Create from text input"
        >
          <i class="fas fa-keyboard"></i> Create from Text
        </button>
        <button
          v-else
          @click="showSignupModal = true"
          class="btn btn-primary"
          title="Sign up to create documents"
        >
          <i class="fas fa-lock"></i> Sign Up to Create
        </button>
      </div>
    </div>

    <!-- Stats Card -->
    <div v-if="documentStore.stats" class="card mb-4 bg-light">
      <div class="card-body">
        <div class="row text-center">
          <div class="col-md-6">
            <h4 class="mb-0">{{ documentStore.stats.unique_documents }}</h4>
            <small class="text-muted">Documents</small>
          </div>
          <div class="col-md-6">
            <h4 class="mb-0">{{ documentStore.stats.total_chunks }}</h4>
            <small class="text-muted">Total Chunks (Nodes)</small>
          </div>
        </div>
      </div>
    </div>

    <!-- Search and Filter Controls -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row g-3">
          <!-- Search Bar -->
          <div class="col-md-6">
            <div class="input-group">
              <span class="input-group-text">
                <i class="fas fa-search"></i>
              </span>
              <input
                type="text"
                class="form-control"
                placeholder="Search by name, path, or content..."
                v-model="searchQuery"
                @input="handleSearch"
              />
              <button
                v-if="searchQuery"
                class="btn btn-outline-secondary"
                @click="clearSearch"
              >
                <i class="fas fa-times"></i>
              </button>
            </div>
          </div>

          <!-- Project Filter -->
          <div class="col-md-3">
            <select class="form-select" v-model="selectedProjectId" @change="handleProjectChange">
              <option value="">All Projects</option>
              <option v-for="project in projects" :key="project.id" :value="project.id">
                {{ project.name }}
              </option>
            </select>
          </div>

          <!-- Sort By -->
          <div class="col-md-3">
            <select class="form-select" v-model="sortBy" @change="handleSortChange">
              <option value="created_at">Sort by Created Time</option>
              <option value="updated_at">Sort by Updated Time</option>
              <option value="name">Sort by Name</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="documentStore.loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="documentStore.documents.length === 0 && searchQuery" class="alert alert-warning">
      <i class="fas fa-search"></i> No documents found matching "{{ searchQuery }}".
      <button class="btn btn-link p-0" @click="clearSearch">Clear search</button>
    </div>

    <div v-else-if="documentStore.documents.length === 0" class="alert alert-info">
      <i class="fas fa-info-circle"></i> No documents found. Upload files or load from URL to get started!
    </div>

    <!-- Documents List (grouped by path) -->
    <div v-else class="documents-list">
      <div
        v-for="(chunks, path) in groupedDocuments"
        :key="path"
        class="card mb-3"
      >
        <div class="card-header d-flex justify-content-between align-items-center">
          <div>
            <h5 class="mb-0">
              <i class="fas fa-file"></i> {{ getDocumentName(chunks[0]) }}
            </h5>
            <small class="text-muted">{{ path }}</small>
          </div>
          <div>
            <span class="badge bg-primary me-2">{{ chunks.length }} chunk(s)</span>
            <button
              class="btn btn-sm btn-outline-info me-2"
              @click="viewChunks(chunks)"
              title="View chunks"
            >
              <i class="fas fa-eye"></i>
            </button>
            <button
              v-if="authStore.canModify"
              class="btn btn-sm btn-outline-danger"
              @click="deleteDocument(chunks[0].project_id, path)"
              title="Delete document"
            >
              <i class="fas fa-trash"></i>
            </button>
          </div>
        </div>
        <div class="card-body">
          <p class="mb-2"><strong>Project:</strong> {{ getProjectName(chunks[0].project_id) }}</p>
          <p class="mb-2"><strong>Created:</strong> {{ formatDate(chunks[0].created_time) }}</p>
          <p class="mb-0 text-truncate"><strong>Preview:</strong> {{ chunks[0].content.substring(0, 200) }}...</p>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <nav v-if="totalPages > 1" aria-label="Documents pagination" class="mt-4">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: currentPage === 1 }">
          <a class="page-link" @click="changePage(currentPage - 1)">Previous</a>
        </li>
        <li v-for="page in totalPages" :key="page" class="page-item" :class="{ active: page === currentPage }">
          <a class="page-link" @click="changePage(page)">{{ page }}</a>
        </li>
        <li class="page-item" :class="{ disabled: currentPage === totalPages }">
          <a class="page-link" @click="changePage(currentPage + 1)">Next</a>
        </li>
      </ul>
    </nav>

    <!-- URL Loader Modal -->
    <div v-if="showURLLoaderModal" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Load Document from URL</h5>
            <button type="button" class="btn-close" @click="showURLLoaderModal = false"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="handleURLLoad">
              <div class="mb-3">
                <label class="form-label">URL</label>
                <input
                  type="url"
                  class="form-control"
                  v-model="urlLoadRequest.url"
                  placeholder="https://example.com/article"
                  required
                />
              </div>
              <div class="mb-3">
                <label class="form-label">Project</label>
                <select class="form-select" v-model="urlLoadRequest.project_id" required>
                  <option value="">Select a project...</option>
                  <option v-for="project in projects" :key="project.id" :value="project.id">
                    {{ project.name }}
                  </option>
                </select>
              </div>
              <div class="row">
                <div class="col-md-6 mb-3">
                  <label class="form-label">Chunk Size</label>
                  <input type="number" class="form-control" v-model.number="urlLoadRequest.chunk_size" />
                </div>
                <div class="col-md-6 mb-3">
                  <label class="form-label">Chunk Overlap</label>
                  <input type="number" class="form-control" v-model.number="urlLoadRequest.chunk_overlap" />
                </div>
              </div>
              <div class="d-grid">
                <button type="submit" class="btn btn-primary" :disabled="documentStore.loading">
                  <span v-if="documentStore.loading" class="spinner-border spinner-border-sm me-2"></span>
                  Load Document
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <!-- File Uploader Modal -->
    <div v-if="showFileUploaderModal" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Upload Files</h5>
            <button type="button" class="btn-close" @click="showFileUploaderModal = false"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="handleFileUpload">
              <div class="mb-3">
                <label class="form-label">Project</label>
                <select class="form-select" v-model="uploadRequest.project_id" required>
                  <option value="">Select a project...</option>
                  <option v-for="project in projects" :key="project.id" :value="project.id">
                    {{ project.name }}
                  </option>
                </select>
              </div>
              <div class="mb-3">
                <label class="form-label">Files</label>
                <div
                  class="dropzone"
                  @drop.prevent="handleDrop"
                  @dragover.prevent="dragOver = true"
                  @dragleave.prevent="dragOver = false"
                  :class="{ 'drag-over': dragOver }"
                >
                  <input
                    type="file"
                    ref="fileInput"
                    multiple
                    @change="handleFileSelect"
                    class="d-none"
                    accept=".pdf,.docx,.pptx,.md,.txt,.go,.py,.java,.js,.ts,.cpp,.c,.h"
                  />
                  <div v-if="selectedFiles.length === 0" class="dropzone-content" @click="fileInput?.click()">
                    <i class="fas fa-cloud-upload-alt fa-3x mb-3 text-muted"></i>
                    <p class="mb-0">Drag & drop files here or click to browse</p>
                    <small class="text-muted">Supported: PDF, DOCX, PPTX, MD, TXT, Code files</small>
                  </div>
                  <div v-else class="selected-files">
                    <div v-for="(file, index) in selectedFiles" :key="index" class="file-item">
                      <i class="fas fa-file me-2"></i>
                      <span>{{ file.name }} ({{ formatFileSize(file.size) }})</span>
                      <button type="button" class="btn btn-sm btn-link text-danger" @click="removeFile(index)">
                        <i class="fas fa-times"></i>
                      </button>
                    </div>
                    <button type="button" class="btn btn-sm btn-link" @click="fileInput?.click()">
                      <i class="fas fa-plus"></i> Add more files
                    </button>
                  </div>
                </div>
              </div>
              <div class="row">
                <div class="col-md-6 mb-3">
                  <label class="form-label">Chunk Size</label>
                  <input type="number" class="form-control" v-model.number="uploadRequest.chunk_size" />
                </div>
                <div class="col-md-6 mb-3">
                  <label class="form-label">Chunk Overlap</label>
                  <input type="number" class="form-control" v-model.number="uploadRequest.chunk_overlap" />
                </div>
              </div>
              <div v-if="documentStore.uploadProgress > 0" class="mb-3">
                <div class="progress">
                  <div
                    class="progress-bar progress-bar-striped progress-bar-animated"
                    :style="{ width: documentStore.uploadProgress + '%' }"
                  >
                    {{ documentStore.uploadProgress }}%
                  </div>
                </div>
              </div>
              <div class="d-grid">
                <button type="submit" class="btn btn-primary" :disabled="documentStore.loading || selectedFiles.length === 0">
                  <span v-if="documentStore.loading" class="spinner-border spinner-border-sm me-2"></span>
                  Upload Files
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <!-- Chunks Viewer Modal -->
    <div v-if="showChunksModal" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Document Chunks (Nodes)</h5>
            <button type="button" class="btn-close" @click="showChunksModal = false"></button>
          </div>
          <div class="modal-body" style="max-height: 70vh; overflow-y: auto;">
            <div v-for="(chunk, index) in viewingChunks" :key="chunk.id" class="card mb-3">
              <div class="card-header">
                <strong>Chunk {{ index + 1 }}</strong>
                <span class="badge bg-secondary ms-2">{{ chunk.content.length }} chars</span>
              </div>
              <div class="card-body">
                <pre class="mb-0" style="white-space: pre-wrap;">{{ chunk.content }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Text Input Modal -->
    <div v-if="showTextInputModal" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Create Document from Text</h5>
            <button type="button" class="btn-close" @click="showTextInputModal = false"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="handleTextCreate">
              <div class="mb-3">
                <label class="form-label">Document Name</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="textInputRequest.name"
                  placeholder="my-document.txt"
                  required
                />
                <small class="form-text text-muted">
                  Include file extension (e.g., .txt, .md, .py)
                </small>
              </div>
              <div class="mb-3">
                <label class="form-label">Project</label>
                <div class="btn-group w-100 mb-2" role="group">
                  <input
                    type="radio"
                    class="btn-check"
                    name="projectMode"
                    id="existingProject"
                    value="existing"
                    v-model="projectMode"
                  />
                  <label class="btn btn-outline-primary" for="existingProject">
                    <i class="fas fa-folder"></i> Select Existing
                  </label>
                  <input
                    type="radio"
                    class="btn-check"
                    name="projectMode"
                    id="newProject"
                    value="new"
                    v-model="projectMode"
                  />
                  <label class="btn btn-outline-primary" for="newProject">
                    <i class="fas fa-plus-circle"></i> Create New
                  </label>
                </div>

                <!-- Existing Project Selector -->
                <select
                  v-if="projectMode === 'existing'"
                  class="form-select"
                  v-model="textInputRequest.project_id"
                  required
                >
                  <option value="">Select a project...</option>
                  <option v-for="project in projects" :key="project.id" :value="project.id">
                    {{ project.name }}
                  </option>
                </select>

                <!-- New Project Name Input -->
                <input
                  v-else
                  type="text"
                  class="form-control"
                  v-model="textInputRequest.project_name"
                  placeholder="Enter new project name..."
                  required
                />
              </div>
              <div class="mb-3">
                <label class="form-label">Content</label>
                <textarea
                  class="form-control font-monospace"
                  v-model="textInputRequest.content"
                  rows="15"
                  placeholder="Paste or type your text content here..."
                  required
                  style="resize: vertical;"
                ></textarea>
                <small class="form-text text-muted">
                  {{ textInputRequest.content.length }} characters
                </small>
              </div>
              <div class="row">
                <div class="col-md-6 mb-3">
                  <label class="form-label">Chunk Size</label>
                  <input type="number" class="form-control" v-model.number="textInputRequest.chunk_size" />
                </div>
                <div class="col-md-6 mb-3">
                  <label class="form-label">Chunk Overlap</label>
                  <input type="number" class="form-control" v-model.number="textInputRequest.chunk_overlap" />
                </div>
              </div>
              <div class="d-grid">
                <button
                  type="submit"
                  class="btn btn-primary"
                  :disabled="documentStore.loading || !textInputRequest.content.trim() || !isProjectValid"
                >
                  <span v-if="documentStore.loading" class="spinner-border spinner-border-sm me-2"></span>
                  Create Document
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <!-- Signup Prompt Modal -->
    <SignupPromptModal :show="showSignupModal" @close="showSignupModal = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useDocumentStore } from '@/stores/documentStore';
import { useProjectStore } from '@/stores/projectStore';
import { useAuthStore } from '@/stores/authStore';
import type { Document } from '@/types';
import SignupPromptModal from '@/components/SignupPromptModal.vue';

const documentStore = useDocumentStore();
const projectStore = useProjectStore();
const authStore = useAuthStore();

// State
const searchQuery = ref('');
const sortBy = ref<'name' | 'created_at' | 'updated_at'>('created_at');
const sortOrder = ref<'asc' | 'desc'>('desc');
const selectedProjectId = ref('');
const currentPage = ref(1);
const itemsPerPage = ref(10);

// Modals
const showURLLoaderModal = ref(false);
const showFileUploaderModal = ref(false);
const showTextInputModal = ref(false);
const showChunksModal = ref(false);
const showSignupModal = ref(false);

// URL Load Request
const urlLoadRequest = ref({
  url: '',
  project_id: '',
  chunk_size: 1000,
  chunk_overlap: 200,
});

// File Upload Request
const uploadRequest = ref({
  project_id: '',
  chunk_size: 1000,
  chunk_overlap: 200,
});

// Text Input Request
const textInputRequest = ref({
  name: '',
  content: '',
  project_id: '',
  project_name: '',
  chunk_size: 1000,
  chunk_overlap: 200,
});

// Project mode for text input
const projectMode = ref<'existing' | 'new'>('existing');

const selectedFiles = ref<File[]>([]);
const dragOver = ref(false);
const viewingChunks = ref<Document[]>([]);
const fileInput = ref<HTMLInputElement>();

// Computed
const projects = computed(() => projectStore.projects);

const groupedDocuments = computed(() => {
  const docs = documentStore.searchAndSortedDocuments;
  const grouped: Record<string, Document[]> = {};

  docs.forEach((doc) => {
    if (!grouped[doc.path]) {
      grouped[doc.path] = [];
    }
    grouped[doc.path].push(doc);
  });

  return grouped;
});

const totalPages = computed(() => {
  const uniquePaths = Object.keys(groupedDocuments.value).length;
  return Math.ceil(uniquePaths / itemsPerPage.value);
});

const isProjectValid = computed(() => {
  if (projectMode.value === 'existing') {
    return textInputRequest.value.project_id !== '';
  } else {
    return textInputRequest.value.project_name.trim() !== '';
  }
});

// Methods
const handleSearch = () => {
  documentStore.setSearchQuery(searchQuery.value);
};

const clearSearch = () => {
  searchQuery.value = '';
  documentStore.setSearchQuery('');
};

const handleSortChange = () => {
  documentStore.setSorting(sortBy.value, sortOrder.value);
  fetchDocuments();
};

const handleProjectChange = () => {
  fetchDocuments();
};

const changePage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
  }
};

const fetchDocuments = async () => {
  const params: any = { limit: 100, offset: 0 };
  if (selectedProjectId.value) {
    params.project_id = selectedProjectId.value;
  }
  await documentStore.fetchDocuments(params);
  await documentStore.fetchStats(selectedProjectId.value || undefined);
};

const handleURLLoad = async () => {
  try {
    await documentStore.loadFromURL(urlLoadRequest.value);
    showURLLoaderModal.value = false;
    urlLoadRequest.value = { url: '', project_id: '', chunk_size: 1000, chunk_overlap: 200 };
  } catch (error) {
    console.error('Failed to load from URL:', error);
  }
};

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files) {
    selectedFiles.value.push(...Array.from(target.files));
  }
};

const handleDrop = (event: DragEvent) => {
  dragOver.value = false;
  if (event.dataTransfer?.files) {
    selectedFiles.value.push(...Array.from(event.dataTransfer.files));
  }
};

const removeFile = (index: number) => {
  selectedFiles.value.splice(index, 1);
};

const handleFileUpload = async () => {
  if (selectedFiles.value.length === 0) return;

  try {
    await documentStore.uploadFiles(
      selectedFiles.value,
      uploadRequest.value.project_id,
      uploadRequest.value.chunk_size,
      uploadRequest.value.chunk_overlap
    );
    showFileUploaderModal.value = false;
    selectedFiles.value = [];
    uploadRequest.value = { project_id: '', chunk_size: 1000, chunk_overlap: 200 };
  } catch (error) {
    console.error('Failed to upload files:', error);
  }
};

const handleTextCreate = async () => {
  if (!textInputRequest.value.content.trim()) return;

  try {
    const request: any = {
      name: textInputRequest.value.name,
      content: textInputRequest.value.content,
      chunk_size: textInputRequest.value.chunk_size,
      chunk_overlap: textInputRequest.value.chunk_overlap,
    };

    // Add either project_id or project_name based on mode
    if (projectMode.value === 'existing') {
      request.project_id = textInputRequest.value.project_id;
    } else {
      request.project_name = textInputRequest.value.project_name;
    }

    await documentStore.createFromText(request);

    // If a new project was created, refresh the projects list
    if (projectMode.value === 'new') {
      await projectStore.fetchProjects();
    }

    showTextInputModal.value = false;
    textInputRequest.value = {
      name: '',
      content: '',
      project_id: '',
      project_name: '',
      chunk_size: 1000,
      chunk_overlap: 200
    };
    projectMode.value = 'existing';
  } catch (error) {
    console.error('Failed to create document from text:', error);
  }
};

const viewChunks = (chunks: Document[]) => {
  viewingChunks.value = chunks;
  showChunksModal.value = true;
};

const deleteDocument = async (projectId: string, path: string) => {
  if (!confirm(`Are you sure you want to delete all chunks for this document?`)) return;

  try {
    await documentStore.deleteDocumentByPath(projectId, path);
    await fetchDocuments();
  } catch (error) {
    console.error('Failed to delete document:', error);
  }
};

const getDocumentName = (doc: Document) => {
  return doc.name || doc.path.split('/').pop() || 'Unnamed';
};

const getProjectName = (projectId: string) => {
  const project = projects.value.find((p) => p.id === projectId);
  return project?.name || projectId;
};

const formatDate = (dateStr?: string) => {
  if (!dateStr) return 'N/A';
  return new Date(dateStr).toLocaleString();
};

const formatFileSize = (bytes: number) => {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
};

// Lifecycle
onMounted(async () => {
  await projectStore.fetchProjects();
  await fetchDocuments();
});
</script>

<style scoped>
.dropzone {
  border: 2px dashed #dee2e6;
  border-radius: 0.375rem;
  padding: 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.dropzone:hover,
.dropzone.drag-over {
  border-color: #0d6efd;
  background-color: #f8f9fa;
}

.dropzone-content {
  cursor: pointer;
}

.selected-files {
  text-align: left;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 0.5rem;
  border-bottom: 1px solid #dee2e6;
}

.file-item:last-child {
  border-bottom: none;
}

.documents-list .card {
  transition: box-shadow 0.3s ease;
}

.documents-list .card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}
</style>
