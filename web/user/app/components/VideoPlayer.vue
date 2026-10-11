<script setup lang="ts">
import { hasAuthSession } from "@/utils/auth.ts"
import { safeStorage } from '@/utils/safeStorage'
import Artplayer from 'artplayer'
import { ElMessage } from 'element-plus'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import { subtitleApi } from '@/api/subtitle.ts'
import SubtitleDisplay from '@/components/SubtitleDisplay.vue'
import SubtitleMenu from '@/components/SubtitleMenu.vue'
import { interactionApi } from '@/api/client'
import { ChatDotRound, CircleClose, View } from '@element-plus/icons-vue'

const props = defineProps({
  currentManuscriptId: { type: Number, required: true },
  manuscriptInfo: { type: Object, required: true },
  videoInfo: { type: Object, required: true },
  currentP: { type: Number, required: true },
  currentVideoIndex: { type: Number, default: 0 },
  resumeTime: { type: Number, default: 0 },
  danmuList: { type: Array, default: () => [] },
  loadingDanmus: { type: Boolean, default: false }
})

const emit = defineEmits<{
  'update:videoInfo': [value: any]
  'update:danmuList': [value: any[]]
  'update:loadingDanmus': [value: boolean]
  'timeUpdate': [value: { currentTime: number; duration: number }]
}>()

const videoId = computed(() => {
  const videos = props.manuscriptInfo?.videos || []
  const video = videos[props.currentVideoIndex]
  return video?.id || null
})

const playerRef = ref<HTMLDivElement | null>(null)
let art: any = null
const currentVideoTime = ref(0)
const pendingResumeTime = ref(props.resumeTime)
const hasAppliedResume = ref(false)

const applyResumeTime = (time: number) => {
  if (!art || !time || time <= 0 || hasAppliedResume.value) return
  const duration =
    Number(art.duration) ||
    Number(props.videoInfo.duration?.split(':').reduce((acc: number, cur: string) => acc * 60 + Number(cur), 0)) ||
    0
  const safeTime = duration > 0 ? Math.min(time, Math.max(0, duration - 1)) : time
  art.currentTime = safeTime
  currentVideoTime.value = safeTime
  hasAppliedResume.value = true
  pendingResumeTime.value = 0
}

const localDanmuList = ref<any[]>([])

watch(() => props.danmuList, (val) => {
  localDanmuList.value = val || []
}, { immediate: true })

const localLoadingDanmus = ref(false)

watch(() => props.loadingDanmus, (val) => {
  localLoadingDanmus.value = val
}, { immediate: true })

const emitDanmuList = () => {
  emit('update:danmuList', localDanmuList.value)
}

const emitLoadingDanmus = () => {
  emit('update:loadingDanmus', localLoadingDanmus.value)
}

const isNativeFullscreen = ref(false)
const isWebFullscreen = ref(false)
const isFullscreenMode = ref(false)

const updateFullscreenMode = () => {
  isFullscreenMode.value = isNativeFullscreen.value || isWebFullscreen.value
}

const subtitleList = ref<any[]>([])
const currentSubtitle = ref<any>(null)
const currentSubtitleContent = ref<any[]>([])
const currentLanguage = ref('')
const SUBTITLE_ENABLED_KEY = 'mybilibili_subtitle_enabled'
const subtitleEnabled = ref(true)
const subtitleDisplayRef = ref<any>(null)
const subtitleSettingsPanelRef = ref<HTMLDivElement | null>(null)
const subtitleSettingsVisible = ref(false)
const subtitleMenuAnchorX = ref(0)
let subtitleToggleGuard = false

