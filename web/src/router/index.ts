import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/HomeView.vue'),
  },
  {
    path: '/auth/callback',
    name: 'OAuthCallback',
    component: () => import('@/views/OAuthCallbackView.vue'),
    meta: { public: true },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/signup',
    name: 'Signup',
    component: () => import('@/views/SignupView.vue'),
    meta: { public: true },
  },
  {
    path: '/assistant',
    name: 'Assistant',
    component: () => import('@/views/AssistantView.vue'),
  },
  {
    path: '/prompts',
    name: 'Prompts',
    component: () => import('@/views/PromptsView.vue'),
  },
  {
    path: '/projects',
    name: 'Projects',
    component: () => import('@/views/ProjectsView.vue'),
  },
  {
    path: '/documents',
    name: 'Documents',
    component: () => import('@/views/DocumentsView.vue'),
  },
  {
    path: '/cursor-rules',
    name: 'CursorRules',
    component: () => import('@/views/CursorRulesView.vue'),
  },
  {
    path: '/cursor-commands',
    name: 'CursorCommands',
    component: () => import('@/views/CursorCommandsView.vue'),
  },
  {
    path: '/asset-library',
    name: 'Assets',
    component: () => import('@/views/AssetsView.vue'),
  },
  {
    path: '/chat',
    name: 'Chat',
    component: () => import('@/views/ChatRecordView.vue'),
  },
  {
    path: '/chat/history',
    name: 'ChatHistory',
    component: () => import('@/views/ChatHistoryView.vue'),
  },
  {
    path: '/codekg',
    name: 'CodeKG',
    component: () => import('@/views/CodeKGView.vue'),
  },
  {
    path: '/tools',
    name: 'Tools',
    component: () => import('@/views/ToolsView.vue'),
    children: [
      {
        path: 'smart-prompt',
        name: 'SmartPrompt',
        component: () => import('@/views/SmartPromptGeneratorView.vue'),
      },
      {
        path: 'encoding',
        name: 'EncodingTools',
        component: () => import('@/views/EncodingToolsView.vue'),
      },
      {
        path: 'update-readme',
        name: 'UpdateReadme',
        component: () => import('@/views/tools/UpdateReadmeView.vue'),
      },
      {
        path: 'write-user-story',
        name: 'WriteUserStory',
        component: () => import('@/views/tools/WriteUserStoryView.vue'),
      },
      {
        path: 'report-bug',
        name: 'ReportBug',
        component: () => import('@/views/tools/ReportBugView.vue'),
      },
      {
        path: 'update-comments',
        name: 'UpdateComments',
        component: () => import('@/views/tools/UpdateCommentsView.vue'),
      },
      {
        path: 'write-rca',
        name: 'WriteRCA',
        component: () => import('@/views/tools/WriteRCAView.vue'),
      },
      {
        path: 'mr-summary',
        name: 'MRSummary',
        component: () => import('@/views/MRSummaryView.vue'),
      },
      {
        path: 'llm-models',
        name: 'LLMModels',
        component: () => import('@/views/LLMModelsView.vue'),
      },
    ],
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue'),
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/help',
    name: 'Help',
    component: () => import('@/views/HelpView.vue'),
  },
  // Admin routes (super admin only)
  {
    path: '/admin/pending-users',
    name: 'admin-pending-users',
    component: () => import('@/views/admin/PendingUsersView.vue'),
    meta: { requiresAuth: true, requiresSuperAdmin: true },
  },
  {
    path: '/admin/users',
    name: 'admin-users',
    component: () => import('@/views/admin/UsersView.vue'),
    meta: { requiresAuth: true, requiresSuperAdmin: true },
  },
  {
    path: '/admin/realms',
    name: 'admin-realms',
    component: () => import('@/views/admin/RealmsView.vue'),
    meta: { requiresAuth: true, requiresSuperAdmin: true },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

// Navigation guard for admin routes
import { useAuthStore } from '@/stores/authStore'

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  // Check authentication requirement
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      console.warn('Authentication required for:', to.path)
      // Redirect to login if not authenticated
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }
  }

  // Check super admin requirement
  if (to.meta.requiresSuperAdmin) {
    if (!authStore.isSuperAdmin) {
      console.warn('Super admin access required for:', to.path)
      // Redirect to home if not super admin
      next({ name: 'Home' })
      return
    }
  }

  next()
})

export default router;
