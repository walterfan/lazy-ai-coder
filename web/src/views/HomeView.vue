<template>
  <div class="home-page">
    <!-- Hero -->
    <section class="hero-section text-center py-4 bg-dark text-white">
      <div class="container">
        <h1 class="display-5 fw-bold mb-1">{{ appStore.title }}</h1>
        <p class="lead text-secondary mb-4">
          Your AI-powered toolkit for every phase of the software development lifecycle.
        </p>

        <!-- Quick-nav cards -->
        <div class="row g-3 justify-content-center mb-4">
          <div class="col-6 col-md-3" v-for="nav in quickNavs" :key="nav.route">
            <router-link :to="nav.route" class="nav-card text-decoration-none">
              <div class="nav-card-inner">
                <div class="nav-card-icon" :style="{ background: nav.gradient }">
                  <i :class="nav.icon"></i>
                </div>
                <div class="nav-card-label">{{ nav.label }}</div>
                <div class="nav-card-desc">{{ nav.desc }}</div>
              </div>
            </router-link>
          </div>
        </div>

        <!-- Search -->
        <div class="row justify-content-center">
          <div class="col-lg-6">
            <div class="input-group shadow">
              <input
                v-model="searchQuery"
                type="text"
                class="form-control"
                placeholder="Search skills, commands, rules..."
                @keyup.enter="fetchAssets"
              />
              <button class="btn btn-primary" @click="fetchAssets" :disabled="loading">
                <i class="fas fa-search"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Tabs + Content -->
    <section class="container mt-n3 position-relative" style="z-index: 1;">
      <div class="card shadow-sm border-0">
        <!-- Tab header -->
        <div class="card-header bg-white border-bottom-0 pt-3 pb-0">
          <ul class="nav nav-tabs card-header-tabs">
            <li class="nav-item" v-for="tab in tabs" :key="tab.value">
              <a
                class="nav-link d-flex align-items-center"
                :class="{ active: activeTab === tab.value }"
                href="#"
                @click.prevent="switchTab(tab.value)"
              >
                <i :class="tab.icon" class="me-2"></i>
                {{ tab.label }}
                <span
                  class="badge rounded-pill ms-2"
                  :class="activeTab === tab.value ? 'bg-primary' : 'bg-secondary'"
                >
                  {{ tabCounts[tab.value] ?? 0 }}
                </span>
              </a>
            </li>
          </ul>
        </div>

        <!-- Provider filter pills -->
        <div v-if="providers.length > 1" class="provider-filter px-3 py-2 border-bottom bg-light">
          <span class="small text-muted me-2">Provider:</span>
          <button
            class="provider-pill"
            :class="{ active: selectedProvider === '' }"
            @click="selectedProvider = ''"
          >
            All
          </button>
          <button
            v-for="prov in providers"
            :key="prov"
            class="provider-pill"
            :class="{ active: selectedProvider === prov }"
            @click="selectedProvider = prov"
          >
            {{ prov }}
          </button>
        </div>

        <div class="card-body">
          <!-- Loading -->
          <div v-if="loading" class="text-center py-5 text-muted">
            <div class="spinner-border text-primary mb-3" role="status"></div>
            <p class="mb-0">Loading {{ activeTab }}...</p>
          </div>

          <!-- Error -->
          <div v-else-if="error" class="alert alert-danger">{{ error }}</div>

          <!-- Empty -->
          <div v-else-if="filteredItems.length === 0" class="text-center py-5 text-muted">
            <i class="fas fa-inbox fa-3x mb-3 d-block"></i>
            <p class="mb-1 fw-semibold">No {{ activeTab }} found</p>
            <p class="small">Try a different search term or switch tabs.</p>
          </div>

          <!-- Grid -->
          <div v-else>
            <p class="small text-muted mb-3">{{ filteredItems.length }} result(s)</p>
            <div class="row g-3">
              <div
                class="col-md-6 col-lg-4"
                v-for="item in filteredItems"
                :key="item.path"
              >
                <div
                  class="card h-100 asset-card"
                  :class="{ 'border-primary': previewItem?.path === item.path }"
                  @click="preview(item)"
                >
                  <div class="card-body d-flex flex-column">
                    <div class="d-flex justify-content-between align-items-start mb-2">
                      <span :class="`badge bg-${typeBadgeColor(item.type)}`">{{ item.type }}</span>
                      <span v-if="item.category" class="badge bg-light text-dark border">{{ item.category }}</span>
                    </div>
                    <h6 class="card-title mb-1 text-truncate" :title="item.name">{{ item.name }}</h6>
                    <p class="card-text small text-muted flex-grow-1 snippet-text">{{ item.snippet }}</p>
                    <div class="d-flex gap-2 mt-auto pt-2 border-top">
                      <button
                        class="btn btn-sm btn-outline-primary flex-fill"
                        @click.stop="preview(item)"
                      >
                        <i class="fas fa-eye me-1"></i>View
                      </button>
                      <button
                        class="btn btn-sm btn-outline-success flex-fill"
                        @click.stop="download(item)"
                      >
                        <i class="fas fa-download me-1"></i>Download
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Preview Modal -->
    <div
      v-if="previewItem"
      class="modal fade show d-block"
      tabindex="-1"
      @click.self="previewItem = null"
    >
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <div>
              <h5 class="modal-title mb-0">{{ previewItem.name }}</h5>
              <small class="text-muted">{{ previewItem.path }}</small>
            </div>
            <button class="btn-close" @click="previewItem = null"></button>
          </div>
          <div class="modal-body">
            <div v-if="previewLoading" class="text-center py-4">
              <div class="spinner-border spinner-border-sm text-primary"></div>
              <span class="ms-2">Loading content...</span>
            </div>
            <div v-else class="markdown-preview" v-html="renderedMarkdown"></div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline-secondary" @click="previewItem = null">Close</button>
            <button class="btn btn-success" @click="download(previewItem!)">
              <i class="fas fa-download me-1"></i>Download
            </button>
          </div>
        </div>
      </div>
    </div>
    <div v-if="previewItem" class="modal-backdrop fade show"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { marked } from 'marked';
