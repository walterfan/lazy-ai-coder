<template>
  <div class="refinement-panel">
    <div class="card">
      <div class="card-header bg-primary text-white">
        <h6 class="mb-0">
          <i class="fas fa-magic"></i> AI-Powered Prompt Refinement
        </h6>
      </div>
      <div class="card-body">
        <p class="small text-muted mb-3">
          Let AI analyze your prompt and suggest specific improvements to make it more effective.
        </p>

        <button
          class="btn btn-primary w-100 mb-3"
          @click="refine"
          :disabled="refining || !prompt"
        >
          <span v-if="refining" class="spinner-border spinner-border-sm me-2"></span>
          <i v-else class="fas fa-wand-magic-sparkles"></i>
          {{ refining ? 'Analyzing...' : 'Refine with AI' }}
        </button>

        <!-- Refinement Results -->
        <div v-if="refinementResponse">
          <!-- Assessment -->
          <div class="alert alert-info mb-3">
            <strong><i class="fas fa-clipboard-check"></i> Assessment:</strong>
            <p class="mb-0 mt-2">{{ refinementResponse.assessment }}</p>
          </div>

          <!-- Quality Improvement -->
          <div class="quality-comparison mb-3">
            <div class="row">
              <div class="col-6">
                <div class="text-center">
                  <small class="text-muted">Before</small>
                  <div class="quality-score">{{ refinementResponse.quality_before.toFixed(1) }}/10</div>
                </div>
              </div>
              <div class="col-6">
                <div class="text-center">
                  <small class="text-success">After</small>
                  <div class="quality-score text-success">
                    {{ refinementResponse.quality_after.toFixed(1) }}/10
                    <i class="fas fa-arrow-up ms-1"></i>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Suggestions -->
          <div v-if="refinementResponse.suggestions.length > 0">
            <h6 class="mb-3">
              <i class="fas fa-lightbulb"></i> Specific Improvements ({{ refinementResponse.suggestions.length }})
            </h6>
            <div
              v-for="(suggestion, index) in refinementResponse.suggestions"
              :key="index"
              class="card mb-3 suggestion-card"
            >
              <div class="card-body">
                <div class="d-flex justify-content-between align-items-start mb-2">
                  <h6 class="mb-0">{{ index + 1 }}. {{ suggestion.title }}</h6>
                  <span class="badge" :class="getImpactBadgeClass(suggestion.impact)">
                    {{ suggestion.impact }} impact
                  </span>
                </div>
                <p class="small text-muted mb-2">{{ suggestion.description }}</p>

                <div class="row g-2">
                  <div class="col-md-6">
                    <div class="diff-box before">
                      <small class="text-muted">Before:</small>
                      <pre class="mb-0">{{ suggestion.before }}</pre>
                    </div>
                  </div>
                  <div class="col-md-6">
                    <div class="diff-box after">
                      <small class="text-success">After:</small>
                      <pre class="mb-0">{{ suggestion.after }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Improvement Tips -->
          <div v-if="refinementResponse.improvement_tips.length > 0" class="mt-3">
            <h6 class="mb-2">
              <i class="fas fa-star"></i> Additional Tips
            </h6>
            <ul class="small mb-0">
              <li v-for="(tip, index) in refinementResponse.improvement_tips" :key="index">
                {{ tip }}
              </li>
            </ul>
          </div>

          <!-- Refined Prompt -->
          <div class="mt-4">
            <div class="d-flex justify-content-between align-items-center mb-2">
              <h6 class="mb-0">
                <i class="fas fa-check-circle text-success"></i> Refined Prompt
              </h6>
              <button
                class="btn btn-sm btn-outline-primary"
                @click="applyRefinedPrompt"
              >
                <i class="fas fa-check"></i> Apply This Version
              </button>
            </div>
            <pre class="refined-prompt">{{ refinementResponse.refined_prompt }}</pre>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else-if="!refining" class="text-center text-muted py-4">
          <i class="fas fa-robot fa-3x mb-3 opacity-50"></i>
          <p class="mb-0">Click "Refine with AI" to get improvement suggestions</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { RefinementResponse } from '@/types/smart-prompt';
import type { Settings } from '@/types';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';

const props = defineProps<{
  prompt: string;
  settings: Settings;
}>();

const emit = defineEmits<{
  (e: 'apply-refined', refinedPrompt: string): void;
}>();

const refining = ref(false);
const refinementResponse = ref<RefinementResponse | null>(null);

async function refine() {
  if (!props.prompt) {
    showToast('Please generate a prompt first', 'warning');
    return;
  }

  try {
    refining.value = true;
    refinementResponse.value = await apiService.refinePrompt(
      props.prompt,
      props.settings
    );
    showToast('Got AI refinement suggestions!', 'success');
  } catch (error) {
    console.error('Refinement error:', error);
    showToast('Failed to refine prompt. Check your LLM settings.', 'danger');
  } finally {
    refining.value = false;
  }
}

function getImpactBadgeClass(impact: string): string {
  switch (impact) {
    case 'high':
      return 'bg-danger';
    case 'medium':
      return 'bg-warning';
    case 'low':
      return 'bg-info';
    default:
      return 'bg-secondary';
  }
}

function applyRefinedPrompt() {
  if (refinementResponse.value) {
    emit('apply-refined', refinementResponse.value.refined_prompt);
    showToast('Applied refined prompt!', 'success');
  }
}

// Expose refine method
defineExpose({ refine });
</script>

<style scoped>
.quality-comparison {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
}

.quality-score {
  font-size: 1.5rem;
  font-weight: bold;
  margin-top: 0.25rem;
}

.suggestion-card {
  border-left: 4px solid #0d6efd;
}

.diff-box {
  background: #f8f9fa;
  padding: 0.75rem;
  border-radius: 4px;
  border: 1px solid #dee2e6;
}

.diff-box.before {
  background: #fff3cd;
  border-color: #ffc107;
}

.diff-box.after {
  background: #d1e7dd;
  border-color: #198754;
}

.diff-box pre {
  font-size: 0.85rem;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.refined-prompt {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 4px;
  border: 1px solid #dee2e6;
  font-size: 0.9rem;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 400px;
  overflow-y: auto;
}
</style>
