<template>
  <div class="ph" :style="{ left: (timeline.playheadPosition * timeline.pixelsPerSecond) + 'px' }">
    <div class="ph-line" />
    <div class="ph-head" @mousedown.stop="onDragStart" />
  </div>
</template>

<script setup>
import { useTimelineStore } from '@/stores/timeline'

const timeline = useTimelineStore()

let drag = null
function onDragStart(e) {
  drag = { startX: e.clientX, baseT: timeline.playheadPosition }
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}
function onMove(e) {
  if (!drag) return
  const dx = e.clientX - drag.startX
  timeline.setPlayhead(Math.max(0, drag.baseT + dx / timeline.pixelsPerSecond))
}
function onUp() {
  drag = null
  window.removeEventListener('mousemove', onMove)
  window.removeEventListener('mouseup', onUp)
}
</script>

<style scoped>
.ph {
  position: absolute; top: 0; bottom: 0; z-index: 5;
  pointer-events: none;
  width: 0;
}
.ph-line {
  position: absolute; top: 0; bottom: 0; left: 0;
  width: 1px; background: #ff5566;
  box-shadow: 0 0 4px rgba(255,85,102,0.5);
}
.ph-head {
  position: absolute; top: -2px; left: -6px;
  width: 12px; height: 12px;
  background: #ff5566; border-radius: 2px;
  pointer-events: auto; cursor: ew-resize;
}
</style>


