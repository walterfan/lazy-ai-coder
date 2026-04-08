<template>
  <div id="app" class="d-flex flex-column min-vh-100">
    <NavigationBar />

    <!-- Guest Mode Banner -->
    <div
      v-if="authStore.isGuest && !bannerDismissed"
      class="alert alert-warning alert-dismissible fade show mb-0 guest-banner"
      role="alert"
    >
      <div class="container">
        <div class="d-flex align-items-center justify-content-between flex-wrap">
          <div class="banner-content">
            <i class="fas fa-info-circle me-2"></i>
            <strong>Guest Mode:</strong> You're viewing in read-only mode.
            <a @click="goToSignup" class="alert-link ms-2 cursor-pointer">
              Sign up for free
            </a>
            to create and manage prompts, rules, and commands.
          </div>
          <button
            type="button"
            class="btn-close"
            @click="dismissBanner"
            aria-label="Close"
          ></button>
        </div>
      </div>
    </div>

    <main class="flex-grow-1">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
    <FooterBar />

    <!-- Global Auth Warning Dialog -->
    <AuthWarningDialog
      :show="authStore.authWarning.show"
      :type="authStore.authWarning.type"
      :message="authStore.authWarning.message"
      @close="authStore.hideAuthWarning"
      @confirm="handleAuthConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import NavigationBar from '@/components/NavigationBar.vue';
import FooterBar from '@/components/FooterBar.vue';
import AuthWarningDialog from '@/components/AuthWarningDialog.vue';
import { useSettingsStore } from '@/stores/settingsStore';
import { useAuthStore } from '@/stores/authStore';
import { storage } from '@/utils/storage';

const settingsStore = useSettingsStore();
const authStore = useAuthStore();
const router = useRouter();

// Load stores synchronously BEFORE component mounts to prevent blank page flash
// This ensures all child components have access to settings/auth immediately
settingsStore.loadFromStorage();
authStore.loadFromStorage();

// Load banner dismissed state from localStorage (must not throw)
const bannerDismissed = ref(storage.getItem('guestBannerDismissed') === 'true');

onMounted(() => {
  // Refresh in case storage changed externally (rare edge case)
  // No longer needed for initial load since we load synchronously above
});

function dismissBanner() {
  bannerDismissed.value = true;
  storage.setItem('guestBannerDismissed', 'true');
}

function handleAuthConfirm() {
  authStore.hideAuthWarning();
  router.push('/login');
}

function goToSignup() {
  router.push('/login');
}
</script>

<style>
/* Import Font Awesome from CDN */
@import 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css';

#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Fade transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 10px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
}

::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 5px;
}

::-webkit-scrollbar-thumb:hover {
  background: #555;
}

/* Loading spinner */
.spinner-border-sm {
  width: 1rem;
  height: 1rem;
  border-width: 0.2em;
}

/* Code blocks */
pre {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
  padding: 1rem;
  overflow-x: auto;
}

code {
  color: #e83e8c;
  background-color: #f8f9fa;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  font-size: 87.5%;
}

pre code {
  color: inherit;
  background-color: transparent;
  padding: 0;
}

/* Card styles */
.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
  transition: box-shadow 0.3s ease;
}

.card:hover {
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
}

/* Button styles */
.btn {
  transition: all 0.3s ease;
}

/* Toast container position */
.toast-container {
  z-index: 1055;
}

/* Guest banner styles */
.guest-banner {
  border-left: 4px solid #ffc107;
  border-radius: 0;
  margin-bottom: 0 !important;
}

.guest-banner .banner-content {
  flex: 1;
  padding: 0.25rem 0;
}

.cursor-pointer {
  cursor: pointer;
  text-decoration: underline;
}

.cursor-pointer:hover {
  text-decoration: none;
}

@media (max-width: 768px) {
  .guest-banner .d-flex {
    flex-direction: column;
    align-items: flex-start !important;
  }

  .guest-banner .btn-close {
    position: absolute;
    right: 1rem;
    top: 1rem;
  }

  .guest-banner .banner-content {
    padding-right: 2rem;
  }
}
</style>
