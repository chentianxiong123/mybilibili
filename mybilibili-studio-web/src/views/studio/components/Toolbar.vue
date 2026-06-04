<template>
  <header class="toolbar">
    <div class="tb-left">
      <span class="logo">mybilibili Studio</span>
      <input
        class="proj-name"
        :value="project.project.name"
        @change="project.setProjectName($event.target.value)"
        size="small"
      />
      <span class="dim">·</span>
      <span class="dim">{{ project.project.width }}×{{ project.project.height }} · {{ project.project.fps }}fps</span>
    </div>

    <div class="tb-center">
      <el-button-group>
        <el-button size="small" :icon="VideoPlay" @click="timeline.togglePlay" :type="isPlaying ? 'primary' : 'default'">
          {{ isPlaying ? '暂停' : '播放' }}
        </el-button>
        <el-button size="small" :icon="RefreshLeft" @click="onRewind">回到开头</el-button>
      </el-button-group>
      <span class="time">{{ formatTime(timeline.playheadPosition) }} / {{ formatTime(totalDuration) }}</span>
    </div>

    <div class="tb-right">
      <router-link to="/workflow" custom v-slot="{ navigate }">
        <el-button size="small" @click="navigate">AI 工作流</el-button>
      </router-link>
      <el-button size="small" :icon="Document" @click="onExport" :loading="engine.isRendering">
        {{ engine.isRendering ? `导出 ${engine.renderProgress}%` : '导出 MP4' }}
      </el-button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { VideoPlay, RefreshLeft, Document } from '@element-plus/icons-vue'
import { useProjectStore } from '@/stores/project'
import { useTimelineStore } from '@/stores/timeline'
import { useEngineStore } from '@/stores/engine'
import { exportToMp4, downloadLastRender } from '@/adapters/webav'

const project = useProjectStore()
const timeline = useTimelineStore()
const engine = useEngineStore()

const isPlaying = computed(() => timeline.playbackState === 'playing')
const totalDuration = computed(() => project.totalDuration || 5)

function formatTime(t) {
  if (!Number.isFinite(t)) return '00:00.00'
  const m = Math.floor(t / 60)
  const s = (t - m * 60).toFixed(2).padStart(5, '0')
  return `${String(m).padStart(2, '0')}:${s}`
}

function onRewind() { timeline.setPlayhead(0) }

async function onExport() {
  try {
    const { url } = await exportToMp4()
    downloadLastRender()
    console.log('[export] done', url)
  } catch (e) {
    console.error('[export] failed', e)
  }
}
</script>

<style scoped>
.toolbar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  padding: 8px 12px;
  background: #15181f;
  border-bottom: 1px solid #232733;
  gap: 12px;
}
.tb-left, .tb-right { display: flex; align-items: center; gap: 8px; }
.tb-right { justify-content: flex-end; }
.tb-center { display: flex; align-items: center; gap: 12px; justify-content: center; }
.logo { font-weight: 700; color: #e6e8eb; }
.proj-name {
  background: #1c2030; border: 1px solid #2a3145; color: #e6e8eb;
  border-radius: 4px; padding: 4px 8px; width: 200px; font-size: 12px;
}
.dim { color: #8a8f9b; font-size: 12px; }
.time { color: #b8bcc7; font-family: ui-monospace, monospace; font-size: 12px; min-width: 130px; text-align: center; }
</style>
