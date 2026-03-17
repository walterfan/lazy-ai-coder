import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type {
  SmartPromptRequest,
  SmartPromptResponse,
  Preset
} from '@/types/smart-prompt';

export const useSmartPromptStore = defineStore('smartPrompt', () => {
  // State
  const loading = ref(false);
  const error = ref<string | null>(null);
  const currentResponse = ref<SmartPromptResponse | null>(null);
  const presets = ref<Preset[]>([]);
  const selectedPresetId = ref<string>('');

  // Getters
  const hasResponse = computed(() => currentResponse.value !== null);
  const selectedPreset = computed(() =>
    presets.value.find(p => p.id === selectedPresetId.value)
  );

  // Actions
  async function loadPresets() {
    try {
      loading.value = true;
      error.value = null;
      const response = await fetch('/api/v1/smart-prompt/presets');
      if (!response.ok) {
        throw new Error('Failed to load presets');
      }
      presets.value = await response.json();
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unknown error';
      console.error('Failed to load presets:', err);
    } finally {
      loading.value = false;
    }
  }

  async function generatePrompt(request: SmartPromptRequest): Promise<SmartPromptResponse | null> {
    try {
      loading.value = true;
      error.value = null;

      const response = await fetch('/api/v1/smart-prompt/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to generate prompt');
      }

      currentResponse.value = await response.json();
      return currentResponse.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unknown error';
      console.error('Failed to generate prompt:', err);
      return null;
    } finally {
      loading.value = false;
    }
  }

  function clearResponse() {
    currentResponse.value = null;
    error.value = null;
  }

  function setSelectedPreset(presetId: string) {
    selectedPresetId.value = presetId;
  }

  return {
    // State
    loading,
    error,
    currentResponse,
    presets,
    selectedPresetId,

    // Getters
    hasResponse,
    selectedPreset,

    // Actions
    loadPresets,
    generatePrompt,
    clearResponse,
    setSelectedPreset,
  };
});
