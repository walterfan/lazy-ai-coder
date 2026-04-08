<template>
  <div
    v-if="show"
    class="modal fade show d-block"
    tabindex="-1"
    style="background: rgba(0,0,0,0.6); z-index: 9999"
    @click.self="handleCancel"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content border-0 shadow-lg">
        <!-- Header -->
        <div :class="['modal-header', headerClass]">
          <h5 class="modal-title d-flex align-items-center">
            <i :class="['me-2', iconClass]"></i>
            {{ title }}
          </h5>
          <button
            type="button"
            class="btn-close btn-close-white"
            @click="handleCancel"
            aria-label="Close"
          ></button>
        </div>

        <!-- Body -->
        <div class="modal-body py-4">
          <div class="text-center mb-3">
            <i :class="['display-1', iconClass, 'opacity-50']"></i>
          </div>

          <p class="lead text-center mb-3">{{ message }}</p>

          <div v-if="type === 'expired'" class="alert alert-warning">
            <i class="fas fa-clock me-2"></i>
            <small>Your session has expired for security reasons. Please login again to continue.</small>
          </div>

          <div v-else-if="type === 'unauthorized'" class="alert alert-danger">
            <i class="fas fa-shield-alt me-2"></i>
            <small>You don't have permission to perform this action. Please login with an account that has the required permissions.</small>
          </div>

          <div v-else class="alert alert-info">
            <i class="fas fa-info-circle me-2"></i>
            <small>Guest mode is read-only. Sign up for a free account to create and modify content.</small>
          </div>

          <!-- Features (only for guest mode) -->
          <div v-if="type === 'guest'" class="features-grid mt-3">
            <div class="feature-badge">
              <i class="fas fa-edit text-primary"></i>
              <span>Edit Content</span>
            </div>
            <div class="feature-badge">
              <i class="fas fa-save text-success"></i>
              <span>Save Changes</span>
            </div>
            <div class="feature-badge">
              <i class="fas fa-sync text-info"></i>
              <span>Sync Data</span>
            </div>
            <div class="feature-badge">
              <i class="fas fa-shield-alt text-warning"></i>
              <span>Secure Storage</span>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="modal-footer bg-light">
          <button
            type="button"
            class="btn btn-secondary"
            @click="handleCancel"
          >
            <i class="fas fa-times me-2"></i>
            {{ cancelText }}
          </button>
          <button
            type="button"
            :class="['btn', confirmButtonClass]"
            @click="handleConfirm"
          >
            <i :class="['me-2', confirmIconClass]"></i>
            {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';

export interface AuthWarningProps {
  show: boolean;
  type: 'expired' | 'unauthorized' | 'guest';
  message?: string;
}

const props = withDefaults(defineProps<AuthWarningProps>(), {
  message: '',
});

const emit = defineEmits<{
  close: [];
  confirm: [];
}>();

const router = useRouter();

const title = computed(() => {
  switch (props.type) {
    case 'expired':
      return 'Session Expired';
    case 'unauthorized':
      return 'Authentication Required';
    case 'guest':
      return 'Sign Up Required';
    default:
      return 'Authentication Required';
  }
});

const headerClass = computed(() => {
  switch (props.type) {
    case 'expired':
      return 'bg-warning text-dark';
    case 'unauthorized':
      return 'bg-danger text-white';
    case 'guest':
      return 'bg-primary text-white';
    default:
      return 'bg-primary text-white';
  }
});

const iconClass = computed(() => {
  switch (props.type) {
    case 'expired':
      return 'fas fa-clock';
    case 'unauthorized':
      return 'fas fa-lock';
    case 'guest':
      return 'fas fa-user-plus';
    default:
      return 'fas fa-exclamation-triangle';
  }
});

const confirmButtonClass = computed(() => {
  switch (props.type) {
    case 'expired':
      return 'btn-warning';
    case 'unauthorized':
      return 'btn-danger';
    case 'guest':
      return 'btn-primary';
    default:
      return 'btn-primary';
  }
});

const confirmIconClass = computed(() => {
  switch (props.type) {
    case 'expired':
    case 'unauthorized':
      return 'fas fa-sign-in-alt';
    case 'guest':
      return 'fas fa-user-plus';
    default:
      return 'fas fa-sign-in-alt';
  }
});

const confirmText = computed(() => {
  switch (props.type) {
    case 'expired':
      return 'Login Again';
    case 'unauthorized':
      return 'Go to Login';
    case 'guest':
      return 'Sign Up Now';
    default:
      return 'Login';
  }
});

const cancelText = computed(() => {
  if (props.type === 'guest') {
    return 'Continue as Guest';
  }
  return 'Cancel';
});

function handleCancel() {
  emit('close');
}

function handleConfirm() {
  emit('confirm');
  emit('close');

  router.push('/login');
}
</script>

<style scoped>
.modal {
  display: block;
  animation: fadeIn 0.2s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.modal-content {
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    transform: translateY(-50px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  border-bottom: none;
  padding: 1.25rem 1.5rem;
}

.modal-body {
  padding: 1.5rem;
}

.modal-footer {
  border-top: 1px solid #dee2e6;
  padding: 1rem 1.5rem;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.feature-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 0.5rem;
  font-size: 0.9rem;
  transition: all 0.2s;
}

.feature-badge:hover {
  background: #e9ecef;
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.feature-badge i {
  font-size: 1.2rem;
  flex-shrink: 0;
}

.btn {
  min-width: 120px;
  transition: all 0.2s;
}

.btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

@media (max-width: 576px) {
  .modal-dialog {
    margin: 0.5rem;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .lead {
    font-size: 1rem;
  }
}

/* Highlight pulse animation for auth section */
:global(.highlight-pulse) {
  animation: pulse 1s ease-in-out 3;
  border: 2px solid #ffc107;
  border-radius: 0.5rem;
  box-shadow: 0 0 20px rgba(255, 193, 7, 0.5);
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 20px rgba(255, 193, 7, 0.5);
  }
  50% {
    box-shadow: 0 0 40px rgba(255, 193, 7, 0.8);
  }
}
</style>
