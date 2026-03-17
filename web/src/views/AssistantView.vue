<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">{{ pageTitle }}</h2>
      <div class="badge bg-secondary">v0.2</div>
    </div>

    <!-- Module and Tags Display -->
    <div v-if="module || tags" class="mb-3">
      <div class="row g-2">
        <div v-if="module" class="col-auto">
          <span class="badge bg-primary">
            <i class="fas fa-cube"></i> {{ module }}
          </span>
        </div>
        <div v-if="tags" class="col-auto">
          <span v-for="tag in tagList" :key="tag" class="badge bg-info me-1">
            <i class="fas fa-tag"></i> {{ tag }}
          </span>
        </div>
      </div>
    </div>

    <hr class="mb-4" />

    <!-- Tag Cloud for Filtering -->
    <div v-if="promptStore.availableTags.length > 0" class="card mb-3">
      <div class="card-body">
        <div class="d-flex align-items-center gap-3">
          <!-- Search Bar -->
          <div class="d-flex align-items-center gap-2 search-container">
            <input
              v-model="searchQuery"
              type="text"
              class="form-control form-control-sm"
              placeholder="Search prompts..."
              @keyup.enter="handleSearch"
            />
            <button
              @click="handleSearch"
              class="btn btn-sm btn-primary"
              :disabled="!searchQuery"
            >
              <i class="fas fa-search"></i>
            </button>
            <button
              v-if="searchQuery"
              @click="clearSearch"
              class="btn btn-sm btn-outline-secondary"
              title="Clear search"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>

          <!-- Tags List -->
          <div class="tags-horizontal-scroll flex-grow-1">
            <button
              v-for="tagCount in promptStore.availableTags"
              :key="tagCount.tag"
              @click="toggleTag(tagCount.tag)"
              :class="[
                'btn btn-sm',
                promptStore.selectedTags.includes(tagCount.tag)
                  ? 'btn-primary'
                  : 'btn-outline-primary'
              ]"
            >
              {{ tagCount.tag }}
              <span class="badge bg-light text-dark ms-1">{{ tagCount.count }}</span>
            </button>
          </div>

          <!-- Clear Filters Button -->
          <button
            v-if="promptStore.selectedTags.length > 0"
            @click="clearTagFilters"
            class="btn btn-sm btn-outline-secondary text-nowrap"
          >
            <i class="fas fa-times"></i> Clear ({{ promptStore.selectedTags.length }})
          </button>
        </div>

        <div v-if="promptStore.selectedTags.length > 0" class="mt-2">
          <small class="text-muted">
            <i class="fas fa-info-circle"></i>
            Showing {{ promptStore.prompts.length }} prompt{{ promptStore.prompts.length !== 1 ? 's' : '' }}
            matching selected tags
          </small>
        </div>
      </div>
    </div>

    <!-- Settings Status Indicator -->
    <div
      class="alert alert-warning alert-dismissible fade show"
      v-if="!settingsStore.isConfigured"
    >
      <i class="fas fa-exclamation-triangle"></i>
      <strong>Settings Required:</strong> Please configure your API keys and settings before
      using the application.
      <router-link to="/settings" class="alert-link">Go to Settings</router-link>
      <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    </div>

    <div
      class="alert alert-success alert-dismissible fade show"
      v-if="settingsStore.isConfigured"
    >
      <i class="fas fa-check-circle"></i>
      <strong>Settings Configured:</strong> Your API keys and settings are properly configured.
      <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    </div>

    <div class="row g-3">
      <!-- Prompt Configuration Card -->
      <div class="col-md-6">
        <div class="card mb-3">
          <div class="card-header bg-primary text-white d-flex justify-content-between align-items-center">
            <h5 class="mb-0">
              <i class="fas fa-comments"></i> Prompt Configuration
            </h5>
            <router-link
              to="/tools/smart-prompt"
              class="btn btn-sm btn-light"
              title="Generate prompts with AI"
            >
              <i class="fas fa-magic"></i> Smart Generator
            </router-link>
          </div>
          <div class="card-body">
            <div class="alert alert-info alert-dismissible fade show mb-3" role="alert">
              <i class="fas fa-lightbulb"></i>
              <strong>Pro Tip:</strong> Use the
              <router-link to="/tools/smart-prompt" class="alert-link fw-bold">
                Smart Prompt Generator
              </router-link>
              to create high-quality prompts with AI assistance using frameworks like R-CAR, RACE, CRISPE, and more!
              <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
            </div>

            <div class="mb-3">
              <label class="form-label d-flex justify-content-between align-items-center">
                <span>Prompt Name</span>
                <div class="d-flex gap-2 align-items-center">
                  <router-link
                    :to="{ path: '/prompts', query: { edit: formData.prompt.name } }"
                    class="badge bg-secondary text-decoration-none"
                    :title="`Edit prompt: ${formData.prompt.name}`"
                  >
                    <i class="fas fa-edit"></i> Edit
                  </router-link>
                  <router-link
                    to="/tools/smart-prompt"
                    class="badge bg-info text-decoration-none"
                    title="Use AI to generate a better prompt"
                  >
                    <i class="fas fa-magic"></i> Smart Prompt Generator
                  </router-link>
                </div>
              </label>
              <select v-model="formData.prompt.name" class="form-select">
                <option
                  v-for="(prompt, index) in sortedPrompts"
                  :key="prompt.name"
                  :value="prompt.name"
                >
                  {{ index + 1 }}. {{ prompt.name }} — {{ prompt.description }}
                </option>
              </select>
              <small class="text-muted mt-1 d-block">
                Select a pre-configured prompt or create a new one using the Smart Generator
              </small>
            </div>

            <div class="mb-3">
              <label class="form-label">System Prompt</label>
              <textarea
                v-model="formData.prompt.system_prompt"
                rows="2"
                class="form-control"
                placeholder="Enter system prompt instructions..."
              ></textarea>
            </div>

            <div class="mb-3">
              <label class="form-label">User Prompt</label>
              <textarea
                v-model="formData.prompt.user_prompt"
                rows="8"
                class="form-control"
                placeholder="Enter user prompt with {{code}} placeholder..."
              ></textarea>
            </div>
          </div>
        </div>
      </div>

      <!-- Code/Path Configuration Card -->
      <div class="col-md-6">
        <div class="card mb-3">
          <div class="card-header bg-success text-white">
            <h5 class="mb-0">
              <i class="fas fa-code"></i> Code/Path Configuration
            </h5>
          </div>
          <div class="card-body">
            <div class="mb-3">
              <label class="form-label">Project</label>
              <select v-model="formData.project" class="form-select">
                <option value="" disabled>Select a project</option>
                <option
                  v-for="projectName in projectStore.projectNames"
                  :key="projectName"
                  :value="projectName"
                >
                  {{ projectName }}
                </option>
              </select>
            </div>

            <div class="mb-3">
              <label class="form-label">Code Path</label>
              <input
                v-model="formData.codePath"
                type="text"
                class="form-control"
                placeholder="e.g., src/main/java/com/example"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">GitLab Repository</label>
              <input
                v-model="formData.gitlab_code_repo"
                type="text"
                class="form-control"
                placeholder="e.g., group/project"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">GitLab Branch</label>
              <input
                v-model="formData.gitlab_code_branch"
                type="text"
                class="form-control"
                placeholder="e.g., main"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">GitLab Code Path</label>
              <input
                v-model="formData.gitlab_code_path"
                type="text"
                class="form-control"
                placeholder="e.g., src/"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">Merge Request ID (Optional)</label>
              <input
                v-model="formData.gitlab_mr_id"
                type="text"
                class="form-control"
                placeholder="e.g., 123"
              />
              <small class="text-muted">Enter MR ID to enable comment feature</small>
            </div>

            <div class="mb-3">
              <label class="form-label">
                <i class="fas fa-code"></i> Computer Language
              </label>
              <select v-model="formData.computer_language" class="form-select">
                <option value="java">☕ Java</option>
                <option value="python">🐍 Python</option>
                <option value="javascript">🟨 JavaScript</option>
                <option value="typescript">🔷 TypeScript</option>
                <option value="go">🐹 Go</option>
                <option value="rust">🦀 Rust</option>
                <option value="cpp">⚡ C++</option>
                <option value="lua">🌙 Lua</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="card mb-3">
      <div class="card-body">
        <div class="d-flex gap-4 align-items-center flex-wrap">
          <!-- LLM Model Selector -->
          <div class="d-flex align-items-center gap-2">
            <label class="form-label mb-0">
              <i class="fas fa-robot"></i> Model:
            </label>
            <select
              v-model="selectedModelId"
              class="form-select form-select-sm"
              style="width: auto; min-width: 180px;"
              @change="onModelChange"
            >
              <option value="">Legacy Settings</option>
              <option
                v-for="model in llmModelStore.enabledModels"
                :key="model.id"
                :value="model.id"
              >
                {{ model.name }} ({{ model.model }})
              </option>
            </select>
            <router-link
              to="/tools/llm-models"
              class="btn btn-sm btn-outline-secondary"
              title="Manage LLM Models"
            >
              <i class="fas fa-cog"></i>
            </router-link>
          </div>
          <div class="form-check form-switch">
            <input
              class="form-check-input"
              type="checkbox"
              v-model="formData.stream"
              id="streamSwitch"
            />
            <label class="form-check-label" for="streamSwitch">
              <i class="fas fa-stream"></i> Enable Streaming
            </label>
          </div>
          <div class="form-check form-switch">
            <input
              class="form-check-input"
              type="checkbox"
              v-model="formData.remember"
              id="rememberSwitch"
            />
            <label class="form-check-label" for="rememberSwitch">
              <i class="fas fa-brain"></i> Remember Conversation
            </label>
          </div>
          <div class="d-flex gap-3 align-items-center">
            <label class="form-label mb-0">
              <i class="fas fa-language"></i> Output Languages:
            </label>
            <div class="form-check">
              <input
                class="form-check-input"
                type="checkbox"
                v-model="formData.output_languages"
                value="english"
                id="langEnglish"
              />
              <label class="form-check-label" for="langEnglish">
                🇺🇸 English
              </label>
            </div>
            <div class="form-check">
              <input
                class="form-check-input"
                type="checkbox"
                v-model="formData.output_languages"
                value="chinese"
                id="langChinese"
              />
              <label class="form-check-label" for="langChinese">
                🇨🇳 Chinese
              </label>
            </div>
          </div>
        </div>

        <div class="d-flex gap-2 mt-3 flex-wrap justify-content-center">
          <button
            @click="submitRequest"
            class="btn btn-primary"
            :disabled="isLoading"
          >
            <span v-if="isLoading" class="spinner-border spinner-border-sm me-2"></span>
            <i v-else class="fas fa-paper-plane me-1"></i>
            {{ isLoading ? loadingMessage : 'Submit Request' }}
          </button>

        </div>
      </div>
    </div>

    <!-- Answer Display -->
    <div v-if="answer" class="card mb-3">
      <div class="card-header bg-info text-white">
        <h5 class="mb-0">
          <i class="fas fa-robot"></i> Answer
        </h5>
      </div>
      <div class="card-body">
        <div v-html="renderedAnswer" class="markdown-content"></div>
      </div>
    </div>

    <!-- Image Container -->
    <div id="imageContainer" class="mt-3"></div>

    <!-- Bottom Action Buttons (Fixed to bottom) -->
    <div class="card mt-4 mb-5 sticky-bottom-actions">
      <div class="card-body bg-light">
        <div class="d-flex gap-2 justify-content-center flex-wrap">

          <button
            v-if="formData.gitlab_mr_id && answer"
            @click="showAddCommentModal"
            class="btn btn-primary"
            :disabled="!answer"
            title="Add comment to GitLab MR"
          >
            <i class="fas fa-comment-medical me-1"></i>
            Add MR Comment
          </button>

          <button
            @click="saveConversation"
            class="btn btn-success"
            :disabled="isSaving || !answer"
          >
            <i class="fas fa-save me-1"></i>
            Save
          </button>

          <button @click="copyAnswer" class="btn btn-secondary" :disabled="!answer">
            <i class="fas fa-copy me-1"></i>
            Copy
          </button>

          <button @click="drawImage" class="btn btn-info" :disabled="isDrawing || !answer">
            <span v-if="isDrawing" class="spinner-border spinner-border-sm me-2"></span>
            <i v-else class="fas fa-image me-1"></i>
            {{ isDrawing ? 'Generating...' : 'Draw Image' }}
          </button>

          <button @click="saveImages" class="btn btn-warning" :disabled="!hasImages">
            <i class="fas fa-download me-1"></i>
            Save Images
          </button>

          <button @click="clearAll" class="btn btn-danger">
            <i class="fas fa-trash me-1"></i>
            Clear
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Add Comment Modal -->
  <div
    v-if="showCommentModal"
    class="modal fade show d-block"
    tabindex="-1"
    style="background-color: rgba(0, 0, 0, 0.5);"
  >
    <div class="modal-dialog modal-lg">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            <i class="fas fa-comment-medical me-2"></i>
            Add Comment to Merge Request #{{ formData.gitlab_mr_id }}
          </h5>
          <button
            type="button"
            class="btn-close"
            @click="closeCommentModal"
          ></button>
        </div>
        <div class="modal-body">
          <div class="alert alert-info">
            <i class="fas fa-info-circle"></i>
            <strong>Tip:</strong> Select text from the AI response below, then click "Add Comment" to pre-fill this textarea. You can edit the comment before submitting.
          </div>
          <div class="mb-3">
            <label class="form-label fw-bold">Comment (Markdown supported)</label>
            <textarea
              v-model="commentText"
              class="form-control"
              rows="10"
              placeholder="Enter your comment here... You can use Markdown formatting."
            ></textarea>
            <small class="text-muted">
              Selected text will appear here. Edit as needed before posting.
            </small>
          </div>
          <div v-if="commentError" class="alert alert-danger">
            <i class="fas fa-exclamation-triangle"></i>
            {{ commentError }}
          </div>
        </div>
        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-secondary"
            @click="closeCommentModal"
          >
            <i class="fas fa-times me-1"></i>
            Cancel
          </button>
          <button
            type="button"
            class="btn btn-primary"
            @click="submitComment"
            :disabled="isSubmittingComment || !commentText.trim()"
          >
            <span v-if="isSubmittingComment" class="spinner-border spinner-border-sm me-2"></span>
            <i v-else class="fas fa-paper-plane me-1"></i>
            {{ isSubmittingComment ? 'Submitting...' : 'Submit Comment' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { marked } from 'marked';
import hljs from 'highlight.js';
import { useSettingsStore } from '@/stores/settingsStore';
import { usePromptStore } from '@/stores/promptStore';
import { useProjectStore } from '@/stores/projectStore';
import { useLLMModelStore } from '@/stores/llmModelStore';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import { getCurrentDateTime, getRandomString } from '@/utils/dateUtils';
import { downloadFile, copyToClipboard, getFileExtension } from '@/utils/fileUtils';
import type { AgentFormData } from '@/types';

const route = useRoute();
const settingsStore = useSettingsStore();
const promptStore = usePromptStore();
const projectStore = useProjectStore();
const llmModelStore = useLLMModelStore();

// URL parameters
const pageTitle = ref('Lazy AI Coder');
const module = ref('');
const tags = ref('');
const tagList = computed(() => (tags.value ? tags.value.split(',').map((tag) => tag.trim()) : []));

// Search
const searchQuery = ref('');

// Computed property for sorted prompts (server-side search already applied)
const sortedPrompts = computed(() => {
  return [...promptStore.filteredPrompts].sort((a, b) =>
    a.name.localeCompare(b.name)
  );
});

// Form data
const formData = ref<AgentFormData>({
  prompt: {
    name: '',
    system_prompt: '',
    user_prompt: '',
    assistant_prompt: '',
  },
  session_id: '',
  computer_language: 'java',
  output_languages: ['chinese'],
  codePath: '',
  stream: true,
  remember: false,
  project: '',
  gitlab_code_repo: '',
  gitlab_code_path: '',
  gitlab_code_branch: '',
  gitlab_mr_id: '',
});

// State
const answer = ref('');
const isLoading = ref(false);
const isDrawing = ref(false);
const isSaving = ref(false);
const loadingMessage = ref('');
const hasImages = ref(false);

// Comment modal state
const showCommentModal = ref(false);
const commentText = ref('');
const isSubmittingComment = ref(false);
const commentError = ref('');

// LLM Model selector
const selectedModelId = ref('');

function onModelChange() {
  llmModelStore.selectModel(selectedModelId.value || null);
}

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
});

