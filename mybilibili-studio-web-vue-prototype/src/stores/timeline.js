import { defineStore } from 'pinia'

/**
 * Timeline store - 改造自 openreel timeline-store.ts
 * 状态:播放头 / 缩放 / 滚动 / 播放状态
 */
export const ZOOM_MIN = 10
export const ZOOM_DEFAULT = 60
export const ZOOM_MAX = 400

export const useTimelineStore = defineStore('timeline', {
  state: () => ({
    playheadPosition: 0,
    playbackState: 'stopped',  // stopped | playing | paused
    playbackRate: 1,
    pixelsPerSecond: ZOOM_DEFAULT,
    scrollX: 0,
    scrollY: 0,
    viewportWidth: 1200,
    viewportHeight: 280,
    trackHeight: 56,
    loopEnabled: false,
    loopStart: 0,
    loopEnd: 10,
    isScrubbing: false
  }),

  actions: {
    setZoom(zoom) {
      this.pixelsPerSecond = Math.max(ZOOM_MIN, Math.min(ZOOM_MAX, zoom))
    },
    setPlayhead(t) {
      this.playheadPosition = Math.max(0, t)
    },
    play() {
      if (this.playbackState === 'playing') return
      this.playbackState = 'playing'
    },
    pause() {
      if (this.playbackState !== 'playing') return
      this.playbackState = 'paused'
    },
    togglePlay() {
      this.playbackState = this.playbackState === 'playing' ? 'paused' : 'playing'
    },
    setScroll(x, y) {
      this.scrollX = x
      this.scrollY = y
    }
  }
})
