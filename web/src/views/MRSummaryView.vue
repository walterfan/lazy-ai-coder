<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-code-branch"></i> Merge Request Summary
      </h2>
      <div class="badge bg-secondary">Beta</div>
    </div>

    <div class="alert alert-info">
      <i class="fas fa-info-circle"></i>
      <strong>MR Summary:</strong> View a detailed summary of all file changes in a GitLab merge request with statistics. 
      Select a project from your configured projects, then enter the merge request ID.
    </div>

    <!-- Loading State -->
    <div v-if="projectsLoading" class="text-center py-4">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading projects...</span>
      </div>
      <p class="mt-2 text-muted">Loading projects...</p>
    </div>

    <!-- No Projects Warning (only show after loading completes) -->
    <div v-else-if="projectStore.projectNames.length === 0" class="alert alert-warning">
      <i class="fas fa-exclamation-triangle"></i>
      <strong>No Projects Configured:</strong> Please configure at least one project with GitLab repository information.
      <router-link to="/projects" class="alert-link ms-2">
        <i class="fas fa-arrow-right"></i> Go to Projects
      </router-link>
    </div>

    <!-- Input Form (only show after projects are loaded) -->
    <div v-if="!projectsLoading && projectStore.projectNames.length > 0" class="card mb-4">
      <div class="card-header bg-primary text-white">
        <h5 class="mb-0">
          <i class="fas fa-cog"></i> Configuration
        </h5>
      </div>
      <div class="card-body">
        <div class="row g-3">
          <div class="col-md-6">
            <label class="form-label">Project</label>
            <select v-model="formData.projectName" class="form-select" @change="onProjectChange">
              <option value="" disabled>Select a project</option>
              <option
                v-for="projectName in projectStore.projectNames"
                :key="projectName"
                :value="projectName"
              >
                {{ projectName }}
              </option>
            </select>
            <small class="text-muted">Select a project configured in Projects page</small>
          </div>
          <div class="col-md-6">
            <label class="form-label">Merge Request ID</label>
            <input
              v-model="formData.mergeRequestId"
              type="text"
              class="form-control"
              placeholder="123"
              required
            />
            <small class="text-muted">Enter the MR number from GitLab</small>
          </div>
          <div class="col-md-6">
            <label class="form-label">GitLab Repository</label>
            <input
              v-model="formData.gitlabRepo"
              type="text"
              class="form-control"
              placeholder="e.g., group/project"
              readonly
            />
            <small class="text-muted">Auto-filled from selected project</small>
          </div>
          <div class="col-md-6">
            <label class="form-label">Output Format</label>
            <select v-model="formData.format" class="form-select">
              <option value="markdown">Markdown Table</option>
              <option value="json">JSON</option>
            </select>
          </div>
        </div>

        <div class="mt-3">
          <button
            @click="fetchMRSummary"
            class="btn btn-primary"
            :disabled="isLoading || !formData.projectName || !formData.mergeRequestId"
          >
            <span v-if="isLoading" class="spinner-border spinner-border-sm me-2"></span>
            <i v-else class="fas fa-search me-1"></i>
            {{ isLoading ? 'Fetching...' : 'Fetch MR Summary' }}
          </button>
          <button
            @click="clearResults"
            class="btn btn-secondary ms-2"
            :disabled="!summary"
          >
            <i class="fas fa-trash me-1"></i>
            Clear
          </button>
        </div>
      </div>
    </div>

    <!-- Results Display -->
    <div v-if="summary" class="card mb-4">
      <div class="card-header bg-success text-white d-flex justify-content-between align-items-center">
        <h5 class="mb-0">
          <i class="fas fa-list"></i> Summary Results
        </h5>
        <button @click="copyToClipboard" class="btn btn-sm btn-light">
          <i class="fas fa-copy me-1"></i>
          Copy
        </button>
      </div>
      <div class="card-body">
        <!-- MR Header Info -->
        <div class="mb-3">
          <h4>{{ summary.title }}</h4>
          <p v-if="summary.description" class="text-muted">{{ summary.description }}</p>
          <div class="badge bg-info">Total Files: {{ summary.total_files }}</div>
        </div>

        <!-- Markdown Table View -->
        <div v-if="formData.format === 'markdown'" class="table-responsive">
          <table class="table table-striped table-hover">
            <thead class="table-dark">
              <tr>
                <th>#</th>
                <th>File Path</th>
                <th>Added</th>
                <th>Deleted</th>
                <th>Renamed</th>
                <th>Changed</th>
                <th>Changed Lines</th>
                <th>Added Lines</th>
                <th>Deleted Lines</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="file in summary.files" :key="file.index">
                <td>{{ file.index }}</td>
                <td>
                  <div v-if="file.old_path === file.new_path">
                    <code class="file-path">{{ file.new_path }}</code>
                  </div>
                  <div v-else>
                    <div v-if="file.old_path" class="text-danger mb-1">
                      <i class="fas fa-minus-circle me-1"></i>
                      <code class="file-path">{{ file.old_path }}</code>
                    </div>
                    <div v-if="file.new_path" class="text-success">
                      <i class="fas fa-plus-circle me-1"></i>
                      <code class="file-path">{{ file.new_path }}</code>
                    </div>
                  </div>
                </td>
                <td>
                  <span v-if="file.added" class="badge bg-success">Y</span>
                  <span v-else class="badge bg-secondary">N</span>
                </td>
                <td>
                  <span v-if="file.deleted" class="badge bg-danger">Y</span>
                  <span v-else class="badge bg-secondary">N</span>
                </td>
                <td>
                  <span v-if="file.renamed" class="badge bg-warning">Y</span>
                  <span v-else class="badge bg-secondary">N</span>
                </td>
                <td>
                  <span v-if="file.changed" class="badge bg-info">Y</span>
                  <span v-else class="badge bg-secondary">N</span>
                </td>
                <td>{{ file.changed_lines }}</td>
                <td class="text-success fw-bold">+{{ file.added_lines }}</td>
                <td class="text-danger fw-bold">-{{ file.deleted_lines }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- JSON View -->
        <div v-else class="json-view">
          <pre><code>{{ JSON.stringify(summary, null, 2) }}</code></pre>
        </div>

        <!-- Statistics Summary -->
        <div class="mt-4 p-3 bg-light rounded">
          <h5>Total Statistics</h5>
          <div class="row">
            <div class="col-md-3">
              <div class="stat-box">
                <div class="stat-value text-success">+{{ totalAddedLines }}</div>
                <div class="stat-label">Lines Added</div>
              </div>
            </div>
            <div class="col-md-3">
              <div class="stat-box">
                <div class="stat-value text-danger">-{{ totalDeletedLines }}</div>
                <div class="stat-label">Lines Deleted</div>
              </div>
            </div>
            <div class="col-md-3">
              <div class="stat-box">
                <div class="stat-value text-info">{{ totalChangedLines }}</div>
                <div class="stat-label">Total Changes</div>
              </div>
            </div>
            <div class="col-md-3">
              <div class="stat-box">
                <div class="stat-value text-primary">{{ summary.total_files }}</div>
                <div class="stat-label">Files Modified</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="alert alert-danger">
      <i class="fas fa-exclamation-triangle"></i>
      <strong>Error:</strong> {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { showToast } from '@/utils/toast';
import { useProjectStore } from '@/stores/projectStore';
import { useSettingsStore } from '@/stores/settingsStore';
import { useInitialLoad } from '@/composables/useInitialLoad';

interface FileSummary {
  index: number;
  old_path: string;
  new_path: string;
  added: boolean;
  deleted: boolean;
  renamed: boolean;
  changed: boolean;
  changed_lines: number;
  added_lines: number;
  deleted_lines: number;
}

interface MRSummary {
  title: string;
  description: string;
  files: FileSummary[];
  total_files: number;
}

const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const { isInitialLoading: projectsLoading, runInitialLoad } = useInitialLoad();

// Ensure settings are loaded from localStorage
settingsStore.loadFromStorage();

const formData = ref({
  projectName: '',
  gitlabURL: '',
  gitlabRepo: '',
  mergeRequestId: '',
  format: 'markdown' as 'markdown' | 'json',
});

const summary = ref<MRSummary | null>(null);
const isLoading = ref(false);
const error = ref('');

const totalAddedLines = computed(() => {
  if (!summary.value) return 0;
  return summary.value.files.reduce((sum, file) => sum + file.added_lines, 0);
});

const totalDeletedLines = computed(() => {
  if (!summary.value) return 0;
  return summary.value.files.reduce((sum, file) => sum + file.deleted_lines, 0);
});

const totalChangedLines = computed(() => {
  if (!summary.value) return 0;
  return summary.value.files.reduce((sum, file) => sum + file.changed_lines, 0);
});

function onProjectChange() {
  const project = projectStore.getProject(formData.value.projectName);
  if (project) {
    // Auto-fill GitLab fields from project configuration
    formData.value.gitlabRepo = project.git_repo || project.gitlab_code_repo || '';
    
    // Note: GitLab URL is typically set via environment variable GITLAB_BASE_URL
    // If project has a custom gitlab URL, we could add it to the project model
  }
}

async function fetchMRSummary() {
  if (!formData.value.projectName || !formData.value.mergeRequestId) {
    showToast('Please select a project and provide merge request ID', 'warning');
    return;
  }

  if (!formData.value.gitlabRepo) {
    showToast('Selected project does not have GitLab repository configured', 'warning');
    return;
  }

  isLoading.value = true;
  error.value = '';
  summary.value = null;

  try {
    const response = await fetch('/mcp/v1/call-tool', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: 'get_gitlab_mr_summary',
        arguments: {
          project: formData.value.gitlabRepo,
          merge_request_id: formData.value.mergeRequestId,
          format: 'json', // Always fetch JSON for parsing
        },
        // Pass GitLab credentials from settings
        settings: {
          gitlab_token: settingsStore.GITLAB_TOKEN,
          gitlab_url: settingsStore.GITLAB_BASE_URL || formData.value.gitlabURL,
        },
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const result = await response.json();

    if (result.isError) {
      throw new Error(result.content[0]?.text || 'Unknown error');
    }

    // Parse the JSON response from the tool
    const textContent = result.content[0]?.text || '{}';
    summary.value = JSON.parse(textContent);

    showToast('MR summary fetched successfully!', 'success');
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred';
    showToast(`Failed to fetch MR summary: ${error.value}`, 'danger');
  } finally {
    isLoading.value = false;
  }
}

function clearResults() {
  summary.value = null;
  error.value = '';
  showToast('Results cleared', 'info');
}

async function copyToClipboard() {
  if (!summary.value) return;

  let text = '';
  if (formData.value.format === 'json') {
    text = JSON.stringify(summary.value, null, 2);
  } else {
    // Generate markdown table
    text = `# Merge Request Summary\n\n`;
    text += `**Title:** ${summary.value.title}\n\n`;
    if (summary.value.description) {
      text += `**Description:** ${summary.value.description}\n\n`;
    }
    text += `**Total Files:** ${summary.value.total_files}\n\n`;
    text += `| # | File Path | Added | Deleted | Renamed | Changed | Changed Lines | Added Lines | Deleted Lines |\n`;
    text += `|---|-----------|-------|---------|---------|---------|---------------|-------------|---------------|\n`;
    
    summary.value.files.forEach(file => {
      // Display file path (with rename indicator if renamed)
      let filePath = '';
      if (file.old_path === file.new_path) {
        filePath = file.new_path;
      } else {
        if (file.old_path && file.new_path) {
          filePath = `~~${file.old_path}~~ → ${file.new_path}`;
        } else if (file.new_path) {
          filePath = file.new_path;
        } else {
          filePath = `~~${file.old_path}~~`;
        }
      }
      
      text += `| ${file.index} | ${filePath} | `;
      text += `${file.added ? 'Y' : 'N'} | ${file.deleted ? 'Y' : 'N'} | `;
      text += `${file.renamed ? 'Y' : 'N'} | ${file.changed ? 'Y' : 'N'} | `;
      text += `${file.changed_lines} | ${file.added_lines} | ${file.deleted_lines} |\n`;
    });
  }

  try {
    await navigator.clipboard.writeText(text);
    showToast('Copied to clipboard!', 'success');
  } catch (err) {
    showToast('Failed to copy to clipboard', 'danger');
  }
}

// Lifecycle - use composable for initial load
runInitialLoad(async () => {
  await projectStore.fetchProjects();

  // Set default project if available
  if (projectStore.projectNames.length > 0) {
    if (!projectStore.projectNames.includes(formData.value.projectName)) {
      // Use default project if set, otherwise first available
      const defaultProj = projectStore.defaultProject;
      formData.value.projectName = defaultProj?.name || projectStore.projectNames[0];
      onProjectChange(); // Auto-fill repo info
    }
  }
});
</script>

<style scoped>
.json-view pre {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
  padding: 1rem;
  overflow-x: auto;
}

.json-view code {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}

.table code {
  background-color: #f8f9fa;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

.table code.file-path {
  display: inline-block;
  max-width: 100%;
  word-break: break-all;
}

.table td > div {
  min-width: 200px;
}

.stat-box {
  text-align: center;
  padding: 1rem;
}

.stat-value {
  font-size: 2rem;
  font-weight: bold;
  line-height: 1;
}

.stat-label {
  font-size: 0.875rem;
  color: #6c757d;
  margin-top: 0.5rem;
}

.table-responsive {
  max-height: 600px;
  overflow-y: auto;
}

.table th {
  position: sticky;
  top: 0;
  background-color: #343a40;
  z-index: 10;
}
</style>