const renderedAnswer = computed(() => {
  const html = marked(answer.value) as string;
  // Apply syntax highlighting to code blocks after rendering
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = html;
  tempDiv.querySelectorAll('pre code').forEach((block) => {
    hljs.highlightElement(block as HTMLElement);
  });
  return tempDiv.innerHTML;
});

// Methods
function getEffectiveSettings() {
  // If a managed model is selected, use its configuration
  const modelConfig = llmModelStore.effectiveConfig;
  if (modelConfig) {
    return {
      ...settingsStore.$state,
      LLM_BASE_URL: modelConfig.baseUrl,
      LLM_MODEL: modelConfig.model,
      LLM_TEMPERATURE: String(modelConfig.temperature),
      // API key still comes from settings store (security: not stored in DB)
    };
  }
  // Fall back to legacy settings
  return settingsStore.$state;
}

async function submitRequest() {
  if (!settingsStore.isConfigured) {
    showToast(
      'Please configure your settings first. Go to Settings page.',
      'warning'
    );
    return;
  }

  try {
    isLoading.value = true;
    loadingMessage.value = formData.value.stream
      ? 'Starting streaming response...'
      : 'Processing your request...';
    answer.value = '';
    formData.value.prompt.assistant_prompt = '';

    if (!formData.value.remember) {
      formData.value.session_id = '';
    } else {
      if (!formData.value.session_id) {
        formData.value.session_id = getRandomString(10);
      }
    }

    const effectiveSettings = getEffectiveSettings();
    const modelName = llmModelStore.selectedModel?.name || 'Legacy Settings';
    showToast(`Request submitted using ${modelName}!`, 'info');

    if (formData.value.stream) {
      await startStreaming();
      return;
    }

    const result = await apiService.executeAgent({
      ...formData.value,
      settings: effectiveSettings,
    });

    answer.value = result.content;
    formData.value.prompt.assistant_prompt = result.content;
    showToast('Request completed successfully!', 'success');
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    answer.value = 'Error: ' + message;
    showToast('Request failed: ' + message, 'danger');
  } finally {
    isLoading.value = false;
    loadingMessage.value = '';
  }
}

