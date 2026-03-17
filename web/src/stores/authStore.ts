import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { useSettingsStore } from './settingsStore';
import { getTokenExpiration, isTokenExpired, isTokenExpiringSoon } from '@/utils/jwtUtils';
import { safeJsonParse, storage } from '@/utils/storage';

export interface RoleInfo {
  id: string;
  name: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  realm_id: string;
  is_active?: boolean;
  roles?: RoleInfo[]; // User's assigned roles
  // OAuth-specific fields (optional)
  gitlab_user_id?: number;
  avatar_url?: string;
}

export type AuthMode = 'guest' | 'password' | 'oauth';

export type AuthWarningType = 'expired' | 'unauthorized' | 'guest';

export interface AuthWarningDialog {
  show: boolean;
  type: AuthWarningType;
  message: string;
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);
  const tokenExpiry = ref<number | null>(null); // Token expiration timestamp in milliseconds
  const mode = ref<AuthMode>('guest'); // Default to guest mode (read-only)
  const loading = ref(false);
  const error = ref<string | null>(null);
  const refreshInterval = ref<number | null>(null); // Interval ID for token refresh checks

  // Auth warning dialog state
  const authWarning = ref<AuthWarningDialog>({
    show: false,
    type: 'expired',
    message: '',
  });

  // Getters
  const isGuest = computed(() => mode.value === 'guest');
  const isPasswordMode = computed(() => mode.value === 'password');
  const isOAuthMode = computed(() => mode.value === 'oauth');

  // Authentication status
  const isAuthenticated = computed(() => {
    // OAuth or password mode with valid token/user
    if (isOAuthMode.value || isPasswordMode.value) {
      return !!token.value && !!user.value;
    }
    // Guest mode is NOT authenticated
    return false;
  });

  // Permission checks
  const canModify = computed(() => isAuthenticated.value);
  const canView = computed(() => true); // Everyone can view (guest + authenticated)

  // Guest users can use the app with their own API keys
  const hasCredentials = computed(() => {
    const settings = useSettingsStore();
    return !!settings.GITLAB_TOKEN || !!settings.LLM_API_KEY;
  });

  // Role-based computed properties
  const userRoles = computed(() => user.value?.roles || []);
  const userRoleNames = computed(() => userRoles.value.map(r => r.name));

  const isSuperAdmin = computed(() => {
    return userRoleNames.value.includes('super_admin');
  });

  const isAdmin = computed(() => {
    return userRoleNames.value.includes('admin') || isSuperAdmin.value;
  });

  const isUser = computed(() => {
    return userRoleNames.value.includes('user');
  });

  const hasRole = (roleName: string) => {
    return userRoleNames.value.includes(roleName);
  };

  // Helper: Update token and expiry
  function setTokenWithExpiry(newToken: string) {
    token.value = newToken;
    const expiry = getTokenExpiration(newToken);
    tokenExpiry.value = expiry;
  }

  // Helper: Check if token needs refresh
  function shouldRefreshToken(): boolean {
    if (!token.value || !tokenExpiry.value) return false;

    // Check if token is expired or expiring soon (within 5 minutes)
    if (isTokenExpired(token.value)) {
      return false; // Already expired, can't refresh
    }

    return isTokenExpiringSoon(token.value, 5); // Refresh if expires in < 5 minutes
  }

  // Actions
  function loadFromStorage() {
    // Try to load authenticated session first (OAuth or password)
    const storedToken = storage.getItem('auth_token');
    const userStr = storage.getItem('user');
    const storedMode = storage.getItem('auth_mode') as AuthMode;

    if (storedToken && userStr && (storedMode === 'oauth' || storedMode === 'password')) {
      // Check if token is expired
      if (isTokenExpired(storedToken)) {
        // Token expired, switch to guest mode
        console.log('Stored token is expired, switching to guest mode');
        switchToGuestMode();
        return;
      }

      token.value = storedToken;
      const parsedUser = safeJsonParse<User>(userStr);
      if (!parsedUser) {
        console.warn('Failed to parse stored user, switching to guest mode');
        switchToGuestMode();
        return;
      }
      user.value = parsedUser;
      mode.value = storedMode;

      // Set token expiry
      const expiry = getTokenExpiration(storedToken);
      tokenExpiry.value = expiry;

      // Start refresh interval
      startTokenRefreshCheck();
    } else {
      // Default to guest mode
      mode.value = 'guest';
      storage.setItem('auth_mode', 'guest');
    }
  }

  function saveToStorage() {
    if ((mode.value === 'oauth' || mode.value === 'password') && token.value && user.value) {
      storage.setItem('auth_token', token.value);
      storage.setItem('user', JSON.stringify(user.value));
      storage.setItem('auth_mode', mode.value);
    } else if (mode.value === 'guest') {
      storage.setItem('auth_mode', 'guest');
    }
  }

  async function handleOAuthCallback(code: string, state: string) {
    try {
      loading.value = true;
      error.value = null;

      const response = await fetch(
        `/api/v1/auth/gitlab/callback?code=${code}&state=${state}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        }
      );

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'OAuth callback failed');
      }

      const data = await response.json();

      switchToOAuthMode(data.token, data.user);
      return true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'OAuth login failed';
      console.error('OAuth callback error:', err);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function fetchCurrentUser() {
    if (!token.value || mode.value !== 'oauth') return;

    try {
      const response = await fetch('/api/v1/auth/user', {
        headers: {
          Authorization: `Bearer ${token.value}`,
        },
      });

      if (response.ok) {
        user.value = await response.json();
        saveToStorage();
      } else {
        // Token invalid or expired
        logout();
      }
    } catch (err) {
      console.error('Failed to fetch user:', err);
      logout();
    }
  }

  async function fetchUserRoles() {
    if (!token.value || !isAuthenticated.value) return;

    try {
      const response = await fetch('/api/v1/profile/roles', {
        headers: {
          Authorization: `Bearer ${token.value}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        // Convert string array to RoleInfo array
        const roleInfos: RoleInfo[] = (data.roles || []).map((roleName: string) => ({
          id: roleName,
          name: roleName,
        }));

        // Update user with roles
        if (user.value) {
          user.value = {
            ...user.value,
            roles: roleInfos,
          };
          saveToStorage();
        }
      }
    } catch (err) {
      console.error('Failed to fetch user roles:', err);
    }
  }

  // Refresh token proactively
  async function refreshToken(): Promise<boolean> {
    if (!token.value) {
      console.log('No token to refresh');
      return false;
    }

    try {
      console.log('Refreshing token...');
      const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token.value}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        console.error('Token refresh failed:', response.status);
        // Token refresh failed, logout user
        logout();
        return false;
      }

      const data = await response.json();

      // Update token and user data
      setTokenWithExpiry(data.token);
      user.value = data.user;
      saveToStorage();

      console.log('Token refreshed successfully');
      return true;
    } catch (err) {
      console.error('Token refresh error:', err);
      logout();
      return false;
    }
  }

  // Start periodic token refresh check
  function startTokenRefreshCheck() {
    // Clear any existing interval
    stopTokenRefreshCheck();

    // Check every minute
    refreshInterval.value = window.setInterval(async () => {
      if (shouldRefreshToken()) {
        console.log('Token expiring soon, refreshing...');
        await refreshToken();
      }
    }, 60 * 1000); // Check every minute
  }

  // Stop token refresh check
  function stopTokenRefreshCheck() {
    if (refreshInterval.value !== null) {
      clearInterval(refreshInterval.value);
      refreshInterval.value = null;
    }
  }

  async function switchToOAuthMode(jwtToken: string, userData: User) {
    setTokenWithExpiry(jwtToken);
    user.value = userData;
    mode.value = 'oauth';
    saveToStorage();
    // Fetch roles after successful login
    await fetchUserRoles();
    // Start refresh check
    startTokenRefreshCheck();
  }

  async function switchToPasswordMode(jwtToken: string, userData: User) {
    setTokenWithExpiry(jwtToken);
    user.value = userData;
    mode.value = 'password';
    saveToStorage();
    // Fetch roles after successful login
    await fetchUserRoles();
    // Start refresh check
    startTokenRefreshCheck();
  }

  function switchToGuestMode() {
    // Stop refresh check
    stopTokenRefreshCheck();
    // Clear authentication data
    token.value = null;
    tokenExpiry.value = null;
    user.value = null;
    mode.value = 'guest';
    storage.removeItem('auth_token');
    storage.removeItem('user');
    storage.setItem('auth_mode', 'guest');
  }

  function logout() {
    // Stop refresh check
    stopTokenRefreshCheck();
    token.value = null;
    tokenExpiry.value = null;
    user.value = null;
    mode.value = 'guest'; // Return to guest mode after logout
    storage.removeItem('auth_token');
    storage.removeItem('user');
    storage.setItem('auth_mode', 'guest');
    error.value = null;
  }

  // Show auth warning dialog
  function showAuthWarning(type: AuthWarningType, message?: string) {
    const defaultMessages: Record<AuthWarningType, string> = {
      expired: 'Your session has expired. Please login again to continue.',
      unauthorized: 'You need to be logged in to access this feature.',
      guest: 'This feature is only available for registered users.',
    };

    authWarning.value = {
      show: true,
      type,
      message: message || defaultMessages[type],
    };
  }

  // Hide auth warning dialog
  function hideAuthWarning() {
    authWarning.value.show = false;
  }

  return {
    // State
    user,
    token,
    tokenExpiry,
    mode,
    loading,
    error,
    authWarning,

    // Getters
    isGuest,
    isPasswordMode,
    isOAuthMode,
    isAuthenticated,
    canModify,
    canView,
    hasCredentials,

    // Role-based getters
    userRoles,
    userRoleNames,
    isSuperAdmin,
    isAdmin,
    isUser,
    hasRole,

    // Actions
    loadFromStorage,
    saveToStorage,
    handleOAuthCallback,
    fetchCurrentUser,
    fetchUserRoles,
    switchToOAuthMode,
    switchToPasswordMode,
    switchToGuestMode,
    logout,
    refreshToken,
    startTokenRefreshCheck,
    stopTokenRefreshCheck,
    showAuthWarning,
    hideAuthWarning,
  };
});
