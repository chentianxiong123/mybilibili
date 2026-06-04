<template>
  <div
    class="clip"
    :class="[track.kind, { active: project.selectedClipId === clip.id }]"
    :style="clipStyle"
    @mousedown.stop="onMouseDown"
    @click.stop="project.selectClip(clip.id)"
  >
    <span class="clip-label">{{ clip.label }}</span>
    <span
      v-if="track.kind !== 'audio'"
      class="resize-handle"
      @mousedown.stop="onResizeStart"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useProjectStore } from '@/stores/project'
import { useTimelineStore } from '@/stores/timeline'
import { syncClipToSprite } from '@/adapters/webav'

const props = defineProps({
  clip: { type: Object, required: true },
  track: { type: Object, required: true }
})

const project = useProjectStore()
const timeline = useTimelineStore()

const clipStyle = computed(() => ({
  left: (props.clip.startTime * timeline.pixelsPerSecond) + 'px',
  width: (props.clip.duration * timeline.pixelsPerSecond) + 'px',
  opacity: props.clip.opacity ?? 1
}))

let drag = null
function onMouseDown(e) {
  project.selectClip(props.clip.id)
  drag = { startX: e.clientX, baseStart: props.clip.startTime }
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}
function onMove(e) {
  if (!drag) return
  const dx = e.clientX - drag.startX
  const newStart = Math.max(0, drag.baseStart + dx / timeline.pixelsPerSecond)
  project.updateClip(props.clip.id, { startTime: newStart })
  syncClipToSprite(props.clip.id, { startTime: newStart })
}
function onUp() {
  drag = null
  window.removeEventListener('mousemove', onMove)
  window.removeEventListener('mouseup', onUp)
}

let resize = null
function onResizeStart(e) {
  resize = { startX: e.clientX, baseDuration: props.clip.duration }
  window.addEventListener('mousemove', onRMove)
  window.addEventListener('mouseup', onRUp)
}
function onRMove(e) {
  if (!resize) return
  const dx = e.clientX - resize.startX
  const newDur = Math.max(0.2, resize.baseDuration + dx / timeline.pixelsPerSecond)
  project.updateClip(props.clip.id, { duration: newDur })
  syncClipToSprite(props.clip.id, { duration: newDur })
}
function onRUp() {
  resize = null
  window.removeEventListener('mousemove', onRMove)
  window.removeEventListener('mouseup', onRUp)
}
</script>

<style scoped>
.clip {
  position: absolute; top: 2px; bottom: 2px;
  border-radius: 3px;
  font-size: 11px;
  color: #d6e4ff;
  padding: 2px 6px;
  cursor: grab;
  user-select: none;
  display: flex; align-items: center; justify-content: space-between;
  overflow: hidden;
  border: 1px solid transparent;
  box-sizing: border-box;
}
.clip.video { background: #2a3a5c; border-color: #3b5fa3; }
.clip.audio { background: #2a4d3a; border-color: #3b8a5a; color: #c8e8d4; }
.clip.subtitle { background: #4d3a2a; border-color: #8a6a3b; color: #e8d4c8; }
.clip.active { border-color: #6fa8ff; box-shadow: 0 0 0 1px #6fa8ff inset; }
.clip-label { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.resize-handle {
  width: 6px; height: 100%; cursor: ew-resize;
  background: rgba(255,255,255,0.1); flex-shrink: 0;
}
</style>

