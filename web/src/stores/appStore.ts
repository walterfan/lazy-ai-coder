import { defineStore } from 'pinia';
import { apiService } from '@/services/apiService';

interface AppState {
  title: string;
  version: string;
  loaded: boolean;
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    title: 'Lazy AI Coder',
    version: '1.0',
    loaded: false,
  }),

  getters: {
    fullTitle: (state) => `${state.title} v${state.version}`,
  },

  actions: {
    async fetchInfo() {
      if (this.loaded) return;
      try {
        const resp = await apiService.get('/app/info');
        if (resp.data?.title) this.title = resp.data.title;
        if (resp.data?.version) this.version = resp.data.version;
      } catch {
        // keep defaults
      }
      this.loaded = true;
    },
  },
});
