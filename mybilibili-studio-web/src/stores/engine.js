import { defineStore } from 'pinia'

/**
 * Engine store - 改造自 openreel engine-store.ts
 * 状态:渲染引擎(WebAV)生命周期
 */
export const useEngineStore = defineStore('engine', {
  state: () => ({
    isReady: false,
    isRendering: false,
    renderProgress: 0,
    renderError: null,
    lastRenderedUrl: null
  }),

  actions: {
    setReady(v) { this.isReady = v },
    setRendering(v) { this.isRendering = v; if (!v) this.renderProgress = 0 },
    setProgress(p) { this.renderProgress = Math.max(0, Math.min(100, p)) },
    setError(e) { this.renderError = e },
    setRenderedUrl(url) { this.lastRenderedUrl = url }
  }
})
