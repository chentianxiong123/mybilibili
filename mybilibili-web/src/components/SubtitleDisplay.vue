<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'

const props = defineProps({
  subtitles: {
    type: Array,
    default: () => []
  },
  currentTime: {
    type: Number,
    default: 0
  },
  enabled: {
    type: Boolean,
    default: true
  }
})

// 字幕设置
const subtitleSettings = ref({
  fontSize: 32,
  color: '#ffffff',
  backgroundColor: 'rgba(0, 0, 0, 0.75)',
  textShadow: '1px 1px 2px rgba(0, 0, 0, 0.5)',
  borderRadius: 4,
  padding: '8px 16px',
  lineHeight: 1.5
})

const BOUNDS_INSET = 12

// 拖拽相关
const subtitleRef = ref(null)
const offsetX = ref(0)
const offsetY = ref(0)
const isDragging = ref(false)
const startX = ref(0)
const startY = ref(0)
const initialOffsetX = ref(0)
const initialOffsetY = ref(0)
let resizeObserver = null
let observedBoundsEl = null
let observedSubtitleEl = null

// 当前显示的字幕文本
const currentSubtitleText = computed(() => {
  if (!props.enabled || !props.subtitles || props.subtitles.length === 0) {
    return ''
  }

  const current = props.subtitles.find(item =>
    props.currentTime >= item.startTime && props.currentTime <= item.endTime
  )

  return current ? current.text : ''
})

// 字幕容器样式
const containerStyle = computed(() => ({
  transform: `translate(calc(-50% + ${offsetX.value}px), ${offsetY.value}px)`
}))

// 字幕文本样式
const textStyle = computed(() => ({
  fontSize: `${subtitleSettings.value.fontSize}px`,
  color: subtitleSettings.value.color,
  backgroundColor: subtitleSettings.value.backgroundColor,
  textShadow: subtitleSettings.value.textShadow,
  borderRadius: `${subtitleSettings.value.borderRadius}px`,
  padding: subtitleSettings.value.padding,
  lineHeight: subtitleSettings.value.lineHeight
}))

const getBoundsElement = () => subtitleRef.value?.parentElement || null

const clampOffsetToBounds = (desiredX, desiredY) => {
  const el = subtitleRef.value
  const boundsEl = getBoundsElement()

  if (!el || !boundsEl || !el.offsetParent) {
    return { x: desiredX, y: desiredY }
  }

  const rect = el.getBoundingClientRect()
  const bounds = boundsEl.getBoundingClientRect()

  if (!rect.width || !rect.height || !bounds.width || !bounds.height) {
    return { x: desiredX, y: desiredY }
  }

  const deltaX = desiredX - offsetX.value
  const deltaY = desiredY - offsetY.value

  let nextX = desiredX
  let nextY = desiredY
  const nextRect = {
    left: rect.left + deltaX,
    right: rect.right + deltaX,
    top: rect.top + deltaY,
    bottom: rect.bottom + deltaY
  }

  const minLeft = bounds.left + BOUNDS_INSET
  const maxRight = bounds.right - BOUNDS_INSET
  const minTop = bounds.top + BOUNDS_INSET
  const maxBottom = bounds.bottom - BOUNDS_INSET

  if (nextRect.left < minLeft) {
    nextX += minLeft - nextRect.left
  }
  if (nextRect.right > maxRight) {
    nextX -= nextRect.right - maxRight
  }
  if (nextRect.top < minTop) {
    nextY += minTop - nextRect.top
  }
  if (nextRect.bottom > maxBottom) {
    nextY -= nextRect.bottom - maxBottom
  }

  return { x: nextX, y: nextY }
}

const clampCurrentPosition = () => {
  const { x, y } = clampOffsetToBounds(offsetX.value, offsetY.value)
  offsetX.value = x
  offsetY.value = y
}

// 开始拖拽
const startDrag = (e) => {
  e.stopPropagation()
  isDragging.value = true
  const clientX = e.type.includes('touch') ? e.touches[0].clientX : e.clientX
  const clientY = e.type.includes('touch') ? e.touches[0].clientY : e.clientY
  startX.value = clientX
  startY.value = clientY
  initialOffsetX.value = offsetX.value
  initialOffsetY.value = offsetY.value
}

