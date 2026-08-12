import { nextTick, ref } from 'vue'
import { ElMessage } from 'element-plus'
import flvJs from 'flv.js'
import Artplayer from 'artplayer'
import ArtplayerPluginDanmuku from 'artplayer-plugin-danmuku'

const DEFAULT_LIVE_THEME = '#fb7299'

export function useLiveRoomPlayer({
  room,
  hostname = typeof window !== 'undefined' ? window.location.hostname : 'localhost',
  ArtplayerCtor = Artplayer,
  flvLib = flvJs,
  DanmukuPlugin = ArtplayerPluginDanmuku,
  message = ElMessage,
  nextTickFn = nextTick,
  setTimeoutFn = (handler, delay) => setTimeout(handler, delay),
  clearTimeoutFn = timerId => clearTimeout(timerId),
  logger = console
}) {
  const playerRef = ref(null)
  const replayPlayerRef = ref(null)
  const isBuffering = ref(false)
  const currentQuality = ref('自动')
  const bufferStallCount = ref(0)
  const showReplayTab = ref(false)
  const replayList = ref([])
  const currentReplay = ref(null)
  let bufferingTimer = null
  let liveArt = null
  let replayArt = null

  const clearBufferingTimer = () => {
    if (!bufferingTimer) return false
    clearTimeoutFn(bufferingTimer)
    bufferingTimer = null
    return true
  }

  const showBuffering = () => {
    clearBufferingTimer()
    isBuffering.value = true
    return true
  }

  const hideBuffering = () => {
    clearBufferingTimer()
    bufferingTimer = setTimeoutFn(() => {
      isBuffering.value = false
      bufferingTimer = null
    }, 500)
    return true
  }

  const destroyFlvPlayer = (player) => {
    if (!player?.flvPlayer) return false
    try {
      player.flvPlayer.pause()
      player.flvPlayer.unload()
      player.flvPlayer.detachMediaElement()
      player.flvPlayer.destroy()
    } catch (e) {}
    player.flvPlayer = null
    return true
  }

  const destroyLivePlayer = () => {
    if (!liveArt) return false
    const player = liveArt
    try { player.destroy() } catch (e) {}
    destroyFlvPlayer(player)
    liveArt = null
    return true
  }

  const initPlayer = () => {
    if (!room.value || !playerRef.value) return false
    destroyLivePlayer()
    const flvUrl = `http://${hostname}:28080/live/${room.value.streamKey}.flv`
    let pendingFlvPlayer = null

    const player = new ArtplayerCtor({
      container: playerRef.value,
      url: flvUrl,
      type: 'flv',
      isLive: true,
      autoplay: true,
      muted: true,
      autoSize: false,
      autoOrientation: true,
      setting: true,
      playbackRate: false,
      aspectRatio: false,
      fullscreen: true,
      fullscreenWeb: true,
      pip: true,
      miniProgressBar: false,
      mutex: true,
      backdrop: true,
      theme: DEFAULT_LIVE_THEME,
      customType: {
        flv: (video, url) => {
          if (!flvLib.isSupported()) return
          const flvPlayer = flvLib.createPlayer({
            type: 'flv',
            url,
            isLive: true,
            config: {
              enableWorker: true,
              enableStashBuffer: false,
              stashInitialSize: 128,
              autoCleanupSourceBuffer: true,
              autoCleanupMaxDuration: 3,
              autoCleanupMinRecoverDuration: 2
            }
          })
          flvPlayer.attachMediaElement(video)
          flvPlayer.load()
          flvPlayer.on(flvLib.Events.ERROR, (errType, errDetail) => {
            logger.error?.('[LiveRoom] FLV 错误:', errType, errDetail)
          })
          if (liveArt) {
            liveArt.flvPlayer = flvPlayer
          } else {
            pendingFlvPlayer = flvPlayer
          }
        }
      },
      plugins: [
        DanmukuPlugin({
          danmuku: [],
          speed: 5,
          opacity: 1,
          fontSize: 22,
          synchronousPlayback: false
        })
      ]
    })

    liveArt = player
    if (pendingFlvPlayer) {
      liveArt.flvPlayer = pendingFlvPlayer
      pendingFlvPlayer = null
    }

    player.on('video:waiting', () => showBuffering())
    player.on('video:playing', () => {
      hideBuffering()
      bufferStallCount.value = 0
    })
    player.on('video:stalled', () => {
      bufferStallCount.value++
      if (bufferStallCount.value >= 3) {
        currentQuality.value = '流畅'
        message.warning('当前网络不稳定，已自动切换到流畅模式')
      }
    })
    player.on('error', (e) => logger.error?.('[LiveRoom] 播放器错误:', e))
    player.on('destroy', () => destroyFlvPlayer(player))
    return true
  }

  const emitDanmaku = (text, color) => {
    const plugin = liveArt?.plugins?.artplayerPluginDanmuku
    if (!plugin || typeof plugin.emit !== 'function') return false
    try {
      plugin.emit({ text, color, border: false })
      return true
    } catch (e) {
      logger.warn?.('emit danmaku failed', e)
      return false
    }
  }

  const loadReplays = async () => {
    // 后端接入回放列表 API 后，可在这里替换为实际请求。
    replayList.value = []
  }

  const destroyReplayPlayer = () => {
    if (!replayArt) return false
    try { replayArt.destroy() } catch (e) {}
    replayArt = null
    return true
  }

  const playReplay = async (replay) => {
    destroyReplayPlayer()
    currentReplay.value = replay
    await nextTickFn()
    if (!replayPlayerRef.value || !replay?.url) return false
    replayArt = new ArtplayerCtor({
      container: replayPlayerRef.value,
      url: replay.url,
      type: replay.url.includes('.m3u8') ? 'm3u8' : undefined,
      autoplay: true,
      setting: true,
      fullscreen: true,
      playbackRate: true
    })
    return true
  }

  const closeReplay = () => {
    currentReplay.value = null
    return destroyReplayPlayer()
  }

  const cleanupLiveRoomPlayer = () => {
    clearBufferingTimer()
    isBuffering.value = false
    closeReplay()
    destroyLivePlayer()
    bufferStallCount.value = 0
    currentQuality.value = '自动'
  }

  return {
    playerRef,
    replayPlayerRef,
    isBuffering,
    currentQuality,
    bufferStallCount,
    showReplayTab,
    replayList,
    currentReplay,
    initPlayer,
    emitDanmaku,
    loadReplays,
    playReplay,
    closeReplay,
    destroyLivePlayer,
    destroyReplayPlayer,
    cleanupLiveRoomPlayer
  }
}
