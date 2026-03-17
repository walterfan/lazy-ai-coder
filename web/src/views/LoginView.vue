<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card shadow">
          <div class="card-header bg-primary text-white text-center">
            <h4 class="mb-0">
              <i class="fas fa-sign-in-alt me-2"></i>
              Sign In
            </h4>
          </div>
          <div class="card-body p-4">
            <!-- OAuth Option -->
            <div class="mb-4">
              <div class="d-grid">
                <a
                  href="/api/v1/auth/gitlab/login"
                  class="btn btn-lg btn-outline-primary"
                >
                  <i class="fab fa-gitlab me-2"></i>
                  Sign In with GitLab OAuth
                </a>
              </div>
              <small class="text-muted d-block text-center mt-2">
                Recommended for secure authentication
              </small>
            </div>

            <div class="position-relative mb-4">
              <hr />
              <span class="position-absolute top-50 start-50 translate-middle bg-white px-3 text-muted">
                or
              </span>
            </div>

            <!-- Username/Password Sign In Form -->
            <form @submit.prevent="handleSignIn">
              <div class="mb-3">
                <label for="signinUsername" class="form-label">
                  <i class="fas fa-user me-1"></i>
                  Username
                </label>
                <input
                  type="text"
                  class="form-control"
                  id="signinUsername"
                  v-model="signinForm.username"
                  placeholder="Enter your username"
                  required
                  autofocus
                />
              </div>

              <div class="mb-3">
                <label for="signinPassword" class="form-label">
                  <i class="fas fa-lock me-1"></i>
                  Password
                </label>
                <input
                  type="password"
                  class="form-control"
                  id="signinPassword"
                  v-model="signinForm.password"
                  placeholder="Enter your password"
                  required
                />
              </div>

              <div class="d-grid mb-3">
                <button type="submit" class="btn btn-primary btn-lg" :disabled="isLoading">
                  <span v-if="isLoading">
                    <span class="spinner-border spinner-border-sm me-2"></span>
                    Signing in...
                  </span>
                  <span v-else>
                    <i class="fas fa-sign-in-alt me-2"></i>
                    Sign In
                  </span>
                </button>
              </div>
            </form>

            <!-- Sign Up Link -->
            <div class="text-center mt-3">
              <p class="text-muted mb-0">
                Don't have an account?
                <router-link to="/signup" class="text-decoration-none fw-bold">
                  Sign Up
                </router-link>
              </p>
            </div>

            <!-- Guest Mode Link -->
            <div class="text-center mt-3 pt-3 border-top">
              <p class="text-muted small mb-0">
                <i class="fas fa-eye me-1"></i>
                Or continue as
                <router-link to="/" class="text-decoration-none">
                  Guest (Read-Only)
                </router-link>
              </p>
            </div>
          </div>
        </div>

        <!-- Info Alert -->
        <div class="alert alert-info mt-3 shadow-sm">
          <i class="fas fa-info-circle me-2"></i>
          <strong>Sign in to unlock full access:</strong>
          <ul class="mb-0 mt-2 small">
            <li>Create and manage prompts, rules, and commands</li>
            <li>Save your configurations</li>
            <li>Sync across devices</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/authStore';
import { showToast } from '@/utils/toast';
import { authService } from '@/services/authService';

const router = useRouter();
const authStore = useAuthStore();
const isLoading = ref(false);

const signinForm = ref({
  username: '',
  password: '',
});

// Redirect if already authenticated
onMounted(() => {
  if (authStore.isAuthenticated) {
    router.push('/');
    showToast('You are already signed in!', 'info');
  }
});

async function handleSignIn() {
  isLoading.value = true;
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

    // Redirect to home page
    router.push('/');
  } catch (error: any) {
    const errorMsg = error.response?.data?.error || error.message || 'Sign in failed';

    // Check for pending approval
    if (errorMsg.includes('pending') || errorMsg.includes('approval') || errorMsg.includes('not active')) {
      showToast('⏳ Your account is pending admin approval. Please wait for an administrator to approve your registration before signing in.', 'warning');
    } else {
      showToast(errorMsg, 'danger');
    }
  } finally {
    isLoading.value = false;
  }
}
</script>

<style scoped>
.card {
  border: none;
}

.form-control:focus {
  border-color: #0d6efd;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}

hr {
  margin: 0;
}
</style>
