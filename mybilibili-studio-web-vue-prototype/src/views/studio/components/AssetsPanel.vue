<template>
  <div class="ap-root">
    <div class="ap-head">
      <span>素材库</span>
      <el-button size="small" @click="pickFile">导入</el-button>
      <input ref="fileInput" type="file" multiple accept="video/*,audio/*,image/*" hidden @change="onFiles" />
    </div>
    <div class="ap-body">
      <el-empty v-if="!project.mediaItems.length" description="尚未导入素材" :image-size="60" />
      <ul class="ap-list">
        <li
          v-for="m in project.mediaItems"
          :key="m.id"
          class="ap-item"
          :class="{ active: project.selectedMediaId === m.id }"
          draggable="true"
          @dragstart="onDragStart($event, m.id)"
          @click="project.selectMedia(m.id)"
        >
          <el-icon class="ap-icon"><component :is="iconOf(m.kind)" /></el-icon>
          <div class="ap-meta">
            <div class="ap-name">{{ m.name }}</div>
            <div class="ap-sub">{{ m.kind }} · {{ (m.duration || 0).toFixed(1) }}s</div>
          </div>
          <el-button size="small" link @click.stop="addToTimeline(m)">+</el-button>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { VideoCamera, Headset, Picture } from '@element-plus/icons-vue'
import { useProjectStore } from '@/stores/project'
import { appendMediaToCanvas } from '@/adapters/webav'

const project = useProjectStore()
const fileInput = ref(null)

function pickFile() { fileInput.value?.click() }

function iconOf(kind) {
  return { video: VideoCamera, audio: Headset, image: Picture }[kind] || VideoCamera
}

async function onFiles(e) {
  const files = Array.from(e.target.files || [])
  for (const f of files) {
    const kind = f.type.startsWith('video/') ? 'video'
      : f.type.startsWith('audio/') ? 'audio'
      : f.type.startsWith('image/') ? 'image' : 'video'
    const url = URL.createObjectURL(f)
    // 先创建 mediaItem 拿到 id,再异步读取时长
    const item = {
      name: f.name, kind, url, _file: f, source: 'local'
    }
    project.addMediaItem(item)
    // 探测时长
    if (kind === 'video' || kind === 'audio') {
      const m = project.mediaItems.find(x => x.url === url)
      if (m) probeDuration(m)
    } else if (kind === 'image') {
      const m = project.mediaItems.find(x => x.url === url)
      if (m) m.duration = 5
    }
  }
  e.target.value = ''
}

function probeDuration(m) {
  const el = document.createElement(m.kind === 'audio' ? 'audio' : 'video')
  el.preload = 'metadata'
  el.src = m.url
  el.onloadedmetadata = () => {
    m.duration = el.duration || 5
  }
}

function onDragStart(e, mediaId) {
  e.dataTransfer.setData('text/media-id', mediaId)
}

async function addToTimeline(m) {
  // 选第一条匹配的轨道
  const trackId = m.kind === 'audio' ? 'a1' : 'v1'
  const projectDur = project.totalDuration
  const start = projectDur
  await appendMediaToCanvas({ mediaId: m.id, trackId, startTime: start })
}
</script>

<style scoped>
.ap-root { display: flex; flex-direction: column; height: 100%; }
.ap-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 12px; font-size: 13px; color: #b8bcc7;
  border-bottom: 1px solid #232733;
}
.ap-body { flex: 1; overflow: auto; padding: 8px; }
.ap-list { list-style: none; padding: 0; margin: 0; }
.ap-item {
  display: flex; align-items: center; gap: 8px; padding: 6px 8px;
  border: 1px solid transparent; border-radius: 4px; cursor: grab;
}
.ap-item:hover { background: #1c2030; border-color: #2a3145; }
.ap-item.active { background: #2a3a5c; border-color: #3b5fa3; }
.ap-icon { font-size: 18px; color: #6fa8ff; }
.ap-name { font-size: 12px; color: #e6e8eb; }
.ap-sub { font-size: 11px; color: #8a8f9b; }
.ap-meta { display: flex; flex-direction: column; flex: 1; }
</style>


