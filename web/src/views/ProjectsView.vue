<template>
  <div class="container mt-4 mb-5">
    <LoadingSpinner :show="isInitialLoading" message="Loading projects..." />
    <div v-if="!isInitialLoading">
      <!-- Header with Add Button -->
      <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-project-diagram"></i> Project Management
      </h2>
      <div class="btn-group">
        <button
          @click="exportProjects"
          class="btn btn-outline-success"
          title="Export projects to YAML"
        >
          <i class="fas fa-file-export"></i> Export
        </button>
        <button
          v-if="authStore.canModify"
          @click="showImportModal"
          class="btn btn-outline-info"
          title="Import projects from YAML"
        >
          <i class="fas fa-file-import"></i> Import
        </button>
        <button
          v-if="authStore.canModify"
          @click="showAddProjectModal"
          class="btn btn-primary"
        >
          <i class="fas fa-plus"></i> Add Project
        </button>
        <button
          v-else
          @click="showSignupModal = true"
          class="btn btn-primary"
          title="Sign up to create and manage projects"
        >
          <i class="fas fa-lock"></i> Sign Up to Add Projects
        </button>
      </div>
    </div>

    <!-- Search, Filter, and View Controls -->
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
                placeholder="Search projects by name, description, language..."
                v-model="searchQuery"
                @input="handleSearch"
              />
              <button
                v-if="searchQuery"
                class="btn btn-outline-secondary"
                type="button"
                @click="clearSearch"
              >
                <i class="fas fa-times"></i>
              </button>
            </div>
          </div>

          <!-- Sort By -->
          <div class="col-md-3">
            <select class="form-select" v-model="sortBy" @change="handleSortChange">
              <option value="name">Sort by Name</option>
              <option value="description">Sort by Description</option>
              <option value="language">Sort by Language</option>
              <option value="created_time">Sort by Created Time</option>
            </select>
          </div>

          <!-- View Toggle -->
          <div class="col-md-3">
            <div class="btn-group w-100" role="group">
              <button
                type="button"
                class="btn"
                :class="viewMode === 'card' ? 'btn-primary' : 'btn-outline-primary'"
                @click="viewMode = 'card'"
              >
                <i class="fas fa-th"></i> Card
              </button>
              <button
                type="button"
                class="btn"
                :class="viewMode === 'list' ? 'btn-primary' : 'btn-outline-primary'"
                @click="viewMode = 'list'"
              >
                <i class="fas fa-list"></i> List
              </button>
              <button
                type="button"
                class="btn btn-outline-secondary"
                @click="toggleSortOrder"
                :title="sortOrder === 'asc' ? 'Ascending' : 'Descending'"
              >
                <i class="fas" :class="sortOrder === 'asc' ? 'fa-sort-alpha-down' : 'fa-sort-alpha-up'"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="projectStore.loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="paginatedProjects.length === 0 && searchQuery" class="alert alert-warning">
      <i class="fas fa-search"></i> No projects found matching "{{ searchQuery }}".
      <button class="btn btn-link p-0" @click="clearSearch">Clear search</button>
    </div>

    <div v-else-if="projectStore.projects.length === 0" class="alert alert-info">
      <i class="fas fa-info-circle"></i> No projects found. Add your first project to get started!
    </div>

    <!-- Card View -->
    <div v-else-if="viewMode === 'card'" class="row g-3 mb-4">
      <div
        v-for="project in paginatedProjects"
        :key="project.id"
        class="col-md-6 col-lg-4"
      >
        <div class="card h-100 project-card">
          <div class="card-header bg-success text-white">
            <div class="d-flex justify-content-between align-items-center">
              <h5 class="mb-0">
                {{ project.name }}
                <i v-if="projectStore.isDefaultProject(project.id)" class="fas fa-star text-warning ms-1" title="Default Project"></i>
              </h5>
              <span v-if="isTemplate(project)" class="badge bg-light text-dark" title="Global template visible to all users">
                <i class="fas fa-globe"></i> Template
              </span>
              <span v-else-if="isPersonal(project)" class="badge bg-warning text-dark" title="Your personal project">
                <i class="fas fa-user"></i> Personal
              </span>
              <span v-else-if="isShared(project)" class="badge bg-info" title="Shared in your realm">
                <i class="fas fa-users"></i> Shared
              </span>
            </div>
          </div>
          <div class="card-body">
            <p class="card-text">{{ project.description || 'No description' }}</p>
            <div class="project-details">
              <div class="mb-2" v-if="project.language">
                <strong><i class="fas fa-code"></i> Language:</strong>
                <span class="badge bg-info ms-2">{{ project.language }}</span>
              </div>
              <div class="mb-2" v-if="project.git_repo">
                <strong><i class="fas fa-code-branch"></i> Repository:</strong>
                <code class="ms-2">{{ project.git_repo }}</code>
              </div>
              <div class="mb-2" v-if="project.git_branch">
                <strong><i class="fas fa-code-branch"></i> Branch:</strong>
                <code class="ms-2">{{ project.git_branch }}</code>
              </div>
              <div class="mb-2" v-if="project.entry_point">
                <strong><i class="fas fa-folder"></i> Entry Point:</strong>
                <code class="ms-2">{{ project.entry_point }}</code>
              </div>
            </div>
          </div>
          <div class="card-footer">
            <button
              @click="selectProject(project)"
              class="btn btn-sm me-2"
              :class="projectStore.selectedProject?.id === project.id ? 'btn-success' : 'btn-outline-success'"
            >
              <i class="fas fa-check-circle"></i>
              {{ projectStore.selectedProject?.id === project.id ? 'Selected' : 'Select' }}
            </button>
            <button
              v-if="!projectStore.isDefaultProject(project.id)"
              @click="setAsDefault(project)"
              class="btn btn-sm btn-outline-warning me-2"
              title="Set as Default Project"
            >
              <i class="fas fa-star"></i>
            </button>
            <button
              v-else
              @click="clearDefault"
              class="btn btn-sm btn-warning me-2"
              title="Clear Default"
            >
              <i class="fas fa-star"></i>
            </button>
            <button
              v-if="canModifyProject(project)"
              @click="editProject(project)"
              class="btn btn-sm btn-outline-primary me-2"
            >
              <i class="fas fa-edit"></i>
            </button>
            <button
              v-else-if="!authStore.canModify"
              @click="handleCUDAction(() => {})"
              class="btn btn-sm btn-outline-secondary me-2"
              :disabled="true"
              title="Sign up to edit projects"
            >
              <i class="fas fa-lock"></i>
            </button>
            <button
              v-if="canModifyProject(project)"
              @click="deleteProject(project.id)"
              class="btn btn-sm btn-outline-danger"
            >
              <i class="fas fa-trash"></i>
            </button>
            <button
              v-else-if="!authStore.canModify"
              @click="handleCUDAction(() => {})"
              class="btn btn-sm btn-outline-secondary"
              :disabled="true"
              title="Sign up to delete projects"
            >
              <i class="fas fa-lock"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-else class="mb-4">
      <div class="table-responsive">
        <table class="table table-hover">
          <thead class="table-light">
            <tr>
              <th style="width: 18%">Name</th>
              <th style="width: 10%">Type</th>
              <th style="width: 20%">Description</th>
              <th style="width: 10%">Language</th>
              <th style="width: 17%">Repository</th>
              <th style="width: 10%">Branch</th>
              <th style="width: 15%">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="project in paginatedProjects" :key="project.id"
                :class="{ 'table-success': projectStore.selectedProject?.id === project.id }">
              <td class="fw-bold">
                {{ project.name }}
                <i v-if="projectStore.isDefaultProject(project.id)" class="fas fa-star text-warning ms-1" title="Default Project"></i>
              </td>
              <td>
                <span v-if="isTemplate(project)" class="badge bg-secondary" title="Global template">
                  <i class="fas fa-globe"></i> Template
                </span>
                <span v-else-if="isPersonal(project)" class="badge bg-warning text-dark" title="Personal">
                  <i class="fas fa-user"></i> Personal
                </span>
                <span v-else-if="isShared(project)" class="badge bg-info" title="Shared">
                  <i class="fas fa-users"></i> Shared
                </span>
              </td>
              <td>{{ project.description || 'No description' }}</td>
              <td>
                <span v-if="project.language" class="badge bg-info">{{ project.language }}</span>
              </td>
              <td>
                <div class="text-truncate" style="max-width: 200px" :title="project.git_repo">
                  {{ project.git_repo || 'N/A' }}
                </div>
              </td>
              <td>
                <code v-if="project.git_branch" class="small">{{ project.git_branch }}</code>
              </td>
              <td>
                <button
                  @click="selectProject(project)"
                  class="btn btn-sm me-1"
                  :class="projectStore.selectedProject?.id === project.id ? 'btn-success' : 'btn-outline-success'"
                  title="Select"
                >
                  <i class="fas fa-check-circle"></i>
                </button>
                <button
                  v-if="!projectStore.isDefaultProject(project.id)"
                  @click="setAsDefault(project)"
                  class="btn btn-sm btn-outline-warning me-1"
                  title="Set as Default"
                >
                  <i class="fas fa-star"></i>
                </button>
                <button
                  v-else
                  @click="clearDefault"
                  class="btn btn-sm btn-warning me-1"
                  title="Clear Default"
                >
                  <i class="fas fa-star"></i>
                </button>
                <button
                  v-if="canModifyProject(project)"
                  @click="editProject(project)"
                  class="btn btn-sm btn-outline-primary me-1"
                  title="Edit"
                >
                  <i class="fas fa-edit"></i>
                </button>
                <button
                  v-else-if="!authStore.canModify"
                  @click="handleCUDAction(() => {})"
                  class="btn btn-sm btn-outline-secondary me-1"
                  :disabled="true"
                  title="Sign up to edit"
                >
                  <i class="fas fa-lock"></i>
                </button>
                <button
                  @click="viewProjectDetails(project)"
                  class="btn btn-sm btn-outline-info me-1"
                  title="View Details"
                >
                  <i class="fas fa-eye"></i>
                </button>
                <button
                  v-if="canModifyProject(project)"
                  @click="deleteProject(project.id)"
                  class="btn btn-sm btn-outline-danger"
                  title="Delete"
                >
                  <i class="fas fa-trash"></i>
                </button>
                <button
                  v-else-if="!authStore.canModify"
                  @click="handleCUDAction(() => {})"
                  class="btn btn-sm btn-outline-secondary"
                  :disabled="true"
                  title="Sign up to delete"
                >
                  <i class="fas fa-lock"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <nav v-if="totalPages > 1" aria-label="Project pagination">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: currentPage === 1 }">
          <a class="page-link" href="#" @click.prevent="goToPage(currentPage - 1)">
            <i class="fas fa-chevron-left"></i> Previous
          </a>
        </li>

        <li
          v-for="page in visiblePages"
          :key="page"
          class="page-item"
          :class="{ active: page === currentPage }"
        >
          <a class="page-link" href="#" @click.prevent="goToPage(page)">{{ page }}</a>
        </li>

        <li class="page-item" :class="{ disabled: currentPage === totalPages }">
          <a class="page-link" href="#" @click.prevent="goToPage(currentPage + 1)">
            Next <i class="fas fa-chevron-right"></i>
          </a>
        </li>
      </ul>

      <div class="text-center text-muted small">
        Showing {{ startIndex + 1 }}-{{ Math.min(endIndex, filteredProjects.length) }} of {{ filteredProjects.length }} projects
        ({{ itemsPerPage }} per page)
      </div>
    </nav>

    <!-- Add/Edit Project Modal -->
    <div
      class="modal fade"
      id="projectModal"
      tabindex="-1"
      aria-labelledby="projectModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="projectModalLabel">
              {{ isEditMode ? 'Edit Project' : 'Add New Project' }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveProject">
              <div class="mb-3">
                <label for="projectName" class="form-label">Name *</label>
                <input
                  type="text"
                  class="form-control"
                  id="projectName"
                  v-model="editingProject.name"
                  required
                  placeholder="e.g., my-awesome-project"
                />
              </div>

              <div class="mb-3">
                <label for="projectDescription" class="form-label">Description</label>
                <textarea
                  class="form-control"
                  id="projectDescription"
                  v-model="editingProject.description"
                  rows="2"
                  placeholder="Brief description of this project"
                ></textarea>
              </div>

              <div class="mb-3">
                <label for="projectGitUrl" class="form-label">Git URL</label>
                <input
                  type="text"
                  class="form-control"
                  id="projectGitUrl"
                  v-model="editingProject.git_url"
                  placeholder="https://github.com/user/repo.git"
                />
              </div>

              <div class="row">
                <div class="col-md-8 mb-3">
                  <label for="projectGitRepo" class="form-label">Git Repository Path</label>
                  <input
                    type="text"
                    class="form-control"
                    id="projectGitRepo"
                    v-model="editingProject.git_repo"
                    placeholder="e.g., user/repo"
                  />
                </div>

                <div class="col-md-4 mb-3">
                  <label for="projectGitBranch" class="form-label">Branch</label>
                  <input
                    type="text"
                    class="form-control"
                    id="projectGitBranch"
                    v-model="editingProject.git_branch"
                    placeholder="e.g., main"
                  />
                </div>
              </div>

              <div class="row">
                <div class="col-md-6 mb-3">
                  <label for="projectLanguage" class="form-label">Language</label>
                  <select class="form-select" id="projectLanguage" v-model="editingProject.language">
                    <option value="">Select Language</option>
                    <option value="JavaScript">JavaScript</option>
                    <option value="TypeScript">TypeScript</option>
                    <option value="Python">Python</option>
                    <option value="Java">Java</option>
                    <option value="Go">Go</option>
                    <option value="Rust">Rust</option>
                    <option value="C++">C++</option>
                    <option value="C#">C#</option>
                    <option value="Ruby">Ruby</option>
                    <option value="PHP">PHP</option>
                  </select>
                </div>

                <div class="col-md-6 mb-3">
                  <label for="projectEntryPoint" class="form-label">Entry Point</label>
                  <input
                    type="text"
                    class="form-control"
                    id="projectEntryPoint"
                    v-model="editingProject.entry_point"
                    placeholder="e.g., src/main.js"
                  />
                </div>
              </div>

              <div class="mb-3">
                <label for="projectScope" class="form-label">Scope</label>
                <select class="form-select" id="projectScope" v-model="editingProject.scope">
                  <option value="personal">Personal (only you can see)</option>
                  <option value="shared">Shared (visible to all in realm)</option>
                </select>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="saveProject"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save me-1"></i>
              {{ isEditMode ? 'Update' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- View Details Modal -->
    <div
      class="modal fade"
      id="viewDetailsModal"
      tabindex="-1"
      aria-labelledby="viewDetailsModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header bg-success text-white">
            <h5 class="modal-title" id="viewDetailsModalLabel">
              <i class="fas fa-eye"></i> {{ viewingProject?.name }}
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="viewingProject">
              <div class="mb-3">
                <h6 class="fw-bold">Description:</h6>
                <p>{{ viewingProject.description || 'No description' }}</p>
              </div>

              <div class="mb-3" v-if="viewingProject.language">
                <h6 class="fw-bold">Language:</h6>
                <span class="badge bg-info">{{ viewingProject.language }}</span>
              </div>

              <div class="mb-3" v-if="viewingProject.git_url">
                <h6 class="fw-bold">Git URL:</h6>
                <code>{{ viewingProject.git_url }}</code>
              </div>

              <div class="mb-3" v-if="viewingProject.git_repo">
                <h6 class="fw-bold">Repository:</h6>
                <code>{{ viewingProject.git_repo }}</code>
              </div>

              <div class="mb-3" v-if="viewingProject.git_branch">
                <h6 class="fw-bold">Branch:</h6>
                <code>{{ viewingProject.git_branch }}</code>
              </div>

              <div class="mb-3" v-if="viewingProject.entry_point">
                <h6 class="fw-bold">Entry Point:</h6>
                <code>{{ viewingProject.entry_point }}</code>
              </div>

              <div class="mb-3" v-if="viewingProject.created_time">
                <h6 class="fw-bold">Created:</h6>
                <span>{{ formatDate(viewingProject.created_time) }}</span>
              </div>

              <div class="mb-3" v-if="viewingProject.updated_time">
                <h6 class="fw-bold">Last Updated:</h6>
                <span>{{ formatDate(viewingProject.updated_time) }}</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Close
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="editProjectFromView"
            >
              <i class="fas fa-edit"></i> Edit
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Import Projects Modal -->
    <div
      class="modal fade"
      id="importModal"
      tabindex="-1"
      aria-labelledby="importModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="importModalLabel">
              <i class="fas fa-file-import"></i> Import Projects from YAML
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label for="importFile" class="form-label">Select YAML File</label>
              <input
                type="file"
                class="form-control"
                id="importFile"
                accept=".yaml,.yml"
                @change="handleFileSelect"
              />
              <div class="form-text">
                Upload a YAML file containing projects to import
              </div>
            </div>

            <div class="mb-3">
              <div class="form-check">
                <input
                  class="form-check-input"
                  type="checkbox"
                  id="updateExisting"
                  v-model="importOptions.updateExisting"
                />
                <label class="form-check-label" for="updateExisting">
                  Update existing projects (if unchecked, existing projects will be skipped)
                </label>
              </div>
            </div>

            <div class="mb-3">
              <label for="importScope" class="form-label">Import as</label>
              <select class="form-select" id="importScope" v-model="importOptions.scope">
                <option value="personal">Personal (only you can see)</option>
                <option value="shared">Shared (visible to all in realm)</option>
              </select>
            </div>

            <div v-if="importPreview.length > 0" class="mb-3">
              <h6>Preview ({{ importPreview.length }} projects):</h6>
              <div class="table-responsive" style="max-height: 300px; overflow-y: auto;">
                <table class="table table-sm">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Language</th>
                      <th>Repository</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(project, index) in importPreview" :key="index">
                      <td>{{ project.name }}</td>
                      <td><span v-if="project.language" class="badge bg-info">{{ project.language }}</span></td>
                      <td><code class="small">{{ project.project || 'N/A' }}</code></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="importProjectsFromFile"
              :disabled="importPreview.length === 0 || isImporting"
            >
              <span v-if="isImporting" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-upload me-1"></i>
              Import {{ importPreview.length }} Projects
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Signup Prompt Modal -->
    <SignupPromptModal :show="showSignupModal" @close="showSignupModal = false" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useProjectStore } from '@/stores/projectStore';
import { useAuthStore } from '@/stores/authStore';
import { showToast } from '@/utils/toast';
import type { Project } from '@/types';
import { Modal } from 'bootstrap';
import SignupPromptModal from '@/components/SignupPromptModal.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';
import { useInitialLoad } from '@/composables/useInitialLoad';

const projectStore = useProjectStore();
const authStore = useAuthStore();
const { isInitialLoading, runInitialLoad } = useInitialLoad();
const showSignupModal = ref(false);
const isEditMode = ref(false);
const isSaving = ref(false);
const viewMode = ref<'card' | 'list'>('list');
const searchQuery = ref('');
const sortBy = ref<'name' | 'description' | 'language' | 'created_time'>('name');
const sortOrder = ref<'asc' | 'desc'>('asc');

// Pagination
const currentPage = ref(1);
const itemsPerPage = ref(12);

const editingProject = ref<Partial<Project & { scope: string }>>({
  name: '',
  description: '',
  git_url: '',
  git_repo: '',
  git_branch: 'main',
  language: '',
  entry_point: '',
  scope: 'personal',
});

const viewingProject = ref<Project | null>(null);

// Import/Export state
const isImporting = ref(false);
const importPreview = ref<any[]>([]);
const importOptions = ref({
  updateExisting: false,
  scope: 'personal',
});

let projectModal: Modal | null = null;
let viewDetailsModal: Modal | null = null;
let importModal: Modal | null = null;

// Computed properties
const filteredProjects = computed(() => {
  return projectStore.searchAndSortedProjects;
});

const totalPages = computed(() => {
  return Math.ceil(filteredProjects.value.length / itemsPerPage.value);
});

const startIndex = computed(() => {
  return (currentPage.value - 1) * itemsPerPage.value;
});

const endIndex = computed(() => {
  return startIndex.value + itemsPerPage.value;
});

const paginatedProjects = computed(() => {
  return filteredProjects.value.slice(startIndex.value, endIndex.value);
});

const visiblePages = computed(() => {
  const pages: number[] = [];
  const maxVisible = 5;
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages.value, start + maxVisible - 1);

  if (end - start < maxVisible - 1) {
    start = Math.max(1, end - maxVisible + 1);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
});

// Helper functions for project ownership/visibility
function isTemplate(project: Project): boolean {
  // Global templates have no user_id and no realm_id (or realm_id is empty)
  return !project.user_id && (!project.realm_id || project.realm_id === '');
}

function isPersonal(project: Project): boolean {
  // Personal projects have user_id set
  return !!project.user_id;
}

function isShared(project: Project): boolean {
  // Shared projects have realm_id but no user_id
  return !project.user_id && !!project.realm_id && project.realm_id !== '';
}

function canModifyProject(project: Project): boolean {
  // Can modify if authenticated and (it's your project or you created it)
  if (!authStore.canModify) return false;
  // Super admins can modify anything (including templates)
  if (authStore.isSuperAdmin) return true;
  // Templates cannot be modified by non-super-admins
  if (isTemplate(project)) return false;
  // Personal projects can only be modified by owner
  if (isPersonal(project)) {
    return project.user_id === authStore.user?.id;
  }
  // Shared projects in your realm can be modified
  if (isShared(project)) {
    return project.realm_id === authStore.user?.realm_id;
  }
  return false;
}

function handleCUDAction(action: () => void) {
  if (!authStore.canModify) {
    showSignupModal.value = true;
    return;
  }
  action();
}

// Methods
function handleSearch() {
  projectStore.setSearchQuery(searchQuery.value);
  currentPage.value = 1;
}

function clearSearch() {
  searchQuery.value = '';
  projectStore.setSearchQuery('');
  currentPage.value = 1;
}

function handleSortChange() {
  projectStore.setSorting(sortBy.value, sortOrder.value);
}

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc';
  projectStore.setSorting(sortBy.value, sortOrder.value);
}

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
}