async function startStreaming() {
  try {
    isLoading.value = true;
    loadingMessage.value = 'Connecting to AI service...';

    const effectiveSettings = getEffectiveSettings();
    const ws = await apiService.createWebSocketConnection(
      formData.value,
      effectiveSettings
    );
    loadingMessage.value = 'Streaming response...';

    const timeoutId = setTimeout(() => {
      if (isLoading.value) {
        console.warn('Streaming timeout reached');
        isLoading.value = false;
        loadingMessage.value = '';
        showToast('Streaming timed out. Please try again.', 'warning');
        ws.close();
      }
    }, 300000);

    ws.onmessage = (event) => {
      const chunk = event.data;
      answer.value += chunk;
      formData.value.prompt.assistant_prompt += chunk;
    };

    ws.onerror = (err) => {
      console.error('WebSocket Error:', err);
      answer.value += '\n[Error occurred during streaming]';
      showToast('Streaming error occurred', 'danger');
      clearTimeout(timeoutId);
      isLoading.value = false;
      loadingMessage.value = '';
      ws.close();
    };

    ws.onclose = () => {
      console.log('WebSocket closed');
      clearTimeout(timeoutId);
      isLoading.value = false;
      loadingMessage.value = '';
      if (answer.value && !answer.value.includes('[Error occurred')) {
        showToast('Streaming completed successfully!', 'success');
      }
    };
  } catch (error) {
    isLoading.value = false;
    loadingMessage.value = '';
    const message = error instanceof Error ? error.message : 'Unknown error';
    showToast('Failed to start streaming: ' + message, 'danger');
  }
}

