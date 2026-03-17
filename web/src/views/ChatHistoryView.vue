<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-code-branch me-2"></i>Coding Mate History
      </h2>
      <router-link to="/chat" class="btn btn-primary">
        <i class="fas fa-plus me-1"></i>New session
      </router-link>
    </div>

    <!-- Stats Panel -->
    <div v-if="store.stats" class="row mb-4">
      <div class="col-md-3">
        <div class="card text-center">
          <div class="card-body">
            <h3 class="text-primary">{{ store.stats.total }}</h3>
            <small class="text-muted">Total sessions</small>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card text-center">
          <div class="card-body">
            <h3 class="text-success">{{ store.stats.streak }}</h3>
            <small class="text-muted">Day Streak</small>
          </div>
        </div>
      </div>
      <div class="col-md-6">
        <div class="card">
          <div class="card-body">
            <small class="text-muted d-block mb-2">Sessions by type</small>
            <div class="d-flex flex-wrap gap-2">
              <span
                v-for="(count, type) in store.stats.by_type"
                :key="type"
                :class="`badge bg-${InputTypeConfig[type as InputType]?.color || 'secondary'}`"
              >
                {{ InputTypeConfig[type as InputType]?.label || type }}: {{ count }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row g-3 align-items-center">
          <div class="col-md-4">
            <div class="input-group">
              <span class="input-group-text">
                <i class="fas fa-search"></i>
              </span>
              <input
                v-model="searchQuery"
                type="text"
                class="form-control"
                placeholder="Search sessions..."
                @keyup.enter="handleSearch"
              />
            </div>
          </div>
          <div class="col-md-4">
            <select v-model="selectedType" class="form-select" @change="handleFilter">
              <option value="">All Types</option>
              <option v-for="(config, type) in InputTypeConfig" :key="type" :value="type">
                {{ config.label }}
              </option>
            </select>
          </div>
          <div class="col-md-4 text-end">
            <button @click="handleSearch" class="btn btn-primary me-2">
              <i class="fas fa-search me-1"></i>Search
            </button>
            <button @click="clearFilters" class="btn btn-outline-secondary">
              <i class="fas fa-times me-1"></i>Clear
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Sessions list -->
    <div class="card">
      <div class="card-body">
        <div v-if="loading" class="text-center py-5">
          <i class="fas fa-spinner fa-spin fa-2x text-primary"></i>
          <p class="mt-2 text-muted">Loading sessions...</p>
        </div>

        <div v-else-if="store.records.length === 0" class="text-center py-5 text-muted">
          <i class="fas fa-inbox fa-3x mb-3"></i>
          <p>No sessions found</p>
          <router-link to="/chat" class="btn btn-primary">
            Create your first session
          </router-link>
        </div>

        <div v-else class="list-group list-group-flush">
          <div
            v-for="record in store.records"
            :key="record.id"
            class="list-group-item list-group-item-action"
          >
            <div class="d-flex justify-content-between align-items-start">
              <div class="flex-grow-1">
                <div class="d-flex align-items-center gap-2 mb-1">
                  <span :class="`badge bg-${InputTypeConfig[record.input_type]?.color}`">
                    <i :class="`fas ${InputTypeConfig[record.input_type]?.icon} me-1`"></i>
                    {{ InputTypeConfig[record.input_type]?.label }}
                  </span>
                  <small class="text-muted">
                    {{ formatDate(record.created_time) }}
                  </small>
                </div>
                <h6 class="mb-1">{{ record.user_input }}</h6>
                <p class="mb-0 text-muted small text-truncate" style="max-width: 500px;">
                  {{ record.response_summary }}
                </p>
              </div>
              <div class="ms-3">
                <button
                  @click="deleteRecord(record.id)"
                  class="btn btn-sm btn-outline-danger"
                  title="Delete session"
                >
                  <i class="fas fa-trash-alt"></i>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <nav v-if="totalPages > 1" class="mt-4">
          <ul class="pagination justify-content-center mb-0">
            <li class="page-item" :class="{ disabled: store.page === 1 }">
              <button class="page-link" @click="goToPage(store.page - 1)">
                <i class="fas fa-chevron-left"></i>
              </button>
            </li>
            <li
              v-for="page in visiblePages"
              :key="page"
              class="page-item"
              :class="{ active: page === store.page }"
            >
              <button class="page-link" @click="goToPage(page)">
                {{ page }}
              </button>
            </li>
            <li class="page-item" :class="{ disabled: store.page === totalPages }">
              <button class="page-link" @click="goToPage(store.page + 1)">
                <i class="fas fa-chevron-right"></i>
              </button>
            </li>
          </ul>
        </nav>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useChatRecordStore } from '@/stores/chatRecordStore';
import { InputTypeConfig, type InputType } from '@/types/chatRecord';

const store = useChatRecordStore();
const loading = ref(false);
const searchQuery = ref('');
const selectedType = ref<InputType | ''>('');

// Computed
const totalPages = computed(() => {
  return Math.ceil(store.totalRecords / store.pageSize);
});

const visiblePages = computed(() => {
  const pages: number[] = [];
  const total = totalPages.value;
  const current = store.page;
  const range = 2; // Show 2 pages before and after current

  for (let i = Math.max(1, current - range); i <= Math.min(total, current + range); i++) {
    pages.push(i);
  }
  return pages;
});

// Methods
const loadRecords = async () => {
  loading.value = true;
  try {
    await store.fetchRecords(
      store.page,
      store.pageSize,
      selectedType.value || undefined,
      searchQuery.value || undefined
    );
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  store.page = 1;
  loadRecords();
};

const handleFilter = () => {
  store.page = 1;
  loadRecords();
};

const clearFilters = () => {
  searchQuery.value = '';
  selectedType.value = '';
  store.page = 1;
  loadRecords();
};

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    store.page = page;
    loadRecords();
  }
};

const deleteRecord = async (id: string) => {
  if (confirm('Are you sure you want to delete this session?')) {
    await store.deleteRecord(id);
  }
};

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

// Lifecycle
onMounted(async () => {
  await Promise.all([loadRecords(), store.fetchStats()]);
});
</script>

<style scoped>
.list-group-item {
  border-left: none;
  border-right: none;
}

.list-group-item:first-child {
  border-top: none;
}

.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
