<template>
  <div class="container mt-4 mb-5">
    <!-- Header -->
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-users"></i> User Management
        </h2>
        <p class="text-muted mb-0">
          <small>Manage all users across realms</small>
        </p>
      </div>
      <button @click="fetchUsers" class="btn btn-outline-primary" :disabled="loading">
        <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i> Refresh
      </button>
    </div>

    <!-- Statistics Cards -->
    <div class="row g-3 mb-4">
      <div class="col-md-3">
        <div class="card stats-card bg-success text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ userManagementStore.activeUsersCount }}</h3>
                <small>Active Users</small>
              </div>
              <i class="fas fa-user-check fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-warning text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ userManagementStore.inactiveUsersCount }}</h3>
                <small>Inactive Users</small>
              </div>
              <i class="fas fa-user-slash fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-info text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ filteredUsers.length }}</h3>
                <small>Filtered Users</small>
              </div>
              <i class="fas fa-filter fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
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
    </div>

    <!-- Filters -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row g-3">
          <div class="col-md-4">
            <label class="form-label"><i class="fas fa-user-check"></i> Status</label>
            <select v-model="filters.status" class="form-select" @change="applyFilters">
              <option value="all">All Users</option>
              <option value="active">Active Only</option>
              <option value="inactive">Inactive Only</option>
            </select>
          </div>
          <div class="col-md-4">
            <label class="form-label"><i class="fas fa-building"></i> Realm</label>
            <select v-model="filters.realm_id" class="form-select" @change="applyFilters">
              <option value="">All Realms</option>
              <option v-for="realm in realms" :key="realm.id" :value="realm.id">
                {{ realm.name }}
              </option>
            </select>
          </div>
          <div class="col-md-4">
            <label class="form-label"><i class="fas fa-search"></i> Search</label>
            <input
              v-model="searchQuery"
              type="text"
              class="form-control"
              placeholder="Search username or email..."
            />
          </div>
        </div>
      </div>
    </div>

    <!--Error Alert -->
    <div v-if="error" class="alert alert-danger alert-dismissible fade show" role="alert">
      <i class="fas fa-exclamation-triangle"></i> {{ error }}
      <button type="button" class="btn-close" @click="clearError"></button>
    </div>

    <!-- Users Table -->
    <div class="card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0">All Users ({{ filteredUsers.length }})</h5>
      </div>
      <div class="card-body">
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Loading...</span>
          </div>
          <p class="mt-3 text-muted">Loading users...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredUsers.length === 0" class="text-center py-5">
          <i class="fas fa-users fa-3x text-muted mb-3"></i>
          <h5>No Users Found</h5>
          <p class="text-muted">No users match your current filters</p>
        </div>

        <!-- Table -->
        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>User</th>
                <th>Realm</th>
                <th>Roles</th>
                <th>Status</th>
                <th>Last Login</th>
                <th class="text-center">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in filteredUsers" :key="user.id">
                <td>
                  <div class="d-flex align-items-center">
                    <img
                      v-if="user.avatar_url"
                      :src="user.avatar_url"
                      alt="Avatar"
                      class="rounded-circle me-2"
                      width="32"
                      height="32"
                    />
                    <i v-else class="fas fa-user-circle fa-2x text-secondary me-2"></i>
                    <div>
                      <strong>{{ user.username }}</strong>
                      <br />
                      <small class="text-muted">{{ user.email }}</small>
                    </div>
                  </div>
                </td>
                <td>
                  <span class="badge bg-info">{{ user.realm_name || 'No Realm' }}</span>
                </td>
                <td>
                  <span
                    v-for="role in user.roles"
                    :key="role.id"
                    class="badge me-1"
                    :class="getRoleBadgeClass(role.name)"
                  >
                    {{ role.name }}
                  </span>
                  <span v-if="user.roles.length === 0" class="text-muted">
                    No roles
                  </span>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="user.is_active ? 'bg-success' : 'bg-secondary'"
                  >
                    {{ user.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td>
                  <small>{{ user.last_login ? formatDate(user.last_login) : 'Never' }}</small>
                </td>
                <td class="text-center">
                  <div class="btn-group btn-group-sm">
                    <button
                      @click="showChangeRealmModal(user)"
                      class="btn btn-info"
                      title="Change realm"
                      :disabled="!user.is_active"
                    >
                      <i class="fas fa-building"></i>
                    </button>
                    <button
                      @click="showChangeRoleModal(user)"
                      class="btn btn-warning"
                      title="Change role"
                      :disabled="!user.is_active"
                    >
                      <i class="fas fa-user-tag"></i>
                    </button>
                    <button
                      @click="showDeactivateConfirmation(user)"
                      class="btn btn-danger"
                      title="Deactivate user"
                      :disabled="!user.is_active || hasRole(user, 'super_admin')"
                    >
                      <i class="fas fa-ban"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Change Realm Modal -->
    <div
      v-if="changeRealmModalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              <i class="fas fa-building"></i> Change User Realm
            </h5>
            <button type="button" class="btn-close" @click="closeChangeRealmModal"></button>
          </div>
          <div class="modal-body">
            <div v-if="selectedUser" class="mb-3">
              <div class="alert alert-info">
                <strong>User:</strong> {{ selectedUser.username }}
                <br />
                <strong>Current Realm:</strong> {{ selectedUser.realm_name || 'None' }}
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">Select New Realm</label>
              <select v-model="changeRealmForm.realm_id" class="form-select">
                <option value="">Select a realm...</option>
                <option v-for="realm in realms" :key="realm.id" :value="realm.id">
                  {{ realm.name }}
                </option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeChangeRealmModal"
              :disabled="changingRealm"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="changeUserRealm"
              :disabled="!changeRealmForm.realm_id || changingRealm"
            >
              <span v-if="changingRealm" class="spinner-border spinner-border-sm me-2"></span>
              Change Realm
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Change Role Modal -->
    <div
      v-if="changeRoleModalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              <i class="fas fa-user-tag"></i> Change User Role
            </h5>
            <button type="button" class="btn-close" @click="closeChangeRoleModal"></button>
          </div>
          <div class="modal-body">
            <div v-if="selectedUser" class="mb-3">
              <div class="alert alert-info">
                <strong>User:</strong> {{ selectedUser.username }}
                <br />
                <strong>Current Roles:</strong>
                <span v-for="role in selectedUser.roles" :key="role.id" class="badge bg-secondary me-1">
                  {{ role.name }}
                </span>
              </div>
            </div>

            <div class="mb-3">
              <label class="form-label">Select New Role</label>
              <select v-model="changeRoleForm.role_id" class="form-select">
                <option value="">Select a role...</option>
                <option value="role_user">User</option>
                <option value="role_admin">Admin</option>
                <option value="role_super_admin">Super Admin</option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeChangeRoleModal"
              :disabled="changingRole"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="changeUserRole"
              :disabled="!changeRoleForm.role_id || changingRole"
            >
              <span v-if="changingRole" class="spinner-border spinner-border-sm me-2"></span>
              Change Role
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Deactivate Confirmation Modal -->
    <div
      v-if="deactivateModalVisible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5)"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header bg-danger text-white">
            <h5 class="modal-title">
              <i class="fas fa-exclamation-triangle"></i> Confirm Deactivation
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              @click="closeDeactivateModal"
            ></button>
          </div>
          <div class="modal-body">
            <p v-if="selectedUser">
              Are you sure you want to deactivate the user
              <strong>{{ selectedUser.username }}</strong>?
            </p>
            <div class="alert alert-warning">
              <i class="fas fa-info-circle"></i>
              The user will no longer be able to sign in until reactivated.
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeDeactivateModal"
              :disabled="deactivating"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-danger"
              @click="deactivateUser"
              :disabled="deactivating"
            >
              <span v-if="deactivating" class="spinner-border spinner-border-sm me-2"></span>
              Deactivate User
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useUserManagementStore } from '@/stores/userManagementStore'
import { useRealmStore } from '@/stores/realmStore'
import type { UserListItem } from '@/services/userManagementService'
import { showToast } from '@/utils/toast'

const route = useRoute()
const userManagementStore = useUserManagementStore()
const realmStore = useRealmStore()

// State
const loading = ref(true) // Start true to prevent blank flash on first load
const changingRealm = ref(false)
const changingRole = ref(false)
const deactivating = ref(false)
const changeRealmModalVisible = ref(false)
const changeRoleModalVisible = ref(false)
const deactivateModalVisible = ref(false)
const selectedUser = ref<UserListItem | null>(null)
const changeRealmForm = ref({ realm_id: '' })
const changeRoleForm = ref({ role_id: '' })
const searchQuery = ref('')
const filters = ref({
  status: 'all',
  realm_id: ''
})

// Computed
const users = computed(() => userManagementStore.allUsers)
const realms = computed(() => realmStore.realms)
const error = computed(() => userManagementStore.error)

const filteredUsers = computed(() => {
  let result = users.value

  // Filter by status
  if (filters.value.status === 'active') {
    result = result.filter(u => u.is_active)
  } else if (filters.value.status === 'inactive') {
    result = result.filter(u => !u.is_active)
  }

  // Filter by realm
  if (filters.value.realm_id) {
    result = result.filter(u => u.realm_id === filters.value.realm_id)
  }

  // Search by username or email
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(
      u =>
        u.username.toLowerCase().includes(query) || u.email.toLowerCase().includes(query)
    )
  }

  return result
})

