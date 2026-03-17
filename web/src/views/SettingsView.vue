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
      </div>
    </div>

    <div class="d-flex gap-2">
      <button @click="saveSettings" class="btn btn-primary">
        <i class="fas fa-save"></i> Save Settings
      </button>
      <button @click="testConnection" class="btn btn-secondary">
        <i class="fas fa-plug"></i> Test Connection
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
import { showToast } from '@/utils/toast';
import type { Settings } from '@/types';

const settingsStore = useSettingsStore();
const authStore = useAuthStore();

// Load settings synchronously before initial render
authStore.loadFromStorage();
settingsStore.loadFromStorage();

const settings = ref<Settings>({ ...settingsStore.$state });

function saveSettings() {
  settingsStore.updateSettings(settings.value);

  // If tokens entered and not authenticated, user is in guest mode
  if ((settings.value.GITLAB_TOKEN || settings.value.LLM_API_KEY) && !authStore.isAuthenticated) {
    authStore.switchToGuestMode();
  }

  showToast('Settings saved successfully!', 'success');
}

function testConnection() {
  if (!settingsStore.isConfigured) {
    showToast('Please configure all settings first.', 'warning');
    return;
  }
  showToast('Testing connection... (to be implemented)', 'info');
}

function clearSettings() {
  if (confirm('Are you sure you want to clear all settings?')) {
    settingsStore.clearSettings();
    settings.value = { ...settingsStore.$state };
    showToast('All settings cleared!', 'info');
  }
}

// onMounted is no longer needed for initial load since we load synchronously above
// But keep it for any future async initialization needs
onMounted(() => {
  // Refresh in case storage changed externally
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
