<template>
  <div class="chat-page d-flex">
    <!-- Left sidebar: SDLC task menu -->
    <aside class="task-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header d-flex align-items-center justify-content-between">
        <span v-if="!sidebarCollapsed" class="sidebar-title">Coding Tasks</span>
        <button class="btn btn-sm btn-link sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'">
          <i :class="sidebarCollapsed ? 'fas fa-angle-right' : 'fas fa-angle-left'"></i>
        </button>
      </div>
      <nav v-if="!sidebarCollapsed" class="sidebar-nav">
        <!-- Active skill indicator -->
        <div v-if="selectedSkillPath" class="active-skill-banner">
          <div class="d-flex align-items-center gap-1">
            <i class="fas fa-graduation-cap"></i>
            <span class="text-truncate">{{ selectedSkillName }}</span>
          </div>
          <button class="btn btn-sm btn-link text-danger p-0" @click="clearSkill" title="Deselect skill">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <!-- Skill picker section -->
        <div class="phase-group">
          <button class="phase-header" @click="skillsExpanded = !skillsExpanded">
            <i class="fas fa-graduation-cap phase-icon"></i>
            <span class="phase-name">Skills</span>
            <span v-if="selectedSkillPath" class="badge bg-warning text-dark ms-auto me-1" style="font-size:0.65rem">1</span>
            <i :class="[skillsExpanded ? 'fas fa-chevron-up' : 'fas fa-chevron-down', 'phase-chevron', { 'ms-auto': !selectedSkillPath }]"></i>
          </button>
          <transition name="slide">
            <div v-if="skillsExpanded" class="skill-picker">
              <div class="d-flex gap-1 mb-1">
                <input
                  v-model="skillSearchQuery"
                  type="text"
                  class="form-control form-control-sm skill-search flex-grow-1"
                  placeholder="Search skills..."
                />
                <button
                  class="btn btn-sm skill-filter-btn"
                  :class="{ active: showFavoritesOnly }"
                  @click="showFavoritesOnly = !showFavoritesOnly"
                  title="Show rated/favorites only"
                >
                  <i class="fas fa-star"></i>
                </button>
              </div>
              <div v-if="loadingSkillContent" class="text-center py-2">
                <div class="spinner-border spinner-border-sm text-primary"></div>
              </div>
              <div class="skill-categories">
                <div v-for="[cat, skills] of skillCategories" :key="cat" class="skill-cat-group">
                  <div class="skill-cat-label">{{ cat }} <span class="text-muted">({{ skills.length }})</span></div>
                  <div
                    v-for="skill in skills"
                    :key="skill.path"
                    class="skill-row"
                  >
                    <button
                      class="skill-option"
                      :class="{ active: selectedSkillPath === skill.path }"
                      @click="selectSkill(skill)"
                      :title="skill.snippet"
                    >
                      <span class="skill-name">{{ skill.name }}</span>
                    </button>
                    <div class="skill-stars" @click.stop>
                      <i
                        v-for="star in 5"
                        :key="star"
                        class="star-icon"
                        :class="star <= getSkillScore(skill.path) ? 'fas fa-star rated' : 'far fa-star'"
                        @click="rateSkillStar(skill, star)"
                      ></i>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </transition>
        </div>

        <!-- SDLC task phases -->
        <div
          v-for="phase in sdlcPhases"
          :key="phase.id"
          class="phase-group"
        >
          <button class="phase-header" @click="togglePhase(phase.id)">
            <i :class="phase.icon" class="phase-icon"></i>
            <span class="phase-name">{{ phase.name }}</span>
            <i :class="expandedPhases.has(phase.id) ? 'fas fa-chevron-up' : 'fas fa-chevron-down'" class="ms-auto phase-chevron"></i>
          </button>
          <transition name="slide">
            <ul v-if="expandedPhases.has(phase.id)" class="task-list">
              <li v-for="task in phase.tasks" :key="task.label">
                <button class="task-item" @click="sendTaskPrompt(phase.name, task)" :title="task.prompt">
                  {{ task.label }}
                </button>
              </li>
            </ul>
          </transition>
        </div>
      </nav>
    </aside>

    <!-- Main chat column -->
    <div class="chat-main d-flex flex-column flex-grow-1">
      <!-- Top bar -->
      <div class="chat-topbar d-flex align-items-center justify-content-between px-3 py-2 border-bottom bg-white">
        <div class="d-flex align-items-center gap-2">
          <i class="fas fa-robot text-primary fs-5"></i>
          <span class="fw-semibold">Coding Mate</span>
          <span v-if="store.turns.length > 0" class="badge bg-secondary">{{ store.turns.length }} turns</span>
          <span v-if="store.sessionId && store.turns.length > 0" class="badge bg-success" title="Conversation memory is active — the AI remembers prior turns in this session">
            <i class="fas fa-brain me-1"></i>Memory on
          </span>
          <span v-if="selectedSkillName" class="badge bg-warning text-dark" :title="`Skill: ${selectedSkillName}`">
            <i class="fas fa-graduation-cap me-1"></i>{{ selectedSkillName }}
          </span>
        </div>
        <div class="d-flex gap-2">
          <button
            v-if="store.turns.length > 0"
            class="btn btn-sm btn-outline-success"
            @click="handleFinishAndSave"
            :disabled="store.confirmLoading"
            title="Save all to history"
          >
            <i class="fas fa-save me-1"></i>Save all
          </button>
          <button
            v-if="store.turns.length > 0"
            class="btn btn-sm btn-outline-danger"
            @click="store.clearSession"
            title="New conversation"
          >
            <i class="fas fa-plus me-1"></i>New chat
          </button>
          <button
            class="btn btn-sm btn-outline-secondary"
            @click="historyPaneOpen = !historyPaneOpen"
            :title="historyPaneOpen ? 'Hide history' : 'Show history'"
          >
            <i class="fas fa-history"></i>
          </button>
        </div>
      </div>

      <!-- Message area -->
      <div class="chat-messages flex-grow-1" ref="messagesContainer">
      <!-- Welcome screen when empty -->
      <div v-if="store.turns.length === 0 && !store.submitLoading" class="welcome-screen">
        <div class="welcome-content">
          <div class="welcome-icon mb-3">
            <i class="fas fa-robot"></i>
          </div>
          <h3 class="fw-bold mb-2">Coding Mate</h3>
          <p class="text-muted mb-4">
            Your AI pair programmer. Ask about architecture, get code reviews,<br>
            explore technologies, or design solutions.
          </p>
          <div class="suggestion-chips d-flex flex-wrap justify-content-center gap-2">
            <button
              v-for="chip in suggestionChips"
              :key="chip.text"
              class="chip-btn"
              @click="submitChip(chip.text)"
            >
              <i :class="chip.icon" class="me-2"></i>{{ chip.text }}
            </button>
          </div>
        </div>
      </div>

      <!-- Conversation messages -->
      <div v-else class="messages-list">
        <template v-for="turn in store.turns" :key="turn.userInput">
          <!-- User message -->
          <div class="message-row user-row">
            <div class="message-bubble user-bubble">
              <div class="message-text">{{ turn.userInput }}</div>
            </div>
            <div class="avatar user-avatar">
              <i class="fas fa-user"></i>
            </div>
          </div>

          <!-- Assistant message -->
          <div class="message-row assistant-row">
            <div class="avatar assistant-avatar">
              <i class="fas fa-robot"></i>
            </div>
            <div class="message-bubble assistant-bubble">
              <!-- Streaming response (live markdown) -->
              <div v-if="turn.streamContent !== undefined" class="response-content">
                <div v-if="turn.streaming" class="streaming-indicator mb-1">
                  <span class="badge bg-info"><i class="fas fa-circle-notch fa-spin me-1"></i>Streaming...</span>
                </div>
                <div class="markdown-body" v-html="renderMarkdown(turn.streamContent || '')"></div>
                <span v-if="turn.streaming" class="streaming-cursor">&#x2588;</span>
              </div>
              <!-- Non-streaming: loading -->
              <div v-else-if="turn.response === null" class="typing-indicator">
                <span></span><span></span><span></span>
              </div>
              <!-- Non-streaming: structured response -->
              <div v-else class="response-content">
                <div v-if="turn.response.input_type" class="response-type-badge mb-2">
                  <span :class="`badge bg-${InputTypeConfig[turn.response.input_type]?.color || 'secondary'}`">
                    <i :class="`fas ${InputTypeConfig[turn.response.input_type]?.icon || 'fa-comment'} me-1`"></i>
                    {{ InputTypeConfig[turn.response.input_type]?.label || turn.response.input_type }}
                  </span>
                </div>
                <div class="markdown-body" v-html="renderResponse(turn)"></div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Input bar -->
    <div class="chat-input-bar border-top bg-white">
      <div class="input-container">
        <div class="input-wrapper">
          <textarea
            ref="inputEl"
            v-model="inputText"
            class="chat-input"
            placeholder="Ask about architecture, coding patterns, tech design..."
            rows="1"
            @keydown="handleKeydown"
            @input="autoResize"
            :disabled="store.submitLoading"
          ></textarea>
          <button
            class="send-btn"
            :class="{ active: inputText.trim() && !store.submitLoading }"
            @click="handleSubmit"
            :disabled="!inputText.trim() || store.submitLoading"
          >
            <i v-if="store.submitLoading" class="fas fa-circle-notch fa-spin"></i>
            <i v-else class="fas fa-arrow-up"></i>
          </button>
        </div>
        <div class="input-hint text-muted">
          <kbd>Enter</kbd> to send, <kbd>Shift+Enter</kbd> for new line
        </div>
      </div>
    </div>
    </div><!-- /chat-main -->

    <!-- Right pane: conversation history -->
    <aside class="history-pane" :class="{ open: historyPaneOpen }">
      <div class="history-header d-flex align-items-center justify-content-between">
        <span class="history-title">History</span>
        <button class="btn btn-sm btn-link p-0" @click="historyPaneOpen = false">
          <i class="fas fa-times text-muted"></i>
        </button>
      </div>
      <div class="history-controls px-2 pb-2">
        <input
          v-model="historySearch"
          type="text"
          class="form-control form-control-sm"
          placeholder="Search history..."
        />
        <select v-model="historyTypeFilter" class="form-select form-select-sm mt-1">
          <option value="">All types</option>
          <option v-for="(cfg, key) in InputTypeConfig" :key="key" :value="key">{{ cfg.label }}</option>
        </select>
      </div>
      <div class="history-list">
        <div v-if="store.records.length === 0" class="text-center text-muted py-4" style="font-size:0.82rem">
          No saved records yet.
        </div>
        <div
          v-for="rec in store.records"
          :key="rec.id"
          class="history-item"
          @click="expandedRecordId === rec.id ? (expandedRecordId = '') : (expandedRecordId = rec.id)"
        >
          <div class="d-flex align-items-center gap-1 mb-1">
            <span :class="`badge bg-${InputTypeConfig[rec.input_type]?.color || 'secondary'}`" style="font-size:0.65rem">
              {{ InputTypeConfig[rec.input_type]?.label || rec.input_type }}
            </span>
            <span class="text-muted ms-auto" style="font-size:0.68rem">{{ formatHistoryDate(rec.created_time) }}</span>
          </div>
          <div class="history-item-text">{{ rec.user_input }}</div>
          <div v-if="expandedRecordId === rec.id" class="history-item-detail mt-1">
            <div class="text-muted" style="font-size:0.78rem">{{ rec.response_summary }}</div>
          </div>
        </div>
        <div v-if="store.totalRecords > store.records.length" class="text-center py-2">
          <button class="btn btn-sm btn-outline-primary" @click="loadMoreHistory">Load more</button>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue';
