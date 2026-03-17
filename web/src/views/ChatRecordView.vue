<template>
  <div class="chat-page d-flex flex-column">
    <!-- Top bar -->
    <div class="chat-topbar d-flex align-items-center justify-content-between px-3 py-2 border-bottom bg-white">
      <div class="d-flex align-items-center gap-2">
        <i class="fas fa-robot text-primary fs-5"></i>
        <span class="fw-semibold">Coding Mate</span>
        <span v-if="store.turns.length > 0" class="badge bg-secondary">{{ store.turns.length }} turns</span>
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
        <router-link to="/chat/history" class="btn btn-sm btn-outline-secondary" title="History">
          <i class="fas fa-history"></i>
        </router-link>
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
              <!-- Loading -->
              <div v-if="turn.response === null" class="typing-indicator">
                <span></span><span></span><span></span>
              </div>
              <!-- Response content -->
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue';
import { marked } from 'marked';
import { useChatRecordStore } from '@/stores/chatRecordStore';
import { InputTypeConfig, type ConversationTurn } from '@/types/chatRecord';

const store = useChatRecordStore();
const inputText = ref('');
const inputEl = ref<HTMLTextAreaElement | null>(null);
const messagesContainer = ref<HTMLDivElement | null>(null);

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
  store.submit(text);
  inputText.value = '';
  nextTick(() => {
    if (inputEl.value) {
      inputEl.value.style.height = 'auto';
    }
  });
}

function submitChip(text: string) {
  store.submit(text);
}

async function handleFinishAndSave() {
  await store.confirmAll();
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

.chat-topbar {
  flex-shrink: 0;
  z-index: 10;
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
}
</style>
