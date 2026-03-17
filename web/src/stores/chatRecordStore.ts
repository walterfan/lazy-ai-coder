import { defineStore } from 'pinia';
import { apiService } from '@/services/apiService';
import { showToast } from '@/utils/toast';
import { useSettingsStore } from '@/stores/settingsStore';
import type {
  InputType,
  ChatRecord,
  ChatRecordSummary,
  SubmitChatRecordResponse,
  ChatRecordStats,
  ConversationTurn,
} from '@/types/chatRecord';

interface ChatRecordState {
  // Multi-turn conversation
  turns: ConversationTurn[];
  submitLoading: boolean;
  confirmLoading: boolean;
  sessionId: string;
  error: string | null;

  // History state (for future use)
  records: ChatRecordSummary[];
  totalRecords: number;
  page: number;
  pageSize: number;
  stats: ChatRecordStats | null;
}

export const useChatRecordStore = defineStore('ChatRecord', {
  state: (): ChatRecordState => ({
    turns: [],
    submitLoading: false,
    confirmLoading: false,
    sessionId: '',
    error: null,
    records: [],
    totalRecords: 0,
    page: 1,
    pageSize: 10,
    stats: null,
  }),

  getters: {
    hasResponse: (state) => state.turns.length > 0 && state.turns[state.turns.length - 1].response !== null,
    currentTurn: (state): ConversationTurn | null =>
      state.turns.length > 0 ? state.turns[state.turns.length - 1] : null,
    currentInput: (state): string => {
      const last = state.turns.length > 0 ? state.turns[state.turns.length - 1] : null;
      return last?.userInput ?? '';
    },
    currentResponse: (state): SubmitChatRecordResponse | null => {
      const last = state.turns.length > 0 ? state.turns[state.turns.length - 1] : null;
      return last?.response ?? null;
    },
    inputType: (state) => state.turns.length > 0 ? state.turns[state.turns.length - 1].response?.input_type ?? null : null,
    responsePayload: (state) => state.turns.length > 0 ? state.turns[state.turns.length - 1].response?.response_payload ?? null : null,
    similarRecords: (state) => state.turns.length > 0 ? (state.turns[state.turns.length - 1].response?.similar_records ?? []) : [],
  },

  actions: {
    // Generate a new session ID
    generateSessionId() {
      this.sessionId = crypto.randomUUID();
    },

    // Submit user input (adds a turn; backend uses session for multi-round context)
    async submit(input: string) {
      if (!input.trim()) {
        showToast('Please enter some text', 'warning');
        return;
      }

      this.submitLoading = true;
      this.error = null;
      this.turns.push({ userInput: input.trim(), response: null });

      try {
        const settingsStore = useSettingsStore();
        const payload: Record<string, string | undefined> = {
          user_input: input.trim(),
          session_id: this.sessionId || undefined,
        };
        if (settingsStore.LLM_API_KEY) {
          payload.LLM_API_KEY = settingsStore.LLM_API_KEY;
          if (settingsStore.LLM_BASE_URL) payload.LLM_BASE_URL = settingsStore.LLM_BASE_URL;
          if (settingsStore.LLM_MODEL) payload.LLM_MODEL = settingsStore.LLM_MODEL;
          if (settingsStore.LLM_TEMPERATURE) payload.LLM_TEMPERATURE = settingsStore.LLM_TEMPERATURE;
        }
        const response = await apiService.post<SubmitChatRecordResponse>(
          '/chat-record/submit',
          payload
        );

        const lastIdx = this.turns.length - 1;
        this.turns[lastIdx] = { ...this.turns[lastIdx], response: response.data };
        if (response.data.session_id) {
          this.sessionId = response.data.session_id;
        }
      } catch (error: unknown) {
        const lastIdx = this.turns.length - 1;
        this.turns.splice(lastIdx, 1);
        const err = error as { response?: { data?: { error?: string }; status?: number }; message?: string };
        const message =
          err.response?.data?.error ||
          (error instanceof Error ? error.message : 'Failed to process input');
        this.error = message;
        showToast(message, 'danger');
      } finally {
        this.submitLoading = false;
      }
    },

    // Save the current (last) response to database
    async confirm() {
      const last = this.turns.length > 0 ? this.turns[this.turns.length - 1] : null;
      if (!last?.response) {
        showToast('No response to save', 'warning');
        return;
      }

      this.confirmLoading = true;
      this.error = null;

      try {
        await apiService.post<ChatRecord>('/chat-record/confirm', {
          user_input: last.userInput,
          input_type: last.response.input_type,
          response_payload: last.response.response_payload,
        });

        showToast('Record saved!', 'success');
        this.turns.pop();
        if (this.turns.length === 0) {
          this.generateSessionId();
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to save record';
        this.error = message;
        showToast(message, 'danger');
      } finally {
        this.confirmLoading = false;
      }
    },

    // Finish conversation: save all turns to database, then clear
    async confirmAll() {
      const toSave = this.turns.filter((t) => t.response !== null);
      if (toSave.length === 0) {
        showToast('No responses to save', 'warning');
        return;
      }

      this.confirmLoading = true;
      this.error = null;

      try {
        for (const turn of toSave) {
          if (!turn.response) continue;
          await apiService.post<ChatRecord>('/chat-record/confirm', {
            user_input: turn.userInput,
            input_type: turn.response.input_type,
            response_payload: turn.response.response_payload,
          });
        }
        showToast(`Saved ${toSave.length} record(s) to history.`, 'success');
        this.clearSession();
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to save records';
        this.error = message;
        showToast(message, 'danger');
      } finally {
        this.confirmLoading = false;
      }
    },

    clearResponse() {
      if (this.turns.length > 0 && this.turns[this.turns.length - 1].response === null) {
        this.turns.pop();
      }
      this.error = null;
    },

    clearSession() {
      this.turns = [];
      this.error = null;
      this.generateSessionId();
    },

    // Fetch chat record history
    async fetchRecords(
      page: number = 1,
      pageSize: number = 10,
      type?: InputType,
      search?: string
    ) {
      try {
        const params = new URLSearchParams();
        params.append('page', page.toString());
        params.append('page_size', pageSize.toString());
        if (type) params.append('type', type);
        if (search) params.append('search', search);

        const response = await apiService.get(`/chat-record/list?${params.toString()}`);
        const data = response.data;

        this.records = data.records || [];
        this.totalRecords = data.total || 0;
        this.page = data.page || 1;
        this.pageSize = data.page_size || 10;
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to fetch records';
        showToast(message, 'danger');
      }
    },

    // Fetch statistics
    async fetchStats() {
      try {
        const response = await apiService.get('/chat-record/stats');
        this.stats = response.data;
      } catch (error) {
        console.error('Failed to fetch stats:', error);
      }
    },

    // Delete a record
    async deleteRecord(id: string) {
      try {
        await apiService.delete(`/chat-record/${id}`);
        showToast('Record deleted', 'success');
        // Refresh the list
        await this.fetchRecords(this.page, this.pageSize);
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to delete record';
        showToast(message, 'danger');
      }
    },
  },
});
