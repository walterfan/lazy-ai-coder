<template>
  <div>
    <!-- Guest Mode Banner - Show only for guests -->
    <div v-if="authStore.isGuest" class="alert alert-info alert-dismissible fade show border-info mb-3" role="alert">
      <h5 class="alert-heading">
        <i class="fas fa-info-circle me-2"></i>
        You're in Guest Mode (Read-Only)
      </h5>
      <p class="mb-3">
        You can view all features and use the app with your own API credentials, but you cannot create, update, or delete resources.
      </p>
      <p class="mb-0">
        <strong>Want full access?</strong>
        <a href="#auth-section" class="alert-link" @click.prevent="scrollToAuth">
          <i class="fas fa-arrow-down me-1"></i>
          Sign up or sign in below
        </a>
        to unlock create, update, and delete capabilities.
      </p>
      <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    </div>

    <!-- Authentication Mode Card -->
    <div class="card">
      <div class="card-header bg-info text-white">
        <h5 class="mb-0">
          <i class="fas fa-shield-alt"></i> Authentication Method
        </h5>
      </div>
      <div class="card-body">
        <p class="text-muted mb-3">
          Choose how you want to authenticate with GitLab:
        </p>

        <!-- OAuth Option -->
        <div class="card mb-3 border" :class="{ 'border-primary border-2': authStore.isOAuthMode }">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-start">
              <div class="flex-grow-1">
                <h6 class="mb-2">
                  <i class="fab fa-gitlab text-primary me-2"></i>
                  GitLab OAuth Login
                  <span v-if="authStore.isOAuthMode" class="badge bg-success ms-2">Active</span>
                  <span class="badge bg-info ms-2">Recommended</span>
                </h6>
                <p class="text-muted small mb-2">
                  Secure authentication via GitLab. Your token is automatically managed and refreshed.
                </p>
                <div v-if="authStore.isOAuthMode && authStore.user" class="mt-2">
                  <div class="d-flex align-items-center">
                    <img
                      :src="authStore.user.avatar_url"
                      width="32"
                      height="32"
                      class="rounded-circle me-2"
                      :alt="authStore.user.username"
                    />
                    <div>
                      <strong>{{ authStore.user.name }}</strong>
                      <br />
                      <small class="text-muted">@{{ authStore.user.username }}</small>
                    </div>
                  </div>
                </div>
              </div>
              <div class="ms-3">
                <a
                  v-if="!authStore.isOAuthMode"
                  href="/api/v1/auth/gitlab/login"
                  class="btn btn-primary"
                >
                  <i class="fab fa-gitlab me-2"></i>
                  Login with GitLab
                </a>
                <button
                  v-else
                  @click="handleLogout"
                  class="btn btn-outline-danger btn-sm"
                >
                  <i class="fas fa-sign-out-alt me-2"></i>
                  Logout
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Username/Password Authentication -->
        <div id="auth-section" class="card border" :class="{ 'border-success border-2': authStore.isPasswordMode }">
          <div class="card-body">
            <h6 class="mb-2">
              <i class="fas fa-user-lock text-success me-2"></i>
              Username & Password Authentication
              <span v-if="authStore.isPasswordMode" class="badge bg-success ms-2">Active</span>
            </h6>
            <p class="text-muted small mb-3">
              Create an account or sign in with username and password to get full access (create, update, delete resources).
            </p>

            <!-- Show user info if authenticated with password -->
            <div v-if="authStore.isPasswordMode && authStore.user" class="alert alert-success mb-3">
              <div class="d-flex justify-content-between align-items-center">
                <div>
                  <i class="fas fa-check-circle me-2"></i>
                  <strong>Signed in as:</strong> {{ authStore.user.username }}
                  <br />
                  <small class="text-muted">{{ authStore.user.email }}</small>
                </div>
                <button @click="handleLogout" class="btn btn-sm btn-outline-danger">
                  <i class="fas fa-sign-out-alt me-1"></i>
                  Logout
                </button>
              </div>
            </div>

            <!-- Sign Up / Sign In Tabs -->
            <div v-else>
              <ul class="nav nav-tabs mb-3" role="tablist">
                <li class="nav-item" role="presentation">
                  <button
                    class="nav-link"
                    :class="{ active: activeAuthTab === 'signup' }"
                    @click="activeAuthTab = 'signup'"
                    type="button"
                  >
                    <i class="fas fa-user-plus me-1"></i>
                    Sign Up
                  </button>
                </li>
                <li class="nav-item" role="presentation">
                  <button
                    class="nav-link"
                    :class="{ active: activeAuthTab === 'signin' }"
                    @click="activeAuthTab = 'signin'"
                    type="button"
                  >
                    <i class="fas fa-sign-in-alt me-1"></i>
                    Sign In
                  </button>
                </li>
              </ul>

              <!-- Sign Up Form -->
              <div v-show="activeAuthTab === 'signup'" class="tab-pane">
                <form @submit.prevent="handleSignUp">
                  <div class="mb-3">
                    <label for="signupUsername" class="form-label">Username *</label>
                    <input
                      type="text"
                      class="form-control"
                      id="signupUsername"
                      v-model="signupForm.username"
                      placeholder="walter_fan"
                      required
                      pattern="[a-zA-Z0-9_]{3,20}"
                      title="3-20 characters, alphanumeric and underscores only"
                    />
                    <small class="form-text text-muted">3-20 characters, alphanumeric and underscores only</small>
                  </div>

                  <div class="mb-3">
                    <label for="signupEmail" class="form-label">Email *</label>
                    <input
                      type="email"
                      class="form-control"
                      id="signupEmail"
                      v-model="signupForm.email"
                      placeholder="walter.fan@example.com"
                      required
                    />
                  </div>

                  <div class="mb-3">
                    <label for="signupPassword" class="form-label">Password *</label>
                    <input
                      type="password"
                      class="form-control"
                      id="signupPassword"
                      v-model="signupForm.password"
                      placeholder="Minimum 8 characters"
                      required
                      minlength="8"
                    />
                    <small class="form-text text-muted">Minimum 8 characters</small>
                  </div>

                  <div class="mb-3">
                    <label for="signupName" class="form-label">Full Name (Optional)</label>
                    <input
                      type="text"
                      class="form-control"
                      id="signupName"
                      v-model="signupForm.name"
                      placeholder="walter fan"
                    />
                  </div>

                  <button type="submit" class="btn btn-success" :disabled="authLoading">
                    <span v-if="authLoading">
                      <span class="spinner-border spinner-border-sm me-2"></span>
                      Signing up...
                    </span>
                    <span v-else>
                      <i class="fas fa-user-plus me-2"></i>
                      Sign Up
                    </span>
                  </button>
                </form>
              </div>

              <!-- Sign In Form -->
              <div v-show="activeAuthTab === 'signin'" class="tab-pane">
                <form @submit.prevent="handleSignIn">
                  <div class="mb-3">
                    <label for="signinUsername" class="form-label">Username *</label>
                    <input
                      type="text"
                      class="form-control"
                      id="signinUsername"
                      v-model="signinForm.username"
                      placeholder="walter_fan"
                      required
                    />
                  </div>

                  <div class="mb-3">
                    <label for="signinPassword" class="form-label">Password *</label>
                    <input
                      type="password"
                      class="form-control"
                      id="signinPassword"
                      v-model="signinForm.password"
                      placeholder="Enter your password"
                      required
                    />
                  </div>

                  <button type="submit" class="btn btn-primary" :disabled="authLoading">
                    <span v-if="authLoading">
                      <span class="spinner-border spinner-border-sm me-2"></span>
                      Signing in...
                    </span>
                    <span v-else>
                      <i class="fas fa-sign-in-alt me-2"></i>
                      Sign In
                    </span>
                  </button>
                </form>
              </div>
            </div>
          </div>
        </div>

        <!-- Guest Mode (Manual Token Entry) -->
        <div class="card border mt-3" :class="{ 'border-secondary border-2': authStore.isGuest }">
          <div class="card-body">
            <h6 class="mb-2">
              <i class="fas fa-eye text-secondary me-2"></i>
              Guest Mode (Read-Only)
              <span v-if="authStore.isGuest" class="badge bg-secondary ms-2">Active</span>
            </h6>
            <p class="text-muted small mb-3">
              Enter your API credentials to use the app as a guest. You can view all features but cannot create, update, or delete resources.
            </p>
            <div class="alert alert-warning small mb-3">
              <i class="fas fa-lock me-2"></i>
              <strong>Guest mode is read-only.</strong> To save prompts, rules, and commands, please
              <a href="#" @click.prevent="scrollToAuth" class="alert-link">sign up or sign in</a>.
            </div>

            <div v-if="authStore.isOAuthMode" class="alert alert-info small mb-0">
              <i class="fas fa-info-circle me-2"></i>
              You're using OAuth. If you want to switch to manual token mode, logout first.
            </div>

            <div v-else>
              <label for="gitlabTokenManual" class="form-label">GitLab Personal Access Token</label>
              <input
                type="password"
                class="form-control"
                id="gitlabTokenManual"
                :value="gitlabToken"
                @input="$emit('update:gitlabToken', ($event.target as HTMLInputElement).value)"
                placeholder="glpat-xxxxxxxxxxxx"
                :disabled="authStore.isOAuthMode"
              />
              <small class="form-text text-muted">
                Create token at: <strong>GitLab → Settings → Access Tokens</strong><br />
                Required scopes: <code>read_user</code>, <code>read_api</code>
              </small>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '@/stores/authStore';
