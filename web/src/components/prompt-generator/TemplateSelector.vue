<template>
  <div class="template-selector">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h5 class="mb-0">
        <i class="fas fa-th-large"></i> Choose a Template (Optional)
      </h5>
      <div class="d-flex gap-2">
        <button class="btn btn-sm btn-primary" @click="showCreateModal">
          <i class="fas fa-plus"></i> New Template
        </button>
        <button class="btn btn-sm btn-outline-secondary" @click="emit('skip')">
          <i class="fas fa-forward"></i> Skip
        </button>
      </div>
    </div>
    <p class="text-muted small mb-4">
      Start with a pre-built template or skip to create from scratch.
    </p>

    <!-- Category Tabs -->
    <ul class="nav nav-pills mb-3" role="tablist">
      <li class="nav-item" role="presentation" v-for="category in categories" :key="category.id">
        <button
          class="nav-link"
          :class="{ active: selectedCategory === category.id }"
          @click="selectCategory(category.id)"
          type="button"
        >
          <i class="fas" :class="`fa-${category.icon}`"></i>
          {{ category.name }}
        </button>
      </li>
    </ul>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-4">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading templates...</span>
      </div>
    </div>

    <!-- Templates Grid -->
    <div v-else-if="filteredTemplates.length > 0" class="row g-3">
      <div
        v-for="template in filteredTemplates"
        :key="template.id"
        class="col-md-6"
      >
        <div class="card template-card h-100">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-start mb-2">
              <h6 class="card-title">{{ template.name }}</h6>
              <span class="badge bg-secondary" v-if="template.use_count > 0">
                {{ template.use_count }} uses
              </span>
            </div>
            <p class="card-text small text-muted">{{ template.description }}</p>

            <div class="d-flex flex-wrap gap-1 mb-2">
              <span
                v-for="tag in template.tags.slice(0, 3)"
                :key="tag"
                class="badge bg-info text-dark"
              >
                {{ tag }}
              </span>
            </div>

            <div class="d-flex justify-content-between align-items-center mt-3">
              <small class="text-muted">
                <i class="fas fa-layer-group"></i>
                {{ getFrameworkName(template.framework) }}
              </small>
              <div class="d-flex gap-1">
                <button
                  class="btn btn-sm btn-primary"
                  @click="useTemplate(template)"
                >
                  <i class="fas fa-play"></i> Use
                </button>
                <button
                  class="btn btn-sm btn-outline-primary"
                  @click="showEditModal(template)"
                  title="Edit template"
                >
                  <i class="fas fa-edit"></i>
                </button>
                <button
                  class="btn btn-sm btn-outline-danger"
                  @click="deleteTemplate(template)"
                  title="Delete template"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="alert alert-info">
      <i class="fas fa-info-circle"></i>
      No templates found in this category.
    </div>

    <!-- Create/Edit Template Modal -->
    <div
      class="modal fade"
      id="templateModal"
      tabindex="-1"
      aria-labelledby="templateModalLabel"
      aria-hidden="true"
      ref="templateModalEl"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="templateModalLabel">
              <i class="fas fa-th-large"></i>
              {{ editingTemplate ? 'Edit Template' : 'Create New Template' }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveTemplate">
              <div class="row">
                <div class="col-md-6 mb-3">
                  <label class="form-label">Template ID * (lowercase, no spaces)</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="templateForm.id"
                    :disabled="!!editingTemplate"
                    required
                    pattern="[a-z-]+"
                    placeholder="e.g., code-review-java"
                  />
                </div>
                <div class="col-md-6 mb-3">
                  <label class="form-label">Template Name *</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="templateForm.name"
                    required
                    placeholder="e.g., Java Code Review"
                  />
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label">Description *</label>
                <textarea
                  class="form-control"
                  v-model="templateForm.description"
                  rows="2"
                  required
                  placeholder="Brief description of the template"
                ></textarea>
              </div>

              <div class="row">
                <div class="col-md-6 mb-3">
                  <label class="form-label">Category *</label>
                  <select
                    class="form-select"
                    v-model="templateForm.category"
                    required
                  >
                    <option value="">Select a category...</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                      {{ cat.name }}
                    </option>
                  </select>
                </div>
                <div class="col-md-6 mb-3">
                  <label class="form-label">Framework *</label>
                  <select
                    class="form-select"
                    v-model="templateForm.framework"
                    required
                  >
                    <option value="">Select a framework...</option>
                    <option v-for="fw in frameworks" :key="fw.id" :value="fw.id">
                      {{ fw.name }}
                    </option>
                  </select>
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label">Tags (comma-separated)</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="templateForm.tagsStr"
                  placeholder="e.g., code, review, java"
                />
                <small class="text-muted">Enter tags separated by commas</small>
              </div>

              <div class="mb-3">
                <label class="form-label">Field Values (JSON) *</label>
                <small class="text-muted d-block mb-2">
                  Define pre-filled values for the selected framework's fields
                </small>
                <textarea
                  class="form-control font-monospace"
                  v-model="templateForm.fieldsJson"
                  rows="10"
                  required
                  placeholder='{&#10;  "field_id_1": "Pre-filled value for field 1",&#10;  "field_id_2": "Pre-filled value for field 2"&#10;}'
                ></textarea>
              </div>

              <div class="alert alert-info">
                <strong>Note:</strong> Changes to templates are temporary and will be reset on server restart.
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
              @click="saveTemplate"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save"></i>
              {{ editingTemplate ? 'Update' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { apiService } from '@/services/apiService';
import type { PromptTemplate, TemplateCategory, Framework } from '@/types/smart-prompt';
import { Modal } from 'bootstrap';
import { showToast } from '@/utils/toast';

const emit = defineEmits<{
  (e: 'template-selected', template: PromptTemplate): void;
  (e: 'skip'): void;
}>();

const loading = ref(false);
const isSaving = ref(false);
const categories = ref<TemplateCategory[]>([]);
const templates = ref<PromptTemplate[]>([]);
const frameworks = ref<Framework[]>([]);
const selectedCategory = ref<string>('');
const editingTemplate = ref<PromptTemplate | null>(null);
const templateForm = ref({
  id: '',
  name: '',
  description: '',
  category: '',
  framework: '',
  tagsStr: '',
  fieldsJson: '{}',
});

let templateModal: Modal | null = null;

const filteredTemplates = computed(() => {
  if (!selectedCategory.value) return templates.value;
  return templates.value.filter(t => t.category === selectedCategory.value);
});

onMounted(async () => {
  await Promise.all([
    loadCategories(),
    loadTemplates(),
    loadFrameworks(),
  ]);
  if (categories.value.length > 0) {
    selectedCategory.value = categories.value[0].id;
  }
});

async function loadCategories() {
  try {
    categories.value = await apiService.getTemplateCategories();
  } catch (error) {
    console.error('Failed to load categories:', error);
  }
}

async function loadTemplates() {
  try {
    loading.value = true;
    templates.value = await apiService.getTemplates();
  } catch (error) {
    console.error('Failed to load templates:', error);
  } finally {
    loading.value = false;
  }
}

async function loadFrameworks() {
  try {
    frameworks.value = await apiService.getFrameworks();
  } catch (error) {
    console.error('Failed to load frameworks:', error);
  }
}

function selectCategory(categoryId: string) {
  selectedCategory.value = categoryId;
}

function getFrameworkName(frameworkId: string): string {
  const framework = frameworks.value.find(f => f.id === frameworkId);
  return framework ? framework.name : frameworkId.toUpperCase();
}

async function useTemplate(template: PromptTemplate) {
  try {
    // Track usage
    await apiService.useTemplate(template.id);
    showToast(`Using template: ${template.name}`, 'success');
    emit('template-selected', template);
  } catch (error) {
    console.error('Failed to use template:', error);
    showToast('Failed to load template', 'danger');
  }
}

function showCreateModal() {
  editingTemplate.value = null;
  templateForm.value = {
    id: '',
    name: '',
    description: '',
    category: categories.value.length > 0 ? categories.value[0].id : '',
    framework: frameworks.value.length > 0 ? frameworks.value[0].id : '',
    tagsStr: '',
    fieldsJson: '{\n  "field1": "value1",\n  "field2": "value2"\n}',
  };
  const modalEl = document.getElementById('templateModal');
  if (modalEl) {
    templateModal = new Modal(modalEl);
    templateModal.show();
  }
}

function showEditModal(template: PromptTemplate) {
  editingTemplate.value = template;
  templateForm.value = {
    id: template.id,
    name: template.name,
    description: template.description,
    category: template.category,
    framework: template.framework,
    tagsStr: template.tags.join(', '),
    fieldsJson: JSON.stringify(template.fields, null, 2),
  };
  const modalEl = document.getElementById('templateModal');
  if (modalEl) {
    templateModal = new Modal(modalEl);
    templateModal.show();
  }
}

async function saveTemplate() {
  try {
    isSaving.value = true;

    // Parse and validate fields JSON
    let fields;
    try {
      fields = JSON.parse(templateForm.value.fieldsJson);
      if (typeof fields !== 'object' || fields === null) {
        throw new Error('Fields must be an object');
      }
    } catch (error) {
      showToast('Invalid fields JSON format', 'danger');
      return;
    }

    // Parse tags
    const tags = templateForm.value.tagsStr
      .split(',')
      .map(tag => tag.trim())
      .filter(tag => tag.length > 0);

    const templateData: PromptTemplate = {
      id: templateForm.value.id,
      name: templateForm.value.name,
      description: templateForm.value.description,
      category: templateForm.value.category,
      framework: templateForm.value.framework,
      tags: tags,
      fields: fields,
      use_count: editingTemplate.value?.use_count || 0,
    };

    if (editingTemplate.value) {
      await apiService.updateTemplate(templateData.id, templateData);
      showToast('Template updated successfully!', 'success');
    } else {
      await apiService.createTemplate(templateData);
      showToast('Template created successfully!', 'success');
    }

    // Reload templates
    await loadTemplates();

    // Close modal
    if (templateModal) {
      templateModal.hide();
    }
  } catch (error) {
    console.error('Failed to save template:', error);
    showToast('Failed to save template. See console for details.', 'danger');
  } finally {
    isSaving.value = false;
  }
}

async function deleteTemplate(template: PromptTemplate) {
  if (!confirm(`Are you sure you want to delete the template "${template.name}"?`)) {
    return;
  }

  try {
    await apiService.deleteTemplate(template.id);
    showToast('Template deleted successfully!', 'success');
    await loadTemplates();
  } catch (error) {
    console.error('Failed to delete template:', error);
    showToast('Failed to delete template', 'danger');
  }
}
</script>

<style scoped>
.template-card {
  transition: all 0.3s ease;
  border: 1px solid #dee2e6;
}

.template-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-color: #0d6efd;
}

.nav-pills .nav-link {
  color: #6c757d;
  margin-right: 0.5rem;
  margin-bottom: 0.5rem;
}

.nav-pills .nav-link.active {
  background-color: #0d6efd;
}

.nav-pills .nav-link:hover:not(.active) {
  background-color: #e9ecef;
}
</style>
