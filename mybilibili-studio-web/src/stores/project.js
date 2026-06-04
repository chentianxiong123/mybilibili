import { defineStore } from 'pinia'

/**
 * 项目 store - 改造自 openreel project-store.ts
 * 状态:project / mediaItems / tracks / clips / selection
 */
let mid = 0
const id = (p) => `${p}_${(++mid).toString(36)}`

export const useProjectStore = defineStore('project', {
  state: () => ({
    project: {
      id: id('proj'),
      name: '未命名项目',
      width: 1920,
      height: 1080,
      fps: 30,
      duration: 60,
      backgroundColor: '#000000'
    },
    mediaItems: [],
    tracks: [
      { id: 'v1', label: 'V1', kind: 'video', height: 56, muted: false, clips: [] },
      { id: 'v2', label: 'V2', kind: 'video', height: 56, muted: false, clips: [] },
      { id: 'a1', label: 'A1', kind: 'audio', height: 48, muted: false, clips: [] },
      { id: 'a2', label: 'A2', kind: 'audio', height: 48, muted: false, clips: [] },
      { id: 't1', label: '字幕', kind: 'subtitle', height: 40, muted: false, clips: [] }
    ],
    selectedClipId: null,
    selectedMediaId: null
  }),

  getters: {
    selectedClip: (s) => {
      for (const t of s.tracks) {
        const c = t.clips.find(c => c.id === s.selectedClipId)
        if (c) return { ...c, trackId: t.id, trackKind: t.kind }
      }
      return null
    },
    totalDuration: (s) => {
      let max = 0
      for (const t of s.tracks) {
        for (const c of t.clips) {
          max = Math.max(max, c.startTime + c.duration)
        }
      }
      return max
    }
  },

  actions: {
    setProjectName(name) { this.project.name = name },

    addMediaItem(item) {
      this.mediaItems.push({ id: id('m'), duration: 0, source: 'local', ...item })
    },

    addClipToTrack(trackId, clip) {
      const track = this.tracks.find(t => t.id === trackId)
      if (!track) return null
      const newClip = { id: id('c'), opacity: 1, volume: 1, ...clip }
      track.clips.push(newClip)
      return newClip
    },

    updateClip(clipId, patch) {
      for (const t of this.tracks) {
        const c = t.clips.find(c => c.id === clipId)
        if (c) { Object.assign(c, patch); return }
      }
    },

    moveClip(clipId, newTrackId, newStartTime) {
      let moved = null
      for (const t of this.tracks) {
        const idx = t.clips.findIndex(c => c.id === clipId)
        if (idx >= 0) {
          moved = t.clips.splice(idx, 1)[0]
          break
        }
      }
      if (!moved) return
      moved.startTime = Math.max(0, newStartTime)
      const target = this.tracks.find(t => t.id === newTrackId)
      if (target) target.clips.push(moved)
    },

    removeClip(clipId) {
      for (const t of this.tracks) {
        const idx = t.clips.findIndex(c => c.id === clipId)
        if (idx >= 0) {
          t.clips.splice(idx, 1)
          if (this.selectedClipId === clipId) this.selectedClipId = null
          return
        }
      }
    },

    splitClipAt(clipId, playhead) {
      for (const t of this.tracks) {
        const c = t.clips.find(c => c.id === clipId)
        if (!c) continue
        const offset = playhead - c.startTime
        if (offset <= 0.05 || offset >= c.duration - 0.05) return null
        const right = { ...c, id: id('c'), startTime: c.startTime + offset, duration: c.duration - offset }
        c.duration = offset
        t.clips.push(right)
        return right.id
      }
      return null
    },

    selectClip(clipId) { this.selectedClipId = clipId },
    selectMedia(mediaId) { this.selectedMediaId = mediaId }
  }
})
