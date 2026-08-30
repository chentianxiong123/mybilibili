import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useLiveRoomPlayer } from '../composables/useLiveRoomPlayer.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    warning: vi.fn()
  }
}))

vi.mock('artplayer', () => ({ default: vi.fn() }))
vi.mock('artplayer-plugin-danmuku', () => ({ default: vi.fn(() => 'danmuku-plugin') }))
vi.mock('flv.js', () => ({
  default: {
    Events: { ERROR: 'error' },
    isSupported: vi.fn(() => false),
    createPlayer: vi.fn()
  }
}))

const createArtplayerHarness = () => {
  const instances = []
  const ArtplayerCtor = vi.fn(function (options) {
    const handlers = {}
    const instance = {
      options,
      plugins: {
        artplayerPluginDanmuku: {
          emit: vi.fn()
        }
      },
      on: vi.fn((event, handler) => {
        handlers[event] = handler
      }),
      trigger: (event, ...args) => handlers[event]?.(...args),
      destroy: vi.fn(() => handlers.destroy?.())
    }
    instances.push(instance)
    return instance
  })
  return { ArtplayerCtor, instances }
}

const createFlvHarness = () => {
  const flvPlayer = {
    attachMediaElement: vi.fn(),
    load: vi.fn(),
    on: vi.fn(),
    pause: vi.fn(),
    unload: vi.fn(),
    detachMediaElement: vi.fn(),
    destroy: vi.fn()
  }
  const flvLib = {
    Events: { ERROR: 'error' },
    isSupported: vi.fn(() => true),
    createPlayer: vi.fn(() => flvPlayer)
  }
  return { flvLib, flvPlayer }
}

describe('useLiveRoomPlayer', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    document.body.innerHTML = ''
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('initializes live playback, handles buffer events, and emits danmaku', () => {
    const { ArtplayerCtor, instances } = createArtplayerHarness()
    const { flvLib } = createFlvHarness()
    const live = useLiveRoomPlayer({
      room: ref({ streamKey: 'room-stream' }),
      hostname: 'example.test',
      ArtplayerCtor,
      flvLib,
      DanmukuPlugin: vi.fn(() => 'plugin')
    })
    live.playerRef.value = document.createElement('div')

    expect(live.initPlayer()).toBe(true)

    expect(ArtplayerCtor).toHaveBeenCalledTimes(1)
    expect(instances[0].options.url).toBe('http://example.test:28080/live/room-stream.flv')
    expect(instances[0].options.type).toBe('flv')

    instances[0].options.customType.flv(document.createElement('video'), instances[0].options.url)
    expect(flvLib.createPlayer).toHaveBeenCalledWith(expect.objectContaining({
      type: 'flv',
      url: 'http://example.test:28080/live/room-stream.flv',
      isLive: true
    }))

    expect(live.emitDanmaku('你好', '#fff')).toBe(true)
    expect(instances[0].plugins.artplayerPluginDanmuku.emit).toHaveBeenCalledWith({
      text: '你好',
      color: '#fff',
      border: false
    })

    instances[0].trigger('video:waiting')
    expect(live.isBuffering.value).toBe(true)

    instances[0].trigger('video:playing')
    expect(live.bufferStallCount.value).toBe(0)
    vi.advanceTimersByTime(500)
    expect(live.isBuffering.value).toBe(false)

    instances[0].trigger('video:stalled')
    instances[0].trigger('video:stalled')
    instances[0].trigger('video:stalled')
    expect(live.currentQuality.value).toBe('流畅')
    expect(ElMessage.warning).toHaveBeenCalledWith('当前网络不稳定，已自动切换到流畅模式')
  })

  it('logs live playback and danmaku failures through the injected logger', () => {
    const logger = { error: vi.fn(), warn: vi.fn() }
    const { ArtplayerCtor, instances } = createArtplayerHarness()
    const { flvLib, flvPlayer } = createFlvHarness()
    const live = useLiveRoomPlayer({
      room: ref({ streamKey: 'room-stream' }),
      ArtplayerCtor,
      flvLib,
      logger
    })
    live.playerRef.value = document.createElement('div')

    expect(live.initPlayer()).toBe(true)
    instances[0].options.customType.flv(document.createElement('video'), instances[0].options.url)
    const flvErrorHandler = flvPlayer.on.mock.calls.find(([event]) => event === 'error')[1]
    const flvError = new Error('flv failed')
    flvErrorHandler('NetworkError', 'detail')
    instances[0].trigger('error', flvError)

    instances[0].plugins.artplayerPluginDanmuku.emit.mockImplementation(() => {
      throw new Error('danmaku failed')
    })

    expect(live.emitDanmaku('你好', '#fff')).toBe(false)
    expect(logger.error).toHaveBeenCalledWith('[LiveRoom] FLV 错误:', 'NetworkError', 'detail')
    expect(logger.error).toHaveBeenCalledWith('[LiveRoom] 播放器错误:', flvError)
    expect(logger.warn).toHaveBeenCalledWith('emit danmaku failed', expect.any(Error))
  })

  it('cleans up buffering timers and FLV resources', () => {
    const { ArtplayerCtor, instances } = createArtplayerHarness()
    const { flvLib, flvPlayer } = createFlvHarness()
    const live = useLiveRoomPlayer({
      room: ref({ streamKey: 'room-stream' }),
      ArtplayerCtor,
      flvLib
    })
    live.playerRef.value = document.createElement('div')

    live.initPlayer()
    instances[0].options.customType.flv(document.createElement('video'), instances[0].options.url)
    instances[0].trigger('video:waiting')
    instances[0].trigger('video:playing')
    expect(vi.getTimerCount()).toBe(1)

    live.cleanupLiveRoomPlayer()

    expect(vi.getTimerCount()).toBe(0)
    expect(live.isBuffering.value).toBe(false)
    expect(live.currentQuality.value).toBe('自动')
    expect(instances[0].destroy).toHaveBeenCalled()
    expect(flvPlayer.pause).toHaveBeenCalled()
    expect(flvPlayer.unload).toHaveBeenCalled()
    expect(flvPlayer.detachMediaElement).toHaveBeenCalled()
    expect(flvPlayer.destroy).toHaveBeenCalled()
  })

  it('creates and replaces replay players', async () => {
    const { ArtplayerCtor, instances } = createArtplayerHarness()
    const live = useLiveRoomPlayer({
      room: ref({ streamKey: 'room-stream' }),
      ArtplayerCtor,
      nextTickFn: vi.fn(async () => true)
    })
    live.replayPlayerRef.value = document.createElement('div')

    expect(await live.playReplay({ title: '回放一', url: 'https://cdn.test/a.m3u8' })).toBe(true)
    expect(live.currentReplay.value.title).toBe('回放一')
    expect(instances[0].options.type).toBe('m3u8')

    expect(await live.playReplay({ title: '回放二', url: 'https://cdn.test/b.mp4' })).toBe(true)
    expect(instances[0].destroy).toHaveBeenCalled()
    expect(instances[1].options.type).toBeUndefined()

    live.closeReplay()
    expect(live.currentReplay.value).toBeNull()
    expect(instances[1].destroy).toHaveBeenCalled()
  })
})
