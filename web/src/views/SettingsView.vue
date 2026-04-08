<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-cog"></i> Settings
      </h2>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-primary text-white d-flex justify-content-between align-items-center">
        <h5 class="mb-0">
          <i class="fas fa-robot"></i> LLM Configuration
        </h5>
        <router-link to="/tools/llm-models" class="btn btn-sm btn-light">
          <i class="fas fa-cogs"></i> Manage Multiple Models
        </router-link>
      </div>
      <div class="card-body">
        <div class="alert alert-info mb-3">
          <i class="fas fa-info-circle"></i>
          <strong>Legacy Settings:</strong> These settings serve as the default fallback when no managed LLM model is selected.
          For multi-model support, visit <router-link to="/tools/llm-models">LLM Model Management</router-link>.
        </div>
        <div class="mb-3">
          <label for="apiKey" class="form-label">API Key</label>
          <input
            v-model="settings.LLM_API_KEY"
            type="password"
            class="form-control"
            id="apiKey"
            placeholder="Enter your LLM API key"
          />
          <small class="text-muted">Your API key for the LLM service</small>
        </div>

        <div class="mb-3">
          <label for="model" class="form-label">Model</label>
          <input
            v-model="settings.LLM_MODEL"
            type="text"
            class="form-control"
            id="model"
            placeholder="e.g., gpt-4, claude-3-opus-20240229"
          />
          <small class="text-muted">The model to use for completions</small>
        </div>

        <div class="mb-3">
          <label for="baseUrl" class="form-label">Base URL</label>
          <input
            v-model="settings.LLM_BASE_URL"
            type="text"
            class="form-control"
            id="baseUrl"
            placeholder="e.g., https://api.openai.com/v1"
          />
          <small class="text-muted">The base URL for the LLM API</small>
        </div>

        <div class="mb-3">
          <label for="temperature" class="form-label">Temperature</label>
          <input
            v-model="settings.LLM_TEMPERATURE"
            type="number"
            step="0.1"
            min="0"
            max="2"
            class="form-control"
            id="temperature"
            placeholder="e.g., 0.7"
          />
          <small class="text-muted"
            >Controls randomness (0.0 = deterministic, 2.0 = very random)</small
          >
        </div>

        <button @click="testLLM" class="btn btn-outline-primary" :disabled="testingLLM">
          <i :class="testingLLM ? 'fas fa-circle-notch fa-spin' : 'fas fa-plug'" class="me-1"></i>
          {{ testingLLM ? 'Testing...' : 'Test LLM Connection' }}
        </button>
        <div v-if="llmResult" class="mt-2">
          <div :class="`alert alert-${llmResult.status === 'healthy' ? 'success' : 'danger'} mb-0`">
            <i :class="llmResult.status === 'healthy' ? 'fas fa-check-circle' : 'fas fa-times-circle'" class="me-1"></i>
            <strong>{{ llmResult.message }}</strong>
            <div v-if="llmResult.details" class="mt-1 small">
              <span v-if="llmResult.details.model_count">Models available: {{ llmResult.details.model_count }}</span>
              <div v-if="llmResult.details.available_models" class="mt-1">
                <span class="badge bg-secondary me-1" v-for="m in llmResult.details.available_models.slice(0, 5)" :key="m">{{ m }}</span>
                <span v-if="llmResult.details.available_models.length > 5" class="text-muted">+{{ llmResult.details.available_models.length - 5 }} more</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-success text-white">
        <h5 class="mb-0">
          <i class="fab fa-gitlab"></i> GitLab Configuration
        </h5>
      </div>
      <div class="card-body">
        <div class="mb-3">
          <label for="gitlabBaseUrl" class="form-label">GitLab Base URL</label>
          <input
            v-model="settings.GITLAB_BASE_URL"
            type="text"
            class="form-control"
            id="gitlabBaseUrl"
            placeholder="e.g., https://gitlab.com"
          />
          <small class="text-muted">The base URL for your GitLab instance</small>
        </div>

        <div class="mb-3" v-if="!authStore.isOAuthMode">
          <label for="gitlabToken" class="form-label">GitLab Personal Access Token</label>
          <input
            v-model="settings.GITLAB_TOKEN"
            type="password"
            class="form-control"
            id="gitlabToken"
            placeholder="glpat-xxxxxxxxxxxx"
          />
          <small class="text-muted">
            Create token at: <strong>GitLab → Settings → Access Tokens</strong><br />
            Required scopes:
            <code>api</code>,
            <code>read_api</code>,
            <code>read_user</code>,
            <code>read_repository</code>,
            <code>write_repository</code>
          </small>
        </div>

        <button @click="testGitLab" class="btn btn-outline-success" :disabled="testingGitLab">
          <i :class="testingGitLab ? 'fas fa-circle-notch fa-spin' : 'fas fa-plug'" class="me-1"></i>
          {{ testingGitLab ? 'Testing...' : 'Test GitLab Connection' }}
        </button>
        <div v-if="gitlabResult" class="mt-2">
          <div :class="`alert alert-${gitlabResult.status === 'healthy' ? 'success' : 'danger'} mb-0`">
            <i :class="gitlabResult.status === 'healthy' ? 'fas fa-check-circle' : 'fas fa-times-circle'" class="me-1"></i>
            <strong>{{ gitlabResult.message }}</strong>
            <div v-if="gitlabResult.details && gitlabResult.details.username" class="mt-1 small">
              Authenticated as: <strong>{{ gitlabResult.details.name }}</strong> (@{{ gitlabResult.details.username }})
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="d-flex gap-2">
      <button @click="saveSettings" class="btn btn-primary">
        <i class="fas fa-save"></i> Save Settings
      </button>
      <button @click="clearSettings" class="btn btn-danger">
        <i class="fas fa-trash"></i> Clear All
      </button>
    </div>

    <div v-if="settingsStore.isConfigured" class="alert alert-success mt-4">
      <i class="fas fa-check-circle"></i>
      <strong>Settings Configured:</strong> Your settings are properly configured.
    </div>

    <div v-else class="alert alert-warning mt-4">
      <i class="fas fa-exclamation-triangle"></i>
      <strong>Settings Incomplete:</strong> Please fill in all required fields.
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useSettingsStore } from '@/stores/settingsStore';
import { useAuthStore } from '@/stores/authStore';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import type { Settings } from '@/types';

