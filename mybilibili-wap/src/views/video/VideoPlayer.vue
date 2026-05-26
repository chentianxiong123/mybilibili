<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import Hls from 'hls.js'
import { getBarrages } from '../../api/video'

const props = defineProps({
  video: { type: Object, default: null },
  live: { type: Boolean, default: false },
  isLive: { type: Boolean, default: false }
})

const videoRef = ref<HTMLVideoElement | null>(null)
const barrageRef = ref<HTMLDivElement | null>(null)
const currentTimeRef = ref<HTMLSpanElement | null>(null)
const progressRef = ref<HTMLDivElement | null>(null)

const isPlaying = ref(false)
const showCover = ref(true)
const duration = ref(0)
const barrageSwitch = ref(true)
const finish = ref(false)
const waiting = ref(false)
const fullscreen = ref(false)

let hls: Hls | null = null
let initBarrages: any[] = []
let barrages: any[] = []
let viewWidth = 0
let viewHeight = 0
let contentHeight = 0
let randomTop = 0
let fixedTop = 0

// 弹幕类型
const BarrageType = { RANDOM: 1, FIXED: 2 }

onMounted(() => {
  if (props.video?.playUrl) {
    initVideo()
  }
  if (!props.live && props.video?.aId) {
    loadBarrages()
  }
})

onUnmounted(() => {
  if (hls) { hls.destroy(); hls = null }
})

const loadBarrages = async () => {
  try {
    const res = await getBarrages(props.video.aId)
    if (res.code === '1') {
      initBarrages = (res.data || []).map(b => ({
        type: BarrageType.RANDOM,
        color: b.color || '#fff',
        content: b.text
      }))
      barrages = initBarrages.slice()
    }
  } catch (e) {}
}

const initVideo = () => {
  const video = videoRef.value
  if (!video) return
  const url = props.video.playUrl
  if (!url) return

  if (url.includes('.m3u8') && Hls.isSupported()) {
    hls = new Hls()
    hls.loadSource(url)
    hls.attachMedia(video)
  } else {
    video.src = url
  }

  video.addEventListener('play', onPlay)
  video.addEventListener('playing', onPlay)
  video.addEventListener('waiting', () => { waiting.value = true })
  video.addEventListener('timeupdate', onTimeUpdate)
  video.addEventListener('ended', onEnded)
}

const onPlay = () => {
  showCover.value = false
  isPlaying.value = true
  waiting.value = false
  if (barrageRef.value && viewWidth === 0) {
    refreshBarrageSize()
  }
}

const onTimeUpdate = () => {
  const video = videoRef.value
  if (!video) return
  if (duration.value === 0) {
    duration.value = video.duration
  }
  if (currentTimeRef.value) {
    currentTimeRef.value.textContent = formatTime(video.currentTime)
  }
  if (progressRef.value) {
    const progress = video.currentTime / video.duration * 100
    progressRef.value.style.width = `${progress}%`
  }
  // 发送时间匹配的弹幕
  if (barrageSwitch.value && barrages.length > 0) {
    const matched = findBarrages(video.currentTime)
    matched.forEach(b => sendBarrage(b))
  }
}

const onEnded = () => {
  isPlaying.value = false
  finish.value = true
  barrages = initBarrages.slice()
  clearBarrage()
}

// ====== 弹幕核心（翻译自 react-bilibili Barrage.tsx） ======

const refreshBarrageSize = () => {
  if (!barrageRef.value) return
  viewWidth = barrageRef.value.offsetWidth
  viewHeight = barrageRef.value.offsetHeight
  // 测量弹幕内容高度
  const div = document.createElement('div')
  div.innerHTML = '测'
  div.style.fontSize = '0.8rem'
  document.body.appendChild(div)
  contentHeight = div.offsetHeight
  document.body.removeChild(div)
}

const sendBarrage = (barrage: { type: number; color: string; content: string }) => {
  const container = barrageRef.value
  if (!container) return

  const el = document.createElement('div')
  el.innerHTML = barrage.content
  el.style.cssText = `
    position: absolute;
    font-family: 黑体;
    font-size: 0.8rem;
    font-weight: bold;
    white-space: pre;
    text-shadow: rgb(0, 0, 0) 1px 1px 2px;
    color: ${barrage.color};
    opacity: 1;
  `

  if (barrage.type !== BarrageType.FIXED) {
    // 滚动弹幕
    el.style.top = `${randomTop}px`
    el.style.left = `${viewWidth}px`
    el.style.transition = 'transform 5s linear'
    container.appendChild(el)

    const targetX = -(viewWidth + el.offsetWidth)
    setTimeout(() => {
      el.style.transform = `translate3d(${targetX}px, 0, 0)`
    }, 10)

    const onEnd = () => {
      el.removeEventListener('transitionend', onEnd)
      if (el.parentNode === container) container.removeChild(el)
      randomTop -= contentHeight
      if (randomTop < 0) randomTop = 0
    }
    el.addEventListener('transitionend', onEnd)

    randomTop += contentHeight
    if (randomTop > viewHeight - contentHeight) randomTop = 0
  } else {
    // 固定弹幕
    el.style.top = `${fixedTop}px`
    container.appendChild(el)
    el.style.left = (viewWidth - el.offsetWidth) / 2 + 'px'
    setTimeout(() => {
      if (el.parentNode === container) container.removeChild(el)
      fixedTop -= contentHeight
      if (fixedTop < 0) fixedTop = 0
    }, 5000)
    fixedTop += contentHeight
    if (fixedTop > viewHeight - contentHeight) fixedTop = 0
  }
}