function showAddProjectModal() {
  isEditMode.value = false;
  editingProject.value = {
    name: '',
    description: '',
    git_url: '',
    git_repo: '',
    git_branch: 'main',
    language: '',
    entry_point: '',
    scope: 'personal',
  };

  const modalEl = document.getElementById('projectModal');
  if (modalEl) {
    projectModal = new Modal(modalEl);
    projectModal.show();
  }
}

function editProject(project: Project) {
  isEditMode.value = true;
  editingProject.value = {
    id: project.id,
    name: project.name,
    description: project.description,
    git_url: project.git_url,
    git_repo: project.git_repo,
    git_branch: project.git_branch,
    language: project.language,
    entry_point: project.entry_point,
    scope: 'personal',
  };

  const modalEl = document.getElementById('projectModal');
  if (modalEl) {
    projectModal = new Modal(modalEl);
    projectModal.show();
  }
}

function viewProjectDetails(project: Project) {
  viewingProject.value = project;

  const modalEl = document.getElementById('viewDetailsModal');
  if (modalEl) {
    viewDetailsModal = new Modal(modalEl);
    viewDetailsModal.show();
  }
}

function editProjectFromView() {
  if (viewDetailsModal) {
    viewDetailsModal.hide();
  }
  if (viewingProject.value) {
    editProject(viewingProject.value);
  }
}

