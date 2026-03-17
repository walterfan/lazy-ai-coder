<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-robot"></i> LLM Model Management
      </h2>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <i class="fas fa-plus"></i> Add Model
      </button>
    </div>

    <div class="alert alert-info mb-4">
      <i class="fas fa-info-circle"></i>
      <strong>Multi-Model Support:</strong> Configure multiple LLM providers and switch between them easily.
      Each model can have its own API key (stored locally in your browser for security).
      If no API key is set for a model, the default API key from Settings will be used.
    </div>

    <!-- Filters -->
    <div class="row mb-3">
      <div class="col-md-4">
        <div class="input-group">
          <span class="input-group-text"><i class="fas fa-search"></i></span>
          <input
            type="text"
            class="form-control"
            placeholder="Search models..."
            v-model="searchQuery"
            @input="debouncedSearch"
          />
        </div>
      </div>
      <div class="col-md-3">
        <select class="form-select" v-model="scope" @change="llmModelStore.setScope(scope)">
          <option value="all">All Models</option>
          <option value="personal">My Models</option>
          <option value="shared">Shared Models</option>
          <option value="templates">Template Models</option>
        </select>
      </div>
      <div class="col-md-3">
        <div class="form-check form-switch mt-2">
          <input
            class="form-check-input"
            type="checkbox"
            id="enabledOnly"
            v-model="enabledOnly"
            @change="llmModelStore.setEnabledOnly(enabledOnly)"
          />
          <label class="form-check-label" for="enabledOnly">Enabled Only</label>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="llmModelStore.loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="mt-2">Loading LLM models...</p>
    </div>

    <!-- Models Table -->
    <div v-else class="card">
      <div class="table-responsive">
        <table class="table table-hover mb-0">
          <thead class="table-light">
            <tr>
              <th style="width: 30px"></th>
              <th>Name</th>
              <th>Type</th>
              <th>Model</th>
              <th>Base URL</th>
              <th>Status</th>
              <th style="width: 180px">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="model in llmModelStore.models" :key="model.id" :class="{ 'table-primary': model.is_default }">
              <td>
                <i v-if="model.is_default" class="fas fa-star text-warning" title="Default Model"></i>
              </td>
              <td>
                <strong>{{ model.name }}</strong>
                <br />
                <small class="text-muted">{{ model.description }}</small>
              </td>
              <td>
                <span class="badge" :class="getLLMTypeBadgeClass(model.llm_type)">
                  {{ model.llm_type }}
                </span>
              </td>
              <td><code>{{ model.model }}</code></td>
              <td>
                <small class="text-truncate d-inline-block" style="max-width: 200px" :title="model.base_url">
                  {{ model.base_url }}
                </small>
              </td>
              <td>
                <span v-if="model.is_enabled" class="badge bg-success">Enabled</span>
                <span v-else class="badge bg-secondary">Disabled</span>
                <span v-if="llmModelStore.hasApiKey(model.id)" class="badge bg-info ms-1" title="Has custom API key">
                  <i class="fas fa-key"></i>
                </span>
              </td>
              <td>
                <div class="btn-group btn-group-sm">
                  <button
                    class="btn btn-outline-primary"
                    @click="editModel(model)"
                    title="Edit"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    v-if="!model.is_default"
                    class="btn btn-outline-warning"
                    @click="setAsDefault(model.id)"
                    title="Set as Default"
                  >
                    <i class="fas fa-star"></i>
                  </button>
                  <button
                    class="btn btn-outline-secondary"
                    @click="toggleEnabled(model)"
                    :title="model.is_enabled ? 'Disable' : 'Enable'"
                  >
                    <i :class="model.is_enabled ? 'fas fa-toggle-on' : 'fas fa-toggle-off'"></i>
                  </button>
                  <button
                    class="btn btn-outline-danger"
                    @click="confirmDelete(model)"
                    title="Delete"
                  >
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="llmModelStore.models.length === 0">
              <td colspan="7" class="text-center py-4 text-muted">
                <i class="fas fa-inbox fa-2x mb-2"></i>
                <p>No LLM models found. Click "Add Model" to create one.</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="llmModelStore.totalPages > 1" class="card-footer">
        <nav>
          <ul class="pagination pagination-sm mb-0 justify-content-center">
            <li class="page-item" :class="{ disabled: llmModelStore.pageNumber === 1 }">
              <a class="page-link" href="#" @click.prevent="llmModelStore.goToPage(llmModelStore.pageNumber - 1)">
                Previous
              </a>
            </li>
            <li
              v-for="page in displayedPages"
              :key="page"
              class="page-item"
              :class="{ active: page === llmModelStore.pageNumber }"
            >
              <a class="page-link" href="#" @click.prevent="llmModelStore.goToPage(page)">
                {{ page }}
              </a>
            </li>
            <li class="page-item" :class="{ disabled: llmModelStore.pageNumber === llmModelStore.totalPages }">
              <a class="page-link" href="#" @click.prevent="llmModelStore.goToPage(llmModelStore.pageNumber + 1)">
                Next
              </a>
            </li>
          </ul>
        </nav>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div class="modal fade" id="modelModal" tabindex="-1" ref="modalRef">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ isEditing ? 'Edit LLM Model' : 'Add LLM Model' }}</h5>
            <button type="button" class="btn-close" @click="closeModal"></button>
          </div>
          <form @submit.prevent="saveModel">
            <div class="modal-body">
              <div class="row g-3">
                <div class="col-md-6">
                  <label class="form-label">Name *</label>
                  <input type="text" class="form-control" v-model="formData.name" required />
                </div>
                <div class="col-md-6">
                  <label class="form-label">LLM Type *</label>
                  <select class="form-select" v-model="formData.llm_type" required>
                    <option value="openai">OpenAI</option>
                    <option value="anthropic">Anthropic</option>
                    <option value="google">Google</option>
                    <option value="alibaba">Alibaba (Qwen)</option>
                    <option value="deepseek">DeepSeek</option>
                    <option value="custom">Custom</option>
                  </select>
                </div>
                <div class="col-12">
                  <label class="form-label">Base URL *</label>
                  <input type="url" class="form-control" v-model="formData.base_url" required placeholder="https://api.openai.com/v1" />
                </div>
                <div class="col-md-6">
                  <label class="form-label">Model *</label>
                  <input type="text" class="form-control" v-model="formData.model" required placeholder="gpt-4" />
                </div>
                <div class="col-md-3">
                  <label class="form-label">Temperature</label>
                  <input type="number" class="form-control" v-model.number="formData.temperature" min="0" max="2" step="0.1" />
                </div>
                <div class="col-md-3">
                  <label class="form-label">Max Tokens</label>
                  <input type="number" class="form-control" v-model.number="formData.max_tokens" min="1" />
                </div>
                <div class="col-12">
                  <label class="form-label">Description</label>
                  <textarea class="form-control" v-model="formData.description" rows="2"></textarea>
                </div>
                <div class="col-12">
                  <label class="form-label">Extra Parameters (JSON)</label>
                  <textarea class="form-control font-monospace" v-model="formData.extra_params" rows="2" placeholder='{"top_p": 0.9}'></textarea>
                </div>
                <div class="col-12">
                  <label class="form-label">
                    API Key
                    <small class="text-muted">(optional, stored locally in browser)</small>
                  </label>
                  <div class="input-group">
                    <input
                      :type="showApiKey ? 'text' : 'password'"
                      class="form-control"
                      v-model="formData.api_key"
                      placeholder="Leave empty to use default API key from Settings"
                    />
                    <button
                      type="button"
                      class="btn btn-outline-secondary"
                      @click="showApiKey = !showApiKey"
                      :title="showApiKey ? 'Hide API Key' : 'Show API Key'"
                    >
                      <i :class="showApiKey ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
                    </button>
                  </div>
                  <small class="text-muted">
                    <i class="fas fa-shield-alt"></i>
                    API keys are stored only in your browser's local storage and never sent to the server.
                  </small>
                </div>
                <div class="col-md-6">
                  <div class="form-check form-switch">
                    <input class="form-check-input" type="checkbox" id="isEnabled" v-model="formData.is_enabled" />
                    <label class="form-check-label" for="isEnabled">Enabled</label>
                  </div>
                </div>
                <div class="col-md-6">
                  <div class="form-check form-switch">
                    <input class="form-check-input" type="checkbox" id="isDefault" v-model="formData.is_default" />
                    <label class="form-check-label" for="isDefault">Set as Default</label>
                  </div>
                </div>
                <div class="col-12" v-if="!isEditing">
                  <label class="form-label">Scope</label>
                  <select class="form-select" v-model="formData.scope">
                    <option value="personal">Personal (only visible to me)</option>
                    <option value="shared">Shared (visible to my organization)</option>
                    <option value="templates" v-if="authStore.isSuperAdmin">Global Template (visible to everyone)</option>
                  </select>
                  <small v-if="formData.scope === 'templates'" class="text-warning">
                    <i class="fas fa-exclamation-triangle"></i>
                    Global templates are visible to all users across all organizations.
                  </small>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeModal">Cancel</button>
              <button type="submit" class="btn btn-primary" :disabled="llmModelStore.loading">
                <span v-if="llmModelStore.loading" class="spinner-border spinner-border-sm me-1"></span>
                {{ isEditing ? 'Update' : 'Create' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div class="modal fade" id="deleteModal" tabindex="-1" ref="deleteModalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Confirm Delete</h5>
            <button type="button" class="btn-close" @click="closeDeleteModal"></button>
          </div>
          <div class="modal-body">
            <p>Are you sure you want to delete the model <strong>{{ modelToDelete?.name }}</strong>?</p>
            <p class="text-muted mb-0">This action cannot be undone.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeDeleteModal">Cancel</button>
            <button type="button" class="btn btn-danger" @click="deleteModel" :disabled="llmModelStore.loading">
              <span v-if="llmModelStore.loading" class="spinner-border spinner-border-sm me-1"></span>
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useLLMModelStore } from '@/stores/llmModelStore';
import { useAuthStore } from '@/stores/authStore';
import type { LLMModel, CreateLLMModelRequest, UpdateLLMModelRequest } from '@/services/llmModelService';
import { Modal } from 'bootstrap';

const llmModelStore = useLLMModelStore();
const authStore = useAuthStore();

// Refs
const modalRef = ref<HTMLElement | null>(null);
const deleteModalRef = ref<HTMLElement | null>(null);
let modalInstance: Modal | null = null;
let deleteModalInstance: Modal | null = null;

// State
const showCreateModal = ref(false);
const isEditing = ref(false);
const editingModelId = ref<string | null>(null);
const modelToDelete = ref<LLMModel | null>(null);
const searchQuery = ref('');
const scope = ref<'all' | 'personal' | 'shared' | 'templates'>('all');
const enabledOnly = ref(false);
const showApiKey = ref(false);

// Form data (api_key is stored locally, not sent to server)
const formData = ref<CreateLLMModelRequest & { is_enabled: boolean; is_default: boolean; api_key: string }>({
  name: '',
  llm_type: 'openai',
  base_url: '',
  model: '',
  extra_params: '',
  temperature: 0.7,
  max_tokens: 4096,
  is_default: false,
  is_enabled: true,
  description: '',
  scope: 'personal',
  api_key: '',
});

// Computed
const displayedPages = computed(() => {
  const pages: number[] = [];
  const total = llmModelStore.totalPages;
  const current = llmModelStore.pageNumber;
  const delta = 2;

  for (let i = Math.max(1, current - delta); i <= Math.min(total, current + delta); i++) {
    pages.push(i);
  }
  return pages;
});

// Methods
function getLLMTypeBadgeClass(type: string): string {
  const classes: Record<string, string> = {
    openai: 'bg-success',
    anthropic: 'bg-primary',
    google: 'bg-info',
    alibaba: 'bg-warning text-dark',
    deepseek: 'bg-danger',
    custom: 'bg-secondary',
  };
  return classes[type] || 'bg-secondary';
}

function resetForm() {
  formData.value = {
    name: '',
    llm_type: 'openai',
    base_url: '',
    model: '',
    extra_params: '',
    temperature: 0.7,
    max_tokens: 4096,
    is_default: false,
    is_enabled: true,
    description: '',
    scope: 'personal',
    api_key: '',
  };
  isEditing.value = false;
  editingModelId.value = null;
  showApiKey.value = false;
}

function openModal() {
  if (modalRef.value && !modalInstance) {
    modalInstance = new Modal(modalRef.value);
  }
  modalInstance?.show();
}

function closeModal() {
  modalInstance?.hide();
  resetForm();
}

function editModel(model: LLMModel) {
  isEditing.value = true;
  editingModelId.value = model.id;
  // Get API key from local storage
  const storedApiKey = llmModelStore.getApiKey(model.id);
  formData.value = {
    name: model.name,
    llm_type: model.llm_type,
    base_url: model.base_url,
    model: model.model,
    extra_params: model.extra_params || '',
    temperature: model.temperature,
    max_tokens: model.max_tokens,
    is_default: model.is_default,
    is_enabled: model.is_enabled,
    description: model.description || '',
    scope: 'personal',
    api_key: storedApiKey || '',
  };
  showApiKey.value = false;
  openModal();
}

async function saveModel() {
  try {
    // Extract api_key before sending to server (it's stored locally only)
    const { api_key, ...serverData } = formData.value;
    
    let modelId: string;
    if (isEditing.value && editingModelId.value) {
      const updateData: UpdateLLMModelRequest = {
        name: serverData.name,
        llm_type: serverData.llm_type,
        base_url: serverData.base_url,
        model: serverData.model,
        extra_params: serverData.extra_params,
        temperature: serverData.temperature,
        max_tokens: serverData.max_tokens,
        is_default: serverData.is_default,
        is_enabled: serverData.is_enabled,
        description: serverData.description,
      };
      await llmModelStore.updateModel(editingModelId.value, updateData);
      modelId = editingModelId.value;
    } else {
      const newModel = await llmModelStore.createModel(serverData);
      modelId = newModel.id;
    }
    
    // Save API key locally (separate from server data)
    if (api_key) {
      llmModelStore.setApiKey(modelId, api_key);
    } else if (isEditing.value) {
      // If editing and API key was cleared, remove it
      llmModelStore.setApiKey(modelId, null);
    }
    
    closeModal();
  } catch (error) {
    // Error is handled in store
  }
}

async function setAsDefault(id: string) {
  await llmModelStore.setDefault(id);
}

async function toggleEnabled(model: LLMModel) {
  await llmModelStore.toggleEnabled(model.id, !model.is_enabled);
}

function confirmDelete(model: LLMModel) {
  modelToDelete.value = model;
  if (deleteModalRef.value && !deleteModalInstance) {
    deleteModalInstance = new Modal(deleteModalRef.value);
  }
  deleteModalInstance?.show();
}

function closeDeleteModal() {
  deleteModalInstance?.hide();
  modelToDelete.value = null;
}

async function deleteModel() {
  if (modelToDelete.value) {
    await llmModelStore.deleteModel(modelToDelete.value.id);
    closeDeleteModal();
  }
}

// Debounced search
let searchTimeout: ReturnType<typeof setTimeout>;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    llmModelStore.setSearchQuery(searchQuery.value);
  }, 300);
}

// Watch for showCreateModal
watch(showCreateModal, (show) => {
  if (show) {
    resetForm();
    openModal();
    showCreateModal.value = false;
  }
});

// Lifecycle
onMounted(() => {
  llmModelStore.fetchModels();
});
</script>

<style scoped>
.table-primary {
  --bs-table-bg: rgba(13, 110, 253, 0.1);
}

code {
  font-size: 0.85em;
}

.font-monospace {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 0.875rem;
}
</style>