import { marked } from 'marked';
import { useChatRecordStore } from '@/stores/chatRecordStore';
import { apiService } from '@/services/apiService';
import { InputTypeConfig, type ConversationTurn } from '@/types/chatRecord';

const store = useChatRecordStore();
const inputText = ref('');
const inputEl = ref<HTMLTextAreaElement | null>(null);
const messagesContainer = ref<HTMLDivElement | null>(null);
const sidebarCollapsed = ref(false);
const expandedPhases = ref<Set<string>>(new Set());

// History pane
const historyPaneOpen = ref(false);
const historySearch = ref('');
const historyTypeFilter = ref('');
const expandedRecordId = ref('');
const historyPage = ref(1);

// Skills
interface SkillItem {
  name: string;
  snippet: string;
  path: string;
  category: string;
  type: string;
}
const availableSkills = ref<SkillItem[]>([]);
const selectedSkillPath = ref<string>('');
const activeSkillContent = ref<string>('');
const skillSearchQuery = ref('');
const skillsExpanded = ref(false);
const loadingSkillContent = ref(false);

const filteredSkills = computed(() => {
  let list = availableSkills.value;
  if (showFavoritesOnly.value) {
    list = list.filter(s => (skillRatingsMap.value[s.path]?.score ?? 0) > 0);
  }
  if (skillSearchQuery.value) {
    const q = skillSearchQuery.value.toLowerCase();
    list = list.filter(
      s => s.name.toLowerCase().includes(q) || s.snippet.toLowerCase().includes(q) || s.category.toLowerCase().includes(q)
    );
  }
  return list;
});

