<template>
  <div class="prompt-builder">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h5 class="mb-0">
        <i class="fas fa-edit"></i> Build Your Prompt
        <span class="badge bg-primary ms-2">{{ framework.name }}</span>
      </h5>
      <button class="btn btn-sm btn-outline-secondary" @click="emit('change-framework')">
        <i class="fas fa-exchange-alt"></i> Change Framework
      </button>
    </div>

    <p class="text-muted small mb-4">
      {{ framework.description }}
    </p>

    <!-- AI Auto-fill Section -->
    <div class="card bg-gradient-primary text-white mb-4">
      <div class="card-body">
        <h6 class="card-title mb-3">
          <i class="fas fa-robot"></i> AI-Powered Auto-fill
        </h6>
        <p class="small mb-3">
          Describe what you want to create, and AI will automatically fill the framework fields for you.
        </p>
        <textarea
          v-model="userInput"
          class="form-control mb-3"
          placeholder="Example: Create a REST API endpoint for user authentication with JWT tokens using FastAPI and PostgreSQL..."
          rows="3"
          :disabled="autoFilling"
        ></textarea>
        <button
          type="button"
          class="btn btn-light"
          @click="autoFillFields"
          :disabled="!userInput.trim() || autoFilling"
        >
          <span v-if="autoFilling" class="spinner-border spinner-border-sm me-2"></span>
          <i v-else class="fas fa-magic"></i>
          {{ autoFilling ? 'Auto-filling...' : 'Auto-fill with AI' }}
        </button>
      </div>
    </div>

    <form @submit.prevent="generate">
      <div v-for="field in framework.fields" :key="field.id" class="mb-4">
        <label :for="field.id" class="form-label">
          {{ field.label }}
          <span class="text-danger" v-if="field.required">*</span>
        </label>

        <div class="field-help mb-2">
          <small class="text-muted">
            <i class="fas fa-info-circle"></i> {{ field.description }}
          </small>
        </div>

        <!-- Text Input -->
        <input
          v-if="field.type === 'text'"
          :id="field.id"
          type="text"
          class="form-control"
          v-model="fieldValues[field.id]"
          :placeholder="field.placeholder"
          :required="field.required"
        />

        <!-- Textarea Input -->
        <textarea
          v-else-if="field.type === 'textarea'"
          :id="field.id"
          class="form-control"
          v-model="fieldValues[field.id]"
          :placeholder="field.placeholder"
          :required="field.required"
          rows="4"
        ></textarea>

        <!-- Select Input -->
        <select
          v-else-if="field.type === 'select' && field.options"
          :id="field.id"
          class="form-select"
          v-model="fieldValues[field.id]"
          :required="field.required"
        >
          <option value="">-- Select --</option>
          <option v-for="option in field.options" :key="option" :value="option">
            {{ option }}
          </option>
        </select>

        <!-- Character count for textareas -->
        <div v-if="field.type === 'textarea' && fieldValues[field.id]" class="text-end">
          <small class="text-muted">
            {{ fieldValues[field.id].length }} characters
          </small>
        </div>
      </div>

      <!-- Quick Tips -->
      <div class="card bg-light border-0 mb-4" v-if="quickTips.length > 0">
        <div class="card-body">
          <h6 class="card-title">
            <i class="fas fa-lightbulb text-warning"></i> Quick Tips for {{ framework.name }}
          </h6>
          <ul class="mb-0 small">
            <li v-for="(tip, index) in quickTips.slice(0, 3)" :key="index">
              {{ tip }}
            </li>
          </ul>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="d-flex gap-2 justify-content-between">
        <div class="d-flex gap-2">
          <button
            type="button"
            class="btn btn-outline-secondary"
            @click="clearFields"
          >
            <i class="fas fa-eraser"></i> Clear All
          </button>
          <button
            type="button"
            class="btn btn-outline-info"
            @click="getQuickRefine"
            :disabled="!hasAnyContent || refining"
          >
            <span v-if="refining" class="spinner-border spinner-border-sm me-2"></span>
            <i v-else class="fas fa-magic"></i>
            Quick Tips
          </button>
        </div>
        <button
          type="submit"
          class="btn btn-primary"
          :disabled="!isValid || generating"
        >
          <span v-if="generating" class="spinner-border spinner-border-sm me-2"></span>
          <i v-else class="fas fa-wand-magic-sparkles"></i>
          Generate Prompt
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import type { Framework } from '@/types/smart-prompt';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import { useSettingsStore } from '@/stores/settingsStore';

const settingsStore = useSettingsStore();

const props = defineProps<{
  framework: Framework;
  initialValues?: Record<string, string>;
}>();