async function saveProject() {
  if (!editingProject.value.name) {
    showToast('Project name is required', 'danger');
    return;
  }

  try {
    isSaving.value = true;

    const projectData: Partial<Project> = {
      name: editingProject.value.name,
      description: editingProject.value.description || '',
      git_url: editingProject.value.git_url || '',
      git_repo: editingProject.value.git_repo || '',
      git_branch: editingProject.value.git_branch || 'main',
      language: editingProject.value.language || '',
      entry_point: editingProject.value.entry_point || '',
    };

    if (isEditMode.value && editingProject.value.id) {
      await projectStore.updateProject(editingProject.value.id, projectData);
    } else {
      await projectStore.createProject(projectData);
    }

    // Close modal
    if (projectModal) {
      projectModal.hide();
    }

    // Refresh list
    await projectStore.fetchProjects();
  } catch (error) {
    console.error('Failed to save project:', error);
  } finally {
    isSaving.value = false;
  }
}

async function deleteProject(id: string) {
  const project = projectStore.getProjectById(id);
  if (confirm(`Are you sure you want to delete the project "${project?.name}"?`)) {
    try {
      await projectStore.deleteProject(id);
    } catch (error) {
      console.error('Failed to delete project:', error);
    }
  }
}

function selectProject(project: Project) {
  projectStore.selectProject(project);
  showToast(`Selected project: ${project.name}`, 'success');
}