const skillCategories = computed(() => {
  const cats = new Map<string, SkillItem[]>();
  for (const s of filteredSkills.value) {
    const cat = s.category || 'other';
    if (!cats.has(cat)) cats.set(cat, []);
    cats.get(cat)!.push(s);
  }
  return cats;
});

const selectedSkillName = computed(() => {
  if (!selectedSkillPath.value) return '';
  const s = availableSkills.value.find(s => s.path === selectedSkillPath.value);
  return s?.name ?? '';
});

async function loadSkills() {
  try {
    const resp = await apiService.get('/codemate/skills');
    availableSkills.value = resp.data?.data ?? [];
  } catch {
    availableSkills.value = [];
  }
}

async function selectSkill(skill: SkillItem) {
  if (selectedSkillPath.value === skill.path) {
    selectedSkillPath.value = '';
    activeSkillContent.value = '';
    return;
  }
  selectedSkillPath.value = skill.path;
  loadingSkillContent.value = true;
  try {
    const resp = await apiService.get(`/codemate/skills/content?path=${encodeURIComponent(skill.path)}`);
    activeSkillContent.value = resp.data?.content ?? '';
  } catch {
    activeSkillContent.value = '';
  } finally {
    loadingSkillContent.value = false;
  }
}

function clearSkill() {
  selectedSkillPath.value = '';
  activeSkillContent.value = '';
}

