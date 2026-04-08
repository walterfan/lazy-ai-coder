<template>
  <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
    <div class="container">
      <router-link to="/" class="navbar-brand d-flex align-items-center">
        <i class="fas fa-robot me-2"></i>
        <span class="brand-text">{{ appStore.title }}</span>
        <span class="version-badge ms-2">v{{ appStore.version }}</span>
      </router-link>
      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarNav"
        aria-controls="navbarNav"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav ms-auto">
          <li class="nav-item">
            <router-link to="/" class="nav-link" :class="{ active: isActiveRoute('Home') }">
              <i class="fas fa-home me-1"></i>
              <span class="nav-text">Home</span>
            </router-link>
          </li>
          <li class="nav-item dropdown">
            <a
              class="nav-link dropdown-toggle"
              :class="{ active: isActiveRoute('Assistant') || isActiveRoute('Chat') || isActiveRoute('ChatHistory') }"
              href="#"
              id="assistantDropdown"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <i class="fas fa-magic me-1"></i>
              <span class="nav-text">Assistant</span>
            </a>
            <ul class="dropdown-menu" aria-labelledby="assistantDropdown">
              <li>
                <router-link to="/chat" class="dropdown-item">
                  <i class="fas fa-graduation-cap me-2"></i>Coding Mate
                </router-link>
              </li>
              <li>
                <router-link
                  to="/assistant?module=code"
                  class="dropdown-item"
                >
                  <i class="fas fa-code me-2"></i>General
                </router-link>
              </li>
              <li>
                <router-link
                  to="/assistant?module=review"
                  class="dropdown-item"
                >
                  <i class="fas fa-search me-2"></i>Review
                </router-link>
              </li>
              <li>
                <router-link
                  to="/assistant?module=write"
                  class="dropdown-item"
                >
                  <i class="fas fa-pen me-2"></i>Write
                </router-link>
              </li>
            </ul>
          </li>
          <li class="nav-item dropdown">
            <a
              class="nav-link dropdown-toggle"
              :class="{ active: isActiveRoute('Tools') || isActiveRoute('CodeKG') || isActiveRoute('Assets') || isActiveRoute('CursorRules') || isActiveRoute('CursorCommands') }"
              href="#"
              id="toolsDropdown"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <i class="fas fa-tools me-1"></i>
              <span class="nav-text">Tools</span>
            </a>
            <ul class="dropdown-menu" aria-labelledby="toolsDropdown">
              <li>
                <router-link to="/codekg" class="dropdown-item">
                  <i class="fas fa-project-diagram me-2"></i>Code Knowledge Base
                </router-link>
              </li>
              <li>
                <router-link to="/asset-library" class="dropdown-item">
                  <i class="fas fa-folder-open me-2"></i>Assets (Commands / Rules / Skills)
                </router-link>
              </li>
              <li>
                <router-link to="/cursor-rules" class="dropdown-item">
                  <i class="fas fa-code me-2"></i>Cursor Rules
                </router-link>
              </li>
              <li>
                <router-link to="/cursor-commands" class="dropdown-item">
                  <i class="fas fa-terminal me-2"></i>Cursor Commands
                </router-link>
              </li>
              <li><hr class="dropdown-divider" /></li>
              <li>
                <router-link to="/tools/smart-prompt" class="dropdown-item">
                  <i class="fas fa-magic me-2"></i>Prompt Generator
                </router-link>
              </li>
              <li>
                <router-link to="/tools/encoding" class="dropdown-item">
                  <i class="fas fa-code me-2"></i>Encoding Tools
                </router-link>
              </li>
              <li>
                <router-link to="/tools/mr-summary" class="dropdown-item">
                  <i class="fas fa-code-branch me-2"></i>MR Summary
                </router-link>
              </li>
            </ul>
          </li>
          <li class="nav-item dropdown">
            <a
              class="nav-link dropdown-toggle"
              :class="{ active: isActiveRoute('Prompts') || isActiveRoute('Projects') || isActiveRoute('Settings') }"
              href="#"
              id="adminDropdown"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <i class="fas fa-user-shield me-1"></i>
              <span class="nav-text">Admin</span>
              <span v-if="authStore.isSuperAdmin && pendingCount > 0" class="badge bg-warning text-dark ms-1">
                {{ pendingCount }}
              </span>
            </a>
            <ul class="dropdown-menu" aria-labelledby="adminDropdown">
              <li>
                <router-link to="/prompts" class="dropdown-item" :class="{ active: isActiveRoute('Prompts') }">
                  <i class="fas fa-comments me-2"></i>Prompts
                </router-link>
              </li>
              <li>
                <router-link to="/projects" class="dropdown-item" :class="{ active: isActiveRoute('Projects') }">
                  <i class="fas fa-project-diagram me-2"></i>Projects
                </router-link>
              </li>
              <li>
                <router-link to="/settings" class="dropdown-item" :class="{ active: isActiveRoute('Settings') }">
                  <i class="fas fa-cog me-2"></i>Settings
                </router-link>
              </li>
              <template v-if="authStore.isSuperAdmin">
                <li><hr class="dropdown-divider" /></li>
                <li>
                  <router-link to="/admin/pending-users" class="dropdown-item">
                    <i class="fas fa-user-clock me-2"></i>Pending Users
                    <span v-if="pendingCount > 0" class="badge bg-warning text-dark ms-2">
                      {{ pendingCount }}
                    </span>
                  </router-link>
                </li>
                <li>
                  <router-link to="/admin/users" class="dropdown-item">
                    <i class="fas fa-users me-2"></i>Manage Users
                  </router-link>
                </li>
                <li>
                  <router-link to="/admin/realms" class="dropdown-item">
                    <i class="fas fa-building me-2"></i>Manage Realms
                  </router-link>
                </li>
                <li><hr class="dropdown-divider" /></li>
                <li>
                  <router-link to="/tools/llm-models" class="dropdown-item">
                    <i class="fas fa-robot me-2"></i>LLM Models
                  </router-link>
                </li>
              </template>
            </ul>
          </li>
          <!-- Notes menu temporarily hidden for testing
          <li class="nav-item">
            <router-link to="/documents" class="nav-link" :class="{ active: isActiveRoute('Documents') }">
              <i class="fas fa-file-alt me-1"></i>
              <span class="nav-text">Notes</span>
            </router-link>
          </li>
          -->

          <li class="nav-item">
            <router-link to="/help" class="nav-link" :class="{ active: isActiveRoute('Help') }">
              <i class="fas fa-question-circle me-1"></i>
              <span class="nav-text">Help</span>
            </router-link>
          </li>

          <!-- OAuth User Profile Dropdown -->
          <li class="nav-item dropdown" v-if="authStore.isOAuthMode && authStore.user">
            <a
              class="nav-link dropdown-toggle"
              href="#"
              id="userDropdown"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <img
                :src="authStore.user.avatar_url"
                class="rounded-circle me-2"
                width="28"
                height="28"
                :alt="authStore.user.username"
              />
              <span class="nav-text">{{ authStore.user.name }}</span>
            </a>
            <ul class="dropdown-menu dropdown-menu-end" aria-labelledby="userDropdown">
              <li>
                <span class="dropdown-item-text">
                  <small class="text-muted">@{{ authStore.user.username }}</small>
                </span>
              </li>
              <li>
                <span class="dropdown-item-text">
                  <span class="badge bg-success">OAuth Mode</span>
                </span>
              </li>
              <li><hr class="dropdown-divider" /></li>
              <li>
                <router-link to="/profile" class="dropdown-item">
                  <i class="fas fa-user me-2"></i>
                  My Profile
                </router-link>
              </li>
              <li>
                <a class="dropdown-item" href="#" @click.prevent="handleLogout">
                  <i class="fas fa-sign-out-alt me-2"></i>
                  Logout
                </a>
              </li>
            </ul>
          </li>

          <!-- Password User Profile Dropdown -->
          <li class="nav-item dropdown" v-else-if="authStore.isPasswordMode && authStore.user">
            <a
              class="nav-link dropdown-toggle"
              href="#"
              id="passwordUserDropdown"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <i class="fas fa-user-circle me-2" style="font-size: 1.5rem;"></i>
              <span class="nav-text">{{ authStore.user.username }}</span>
            </a>
            <ul class="dropdown-menu dropdown-menu-end" aria-labelledby="passwordUserDropdown">
              <li>
                <span class="dropdown-item-text">
                  <small class="text-muted">{{ authStore.user.email }}</small>
                </span>
              </li>
              <li>
                <span class="dropdown-item-text">
                  <span class="badge bg-primary">Password Mode</span>
                </span>
              </li>
              <li><hr class="dropdown-divider" /></li>
              <li>
                <router-link to="/profile" class="dropdown-item">
                  <i class="fas fa-user me-2"></i>
                  My Profile
                </router-link>
              </li>
              <li>
                <a class="dropdown-item" href="#" @click.prevent="handleLogout">
                  <i class="fas fa-sign-out-alt me-2"></i>
                  Logout
                </a>
              </li>
            </ul>
          </li>

          <!-- Account Dropdown for Guest Users -->
          <li class="nav-item dropdown" v-else-if="authStore.isGuest">
            <a
              class="nav-link dropdown-toggle"
              href="#"
              id="accountDropdown"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <i class="fas fa-user-circle me-1"></i>
              <span class="nav-text">Account</span>
            </a>
            <ul class="dropdown-menu dropdown-menu-end" aria-labelledby="accountDropdown">
              <li>
                <router-link to="/login" class="dropdown-item">
                  <i class="fas fa-sign-in-alt me-2"></i>
                  Sign In
                </router-link>
              </li>
              <li>
                <router-link to="/signup" class="dropdown-item">
                  <i class="fas fa-user-plus me-2"></i>
                  Sign Up
                </router-link>
              </li>
              <li><hr class="dropdown-divider" /></li>
              <li>
                <span class="dropdown-item-text">
                  <small class="text-muted">
                    <i class="fas fa-info-circle me-1"></i>
                    Get full access with an account
                  </small>
                </span>
              </li>
            </ul>
          </li>

          <!-- Guest Mode Indicator -->
          <li class="nav-item" v-if="authStore.isGuest">
            <a href="/login" class="nav-link" @click.prevent="$router.push('/login')">
              <span class="badge bg-warning text-dark px-3 py-2 guest-badge" title="Click to sign in for full access">
                <i class="fas fa-eye me-2"></i>
                Guest (Read-Only)
                <i class="fas fa-arrow-right ms-2"></i>
              </span>
            </a>
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, nextTick } from 'vue';
import { useAuthStore } from '@/stores/authStore';
import { useAppStore } from '@/stores/appStore';
import { useUserManagementStore } from '@/stores/userManagementStore';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const appStore = useAppStore();
const userManagementStore = useUserManagementStore();

