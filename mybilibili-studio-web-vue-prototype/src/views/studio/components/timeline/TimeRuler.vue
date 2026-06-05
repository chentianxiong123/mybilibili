<template>
  <div class="ruler" :style="{ width: totalWidth + 'px' }">
    <div
      v-for="tick in ticks"
      :key="tick.t"
      class="tick"
      :class="{ major: tick.major }"
      :style="{ left: (tick.t * timeline.pixelsPerSecond) + 'px' }"
    >
      <span v-if="tick.major" class="label">{{ formatTime(tick.t) }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useTimelineStore } from '@/stores/timeline'
import { useProjectStore } from '@/stores/project'

const timeline = useTimelineStore()
const project = useProjectStore()

const totalWidth = computed(() => {
  const dur = (project.totalDuration || 30) + 10
  return Math.max(2000, dur * timeline.pixelsPerSecond)
})

const ticks = computed(() => {
  const dur = (project.totalDuration || 30) + 10
  const pps = timeline.pixelsPerSecond
  const majorStep = Math.max(1, Math.ceil(80 / pps))
  const minorStep = Math.max(0.1, majorStep / 5)
  const out = []
  for (let t = 0; t <= dur; t += minorStep) {
    const major = Math.abs(t - Math.round(t / majorStep) * majorStep) < 0.01
    out.push({ t: Math.round(t * 100) / 100, major })
  }
  return out
})

function formatTime(t) {
  const m = Math.floor(t / 60)
  const s = (t - m * 60).toFixed(0)
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
</script>

<style scoped>
.ruler {
  position: relative;
  height: 24px;
  background: #1a1d24;
  border-bottom: 1px solid #232733;
  min-width: 100%;
}
.tick {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 1px;
  background: #2a2f3a;
}
.tick.major { background: #4a4f5a; }
.tick.major .label {
  position: absolute;
  top: 2px;
  left: 4px;
  font-size: 10px;
  color: #8a8f9b;
  font-family: ui-monospace, monospace;
  white-space: nowrap;
}
</style>

