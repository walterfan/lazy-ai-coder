<template>
  <div class="card">
    <div class="card-header bg-primary text-white d-flex justify-content-between align-items-center">
      <h5 class="mb-0">
        <i class="bi bi-file-earmark-text me-2"></i>
        Generated CAR Prompt
      </h5>
      <div>
        <button class="btn btn-sm btn-light me-2" @click="copyToClipboard" title="Copy to clipboard">
          <i class="bi bi-clipboard"></i> Copy
        </button>
        <button class="btn btn-sm btn-light" @click="toggleEdit" :title="editing ? 'Done editing' : 'Edit'">
          <i :class="editing ? 'bi bi-check-lg' : 'bi bi-pencil'"></i>
          {{ editing ? 'Done' : 'Edit' }}
        </button>
      </div>
    </div>
    <div class="card-body">
      <!-- Context Section -->
      <div class="mb-4">
        <h6 class="text-primary">
          <strong>Context:</strong>
        </h6>
        <div v-if="!editing" class="prompt-section">
          <pre class="mb-0">{{ localContext }}</pre>
        </div>
        <textarea
          v-else
          v-model="localContext"
          class="form-control"
          rows="5"
        ></textarea>
      </div>

      <!-- Action Section -->
      <div class="mb-4">
        <h6 class="text-success">
          <strong>Action:</strong>
        </h6>
        <div v-if="!editing" class="prompt-section">
          <pre class="mb-0">{{ localAction }}</pre>
        </div>
        <textarea
          v-else
          v-model="localAction"
          class="form-control"
          rows="3"
        ></textarea>
      </div>

      <!-- Result Section -->
      <div class="mb-4">
        <h6 class="text-warning">
          <strong>Result:</strong>
        </h6>
        <div v-if="!editing" class="prompt-section">
          <pre class="mb-0">{{ localResult }}</pre>
        </div>
        <textarea
          v-else
          v-model="localResult"
          class="form-control"
          rows="8"
        ></textarea>
      </div>

      <!-- Copy status -->
      <div v-if="copySuccess" class="alert alert-success alert-dismissible fade show" role="alert">
        Prompt copied to clipboard!
        <button type="button" class="btn-close" @click="copySuccess = false"></button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
  context: string;
  action: string;
  result: string;
  fullPrompt: string;
}>();

const editing = ref(false);
const copySuccess = ref(false);

// Local editable copies
const localContext = ref(props.context);
const localAction = ref(props.action);
const localResult = ref(props.result);

function toggleEdit() {
  editing.value = !editing.value;
}

async function copyToClipboard() {
  const fullText = `**Context:**
${localContext.value}

**Action:**
${localAction.value}

**Result:**
${localResult.value}`;

  try {
    await navigator.clipboard.writeText(fullText);
    copySuccess.value = true;
    setTimeout(() => {
      copySuccess.value = false;
    }, 3000);
  } catch (err) {
    console.error('Failed to copy:', err);
  }
}
</script>

<style scoped>
.prompt-section {
  background-color: #f8f9fa;
  padding: 1rem;
  border-radius: 0.25rem;
  border: 1px solid #dee2e6;
}

.prompt-section pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
}

textarea.form-control {
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
}

.card-header {
  background-color: #0d6efd;
}
</style>
