import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as userManagementService from '@/services/userManagementService'
import type { UserListItem, UserApprovalRequest, UserUpdateRealmRequest, UserUpdateRoleRequest } from '@/services/userManagementService'

export const useUserManagementStore = defineStore('userManagement', () => {
  // State
  const pendingUsers = ref<UserListItem[]>([])
  const allUsers = ref<UserListItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const pendingUsersCount = computed(() => pendingUsers.value.length)
  const activeUsersCount = computed(() => allUsers.value.filter(u => u.is_active).length)
  const inactiveUsersCount = computed(() => allUsers.value.filter(u => !u.is_active).length)

  // Actions
  async function fetchPendingUsers() {
    loading.value = true
    error.value = null
    try {
      pendingUsers.value = await userManagementService.getPendingUsers()
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to fetch pending users'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchAllUsers(filters?: { is_active?: boolean; realm_id?: string }) {
    loading.value = true
    error.value = null
    try {
      allUsers.value = await userManagementService.getAllUsers(filters)
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to fetch users'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function approveUser(userId: string, request: UserApprovalRequest) {
    loading.value = true
    error.value = null
    try {
      const approvedUser = await userManagementService.approveUser(userId, request)

      // Remove from pending list
      pendingUsers.value = pendingUsers.value.filter(u => u.id !== userId)

      // Add to all users list if it's loaded
      if (allUsers.value.length > 0) {
        allUsers.value.unshift(approvedUser)
      }

      return approvedUser
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to approve user'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function rejectUser(userId: string) {
    loading.value = true
    error.value = null
    try {
      await userManagementService.rejectUser(userId)

      // Remove from pending list
      pendingUsers.value = pendingUsers.value.filter(u => u.id !== userId)
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to reject user'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateUserRealm(userId: string, request: UserUpdateRealmRequest) {
    loading.value = true
    error.value = null
    try {
      await userManagementService.updateUserRealm(userId, request)

      // Update in allUsers list
      const user = allUsers.value.find(u => u.id === userId)
      if (user) {
        user.realm_id = request.realm_id
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to update user realm'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateUserRole(userId: string, request: UserUpdateRoleRequest) {
    loading.value = true
    error.value = null
    try {
      await userManagementService.updateUserRole(userId, request)

      // Refresh users to get updated role information
      await fetchAllUsers()
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to update user role'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deactivateUser(userId: string) {
    loading.value = true
    error.value = null
    try {
      await userManagementService.deactivateUser(userId)

      // Update in allUsers list
      const user = allUsers.value.find(u => u.id === userId)
      if (user) {
        user.is_active = false
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to deactivate user'
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  function reset() {
    pendingUsers.value = []
    allUsers.value = []
    loading.value = false
    error.value = null
  }

  return {
    // State
    pendingUsers,
    allUsers,
    loading,
    error,

    // Computed
    pendingUsersCount,
    activeUsersCount,
    inactiveUsersCount,

    // Actions
    fetchPendingUsers,
    fetchAllUsers,
    approveUser,
    rejectUser,
    updateUserRealm,
    updateUserRole,
    deactivateUser,
    clearError,
    reset
  }
})