function setAsDefault(project: Project) {
  projectStore.setDefaultProject(project.id);
}

function clearDefault() {
  projectStore.clearDefaultProject();
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString();
}

// Import/Export functions
async function exportProjects() {
  try {
    const scope = 'all'; // Export all accessible projects
    const blob = await projectStore.exportProjects(scope);

    // Create download link
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    const date = new Date().toISOString().split('T')[0];
    link.download = `projects_${date}.yaml`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);

    showToast('Projects exported successfully', 'success');
  } catch (error) {
    console.error('Failed to export projects:', error);
    showToast('Failed to export projects', 'danger');
  }
}

function showImportModal() {
  importPreview.value = [];
  importOptions.value = {
    updateExisting: false,
    scope: 'personal',
  };

  const modalEl = document.getElementById('importModal');
  if (modalEl) {
    importModal = new Modal(modalEl);
    importModal.show();
  }
}

async function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];

  if (!file) {
    importPreview.value = [];
    return;
  }

  try {
    const text = await file.text();
    // Parse YAML - simple parsing for projects structure
    const data = parseYAML(text);

    if (data.projects) {
      // Convert from config.yaml format to array
      const projectsArray = Object.entries(data.projects).map(([name, config]: [string, any]) => ({
        name,
        description: config.description || `Imported from ${file.name}`,
        gitUrl: config.gitUrl || '',
        project: config.project || '',
        branch: config.branch || 'main',
        codePath: config.codePath || '',
        language: config.language || '',
      }));
      importPreview.value = projectsArray;
    } else {
      showToast('Invalid YAML format. Expected "projects" key.', 'danger');
      importPreview.value = [];
    }
  } catch (error) {
    console.error('Failed to parse file:', error);
    showToast('Failed to parse YAML file', 'danger');
    importPreview.value = [];
  }
}

