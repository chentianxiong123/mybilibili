import { ref } from 'vue'

export const desktopCompact = ref(false)

export function useZoomCompact() {
  return { desktopCompact }
}