// Methods
async function fetchUsers() {
  loading.value = true
  try {
    await userManagementStore.fetchAllUsers()
    await realmStore.fetchRealms()
  } catch (err) {
    console.error('Failed to fetch users:', err)
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  // Filters are reactive, just trigger re-computation
}

function showChangeRealmModal(user: UserListItem) {
  selectedUser.value = user
  changeRealmForm.value.realm_id = user.realm_id
  changeRealmModalVisible.value = true
}

function closeChangeRealmModal() {
  changeRealmModalVisible.value = false
  selectedUser.value = null
  changeRealmForm.value.realm_id = ''
}

async function changeUserRealm() {
  if (!selectedUser.value) return

  changingRealm.value = true
  try {
    await userManagementStore.updateUserRealm(
      selectedUser.value.id,
      changeRealmForm.value
    )
    showToast('User realm updated successfully', 'success')
    closeChangeRealmModal()
    await fetchUsers()
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to update realm', 'danger')
  } finally {
    changingRealm.value = false
  }
}

function showChangeRoleModal(user: UserListItem) {
  selectedUser.value = user
  changeRoleForm.value.role_id = user.roles[0]?.id || ''
  changeRoleModalVisible.value = true
}

function closeChangeRoleModal() {
  changeRoleModalVisible.value = false
  selectedUser.value = null
  changeRoleForm.value.role_id = ''
}

async function changeUserRole() {
  if (!selectedUser.value) return

  changingRole.value = true
  try {
    await userManagementStore.updateUserRole(selectedUser.value.id, changeRoleForm.value)
    showToast('User role updated successfully', 'success')
    closeChangeRoleModal()
    await fetchUsers()
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to update role', 'danger')
  } finally {
    changingRole.value = false
  }
}

function showDeactivateConfirmation(user: UserListItem) {
  selectedUser.value = user
  deactivateModalVisible.value = true
}

function closeDeactivateModal() {
  deactivateModalVisible.value = false
  selectedUser.value = null
}

async function deactivateUser() {
  if (!selectedUser.value) return

  deactivating.value = true
  try {
    await userManagementStore.deactivateUser(selectedUser.value.id)
    showToast('User deactivated successfully', 'info')
    closeDeactivateModal()
    await fetchUsers()
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to deactivate user', 'danger')
  } finally {
    deactivating.value = false
  }
}

function getRoleBadgeClass(roleName: string): string {
  switch (roleName) {
    case 'super_admin':
      return 'bg-danger'
    case 'admin':
      return 'bg-warning text-dark'
    case 'user':
      return 'bg-primary'
    default:
      return 'bg-secondary'
  }
}

function hasRole(user: UserListItem, roleName: string): boolean {
  return user.roles.some(r => r.name === roleName)
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
  fetchUsers()

  // Check for realm filter from query params
  if (route.query.realm_id) {
    filters.value.realm_id = route.query.realm_id as string
  }
})

// Watch for route query changes
watch(
  () => route.query.realm_id,
  (newRealmId) => {
    if (newRealmId) {
      filters.value.realm_id = newRealmId as string
    }
  }
)
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
