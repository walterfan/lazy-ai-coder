<template>
  <div class="card">
    <div class="card-header bg-success text-white">
      <h5 class="mb-0">
        <i class="bi bi-code-square me-2"></i>
        Code Examples
      </h5>
    </div>
    <div class="card-body">
      <div v-for="(example, index) in examples" :key="index" class="mb-4">
        <h6 class="text-success">
          <strong>{{ example.title }}</strong>
        </h6>
        <p class="text-muted small">{{ example.description }}</p>

        <!-- Code block with syntax highlighting -->
        <div class="code-container position-relative">
          <button
            class="btn btn-sm btn-outline-secondary copy-btn"
            @click="copyCode(example.code, index)"
            :title="'Copy code'"
          >
            <i class="bi bi-clipboard"></i>
          </button>
          <pre><code :class="`language-${example.language}`" v-html="highlightedCode(example.code, example.language)"></code></pre>
        </div>

        <!-- Copy success indicator -->
        <div v-if="copiedIndex === index" class="alert alert-success alert-sm mt-2">
          Code copied to clipboard!
        </div>

        <hr v-if="index < examples.length - 1" class="my-4" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import hljs from 'highlight.js';
import type { CodeExample } from '@/types/smart-prompt';

defineProps<{
  examples: CodeExample[];
}>();

const copiedIndex = ref<number | null>(null);

function highlightedCode(code: string, language: string): string {
  try {
    if (hljs.getLanguage(language)) {
      return hljs.highlight(code, { language }).value;
    } else {
      return hljs.highlightAuto(code).value;
    }
  } catch (err) {
    console.error('Highlight error:', err);
    return code;
  }
}

async function copyCode(code: string, index: number) {
  try {
    await navigator.clipboard.writeText(code);
    copiedIndex.value = index;
    setTimeout(() => {
      copiedIndex.value = null;
    }, 2000);
  } catch (err) {
    console.error('Failed to copy code:', err);
  }
}
</script>

<style scoped>
.card-header {
  background-color: #28a745;
}

.code-container {
  position: relative;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 0.25rem;
  padding: 1rem;
}

.code-container pre {
  margin-bottom: 0;
  overflow-x: auto;
}

.code-container code {
  font-family: 'Courier New', Consolas, Monaco, monospace;
  font-size: 0.875rem;
}

.copy-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  z-index: 10;
}

.alert-sm {
  padding: 0.5rem;
  font-size: 0.875rem;
}
</style>
