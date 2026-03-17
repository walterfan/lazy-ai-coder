<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-folder-open me-2"></i>Assets
        </h2>
        <p class="text-muted mb-0 small">
          Search and download commands, rules, and skills from the assets folder.
        </p>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
        <div class="row g-3 mb-4">
          <div class="col-md-4">
            <label class="form-label">Type</label>
            <select v-model="filterType" class="form-select">
              <option value="all">All</option>
              <option value="command">Commands</option>
              <option value="rule">Rules</option>
              <option value="skill">Skills</option>
            </select>
          </div>
          <div class="col-md-4">
            <label class="form-label">Search</label>
            <input
              v-model="searchQuery"
              type="text"
              class="form-control"
              placeholder="Name, path, or content..."
              @keyup.enter="fetchAssets"
            />
          </div>
          <div class="col-md-4 d-flex align-items-end">
            <button class="btn btn-primary me-2" @click="fetchAssets" :disabled="loading">
              <i class="fas fa-search me-1"></i>Search
            </button>
            <button class="btn btn-outline-secondary" @click="clearFilters">
              <i class="fas fa-times me-1"></i>Clear
            </button>
          </div>
        </div>

        <div v-if="error" class="alert alert-danger">{{ error }}</div>
        <div v-else-if="loading" class="text-center py-5 text-muted">
          <i class="fas fa-spinner fa-spin fa-2x mb-2"></i>
          <p class="mb-0">Loading assets...</p>
        </div>
        <div v-else-if="items.length === 0" class="text-center py-5 text-muted">
          <i class="fas fa-inbox fa-2x mb-2"></i>
          <p class="mb-0">No assets found. Try changing filters or search.</p>
        </div>
        <div v-else>
          <p class="small text-muted mb-3">Showing {{ items.length }} item(s)</p>
          <div class="table-responsive">
            <table class="table table-hover align-middle">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Name</th>
                  <th>Category</th>
                  <th>Path</th>
                  <th class="text-end">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in items" :key="item.path">
                  <td>
                    <span :class="`badge bg-${typeBadgeColor(item.type)}`">
                      {{ item.type }}
                    </span>
                  </td>
                  <td>
                    <strong>{{ item.name }}</strong>
                  </td>
                  <td>
                    <span v-if="item.category" class="text-muted small">{{ item.category }}</span>
                    <span v-else class="text-muted">—</span>
                  </td>
                  <td>
                    <code class="small">{{ item.path }}</code>
                  </td>
                  <td class="text-end">
                    <button
                      class="btn btn-sm btn-outline-primary me-1"
                      @click="preview(item)"
                      title="Preview"
                    >
                      <i class="fas fa-eye"></i>
                    </button>
                    <button
                      class="btn btn-sm btn-outline-success"
                      @click="download(item)"
                      title="Download"
                    >
                      <i class="fas fa-download"></i>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="previewItem" class="mt-4 pt-3 border-top">
            <h6 class="mb-2">
              Preview: <code>{{ previewItem.path }}</code>
              <button class="btn btn-sm btn-link p-0 ms-2" @click="previewItem = null">
                <i class="fas fa-times"></i>
              </button>
            </h6>
            <pre class="bg-light p-3 rounded small" style="max-height: 300px; overflow: auto;">{{ previewContent }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import type { AssetItem } from '@/types';

const filterType = ref('all');
const searchQuery = ref('');
const items = ref<AssetItem[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const previewItem = ref<AssetItem | null>(null);
const previewContent = ref('');

function typeBadgeColor(type: string): string {
  switch (type) {
    case 'command':
      return 'primary';
    case 'rule':
      return 'info';
    case 'skill':
      return 'success';
    default:
      return 'secondary';
  }
}

async function fetchAssets() {
  loading.value = true;
  error.value = null;
  try {
    const res = await apiService.listAssets({
      type: filterType.value === 'all' ? undefined : filterType.value,
      q: searchQuery.value || undefined,
    });
    items.value = res.data || [];
  } catch (e) {
    const msg = e instanceof Error ? e.message : 'Failed to load assets';
    error.value = msg;
    showToast(msg, 'danger');
  } finally {
    loading.value = false;
  }
}

function clearFilters() {
  filterType.value = 'all';
  searchQuery.value = '';
  previewItem.value = null;
  fetchAssets();
}

async function preview(item: AssetItem) {
  previewItem.value = item;
  previewContent.value = 'Loading...';
  try {
    const response = await apiService.get<string>('/assets/download', {
      params: { path: item.path },
      responseType: 'text',
    });
    previewContent.value = response.data ?? '';
  } catch {
    previewContent.value = 'Failed to load content.';
  }
}

async function download(item: AssetItem) {
  try {
    if (item.type === 'skill') {
      await apiService.downloadSkillZip(item.path);
      showToast(`Downloaded ${item.name}.zip`, 'success');
    } else {
      const filename = item.path.split('/').pop() || item.name;
      await apiService.downloadAsset(item.path, filename);
      showToast(`Downloaded ${filename}`, 'success');
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : 'Download failed';
    showToast(msg, 'danger');
  }
}

onMounted(() => {
  fetchAssets();
});
</script>