// Skill ratings
const skillRatingsMap = ref<Record<string, any>>({});
const showFavoritesOnly = ref(false);

async function loadSkillRatings() {
  try {
    skillRatingsMap.value = await apiService.getSkillRatingsMap();
  } catch {
    skillRatingsMap.value = {};
  }
}

function getSkillScore(path: string): number {
  return skillRatingsMap.value[path]?.score ?? 0;
}

async function rateSkillStar(skill: SkillItem, star: number) {
  const current = getSkillScore(skill.path);
  const newScore = current === star ? 0 : star;
  if (newScore === 0) {
    const existing = skillRatingsMap.value[skill.path];
    if (existing?.id) {
      await apiService.deleteSkillRating(existing.id);
      delete skillRatingsMap.value[skill.path];
    }
    return;
  }
  try {
    const result = await apiService.rateSkill({
      skill_path: skill.path,
      skill_name: skill.name,
      category: skill.category,
      score: newScore,
      favorited: newScore >= 4,
    });
    skillRatingsMap.value[skill.path] = result;
  } catch { /* ignore */ }
}

interface TaskItem {
  label: string;
  prompt: string;
}

interface SdlcPhase {
  id: string;
  name: string;
  icon: string;
  tasks: TaskItem[];
}

const sdlcPhases: SdlcPhase[] = [
  {
    id: 'planning',
    name: 'Planning & Requirements',
    icon: 'fas fa-clipboard-list',
    tasks: [
      { label: 'Define project scope', prompt: 'Help me define the project scope including goals, deliverables, constraints, and assumptions for: ' },
      { label: 'Estimate effort & cost', prompt: 'Help me estimate the effort, timeline, and cost for a software project that involves: ' },
      { label: 'Identify risks', prompt: 'Identify potential risks, dependencies, and mitigation strategies for a project that: ' },
      { label: 'Write user stories', prompt: 'Write well-structured user stories with acceptance criteria for: ' },
    ],
  },
  {
    id: 'requirements',
    name: 'Defining Requirements',
    icon: 'fas fa-file-alt',
    tasks: [
      { label: 'Write SRS document', prompt: 'Help me draft a Software Requirement Specification (SRS) document for: ' },
      { label: 'Define functional requirements', prompt: 'List and describe the functional requirements for: ' },
      { label: 'Define non-functional requirements', prompt: 'Define the non-functional requirements (performance, security, scalability, usability) for: ' },
      { label: 'Create use cases', prompt: 'Create detailed use cases with actors, preconditions, main flow, and alternate flows for: ' },
    ],
  },
  {
    id: 'design',
    name: 'System Design',
    icon: 'fas fa-drafting-compass',
    tasks: [
      { label: 'Design architecture', prompt: 'Propose a software architecture (components, layers, data flow) for: ' },
      { label: 'Design API endpoints', prompt: 'Design RESTful API endpoints with request/response schemas for: ' },
      { label: 'Design database schema', prompt: 'Design a database schema (tables, relationships, indexes) for: ' },
      { label: 'Create sequence diagram', prompt: 'Describe a sequence diagram showing interactions between components for: ' },
      { label: 'Evaluate tech stack', prompt: 'Compare and recommend a technology stack for: ' },
    ],
  },
  {
    id: 'implementation',
    name: 'Implementation',
    icon: 'fas fa-code',
    tasks: [
      { label: 'Generate boilerplate code', prompt: 'Generate boilerplate/scaffolding code for: ' },
      { label: 'Implement a feature', prompt: 'Implement the following feature with clean, production-ready code: ' },
      { label: 'Refactor code', prompt: 'Refactor the following code for better readability, performance, and maintainability: ' },
      { label: 'Write utility function', prompt: 'Write a well-tested utility function that: ' },
      { label: 'Explain code snippet', prompt: 'Explain the following code step by step: ' },
    ],
  },
  {
    id: 'testing',
    name: 'Testing',
    icon: 'fas fa-vial',
    tasks: [
      { label: 'Write unit tests', prompt: 'Write comprehensive unit tests with edge cases for: ' },
      { label: 'Write integration tests', prompt: 'Write integration tests covering the interaction between: ' },
      { label: 'Generate test cases', prompt: 'Generate a test case matrix (positive, negative, boundary) for: ' },
      { label: 'Review test coverage', prompt: 'Analyze and suggest improvements for test coverage of: ' },
    ],
  },
  {
    id: 'deployment',
    name: 'Deployment',
    icon: 'fas fa-rocket',
    tasks: [
      { label: 'Write Dockerfile', prompt: 'Write an optimized, multi-stage Dockerfile for: ' },
      { label: 'Create CI/CD pipeline', prompt: 'Design a CI/CD pipeline (build, test, deploy stages) for: ' },
      { label: 'Write deployment script', prompt: 'Write a deployment script or Kubernetes manifests for: ' },
      { label: 'Create rollback plan', prompt: 'Create a rollback plan and checklist for deploying: ' },
    ],
  },
  {
    id: 'maintenance',
    name: 'Maintenance',
    icon: 'fas fa-wrench',
    tasks: [
      { label: 'Debug an issue', prompt: 'Help me debug the following issue. The symptoms are: ' },
      { label: 'Write RCA document', prompt: 'Write a Root Cause Analysis (RCA) document for an incident where: ' },
      { label: 'Optimize performance', prompt: 'Suggest performance optimizations for: ' },
      { label: 'Plan migration', prompt: 'Create a migration plan (steps, rollback, validation) to: ' },
      { label: 'Update documentation', prompt: 'Write or update the developer documentation for: ' },
    ],
  },
];

