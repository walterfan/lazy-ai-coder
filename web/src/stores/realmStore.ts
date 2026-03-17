import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as realmService from '@/services/realmService'
import type { Realm, RealmWithUserCount, CreateRealmRequest, UpdateRealmRequest, User } from '@/services/realmService'

export const useRealmStore = defineStore('realm', () => {
  // State
  const realms = ref<RealmWithUserCount[]>([])
  const currentRealm = ref<Realm | null>(null)
  const realmUsers = ref<User[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const realmsCount = computed(() => realms.value.length)
  const totalUsersAcrossRealms = computed(() =>
    realms.value.reduce((sum, realm) => sum + realm.user_count, 0)
  )

  // Get realm by ID from cached list
  const getRealmById = computed(() => (realmId: string) => {
    return realms.value.find(r => r.id === realmId)
  })

  // Actions
  async function fetchRealms() {
    loading.value = true
    error.value = null
    try {
      realms.value = await realmService.getAllRealms()
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to fetch realms'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchRealmDetails(realmId: string) {
    loading.value = true
    error.value = null
    try {
      currentRealm.value = await realmService.getRealmById(realmId)
      return currentRealm.value
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to fetch realm details'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchRealmUsers(realmId: string) {
    loading.value = true
    error.value = null
    try {
      realmUsers.value = await realmService.getUsersInRealm(realmId)
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to fetch realm users'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createRealm(request: CreateRealmRequest) {
    loading.value = true
    error.value = null
    try {
      const newRealm = await realmService.createRealm(request)

      // Add to realms list with user_count = 0
      realms.value.unshift({
        ...newRealm,
        user_count: 0,
        created_at: newRealm.created_time
      })

      return newRealm
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to create realm'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateRealm(realmId: string, request: UpdateRealmRequest) {
    loading.value = true
    error.value = null
    try {
      const updatedRealm = await realmService.updateRealm(realmId, request)

      // Update in realms list
      const index = realms.value.findIndex(r => r.id === realmId)
      if (index !== -1) {
        realms.value[index] = {
          ...realms.value[index],
          name: updatedRealm.name,
          description: updatedRealm.description
        }
      }

      // Update current realm if it's the one being edited
      if (currentRealm.value?.id === realmId) {
        currentRealm.value = updatedRealm
      }

      return updatedRealm
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to update realm'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteRealm(realmId: string) {
    loading.value = true
    error.value = null
    try {
      await realmService.deleteRealm(realmId)

      // Remove from realms list
      realms.value = realms.value.filter(r => r.id !== realmId)

      // Clear current realm if it's the deleted one
      if (currentRealm.value?.id === realmId) {
        currentRealm.value = null
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Failed to delete realm'
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  function clearCurrentRealm() {
    currentRealm.value = null
  }

  function clearRealmUsers() {
    realmUsers.value = []
  }

  function reset() {
    realms.value = []
    currentRealm.value = null
    realmUsers.value = []
    loading.value = false
    error.value = null
  }

  return {
    // State
    realms,
    currentRealm,
    realmUsers,
    loading,
    error,

    // Computed
    realmsCount,
    totalUsersAcrossRealms,
    getRealmById,

    // Actions
    fetchRealms,
    fetchRealmDetails,
    fetchRealmUsers,
    createRealm,
    updateRealm,
    deleteRealm,
    clearError,
    clearCurrentRealm,
    clearRealmUsers,
    reset
  }
})
