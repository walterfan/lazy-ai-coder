<template>
  <div class="framework-selector">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <div>
        <h5 class="mb-0">
          <i class="fas fa-layer-group"></i> Select Prompt Engineering Framework
        </h5>
        <p class="text-muted small mb-0">
          Choose a framework that best suits your task.
        </p>
      </div>
      <button
        class="btn btn-sm btn-primary"
        @click="showCreateModal"
        title="Create new framework"
      >
        <i class="fas fa-plus"></i> New Framework
      </button>
    </div>

    <div v-if="loading" class="text-center py-4">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading frameworks...</span>
      </div>
    </div>

    <div v-else class="row g-3">
      <div
        v-for="framework in frameworks"
        :key="framework.id"
        class="col-md-6 col-lg-4"
      >
        <div
          class="card framework-card h-100"
          :class="{ 'selected': selectedFrameworkId === framework.id }"
          @click="selectFramework(framework.id)"
          role="button"
          tabindex="0"
        >
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-start mb-2">
              <h6 class="card-title mb-0">
                <i class="fas fa-check-circle text-success me-2" v-if="selectedFrameworkId === framework.id"></i>
                {{ framework.name }}
              </h6>
              <div class="d-flex gap-1">
                <span class="badge bg-primary">{{ framework.fields.length }} fields</span>
                <button
                  class="btn btn-sm btn-outline-primary p-0 px-1"
                  @click.stop="showEditModal(framework)"
                  title="Edit framework"
                >
                  <i class="fas fa-edit"></i>
                </button>
                <button
                  class="btn btn-sm btn-outline-danger p-0 px-1"
                  @click.stop="deleteFramework(framework)"
                  title="Delete framework"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </div>
            <p class="card-text small text-muted mb-2">
              {{ framework.description }}
            </p>
            <div class="best-for mt-2">
              <strong class="small text-primary">Best for:</strong>
              <p class="small mb-0">{{ framework.best_for }}</p>
            </div>
          </div>
          <div class="card-footer bg-transparent border-0 pt-0">
            <button
              class="btn btn-sm btn-outline-secondary w-100"
              @click.stop="showExample(framework)"
            >
              <i class="fas fa-eye"></i> View Example
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Example Modal -->
    <div
      class="modal fade"
      id="exampleModal"
      tabindex="-1"
      aria-labelledby="exampleModalLabel"
      aria-hidden="true"
      ref="exampleModal"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header bg-primary text-white">
            <h5 class="modal-title" id="exampleModalLabel">
              <i class="fas fa-lightbulb"></i> {{ viewingFramework?.name }} Example
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="viewingFramework">
              <div class="mb-3">
                <h6 class="fw-bold text-primary">Use Case:</h6>
                <p>{{ viewingFramework.example.use_case }}</p>
              </div>
              <div class="mb-3">
                <h6 class="fw-bold text-primary">Input:</h6>
                <p class="bg-light p-2 rounded">{{ viewingFramework.example.input }}</p>
              </div>
              <div class="mb-3">
                <h6 class="fw-bold text-primary">Generated Prompt:</h6>
                <pre class="bg-light p-3 rounded" style="max-height: 400px; overflow-y: auto;">{{ viewingFramework.example.generated_prompt }}</pre>
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
              @click="selectAndClose"
              v-if="selectedFrameworkId !== viewingFramework?.id"
            >
              <i class="fas fa-check"></i> Use This Framework
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Framework Modal -->
    <div
      class="modal fade"
      id="frameworkModal"
      tabindex="-1"
      aria-labelledby="frameworkModalLabel"
      aria-hidden="true"
      ref="frameworkModalEl"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="frameworkModalLabel">
              <i class="fas fa-layer-group"></i>
              {{ editingFramework ? 'Edit Framework' : 'Create New Framework' }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveFramework">
              <div class="row">
                <div class="col-md-6 mb-3">
                  <label class="form-label">Framework ID * (lowercase, no spaces)</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="frameworkForm.id"
                    :disabled="!!editingFramework"
                    required
                    pattern="[a-z-]+"
                    placeholder="e.g., crispe"
                  />
                </div>
                <div class="col-md-6 mb-3">
                  <label class="form-label">Framework Name *</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="frameworkForm.name"
                    required
                    placeholder="e.g., CRISPE"
                  />
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label">Description *</label>
                <textarea
                  class="form-control"
                  v-model="frameworkForm.description"
                  rows="2"
                  required
                  placeholder="Brief description of the framework"
                ></textarea>
              </div>

              <div class="mb-3">
                <label class="form-label">Best For *</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="frameworkForm.best_for"
                  required
                  placeholder="What is this framework best suited for?"
                />
              </div>

              <div class="mb-3">
                <label class="form-label">Fields *</label>
                <small class="text-muted d-block mb-2">Define the framework fields (one JSON object per line)</small>
                <textarea
                  class="form-control font-monospace"
                  v-model="frameworkForm.fieldsJson"
                  rows="8"
                  required
                  placeholder='[&#10;  {"id": "capacity", "label": "Capacity", "description": "Role and expertise", "placeholder": "You are..."},&#10;  {"id": "insight", "label": "Insight", "description": "Context and background", "placeholder": "The context is..."}&#10;]'
                ></textarea>
              </div>

              <div class="alert alert-info">
                <strong>Note:</strong> Changes to frameworks are temporary and will be reset on server restart.
                For permanent changes, update the backend code.
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
              @click="saveFramework"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save"></i>
              {{ editingFramework ? 'Update' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { apiService } from '@/services/apiService';
import type { Framework } from '@/types/smart-prompt';
import { Modal } from 'bootstrap';
import { showToast } from '@/utils/toast';

defineProps<{
  selectedFrameworkId?: string;
}>();

const emit = defineEmits<{
  (e: 'framework-selected', frameworkId: string): void;
}>();

const loading = ref(false);
const isSaving = ref(false);
const frameworks = ref<Framework[]>([]);
const viewingFramework = ref<Framework | null>(null);
const editingFramework = ref<Framework | null>(null);
const frameworkForm = ref({
  id: '',
  name: '',
  description: '',
  best_for: '',
  fieldsJson: '[]',
});

let exampleModal: Modal | null = null;
let frameworkModal: Modal | null = null;

onMounted(async () => {
  await loadFrameworks();
});

async function loadFrameworks() {
  try {
    loading.value = true;
    frameworks.value = await apiService.getFrameworks();
  } catch (error) {
    console.error('Failed to load frameworks:', error);
    showToast('Failed to load frameworks', 'danger');
  } finally {
    loading.value = false;
  }
}

function selectFramework(frameworkId: string) {
  emit('framework-selected', frameworkId);
}

function showExample(framework: Framework) {
  viewingFramework.value = framework;
  const modalEl = document.getElementById('exampleModal');
  if (modalEl) {
    exampleModal = new Modal(modalEl);
    exampleModal.show();
  }
}

function selectAndClose() {
  if (viewingFramework.value) {
    selectFramework(viewingFramework.value.id);
  }
  if (exampleModal) {
    exampleModal.hide();
  }
}

function showCreateModal() {
  editingFramework.value = null;
  frameworkForm.value = {
    id: '',
    name: '',
    description: '',
    best_for: '',
    fieldsJson: '[\n  {\n    "id": "field1",\n    "label": "Field 1",\n    "description": "Description",\n    "placeholder": "Enter..."\n  }\n]',
  };
  const modalEl = document.getElementById('frameworkModal');
  if (modalEl) {
    frameworkModal = new Modal(modalEl);
    frameworkModal.show();
  }
}

function showEditModal(framework: Framework) {
  editingFramework.value = framework;
  frameworkForm.value = {
    id: framework.id,
    name: framework.name,
    description: framework.description,
    best_for: framework.best_for,
    fieldsJson: JSON.stringify(framework.fields, null, 2),
  };
  const modalEl = document.getElementById('frameworkModal');
  if (modalEl) {
    frameworkModal = new Modal(modalEl);
    frameworkModal.show();
  }
}

async function saveFramework() {
  try {
    isSaving.value = true;

    // Parse and validate fields JSON
    let fields;
    try {
      fields = JSON.parse(frameworkForm.value.fieldsJson);
      if (!Array.isArray(fields)) {
        throw new Error('Fields must be an array');
      }
    } catch (error) {
      showToast('Invalid fields JSON format', 'danger');
      return;
    }

    const frameworkData: Framework = {
      id: frameworkForm.value.id,
      name: frameworkForm.value.name,
      description: frameworkForm.value.description,
      best_for: frameworkForm.value.best_for,
      fields: fields,
      template: editingFramework.value?.template || '', // Add template field
      example: editingFramework.value?.example || {
        use_case: 'Example use case',
        input: 'Example input',
        generated_prompt: 'Example generated prompt',
      },
    };

    if (editingFramework.value) {
      await apiService.updateFramework(frameworkData.id, frameworkData);
      showToast('Framework updated successfully!', 'success');
    } else {
      await apiService.createFramework(frameworkData);
      showToast('Framework created successfully!', 'success');
    }

    // Reload frameworks
    await loadFrameworks();

    // Close modal
    if (frameworkModal) {
      frameworkModal.hide();
    }
  } catch (error) {
    console.error('Failed to save framework:', error);
    showToast('Failed to save framework. See console for details.', 'danger');
  } finally {
    isSaving.value = false;
  }
}

async function deleteFramework(framework: Framework) {
  if (!confirm(`Are you sure you want to delete the framework "${framework.name}"?`)) {
    return;
  }

  try {
    await apiService.deleteFramework(framework.id);
    showToast('Framework deleted successfully!', 'success');
    await loadFrameworks();
  } catch (error) {
    console.error('Failed to delete framework:', error);
    showToast('Failed to delete framework', 'danger');
  }
}
</script>

<style scoped>
.framework-card {
  cursor: pointer;
  transition: all 0.3s ease;
  border: 2px solid #dee2e6;
}

.framework-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-color: #0d6efd;
}

.framework-card.selected {
  border-color: #0d6efd;
  background-color: #f8f9ff;
  box-shadow: 0 4px 12px rgba(13, 110, 253, 0.2);
}

.best-for {
  border-left: 3px solid #0d6efd;
  padding-left: 10px;
  margin-top: 10px;
}

pre {
  font-size: 0.85rem;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>