function togglePhase(id: string) {
  const next = new Set(expandedPhases.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  expandedPhases.value = next;
}

function sendTaskPrompt(_phase: string, task: TaskItem) {
  inputText.value = task.prompt;
  nextTick(() => {
    inputEl.value?.focus();
    autoResize();
  });
}

const suggestionChips = [
  { text: 'gRPC vs REST for microservices', icon: 'fas fa-search' },
  { text: 'Design API for file sync service', icon: 'fas fa-drafting-compass' },
  { text: 'Best practices for Go error handling', icon: 'fas fa-code' },
  { text: 'How to structure integration tests', icon: 'fas fa-vial' },
  { text: 'CI/CD pipeline for a Go service', icon: 'fas fa-rocket' },
  { text: 'Learn Kubernetes basics', icon: 'fas fa-book' },
];

onMounted(() => {
  if (!store.sessionId) store.generateSessionId();
  inputEl.value?.focus();
  loadSkills();
  loadSkillRatings();
});

watch(
  () => store.turns.length,
  () => nextTick(scrollToBottom),
);

watch(
  () => store.turns.map((t) => t.response),
  () => nextTick(scrollToBottom),
  { deep: true },
);

// Scroll as streaming tokens arrive
watch(
  () => store.streamingText,
  () => nextTick(scrollToBottom),
);

// Load history when pane opens, or when filters change
watch(historyPaneOpen, (open) => {
  if (open) fetchHistory();
});
watch([historySearch, historyTypeFilter], () => {
  if (historyPaneOpen.value) {
    historyPage.value = 1;
    fetchHistory();
  }
});

function scrollToBottom() {
  const el = messagesContainer.value;
  if (el) el.scrollTop = el.scrollHeight;
}

function autoResize() {
  const el = inputEl.value;
  if (!el) return;
  el.style.height = 'auto';
  el.style.height = Math.min(el.scrollHeight, 200) + 'px';
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleSubmit();
  }
}

function handleSubmit() {
  const text = inputText.value.trim();
  if (!text || store.submitLoading) return;
  store.submitStream(text, activeSkillContent.value || undefined);
  inputText.value = '';
  nextTick(() => {
    if (inputEl.value) {
      inputEl.value.style.height = 'auto';
    }
  });
}

function submitChip(text: string) {
  store.submitStream(text, activeSkillContent.value || undefined);
}

async function handleFinishAndSave() {
  await store.confirmAll();
}

function renderMarkdown(md: string): string {
  if (!md) return '';
  return marked.parse(md, { async: false }) as string;
}

async function fetchHistory() {
  const type = historyTypeFilter.value || undefined;
  const search = historySearch.value || undefined;
  await store.fetchRecords(historyPage.value, 20, type as any, search);
}