const onDocClickForSubtitlePanel = (e: Event) => {
  if (subtitleToggleGuard) return
  // 点击菜单内部不关闭；点控制栏字幕按钮由 guard 处理
  const target = e.target as HTMLElement | null
  if (target?.closest?.('.bsub-menu') || target?.closest?.('.art-controls')) return
  subtitleSettingsVisible.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocClickForSubtitlePanel)
})
const SUBTITLE_SETTINGS_KEY = 'mybilibili_subtitle_settings'
const FULLSCREEN_FONT_SCALE = 1.5
const defaultSubtitleSettings = {
  fontSize: 32,
  fontScale: 1,
  color: '#ffffff',
  stroke: 0,
  position: 'bottom-center',
  backgroundOpacity: 0.75,
  fitScale: true,
  backgroundColor: 'rgba(0, 0, 0, 0.75)',
  textShadow: '1px 1px 2px rgba(0, 0, 0, 0.5)',
  borderRadius: 4,
  padding: '8px 16px',
  lineHeight: 1.5
}

const subtitleSettings = ref({ ...defaultSubtitleSettings })

if (import.meta.client) {
  subtitleEnabled.value = safeStorage.getItem(SUBTITLE_ENABLED_KEY) !== 'false'
  const saved = safeStorage.getItem(SUBTITLE_SETTINGS_KEY)
  if (saved) {
    try {
      const parsed = JSON.parse(saved)
      // 迁移旧格式：fontSize 像素值 → fontScale 倍率
      if (parsed.fontScale === undefined && typeof parsed.fontSize === 'number') {
        parsed.fontScale = Math.min(1.6, Math.max(0.6, parsed.fontSize / 32))
      }
      subtitleSettings.value = { ...defaultSubtitleSettings, ...parsed }
    } catch {}
  }
}

const saveSubtitleSettings = () => {
  try {
    safeStorage.setItem(SUBTITLE_SETTINGS_KEY, JSON.stringify(subtitleSettings.value))
  } catch (e) {
  }
}

const pushSettingsToSubtitle = () => {
  if (!subtitleDisplayRef.value) return
  // 等比缩放：字体随播放器宽度缩放（1280 为基准），全屏再乘系数
  let widthScale = 1
  if (subtitleSettings.value.fitScale && playerRef.value) {
    const w = playerRef.value.clientWidth || 1280
    widthScale = Math.min(2, Math.max(0.6, w / 1280))
  }
  const fullscreenScale = isFullscreenMode.value ? FULLSCREEN_FONT_SCALE : 1
  const baseFont = (subtitleSettings.value.fontSize || 32) * (subtitleSettings.value.fontScale || 1)
  subtitleDisplayRef.value.updateSettings({
    fontSize: Math.round(baseFont * widthScale * fullscreenScale),
    color: subtitleSettings.value.color,
    backgroundColor: `rgba(0, 0, 0, ${subtitleSettings.value.backgroundOpacity ?? 0.75})`,
    textShadow: subtitleSettings.value.textShadow,
    borderRadius: subtitleSettings.value.borderRadius,
    padding: subtitleSettings.value.padding,
    lineHeight: subtitleSettings.value.lineHeight,
    position: subtitleSettings.value.position,
    stroke: subtitleSettings.value.stroke
  })
}

const refreshSubtitleLayout = (resetPosition = false) => {
  nextTick(() => {
    pushSettingsToSubtitle()
    if (!subtitleDisplayRef.value) return
    if (resetPosition) {
      subtitleDisplayRef.value.centerSubtitle()
    } else {
      subtitleDisplayRef.value.updatePosition()
    }
    requestAnimationFrame(() => {
      subtitleDisplayRef.value?.updatePosition()
    })
  })
}

const handleSubtitlePlayerResize = () => {
  refreshSubtitleLayout(false)
}

const updateSubtitleSettings = (settings: any) => {
  subtitleSettings.value = { ...subtitleSettings.value, ...settings }
  saveSubtitleSettings()
  pushSettingsToSubtitle()
}

watch(
  subtitleSettings,
  () => {
    saveSubtitleSettings()
    pushSettingsToSubtitle()
  },
  { deep: true }
)