import { apiService } from '@/services/apiService';
import { useAppStore } from '@/stores/appStore';
import { showToast } from '@/utils/toast';
import type { AssetItem } from '@/types';

const appStore = useAppStore();

const quickNavs = [
  {
    label: 'Coding Mate',
    desc: 'AI pair programmer',
    icon: 'fas fa-graduation-cap',
    route: '/chat',
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  },
  {
    label: 'Code Assistant',
    desc: 'Write & review code',
    icon: 'fas fa-code',
    route: '/assistant?module=code',
    gradient: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)',
  },
  {
    label: 'Code Review',
    desc: 'Analyze & improve',
    icon: 'fas fa-search',
    route: '/assistant?module=review',
    gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  },
  {
    label: 'Knowledge Base',
    desc: 'Index & search repos',
    icon: 'fas fa-project-diagram',
    route: '/codekg',
    gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
  },
];

const tabs = [
  { value: 'skill', label: 'Skills', icon: 'fas fa-graduation-cap' },
  { value: 'command', label: 'Commands', icon: 'fas fa-terminal' },
  { value: 'rule', label: 'Rules', icon: 'fas fa-gavel' },
] as const;

type TabValue = (typeof tabs)[number]['value'];

const activeTab = ref<TabValue>('skill');
const searchQuery = ref('');
const selectedProvider = ref<string>('');
const allItems = ref<AssetItem[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const previewItem = ref<AssetItem | null>(null);
const previewContent = ref('');
const previewLoading = ref(false);

const tabCounts = computed(() => {
  const counts: Record<string, number> = { skill: 0, command: 0, rule: 0 };
  for (const it of allItems.value) {
    if (counts[it.type] !== undefined) counts[it.type]++;
  }
  return counts;
});

const providers = computed(() => {
  const cats = new Set<string>();
  for (const it of allItems.value) {
    if (it.type === activeTab.value && it.category) {
      cats.add(it.category);
    }
  }
  return Array.from(cats).sort();
});

const filteredItems = computed(() => {
  return allItems.value.filter((it) => {
    if (it.type !== activeTab.value) return false;
    if (selectedProvider.value && it.category !== selectedProvider.value) return false;
    return true;
  });
});

function typeBadgeColor(type: string): string {
  switch (type) {
    case 'command': return 'primary';
    case 'rule': return 'info';
    case 'skill': return 'success';
    default: return 'secondary';
  }
}

async function fetchAssets() {
  loading.value = true;
  error.value = null;
  try {
    const res = await apiService.listAssets({
      q: searchQuery.value || undefined,
    });
    allItems.value = res.data || [];
  } catch (e) {
    const msg = e instanceof Error ? e.message : 'Failed to load assets';
    error.value = msg;
    showToast(msg, 'danger');
  } finally {
    loading.value = false;
  }
}

function switchTab(tab: TabValue) {
  activeTab.value = tab;
  selectedProvider.value = '';
  previewItem.value = null;
}

const renderedMarkdown = computed(() => {
  if (!previewContent.value || previewLoading.value) return '';
  return marked.parse(previewContent.value, { async: false }) as string;
});

async function preview(item: AssetItem) {
  previewItem.value = item;
  previewLoading.value = true;
  previewContent.value = '';
  try {
    const response = await apiService.get<string>('/assets/download', {
      params: { path: item.path },
      responseType: 'text',
    });
    previewContent.value = response.data ?? '';
  } catch {
    previewContent.value = 'Failed to load content.';
  } finally {
    previewLoading.value = false;
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

watch(searchQuery, (val) => {
  if (val === '') fetchAssets();
});

onMounted(() => {
  fetchAssets();
});
</script>

<style scoped>
.hero-section {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}

/* Quick-nav cards */
.nav-card {
  display: block;
  color: #fff;
}

.nav-card-inner {
  background: rgba(255, 255, 255, 0.07);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 14px;
  padding: 1.25rem 0.75rem;
  text-align: center;
  transition: transform 0.18s, background 0.18s, box-shadow 0.18s;
}

.nav-card-inner:hover {
  transform: translateY(-4px);
  background: rgba(255, 255, 255, 0.13);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
}

.nav-card-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
  border-radius: 14px;
  font-size: 1.4rem;
  color: #fff;
  margin-bottom: 0.65rem;
}

.nav-card-label {
  font-weight: 600;
  font-size: 0.95rem;
  margin-bottom: 0.2rem;
}

.nav-card-desc {
  font-size: 0.78rem;
  color: rgba(255, 255, 255, 0.55);
}

/* Provider filter */
.provider-filter {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.provider-pill {
  display: inline-block;
  padding: 0.2rem 0.65rem;
  border: 1px solid #dee2e6;
  border-radius: 20px;
  background: #fff;
  font-size: 0.78rem;
  color: #555;
  cursor: pointer;
  transition: all 0.12s;
}

.provider-pill:hover {
  border-color: #667eea;
  color: #667eea;
}

.provider-pill.active {
  background: #667eea;
  border-color: #667eea;
  color: #fff;
}

.mt-n3 {
  margin-top: -1.5rem;
}

.asset-card {
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.asset-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.snippet-text {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 0.82rem;
  line-height: 1.5;
}

.nav-tabs .nav-link {
  font-weight: 500;
  color: #555;
}

.nav-tabs .nav-link.active {
  color: #0d6efd;
  border-color: #dee2e6 #dee2e6 #fff;
}

.markdown-preview {
  font-size: 0.92rem;
  line-height: 1.7;
}

.markdown-preview :deep(pre) {
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  padding: 0.75rem 1rem;
  font-size: 0.85rem;
  overflow-x: auto;
}

.markdown-preview :deep(code) {
  font-size: 0.88em;
  color: #d63384;
  background: #f8f9fa;
  padding: 0.1em 0.3em;
  border-radius: 3px;
}

.markdown-preview :deep(pre code) {
  color: inherit;
  background: none;
  padding: 0;
}

.markdown-preview :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
  font-size: 0.88rem;
}

.markdown-preview :deep(th),
.markdown-preview :deep(td) {
  border: 1px solid #dee2e6;
  padding: 0.5rem 0.75rem;
  text-align: left;
}

.markdown-preview :deep(th) {
  background: #f1f3f5;
  font-weight: 600;
}

.markdown-preview :deep(tr:nth-child(even)) {
  background: #f8f9fa;
}

.markdown-preview :deep(blockquote) {
  border-left: 4px solid #dee2e6;
  padding: 0.5rem 1rem;
  margin: 1rem 0;
  color: #6c757d;
  background: #f8f9fa;
  border-radius: 0 4px 4px 0;
}

.markdown-preview :deep(img) {
  max-width: 100%;
  height: auto;
}
</style>
