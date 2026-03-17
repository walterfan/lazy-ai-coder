import { defineStore } from 'pinia';
import type { Settings } from '@/types';
import { storage } from '@/utils/storage';

export const useSettingsStore = defineStore('settings', {
  state: (): Settings => ({
    LLM_API_KEY: '',
    LLM_MODEL: '',
    LLM_BASE_URL: '',
    LLM_TEMPERATURE: '',
    GITLAB_BASE_URL: '',
    GITLAB_TOKEN: '',
  }),

  getters: {
    isConfigured(): boolean {
      return !!(
        this.LLM_API_KEY &&
        this.LLM_MODEL &&
        this.LLM_BASE_URL &&
        this.GITLAB_BASE_URL &&
        this.GITLAB_TOKEN
      );
    },
  },

  actions: {
    loadFromStorage() {
      try {
        this.LLM_API_KEY = storage.getItem('LLM_API_KEY') || '';
        this.LLM_MODEL = storage.getItem('LLM_MODEL') || '';
        this.LLM_BASE_URL = storage.getItem('LLM_BASE_URL') || '';
        this.LLM_TEMPERATURE = storage.getItem('LLM_TEMPERATURE') || '';
        this.GITLAB_BASE_URL = storage.getItem('GITLAB_BASE_URL') || '';
        this.GITLAB_TOKEN = storage.getItem('GITLAB_TOKEN') || '';
      } catch (error) {
        console.error('Failed to load settings:', error);
      }
    },

    saveToStorage() {
      try {
        storage.setItem('LLM_API_KEY', this.LLM_API_KEY);
        storage.setItem('LLM_MODEL', this.LLM_MODEL);
        storage.setItem('LLM_BASE_URL', this.LLM_BASE_URL);
        storage.setItem('LLM_TEMPERATURE', this.LLM_TEMPERATURE);
        storage.setItem('GITLAB_BASE_URL', this.GITLAB_BASE_URL);
        storage.setItem('GITLAB_TOKEN', this.GITLAB_TOKEN);
      } catch (error) {
        console.error('Failed to save settings:', error);
      }
    },

    updateSettings(settings: Partial<Settings>) {
      Object.assign(this, settings);
      this.saveToStorage();
    },

    clearSettings() {
      this.LLM_API_KEY = '';
      this.LLM_MODEL = '';
      this.LLM_BASE_URL = '';
      this.LLM_TEMPERATURE = '';
      this.GITLAB_BASE_URL = '';
      this.GITLAB_TOKEN = '';
      // Only clear settings keys; do NOT wipe auth/session/other app flags.
      storage.removeItem('LLM_API_KEY');
      storage.removeItem('LLM_MODEL');
      storage.removeItem('LLM_BASE_URL');
      storage.removeItem('LLM_TEMPERATURE');
      storage.removeItem('GITLAB_BASE_URL');
      storage.removeItem('GITLAB_TOKEN');
    },
  },
});