const resetSubtitlePosition = () => {
  if (subtitleDisplayRef.value) {
    subtitleDisplayRef.value.resetPosition()
  }
}

const resetSubtitleSettings = () => {
  subtitleSettings.value = { ...defaultSubtitleSettings }
  saveSubtitleSettings()
  pushSettingsToSubtitle()
  if (subtitleDisplayRef.value) {
    subtitleDisplayRef.value.resetPosition()
  }
  ElMessage.success('已恢复默认设置')
}

// 传给菜单的语言列表：过滤掉空壳占位记录（同语言重复时只留非空）
const menuSubtitles = computed(() =>
  subtitleList.value.filter((sub: any) => {
    const c = sub.content || sub.cues || []
    return Array.isArray(c) && c.length > 1
  })
)

const loadSubtitles = async () => {
  try {
    console.log('[字幕] 开始加载字幕列表, videoId:', videoId.value)
    const response = await subtitleApi.getSubtitles(videoId.value)
    console.log('[字幕] 字幕列表响应:', response)
    if (response.code === 200) {
      subtitleList.value = response.data || []
      console.log('[字幕] 字幕列表:', subtitleList.value)
      // 列表接口已带 content，直接过滤空壳（如 whisper 占位记录），避免选中空字幕
      const nonEmpty = subtitleList.value.filter((sub: any) => {
        const content = sub.content || sub.cues || []
        return Array.isArray(content) && content.length > 1
      })
      const defaultSub = nonEmpty.find((sub: any) => sub.isDefault || sub.is_default)
      const target = defaultSub || nonEmpty[0] || null
      if (target) {
        console.log('[字幕] 加载默认字幕:', target.language, '条数:', (target.content || []).length)
        await loadSubtitleContent(target.language)
      } else {
        console.log('[字幕] 没有可用字幕')
      }
    }
} catch (error) {
    console.error('[字幕] 加载字幕列表失败:', error)
    ElMessage.warning('字幕服务暂时不可用')
  }
}

const loadSubtitleContent = async (language: string) => {
  try {
    console.log('[字幕] 开始加载字幕内容, language:', language)
    // 优先直接用列表接口携带的 content（同语言可能有空壳占位记录，取非空的）
    const fromList = subtitleList.value.find((sub: any) => {
      if (sub.language !== language) return false
      const c = sub.content || sub.cues || []
      return Array.isArray(c) && c.length > 1
    })
    const listContent = fromList?.content || fromList?.cues || []
    if (Array.isArray(listContent) && listContent.length > 1) {
      currentSubtitle.value = fromList
      currentSubtitleContent.value = listContent
      currentLanguage.value = language
      console.log('[字幕] 从列表载入字幕, 条数:', currentSubtitleContent.value.length)
      return
    }
    const response = await subtitleApi.getSubtitle(videoId.value, language)
    console.log('[字幕] 字幕内容响应:', response)
    if (response && response.code === 200 && response.data) {
      currentSubtitle.value = response.data
      currentSubtitleContent.value = response.data.content || response.data.cues || []
      currentLanguage.value = language
      console.log('[字幕] 字幕内容已加载, 条数:', currentSubtitleContent.value.length)
    } else {
      console.log('[字幕] 字幕内容为空')
    }
  } catch (error) {
    console.error('[字幕] 加载字幕内容失败:', error)
    ElMessage.warning('字幕内容加载失败')
  }
}

const switchLanguage = async (language: string) => {
  if (language === currentLanguage.value || !language) return
  await loadSubtitleContent(language)
}

const formatDate = (dateStr: any) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