function loadMoreHistory() {
  historyPage.value += 1;
  fetchHistory();
}

function formatHistoryDate(dateStr: string): string {
  try {
    const d = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    if (diff < 60_000) return 'just now';
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
    return d.toLocaleDateString();
  } catch {
    return dateStr;
  }
}

function renderResponse(turn: ConversationTurn): string {
  if (!turn.response) return '';
  const p = turn.response.response_payload;
  if (!p) return '<em>Empty response</em>';

  const parts: string[] = [];

  if (p.summary) parts.push(`**Summary:** ${p.summary}`);
  if (p.introduction) parts.push(`${p.introduction}`);
  if (p.problem_statement) parts.push(`**Problem:** ${p.problem_statement}`);
  if (p.explanation) parts.push(p.explanation);
  if (p.answer) parts.push(p.answer);

  if (p.options?.length) {
    parts.push('### Options\n');
    for (const opt of p.options) {
      let line = `**${opt.name || 'Option'}**`;
      if (opt.pros) line += `\n- Pros: ${opt.pros}`;
      if (opt.cons) line += `\n- Cons: ${opt.cons}`;
      if (opt.summary) line += `\n- ${opt.summary}`;
      parts.push(line);
    }
  }

  if (p.key_concepts?.length) {
    parts.push('### Key Concepts\n');
    for (const c of p.key_concepts) {
      parts.push(`- **${c.name}**: ${c.description}`);
    }
  }

  if (p.approach_options?.length) {
    parts.push('### Approach Options\n');
    for (const a of p.approach_options) parts.push(`- ${a}`);
  }
  if (p.chosen_approach) parts.push(`**Chosen Approach:** ${p.chosen_approach}`);

  if (p.components?.length) {
    parts.push(`**Components:** ${p.components.join(', ')}`);
  }

  if (p.trade_offs) parts.push(`**Trade-offs:** ${p.trade_offs}`);
  if (p.recommendation) parts.push(`**Recommendation:** ${p.recommendation}`);
  if (p.risks) parts.push(`**Risks:** ${p.risks}`);

  if (p.chat_path?.length) {
    parts.push('### Learning Path\n');
    for (const step of p.chat_path) {
      let line = `${step.order}. **${step.title}**`;
      if (step.duration) line += ` *(${step.duration})*`;
      line += `\n   ${step.description}`;
      parts.push(line);
    }
  }

  if (p.plan?.length) {
    parts.push('### Execution Plan\n');
    for (let i = 0; i < p.plan.length; i++) {
      parts.push(`${i + 1}. ${p.plan[i]}`);
    }
  }

  if (p.resources?.length) {
    parts.push('### Resources\n');
    for (const r of p.resources) {
      const badge = `\`${r.type}\``;
      const link = r.url ? `[${r.title}](${r.url})` : r.title;
      parts.push(`- ${badge} ${link}`);
    }
  }

  if (p.prerequisites?.length) {
    parts.push(`**Prerequisites:** ${p.prerequisites.join(', ')}`);
  }
  if (p.time_estimate) parts.push(`**Estimated Time:** ${p.time_estimate}`);
  if (p.pronunciation) parts.push(`**Pronunciation:** \`${p.pronunciation}\``);
  if (p.example) parts.push(`> ${p.example}`);

  if (p.references?.length) {
    parts.push('### References\n');
    for (const ref of p.references) parts.push(`- ${ref}`);
  }

  const md = parts.join('\n\n');
  return marked.parse(md, { async: false }) as string;
}
</script>

<style scoped>
.chat-page {
  height: calc(100vh - 56px);
  background: #f7f7f8;
}

/* Sidebar */
.task-sidebar {
  width: 260px;
  min-width: 260px;
  background: #fff;
  border-right: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  transition: width 0.2s, min-width 0.2s;
}

.task-sidebar.collapsed {
  width: 44px;
  min-width: 44px;
}

.sidebar-header {
  padding: 0.6rem 0.75rem;
  border-bottom: 1px solid #eee;
}

.sidebar-title {
  font-weight: 600;
  font-size: 0.88rem;
  color: #333;
}

.sidebar-toggle {
  color: #888;
  padding: 0.15rem 0.35rem;
  font-size: 0.85rem;
}

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 0.25rem 0;
}

.phase-group {
  border-bottom: 1px solid #f0f0f0;
}

.phase-header {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: none;
  background: none;
  font-size: 0.82rem;
  font-weight: 600;
  color: #444;
  cursor: pointer;
  text-align: left;
  gap: 8px;
  transition: background 0.12s;
}

