import { apiService } from './apiService'

export interface RoleInfo {
  id: string
  name: string
}

export interface UserListItem {
  id: string
  username: string
  email: string
  name: string
  realm_id: string
  realm_name: string
  is_active: boolean
  roles: RoleInfo[]
  created_at: string
  last_login: string | null
  avatar_url?: string
}

export interface UserApprovalRequest {
  realm_id: string
  role_id: string
}

export interface UserUpdateRealmRequest {
  realm_id: string
}

export interface UserUpdateRoleRequest {
  role_id: string
}

/**
 * Get all pending users (is_active=false)
 */
export async function getPendingUsers(): Promise<UserListItem[]> {
  const response = await apiService.get('/admin/pending-users')
  return response.data
}

/**
 * Approve a pending user and assign realm and role
 */
export async function approveUser(userId: string, request: UserApprovalRequest): Promise<UserListItem> {
  const response = await apiService.post(`/admin/users/${userId}/approve`, request)
  return response.data
}

/**
 * Reject a pending user
 */
export async function rejectUser(userId: string): Promise<{ message: string }> {
  const response = await apiService.post(`/admin/users/${userId}/reject`)
  return response.data
}

/**
 * Get all users with optional filters
 */
export async function getAllUsers(filters?: {
  is_active?: boolean
  realm_id?: string
}): Promise<UserListItem[]> {
  const params = new URLSearchParams()
  if (filters?.is_active !== undefined) {
    params.append('is_active', String(filters.is_active))
  }
  if (filters?.realm_id) {
    params.append('realm_id', filters.realm_id)
  }

  const query = params.toString()
  const url = query ? `/admin/users?${query}` : '/admin/users'
  const response = await apiService.get(url)
  return response.data
}

/**
 * Update user's realm
 */
export async function updateUserRealm(
  userId: string,
  request: UserUpdateRealmRequest
): Promise<{ message: string }> {
  const response = await apiService.put(`/admin/users/${userId}/realm`, request)
  return response.data
}

/**
 * Update user's role
 */
export async function updateUserRole(
  userId: string,
  request: UserUpdateRoleRequest
): Promise<{ message: string }> {
  const response = await apiService.put(`/admin/users/${userId}/role`, request)
  return response.data
}

/**
 * Deactivate an active user
 */
export async function deactivateUser(userId: string): Promise<{ message: string }> {
  const response = await apiService.post(`/admin/users/${userId}/deactivate`)
  return response.data
}

export default {
  getPendingUsers,
  approveUser,
  rejectUser,
  getAllUsers,
  updateUserRealm,
  updateUserRole,
  deactivateUser
}