// Simple YAML parser for our specific format
function parseYAML(text: string): any {
  const result: any = { projects: {} };
  const lines = text.split('\n');
  let currentProject: string | null = null;
  let inProjects = false;

  for (const line of lines) {
    const trimmed = line.trim();

    if (trimmed.startsWith('projects:')) {
      inProjects = true;
      continue;
    }

    if (!inProjects || trimmed === '' || trimmed.startsWith('#')) {
      continue;
    }

    // Check if this is a project name (not indented much)
    if (line.match(/^  \w+:/)) {
      currentProject = trimmed.replace(':', '');
      result.projects[currentProject] = {};
      continue;
    }

    // Parse project properties
    if (currentProject && line.match(/^    \w+:/)) {
      const match = trimmed.match(/(\w+):\s*(.+)/);
      if (match) {
        const [, key, value] = match;
        result.projects[currentProject][key] = value;
      }
    }
  }

  return result;
}

async function importProjectsFromFile() {
  if (importPreview.value.length === 0) {
    showToast('No projects to import', 'warning');
    return;
  }

  try {
    isImporting.value = true;

    const result = await projectStore.importProjects(
      importPreview.value,
      importOptions.value.updateExisting,
      importOptions.value.scope
    );

    // Close modal
    if (importModal) {
      importModal.hide();
    }

    // Clear file input
    const fileInput = document.getElementById('importFile') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }

    // Show success message
    showToast(
      `Import completed: ${result.created} created, ${result.updated} updated, ${result.skipped} skipped`,
      'success'
    );

    // Refresh projects list
    await projectStore.fetchProjects();
  } catch (error) {
    console.error('Failed to import projects:', error);
    showToast('Failed to import projects', 'danger');
  } finally {
    isImporting.value = false;
  }
}