.phase-header:hover {
  background: #f5f5ff;
}

.phase-icon {
  font-size: 0.85rem;
  width: 18px;
  text-align: center;
  color: #667eea;
}

.phase-chevron {
  font-size: 0.7rem;
  color: #aaa;
}

.task-list {
  list-style: none;
  padding: 0;
  margin: 0 0 0.25rem;
}

.task-item {
  display: block;
  width: 100%;
  padding: 0.4rem 0.75rem 0.4rem 2.2rem;
  border: none;
  background: none;
  font-size: 0.8rem;
  color: #555;
  text-align: left;
  cursor: pointer;
  transition: background 0.1s, color 0.1s;
}

.task-item:hover {
  background: #eef0ff;
  color: #667eea;
}

.slide-enter-active,
.slide-leave-active {
  transition: max-height 0.2s ease, opacity 0.2s ease;
  overflow: hidden;
}

.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
}

.slide-enter-to,
.slide-leave-from {
  max-height: 300px;
  opacity: 1;
}

/* Main chat column */
.chat-main {
  min-width: 0;
}

.chat-topbar {
  flex-shrink: 0;
  z-index: 10;
}

@media (max-width: 768px) {
  .task-sidebar {
    display: none;
  }
}

.chat-messages {
  overflow-y: auto;
  padding: 0;
}

/* Welcome screen */
.welcome-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 2rem;
}

.welcome-content {
  text-align: center;
  max-width: 600px;
}

.welcome-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-size: 1.8rem;
}

.chip-btn {
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 20px;
  padding: 0.5rem 1rem;
  font-size: 0.88rem;
  color: #333;
  cursor: pointer;
  transition: all 0.15s;
}

.chip-btn:hover {
  border-color: #667eea;
  color: #667eea;
  background: #f0f0ff;
}

/* Messages */
.messages-list {
  max-width: 800px;
  margin: 0 auto;
  padding: 1.5rem 1rem;
}

.message-row {
  display: flex;
  gap: 12px;
  margin-bottom: 1.5rem;
  align-items: flex-start;
}

.user-row {
  justify-content: flex-end;
}

.assistant-row {
  justify-content: flex-start;
}

.avatar {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
}

.user-avatar {
  background: #667eea;
  color: #fff;
  order: 2;
}

.assistant-avatar {
  background: #10a37f;
  color: #fff;
}

.message-bubble {
  max-width: 75%;
  border-radius: 16px;
  padding: 0.75rem 1rem;
  line-height: 1.6;
  font-size: 0.93rem;
}

.user-bubble {
  background: #667eea;
  color: #fff;
  border-bottom-right-radius: 4px;
}

.assistant-bubble {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-bottom-left-radius: 4px;
}

.message-text {
  white-space: pre-wrap;
  word-break: break-word;
}

/* Typing indicator */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 4px 0;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #999;
  border-radius: 50%;
  animation: typing-bounce 1.4s infinite both;
}

.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing-bounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-6px); opacity: 1; }
}

/* Response markdown */
.response-content :deep(h3) {
  font-size: 1rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 0.5rem;
  color: #333;
}

.response-content :deep(p) {
  margin-bottom: 0.5rem;
}

.response-content :deep(ul),
.response-content :deep(ol) {
  padding-left: 1.2rem;
  margin-bottom: 0.5rem;
}

.response-content :deep(li) {
  margin-bottom: 0.25rem;
}

.response-content :deep(code) {
  font-size: 0.85em;
  background: #f0f0f0;
  padding: 0.1em 0.35em;
  border-radius: 3px;
  color: #c7254e;
}

.response-content :deep(pre) {
  background: #f6f8fa;
  border: 1px solid #e1e4e8;
  border-radius: 8px;
  padding: 0.75rem;
  overflow-x: auto;
  font-size: 0.85rem;
  margin: 0.5rem 0;
}

.response-content :deep(pre code) {
  background: none;
  padding: 0;
  color: inherit;
}

.response-content :deep(blockquote) {
  border-left: 3px solid #667eea;
  padding: 0.5rem 0.75rem;
  margin: 0.5rem 0;
  background: #f8f8ff;
  border-radius: 0 6px 6px 0;
  color: #555;
  font-style: italic;
}

.response-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 0.5rem 0;
  font-size: 0.88rem;
}

.response-content :deep(th),
.response-content :deep(td) {
  border: 1px solid #e1e4e8;
  padding: 0.4rem 0.6rem;
}

.response-content :deep(th) {
  background: #f6f8fa;
  font-weight: 600;
}

.response-content :deep(a) {
  color: #667eea;
}

