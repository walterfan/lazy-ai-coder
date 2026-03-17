<template>
  <div class="container mt-4 mb-5">
    <!-- Header -->
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-building"></i> Realm Management
        </h2>
        <p class="text-muted mb-0">
          <small>Manage organizational realms and teams</small>
        </p>
      </div>
      <div class="btn-group">
        <button @click="showCreateModal" class="btn btn-primary">
          <i class="fas fa-plus"></i> Create Realm
        </button>
        <button @click="fetchRealms" class="btn btn-outline-primary" :disabled="loading">
          <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i> Refresh
        </button>
      </div>
    </div>

    <!-- Statistics Cards -->
    <div class="row g-3 mb-4">
      <div class="col-md-4">
        <div class="card stats-card bg-primary text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ realmStore.realmsCount }}</h3>
                <small>Total Realms</small>
              </div>
              <i class="fas fa-building fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card stats-card bg-success text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ realmStore.totalUsersAcrossRealms }}</h3>
                <small>Total Users</small>
              </div>
              <i class="fas fa-users fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card stats-card bg-info text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ averageUsersPerRealm }}</h3>
                <small>Avg Users/Realm</small>
              </div>
              <i class="fas fa-chart-bar fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="alert alert-danger alert-dismissible fade show" role="alert">
      <i class="fas fa-exclamation-triangle"></i> {{ error }}
      <button type="button" class="btn-close" @click="clearError"></button>
    </div>

    <!-- Realms Table -->
    <div class="card">
      <div class="card-header">
        <h5 class="mb-0">All Realms</h5>
      </div>
      <div class="card-body">
        <!-- Loading State -->
        <div v-if="loading && realms.length === 0" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Loading...</span>
          </div>
          <p class="mt-3 text-muted">Loading realms...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="!loading && realms.length === 0" class="text-center py-5">
          <i class="fas fa-building fa-3x text-muted mb-3"></i>
          <h5>No Realms Found</h5>
          <p class="text-muted">Create your first realm to get started</p>
          <button @click="showCreateModal" class="btn btn-primary">
            <i class="fas fa-plus"></i> Create Realm
          </button>
        </div>

        <!-- Table -->
        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Users</th>
                <th>Created By</th>
                <th>Created At</th>
                <th class="text-center">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="realm in realms" :key="realm.id">
                <td>
                  <strong>{{ realm.name }}</strong>
                  <span
                    v-if="realm.id === 'system'"
                    class="badge bg-danger ms-2"
                    title="System realm for super admins"
                  >
                    System
                  </span>
                </td>
                <td>{{ realm.description || '-' }}</td>
                <td>
                  <span class="badge bg-primary">{{ realm.user_count }} users</span>
                </td>
                <td>
                  <small>{{ realm.created_by }}</small>
                </td>
                <td>
                  <small>{{ formatDate(realm.created_at) }}</small>
                </td>
                <td class="text-center">
                  <div class="btn-group btn-group-sm">
                    <button
                      @click="viewRealmUsers(realm)"
                      class="btn btn-info"
                      title="View users in this realm"
                    >
                      <i class="fas fa-users"></i>
                    </button>
                    <button
                      @click="showEditModal(realm)"
                      class="btn btn-warning"
                      title="Edit realm"
                      :disabled="realm.id === 'system'"
                    >
                      <i class="fas fa-edit"></i>
                    </button>
                    <button
                      @click="showDeleteConfirmation(realm)"
                      class="btn btn-danger"
                      title="Delete realm"
                      :disabled="realm.id === 'system' || realm.user_count > 0"
                    >
                      <i class="fas fa-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div
      v-if="modalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              <i class="fas" :class="isEditMode ? 'fa-edit' : 'fa-plus'"></i>
              {{ isEditMode ? 'Edit Realm' : 'Create Realm' }}
            </h5>
            <button type="button" class="btn-close" @click="closeModal"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveRealm">
              <div class="mb-3">
                <label class="form-label">
                  <i class="fas fa-tag"></i> Realm Name
                  <span class="text-danger">*</span>
                </label>
                <input
                  v-model="realmForm.name"
                  type="text"
                  class="form-control"
                  placeholder="Enter realm name"
                  required
                  maxlength="100"
                />
                <small class="text-muted">
                  A unique name for this organizational realm
                </small>
              </div>

              <div class="mb-3">
                <label class="form-label">
                  <i class="fas fa-align-left"></i> Description
                </label>
                <textarea
                  v-model="realmForm.description"
                  class="form-control"
                  rows="3"
                  placeholder="Enter realm description (optional)"
                  maxlength="500"
                ></textarea>
                <small class="text-muted">
                  Describe the purpose or scope of this realm
                </small>
              </div>

              <div v-if="modalError" class="alert alert-danger">
                <i class="fas fa-exclamation-triangle"></i> {{ modalError }}
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeModal"
              :disabled="saving"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="saveRealm"
              :disabled="!realmForm.name || saving"
            >
              <span v-if="saving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas" :class="isEditMode ? 'fa-save' : 'fa-plus'"></i>
              {{ saving ? 'Saving...' : (isEditMode ? 'Update Realm' : 'Create Realm') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div
      v-if="deleteModalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header bg-danger text-white">
            <h5 class="modal-title">
              <i class="fas fa-exclamation-triangle"></i> Confirm Deletion
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              @click="closeDeleteModal"
            ></button>
          </div>
          <div class="modal-body">
            <p v-if="selectedRealm">
              Are you sure you want to delete the realm
              <strong>{{ selectedRealm.name }}</strong>?
            </p>
            <div class="alert alert-warning">
              <i class="fas fa-info-circle"></i>
              This action cannot be undone. The realm must have no users before it can
              be deleted.
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeDeleteModal"
              :disabled="deleting"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-danger"
              @click="deleteRealm"
              :disabled="deleting"
            >
              <span v-if="deleting" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-trash"></i>
              {{ deleting ? 'Deleting...' : 'Delete Realm' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRealmStore } from '@/stores/realmStore'
import type { RealmWithUserCount } from '@/services/realmService'
import { showToast } from '@/utils/toast'

const router = useRouter()
const realmStore = useRealmStore()

// State
const loading = ref(true) // Start true to prevent blank flash on first load
const saving = ref(false)
const deleting = ref(false)
const modalVisible = ref(false)
const deleteModalVisible = ref(false)
const isEditMode = ref(false)
const selectedRealm = ref<RealmWithUserCount | null>(null)
const realmForm = ref({
  name: '',
  description: ''
})
const modalError = ref<string | null>(null)

// Computed
const realms = computed(() => realmStore.realms)
const error = computed(() => realmStore.error)
const averageUsersPerRealm = computed(() => {
  if (realmStore.realmsCount === 0) return 0
  return Math.round(realmStore.totalUsersAcrossRealms / realmStore.realmsCount)
})

// Methods
async function fetchRealms() {
  loading.value = true
  try {
    await realmStore.fetchRealms()
  } catch (err) {
    console.error('Failed to fetch realms:', err)
  } finally {
    loading.value = false
  }
}

function showCreateModal() {
  isEditMode.value = false
  selectedRealm.value = null
  realmForm.value = { name: '', description: '' }
  modalError.value = null
  modalVisible.value = true
}

function showEditModal(realm: RealmWithUserCount) {
  isEditMode.value = true
  selectedRealm.value = realm
  realmForm.value = {
    name: realm.name,
    description: realm.description
  }
  modalError.value = null
  modalVisible.value = true
}

function closeModal() {
  modalVisible.value = false
  isEditMode.value = false
  selectedRealm.value = null
  realmForm.value = { name: '', description: '' }
  modalError.value = null
}

async function saveRealm() {
  saving.value = true
  modalError.value = null

  try {
    if (isEditMode.value && selectedRealm.value) {
      await realmStore.updateRealm(selectedRealm.value.id, realmForm.value)
      showToast('Realm updated successfully', 'success')
    } else {
      await realmStore.createRealm(realmForm.value)
      showToast('Realm created successfully', 'success')
    }
    closeModal()
  } catch (err: any) {
    modalError.value = err.response?.data?.error || err.message || 'Failed to save realm'
  } finally {
    saving.value = false
  }
}

function showDeleteConfirmation(realm: RealmWithUserCount) {
  selectedRealm.value = realm
  deleteModalVisible.value = true
}

function closeDeleteModal() {
  deleteModalVisible.value = false
  selectedRealm.value = null
}

async function deleteRealm() {
  if (!selectedRealm.value) return

  deleting.value = true

  try {
    await realmStore.deleteRealm(selectedRealm.value.id)
    showToast('Realm deleted successfully', 'info')
    closeDeleteModal()
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to delete realm', 'danger')
  } finally {
    deleting.value = false
  }
}

function viewRealmUsers(realm: RealmWithUserCount) {
  router.push({ name: 'admin-users', query: { realm_id: realm.id } })
}

function clearError() {
  realmStore.clearError()
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString()
}

// Lifecycle
onMounted(() => {
  fetchRealms()
})
</script>

<style scoped>
.stats-card {
  transition: transform 0.2s;
}

.stats-card:hover {
  transform: translateY(-2px);
}

.table-responsive {
  max-height: 600px;
  overflow-y: auto;
}

.modal.show {
  display: block;
}
</style>
