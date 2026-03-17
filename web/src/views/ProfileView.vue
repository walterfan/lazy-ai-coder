<template>
  <div class="container mt-4">
    <div class="row">
      <div class="col-md-8 offset-md-2">
        <h2 class="mb-4">
          <i class="bi bi-person-circle"></i> My Profile
        </h2>

        <!-- Alert messages -->
        <div v-if="successMessage" class="alert alert-success alert-dismissible fade show" role="alert">
          <i class="bi bi-check-circle"></i> {{ successMessage }}
          <button type="button" class="btn-close" @click="successMessage = ''"></button>
        </div>
        <div v-if="errorMessage" class="alert alert-danger alert-dismissible fade show" role="alert">
          <i class="bi bi-exclamation-triangle"></i> {{ errorMessage }}
          <button type="button" class="btn-close" @click="errorMessage = ''"></button>
        </div>

        <!-- Profile Information Card -->
        <div class="card mb-4">
          <div class="card-header">
            <h5 class="mb-0"><i class="bi bi-info-circle"></i> Profile Information</h5>
          </div>
          <div class="card-body">
            <form @submit.prevent="updateProfile">
              <div class="mb-3">
                <label for="username" class="form-label">Username</label>
                <input
                  type="text"
                  class="form-control"
                  id="username"
                  :value="profile.username"
                  disabled
                  readonly
                />
                <div class="form-text">Username cannot be changed</div>
              </div>

              <div class="mb-3">
                <label for="email" class="form-label">Email</label>
                <input
                  type="email"
                  class="form-control"
                  id="email"
                  v-model="profileForm.email"
                  required
                />
              </div>

              <div class="mb-3">
                <label for="name" class="form-label">Full Name</label>
                <input
                  type="text"
                  class="form-control"
                  id="name"
                  v-model="profileForm.name"
                  required
                />
              </div>

              <div class="mb-3">
                <label class="form-label">Realm ID</label>
                <input
                  type="text"
                  class="form-control"
                  :value="profile.realm_id || 'N/A'"
                  disabled
                  readonly
                />
              </div>

              <div class="mb-3" v-if="profile.last_login_at">
                <label class="form-label">Last Login</label>
                <input
                  type="text"
                  class="form-control"
                  :value="formatDate(profile.last_login_at)"
                  disabled
                  readonly
                />
              </div>

              <div class="mb-3">
                <label class="form-label">Account Created</label>
                <input
                  type="text"
                  class="form-control"
                  :value="formatDate(profile.created_time)"
                  disabled
                  readonly
                />
              </div>

              <button type="submit" class="btn btn-primary" :disabled="isSaving">
                <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="bi bi-save me-2"></i>
                {{ isSaving ? 'Saving...' : 'Save Changes' }}
              </button>
            </form>
          </div>
        </div>

        <!-- Change Password Card (only for password users) -->
        <div class="card" v-if="canChangePassword">
          <div class="card-header">
            <h5 class="mb-0"><i class="bi bi-shield-lock"></i> Change Password</h5>
          </div>
          <div class="card-body">
            <form @submit.prevent="changePassword">
              <div class="mb-3">
                <label for="oldPassword" class="form-label">Current Password</label>
                <input
                  type="password"
                  class="form-control"
                  id="oldPassword"
                  v-model="passwordForm.oldPassword"
                  required
                />
              </div>

              <div class="mb-3">
                <label for="newPassword" class="form-label">New Password</label>
                <input
                  type="password"
                  class="form-control"
                  id="newPassword"
                  v-model="passwordForm.newPassword"
                  required
                  minlength="8"
                />
                <div class="form-text">Password must be at least 8 characters long</div>
              </div>

              <div class="mb-3">
                <label for="confirmPassword" class="form-label">Confirm New Password</label>
                <input
                  type="password"
                  class="form-control"
                  id="confirmPassword"
                  v-model="passwordForm.confirmPassword"
                  required
                  :class="{ 'is-invalid': passwordForm.newPassword !== passwordForm.confirmPassword && passwordForm.confirmPassword }"
                />
                <div class="invalid-feedback" v-if="passwordForm.newPassword !== passwordForm.confirmPassword">
                  Passwords do not match
                </div>
              </div>

              <button
                type="submit"
                class="btn btn-warning"
                :disabled="isChangingPassword || passwordForm.newPassword !== passwordForm.confirmPassword"
              >
                <span v-if="isChangingPassword" class="spinner-border spinner-border-sm me-2"></span>
                <i v-else class="bi bi-key me-2"></i>
                {{ isChangingPassword ? 'Changing...' : 'Change Password' }}
              </button>
            </form>
          </div>
        </div>

        <!-- OAuth Info Card (for OAuth users) -->
        <div class="card mt-4" v-if="!canChangePassword">
          <div class="card-header">
            <h5 class="mb-0"><i class="bi bi-gitlab"></i> Authentication Method</h5>
          </div>
          <div class="card-body">
            <p class="mb-0">
              <i class="bi bi-info-circle text-info"></i>
              You are signed in with OAuth (GitLab). Password management is handled by your OAuth provider.
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import { formatDate } from '@/utils/dateUtils';

