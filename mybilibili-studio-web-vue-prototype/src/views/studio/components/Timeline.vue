<template>
  <div class="tl-root">
    <div class="tl-head">
      <span>时间轴</span>
      <div class="tl-tools">
        <el-button-group>
          <el-button size="small" @click="zoomOut">-</el-button>
          <el-button size="small" @click="zoomReset">{{ timeline.pixelsPerSecond }}px/s</el-button>
          <el-button size="small" @click="zoomIn">+</el-button>
        </el-button-group>
        <el-button size="small" :disabled="!project.selectedClipId" @click="onSplit">分割</el-button>
        <el-button size="small" :disabled="!project.selectedClipId" type="danger" @click="onDelete">删除</el-button>
      </div>
    </div>
    <div class="tl-body" ref="bodyEl" @scroll="onScroll">
      <TimeRuler class="tl-ruler" />
      <div class="tl-tracks">
        <TrackHeader
          v-for="track in project.tracks"
          :key="track.id"
          :track="track"
        />
        <div class="tl-lanes">
          <div
            v-for="track in project.tracks"
            :key="track.id"
            class="tl-lane"
            :style="{ height: track.height + 'px' }"
            @dragover.prevent
            @drop="onDrop($event, track.id)"
          >
            <ClipComponent
              v-for="clip in track.clips"
              :key="clip.id"
              :clip="clip"
              :track="track"
            />
          </div>
          <Playhead />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useProjectStore } from '@/stores/project'
import { useTimelineStore, ZOOM_MIN, ZOOM_MAX } from '@/stores/timeline'
import { syncClipToSprite, removeSpriteOfClip } from '@/adapters/webav'
import { appendMediaToCanvas } from '@/adapters/webav'
import TimeRuler from './timeline/TimeRuler.vue'
import TrackHeader from './timeline/TrackHeader.vue'
import ClipComponent from './timeline/ClipComponent.vue'
import Playhead from './timeline/Playhead.vue'

const project = useProjectStore()
const timeline = useTimelineStore()
const bodyEl = ref(null)

function onScroll() {
  if (!bodyEl.value) return
  timeline.setScroll(bodyEl.value.scrollLeft, bodyEl.value.scrollTop)
}

function zoomIn() { timeline.setZoom(timeline.pixelsPerSecond * 1.25) }
function zoomOut() { timeline.setZoom(timeline.pixelsPerSecond / 1.25) }
function zoomReset() { timeline.setZoom(60) }

function onDelete() {
  if (!project.selectedClipId) return
  removeSpriteOfClip(project.selectedClipId)
  project.removeClip(project.selectedClipId)
}

function onSplit() {
  if (!project.selectedClipId) return
  project.splitClipAt(project.selectedClipId, timeline.playheadPosition)
}

async function onDrop(e, trackId) {
  const mediaId = e.dataTransfer.getData('text/media-id')
  if (!mediaId) return
  const start = timeline.playheadPosition || 0
  await appendMediaToCanvas({ mediaId, trackId, startTime: start })
}

// 播放循环 - 简单的 rAF 推进播放头
let raf = null
let lastTs = 0
function tick(ts) {
  if (timeline.playbackState === 'playing') {
    if (lastTs) {
      const dt = (ts - lastTs) / 1000 * timeline.playbackRate
      const next = timeline.playheadPosition + dt
      const total = project.totalDuration || 5
      if (next >= total) {
        timeline.setPlayhead(0)
        timeline.pause()
      } else {
        timeline.setPlayhead(next)
      }
    }
    lastTs = ts
    raf = requestAnimationFrame(tick)
  } else {
    lastTs = 0
    raf = null
  }
}

onMounted(() => { watchPlayState() })
onBeforeUnmount(() => { if (raf) cancelAnimationFrame(raf) })

import { watch } from 'vue'
function watchPlayState() {
  watch(() => timeline.playbackState, (s) => {
    if (s === 'playing' && !raf) {
      lastTs = 0
      raf = requestAnimationFrame(tick)
    } else if (s !== 'playing' && raf) {
      cancelAnimationFrame(raf)
      raf = null
    }
  })
}

// 监听 clip 改动,同步到 sprite
import { watch as w2 } from 'vue'
w2(() => project.tracks.map(t => t.clips.map(c => [c.id, c.startTime, c.duration, c.opacity, c.volume])), () => {
  for (const t of project.tracks) {
    for (const c of t.clips) {
      syncClipToSprite(c.id, c)
    }
  }
}, { deep: true })
</script>

<style scoped>
.tl-root { display: flex; flex-direction: column; height: 100%; }
.tl-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px; font-size: 12px; color: #b8bcc7;
  border-bottom: 1px solid #232733;
}
.tl-tools { display: flex; align-items: center; gap: 8px; }
.tl-body { flex: 1; overflow: auto; position: relative; min-height: 0; }
.tl-ruler { position: sticky; top: 0; z-index: 4; background: #15181f; }
.tl-tracks { display: flex; }
.tl-lanes { flex: 1; position: relative; }
.tl-lane {
  position: relative;
  border-bottom: 1px solid #1c2030;
  background: #11141a;
}
</style>
