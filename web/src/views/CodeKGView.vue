<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-project-diagram"></i> Code Knowledge Base
      </h2>
    </div>

    <!-- Repository Input Section -->
    <div class="card mb-4">
      <div class="card-header bg-primary text-white">
        <h5 class="mb-0"><i class="fas fa-database me-2"></i>Index a Code Repository</h5>
      </div>
      <div class="card-body">
        <form @submit.prevent="handleSubmitRepo">
          <div class="row g-3 align-items-end">
            <div class="col-md-5">
              <label for="repoInput" class="form-label fw-bold">Repository Path or Git URL</label>
              <div class="input-group">
                <span class="input-group-text">
                  <i class="fas" :class="isGitUrl ? 'fa-globe' : 'fa-folder'"></i>
                </span>
                <input
                  type="text"
                  class="form-control"
                  id="repoInput"
                  v-model="repoInput"
                  placeholder="e.g., /path/to/project or https://github.com/user/repo.git"
                  :disabled="isIndexing"
                />
              </div>
              <div class="form-text">
                Enter a local file path or a Git clone URL
              </div>
            </div>
            <div class="col-md-2">
              <label for="repoBranch" class="form-label">Branch</label>
              <input
                type="text"
                class="form-control"
                id="repoBranch"
                v-model="repoBranch"
                placeholder="main"
                :disabled="isIndexing"
              />
            </div>
            <div class="col-md-3">
              <label for="repoName" class="form-label">Name</label>
              <input
                type="text"
                class="form-control"
                id="repoName"
                v-model="repoName"
                placeholder="Auto-detected from path"
                :disabled="isIndexing"
              />
            </div>
            <div class="col-md-2">
              <button
                type="submit"
                class="btn btn-primary w-100"
                :disabled="!repoInput.trim() || isIndexing"
              >
                <span v-if="isIndexing" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="fas fa-cogs me-1"></i>
                {{ isIndexing ? 'Indexing...' : 'Build KB' }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <!-- Indexing Progress -->
    <div v-if="syncStatus && syncStatus.status !== 'idle'" class="card mb-4 border-info">
      <div class="card-body">
        <div class="d-flex justify-content-between align-items-center mb-2">
          <h6 class="mb-0">
            <i class="fas fa-sync-alt me-2" :class="{ 'fa-spin': syncStatus.status === 'running' }"></i>
            Indexing Progress
          </h6>
          <span class="badge" :class="statusBadgeClass">{{ syncStatus.status }}</span>
        </div>
        <div class="progress mb-2" style="height: 20px;">
          <div
            class="progress-bar progress-bar-striped"
            :class="{ 'progress-bar-animated': syncStatus.status === 'running', 'bg-success': syncStatus.status === 'completed', 'bg-danger': syncStatus.status === 'failed' }"
            role="progressbar"
            :style="{ width: progressPercent + '%' }"
          >
            {{ progressPercent }}%
          </div>
        </div>
        <div class="row text-muted small">
          <div class="col">
            <i class="fas fa-file me-1"></i>
            Files: {{ syncStatus.processed_files }} / {{ syncStatus.total_files }}
          </div>
          <div class="col">
            <i class="fas fa-plus-circle text-success me-1"></i>
            Created: {{ syncStatus.entities_created }}
          </div>
          <div class="col">
            <i class="fas fa-edit text-warning me-1"></i>
            Updated: {{ syncStatus.entities_updated }}
          </div>
          <div class="col">
            <i class="fas fa-trash text-danger me-1"></i>
            Deleted: {{ syncStatus.entities_deleted }}
          </div>
        </div>
        <div v-if="syncStatus.error" class="alert alert-danger mt-2 mb-0 small">
          {{ syncStatus.error }}
        </div>
      </div>
    </div>

    <!-- Indexed Repositories -->
    <div class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0"><i class="fas fa-list me-2"></i>Indexed Repositories</h5>
        <button class="btn btn-sm btn-outline-primary" @click="loadRepos" :disabled="loadingRepos">
          <i class="fas fa-refresh" :class="{ 'fa-spin': loadingRepos }"></i> Refresh
        </button>
      </div>
      <div class="card-body">
        <div v-if="loadingRepos" class="text-center py-3">
          <div class="spinner-border text-primary spinner-border-sm"></div>
          <span class="ms-2">Loading repositories...</span>
        </div>
        <div v-else-if="repos.length === 0" class="text-center text-muted py-4">
          <i class="fas fa-inbox fa-3x mb-3 d-block"></i>
          <p>No repositories indexed yet. Enter a path or URL above to build your first knowledge base.</p>
        </div>
        <div v-else class="table-responsive">
          <table class="table table-hover mb-0">
            <thead class="table-light">
              <tr>
                <th>Name</th>
                <th>Source</th>
                <th>Branch</th>
                <th>Status</th>
                <th>Last Sync</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="repo in repos" :key="repo.id">
                <td class="fw-bold">{{ repo.name }}</td>
                <td>
                  <code class="small text-truncate d-inline-block" style="max-width: 300px;">{{ repo.local_path || repo.url }}</code>
                </td>
                <td><span class="badge bg-secondary">{{ repo.branch || 'main' }}</span></td>
                <td>
                  <span class="badge" :class="repo.status === 'syncing' ? 'bg-warning text-dark' : 'bg-light text-dark'">
                    {{ repo.status || 'idle' }}
                  </span>
                </td>
                <td>
                  <span v-if="repo.last_sync" class="small">{{ formatDate(repo.last_sync) }}</span>
                  <span v-else class="text-muted small">Never</span>
                </td>
                <td>
                  <button
                    class="btn btn-sm btn-outline-success me-1"
                    @click="triggerSync(repo)"
                    title="Re-sync"
                    :disabled="isIndexing"
                  >
                    <i class="fas fa-sync-alt"></i>
                  </button>
                  <button
                    class="btn btn-sm btn-outline-info me-1"
                    @click="browseEntities(repo)"
                    title="Browse entities"
                  >
                    <i class="fas fa-th-list"></i>
                  </button>
                  <button
                    class="btn btn-sm btn-outline-danger"
                    @click="confirmDeleteRepo(repo)"
                    title="Delete repository"
                    :disabled="isIndexing"
                  >
                    <i class="fas fa-trash"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Repository Browser (Tabs: Knowledge / Entities) -->
    <div v-if="showEntityBrowser" class="card mb-4">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0">
          <i class="fas fa-cubes me-2"></i>
          <span v-if="browsingRepoName">{{ browsingRepoName }}</span>
        </h5>
        <button class="btn btn-sm btn-outline-secondary" @click="showEntityBrowser = false">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <!-- Tabs -->
      <ul class="nav nav-tabs px-3 pt-2">
        <li class="nav-item">
          <a class="nav-link" :class="{ active: browserTab === 'knowledge' }" href="#"
             @click.prevent="browserTab = 'knowledge'">
            <i class="fas fa-book me-1"></i>Knowledge Docs
            <span v-if="knowledgeDocs.length" class="badge bg-info ms-1">{{ knowledgeDocs.length }}</span>
          </a>
        </li>
        <li class="nav-item">
          <a class="nav-link" :class="{ active: browserTab === 'entities' }" href="#"
             @click.prevent="browserTab = 'entities'">
            <i class="fas fa-th-list me-1"></i>Entities
            <span class="badge bg-primary ms-1">{{ entityTotal }}</span>
          </a>
        </li>
      </ul>

      <div class="card-body">
        <!-- Knowledge Docs Tab -->
        <div v-if="browserTab === 'knowledge'">
          <div v-if="loadingKnowledgeDocs" class="text-center py-3">
            <div class="spinner-border text-primary spinner-border-sm"></div>
            <span class="ms-2">Loading knowledge documents...</span>
          </div>
          <div v-else-if="knowledgeDocs.length === 0" class="text-center text-muted py-4">
            <i class="fas fa-file-alt fa-2x mb-2 d-block"></i>
            No knowledge docs generated yet. Trigger a sync to auto-generate docs.
          </div>
          <div v-else>
            <!-- Doc selector -->
            <div class="d-flex gap-2 mb-3 flex-wrap">
              <button v-for="doc in knowledgeDocs" :key="doc.id"
                class="btn btn-sm"
                :class="selectedDocId === doc.id ? 'btn-primary' : 'btn-outline-primary'"
                @click="selectedDocId = doc.id">
                <i class="fas me-1" :class="docTypeIcon(doc.doc_type)"></i>
                {{ doc.title }}
              </button>
            </div>
            <!-- Doc content -->
            <div v-if="selectedDoc" class="knowledge-doc-content">
              <div class="d-flex justify-content-between align-items-center mb-2">
                <h5 class="mb-0">{{ selectedDoc.title }}</h5>
                <span class="small text-muted">
                  <i class="fas fa-clock me-1"></i>{{ formatDate(selectedDoc.updated_at) }}
                </span>
              </div>
              <div class="doc-body" v-html="renderMarkdown(selectedDoc.content)"></div>
            </div>
          </div>
        </div>

        <!-- Entities Tab -->
        <div v-if="browserTab === 'entities'">
          <div class="d-flex justify-content-end mb-2">
            <select class="form-select form-select-sm" style="width: 140px;" v-model="entityTypeFilter" @change="loadEntities(1)">
              <option value="">All Types</option>
              <option value="function">Functions</option>
              <option value="class">Classes</option>
            </select>
          </div>
          <div v-if="loadingEntities" class="text-center py-3">
            <div class="spinner-border text-primary spinner-border-sm"></div>
            <span class="ms-2">Loading entities...</span>
          </div>
          <div v-else-if="entities.length === 0" class="text-center text-muted py-3">
            No entities found. Try indexing a repository first.
          </div>
          <div v-else>
            <div class="row g-2">
              <div v-for="entity in entities" :key="entity.id" class="col-md-6 col-lg-4">
                <div class="card entity-card h-100" :class="{ 'border-primary': selectedEntity?.id === entity.id }"
                     @click="selectedEntity = selectedEntity?.id === entity.id ? null : entity"
                     role="button">
                  <div class="card-body py-2 px-3">
                    <div class="d-flex align-items-center mb-1">
                      <span class="badge me-2" :class="entityTypeBadge(entity.entity_type)">
                        {{ entity.entity_type }}
                      </span>
                      <strong class="text-truncate">{{ entity.name }}</strong>
                    </div>
                    <div class="small text-muted">
                      <i class="fas fa-file-code me-1"></i>
                      {{ entity.file_path }}:{{ entity.start_line }}
                    </div>
                    <div v-if="entity.language" class="small text-muted">
                      <i class="fas fa-code me-1"></i>{{ entity.language }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Entity Detail -->
            <div v-if="selectedEntity" class="mt-3">
              <div class="card border-primary">
                <div class="card-header bg-primary bg-opacity-10 d-flex justify-content-between">
                  <span>
                    <span class="badge me-2" :class="entityTypeBadge(selectedEntity.entity_type)">{{ selectedEntity.entity_type }}</span>
                    <strong>{{ selectedEntity.name }}</strong>
                  </span>
                  <button class="btn btn-sm btn-outline-secondary" @click="selectedEntity = null"><i class="fas fa-times"></i></button>
                </div>
                <div class="card-body">
                  <table class="table table-sm table-borderless mb-2">
                    <tr><td class="text-muted" style="width:100px">File</td><td><code>{{ selectedEntity.file_path }}:{{ selectedEntity.start_line }}-{{ selectedEntity.end_line }}</code></td></tr>
                    <tr><td class="text-muted">Language</td><td>{{ selectedEntity.language }}</td></tr>
                    <tr v-if="selectedEntity.signature"><td class="text-muted">Signature</td><td><code>{{ selectedEntity.signature }}</code></td></tr>
                    <tr v-if="selectedEntity.doc_string"><td class="text-muted">Doc</td><td>{{ selectedEntity.doc_string }}</td></tr>
                  </table>
                </div>
              </div>
            </div>

            <!-- Pagination -->
            <nav v-if="entityTotal > entityPerPage" class="mt-3">
              <ul class="pagination pagination-sm justify-content-center mb-0">
                <li class="page-item" :class="{ disabled: entityPage <= 1 }">
                  <a class="page-link" href="#" @click.prevent="loadEntities(entityPage - 1)">Prev</a>
                </li>
                <li class="page-item disabled">
                  <span class="page-link">{{ entityPage }} / {{ Math.ceil(entityTotal / entityPerPage) }}</span>
                </li>
                <li class="page-item" :class="{ disabled: entityPage >= Math.ceil(entityTotal / entityPerPage) }">
                  <a class="page-link" href="#" @click.prevent="loadEntities(entityPage + 1)">Next</a>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </div>
    </div>

    <!-- Search Section -->
    <div class="card mb-4">
      <div class="card-header bg-success text-white">
        <h5 class="mb-0"><i class="fas fa-search me-2"></i>Search Knowledge Base</h5>
      </div>
      <div class="card-body">
        <form @submit.prevent="handleSearch">
          <div class="row g-3 align-items-end">
            <div class="col-md-8">
              <label for="searchInput" class="form-label">Ask about your codebase</label>
              <input
                type="text"
                class="form-control form-control-lg"
                id="searchInput"
                v-model="searchQuery"
                placeholder="e.g., How does authentication work? / What calls HandleRequest?"
                :disabled="isSearching"
              />
            </div>
            <div class="col-md-2">
              <label for="topK" class="form-label">Top K</label>
              <select class="form-select" id="topK" v-model.number="searchTopK">
                <option :value="5">5</option>
                <option :value="10">10</option>
                <option :value="20">20</option>
              </select>
            </div>
            <div class="col-md-2">
              <button
                type="submit"
                class="btn btn-success w-100 btn-lg"
                :disabled="!searchQuery.trim() || isSearching"
              >
                <span v-if="isSearching" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="fas fa-search me-1"></i>
                Search
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <!-- Search Results -->
    <div v-if="searchResult" class="card mb-4">
      <div class="card-header">
        <h5 class="mb-0"><i class="fas fa-lightbulb me-2"></i>Answer</h5>
      </div>
      <div class="card-body">
        <div class="answer-content mb-4" v-html="renderedAnswer"></div>

        <div v-if="searchResult.entities && searchResult.entities.length > 0">
          <h6 class="border-bottom pb-2 mb-3">
            <i class="fas fa-code me-2"></i>Related Code Entities ({{ searchResult.entities.length }})
          </h6>
          <div class="row g-2">
            <div
              v-for="entity in searchResult.entities"
              :key="entity.id"
              class="col-md-6"
            >
              <div class="card entity-card h-100">
                <div class="card-body py-2 px-3">
                  <div class="d-flex justify-content-between align-items-start">
                    <div>
                      <span class="badge me-2" :class="entityTypeBadge(entity.entity_type)">
                        {{ entity.entity_type }}
                      </span>
                      <strong>{{ entity.name }}</strong>
                    </div>
                    <span v-if="entity.language" class="badge bg-light text-dark">{{ entity.language }}</span>
                  </div>
                  <div class="small text-muted mt-1">
                    <i class="fas fa-file-code me-1"></i>
                    {{ entity.file_path }}:{{ entity.start_line }}-{{ entity.end_line }}
                  </div>
                  <div v-if="entity.signature" class="small mt-1"><code>{{ entity.signature }}</code></div>
                  <div v-if="entity.doc_string" class="small mt-1 text-muted">{{ entity.doc_string }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="text-center text-muted py-3">
          No matching code entities found.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import type { CodeKGRepo, CodeKGEntity, CodeKGSearchResult, CodeKGSyncStatus, CodeKGKnowledgeDoc } from '@/types';

// --- Repo registration ---
const repoInput = ref('');
const repoBranch = ref('main');
const repoName = ref('');
const isIndexing = ref(false);
const syncStatus = ref<CodeKGSyncStatus | null>(null);
let pollTimer: ReturnType<typeof setInterval> | null = null;

// --- Repo list ---
const repos = ref<CodeKGRepo[]>([]);
const loadingRepos = ref(false);

// --- Repository browser ---
const showEntityBrowser = ref(false);
const browserTab = ref<'knowledge' | 'entities'>('knowledge');
const browsingRepoId = ref('');
const browsingRepoName = ref('');

// --- Knowledge docs ---
const knowledgeDocs = ref<CodeKGKnowledgeDoc[]>([]);
const loadingKnowledgeDocs = ref(false);
const selectedDocId = ref('');
const selectedDoc = computed(() => knowledgeDocs.value.find(d => d.id === selectedDocId.value) || null);

// --- Entity browser ---
const entities = ref<CodeKGEntity[]>([]);
const entityTotal = ref(0);
const entityPage = ref(1);
const entityPerPage = 18;
const entityTypeFilter = ref('');
const loadingEntities = ref(false);
const selectedEntity = ref<CodeKGEntity | null>(null);

// --- Search ---
const searchQuery = ref('');
const searchTopK = ref(10);
const isSearching = ref(false);
const searchResult = ref<CodeKGSearchResult | null>(null);

const isGitUrl = computed(() => {
  const v = repoInput.value.trim();
  return v.startsWith('http://') || v.startsWith('https://') || v.startsWith('git@') || v.endsWith('.git');
});

const progressPercent = computed(() => {
  if (!syncStatus.value || syncStatus.value.total_files === 0) return 0;
  return Math.round((syncStatus.value.processed_files / syncStatus.value.total_files) * 100);
});

const statusBadgeClass = computed(() => {
  if (!syncStatus.value) return 'bg-secondary';
  switch (syncStatus.value.status) {
    case 'running': return 'bg-primary';
    case 'completed': return 'bg-success';
    case 'failed': return 'bg-danger';
    default: return 'bg-secondary';
  }
});

const renderedAnswer = computed(() => {
  if (!searchResult.value?.answer) return '';
  return searchResult.value.answer
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>');
});

function deriveRepoName(input: string): string {
  const trimmed = input.trim().replace(/\/$/, '').replace(/\.git$/, '');
  const parts = trimmed.split('/');
  return parts[parts.length - 1] || 'repo';
}

async function handleSubmitRepo() {
  if (!repoInput.value.trim()) return;

  const name = repoName.value.trim() || deriveRepoName(repoInput.value);
  const data: Record<string, string | undefined> = {
    name,
    url: repoInput.value.trim(),
    branch: repoBranch.value || 'main',
  };
  if (!isGitUrl.value) {
    data.local_path = repoInput.value.trim();
  }

  try {
    isIndexing.value = true;
    const repo = await apiService.codekgRegisterRepo(data as any);
    showToast(`Repository "${name}" registered. Starting indexing...`, 'success');

    const repoId = repo.id || repo.ID;
    const syncResp = await apiService.codekgTriggerSync(repoId);
    syncStatus.value = {
      job_id: syncResp.job_id,
      status: 'running',
      total_files: 0,
      processed_files: 0,
      entities_created: 0,
      entities_updated: 0,
      entities_deleted: 0,
    };

    startPolling(repoId);
  } catch (error: any) {
    showToast(error?.response?.data?.error || 'Failed to register repository', 'danger');
    isIndexing.value = false;
  }
}

function startPolling(repoId: string) {
  stopPolling();
  pollTimer = setInterval(async () => {
    try {
      const status = await apiService.codekgGetSyncStatus(repoId);
      syncStatus.value = status;
      if (status.status === 'completed' || status.status === 'failed' || status.status === 'idle') {
        stopPolling();
        isIndexing.value = false;
        await loadRepos();
        if (status.status === 'completed' || (status.status === 'idle' && status.entities_created > 0)) {
          showToast(`Knowledge base built! ${status.entities_created} entities indexed.`, 'success');
          browseEntitiesById(repoId, repos.value.find(r => r.id === repoId)?.name || '');
        }
      }
    } catch {
      stopPolling();
      isIndexing.value = false;
    }
  }, 2000);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

async function loadRepos() {
  try {
    loadingRepos.value = true;
    repos.value = await apiService.codekgListRepos();
  } catch {
    repos.value = [];
  } finally {
    loadingRepos.value = false;
  }
}

async function triggerSync(repo: CodeKGRepo) {
  try {
    isIndexing.value = true;
    const syncResp = await apiService.codekgTriggerSync(repo.id);
    syncStatus.value = {
      job_id: syncResp.job_id,
      status: 'running',
      total_files: 0,
      processed_files: 0,
      entities_created: 0,
      entities_updated: 0,
      entities_deleted: 0,
    };
    startPolling(repo.id);
    showToast(`Re-syncing "${repo.name}"...`, 'info');
  } catch (error: any) {
    showToast(error?.response?.data?.error || 'Failed to trigger sync', 'danger');
    isIndexing.value = false;
  }
}

async function confirmDeleteRepo(repo: CodeKGRepo) {
  if (!confirm(`Delete "${repo.name}" and all its entities, embeddings, and knowledge docs?`)) return;
  try {
    await apiService.codekgDeleteRepo(repo.id);
    showToast(`Repository "${repo.name}" deleted.`, 'success');
    if (browsingRepoId.value === repo.id) {
      showEntityBrowser.value = false;
    }
    await loadRepos();
  } catch (error: any) {
    showToast(error?.response?.data?.error || 'Failed to delete repository', 'danger');
  }
}

function browseEntities(repo: CodeKGRepo) {
  browseEntitiesById(repo.id, repo.name);
}

function browseEntitiesById(repoId: string, name: string) {
  browsingRepoId.value = repoId;
  browsingRepoName.value = name;
  showEntityBrowser.value = true;
  browserTab.value = 'knowledge';
  entityTypeFilter.value = '';
  selectedEntity.value = null;
  loadKnowledgeDocs(repoId);
  loadEntities(1);
}

async function loadKnowledgeDocs(repoId: string) {
  try {
    loadingKnowledgeDocs.value = true;
    knowledgeDocs.value = await apiService.codekgGetKnowledgeDocs(repoId);
    if (knowledgeDocs.value.length > 0) {
      selectedDocId.value = knowledgeDocs.value[0].id;
    }
  } catch {
    knowledgeDocs.value = [];
  } finally {
    loadingKnowledgeDocs.value = false;
  }
}

async function loadEntities(page: number) {
  try {
    loadingEntities.value = true;
    entityPage.value = page;
    const resp = await apiService.codekgGetEntities({
      repo_id: browsingRepoId.value || undefined,
      type: entityTypeFilter.value || undefined,
      page,
      per_page: entityPerPage,
    });
    entities.value = resp.data || [];
    entityTotal.value = resp.total || 0;
  } catch {
    entities.value = [];
    entityTotal.value = 0;
  } finally {
    loadingEntities.value = false;
  }
}

async function handleSearch() {
  if (!searchQuery.value.trim()) return;
  try {
    isSearching.value = true;
    searchResult.value = await apiService.codekgSearch(searchQuery.value, { top_k: searchTopK.value });
  } catch (error: any) {
    showToast(error?.response?.data?.error || 'Search failed', 'danger');
    searchResult.value = null;
  } finally {
    isSearching.value = false;
  }
}

function entityTypeBadge(entityType: string): string {
  switch (entityType?.toLowerCase()) {
    case 'function': return 'bg-primary';
    case 'struct': case 'class': return 'bg-success';
    case 'interface': return 'bg-warning text-dark';
    case 'package': return 'bg-info';
    case 'file': return 'bg-secondary';
    default: return 'bg-dark';
  }
}

function docTypeIcon(docType: string): string {
  switch (docType) {
    case 'repo-map': return 'fa-sitemap';
    case 'overview': return 'fa-eye';
    case 'architecture': return 'fa-project-diagram';
    default: return 'fa-file-alt';
  }
}

function renderMarkdown(content: string): string {
  if (!content) return '';
  return content
    .replace(/^### (.+)$/gm, '<h5 class="mt-3 mb-2">$1</h5>')
    .replace(/^## (.+)$/gm, '<h4 class="mt-4 mb-2">$1</h4>')
    .replace(/^# (.+)$/gm, '<h3 class="mt-4 mb-3">$1</h3>')
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre class="doc-pre"><code class="language-$1">$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/^\|(.+)\|$/gm, (match) => {
      const cells = match.split('|').filter(c => c.trim() !== '');
      if (cells.every(c => /^[\s-]+$/.test(c))) return '';
      const tds = cells.map(c => `<td class="px-2 py-1">${c.trim()}</td>`).join('');
      return `<tr>${tds}</tr>`;
    })
    .replace(/(<tr>[\s\S]*?<\/tr>)/g, '<table class="table table-sm table-bordered my-2">$1</table>')
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>[\s\S]*?<\/li>)/g, '<ul class="mb-2">$1</ul>')
    .replace(/<\/ul>\s*<ul[^>]*>/g, '')
    .replace(/\n\n/g, '<br><br>')
    .replace(/\n/g, '<br>');
}

function formatDate(dateString: string): string {
  if (!dateString) return '';
  return new Date(dateString).toLocaleString();
}

onMounted(() => {
  loadRepos();
});

onUnmounted(() => {
  stopPolling();
});
</script>

<style scoped>
.entity-card {
  transition: transform 0.15s, box-shadow 0.15s;
  border-left: 3px solid #0d6efd;
  cursor: pointer;
}

.entity-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.answer-content {
  line-height: 1.7;
}

.answer-content pre {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 0.375rem;
  padding: 1rem;
  overflow-x: auto;
}

.answer-content code {
  background: #f0f0f0;
  padding: 0.15rem 0.35rem;
  border-radius: 0.2rem;
  font-size: 0.875em;
}

.answer-content pre code {
  background: none;
  padding: 0;
}

.knowledge-doc-content .doc-body {
  line-height: 1.7;
  max-height: 600px;
  overflow-y: auto;
}

.knowledge-doc-content .doc-pre {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 0.375rem;
  padding: 1rem;
  overflow-x: auto;
  font-size: 0.85em;
}
</style>