const emit = defineEmits<{
  (e: 'change-framework'): void;
  (e: 'generate', fields: Record<string, string>): void;
  (e: 'quick-tips', tips: string[]): void;
}>();

const fieldValues = ref<Record<string, string>>({});
const generating = ref(false);
const refining = ref(false);
const quickTips = ref<string[]>([]);
const userInput = ref('');
const autoFilling = ref(false);

// Initialize field values
onMounted(() => {
  initializeFields();
});

// Watch for framework changes
watch(() => props.framework, () => {
  initializeFields();
}, { immediate: true });

// Watch for initial values (from template)
watch(() => props.initialValues, (newValues) => {
  if (newValues) {
    fieldValues.value = { ...fieldValues.value, ...newValues };
  }
}, { immediate: true });

function initializeFields() {
  const values: Record<string, string> = {};
  props.framework.fields.forEach(field => {
    values[field.id] = props.initialValues?.[field.id] || '';
  });
  fieldValues.value = values;
}

const isValid = computed(() => {
  return props.framework.fields
    .filter(f => f.required)
    .every(f => fieldValues.value[f.id]?.trim());
});

const hasAnyContent = computed(() => {
  return Object.values(fieldValues.value).some(v => v?.trim());
});

function clearFields() {
  if (confirm('Clear all fields?')) {
    initializeFields();
    quickTips.value = [];
  }
}

async function generate() {
  if (!isValid.value) {
    showToast('Please fill in all required fields', 'warning');
    return;
  }

  try {
    generating.value = true;
    emit('generate', fieldValues.value);
  } catch (error) {
    console.error('Generation error:', error);
    showToast('Failed to generate prompt', 'danger');
  } finally {
    generating.value = false;
  }
}

async function getQuickRefine() {
  try {
    refining.value = true;

    // Build a simple prompt from current values
    const promptText = Object.entries(fieldValues.value)
      .filter(([_, value]) => value?.trim())
      .map(([key, value]) => `${key}: ${value}`)
      .join('\n');

    const response = await apiService.quickRefine(promptText, props.framework.id);
    quickTips.value = response.suggestions;
    emit('quick-tips', response.suggestions);
    showToast('Got quick tips!', 'success');
  } catch (error) {
    console.error('Quick refine error:', error);
    showToast('Failed to get tips', 'danger');
  } finally {
    refining.value = false;
  }
}

async function autoFillFields() {
  if (!userInput.value.trim()) {
    showToast('Please enter your requirement', 'warning');
    return;
  }

  if (!settingsStore.LLM_API_KEY) {
    showToast('Please configure LLM API key in Settings first', 'danger');
    return;
  }

  try {
    autoFilling.value = true;

    const settings = {
      LLM_API_KEY: settingsStore.LLM_API_KEY,
      LLM_MODEL: settingsStore.LLM_MODEL || 'gpt-4',
      LLM_BASE_URL: settingsStore.LLM_BASE_URL || 'https://api.openai.com/v1',
      LLM_TEMPERATURE: settingsStore.LLM_TEMPERATURE || '0.7',
      GITLAB_BASE_URL: settingsStore.GITLAB_BASE_URL,
      GITLAB_TOKEN: settingsStore.GITLAB_TOKEN,
    };

    const response = await apiService.autoFillFields(
      props.framework.id,
      userInput.value,
      settings
    );

    // Update field values with AI-generated content
    fieldValues.value = { ...fieldValues.value, ...response.fields };

    showToast('Fields auto-filled successfully!', 'success');
    userInput.value = ''; // Clear input after successful auto-fill
  } catch (error) {
    console.error('Auto-fill error:', error);
    showToast('Failed to auto-fill fields. Please check your LLM settings.', 'danger');
  } finally {
    autoFilling.value = false;
  }
}
</script>

<style scoped>
.field-help {
  padding: 0.25rem 0.5rem;
  background-color: #f8f9fa;
  border-left: 3px solid #0d6efd;
  border-radius: 4px;
}

textarea.form-control {
  resize: vertical;
  min-height: 100px;
}

.form-label {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.bg-gradient-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.bg-gradient-primary .form-control {
  border: 2px solid rgba(255, 255, 255, 0.3);
  background-color: rgba(255, 255, 255, 0.9);
}

.bg-gradient-primary .form-control:focus {
  border-color: rgba(255, 255, 255, 0.8);
  background-color: white;
  box-shadow: 0 0 0 0.2rem rgba(255, 255, 255, 0.3);
}

.bg-gradient-primary .btn-light {
  font-weight: 600;
}
</style>
