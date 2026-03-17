import axios from 'axios';
import { storage } from '@/utils/storage';

const env = import.meta.env.VITE_API_BASE_URL || '';
const API_BASE_URL = env ? env.replace(/\/api\/v1\/?$/, '') + '/api/v1' : '/api/v1';

export interface SignUpRequest {
  username: string;
  email: string;
  password: string;
  name?: string;
}

export interface SignInRequest {
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: string;
    username: string;
    email: string;
    name: string;
    realm_id: string;
  };
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

class AuthService {
  /**
   * Sign up a new user with username/password
   */
  async signUp(request: SignUpRequest): Promise<AuthResponse> {
    const response = await axios.post<AuthResponse>(
      `${API_BASE_URL}/auth/signup`,
      request
    );
    return response.data;
  }

  /**
   * Sign in an existing user with username/password
   */
  async signIn(request: SignInRequest): Promise<AuthResponse> {
    const response = await axios.post<AuthResponse>(
      `${API_BASE_URL}/auth/signin`,
      request
    );
    return response.data;
  }

  /**
   * Change password for authenticated user
   */
  async changePassword(request: ChangePasswordRequest): Promise<void> {
    const token = storage.getItem('auth_token');
    if (!token) {
      throw new Error('Not authenticated');
    }

    await axios.post(
      `${API_BASE_URL}/auth/change-password`,
      request,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
  }

  /**
   * Get current authenticated user
   */
  async getCurrentUser(): Promise<AuthResponse['user']> {
    const token = storage.getItem('auth_token');
    if (!token) {
      throw new Error('Not authenticated');
    }

    const response = await axios.get<AuthResponse['user']>(
      `${API_BASE_URL}/auth/user`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    return response.data;
  }

  /**
   * Logout (client-side only for now)
   */
  async logout(): Promise<void> {
    const token = storage.getItem('auth_token');
    if (token) {
      try {
        await axios.post(
          `${API_BASE_URL}/auth/logout`,
          {},
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );
      } catch (error) {
        console.error('Logout error:', error);
      }
    }

    // Clear local storage
    storage.removeItem('auth_token');
    storage.removeItem('user');
    storage.setItem('auth_mode', 'guest');
  }
}

export const authService = new AuthService();
