<script setup lang="ts">
import { safeStorage } from '@/utils/safeStorage'
import Artplayer from 'artplayer'
import { ElMessage } from 'element-plus'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import { subtitleApi } from '@/api/subtitle.ts'
import SubtitleDisplay from '@/components/SubtitleDisplay.vue'
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
let subtitleToggleGuard = false

const onDocClickForSubtitlePanel = (e: Event) => {
  if (subtitleToggleGuard) return
  const panel = subtitleSettingsPanelRef.value
  if (panel && !panel.contains(e.target as Node)) {
    subtitleSettingsVisible.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocClickForSubtitlePanel)
})
const SUBTITLE_SETTINGS_KEY = 'mybilibili_subtitle_settings'
const FULLSCREEN_FONT_SCALE = 1.5
const defaultSubtitleSettings = {
  fontSize: 32,
  color: '#ffffff',
  backgroundColor: 'rgba(0, 0, 0, 0.75)',
  backgroundOpacity: 0.75,
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
      subtitleSettings.value = { ...defaultSubtitleSettings, ...JSON.parse(saved) }
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
  const scale = isFullscreenMode.value ? FULLSCREEN_FONT_SCALE : 1
  subtitleDisplayRef.value.updateSettings({
    ...subtitleSettings.value,
    fontSize: Math.round(subtitleSettings.value.fontSize * scale)
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

const loadSubtitles = async () => {
  try {
    console.log('[字幕] 开始加载字幕列表, videoId:', videoId.value)
    const response = await subtitleApi.getSubtitles(videoId.value)
    console.log('[字幕] 字幕列表响应:', response)
    if (response.code === 200) {
      subtitleList.value = response.data || []
      console.log('[字幕] 字幕列表:', subtitleList.value)
      const defaultSub = subtitleList.value.find((sub: any) => sub.isDefault)
      if (defaultSub) {
        console.log('[字幕] 加载默认字幕:', defaultSub.language)
        await loadSubtitleContent(defaultSub.language)
      } else if (subtitleList.value.length > 0) {
        console.log('[字幕] 加载第一个字幕:', subtitleList.value[0].language)
        await loadSubtitleContent(subtitleList.value[0].language)
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
    playbackRate: true,
    aspectRatio: true,
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
          const token = safeStorage.getItem("token")
          if (!token) {
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

  playerConfig.controls = [
    {
      position: 'right',
      html: '<span class="art-icon">字幕</span>',
      tooltip: '字幕设置',
      click: () => {
        subtitleSettingsVisible.value = !subtitleSettingsVisible.value
        subtitleToggleGuard = true
        setTimeout(() => { subtitleToggleGuard = false }, 0)
      },
      mounted: function (this: any) {
        this.innerHTML = `<span class="art-icon">字幕</span>`
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
})

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
    <div ref="subtitleSettingsPanelRef" v-show="subtitleSettingsVisible" class="subtitle-settings-panel" @click.stop>
      <div class="settings-header">
        <span>字幕设置</span>
        <el-button link size="small" @click="subtitleSettingsVisible = false">
          <el-icon><CircleClose /></el-icon>
        </el-button>
      </div>
      <div class="settings-content">
        <div class="setting-item">
          <span class="setting-label">字幕开关</span>
          <el-switch v-model="subtitleEnabled" />
        </div>
        <div class="setting-item" v-if="subtitleList.length > 1">
          <span class="setting-label">字幕语言</span>
          <el-select v-model="currentLanguage" placeholder="选择语言" size="small" @change="switchLanguage">
            <el-option
              v-for="sub in subtitleList"
              :key="sub.language"
              :label="sub.languageName || sub.language_name || sub.language"
              :value="sub.language"
            />
          </el-select>
        </div>
        <div class="setting-item">
          <span class="setting-label">字体大小</span>
          <el-slider v-model="subtitleSettings.fontSize" :min="12" :max="40" :step="1" @change="updateSubtitleSettings({ fontSize: $event })" />
          <span class="setting-value">{{ subtitleSettings.fontSize }}px</span>
        </div>
        <div class="setting-item">
          <span class="setting-label">字体颜色</span>
          <el-color-picker :teleported="false" v-model="subtitleSettings.color" @change="updateSubtitleSettings({ color: $event })" />
        </div>
        <div class="setting-item">
          <span class="setting-label">背景透明度</span>
          <el-slider v-model="subtitleSettings.backgroundOpacity" :min="0" :max="1" :step="0.1" @change="updateSubtitleSettings({ backgroundColor: `rgba(0, 0, 0, ${$event})` })" />
        </div>
        <div class="setting-item">
          <span class="setting-label">圆角大小</span>
          <el-slider v-model="subtitleSettings.borderRadius" :min="0" :max="20" :step="1" @change="updateSubtitleSettings({ borderRadius: $event })" />
        </div>
        <div class="setting-actions">
          <el-button size="small" @click="resetSubtitleSettings">恢复默认</el-button>
          <div style="display:flex;gap:8px;">
            <el-button size="small" @click="resetSubtitlePosition">重置位置</el-button>
            <el-button size="small" type="primary" @click="subtitleSettingsVisible = false">完成</el-button>
          </div>
        </div>
      </div>
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
  flex: 1;
  aspect-ratio: 16/9;
  background-color: #000;
  min-height: 450px;
  position: relative;
  overflow: hidden;
}

.subtitle-settings-panel {
  position: absolute;
  bottom: 50px;
  right: 10px;
  width: 280px;
  max-width: calc(100% - 20px);
  max-height: calc(100% - 70px);
  background: rgba(28, 28, 28, 0.95);
  border-radius: 8px;
  padding: 16px;
  z-index: 120;
  color: #fff;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  overflow: auto;
  pointer-events: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.subtitle-settings-panel::-webkit-scrollbar {
  display: none;
}

.subtitle-settings-panel .settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.subtitle-settings-panel .settings-header span {
  font-size: 16px;
  font-weight: 500;
}

.subtitle-settings-panel .setting-item {
  margin-bottom: 16px;
}

.subtitle-settings-panel .setting-label {
  display: block;
  font-size: 13px;
  color: #ccc;
  margin-bottom: 8px;
}

.subtitle-settings-panel .setting-value {
  font-size: 12px;
  color: #999;
  margin-left: 8px;
}

.subtitle-settings-panel .setting-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
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