const loadDanmakus = async () => {
  try {
    localLoadingDanmus.value = true
    emitLoadingDanmus()
    const currentVideo = props.manuscriptInfo?.videos?.[props.currentVideoIndex]
    if (!currentVideo) return
    const danmakuResponse = await interactionApi.getDanmakus(currentVideo.id)
    if (danmakuResponse.code === 200) {
      const danmakuData = danmakuResponse.data || []
      localDanmuList.value = danmakuData.map((danmaku: any) => ({
        text: danmaku.content,
        time: Number.parseFloat(danmaku.time) || 0,
        color: danmaku.color || '#ffffff',
        sendTime: formatDate(danmaku.created_at || danmaku.createTime)
      }))
      emitDanmuList()
      const updatedVideoInfo = { ...props.videoInfo, danmuLoadedCount: localDanmuList.value.length }
      emit('update:videoInfo', updatedVideoInfo)
      if (art) {
        art.plugins.artplayerPluginDanmuku.config({
          danmuku: localDanmuList.value
        })
      }
    }
  } catch (error) {
    console.error('获取弹幕失败:', error)
  } finally {
    localLoadingDanmus.value = false
    emitLoadingDanmus()
  }
}

const handleVideoTimeUpdate = () => {
  if (art) {
    const currentTime = art.currentTime || 0
    const duration = art.duration || 0
    emit('timeUpdate', { currentTime, duration })
  }
}

let playerInitialized = false
let lastAppliedUrl = ''
let timeUpdateInterval: any = null

const VOL_ICONS = {
  high: 'M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z',
  low: 'M18.5 12c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM5 9v6h4l5 5V4L9 9H5z',
  muted: 'M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51A8.796 8.796 0 0 0 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3 3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06a8.99 8.99 0 0 0 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4 9.91 6.09 12 8.18V4z'
}
const updateVolIcon = (el: HTMLElement, art: any) => {
  const icon = el.querySelector('.art-vol-icon')
  if (!icon) return
  const vol = art.muted ? 0 : art.volume
  const path = vol === 0 ? VOL_ICONS.muted : vol > 0.5 ? VOL_ICONS.high : VOL_ICONS.low
  icon.innerHTML = `<svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor"><path d="${path}"/></svg>`
}