// 拖拽中
const onDrag = (e) => {
  if (!isDragging.value) return
  e.preventDefault()
  const clientX = e.type.includes('touch') ? e.touches[0].clientX : e.clientX
  const clientY = e.type.includes('touch') ? e.touches[0].clientY : e.clientY

  const nextX = initialOffsetX.value + (clientX - startX.value)
  const nextY = initialOffsetY.value + (clientY - startY.value)
  const clamped = clampOffsetToBounds(nextX, nextY)

  offsetX.value = clamped.x
  offsetY.value = clamped.y
}

// 结束拖拽
const endDrag = () => {
  isDragging.value = false
}

// 更新设置
const updateSettings = (newSettings) => {
  subtitleSettings.value = { ...subtitleSettings.value, ...newSettings }
  nextTick(clampCurrentPosition)
}

// 重置位置
const resetPosition = () => {
  offsetX.value = 0
  offsetY.value = 0
  nextTick(clampCurrentPosition)
}

// 居中字幕
const centerSubtitle = () => {
  offsetX.value = 0
  offsetY.value = 0
  nextTick(clampCurrentPosition)
}

// 更新位置（兼容旧接口）
const updatePosition = () => {
  nextTick(clampCurrentPosition)
}

const attachResizeObserver = () => {
  if (typeof ResizeObserver === 'undefined') return

  if (!resizeObserver) {
    resizeObserver = new ResizeObserver(() => {
      clampCurrentPosition()
    })
  }

  const boundsEl = getBoundsElement()
  if (boundsEl && boundsEl !== observedBoundsEl) {
    if (observedBoundsEl) resizeObserver.unobserve(observedBoundsEl)
    resizeObserver.observe(boundsEl)
    observedBoundsEl = boundsEl
  }

  if (subtitleRef.value && subtitleRef.value !== observedSubtitleEl) {
    if (observedSubtitleEl) resizeObserver.unobserve(observedSubtitleEl)
    resizeObserver.observe(subtitleRef.value)
    observedSubtitleEl = subtitleRef.value
  }
}

onMounted(() => {
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', endDrag)
  document.addEventListener('touchmove', onDrag, { passive: false })
  document.addEventListener('touchend', endDrag)

  nextTick(() => {
    attachResizeObserver()
    clampCurrentPosition()
  })
})

onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
  document.removeEventListener('touchmove', onDrag)
  document.removeEventListener('touchend', endDrag)
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
    observedBoundsEl = null
    observedSubtitleEl = null
  }
})

watch(currentSubtitleText, () => {
  nextTick(() => {
    attachResizeObserver()
    clampCurrentPosition()
  })
})

// 暴露方法给父组件
defineExpose({
  updateSettings,
  resetPosition,
  getSettings: () => subtitleSettings.value,
  centerSubtitle,
  updatePosition
})
</script>

<template>
  <div
    v-show="enabled && currentSubtitleText"
    ref="subtitleRef"
    class="subtitle-display"
    :class="{ 'is-dragging': isDragging }"
    :style="containerStyle"
  >
    <div
      class="subtitle-text"
      :style="textStyle"
      @mousedown="startDrag"
      @touchstart="startDrag"
      v-html="currentSubtitleText.replace(/\\n/g, '<br>')"
    ></div>
    <div class="drag-hint" v-if="isDragging">拖拽中</div>
  </div>
</template>

<style scoped>
.subtitle-display {
  position: absolute;
  left: 50%;
  bottom: 64px;
  display: block;
  pointer-events: auto;
  text-align: center;
  user-select: none;
  width: max-content;
  max-width: calc(100% - 32px);
  z-index: 2;
  will-change: transform;
}

.subtitle-text {
  display: block;
  cursor: move;
  transition: opacity 0.2s ease;
  max-width: 100%;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
  text-align: center;
}

.subtitle-text:hover {
  box-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
}

.subtitle-text:active {
  cursor: grabbing;
}

.drag-hint {
  position: absolute;
  top: -25px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
}
</style>
