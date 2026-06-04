import { defineStore } from 'pinia'

/**
 * UI store - 改造自 openreel ui-store.ts
 * 状态:面板宽度 / 弹窗 / 主题
 */
const PANEL_DEFAULTS = {
  mediaWidth: 320,
  inspectorWidth: 320,
  timelineHeight: 280
}

export const useUIStore = defineStore('ui', {
  state: () => ({
    ...PANEL_DEFAULTS,
    activeModal: null,        // export | settings | null
    isFullscreenPreview: false,
    showKeyframes: true,
    showRuler: true
  }),

  actions: {
    setMediaWidth(w) { this.mediaWidth = Math.max(240, Math.min(640, w)) },
    setInspectorWidth(w) { this.inspectorWidth = Math.max(260, Math.min(560, w)) },
    setTimelineHeight(h) { this.timelineHeight = Math.max(160, Math.min(560, h)) },
    openModal(name) { this.activeModal = name },
    closeModal() { this.activeModal = null },
    toggleKeyframes() { this.showKeyframes = !this.showKeyframes },
    toggleRuler() { this.showRuler = !this.showRuler }
  }
})
