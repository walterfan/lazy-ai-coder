<template>
  <div class="container">
    <div class="row justify-content-center mt-5">
      <div class="col-md-6 text-center">
        <div v-if="loading" class="mb-4">
          <div class="spinner-border text-primary" style="width: 3rem; height: 3rem;" role="status">
            <span class="visually-hidden">Loading...</span>
          </div>
          <h4 class="mt-4">Completing GitLab OAuth login...</h4>
          <p class="text-muted">Please wait while we process your authentication.</p>
        </div>

        <div v-else-if="error" class="alert alert-danger">
          <h5>
            <i class="fas fa-exclamation-triangle me-2"></i>
            Authentication Failed
          </h5>
          <p class="mb-3">{{ error }}</p>
          <router-link to="/settings" class="btn btn-primary">
            <i class="fas fa-arrow-left me-2"></i>
            Back to Settings
          </router-link>
        </div>

        <div v-else-if="success" class="alert alert-success">
          <h5>
            <i class="fas fa-check-circle me-2"></i>
            Login Successful!
          </h5>
          <p class="mb-0">Redirecting to home page...</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/authStore';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const loading = ref(true);
const error = ref('');
const success = ref(false);

onMounted(async () => {
  const code = route.query.code as string;
  const state = route.query.state as string;
  const errorParam = route.query.error as string;

  // Check for OAuth error
  if (errorParam) {
    error.value = `GitLab OAuth error: ${errorParam}`;
    loading.value = false;
    return;
  }

  // Check for authorization code
  if (!code) {
    error.value = 'Missing authorization code from GitLab';
    loading.value = false;
    return;
  }

  try {
    const result = await authStore.handleOAuthCallback(code, state);

    if (result) {
      success.value = true;
      // Redirect to home page after short delay
      setTimeout(() => {
        router.push('/');
      }, 1500);
    } else {
      error.value = authStore.error || 'OAuth login failed. Please try again.';
      loading.value = false;
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An unexpected error occurred';
    loading.value = false;
  }
});
</script>

<style scoped>
.spinner-border {
  border-width: 0.3rem;
}

.alert {
  border-radius: 0.5rem;
  padding: 2rem;
}
</style>
