<template>
  <div class="container mt-4 mb-5">
    <!-- Header -->
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-user-clock"></i> Pending User Approvals
        </h2>
        <p class="text-muted mb-0">
          <small>Review and approve new user registrations</small>
        </p>
      </div>
      <button @click="fetchPendingUsers" class="btn btn-outline-primary" :disabled="loading">
        <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i> Refresh
      </button>
    </div>

    <!-- Statistics Cards -->
    <div class="row g-3 mb-4">
      <div class="col-md-4">
        <div class="card stats-card bg-warning text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ userManagementStore.pendingUsersCount }}</h3>
                <small>Pending Approvals</small>
              </div>
              <i class="fas fa-user-clock fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card stats-card bg-success text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ approvedCount }}</h3>
                <small>Approved Today</small>
              </div>
              <i class="fas fa-user-check fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card stats-card bg-danger text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ rejectedCount }}</h3>
                <small>Rejected Today</small>
              </div>
              <i class="fas fa-user-times fa-2x opacity-50"></i>
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

    <!-- Pending Users Table -->
    <div class="card">
      <div class="card-header">
        <h5 class="mb-0">Pending User Registrations</h5>
      </div>
      <div class="card-body">
        <!-- Loading State -->
        <div v-if="loading && pendingUsers.length === 0" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Loading...</span>
          </div>
          <p class="mt-3 text-muted">Loading pending users...</p>
        </div>

        <!-- Empty State -->
        <div
          v-else-if="!loading && pendingUsers.length === 0"
          class="text-center py-5"
        >
          <i class="fas fa-check-circle fa-3x text-success mb-3"></i>
          <h5>No Pending Approvals</h5>
          <p class="text-muted">All user registrations have been processed</p>
        </div>

        <!-- Table -->
        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>Username</th>
                <th>Email</th>
                <th>Name</th>
                <th>Created At</th>
                <th>Avatar</th>
                <th class="text-center">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in pendingUsers" :key="user.id">
                <td>
                  <strong>{{ user.username }}</strong>
                </td>
                <td>
                  <i class="fas fa-envelope text-muted"></i>
                  {{ user.email }}
                </td>
                <td>{{ user.name || '-' }}</td>
                <td>
                  <small>{{ formatDate(user.created_at) }}</small>
                </td>
                <td>
                  <img
                    v-if="user.avatar_url"
                    :src="user.avatar_url"
                    alt="Avatar"
                    class="rounded-circle"
                    width="32"
                    height="32"
                  />
                  <i v-else class="fas fa-user-circle fa-2x text-secondary"></i>
                </td>
                <td class="text-center">
                  <div class="btn-group btn-group-sm">
                    <button
                      @click="showApproveModal(user)"
                      class="btn btn-success"
                      title="Approve user"
                    >
                      <i class="fas fa-check"></i> Approve
                    </button>
                    <button
                      @click="showRejectConfirmation(user)"
                      class="btn btn-danger"
                      title="Reject user"
                    >
                      <i class="fas fa-times"></i> Reject
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Approve Modal -->
    <div
      v-if="approveModalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              <i class="fas fa-user-check"></i> Approve User
            </h5>
            <button
              type="button"
              class="btn-close"
              @click="closeApproveModal"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="selectedUser" class="mb-3">
              <div class="alert alert-info">
                <strong>User:</strong> {{ selectedUser.username }} ({{
                  selectedUser.email
                }})
              </div>
            </div>

            <!-- Realm Selection -->
            <div class="mb-3">
              <label class="form-label">
                <i class="fas fa-building"></i> Assign Realm
                <span class="text-danger">*</span>
              </label>
              <select v-model="approvalForm.realm_id" class="form-select" required>
                <option value="">Select a realm...</option>
                <option v-for="realm in realms" :key="realm.id" :value="realm.id">
                  {{ realm.name }} ({{ realm.user_count }} users)
                </option>
              </select>
              <small class="text-muted">
                The realm determines which organization/team the user belongs to
              </small>
            </div>

            <!-- Role Selection -->
            <div class="mb-3">
              <label class="form-label">
                <i class="fas fa-user-tag"></i> Assign Role
                <span class="text-danger">*</span>
              </label>
              <select v-model="approvalForm.role_id" class="form-select" required>
                <option value="">Select a role...</option>
                <option value="role_user">User (Standard permissions)</option>
                <option value="role_admin">Admin (Can manage their realm)</option>
                <option value="role_super_admin">
                  Super Admin (Full system access)
                </option>
              </select>
              <small class="text-muted">
                The role determines what permissions the user will have
              </small>
            </div>

            <div v-if="approvalError" class="alert alert-danger">
              <i class="fas fa-exclamation-triangle"></i> {{ approvalError }}
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeApproveModal"
              :disabled="approving"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-success"
              @click="approveUser"
              :disabled="
                !approvalForm.realm_id || !approvalForm.role_id || approving
              "
            >
              <span
                v-if="approving"
                class="spinner-border spinner-border-sm me-2"
              ></span>
              <i v-else class="fas fa-check"></i>
              {{ approving ? 'Approving...' : 'Approve User' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Reject Confirmation Modal -->
    <div
      v-if="rejectModalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header bg-danger text-white">
            <h5 class="modal-title">
              <i class="fas fa-exclamation-triangle"></i> Confirm Rejection
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              @click="closeRejectModal"
            ></button>
          </div>
          <div class="modal-body">
            <p v-if="selectedUser">
              Are you sure you want to reject the user registration for
              <strong>{{ selectedUser.username }}</strong> ({{ selectedUser.email }})?
            </p>
            <div class="alert alert-warning">
              <i class="fas fa-info-circle"></i>
              This action will permanently delete the user account. The user will need
              to sign up again.
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeRejectModal"
              :disabled="rejecting"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-danger"
              @click="rejectUser"
              :disabled="rejecting"
            >
              <span
                v-if="rejecting"
                class="spinner-border spinner-border-sm me-2"
              ></span>
              <i v-else class="fas fa-times"></i>
              {{ rejecting ? 'Rejecting...' : 'Reject User' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserManagementStore } from '@/stores/userManagementStore'
import { useRealmStore } from '@/stores/realmStore'
import type { UserListItem } from '@/services/userManagementService'
import { showToast } from '@/utils/toast'

const userManagementStore = useUserManagementStore()
const realmStore = useRealmStore()

// State
const loading = ref(true) // Start true to prevent blank flash on first load
const approving = ref(false)
const rejecting = ref(false)
const approveModalVisible = ref(false)
const rejectModalVisible = ref(false)
const selectedUser = ref<UserListItem | null>(null)
const approvalForm = ref({
  realm_id: '',
  role_id: ''
})
const approvalError = ref<string | null>(null)
const approvedCount = ref(0)
const rejectedCount = ref(0)

// Computed
const pendingUsers = computed(() => userManagementStore.pendingUsers)
const realms = computed(() => realmStore.realms)
const error = computed(() => userManagementStore.error)

// Methods
async function fetchPendingUsers() {
  loading.value = true
  try {
    await userManagementStore.fetchPendingUsers()
    await realmStore.fetchRealms() // Need realms for approval
  } catch (err) {
    console.error('Failed to fetch pending users:', err)
  } finally {
    loading.value = false
  }
}

function showApproveModal(user: UserListItem) {
  selectedUser.value = user
  approvalForm.value = {
    realm_id: '',
    role_id: 'role_user' // Default to user role
  }
  approvalError.value = null
  approveModalVisible.value = true
}

function closeApproveModal() {
  approveModalVisible.value = false
  selectedUser.value = null
  approvalForm.value = { realm_id: '', role_id: '' }
  approvalError.value = null
}

async function approveUser() {
  if (!selectedUser.value) return

  approving.value = true
  approvalError.value = null

  try {
    await userManagementStore.approveUser(selectedUser.value.id, approvalForm.value)
    showToast(`User ${selectedUser.value.username} approved successfully`, 'success')
    approvedCount.value++
    closeApproveModal()
  } catch (err: any) {
    approvalError.value = err.response?.data?.error || err.message || 'Failed to approve user'
  } finally {
    approving.value = false
  }
}

function showRejectConfirmation(user: UserListItem) {
  selectedUser.value = user
  rejectModalVisible.value = true
}

function closeRejectModal() {
  rejectModalVisible.value = false
  selectedUser.value = null
}

async function rejectUser() {
  if (!selectedUser.value) return

  rejecting.value = true

  try {
    await userManagementStore.rejectUser(selectedUser.value.id)
    showToast(`User ${selectedUser.value.username} rejected`, 'info')
    rejectedCount.value++
    closeRejectModal()
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to reject user', 'danger')
  } finally {
    rejecting.value = false
  }
}

function clearError() {
  userManagementStore.clearError()
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleString()
}

// Lifecycle
onMounted(() => {
  fetchPendingUsers()
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
