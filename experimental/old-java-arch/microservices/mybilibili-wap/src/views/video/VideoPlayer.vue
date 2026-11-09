<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { getBarrages } from '../../api/video'

const props = defineProps({
  video: { type: Object, default: null },
  live: { type: Boolean, default: false },
  isLive: { type: Boolean, default: false },
  danmakuMount: { type: String, default: '' }
})

const emit = defineEmits<{
  (e: 'timeupdate', time: number): void
  (e: 'play'): void
  (e: 'pause'): void
  (e: 'ended'): void
}>()

const playerRef = ref<HTMLDivElement | null>(null)
let art: any = null
let artReady = false
let ArtplayerClass: any = null
let DanmukuPluginClass: any = null
let HlsClass: any = null

const isPlaying = ref(false)
const showCover = ref(true)
const finish = ref(false)
const duration = ref(0)
const currentTime = ref(0)

const buildQualityOptions = () => {
  const opts: any[] = []
  const v = props.video
  if (!v) return opts
  if (v.playUrlHd) opts.push({ default: true, name: '1080P 高清', html: '1080P 高清', url: v.playUrlHd })
  if (v.playUrlSd) opts.push({ default: !v.playUrlHd, name: '720P 标清', html: '720P 标清', url: v.playUrlSd })
  if (v.playUrlLd) opts.push({ default: !v.playUrlHd && !v.playUrlSd, name: '480P 流畅', html: '480P 流畅', url: v.playUrlLd })
  if (opts.length === 0 && v.playUrl) opts.push({ default: true, name: '默认', html: '默认', url: v.playUrl })
  return opts
}

const initPlayer = async () => {
  if (!playerRef.value || !props.video) return

  const defaultUrl = props.video.playUrl || props.video.playUrlHd || ''
  if (!defaultUrl) return

  const qualityOptions = buildQualityOptions()
  const isHlsVideo = defaultUrl.endsWith('.m3u8') || qualityOptions.some(option => option.url?.endsWith('.m3u8'))

  // Dynamic import player libraries only when the video detail page needs them.
  if (!ArtplayerClass) {
    const [Art, Danmuku] = await Promise.all([
      import('artplayer'),
      import('artplayer-plugin-danmuku')
    ])
    ArtplayerClass = Art.default
    DanmukuPluginClass = Danmuku.default
  }
  if (isHlsVideo && !HlsClass) {
    const Hls = await import('hls.js')
    HlsClass = Hls.default
  }

  // Load barrages
  let barrages: any[] = []
  if (!props.live && props.video.aId) {
    try {
      const res = await getBarrages(props.video.aId)
      if (res.code === '1') {
        barrages = (res.data || []).map((b: any) => ({
          text: b.text || b.content,
          time: parseFloat(b.time) || 0,
          color: b.color || '#ffffff'
        }))
      }
    } catch (e) {}
  }

  art = new ArtplayerClass({
    container: playerRef.value,
    url: defaultUrl,
    poster: props.video.pic || '',
    volume: 0.7,
    isLive: !!props.live,
    muted: false,
    autoplay: false,
    pip: false,
    autoSize: false,
    autoMini: false,
    screenshot: false,
    setting: true,
    loop: false,
    flip: false,
    playbackRate: true,
    aspectRatio: false,
    fullscreen: true,
    fullscreenWeb: false,
    miniProgressBar: false,
    quality: qualityOptions,
    theme: '#fb7299',
    lang: 'zh-cn',
    type: defaultUrl.endsWith('.m3u8') ? 'm3u8' : 'mp4',
    customType: {
      m3u8(video: HTMLVideoElement, url: string) {
        if (video.canPlayType('application/x-mpegURL')) {
          video.src = url
        } else if (HlsClass?.isSupported()) {
          const hls = new HlsClass()
          hls.loadSource(url)
          hls.attachMedia(video)
        }
      }
    },
    plugins: [
      DanmukuPluginClass({
        danmuku: barrages,
        speed: 5,
        opacity: 1,
        color: '#ffffff',
        fontSize: 20,
        synchronousPlayback: true,
        maxLength: 50,
        margin: [6, 6, 6, 6],
        theme: props.danmakuMount ? 'light' : 'dark',
        width: 0,
        lockTime: 3,
        ...(props.danmakuMount ? { mount: props.danmakuMount } : {})
      })
    ]
  })

  art.on('ready', () => {
    artReady = true
    duration.value = art!.duration
    showCover.value = false
  })

  art.on('play', () => {
    isPlaying.value = true
    finish.value = false
    emit('play')
  })

  art.on('pause', () => {
    isPlaying.value = false
    emit('pause')
  })

  art.on('ended', () => {
    isPlaying.value = false
    finish.value = true
    emit('ended')
  })

  art.on('timeupdate', () => {
    if (!art) return
    currentTime.value = art.currentTime
    duration.value = art.duration
    emit('timeupdate', art.currentTime)
  })
}

const destroyPlayer = () => {
  if (art) {
    art.destroy()
    art = null
    artReady = false
  }
}

watch(() => props.video, () => {
  destroyPlayer()
  if (props.video?.playUrl) {
    nextTick(() => initPlayer())
  }
})

onMounted(() => {
  if (props.video?.playUrl) {
    initPlayer()
  }
})

onUnmounted(() => {
  destroyPlayer()
})

defineExpose({
  play: () => art?.play(),
  pause: () => art?.pause(),
  seek: (time: number) => { if (art) art.currentTime = time },
  getArt: () => art
})
</script>

<template>
  <div class="artplayer-mobile-wrap">
    <div ref="playerRef" class="artplayer-mobile" />
  </div>
</template>

<style scoped lang="scss">
.artplayer-mobile-wrap {
  width: 100%;
  background: #000;
  position: relative;
}

.artplayer-mobile {
  width: 100%;
  aspect-ratio: 16/9;

  // Override Artplayer default styles for mobile
  :deep(.art-video) {
    object-fit: contain;
  }

  :deep(.art-controls) {
    .art-controls-left,
    .art-controls-right {
      .art-control {
        min-width: 32px;
        height: 36px;
      }
    }
  }

  // 当弹幕控制挂载到外部时，隐藏播放器内部的弹幕控制
  :deep(.apd-toggle-on),
  :deep(.apd-toggle-off),
  :deep(.apd-emitter) {
    display: none !important;
  }
}
</style>