const settingsStore = useSettingsStore();
const authStore = useAuthStore();

authStore.loadFromStorage();
settingsStore.loadFromStorage();

const settings = ref<Settings>({ ...settingsStore.$state });

const testingLLM = ref(false);
const testingGitLab = ref(false);
const llmResult = ref<any>(null);
const gitlabResult = ref<any>(null);

function saveSettings() {
  settingsStore.updateSettings(settings.value);

  if ((settings.value.GITLAB_TOKEN || settings.value.LLM_API_KEY) && !authStore.isAuthenticated) {
    authStore.switchToGuestMode();
  }

  showToast('Settings saved successfully!', 'success');
}

async function testLLM() {
  if (!settings.value.LLM_BASE_URL || !settings.value.LLM_API_KEY) {
    showToast('Please fill in LLM Base URL and API Key first.', 'warning');
    return;
  }
  testingLLM.value = true;
  llmResult.value = null;
  try {
    llmResult.value = await apiService.testLLMConnection(
      settings.value.LLM_BASE_URL,
      settings.value.LLM_API_KEY,
    );
  } catch (error: any) {
    llmResult.value = {
      status: 'unhealthy',
      message: error?.response?.data?.message || error?.message || 'Connection test failed',
    };
  } finally {
    testingLLM.value = false;
  }
}

async function testGitLab() {
  if (!settings.value.GITLAB_BASE_URL || !settings.value.GITLAB_TOKEN) {
    showToast('Please fill in GitLab Base URL and Token first.', 'warning');
    return;
  }
  testingGitLab.value = true;
  gitlabResult.value = null;
  try {
    gitlabResult.value = await apiService.testGitLabConnection(
      settings.value.GITLAB_BASE_URL,
      settings.value.GITLAB_TOKEN,
    );
  } catch (error: any) {
    gitlabResult.value = {
      status: 'unhealthy',
      message: error?.response?.data?.message || error?.message || 'Connection test failed',
    };
  } finally {
    testingGitLab.value = false;
  }
}

function clearSettings() {
  if (confirm('Are you sure you want to clear all settings?')) {
    settingsStore.clearSettings();
    settings.value = { ...settingsStore.$state };
    llmResult.value = null;
    gitlabResult.value = null;
    showToast('All settings cleared!', 'info');
  }
}

onMounted(() => {
  settings.value = { ...settingsStore.$state };
});
</script>

<style scoped>
.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
}

.form-control:focus {
  border-color: #0d6efd;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}

.alert {
  border-radius: 0.375rem;
}
</style>
