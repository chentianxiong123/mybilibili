<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import Hls from 'hls.js'

const props = defineProps({
  video: { type: Object, default: null }
})

const videoRef = ref(null)
const isPlaying = ref(false)
const showControls = ref(true)
const currentTime = ref(0)
const duration = ref(0)
const barrageSwitch = ref(true)
const isFullscreen = ref(false)
const showCover = ref(true)
let controlsTimer = null
let hls = null
const barrageList = ref([])
const barrageContainerRef = ref(null)

onMounted(() => {
  if (props.video?.url && videoRef.value) {
    initPlayer()
  }
})

onUnmounted(() => {
  if (hls) { hls.destroy(); hls = null }
})

const initPlayer = () => {
  const video = videoRef.value
  if (!video) return
  const url = props.video.url
  if (url.includes('.m3u8') && Hls.isSupported()) {
    hls = new Hls()
    hls.loadSource(url)
    hls.attachMedia(video)
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      showCover.value = false
    })
  } else {
    video.src = url
  }
  video.addEventListener('loadedmetadata', () => {
    duration.value = video.duration
  })
  video.addEventListener('timeupdate', () => {
    currentTime.value = video.currentTime
    renderBarrages()
  })
  video.addEventListener('play', () => { isPlaying.value = true; showCover.value = false })
  video.addEventListener('pause', () => { isPlaying.value = false })
}

const togglePlay = () => {
  const video = videoRef.value
  if (!video) return
  if (video.paused) {
    video.play()
  } else {
    video.pause()
  }
}

const toggleFullscreen = () => {
  const el = videoRef.value.parentElement
  if (!document.fullscreenElement) {
    el.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

const toggleBarrage = () => {
  barrageSwitch.value = !barrageSwitch.value
}

const renderBarrages = () => {
  // 简单弹幕渲染：通过 CSS animation
}

const sendBarrage = (text, color = '#fff') => {
  if (!barrageSwitch.value) return
  const container = barrageContainerRef.value
  if (!container) return
  const el = document.createElement('div')
  el.className = 'barrage-item'
  el.textContent = text
  el.style.color = color
  el.style.top = (Math.random() * 70 + 10) + '%'
  container.appendChild(el)
  setTimeout(() => el.remove(), 6000)
}

defineExpose({ sendBarrage })

function formatTime(seconds) {
  if (!seconds || isNaN(seconds)) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="video-player">
    <div class="player-wrapper">
      <video
        ref="videoRef"
        class="video-element"
        @click="togglePlay"
        playsinline
      />
      <div ref="barrageContainerRef" class="barrage-layer" :class="{ off: !barrageSwitch }" />

      <img v-if="showCover && video?.cover" :src="video.cover" class="video-cover" />

      <div class="play-btn" @click="togglePlay" v-if="!isPlaying">
        <svg viewBox="0 0 24 24" width="48" height="48">
          <path d="M8 5v14l11-7z" fill="#fff"/>
        </svg>
      </div>

      <div class="controls" v-show="isPlaying || true">
        <div class="control-bar">
          <span class="time">{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</span>
          <div class="control-btns">
            <button @click.stop="toggleBarrage">
              {{ barrageSwitch ? '弹' : '关' }}
            </button>
            <button @click.stop="toggleFullscreen">⛶</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.video-player {
  background: #000;
}

.player-wrapper {
  position: relative;
  width: 100%;
  padding-top: 56.25%;
}

.video-element {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.barrage-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;

  &.off { display: none; }
}

.video-cover {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.play-btn {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  cursor: pointer;
  opacity: 0.8;
}

.controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(transparent, rgba(0,0,0,0.7));
}

.control-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  color: #fff;
  font-size: 12px;
}

.control-btns {
  display: flex;
  gap: 12px;

  button {
    background: none;
    border: none;
    color: #fff;
    font-size: 14px;
    cursor: pointer;
  }
}

@keyframes barrage-fly {
  from { transform: translateX(100vw); }
  to { transform: translateX(-100%); }
}

:deep(.barrage-item) {
  position: absolute;
  white-space: nowrap;
  font-size: 16px;
  text-shadow: 1px 1px 0 #000;
  animation: barrage-fly 6s linear forwards;
  pointer-events: none;
}
</style>