const initPlayer = async () => {
  if (playerInitialized) return
  playerInitialized = true

  const { default: ArtplayerPluginDanmuku } = await import('artplayer-plugin-danmuku')

  let defaultUrl = ''
  const qualityOptions: any[] = []

  if (props.videoInfo.playUrlHd) {
    qualityOptions.push({
      default: true,
      name: '1080P 高清',
      html: '1080P 高清',
      url: props.videoInfo.playUrlHd
    })
  }

  if (props.videoInfo.playUrlSd) {
    qualityOptions.push({
      default: !props.videoInfo.playUrlHd,
      name: '720P 标清',
      html: '720P 标清',
      url: props.videoInfo.playUrlSd
    })
  }

  if (props.videoInfo.playUrlLd) {
    qualityOptions.push({
      default: !props.videoInfo.playUrlHd && !props.videoInfo.playUrlSd,
      name: '480P 流畅',
      html: '480P 流畅',
      url: props.videoInfo.playUrlLd
    })
  }

  if (qualityOptions.length === 0) {
    ElMessage.error('视频播放地址缺失')
    return
  }
  defaultUrl = qualityOptions[0].url
  lastAppliedUrl = defaultUrl

  const playerConfig: any = {
    container: playerRef.value,
    url: defaultUrl,
    poster: props.videoInfo.coverUrl,
    volume: 0.7,
    isLive: false,
    muted: false,
    autoplay: false,
    pip: true,
    autoSize: false,
    autoMini: true,
    screenshot: false,
    setting: true,
    loop: false,
    flip: true,
    playbackRate: false,
    volume: 0.7,
    aspectRatio: false,
    fullscreen: true,
    fullscreenWeb: true,
    miniProgressBar: true,
    quality: qualityOptions,
    theme: '#23ade5',
    lang: 'zh-cn',
    type: defaultUrl.endsWith('.m3u8') ? 'm3u8' : 'mp4',
    customType: {
      m3u8: async (video: any, url: string, art: any) => {
        if (video.canPlayType('application/x-mpegURL')) {
          video.src = url
        } else {
          const { default: Hls } = await import('hls.js')
          const hls = new Hls()
          hls.loadSource(url)
          hls.attachMedia(video)
        }
      }
    },
    plugins: [
      ArtplayerPluginDanmuku({
        danmuku: localDanmuList.value,
        speed: 5,
        opacity: 1,
        color: '#ffffff',
        fontSize: 25,
        synchronousPlayback: true,
        maxlength: 50,
        margin: [10, 10, 10, 10],
        emitter: true,
        beforeEmit: async (danmu: any) => {
          if (!hasAuthSession()) {
            ElMessage.warning('请先登录')
            return false
          }
          const currentVideo = props.manuscriptInfo?.videos?.[props.currentVideoIndex]
          if (!currentVideo) return false
          try {
            const response = await interactionApi.sendDanmaku(
              currentVideo.id,
              props.manuscriptInfo?.id,
              danmu.text,
              Number(danmu.time) || 0,
              danmu.color || '#ffffff',
              danmu.mode || 0
            )
            if (response.code === 200) {
              localDanmuList.value.push({
                text: danmu.text,
                time: danmu.time,
                color: danmu.color || '#ffffff',
                sendTime: formatDate(new Date())
              })
              emitDanmuList()
              const updatedVideoInfo = { ...props.videoInfo, danmuLoadedCount: localDanmuList.value.length }
              emit('update:videoInfo', updatedVideoInfo)
              ElMessage.success('发送成功')
              return true
            } else {
              ElMessage.error(response.message || '发送失败')
              return false
            }
          } catch (error) {
            console.error('发送弹幕失败:', error)
            ElMessage.error('发送失败，请稍后重试')
            return false
          }
        }
      })
    ]
  }

  let volEl: HTMLElement | null = null
  let volSlider: HTMLElement | null = null

  playerConfig.controls = [
    // 倍速（B站右侧顺序：清晰度 → 倍速 → 字幕 → 音量 → 设置 → …）
    {
      position: 'right',
      index: 20,
      html: '倍速',
      tooltip: '倍速',
      selector: [2, 1.5, 1.25, 1, 0.75, 0.5].map(v => ({
        value: v,
        name: `ctrl-speed-${v}`,
        html: v === 1 ? '1.0x' : `${v}x`,
        default: v === 1
      })),
      onSelect: function (this: any, item: any) {
        this.playbackRate = item.value
        return item.html
      }
    },
    // 字幕（原有功能）
    {
      position: 'right',
      index: 22,
      html: '字幕',
      tooltip: '字幕设置',
      click: function (this: any, _mgr: any, ev: Event) {
        const btn = (ev?.currentTarget as HTMLElement) || null
        if (btn) {
          const btnRect = btn.getBoundingClientRect()
          const host = playerRef.value?.getBoundingClientRect()
          if (host) {
            subtitleMenuAnchorX.value = btnRect.left + btnRect.width / 2 - host.left
          }
        }
        subtitleSettingsVisible.value = !subtitleSettingsVisible.value
        subtitleToggleGuard = true
        setTimeout(() => { subtitleToggleGuard = false }, 0)
      }
    },
    // 音量：图标 + 点击静音切换 + hover 弹竖向音量条
    {
      position: 'right',
      index: 24,
      html: '<span class="art-icon art-vol-icon"></span>',
      tooltip: '音量',
      click: function (this: any) {
        this.muted = !this.muted
        if (volEl) updateVolIcon(volEl, this)
      },
      mounted: function (this: any, el: HTMLElement) {
        const art = this
        volEl = el
        el.style.position = 'relative'
        updateVolIcon(el, art)
        art.on('video:volumechange', () => {
          updateVolIcon(el, art)
          if (volSlider) {
            const input = volSlider.querySelector('input')
            if (input && document.activeElement !== input) {
              input.value = String(Math.round((art.volume || 0) * 100))
            }
          }
        })
        art.proxy(el, 'mouseenter', () => {
          if (!volSlider) {
            volSlider = document.createElement('div')
            volSlider.className = 'art-vol-slider'
            volSlider.style.cssText = 'position:absolute;bottom:100%;left:50%;transform:translateX(-50%);background:rgba(0,0,0,0.8);border-radius:4px;padding:10px 0;width:32px;height:100px;display:flex;justify-content:center;z-index:310;'
            const input = document.createElement('input')
            input.type = 'range'
            input.min = '0'
            input.max = '100'
            input.value = String(Math.round((art.volume || 0) * 100))
            input.style.cssText = 'writing-mode:vertical-lr;direction:rtl;height:76px;width:4px;margin:0;accent-color:#fff;cursor:pointer;'
            input.addEventListener('input', () => {
              art.volume = Number(input.value) / 100
              if (art.muted && art.volume > 0) art.muted = false
              updateVolIcon(el, art)
            })
            volSlider.appendChild(input)
            el.appendChild(volSlider)
          }
          volSlider.style.display = 'flex'
        })
        art.proxy(el, 'mouseleave', () => {
          if (volSlider) volSlider.style.display = 'none'
        })
      }
    }
  ]

  playerConfig.layers = [
    {
      name: 'customSubtitle',
      html: '',
      style: {
        position: 'absolute',
        inset: '0',
        overflow: 'hidden',
        pointerEvents: 'none',
        zIndex: 42
      },
      mounted: (layer: any) => {
        if (subtitleDisplayRef.value) {
          layer.appendChild(subtitleDisplayRef.value.$el)
          refreshSubtitleLayout(false)
        }
      }
    },
    {
      name: 'subtitleSettingsPanel',
      html: '',
      style: {
        position: 'absolute',
        inset: '0',
        zIndex: 100,
        pointerEvents: 'none',
        overflow: 'hidden'
      },
      mounted: (layer: any) => {
        if (subtitleSettingsPanelRef.value) {
          layer.appendChild(subtitleSettingsPanelRef.value)
        }
      }
    }
  ]

  art = new Artplayer(playerConfig)

  art.on('ready', () => {
    refreshSubtitleLayout(true)
    if (props.resumeTime > 0) {
      setTimeout(() => applyResumeTime(props.resumeTime), 100)
    }
  })

  // 监听容器尺寸变化（缩放/窗口变化时），让 art-controls 跟着重新布局
  if (playerRef.value && typeof ResizeObserver !== 'undefined') {
    let resizeTimer: any = null
    const ro = new ResizeObserver(() => {
      if (resizeTimer) clearTimeout(resizeTimer)
      resizeTimer = setTimeout(() => {
        try {
          if (art && typeof art.resize === 'function') {
            art.resize()
          }
        } catch (e) {
          // ignore
        }
      }, 100)
    })
    ro.observe(playerRef.value)
    ;(playerRef.value as any).__artResizeObserver = ro
  }

  let lastLoggedSecond = -1

  const updateCurrentTime = () => {
    if (art && art.currentTime !== undefined) {
      const time = art.currentTime
      currentVideoTime.value = time
      const currentSecond = Math.floor(time)
      if (currentSecond !== lastLoggedSecond) {
        console.log('[字幕] 时间更新, 当前时间:', time.toFixed(2))
        lastLoggedSecond = currentSecond
      }
      handleVideoTimeUpdate()
    }
  }

  art.on('timeupdate', () => {
    updateCurrentTime()
  })

  timeUpdateInterval = setInterval(updateCurrentTime, 100)

  art.on('play', () => {
    console.log('[字幕] 视频开始播放')
    if (!timeUpdateInterval) {
      timeUpdateInterval = setInterval(updateCurrentTime, 100)
    }
  })

  art.on('pause', () => {
    console.log('[字幕] 视频暂停, 当前时间:', art.currentTime)
  })

  art.on('fullscreen', (isFullscreen: boolean) => {
    isNativeFullscreen.value = isFullscreen
    updateFullscreenMode()
    setTimeout(() => refreshSubtitleLayout(true), 100)
  })

  art.on('fullscreenWeb', (isFullscreen: boolean) => {
    isWebFullscreen.value = isFullscreen
    updateFullscreenMode()
    setTimeout(() => refreshSubtitleLayout(true), 100)
  })

  art.on('resize', () => {
    refreshSubtitleLayout(false)
  })

  loadSubtitles()
  window.addEventListener('resize', handleSubtitlePlayerResize)
}