interface UserProfile {
  id: string;
  username: string;
  email: string;
  name: string;
  realm_id: string;
  avatar_url?: string;
  is_active: boolean;
  last_login_at?: string;
  created_time: string;
}

const profile = ref<UserProfile>({
  id: '',
  username: '',
  email: '',
  name: '',
  realm_id: '',
  is_active: true,
  created_time: new Date().toISOString(),
});

const profileForm = reactive({
  email: '',
  name: '',
});

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
});

const isSaving = ref(false);
const isChangingPassword = ref(false);
const successMessage = ref('');
const errorMessage = ref('');

// Check if user can change password (password users only, not OAuth users)
const canChangePassword = computed(() => {
  // We'll determine this based on the response from change-password endpoint
  // For now, assume password users can change password
  return true;
});

const loadProfile = async () => {
  try {
    const response = await apiService.get<UserProfile>('/profile');
    profile.value = response.data;
    profileForm.email = profile.value.email;
    profileForm.name = profile.value.name;
  } catch (error: any) {
    console.error('Failed to load profile:', error);
    errorMessage.value = 'Failed to load profile. Please try again.';
  }
};

const updateProfile = async () => {
  try {
    isSaving.value = true;
    errorMessage.value = '';
    successMessage.value = '';

    const response = await apiService.put<UserProfile>('/profile', {
      email: profileForm.email,
      name: profileForm.name,
    });

    profile.value = response.data;
    profileForm.email = profile.value.email;
    profileForm.name = profile.value.name;

    successMessage.value = 'Profile updated successfully!';
    showToast('Profile updated successfully', 'success');
  } catch (error: any) {
    console.error('Failed to update profile:', error);
    const message = error.response?.data?.error || 'Failed to update profile. Please try again.';
    errorMessage.value = message;
    showToast(message, 'danger');
  } finally {
    isSaving.value = false;
  }
};

const changePassword = async () => {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    errorMessage.value = 'Passwords do not match';
    return;
  }

  if (passwordForm.newPassword.length < 8) {
    errorMessage.value = 'Password must be at least 8 characters long';
    return;
  }

  try {
    isChangingPassword.value = true;
    errorMessage.value = '';
    successMessage.value = '';

    await apiService.post('/profile/change-password', {
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword,
    });

    // Clear form
    passwordForm.oldPassword = '';
    passwordForm.newPassword = '';
    passwordForm.confirmPassword = '';

    successMessage.value = 'Password changed successfully!';
    showToast('Password changed successfully', 'success');
  } catch (error: any) {
    console.error('Failed to change password:', error);
    const message = error.response?.data?.error || 'Failed to change password. Please try again.';

    // Handle OAuth users trying to change password
    if (error.response?.status === 403) {
      errorMessage.value = message;
      // Hide password change form for OAuth users
      // We could set a flag here if needed
    } else {
      errorMessage.value = message;
    }

    showToast(message, 'danger');
  } finally {
    isChangingPassword.value = false;
  }
};

onMounted(() => {
  loadProfile();
});
</script>

<style scoped>
.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
}

.card-header {
  background-color: #f8f9fa;
  border-bottom: 1px solid #dee2e6;
}

.form-control:disabled,
.form-control[readonly] {
  background-color: #e9ecef;
  cursor: not-allowed;
}

.btn {
  transition: all 0.3s ease;
}

.btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.spinner-border-sm {
  width: 1rem;
  height: 1rem;
  border-width: 0.15em;
}
</style>
