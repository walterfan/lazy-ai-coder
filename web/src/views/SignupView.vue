<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card shadow">
          <div class="card-header bg-success text-white text-center">
            <h4 class="mb-0">
              <i class="fas fa-user-plus me-2"></i>
              Sign Up
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
                  Sign Up with GitLab OAuth
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

            <!-- Username/Password Sign Up Form -->
            <form @submit.prevent="handleSignUp">
              <div class="mb-3">
                <label for="signupUsername" class="form-label">
                  <i class="fas fa-user me-1"></i>
                  Username *
                </label>
                <input
                  type="text"
                  class="form-control"
                  id="signupUsername"
                  v-model="signupForm.username"
                  placeholder="walter_fan"
                  required
                  pattern="[a-zA-Z0-9_]{3,20}"
                  title="3-20 characters, alphanumeric and underscores only"
                  autofocus
                />
                <small class="form-text text-muted">3-20 characters, alphanumeric and underscores only</small>
              </div>

              <div class="mb-3">
                <label for="signupEmail" class="form-label">
                  <i class="fas fa-envelope me-1"></i>
                  Email *
                </label>
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
                <label for="signupPassword" class="form-label">
                  <i class="fas fa-lock me-1"></i>
                  Password *
                </label>
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
                <label for="signupName" class="form-label">
                  <i class="fas fa-id-card me-1"></i>
                  Full Name (Optional)
                </label>
                <input
                  type="text"
                  class="form-control"
                  id="signupName"
                  v-model="signupForm.name"
                  placeholder="walter fan"
                />
              </div>

              <div class="d-grid mb-3">
                <button type="submit" class="btn btn-success btn-lg" :disabled="isLoading">
                  <span v-if="isLoading">
                    <span class="spinner-border spinner-border-sm me-2"></span>
                    Creating account...
                  </span>
                  <span v-else>
                    <i class="fas fa-user-plus me-2"></i>
                    Create Account
                  </span>
                </button>
              </div>
            </form>

            <!-- Sign In Link -->
            <div class="text-center mt-3">
              <p class="text-muted mb-0">
                Already have an account?
                <router-link to="/login" class="text-decoration-none fw-bold">
                  Sign In
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
        <div class="alert alert-success mt-3 shadow-sm">
          <i class="fas fa-shield-alt me-2"></i>
          <strong>Create an account to unlock:</strong>
          <ul class="mb-0 mt-2 small">
            <li>Create and manage prompts, rules, and commands</li>
            <li>Save your configurations across devices</li>
            <li>Full access to all features</li>
          </ul>
        </div>

        <!-- Privacy Notice -->
        <div class="alert alert-light border mt-3 shadow-sm">
          <small class="text-muted">
            <i class="fas fa-lock me-1"></i>
            <strong>Your privacy matters:</strong> We never share your information.
            Your data is encrypted and secure.
          </small>
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

const signupForm = ref({
  username: '',
  email: '',
  password: '',
  name: '',
});

// Redirect if already authenticated
onMounted(() => {
  if (authStore.isAuthenticated) {
    router.push('/');
    showToast('You are already signed in!', 'info');
  }
});

async function handleSignUp() {
  isLoading.value = true;
  try {
    await authService.signUp({
      username: signupForm.value.username,
      email: signupForm.value.email,
      password: signupForm.value.password,
      name: signupForm.value.name || signupForm.value.username,
    });

    // Clear form
    signupForm.value = { username: '', email: '', password: '', name: '' };

    // Show approval wait message
    showToast('✅ Account created! Your registration is pending admin approval. You will be able to sign in once an administrator approves your account.', 'info');

    // Redirect to login page with registered query parameter
    router.push({ name: 'Login', query: { registered: 'true' } });
  } catch (error: any) {
    const message = error.response?.data?.error || error.message || 'Signup failed';
    showToast(message, 'danger');
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
  border-color: #198754;
  box-shadow: 0 0 0 0.25rem rgba(25, 135, 84, 0.25);
}

hr {
  margin: 0;
}
</style>