onMounted(() => {
  runInitialLoad(async () => {
    await projectStore.fetchProjects();

    // Load preferences from localStorage
    const savedViewMode = localStorage.getItem('projectViewMode');
    if (savedViewMode === 'list' || savedViewMode === 'card') {
      viewMode.value = savedViewMode;
    }

    const savedItemsPerPage = localStorage.getItem('projectItemsPerPage');
    if (savedItemsPerPage) {
      itemsPerPage.value = parseInt(savedItemsPerPage, 10);
    }
  });
});

// Watch for view mode changes to save preference
watch(viewMode, (newMode) => {
  localStorage.setItem('projectViewMode', newMode);
});
</script>

<style scoped>
.project-card {
  transition: transform 0.2s, box-shadow 0.2s;
}

.project-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.project-details code {
  background-color: #f8f9fa;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

.table-responsive {
  border-radius: 0.375rem;
  overflow: hidden;
}

.table tbody tr {
  transition: background-color 0.2s;
}

.table tbody tr:hover {
  background-color: rgba(25, 135, 84, 0.05);
}

.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination {
  margin-top: 1rem;
}

.badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
}

.card-header {
  padding: 0.75rem 1rem;
}

.card-body {
  padding: 1rem;
}

.card-footer {
  padding: 0.75rem 1rem;
  background-color: rgba(0, 0, 0, 0.03);
}

.input-group-text {
  background-color: #f8f9fa;
}

.modal-dialog {
  max-width: 800px;
}

.btn-group .btn {
  border-radius: 0;
}

.btn-group .btn:first-child {
  border-top-left-radius: 0.375rem;
  border-bottom-left-radius: 0.375rem;
}

.btn-group .btn:last-child {
  border-top-right-radius: 0.375rem;
  border-bottom-right-radius: 0.375rem;
}
</style>