const updatePlayer = async () => {
  if (!art) {
    hasAppliedResume.value = false
    pendingResumeTime.value = props.resumeTime
    await initPlayer()
    return
  }

  const qualityOptions: any[] = []

  if (props.videoInfo.playUrlHd) {
    qualityOptions.push({
      default: true,
      name: '1080P 高清',
      html: '1080P 高清',
      url: props.videoInfo.playUrlHd
    })
  }

  if (props.videoInfo.playUrlSd) {
    qualityOptions.push({
      default: !props.videoInfo.playUrlHd,
      name: '720P 标清',
      html: '720P 标清',
      url: props.videoInfo.playUrlSd
    })
  }

  if (props.videoInfo.playUrlLd) {
    qualityOptions.push({
      default: !props.videoInfo.playUrlHd && !props.videoInfo.playUrlSd,
      name: '480P 流畅',
      html: '480P 流畅',
      url: props.videoInfo.playUrlLd
    })
  }

  if (qualityOptions.length === 0) {
    ElMessage.error('视频播放地址缺失')
    return
  }

  const targetUrl = qualityOptions[0].url
  if (playerInitialized && targetUrl === lastAppliedUrl) return

  lastAppliedUrl = targetUrl
  art.switchUrl(qualityOptions[0].url)
  if (qualityOptions.length > 0) {
    art.quality = qualityOptions
  }

  art.currentTime = 0
  hasAppliedResume.value = false
  pendingResumeTime.value = props.resumeTime

  if (props.resumeTime > 0) {
    setTimeout(() => applyResumeTime(props.resumeTime), 100)
  }

  loadDanmakus()
  loadSubtitles()
}

