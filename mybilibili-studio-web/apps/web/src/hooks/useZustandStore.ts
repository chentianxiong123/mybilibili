import { ref, onUnmounted } from "vue";

/**
 * A helper to use a Zustand store reactively inside Vue 3 components.
 * Subscribes to store changes and updates a Vue ref.
 */
export function useZustandStore<T>(store: any) {
  const state = ref<T>(store.getState());

  const unsubscribe = store.subscribe((newState: T) => {
    state.value = newState;
  });

  onUnmounted(() => {
    unsubscribe();
  });

  return state;
}
