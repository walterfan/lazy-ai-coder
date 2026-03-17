<template>
  <div class="container mt-4 mb-5">
    <h1 class="mb-2">
      <i class="fas fa-wand-magic-sparkles"></i> Smart Prompt Generator
    </h1>
    <p class="text-muted mb-4">
      Create professional ChatGPT prompts using industry-standard frameworks and templates
    </p>

    <!-- Step Indicator -->
    <div class="steps-indicator mb-4">
      <div
        class="step"
        :class="{ active: currentStep === 1, completed: currentStep > 1, clickable: true }"
        @click="goToStep(1)"
        role="button"
        tabindex="0"
        title="Go to Choose Method"
      >
        <div class="step-number">1</div>
        <div class="step-label">Choose Method</div>
      </div>
      <div class="step-line" :class="{ active: currentStep > 1 }"></div>
      <div
        class="step"
        :class="{
          active: currentStep === 2,
          completed: currentStep > 2,
          clickable: canGoToStep(2),
          disabled: !canGoToStep(2)
        }"
        @click="goToStepIfAllowed(2)"
        role="button"
        tabindex="0"
        :title="canGoToStep(2) ? 'Go to Build Prompt' : 'Select a framework first'"
      >
        <div class="step-number">2</div>
        <div class="step-label">Build Prompt</div>
      </div>
      <div class="step-line" :class="{ active: currentStep > 2 }"></div>
      <div
        class="step"
        :class="{
          active: currentStep === 3,
          completed: currentStep > 3,
          clickable: canGoToStep(3),
          disabled: !canGoToStep(3)
        }"
        @click="goToStepIfAllowed(3)"
        role="button"
        tabindex="0"
        :title="canGoToStep(3) ? 'Go to Review, Refine & Try' : 'Generate a prompt first'"
      >
        <div class="step-number">3</div>
        <div class="step-label">Review, Refine & Try</div>
      </div>
    </div>

    <!-- Step 1: Choose Framework or Template -->
    <div v-if="currentStep === 1" class="step-content">
      <div class="card mb-4">
        <div class="card-body">
          <div class="text-center mb-4">
            <h4>How would you like to start?</h4>
            <p class="text-muted">Choose a framework for structured guidance or pick a pre-built template</p>
          </div>

          <div class="row g-3 mb-4">
            <div class="col-md-4">
              <button
                class="method-card"
                :class="{ active: startMethod === 'framework' }"
                @click="startMethod = 'framework'"
              >
                <i class="fas fa-layer-group fa-2x mb-2 text-primary"></i>
                <h6 class="mb-2">Choose Framework</h6>
                <p class="text-muted small mb-0">
                  Select from CRISPE, RISEN, CO-STAR, APE, R-CAR, or RACE frameworks
                </p>
              </button>
            </div>
            <div class="col-md-4">
              <button
                class="method-card"
                :class="{ active: startMethod === 'template' }"
                @click="startMethod = 'template'"
              >
                <i class="fas fa-th-large fa-2x mb-2 text-success"></i>
                <h6 class="mb-2">Use Template</h6>
                <p class="text-muted small mb-0">
                  Start with a pre-built template for common coding tasks
                </p>
              </button>
            </div>
            <div class="col-md-4">
              <button
                class="method-card"
                :class="{ active: startMethod === 'scratch' }"
                @click="handleStartFromScratch"
              >
                <i class="fas fa-pen-to-square fa-2x mb-2 text-warning"></i>
                <h6 class="mb-2">Start from Scratch</h6>
                <p class="text-muted small mb-0">
                  Jump directly to step 3 and manually write your prompts
                </p>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Framework Selection -->
      <FrameworkSelector
        v-if="startMethod === 'framework'"
        :selected-framework-id="selectedFrameworkId"
        @framework-selected="handleFrameworkSelected"
      />

      <!-- Template Selection -->
      <TemplateSelector
        v-if="startMethod === 'template'"
        @template-selected="handleTemplateSelected"
        @skip="handleTemplateSkip"
      />

      <!-- Next Button -->
      <div class="text-end mt-4" v-if="selectedFrameworkId">
        <button class="btn btn-primary btn-lg" @click="goToStep(2)">
          <i class="fas fa-arrow-right"></i> Continue to Build Prompt
        </button>
      </div>
    </div>

    <!-- Step 2: Build Prompt -->
    <div v-if="currentStep === 2 && selectedFramework" class="step-content">
      <div class="row">
        <div class="col-lg-8">
          <div class="card mb-4">
            <div class="card-body">
              <PromptBuilder
                :framework="selectedFramework"
                :initial-values="templateValues"
                @change-framework="goToStep(1)"
                @generate="handleGenerate"
                @quick-tips="handleQuickTips"
              />
            </div>
          </div>
        </div>
        <div class="col-lg-4">
          <!-- Quick Tips Sidebar -->
          <div class="card mb-3" v-if="quickTips.length > 0">
            <div class="card-header bg-warning text-dark">
              <h6 class="mb-0">
                <i class="fas fa-lightbulb"></i> Quick Tips
              </h6>
            </div>
            <div class="card-body">
              <ul class="small mb-0">
                <li v-for="(tip, index) in quickTips" :key="index" class="mb-2">
                  {{ tip }}
                </li>
              </ul>
            </div>
          </div>

          <!-- Framework Info -->
          <div class="card">
            <div class="card-header bg-info text-white">
              <h6 class="mb-0">
                <i class="fas fa-info-circle"></i> About {{ selectedFramework.name }}
              </h6>
            </div>
            <div class="card-body">
              <p class="small mb-2">{{ selectedFramework.description }}</p>
              <hr>
              <p class="small mb-0">
                <strong>Best for:</strong><br>
                {{ selectedFramework.best_for }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 3: Review, Refine & Try -->
    <div v-if="currentStep === 3" class="step-content">
      <!-- Manual Mode Info Banner -->
      <div v-if="isManualMode" class="alert alert-info mb-4">
        <div class="d-flex align-items-center">
          <i class="fas fa-pen-to-square fa-2x me-3"></i>
          <div>
            <h6 class="mb-1">
              <strong>Manual Mode</strong>
            </h6>
            <p class="mb-0">
              You're writing prompts from scratch. Fill in the system and user prompts below,
              then test them with the LLM. You can refine and retry as many times as needed.
            </p>
          </div>
          <button
            class="btn btn-outline-primary btn-sm ms-auto"
            @click="goToStep(1)"
            title="Return to step 1"
          >
            <i class="fas fa-arrow-left"></i> Back
          </button>
        </div>
      </div>

      <!-- System Prompt (Editable) -->
      <div class="card mb-3">
        <div class="card-header bg-primary text-white">
          <div class="d-flex justify-content-between align-items-center">
            <h5 class="mb-0">
              <i class="fas fa-robot"></i> System Prompt
              <span class="badge bg-light text-dark ms-2">
                <i class="fas fa-edit"></i> Editable
              </span>
            </h5>
            <button
              class="btn btn-sm btn-light"
              @click="copySystemPrompt"
            >
              <i class="fas fa-copy"></i> Copy
            </button>
          </div>
        </div>
        <div class="card-body">
          <p class="text-muted small mb-2">
            <i class="fas fa-info-circle"></i>
            This defines the role, constraints, and behavior for the AI. You can edit it directly.
          </p>
          <textarea
            v-model="systemPrompt"
            class="form-control generated-prompt-textarea"
            rows="5"
            placeholder="Enter system prompt..."
          ></textarea>
        </div>
      </div>

      <!-- User Prompt (Editable) -->
      <div class="card mb-3">
        <div class="card-header bg-success text-white">
          <div class="d-flex justify-content-between align-items-center">
            <h5 class="mb-0">
              <i class="fas fa-user"></i> User Prompt
              <span class="badge bg-light text-dark ms-2">
                <i class="fas fa-edit"></i> Editable
              </span>
            </h5>
            <button
              class="btn btn-sm btn-light"
              @click="copyUserPrompt"
            >
              <i class="fas fa-copy"></i> Copy
            </button>
          </div>
        </div>
        <div class="card-body">
          <p class="text-muted small mb-2">
            <i class="fas fa-info-circle"></i>
            This is your actual request with context and requirements. You can edit it directly.
          </p>
          <textarea
            v-model="userPrompt"
            class="form-control generated-prompt-textarea"
            rows="10"
            placeholder="Enter user prompt..."
          ></textarea>
        </div>
      </div>

      <!-- Actions -->
      <div class="d-flex gap-2 justify-content-between mb-4">
        <button class="btn btn-outline-secondary" @click="startOver">
          <i class="fas fa-redo"></i> Start Over
        </button>
        <div class="d-flex gap-2">
          <button
            v-if="!isManualMode && selectedFramework"
            class="btn btn-outline-primary"
            @click="goToStep(2)"
          >
            <i class="fas fa-arrow-left"></i> Back to Edit
          </button>
          <button class="btn btn-success" @click="showSaveDialog = true">
            <i class="fas fa-save"></i> Save
          </button>
          <button
            v-if="settingsStore.LLM_API_KEY"
            class="btn btn-warning"
            @click="showRefineModal = true"
            :disabled="!systemPrompt && !userPrompt"
          >
            <i class="fas fa-wand-magic-sparkles"></i> Refine
          </button>
          <button
            v-if="settingsStore.LLM_API_KEY"
            class="btn btn-info"
            @click="showRefinement = !showRefinement"
          >
            <i class="fas fa-magic"></i> {{ showRefinement ? 'Hide' : 'AI Analysis' }}
          </button>
          <button
            class="btn btn-primary"
            @click="submitPrompt"
            :disabled="isSubmitting || !settingsStore.LLM_API_KEY"
          >
            <i class="fas fa-play"></i>
            {{ isSubmitting ? 'Trying...' : 'Try' }}
          </button>
        </div>
      </div>

      <!-- Save Dialog -->
      <div v-if="showSaveDialog" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">Save Prompt</h5>
              <button type="button" class="btn-close" @click="closeSaveDialog"></button>
            </div>
            <div class="modal-body">
              <form @submit.prevent="savePrompt">
                <div class="mb-3">
                  <label for="promptName" class="form-label">Name <span class="text-danger">*</span></label>
                  <input
                    type="text"
                    class="form-control"
                    id="promptName"
                    v-model="saveForm.name"
                    placeholder="e.g., Python API Developer Prompt"
                    required
                  />
                </div>
                <div class="mb-3">
                  <label for="promptDescription" class="form-label">Description</label>
                  <textarea
                    class="form-control"
                    id="promptDescription"
                    v-model="saveForm.description"
                    rows="3"
                    placeholder="Describe what this prompt does..."
                  ></textarea>
                </div>
                <div class="mb-3">
                  <label for="promptTags" class="form-label">Tags</label>
                  <input
                    type="text"
                    class="form-control"
                    id="promptTags"
                    v-model="saveForm.tags"
                    placeholder="e.g., python, api, development (comma-separated)"
                  />
                  <div class="form-text">Separate tags with commas</div>
                </div>
              </form>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeSaveDialog">Cancel</button>
              <button
                type="button"
                class="btn btn-success"
                @click="savePrompt"
                :disabled="!saveForm.name || isSaving"
              >
                <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="fas fa-save"></i>
                {{ isSaving ? 'Saving...' : 'Save Prompt' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Template Placeholder Modal -->
      <div v-if="showPlaceholderModal" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-lg">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">
                <i class="fas fa-edit"></i> Fill Template Placeholders
              </h5>
              <button type="button" class="btn-close" @click="closePlaceholderModal"></button>
            </div>
            <div class="modal-body">
              <p class="text-muted mb-4">
                <i class="fas fa-info-circle"></i>
                This template contains placeholders. Please provide values for each placeholder below.
                The AI will then intelligently fill in the framework fields based on your input.
              </p>
              <form @submit.prevent="applyPlaceholders">
                <div v-for="placeholder in templatePlaceholders" :key="placeholder.name" class="mb-3">
                  <label :for="`placeholder-${placeholder.name}`" class="form-label">
                    <code v-text="`{{${placeholder.name}}}`"></code>
                    <span class="text-danger">*</span>
                  </label>
                  <input
                    :id="`placeholder-${placeholder.name}`"
                    type="text"
                    class="form-control"
                    v-model="placeholder.value"
                    :placeholder="`Enter value for ${placeholder.name}...`"
                    required
                  />
                  <small class="text-muted">
                    Example: {{ getPlaceholderExample(placeholder.name) }}
                  </small>
                </div>
              </form>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closePlaceholderModal">Cancel</button>
              <button
                type="button"
                class="btn btn-primary"
                @click="applyPlaceholders"
                :disabled="!allPlaceholdersFilled || fillingPlaceholders"
              >
                <span v-if="fillingPlaceholders" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="fas fa-magic"></i>
                {{ fillingPlaceholders ? 'Processing...' : 'Auto-fill with AI' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Refine Modal -->
      <div v-if="showRefineModal" class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-lg">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">
                <i class="fas fa-wand-magic-sparkles"></i> Refine Prompt with AI
              </h5>
              <button type="button" class="btn-close" @click="closeRefineModal"></button>
            </div>
            <div class="modal-body">
              <p class="text-muted mb-3">
                <i class="fas fa-info-circle"></i>
                Describe what improvements you want to make to your prompt. The AI will analyze your requirements and refine the prompt accordingly.
              </p>
              <form @submit.prevent="refinePromptWithRequirements">
                <div class="mb-3">
                  <label for="refineRequirements" class="form-label">
                    Refinement Requirements <span class="text-danger">*</span>
                  </label>
                  <textarea
                    id="refineRequirements"
                    class="form-control"
                    v-model="refineRequirements"
                    rows="6"
                    placeholder="Example:&#10;- Make the tone more professional&#10;- Add error handling instructions&#10;- Include examples in the output format&#10;- Make it more concise"
                    required
                  ></textarea>
                  <small class="text-muted">
                    Be specific about what you want to improve or change in the prompt.
                  </small>
                </div>
              </form>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeRefineModal">Cancel</button>
              <button
                type="button"
                class="btn btn-warning"
                @click="refinePromptWithRequirements"
                :disabled="!refineRequirements.trim() || isRefining"
              >
                <span v-if="isRefining" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="fas fa-wand-magic-sparkles"></i>
                {{ isRefining ? 'Refining...' : 'Refine Prompt' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Refinement Panel (Collapsible) -->
      <RefinementPanel
        v-if="showRefinement && settingsStore.LLM_API_KEY"
        :prompt="generatedPrompt"
        :settings="settings"
        @apply-refined="applyRefinedPrompt"
        ref="refinementPanel"
        class="mb-4"
      />

      <!-- Conversation History -->
      <div v-if="conversationHistory.length > 0 || isSubmitting" class="card mb-3">
        <div class="card-header bg-info text-white">
          <div class="d-flex justify-content-between align-items-center">
            <h5 class="mb-0">
              <i class="fas fa-comments"></i> Conversation History
            </h5>
            <button
              v-if="conversationHistory.length > 0"
              class="btn btn-sm btn-light"
              @click="clearConversation"
              title="Clear conversation and start fresh"
            >
              <i class="fas fa-trash"></i> Clear
            </button>
          </div>
        </div>
        <div class="card-body conversation-container">
          <!-- Conversation turns -->
          <div
            v-for="(turn, index) in conversationHistory"
            :key="index"
            class="conversation-turn mb-3"
          >
            <!-- User message -->
            <div class="message user-message">
              <div class="message-header">
                <i class="fas fa-user"></i>
                <strong>You</strong>
                <span class="text-muted small ms-2">{{ turn.timestamp }}</span>
              </div>
              <div class="message-content">{{ turn.userMessage }}</div>
            </div>

            <!-- Assistant response -->
            <div class="message assistant-message">
              <div class="message-header">
                <i class="fas fa-robot"></i>
                <strong>Assistant</strong>
              </div>
              <div class="message-content">{{ turn.assistantResponse }}</div>
            </div>
          </div>

          <!-- Loading state for current request -->
          <div v-if="isSubmitting" class="message assistant-message">
            <div class="message-header">
              <i class="fas fa-robot"></i>
              <strong>Assistant</strong>
            </div>
            <div class="message-content">
              <div class="spinner-border spinner-border-sm text-primary me-2" role="status">
                <span class="visually-hidden">Loading...</span>
              </div>
              Thinking...
            </div>
          </div>

          <!-- Current streaming response -->
          <div v-else-if="llmResponse && !isSubmitting" class="message assistant-message">
            <div class="message-header">
              <i class="fas fa-robot"></i>
              <strong>Assistant</strong>
            </div>
            <div class="message-content">{{ llmResponse }}</div>
          </div>
        </div>
      </div>

      <!-- Follow-up Question Input -->
      <div v-if="conversationHistory.length > 0" class="card mb-3">
        <div class="card-header bg-light">
          <h6 class="mb-0">
            <i class="fas fa-comment-dots"></i> Continue Conversation
          </h6>
        </div>
        <div class="card-body">
          <div class="mb-3">
            <label for="followUpInput" class="form-label">Ask a follow-up question:</label>
            <textarea
              id="followUpInput"
              v-model="followUpMessage"
              class="form-control"
              rows="3"
              placeholder="Type your follow-up question here..."
              :disabled="isSubmitting"
              @keydown.ctrl.enter="submitFollowUp"
            ></textarea>
            <small class="text-muted">Press Ctrl+Enter to send</small>
          </div>
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <button
                class="btn btn-outline-secondary btn-sm"
                @click="copyConversation"
              >
                <i class="fas fa-copy"></i> Copy Conversation
              </button>
            </div>
            <button
              class="btn btn-primary"
              @click="submitFollowUp"
              :disabled="isSubmitting || !followUpMessage.trim()"
            >
              <i class="fas fa-paper-plane"></i>
              {{ isSubmitting ? 'Sending...' : 'Send' }}
            </button>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import { useSettingsStore } from '@/stores/settingsStore';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import type { Framework, PromptTemplate, GenerateFromFrameworkResponse } from '@/types/smart-prompt';

// Components
import FrameworkSelector from '@/components/prompt-generator/FrameworkSelector.vue';
import TemplateSelector from '@/components/prompt-generator/TemplateSelector.vue';
import PromptBuilder from '@/components/prompt-generator/PromptBuilder.vue';
import RefinementPanel from '@/components/prompt-generator/RefinementPanel.vue';

const route = useRoute();
const settingsStore = useSettingsStore();

// State
const currentStep = ref(1);
const startMethod = ref<'framework' | 'template' | 'scratch'>('framework');
const selectedFrameworkId = ref<string>('');
const selectedFramework = ref<Framework | null>(null);
const templateValues = ref<Record<string, string>>({});
const systemPrompt = ref(`你是一位专业的 {{角色}}, 专注于 {{具体领域}}

## 核心能力
- 能力1
- 能力2
- 能力3

## 工作原则
1. 专业性：始终保持专业的态度和语言
2. 准确性：确保提供的信息准确可靠
3. 安全性：注意保护用户隐私和数据安全
4. 友好性：使用亲切但不过分熟络的语气

## 行为规范
- 当不确定时，主动询问用户
- 当遇到敏感信息时，提醒用户注意安全
- 当超出能力范围时，明确告知限制
- 当需要更多信息时，有条理地提出问题

## 输出格式
1. 回答结构清晰, 层次分明
2. 重要信息使用加粗或列表突出
3. 专业术语配有通俗的解释
4. 代码示例包含完整注释

## 工具使用
- 允许使用：{{工具列表}}
- 使用限制：{{限制说明}}
- 异常处理：{{处理方法}}

{{具体任务要求和个性化设置}}`);
const userPrompt = ref('');
const generatedPrompt = ref(''); // Full prompt for backward compatibility
const isManualMode = ref(false); // Track if user started from scratch
const qualityScore = ref<{ score: number; max_score: number; feedback?: string[]; suggestions?: string[] } | null>(null);
const quickTips = ref<string[]>([]);
const refinementPanel = ref();
const showRefinement = ref(false);

// Template placeholder state
const showPlaceholderModal = ref(false);
const selectedTemplate = ref<PromptTemplate | null>(null);
const templatePlaceholders = ref<Array<{ name: string; value: string }>>([]);
const fillingPlaceholders = ref(false);

// Save state
const showSaveDialog = ref(false);
const isSaving = ref(false);
const saveForm = ref({
  name: '',
  description: '',
  tags: ''
});

// Refine state
const showRefineModal = ref(false);
const isRefining = ref(false);
const refineRequirements = ref('');

// Try state
const isSubmitting = ref(false);
const llmResponse = ref('');
const followUpMessage = ref('');
const conversationHistory = ref<Array<{
  userMessage: string;
  assistantResponse: string;
  timestamp: string;
}>>([]);
let websocket: WebSocket | null = null;

const settings = computed(() => ({
  LLM_API_KEY: settingsStore.LLM_API_KEY,
  LLM_MODEL: settingsStore.LLM_MODEL,
  LLM_BASE_URL: settingsStore.LLM_BASE_URL,
  LLM_TEMPERATURE: settingsStore.LLM_TEMPERATURE || '0.7',
  GITLAB_BASE_URL: settingsStore.GITLAB_BASE_URL,
  GITLAB_TOKEN: settingsStore.GITLAB_TOKEN,
}));

onMounted(() => {
  // Load settings from localStorage
  settingsStore.loadFromStorage();

  // Check if route has query params for pre-filling prompt content
  if (route.query.systemPrompt || route.query.userPrompt) {
    // Enable manual mode
    isManualMode.value = true;
    startMethod.value = 'scratch';

    // Pre-fill prompts from query params
    systemPrompt.value = (route.query.systemPrompt as string) || '';
    userPrompt.value = (route.query.userPrompt as string) || '';
    generatedPrompt.value = 'manual'; // Set to non-empty to satisfy step progression

    // Go to step 3 (Review, Refine & Try)
    currentStep.value = 3;

    // Show toast to inform user
    showToast('Prompt loaded for refinement', 'success');
  }
});

// Auto-scroll to bottom of conversation when new messages arrive
watch(
  () => conversationHistory.value.length,
  async () => {
    await nextTick();
    const container = document.querySelector('.conversation-container');
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  }
);

// Auto-scroll when streaming response updates
watch(
  () => llmResponse.value,
  async () => {
    await nextTick();
    const container = document.querySelector('.conversation-container');
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  }
);

function canGoToStep(step: number): boolean {
  if (step === 1) return true; // Can always go to step 1
  if (step === 2) return !!selectedFrameworkId.value; // Need framework for step 2
  if (step === 3) return !!generatedPrompt.value || isManualMode.value; // Need generated prompt or manual mode for step 3
  return false;
}

function goToStep(step: number) {
  currentStep.value = step;

  // Reset manual mode when returning to step 1
  if (step === 1) {
    isManualMode.value = false;
    if (startMethod.value === 'scratch') {
      startMethod.value = 'framework';
    }
  }

  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function goToStepIfAllowed(step: number) {
  if (canGoToStep(step)) {
    goToStep(step);
  } else {
    // Provide feedback about why they can't go to this step
    if (step === 2 && !selectedFrameworkId.value) {
      showToast('Please select a framework first', 'warning');
    } else if (step === 3 && !generatedPrompt.value) {
      showToast('Please generate a prompt first', 'warning');
    }
  }
}

async function handleFrameworkSelected(frameworkId: string) {
  selectedFrameworkId.value = frameworkId;
  try {
    selectedFramework.value = await apiService.getFramework(frameworkId);
  } catch (error) {
    console.error('Failed to load framework:', error);
    showToast('Failed to load framework details', 'danger');
  }
}

async function handleTemplateSelected(template: PromptTemplate) {
  // Set framework first
  await handleFrameworkSelected(template.framework);

  // Extract all template field values as text for placeholder detection
  const templateContent = Object.values(template.fields).join(' ');

  // Check if template has placeholders (pattern: {{placeholder_name}})
  const placeholderPattern = /\{\{([^}]+)\}\}/g;
  const placeholders = new Set<string>();
  let match;

  while ((match = placeholderPattern.exec(templateContent)) !== null) {
    placeholders.add(match[1].trim());
  }

  // If template has placeholders, show modal to collect values
  if (placeholders.size > 0) {
    selectedTemplate.value = template;
    templatePlaceholders.value = Array.from(placeholders).map(name => ({
      name,
      value: ''
    }));
    showPlaceholderModal.value = true;
  } else {
    // No placeholders, just load template normally
    templateValues.value = template.fields;
    startMethod.value = 'framework';
    showToast(`Loaded template: ${template.name}`, 'success');
    goToStep(2);
  }
}

function handleTemplateSkip() {
  startMethod.value = 'framework';
}

// Computed property to check if all placeholders are filled
const allPlaceholdersFilled = computed(() => {
  return templatePlaceholders.value.every(p => p.value.trim() !== '');
});

// Get example value for a placeholder based on its name
function getPlaceholderExample(name: string): string {
  const examples: Record<string, string> = {
    'project_name': 'E-commerce Platform',
    'task_description': 'Implement user authentication',
    'language': 'Python',
    'framework': 'FastAPI',
    'feature': 'Shopping Cart',
    'api_endpoint': '/api/v1/users',
    'database': 'PostgreSQL',
    'technology': 'React'
  };

  return examples[name] || 'Your value here';
}

// Close placeholder modal
function closePlaceholderModal() {
  showPlaceholderModal.value = false;
  templatePlaceholders.value = [];
  selectedTemplate.value = null;
}

// Apply placeholders and auto-fill with AI
async function applyPlaceholders() {
  if (!selectedTemplate.value || !allPlaceholdersFilled.value) {
    return;
  }

  if (!settingsStore.LLM_API_KEY) {
    showToast('Please configure LLM API key in Settings first', 'danger');
    return;
  }

  try {
    fillingPlaceholders.value = true;

    // Replace placeholders in template fields
    const filledFields: Record<string, string> = {};
    for (const [key, value] of Object.entries(selectedTemplate.value.fields)) {
      let filledValue = value;
      templatePlaceholders.value.forEach(placeholder => {
        const pattern = new RegExp(`\\{\\{\\s*${placeholder.name}\\s*\\}\\}`, 'g');
        filledValue = filledValue.replace(pattern, placeholder.value);
      });
      filledFields[key] = filledValue;
    }

    // Build user input for auto-fill from filled template content
    const userInputText = Object.values(filledFields).filter(v => v.trim()).join('\n');

    // Prepare settings for API call
    const settings = {
      LLM_API_KEY: settingsStore.LLM_API_KEY,
      LLM_MODEL: settingsStore.LLM_MODEL || 'gpt-4',
      LLM_BASE_URL: settingsStore.LLM_BASE_URL || 'https://api.openai.com/v1',
      LLM_TEMPERATURE: settingsStore.LLM_TEMPERATURE || '0.7',
      GITLAB_BASE_URL: settingsStore.GITLAB_BASE_URL,
      GITLAB_TOKEN: settingsStore.GITLAB_TOKEN,
    };

    // Call auto-fill API with the filled template content
    const response = await apiService.autoFillFields(
      selectedTemplate.value.framework,
      userInputText,
      settings
    );

    // Update field values with AI-generated content
    templateValues.value = response.fields;

    // Close modal and proceed to step 2
    showPlaceholderModal.value = false;
    startMethod.value = 'framework';
    showToast(`Template loaded and auto-filled successfully!`, 'success');
    goToStep(2);

  } catch (error) {
    console.error('Failed to apply placeholders:', error);
    showToast('Failed to auto-fill from template. Please try again.', 'danger');
  } finally {
    fillingPlaceholders.value = false;
  }
}

function handleStartFromScratch() {
  // Enable manual mode
  isManualMode.value = true;
  startMethod.value = 'scratch';

  // Initialize with empty prompts
  systemPrompt.value = '';
  userPrompt.value = '';
  generatedPrompt.value = 'manual'; // Set to non-empty to satisfy step progression

  // Reset other state
  selectedFrameworkId.value = '';
  selectedFramework.value = null;
  templateValues.value = {};
  qualityScore.value = null;
  conversationHistory.value = [];
  llmResponse.value = '';

  showToast('Manual mode enabled - write your prompts directly!', 'success');
  goToStep(3);
}

async function handleGenerate(fields: Record<string, string>) {
  if (!selectedFrameworkId.value) return;

  try {
    const response: GenerateFromFrameworkResponse = await apiService.generateFromFramework(
      selectedFrameworkId.value,
      fields
    );

    // Store all three prompt variations
    systemPrompt.value = response.system_prompt;
    userPrompt.value = response.user_prompt;
    generatedPrompt.value = response.full_prompt; // For backward compatibility

    qualityScore.value = {
      score: response.quality_score,
      max_score: response.max_score,
    };

    showToast('Prompt generated successfully!', 'success');
    goToStep(3);
  } catch (error) {
    console.error('Generation error:', error);
    showToast('Failed to generate prompt', 'danger');
  }
}

function handleQuickTips(tips: string[]) {
  quickTips.value = tips;
}

async function copySystemPrompt() {
  try {
    await navigator.clipboard.writeText(systemPrompt.value);
    showToast('System prompt copied to clipboard!', 'success');
  } catch (error) {
    console.error('Copy error:', error);
    showToast('Failed to copy to clipboard', 'danger');
  }
}

async function copyUserPrompt() {
  try {
    await navigator.clipboard.writeText(userPrompt.value);
    showToast('User prompt copied to clipboard!', 'success');
  } catch (error) {
    console.error('Copy error:', error);
    showToast('Failed to copy to clipboard', 'danger');
  }
}

function applyRefinedPrompt(refinedPrompt: string) {
  generatedPrompt.value = refinedPrompt;
  showToast('Applied refined prompt!', 'success');
}

function closeSaveDialog() {
  showSaveDialog.value = false;
  saveForm.value = {
    name: '',
    description: '',
    tags: ''
  };
}

function closeRefineModal() {
  showRefineModal.value = false;
  refineRequirements.value = '';
}

async function refinePromptWithRequirements() {
  if (!refineRequirements.value.trim()) {
    showToast('Please enter your refinement requirements', 'warning');
    return;
  }

  try {
    isRefining.value = true;

    // Prepare the refinement request
    const refineRequest = {
      system_prompt: systemPrompt.value,
      user_prompt: userPrompt.value,
      requirements: refineRequirements.value.trim(),
      settings: {
        LLM_API_KEY: settingsStore.LLM_API_KEY,
        LLM_MODEL: settingsStore.LLM_MODEL,
        LLM_BASE_URL: settingsStore.LLM_BASE_URL,
        LLM_TEMPERATURE: settingsStore.LLM_TEMPERATURE || '0.7',
      }
    };

    // Call API to refine prompt
    const response = await apiService.refinePromptWithRequirements(refineRequest);

    // Update prompts with refined versions
    if (response.system_prompt) {
      systemPrompt.value = response.system_prompt;
    }
    if (response.user_prompt) {
      userPrompt.value = response.user_prompt;
    }

    showToast('Prompt refined successfully!', 'success');
    closeRefineModal();
  } catch (error) {
    console.error('Failed to refine prompt:', error);
    const message = error instanceof Error ? error.message : 'Unknown error';
    showToast('Failed to refine prompt: ' + message, 'danger');
  } finally {
    isRefining.value = false;
  }
}

async function savePrompt() {
  if (!saveForm.value.name.trim()) {
    showToast('Please enter a prompt name', 'warning');
    return;
  }

  try {
    isSaving.value = true;

    const promptData = {
      name: saveForm.value.name.trim(),
      description: saveForm.value.description.trim(),
      system_prompt: systemPrompt.value,
      user_prompt: userPrompt.value,
      tags: saveForm.value.tags.trim()
    };

    await apiService.createPrompt(promptData);

    showToast('Prompt saved successfully!', 'success');
    closeSaveDialog();
  } catch (error) {
    console.error('Save prompt error:', error);
    showToast('Failed to save prompt. Please try again.', 'danger');
  } finally {
    isSaving.value = false;
  }
}

function startOver() {
  if (confirm('Start over? This will clear your current prompt and conversation.')) {
    currentStep.value = 1;
    selectedFrameworkId.value = '';
    selectedFramework.value = null;
    templateValues.value = {};
    systemPrompt.value = '';
    userPrompt.value = '';
    generatedPrompt.value = '';
    qualityScore.value = null;
    quickTips.value = [];
    showRefinement.value = false;
    showSaveDialog.value = false;
    saveForm.value = { name: '', description: '', tags: '' };
    llmResponse.value = '';
    followUpMessage.value = '';
    conversationHistory.value = [];
    closeWebSocket();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
}

async function submitPrompt() {
  if (!settingsStore.LLM_API_KEY) {
    showToast('Please configure LLM API key in Settings', 'danger');
    return;
  }

  isSubmitting.value = true;
  llmResponse.value = '';
  const currentUserMessage = userPrompt.value;

  try {
    // Connect to WebSocket
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = window.location.host;
    const wsUrl = `${wsProtocol}//${wsHost}/api/v1/stream`;

    websocket = new WebSocket(wsUrl);

    websocket.onopen = () => {
      console.log('WebSocket connected');

      // Build conversation messages for context
      const messages: Array<{ role: string; content: string }> = [];

      // Add conversation history
      conversationHistory.value.forEach(turn => {
        messages.push({
          role: 'user',
          content: turn.userMessage
        });
        messages.push({
          role: 'assistant',
          content: turn.assistantResponse
        });
      });

      // Send the request in the format expected by the backend
      const message = {
        settings: {
          LLM_API_KEY: settingsStore.LLM_API_KEY || '',
          LLM_MODEL: settingsStore.LLM_MODEL || 'gpt-4',
          LLM_BASE_URL: settingsStore.LLM_BASE_URL || 'https://api.openai.com/v1',
          LLM_TEMPERATURE: settingsStore.LLM_TEMPERATURE || '0.7',
          GITLAB_BASE_URL: settingsStore.GITLAB_BASE_URL || '',
          GITLAB_TOKEN: settingsStore.GITLAB_TOKEN || ''
        },
        prompt: {
          name: 'smart-prompt-test',
          system_prompt: systemPrompt.value,
          user_prompt: currentUserMessage,
          assistant_prompt: messages.length > 0 ? JSON.stringify(messages) : ''
        },
        stream: true,
        remember: false,
        session_id: `smart-prompt-${Date.now()}`
      };
      websocket?.send(JSON.stringify(message));
    };

    websocket.onmessage = (event) => {
      const data = event.data;
      // Append to response
      llmResponse.value += data;
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
      showToast('Connection error. Please try again.', 'danger');
      isSubmitting.value = false;
    };

    websocket.onclose = () => {
      console.log('WebSocket closed');
      isSubmitting.value = false;

      // Save to conversation history
      if (llmResponse.value) {
        conversationHistory.value.push({
          userMessage: currentUserMessage,
          assistantResponse: llmResponse.value,
          timestamp: new Date().toLocaleTimeString()
        });
        llmResponse.value = ''; // Clear for next turn
      }
    };
  } catch (error) {
    console.error('Submit error:', error);
    showToast('Failed to submit prompt', 'danger');
    isSubmitting.value = false;
  }
}

async function submitFollowUp() {
  if (!followUpMessage.value.trim()) {
    showToast('Please enter a follow-up question', 'warning');
    return;
  }

  // Set the user prompt to the follow-up message
  const originalUserPrompt = userPrompt.value;
  userPrompt.value = followUpMessage.value.trim();
  followUpMessage.value = '';

  // Submit the prompt
  await submitPrompt();

  // Restore original user prompt after submission
  setTimeout(() => {
    userPrompt.value = originalUserPrompt;
  }, 100);
}

function clearConversation() {
  if (confirm('Clear conversation history? This cannot be undone.')) {
    conversationHistory.value = [];
    llmResponse.value = '';
    showToast('Conversation cleared', 'info');
  }
}

async function copyConversation() {
  let text = `System Prompt:\n${systemPrompt.value}\n\n`;
  text += `Initial User Prompt:\n${userPrompt.value}\n\n`;
  text += '=== Conversation ===\n\n';

  conversationHistory.value.forEach((turn) => {
    text += `[${turn.timestamp}] You: ${turn.userMessage}\n\n`;
    text += `Assistant: ${turn.assistantResponse}\n\n`;
    text += '---\n\n';
  });

  try {
    await navigator.clipboard.writeText(text);
    showToast('Conversation copied to clipboard!', 'success');
  } catch (error) {
    console.error('Copy error:', error);
    showToast('Failed to copy to clipboard', 'danger');
  }
}

function closeWebSocket() {
  if (websocket) {
    websocket.close();
    websocket = null;
  }
}
</script>

<style scoped>
.steps-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 0;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}

.step.clickable {
  cursor: pointer;
}

.step.disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.step-number {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: #e9ecef;
  color: #6c757d;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 1.25rem;
  margin-bottom: 0.5rem;
  transition: all 0.3s;
}

.step.clickable:hover:not(.disabled) .step-number {
  transform: scale(1.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.step.clickable:hover:not(.disabled):not(.active) .step-number {
  background: #d0d5dd;
}

.step.active .step-number {
  background: #0d6efd;
  color: white;
  box-shadow: 0 0 0 4px rgba(13, 110, 253, 0.2);
}

.step.completed .step-number {
  background: #198754;
  color: white;
}

.step.completed:hover:not(.active):not(.disabled) .step-number {
  background: #157347;
}

.step-label {
  font-size: 0.875rem;
  color: #6c757d;
  font-weight: 500;
  transition: color 0.3s;
}

.step.active .step-label {
  color: #0d6efd;
  font-weight: 600;
}

.step.clickable:hover:not(.disabled) .step-label {
  color: #0d6efd;
}

.step-line {
  width: 100px;
  height: 2px;
  background: #e9ecef;
  margin: 0 1rem;
  margin-bottom: 2rem;
}

.step-line.active {
  background: #0d6efd;
}

.method-card {
  width: 100%;
  padding: 1rem;
  border: 2px solid #dee2e6;
  border-radius: 8px;
  background: white;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
}

.method-card:hover {
  border-color: #0d6efd;
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.method-card.active {
  border-color: #0d6efd;
  background: #f8f9ff;
  box-shadow: 0 2px 8px rgba(13, 110, 253, 0.15);
}

.method-card h6 {
  margin-bottom: 0.25rem;
  font-weight: 600;
}

.generated-prompt {
  background: #f8f9fa;
  padding: 1.5rem;
  border-radius: 8px;
  font-size: 0.95rem;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 500px;
  overflow-y: auto;
  line-height: 1.6;
}

.step-content {
  animation: fadeIn 0.3s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .steps-indicator {
    flex-wrap: wrap;
  }

  .step-line {
    width: 50px;
    margin: 0 0.5rem;
  }
}

.response-content {
  max-height: 500px;
  overflow-y: auto;
}

.response-text {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
  font-size: 0.95rem;
  white-space: pre-wrap;
  word-wrap: break-word;
  line-height: 1.6;
}

/* Editable Prompt Textareas */
.generated-prompt-textarea {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9rem;
  line-height: 1.6;
  resize: vertical;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
}

.generated-prompt-textarea:focus {
  background: #ffffff;
  border-color: #86b7fe;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}

/* Conversation Styles */
.conversation-container {
  max-height: 600px;
  overflow-y: auto;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.conversation-turn {
  margin-bottom: 1.5rem;
}

.message {
  margin-bottom: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.user-message {
  background: #e3f2fd;
  border-left: 4px solid #2196f3;
  margin-left: 2rem;
}

.assistant-message {
  background: #f1f8e9;
  border-left: 4px solid #4caf50;
  margin-right: 2rem;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #495057;
}

.message-header i {
  font-size: 1.1rem;
}

.user-message .message-header {
  color: #1976d2;
}

.assistant-message .message-header {
  color: #388e3c;
}

.message-content {
  color: #212529;
  white-space: pre-wrap;
  word-wrap: break-word;
  line-height: 1.6;
  font-size: 0.95rem;
}

/* Scrollbar styling for conversation */
.conversation-container::-webkit-scrollbar {
  width: 8px;
}

.conversation-container::-webkit-scrollbar-track {
  background: #e9ecef;
  border-radius: 4px;
}

.conversation-container::-webkit-scrollbar-thumb {
  background: #adb5bd;
  border-radius: 4px;
}

.conversation-container::-webkit-scrollbar-thumb:hover {
  background: #6c757d;
}
</style>