/* Input bar */
.chat-input-bar {
  flex-shrink: 0;
  padding: 0.75rem 1rem 0.5rem;
}

.input-container {
  max-width: 800px;
  margin: 0 auto;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  background: #fff;
  border: 1px solid #d9d9e3;
  border-radius: 16px;
  padding: 0.5rem 0.5rem 0.5rem 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-wrapper:focus-within {
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.15);
}

.chat-input {
  flex: 1;
  border: none;
  outline: none;
  resize: none;
  font-size: 0.95rem;
  line-height: 1.5;
  max-height: 200px;
  background: transparent;
  font-family: inherit;
}

.send-btn {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  background: #d9d9e3;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 0.9rem;
}

.send-btn.active {
  background: #667eea;
}

.send-btn:disabled {
  cursor: default;
}

.input-hint {
  font-size: 0.75rem;
  text-align: center;
  padding-top: 0.35rem;
}

.input-hint kbd {
  background: #eee;
  border-radius: 3px;
  padding: 0 4px;
  font-size: 0.72rem;
  border: 1px solid #ddd;
}

/* Skill picker */
.active-skill-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.35rem 0.75rem;
  background: #fff3cd;
  border-bottom: 1px solid #ffe69c;
  font-size: 0.78rem;
  font-weight: 600;
  color: #664d03;
}

.skill-picker {
  padding: 0.35rem 0.5rem 0.5rem;
}

.skill-search {
  margin-bottom: 0.4rem;
  font-size: 0.78rem;
}

.skill-categories {
  max-height: 280px;
  overflow-y: auto;
}

.skill-cat-group {
  margin-bottom: 0.35rem;
}

.skill-cat-label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  color: #999;
  padding: 0.15rem 0.25rem;
  letter-spacing: 0.03em;
}

.skill-filter-btn {
  border: 1px solid #ddd;
  background: #fff;
  color: #ccc;
  padding: 0.15rem 0.4rem;
  font-size: 0.78rem;
}

.skill-filter-btn.active {
  color: #f5a623;
  border-color: #f5a623;
  background: #fef9ef;
}

.skill-row {
  display: flex;
  align-items: center;
  gap: 2px;
}

.skill-option {
  flex: 1;
  min-width: 0;
  text-align: left;
  border: none;
  background: transparent;
  padding: 0.25rem 0.5rem;
  font-size: 0.78rem;
  border-radius: 4px;
  cursor: pointer;
  color: #444;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.skill-option:hover {
  background: #f0f0f5;
}

.skill-option.active {
  background: #667eea;
  color: #fff;
}

.skill-name {
  pointer-events: none;
}

.skill-stars {
  flex-shrink: 0;
  display: flex;
  gap: 1px;
  padding-right: 0.25rem;
}

.star-icon {
  font-size: 0.6rem;
  cursor: pointer;
  color: #ddd;
  transition: color 0.1s;
}

.star-icon.rated {
  color: #f5a623;
}

.star-icon:hover {
  color: #f5a623;
}

/* Streaming cursor */
.streaming-cursor {
  display: inline-block;
  animation: blink 1s step-end infinite;
  color: #667eea;
  font-size: 0.9rem;
  vertical-align: text-bottom;
}

@keyframes blink {
  50% { opacity: 0; }
}

.streaming-indicator .badge {
  font-size: 0.7rem;
  font-weight: 500;
}

/* History right pane */
.history-pane {
  width: 0;
  min-width: 0;
  overflow: hidden;
  background: #fff;
  border-left: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  transition: width 0.25s ease, min-width 0.25s ease;
}

.history-pane.open {
  width: 300px;
  min-width: 300px;
}

.history-header {
  padding: 0.6rem 0.75rem;
  border-bottom: 1px solid #eee;
  flex-shrink: 0;
}

.history-title {
  font-weight: 600;
  font-size: 0.88rem;
  color: #333;
}

.history-controls {
  flex-shrink: 0;
  padding-top: 0.5rem;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.25rem 0;
}

.history-item {
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid #f2f2f2;
  cursor: pointer;
  transition: background 0.12s;
}

.history-item:hover {
  background: #f8f8ff;
}

.history-item-text {
  font-size: 0.82rem;
  color: #444;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.history-item-detail {
  border-top: 1px solid #f0f0f0;
  padding-top: 0.35rem;
  max-height: 120px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .message-bubble {
    max-width: 90%;
  }

  .welcome-content p br {
    display: none;
  }

  .suggestion-chips {
    padding: 0 0.5rem;
  }

  .history-pane.open {
    width: 100%;
    min-width: 100%;
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    z-index: 20;
  }
}
</style>
