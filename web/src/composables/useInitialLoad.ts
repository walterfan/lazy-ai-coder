import { ref, onMounted } from 'vue';

/**
 * Composable for handling initial data loading in views.
 * Prevents blank page flash by tracking loading state.
 * 
 * Usage:
 * ```ts
 * const { isInitialLoading, runInitialLoad } = useInitialLoad();
 * 
 * onMounted(async () => {
 *   await runInitialLoad(async () => {
 *     await fetchData();
 *   });
 * });
 * ```
 * 
 * In template:
 * ```html
 * <div v-if="isInitialLoading" class="text-center py-5">
 *   <div class="spinner-border text-primary"></div>
 *   <p class="mt-2">Loading...</p>
 * </div>
 * <div v-else>
 *   <!-- Your content -->
 * </div>
 * ```
 */
export function useInitialLoad() {
  const isInitialLoading = ref(true);
  const loadError = ref<string | null>(null);

  /**
   * Run the initial load function and manage loading state
   * @param loadFn - Async function that loads initial data
   * @returns Promise that resolves when loading is complete
   */
  async function runInitialLoad(loadFn: () => Promise<void>): Promise<void> {
    isInitialLoading.value = true;
    loadError.value = null;
    try {
      await loadFn();
    } catch (error) {
      loadError.value = error instanceof Error ? error.message : 'Failed to load data';
      console.error('Initial load error:', error);
    } finally {
      isInitialLoading.value = false;
    }
  }

  /**
   * Run initial load automatically in onMounted
   * @param loadFn - Async function that loads initial data
   */
  function autoLoad(loadFn: () => Promise<void>): void {
    onMounted(() => runInitialLoad(loadFn));
  }

  return {
    isInitialLoading,
    loadError,
    runInitialLoad,
    autoLoad,
  };
}

/**
 * Higher-order composable that combines multiple store loading states
 * 
 * Usage:
 * ```ts
 * const isLoading = useCombinedLoading(
 *   () => projectStore.loading,
 *   () => promptStore.loading
 * );
 * ```
 */
export function useCombinedLoading(...loadingGetters: (() => boolean)[]): () => boolean {
  return () => loadingGetters.some(getter => getter());
}