async function drawImage() {
  if (!settingsStore.isConfigured) {
    showToast('Please configure your settings first.', 'warning');
    return;
  }

  if (!answer.value) {
    showToast(
      'Please submit a request first to generate content for image creation.',
      'warning'
    );
    return;
  }

  try {
    isDrawing.value = true;
    showToast('Starting image generation...', 'info');

    const response = await fetch('/api/v1/draw', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...formData.value.prompt,
        settings: settingsStore.$state,
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const imageContainer = document.getElementById('imageContainer');
    if (!imageContainer) {
      throw new Error('Image container not found');
    }

    imageContainer.innerHTML = '';

    const result = await response.json();

    if (!result || result.length === 0) {
      showToast(
        'No diagrams found in the response. Make sure your content contains PlantUML or mindmap syntax.',
        'warning'
      );
      return;
    }

    result.forEach((image: any) => {
      const linkElement = document.createElement('a');
      linkElement.href = image.url.replace('/png/', '/uml/');
      linkElement.target = '_blank';
      linkElement.rel = 'noopener noreferrer';
      linkElement.className = 'image-item';
      linkElement.style.display = 'block';
      linkElement.style.marginBottom = '20px';

      const imgElement = document.createElement('img');
      const imageSrc = image.path.replace('./web', '');
      imgElement.src = imageSrc;
      imgElement.alt = image.type;
      imgElement.style.maxWidth = '100%';
      imgElement.style.marginTop = '10px';
      imgElement.style.borderRadius = '8px';
      imgElement.style.boxShadow = '0 4px 8px rgba(0,0,0,0.1)';

      const codeElement = document.createElement('code');
      codeElement.textContent = image.script;
      codeElement.style.display = 'block';
      codeElement.style.marginTop = '10px';
      codeElement.style.fontFamily = 'monospace';
      codeElement.style.whiteSpace = 'pre-wrap';
      codeElement.style.backgroundColor = '#f8f9fa';
      codeElement.style.padding = '15px';
      codeElement.style.borderRadius = '8px';

      linkElement.appendChild(imgElement);
      linkElement.appendChild(codeElement);
      imageContainer.appendChild(linkElement);
    });

    hasImages.value = true;
    showToast(
      `Successfully generated ${result.length} diagram${result.length > 1 ? 's' : ''}!`,
      'success'
    );
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    showToast('Failed to generate images: ' + message, 'danger');
  } finally {
    isDrawing.value = false;
  }
}

async function saveConversation() {
  if (!answer.value) {
    showToast('No content to save. Please submit a request first.', 'warning');
    return;
  }

  try {
    isSaving.value = true;

    const filename =
      prompt('Enter filename:', `conversation-${getCurrentDateTime()}.md`) ||
      `conversation-${getCurrentDateTime()}.md`;

    const markdownContent =
      `# Conversation at ${getCurrentDateTime()}\n\n` +
      `* Module: ${module.value || 'N/A'}\n` +
      `* Tags: ${tags.value || 'N/A'}\n` +
      `* name: ${formData.value.prompt.name}\n` +
      `* system prompt: ${formData.value.prompt.system_prompt}\n` +
      `* user prompt: ${formData.value.prompt.user_prompt}\n` +
      `* computer language: ${formData.value.computer_language}\n` +
      `* output languages: ${formData.value.output_languages.join(', ')}\n` +
      `* project: ${formData.value.project}\n` +
      `\n\n---\n\n` +
      `## Answer:\n\n` +
      `${answer.value}`;

    downloadFile(markdownContent, filename);
    showToast('Conversation saved successfully!', 'success');
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    showToast('Failed to save conversation: ' + message, 'danger');
  } finally {
    isSaving.value = false;
  }
}

async function copyAnswer() {
  if (!answer.value) {
    showToast('No content to copy. Please submit a request first.', 'warning');
    return;
  }

  const markdownContent =
    `# Conversation at ${getCurrentDateTime()}\n\n` +
    `* Module: ${module.value || 'N/A'}\n` +
    `* Tags: ${tags.value || 'N/A'}\n` +
    `* name: ${formData.value.prompt.name}\n` +
    `* system prompt: ${formData.value.prompt.system_prompt}\n` +
    `* user prompt: ${formData.value.prompt.user_prompt}\n` +
    `\n\n---\n\n` +
    `## Answer:\n\n` +
    `${answer.value}`;

  await copyToClipboard(markdownContent, 'Answer copied to clipboard successfully!');
}

async function saveImages() {
  const imageContainer = document.getElementById('imageContainer');
  if (!imageContainer) return;

  const links = imageContainer.querySelectorAll('a');

  if (links.length === 0) {
    showToast(
      'No images to save. Please generate images first using "Draw Image" button.',
      'warning'
    );
    return;
  }

  const timestamp = getCurrentDateTime();
  let savedCount = 0;

  links.forEach((link, index) => {
    const img = link.querySelector('img') as HTMLImageElement;
    if (img && img.src) {
      const a = document.createElement('a');
      a.href = img.src;
      a.download = `${timestamp}-${index + 1}${getFileExtension(img.src)}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      savedCount++;
    }
  });

  if (savedCount > 0) {
    showToast(
      `Successfully saved ${savedCount} image${savedCount > 1 ? 's' : ''}!`,
      'success'
    );
  } else {
    showToast('No valid images found to save.', 'warning');
  }
}

function clearAll() {
  answer.value = '';
  formData.value.prompt.assistant_prompt = '';
  const imageContainer = document.getElementById('imageContainer');
  if (imageContainer) {
    imageContainer.innerHTML = '';
  }
  hasImages.value = false;
  showToast('Cleared all content', 'info');
}

// Tag filtering methods
async function toggleTag(tag: string) {
  promptStore.toggleTag(tag);
  await promptStore.fetchPromptsByTags();

  // Update sorted prompts and set default if needed
  if (sortedPrompts.value.length > 0) {
    const currentPrompt = promptStore.getPromptByName(formData.value.prompt.name);
    if (!currentPrompt) {
      // Current prompt not in filtered results, select first one
      const defaultPrompt = promptStore.getPromptByName('summarize') || sortedPrompts.value[0];
      formData.value.prompt.name = defaultPrompt.name;
      formData.value.prompt.system_prompt = defaultPrompt.system_prompt.replace(/\\n/g, '\n');
      formData.value.prompt.user_prompt = defaultPrompt.user_prompt.replace(/\\n/g, '\n');
    }
  }

  showToast(
    `Filter ${promptStore.selectedTags.includes(tag) ? 'applied' : 'removed'}: ${tag}`,
    'info'
  );
}

async function clearTagFilters() {
  promptStore.clearTags();
  await promptStore.fetchPrompts(module.value || undefined);

  // Reapply client-side tag filters if needed
  if (tagList.value.length > 0) {
    promptStore.filterPrompts(module.value, tagList.value);
  }

  // Reset to default prompt
  if (sortedPrompts.value.length > 0) {
    const defaultPrompt = promptStore.getPromptByName('summarize') || sortedPrompts.value[0];
    formData.value.prompt.name = defaultPrompt.name;
    formData.value.prompt.system_prompt = defaultPrompt.system_prompt.replace(/\\n/g, '\n');
    formData.value.prompt.user_prompt = defaultPrompt.user_prompt.replace(/\\n/g, '\n');
  }

  showToast('Tag filters cleared', 'info');
}

// Search methods
async function handleSearch() {
  if (!searchQuery.value.trim()) {
    return;
  }

  showToast(`Searching for: ${searchQuery.value}`, 'info');

  // Perform server-side search
  await promptStore.fetchPrompts(module.value || undefined, searchQuery.value);

  // Update first matching prompt if results exist
  if (sortedPrompts.value.length > 0) {
    const firstPrompt = sortedPrompts.value[0];
    formData.value.prompt.name = firstPrompt.name;
    formData.value.prompt.system_prompt = firstPrompt.system_prompt.replace(/\\n/g, '\n');
    formData.value.prompt.user_prompt = firstPrompt.user_prompt.replace(/\\n/g, '\n');
  } else {
    showToast('No prompts found matching your search', 'warning');
  }
}

async function clearSearch() {
  searchQuery.value = '';

  // Re-fetch all prompts without search filter
  await promptStore.fetchPrompts(module.value || undefined);

  // Reapply client-side tag filters if needed
  if (tagList.value.length > 0) {
    promptStore.filterPrompts(module.value, tagList.value);
  }

  // Reset to default prompt
  if (sortedPrompts.value.length > 0) {
    const defaultPrompt = promptStore.getPromptByName('summarize') || sortedPrompts.value[0];
    formData.value.prompt.name = defaultPrompt.name;
    formData.value.prompt.system_prompt = defaultPrompt.system_prompt.replace(/\\n/g, '\n');
    formData.value.prompt.user_prompt = defaultPrompt.user_prompt.replace(/\\n/g, '\n');
  }

  showToast('Search cleared', 'info');
}

// Comment modal functions
function showAddCommentModal() {
  // Get selected text if any
  const selectedText = window.getSelection()?.toString() || '';
  
  if (selectedText.trim()) {
    commentText.value = selectedText.trim();
  } else {
    // If no text selected, suggest selecting from answer
    commentText.value = '';
  }
  
  commentError.value = '';
  showCommentModal.value = true;
}

function closeCommentModal() {
  showCommentModal.value = false;
  commentText.value = '';
  commentError.value = '';
}

async function submitComment() {
  if (!commentText.value.trim()) {
    commentError.value = 'Please enter a comment';
    return;
  }

  if (!formData.value.gitlab_mr_id) {
    commentError.value = 'Merge Request ID is required';
    return;
  }

  if (!formData.value.gitlab_code_repo) {
    commentError.value = 'GitLab repository is not configured for this project';
    return;
  }

  isSubmittingComment.value = true;
  commentError.value = '';

  try {
    const response = await fetch('/mcp/v1/call-tool', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: 'post_gitlab_mr_comment',
        arguments: {
          project: formData.value.gitlab_code_repo,
          merge_request_id: formData.value.gitlab_mr_id,
          comment: commentText.value,
        },
        // Pass GitLab credentials from settings
        settings: {
          gitlab_token: settingsStore.GITLAB_TOKEN,
          gitlab_url: settingsStore.GITLAB_BASE_URL,
        },
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const result = await response.json();

    if (result.isError) {
      throw new Error(result.content[0]?.text || 'Failed to post comment');
    }

    showToast('Comment posted successfully to GitLab MR!', 'success');
    closeCommentModal();
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    commentError.value = `Failed to post comment: ${message}`;
    showToast(`Failed to post comment: ${message}`, 'danger');
  } finally {
    isSubmittingComment.value = false;
  }
}

// Watchers
watch(
  () => formData.value.prompt.name,
  (newName) => {
    const selected = promptStore.getPromptByName(newName);
    if (selected) {
      formData.value.prompt.system_prompt = selected.system_prompt.replace(/\\n/g, '\n');
      formData.value.prompt.user_prompt = selected.user_prompt.replace(/\\n/g, '\n');
    }
  }
);

watch(
  () => formData.value.project,
  (newProject) => {
    const config = projectStore.getProject(newProject);
    if (config) {
      // Auto-fill Computer Language
      if (config.language) {
        formData.value.computer_language = config.language;
      }

      // Auto-fill GitLab fields
      formData.value.gitlab_code_repo = config.git_repo || config.gitlab_code_repo || '';
      formData.value.gitlab_code_branch = config.git_branch || config.gitlab_code_branch || '';
      formData.value.gitlab_code_path = config.entry_point || config.gitlab_code_path || '';

      // Note: codePath is NOT auto-filled to allow user flexibility
    }
  }
);

// Watch for route changes to re-fetch prompts when module/tags change
watch(
  () => route.query,
  async (newQuery) => {
    // Update module and tags from URL parameters
    const newModule = (newQuery.module as string) || '';
    const newTags = (newQuery.tags as string) || '';
    const newTitle = (newQuery.title as string) || 'Lazy AI Coder';

    // Only re-fetch if module or tags actually changed
    if (newModule !== module.value || newTags !== tags.value) {
      module.value = newModule;
      tags.value = newTags;
      pageTitle.value = newTitle;

      // Re-fetch prompts from backend with new module/tags
      await promptStore.fetchPrompts(module.value || undefined);

      // Apply client-side filters if additional tags are provided
      if (tagList.value.length > 0) {
        promptStore.filterPrompts(module.value, tagList.value);
      }

      // Set default prompt after fetching new prompts
      if (sortedPrompts.value.length > 0) {
        const defaultPrompt = promptStore.getPromptByName('summarize') || sortedPrompts.value[0];
        formData.value.prompt.name = defaultPrompt.name;
        formData.value.prompt.system_prompt = defaultPrompt.system_prompt.replace(/\\n/g, '\n');
        formData.value.prompt.user_prompt = defaultPrompt.user_prompt.replace(/\\n/g, '\n');
      }
    }
  }
);

// Lifecycle
onMounted(async () => {
  // Parse URL parameters
  pageTitle.value = (route.query.title as string) || 'Lazy AI Coder';
  module.value = (route.query.module as string) || '';
  tags.value = (route.query.tags as string) || '';

  // Load data - pass module as tags to backend API
  await Promise.all([
    promptStore.fetchPrompts(module.value || undefined),
    projectStore.fetchProjects(),
    promptStore.fetchAllPromptsForTags(), // Fetch all prompts for accurate tag counts
    llmModelStore.initialize(), // Initialize LLM models
  ]);

  // Sync selected model ID with store
  if (llmModelStore.selectedModel) {
    selectedModelId.value = llmModelStore.selectedModel.id;
  }

  // Apply client-side filters if additional tags are provided
  if (tagList.value.length > 0) {
    promptStore.filterPrompts(module.value, tagList.value);
  }

  // Set default prompt
  if (sortedPrompts.value.length > 0) {
    const defaultPrompt = promptStore.getPromptByName('summarize') || sortedPrompts.value[0];
    formData.value.prompt.name = defaultPrompt.name;
    formData.value.prompt.system_prompt = defaultPrompt.system_prompt.replace(/\\n/g, '\n');
    formData.value.prompt.user_prompt = defaultPrompt.user_prompt.replace(/\\n/g, '\n');
  }

  // Set default project - prefer default project, then first available
  if (projectStore.projectNames.length > 0) {
    // Check if current project exists in the list
    if (!projectStore.projectNames.includes(formData.value.project)) {
      // Use default project if set, otherwise use first available
      const defaultProj = projectStore.defaultProject;
      formData.value.project = defaultProj?.name || projectStore.projectNames[0];
    }
  }

  // Reload settings when the page gains focus
  window.addEventListener('focus', () => {
    settingsStore.loadFromStorage();
  });
});
</script>

<style scoped>
.markdown-content {
  font-size: 1rem;
  line-height: 1.6;
}

.markdown-content :deep(pre) {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
  padding: 1rem;
  overflow-x: auto;
}

.sticky-bottom-actions {
  position: sticky;
  bottom: 20px;
  z-index: 1000;
  box-shadow: 0 -4px 6px -1px rgba(0, 0, 0, 0.1), 0 -2px 4px -1px rgba(0, 0, 0, 0.06);
}

.sticky-bottom-actions .btn {
  min-width: 120px;
}

.markdown-content :deep(code) {
  background-color: #f8f9fa;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  font-size: 87.5%;
}

.markdown-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
}

.image-item {
  display: block;
  margin-bottom: 20px;
}

.image-item img {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

/* Search container styles */
.search-container {
  min-width: 250px;
}

/* Tag cloud styles */
.tags-horizontal-scroll {
  display: flex;
  gap: 0.5rem;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  padding-bottom: 0.5rem;
  /* Custom scrollbar styling */
  scrollbar-width: thin;
  scrollbar-color: #dee2e6 #f8f9fa;
}

.tags-horizontal-scroll::-webkit-scrollbar {
  height: 8px;
}

.tags-horizontal-scroll::-webkit-scrollbar-track {
  background: #f8f9fa;
  border-radius: 4px;
}

.tags-horizontal-scroll::-webkit-scrollbar-thumb {
  background: #dee2e6;
  border-radius: 4px;
}

.tags-horizontal-scroll::-webkit-scrollbar-thumb:hover {
  background: #adb5bd;
}

.tags-horizontal-scroll .btn {
  flex-shrink: 0;
}

.btn-outline-primary:hover {
  transform: translateY(-2px);
  transition: all 0.2s ease;
}

.btn-primary {
  transform: translateY(-2px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.gap-2 {
  gap: 0.5rem !important;
}

/* Responsive styles */
@media (max-width: 992px) {
  .search-container {
    width: 100%;
    min-width: auto;
  }

  .tags-horizontal-scroll {
    width: 100%;
  }
}

/* Comment Modal Styles */
.modal.show {
  display: block;
}

.modal textarea {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 0.9rem;
}

.modal-dialog {
  margin-top: 3rem;
}

.modal-content {
  border-radius: 0.5rem;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.modal-header {
  background-color: #f8f9fa;
  border-bottom: 2px solid #dee2e6;
}

.modal-footer {
  background-color: #f8f9fa;
  border-top: 2px solid #dee2e6;
}
</style>
