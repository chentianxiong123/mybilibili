import { ref, computed } from 'vue'

export const compactLevel = ref(0)

export const desktopCompact = computed(() => compactLevel.value > 0)

export function useZoomCompact() {
  return { compactLevel, desktopCompact }
}