// Pending users count for super admin badge
const pendingCount = computed(() => userManagementStore.pendingUsersCount);

const isActiveRoute = (routeName: string): boolean => {
  return computed(() => route.name === routeName).value;
};

function handleLogout() {
  authStore.logout();
  router.push('/settings');
}

// Load auth state on mount
onMounted(() => {
  authStore.loadFromStorage();
});

// Initialize Bootstrap dropdowns when component mounts
onMounted(async () => {
  await nextTick();

  // Wait for Bootstrap to be available
  const initDropdowns = () => {
    // @ts-ignore - Bootstrap is loaded globally
    if (typeof window.bootstrap !== 'undefined') {
      const dropdownElements = document.querySelectorAll('[data-bs-toggle="dropdown"]');
      dropdownElements.forEach((element) => {
        // @ts-ignore
        new window.bootstrap.Dropdown(element);
      });
    } else {
      // Retry after a short delay if Bootstrap isn't loaded yet
      setTimeout(initDropdowns, 100);
    }
  };

  initDropdowns();
});
</script>

<style scoped>
.navbar-brand .version-badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 0.25rem;
}

.nav-link {
  transition: all 0.3s ease;
}

.nav-link:hover {
  color: #fff !important;
}

.nav-link.active {
  color: #fff !important;
  font-weight: 500;
}

/* Guest badge styling */
.guest-badge {
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.3s ease;
  animation: subtle-pulse 3s ease-in-out infinite;
}

.guest-badge:hover {
  background-color: #ffc107 !important;
  transform: scale(1.05);
  box-shadow: 0 0 10px rgba(255, 193, 7, 0.5);
}

@keyframes subtle-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.85;
  }
}

@media (max-width: 768px) {
  .nav-text {
    display: inline;
  }
}
</style>