watch(() => props.videoInfo?.playUrl, (newUrl) => {
  if (!newUrl) return
  if (!playerInitialized) {
    initPlayer()
    loadDanmakus()
  }
}, { immediate: true })

watch(() => [props.videoInfo?.playUrlHd, props.videoInfo?.playUrlSd, props.videoInfo?.playUrlLd], () => {
  if (playerInitialized && props.videoInfo?.playUrl) {
    updatePlayer()
  }
}, { deep: true })

watch(subtitleEnabled, (val) => {
  try {
    safeStorage.setItem(SUBTITLE_ENABLED_KEY, String(val))
  } catch (e) {
  }
})

// 仅 t 参数变化（同一分P续播）时应用续播时间
watch(() => props.resumeTime, (val) => {
  if (art && playerInitialized && val > 0) {
    setTimeout(() => applyResumeTime(val), 100)
  }
})

const seekTo = (time: number) => {
  if (art && time !== undefined) {
    art.currentTime = time
    console.log('跳转到时间:', time)
  }
}

defineExpose({
  seekTo
})

onUnmounted(() => {
  if (timeUpdateInterval) {
    clearInterval(timeUpdateInterval)
    timeUpdateInterval = null
  }
  window.removeEventListener('resize', handleSubtitlePlayerResize)
  document.removeEventListener('click', onDocClickForSubtitlePanel)
  if (art) {
    art.destroy()
    art = null
  }
})
</script>

