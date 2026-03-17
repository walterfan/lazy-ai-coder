<template>
  <div
    v-if="show"
    class="modal fade show d-block"
    tabindex="-1"
    style="background: rgba(0,0,0,0.5)"
    @click.self="close"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header bg-light">
          <h5 class="modal-title">
            <i class="fas fa-lock text-warning me-2"></i>
            Sign Up to Continue
          </h5>
          <button
            type="button"
            class="btn-close"
            @click="close"
            aria-label="Close"
          ></button>
        </div>
        <div class="modal-body">
          <p class="lead text-center mb-4">
            This feature requires a free account.
          </p>
          <div class="row g-3">
            <div class="col-md-6">
              <div class="feature-item">
                <i class="fas fa-plus-circle text-success me-2"></i>
                Create custom prompts
              </div>
            </div>
            <div class="col-md-6">
              <div class="feature-item">
                <i class="fas fa-edit text-primary me-2"></i>
                Manage your content
              </div>
            </div>
            <div class="col-md-6">
              <div class="feature-item">
                <i class="fas fa-save text-info me-2"></i>
                Save configurations
              </div>
            </div>
            <div class="col-md-6">
              <div class="feature-item">
                <i class="fas fa-sync text-purple me-2"></i>
                Sync across devices
              </div>
            </div>
          </div>
          <div class="alert alert-info mt-4 mb-0">
            <i class="fas fa-shield-alt me-2"></i>
            <small>Your data is private and secure. We never share your information.</small>
          </div>
        </div>
        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-secondary"
            @click="close"
          >
            <i class="fas fa-times me-2"></i>
            Maybe Later
          </button>
          <button
            type="button"
            class="btn btn-primary btn-lg"
            @click="goToSignup"
          >
            <i class="fas fa-user-plus me-2"></i>
            Sign Up Now - It's Free!
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

defineProps<{
  show: boolean;
}>();

const emit = defineEmits<{
  close: [];
}>();

const router = useRouter();

function close() {
  emit('close');
}

function goToSignup() {
  emit('close');
  router.push('/settings');
  // Scroll to auth section after navigation
  setTimeout(() => {
    const authSection = document.getElementById('auth-section');
    if (authSection) {
      authSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
      // Add highlight effect
      authSection.classList.add('highlight-pulse');
      setTimeout(() => {
        authSection.classList.remove('highlight-pulse');
      }, 3000);
    }
  }, 100);
}
</script>

<style scoped>
.modal {
  display: block;
}

.modal-content {
  border: none;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.modal-header {
  border-bottom: 2px solid #dee2e6;
}

.feature-item {
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 0.5rem;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
}

.feature-item i {
  font-size: 1.2rem;
  flex-shrink: 0;
}

.text-purple {
  color: #6f42c1;
}

.btn-lg {
  padding: 0.75rem 1.5rem;
  font-size: 1.1rem;
}

@media (max-width: 576px) {
  .modal-dialog {
    margin: 1rem;
  }

  .btn-lg {
    font-size: 1rem;
    padding: 0.6rem 1.2rem;
  }
}
</style>