import { showToast } from '@/utils/toast';
import { authService } from '@/services/authService';

// Props
defineProps<{
  gitlabToken?: string;
}>();

// Emits
const emit = defineEmits<{
  'update:gitlabToken': [value: string];
}>();

const authStore = useAuthStore();

// Authentication forms
const activeAuthTab = ref<'signup' | 'signin'>('signup');
const authLoading = ref(false);

const signupForm = ref({
  username: '',
  email: '',
  password: '',
  name: '',
});

const signinForm = ref({
  username: '',
  password: '',
});

// Scroll to authentication section
function scrollToAuth() {
  const authSection = document.getElementById('auth-section');
  if (authSection) {
    authSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
    // Optional: Add a highlight effect
    authSection.classList.add('highlight-pulse');
    setTimeout(() => {
      authSection.classList.remove('highlight-pulse');
    }, 2000);
  }
}

async function handleSignUp() {
  authLoading.value = true;
  try {
    const response = await authService.signUp({
      username: signupForm.value.username,
      email: signupForm.value.email,
      password: signupForm.value.password,
      name: signupForm.value.name || signupForm.value.username,
    });

    // Store auth data
    authStore.switchToPasswordMode(response.token, response.user);

    // Clear form
    signupForm.value = { username: '', email: '', password: '', name: '' };

    showToast(`Welcome, ${response.user.username}! Your account has been created.`, 'success');
  } catch (error: any) {
    const message = error.response?.data?.error || error.message || 'Signup failed';
    showToast(message, 'danger');
  } finally {
    authLoading.value = false;
  }
}

async function handleSignIn() {
  authLoading.value = true;
  try {
    const response = await authService.signIn({
      username: signinForm.value.username,
      password: signinForm.value.password,
    });

    // Store auth data
    authStore.switchToPasswordMode(response.token, response.user);

    // Clear form
    signinForm.value = { username: '', password: '' };

    showToast(`Welcome back, ${response.user.username}!`, 'success');
  } catch (error: any) {
    const message = error.response?.data?.error || error.message || 'Sign in failed';
    showToast(message, 'danger');
  } finally {
    authLoading.value = false;
  }
}

function handleLogout() {
  if (confirm('Are you sure you want to logout? You will return to guest mode (read-only).')) {
    authStore.logout();
    showToast('Logged out successfully. You are now in guest mode (read-only).', 'info');
  }
}
</script>

<style scoped>
@keyframes highlight-pulse {
  0%, 100% {
    background-color: transparent;
  }
  50% {
    background-color: rgba(13, 202, 240, 0.1);
  }
}

.highlight-pulse {
  animation: highlight-pulse 2s ease-in-out;
}
</style>