<template>
  <div class="video-player-wrapper">
    <div ref="playerRef" class="video-player"></div>
    <SubtitleDisplay
      ref="subtitleDisplayRef"
      :subtitles="currentSubtitleContent"
      :current-time="currentVideoTime"
      :enabled="subtitleEnabled"
    />
    <div ref="subtitleSettingsPanelRef" class="subtitle-menu-anchor" @click.stop>
      <SubtitleMenu
        v-model:visible="subtitleSettingsVisible"
        :anchor-x="subtitleMenuAnchorX"
        :subtitles="menuSubtitles"
        :current-language="currentLanguage"
        :enabled="subtitleEnabled"
        :settings="subtitleSettings"
        @update:enabled="(v: boolean) => { subtitleEnabled = v; safeStorage.setItem(SUBTITLE_ENABLED_KEY, String(v)) }"
        @select-language="switchLanguage"
        @update-settings="updateSubtitleSettings"
        @reset-settings="resetSubtitleSettings"
      />
    </div>
  </div>
  <div class="video-status-bar-simple">
    <div class="status-info">
      <span class="status-item">
        <el-icon><View /></el-icon>
        {{ Math.max(1, props.videoInfo.watchingCount || 0).toLocaleString() }}人正在看
      </span>
      <span class="status-item">
        <el-icon><ChatDotRound /></el-icon>
        已装载{{ (props.videoInfo.danmuLoadedCount || 0).toLocaleString() }}条弹幕
      </span>
    </div>
  </div>
</template>

<style scoped>
.video-player-wrapper {
  position: relative;
  width: 100%;
}

/* 控制栏仅在鼠标悬浮播放器时显示，不悬浮即隐藏（含暂停状态） */
.video-player :deep(.art-video-player:not(.art-hover)) .art-bottom {
  opacity: 0 !important;
  pointer-events: none !important;
}

.video-player {
  width: 100% !important;
  height: auto !important;
  max-width: 100%;
  aspect-ratio: 16/9;
  background-color: #000;
  position: relative;
  overflow: hidden;
}

/* ArtPlayer 内部也强制跟随容器 */
.video-player :deep(.art-video-player) {
  width: 100% !important;
  height: 100% !important;
  max-width: 100%;
}

/* ArtPlayer 内部控件不要换行（保持单行布局，用 CSS var 缩放解决空间问题） */
.video-player :deep(.art-controls) {
    flex-wrap: nowrap !important;
}
.video-player :deep(.art-control-volume) {
    display: none !important;
}

/* artplayer-plugin-danmuku 设置面板：强制覆盖全局 a/inherit 的深色字 */
.video-player :deep(.apd-config-panel-inner),
.video-player :deep(.apd-config-panel-inner *),
.video-player :deep(.apd-style-panel-inner),
.video-player :deep(.apd-style-panel-inner *) {
    color: #fff !important;
}
.video-player :deep(.apd-config-panel-inner .apd-icon),
.video-player :deep(.apd-style-panel-inner .apd-icon) {
    fill: #fff !important;
}
.video-player :deep(.apd-slider-steps),
.video-player :deep(.apd-slider-step),
.video-player :deep(.apd-mode),
.video-player :deep(.apd-other),
.video-player :deep(.apd-value),
.video-player :deep(.apd-check-on),
.video-player :deep(.apd-check-off) {
    color: #fff !important;
}
.video-player :deep(.apd-config-panel-inner input[type=range]) {
    background: rgba(255,255,255,0.15) !important;
}

.subtitle-menu-anchor {
  position: absolute;
  inset: 0;
  z-index: 120;
  pointer-events: none;
}
.subtitle-menu-anchor :deep(.bsub-menu) {
  pointer-events: auto;
}

.video-status-bar-simple {
  background-color: #fff;
  padding: 20px 0;
}

.video-status-bar-simple .status-info {
  display: flex;
  gap: 20px;
  padding-left: 10px;
}

.video-status-bar-simple .status-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #666;
}

.video-status-bar-simple .status-item .el-icon {
  font-size: 16px;
}
</style>