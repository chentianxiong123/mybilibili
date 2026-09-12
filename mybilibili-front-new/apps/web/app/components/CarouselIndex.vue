<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'

interface BannerItem {
  id?: string | number
  img: string
  title?: string
  link?: string
}

const props = withDefaults(defineProps<{
  banners: BannerItem[]
  interval?: number
  showIndicators?: boolean
  showArrows?: boolean
  showTitle?: boolean
}>(), {
  interval: 3500,
  showIndicators: true,
  showArrows: true,
  showTitle: true
})

const emit = defineEmits<{
  (e: 'change', index: number): void
}>()

const current = ref(0)
const isTransition = ref(false)
const isPrev = ref(false)
const slideWidth = computed(() => 100 / Math.max(1, props.banners.length))

let timer: ReturnType<typeof setTimeout> | null = null

const start = () => {
  stop()
  if (props.banners.length <= 1) return
  timer = setTimeout(() => next(), props.interval)
}

const stop = () => {
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
}

const goTo = (i: number) => {
  if (i === current.value) return
  isTransition.value = true
  isPrev.value = i < current.value
  current.value = i
  emit('change', i)
  start()
}

const next = () => {
  const n = (current.value + 1) % props.banners.length
  isTransition.value = true
  isPrev.value = false
  current.value = n
  emit('change', n)
  start()
}

const prev = () => {
  const n = (current.value - 1 + props.banners.length) % props.banners.length
  isTransition.value = true
  isPrev.value = true
  current.value = n
  emit('change', n)
  start()
}

const onTransitionEnd = () => {
  isTransition.value = false
}

onMounted(() => {
  start()
})

onBeforeUnmount(() => {
  stop()
})

const transformStyle = computed(() => {
  const offset = isPrev.value
    ? 0
    : (current.value + 1) * slideWidth.value
  return {
    transform: `translateX(-${offset}%)`,
    transition: isTransition.value ? 'transform 400ms ease' : 'none',
    width: `${props.banners.length * 100}%`
  }
})
</script>

<template>
  <div class="v-carousel" @mouseenter="stop" @mouseleave="start">
    <div class="v-carousel-viewport">
      <div class="v-carousel-track" :style="transformStyle" @transitionend="onTransitionEnd">
        <div
          class="v-carousel-slide"
          :style="{ width: `${slideWidth}%` }"
          v-for="(b, i) in banners"
          :key="b.id ?? i"
        >
          <a class="v-carousel-inner" :href="b.link || '#'" target="_blank">
            <img loading="lazy" decoding="async" :src="b.img" :alt="b.title || `banner-${i}`">
          </a>
        </div>
      </div>
      <div class="v-carousel-shadow"></div>
    </div>

    <div v-if="showIndicators && banners.length > 1" class="v-carousel-indicators">
      <span
        v-for="(_, i) in banners"
        :key="i"
        :class="['v-carousel-dot', { 'is-active': i === current }]"
        @click="goTo(i)"
      ></span>
    </div>

    <div v-if="showArrows && banners.length > 1" class="v-carousel-arrows">
      <button type="button" class="v-carousel-arrow v-carousel-arrow-prev" @click="prev" aria-label="prev">
        <i class="iconfont icon-zuojiantou"></i>
      </button>
      <button type="button" class="v-carousel-arrow v-carousel-arrow-next" @click="next" aria-label="next">
        <i class="iconfont icon-youjiantou"></i>
      </button>
    </div>

    <div v-if="showTitle && banners[current]?.title" class="v-carousel-title">
      <span>{{ banners[current]?.title }}</span>
    </div>
  </div>
</template>

<style scoped>
.v-carousel {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: var(--v-radius-md, 8px);
  background: var(--v-bg2, #F6F7F8);
}

.v-carousel-viewport {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.v-carousel-track {
  display: flex;
  height: 100%;
}

.v-carousel-slide {
  position: relative;
  height: 100%;
  flex-shrink: 0;
}

.v-carousel-inner {
  display: block;
  width: 100%;
  height: 100%;
}

.v-carousel-inner img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.v-carousel-shadow {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 80px;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.7) 0%, rgba(0, 0, 0, 0.4) 50%, transparent 100%);
  pointer-events: none;
}

.v-carousel-indicators {
  position: absolute;
  bottom: 16px;
  left: 16px;
  display: flex;
  gap: 6px;
  z-index: 2;
}

.v-carousel-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: all 0.3s ease;
}

.v-carousel-dot:hover {
  background-color: rgba(255, 255, 255, 0.8);
}

.v-carousel-dot.is-active {
  background-color: var(--v-bg1, #FFFFFF);
  transform: scale(1.2);
}

.v-carousel-arrows {
  position: absolute;
  bottom: 16px;
  right: 16px;
  display: flex;
  gap: 8px;
  z-index: 2;
}

.v-carousel-arrow {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.15);
  border: none;
  color: var(--v-bg1, #FFFFFF);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.v-carousel-arrow:hover {
  background-color: rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}

.v-carousel-arrow:active {
  transform: scale(0.95);
}

.v-carousel-arrow .iconfont {
  font-size: 14px;
  font-weight: 600;
}

.v-carousel-title {
  position: absolute;
  top: 16px;
  left: 16px;
  right: 80px;
  color: var(--v-bg1, #FFFFFF);
  font-size: 16px;
  font-weight: 600;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  z-index: 2;
}
</style>
