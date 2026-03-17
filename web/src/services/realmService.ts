import { apiService } from './apiService'

export interface Realm {
  id: string
  name: string
  description: string
  created_by: string
  created_time: string
  updated_by: string
  updated_time: string
}

export interface RealmWithUserCount {
  id: string
  name: string
  description: string
  user_count: number
  created_by: string
  created_at: string
}

export interface CreateRealmRequest {
  name: string
  description: string
}

export interface UpdateRealmRequest {
  name?: string
  description?: string
}

export interface User {
  id: string
  username: string
  email: string
  name: string
  realm_id: string
  is_active: boolean
  created_time: string
  last_login_at: string | null
}

/**
 * Get all realms with user counts
 */
export async function getAllRealms(): Promise<RealmWithUserCount[]> {
  const response = await apiService.get('/admin/realms')
  return response.data
}

/**
 * Get a specific realm by ID
 */
export async function getRealmById(realmId: string): Promise<Realm> {
  const response = await apiService.get(`/admin/realms/${realmId}`)
  return response.data
}

/**
 * Create a new realm
 */
export async function createRealm(request: CreateRealmRequest): Promise<Realm> {
  const response = await apiService.post('/admin/realms', request)
  return response.data
}

/**
 * Update an existing realm
 */
export async function updateRealm(realmId: string, request: UpdateRealmRequest): Promise<Realm> {
  const response = await apiService.put(`/admin/realms/${realmId}`, request)
  return response.data
}

/**
 * Delete a realm (soft delete, must have no users)
 */
export async function deleteRealm(realmId: string): Promise<{ message: string }> {
  const response = await apiService.delete(`/admin/realms/${realmId}`)
  return response.data
}

/**
 * Get all active users in a realm
 */
export async function getUsersInRealm(realmId: string): Promise<User[]> {
  const response = await apiService.get(`/admin/realms/${realmId}/users`)
  return response.data
}

export default {
  getAllRealms,
  getRealmById,
  createRealm,
  updateRealm,
  deleteRealm,
  getUsersInRealm
}