const clearBarrage = () => {
  randomTop = 0
  fixedTop = 0
  const container = barrageRef.value
  if (container) container.innerHTML = ''
}

// 查找当前时间点应该显示的弹幕
const findBarrages = (currentTime: number) => {
  // 简化：每次 timeupdate 取 1-2 条
  if (barrages.length === 0) return []
  const result = barrages.splice(0, 1)
  return result
}

// ====== 播放控制 ======

const togglePlay = () => {
  const video = videoRef.value
  if (!video) return
  if (video.paused) video.play()
  else video.pause()
}

const toggleFullscreen = () => {
  const el = videoRef.value?.parentElement
  if (!el) return
  if (!document.fullscreenElement) {
    el.requestFullscreen()
    fullscreen.value = true
  } else {
    document.exitFullscreen()
    fullscreen.value = false
  }
}

const toggleBarrage = () => {
  barrageSwitch.value = !barrageSwitch.value
  if (!barrageSwitch.value) clearBarrage()
}

const seekTo = (e: MouseEvent) => {
  const bar = e.currentTarget as HTMLElement
  const rect = bar.getBoundingClientRect()
  const ratio = (e.clientX - rect.left) / rect.width
  const video = videoRef.value
  if (video && video.duration) {
    video.currentTime = ratio * video.duration
  }
}

const formatTime = (seconds: number) => {
  if (!seconds || isNaN(seconds)) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

// 供外部调用发送弹幕
const sendExternalBarrage = (text: string, color = '#fff') => {
  if (!barrageSwitch.value) return
  sendBarrage({ type: BarrageType.RANDOM, color, content: text })
}

defineExpose({ sendBarrage: sendExternalBarrage })
</script>

<template>
  <div class="video-player">
    <div class="player-wrapper">
      <video
        ref="videoRef"
        class="video-element"
        @click="togglePlay"
        playsinline
        webkit-playsinline
      />
      <!-- 弹幕层 -->
      <div ref="barrageRef" class="barrage-layer" :class="{ off: !barrageSwitch }" />

      <!-- 封面 -->
      <img v-if="showCover && video?.pic" :src="video.pic" class="video-cover" />

      <!-- 缓冲 -->
      <div v-if="waiting && !showCover" class="loading-overlay">
        <img src="../../assets/loading.svg" class="loading-icon" />
      </div>

      <!-- 播放按钮 -->
      <div v-if="!isPlaying && !finish" class="play-btn" @click="togglePlay">
        <img src="../../assets/ico-play.png" />
      </div>

      <!-- 重播 -->
      <div v-if="finish" class="replay-overlay" @click="togglePlay">
        <img src="../../assets/ico-replay.png" />
        <span>重播</span>
      </div>

      <!-- 控制栏 -->
      <div class="controls">
        <div class="progress-bar" @click="seekTo">
          <div ref="progressRef" class="progress-fill" />
        </div>
        <div class="control-bar">
          <span ref="currentTimeRef" class="time">00:00</span>
          <span class="duration">/ {{ formatTime(duration) }}</span>
          <div class="control-btns">
            <span class="barrage-btn" :class="{ off: !barrageSwitch }" @click="toggleBarrage">
              <img :src="barrageSwitch ? '../../assets/barrage-on.png' : '../../assets/barrage-off.png'" />
            </span>
            <span class="fullscreen-btn" @click="toggleFullscreen">
              <img src="../../assets/fullscreen.png" />
            </span>
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

.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.loading-icon {
  width: 36px;
  height: 36px;
}

.play-btn, .replay-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  cursor: pointer;

  img { width: 48px; height: 48px; }
}

.replay-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 14px;
}

.controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(transparent, rgba(0,0,0,0.7));
}

.progress-bar {
  height: 3px;
  background: rgba(255,255,255,0.3);
  cursor: pointer;
}

.progress-fill {
  height: 100%;
  background: #fb7299;
  width: 0;
}

.control-bar {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  color: #fff;
  font-size: 12px;
}

.duration { color: rgba(255,255,255,0.6); margin-left: 4px; }

.control-btns {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;

  img { width: 20px; height: 20px; }
}

.barrage-btn.off { opacity: 0.5; }
</style>