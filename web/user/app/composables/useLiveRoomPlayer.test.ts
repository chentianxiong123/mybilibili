import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useLiveRoomPlayer } from './useLiveRoomPlayer'

class FakeFlvPlayer {
  static instances: FakeFlvPlayer[] = []
  attached: any = null
  loaded = false
  handlers: Record<string, any> = {}
  constructor(public cfg: any) {
    FakeFlvPlayer.instances.push(this)
  }
  attachMediaElement(el: any) { this.attached = el }
  load() { this.loaded = true }
  unload() {}
  detachMediaElement() {}
  destroy() {}
  pause() {}
  on(evt: string, cb: any) { this.handlers[evt] = cb }
}

function makeFakeArtPlayer() {
  const handlers: Record<string, any> = {}
  const player: any = {
    destroy: vi.fn(),
    flvPlayer: null,
    plugins: {
      artplayerPluginDanmuku: {
        emit: vi.fn(),
      }
    },
    on(evt: string, cb: any) { handlers[evt] = cb },
    trigger(evt: string, ...args: any[]) { handlers[evt]?.(...args) },
  }
  return player
}

const FakeFlvLib = {
  isSupported: vi.fn(() => true),
  createPlayer: vi.fn((cfg: any) => {
    const p = new FakeFlvPlayer(cfg)
    return p
  }),
  Events: { ERROR: 'ERROR' },
}

const FakeDanmukuPlugin = vi.fn(() => ({}))

describe('useLiveRoomPlayer', () => {
  let ArtplayerCtor: any
  let message: any
  let logger: any

  beforeEach(() => {
    FakeFlvPlayer.instances = []
    FakeFlvLib.isSupported.mockClear()
    FakeFlvLib.createPlayer.mockClear()
    ArtplayerCtor = vi.fn(function (this: any) { return makeFakeArtPlayer() })
    message = { warning: vi.fn(), error: vi.fn(), success: vi.fn() }
    logger = { error: vi.fn(), warn: vi.fn() }
  })

  describe('初始状态', () => {
    it('默认 isBuffering=false，bufferStallCount=0', () => {
      const r = useLiveRoomPlayer({ room: ref(null), ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      expect(r.isBuffering.value).toBe(false)
      expect(r.bufferStallCount.value).toBe(0)
      expect(r.currentQuality.value).toBe('自动')
    })
  })

  describe('initPlayer', () => {
    it('room 为 null 时返回 false，不创建 player', () => {
      const r = useLiveRoomPlayer({ room: ref(null), ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.playerRef.value = document.createElement('div')
      expect(r.initPlayer()).toBe(false)
      expect(ArtplayerCtor).not.toHaveBeenCalled()
    })

    it('playerRef 为 null 时返回 false', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      expect(r.initPlayer()).toBe(false)
    })

    it('正常初始化时创建 Artplayer 并使用正确的 url', () => {
      const room = ref({ streamKey: 'skey' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message, logger })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      expect(ArtplayerCtor).toHaveBeenCalledTimes(1)
      const opts = ArtplayerCtor.mock.calls[0][0]
      expect(opts.url).toBe('http://localhost:28080/live/skey.flv')
      expect(opts.type).toBe('flv')
      expect(opts.theme).toBe('#fb7299')
    })

    it('自定义 hostname 生效', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, hostname: 'example.com', ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      const opts = ArtplayerCtor.mock.calls[0][0]
      expect(opts.url).toContain('example.com')
    })

    it('flvLib.isSupported=false 时不创建 flvPlayer', () => {
      FakeFlvLib.isSupported.mockReturnValueOnce(false)
      const room = ref({ streamKey: 'k' })
      const r = useLivePushRoomControlsWith(room)
      r.initPlayer()
      expect(FakeFlvPlayer.instances.length).toBe(0)
    })

    it('video:stalled 触发 3 次后切换到流畅模式', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message, logger })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      const player = ArtplayerCtor.mock.results[0].value
      player.trigger('video:stalled')
      player.trigger('video:stalled')
      expect(r.currentQuality.value).toBe('自动')
      player.trigger('video:stalled')
      expect(r.currentQuality.value).toBe('流畅')
      expect(message.warning).toHaveBeenCalled()
    })

    it('video:waiting 触发 isBuffering=true', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      const player = ArtplayerCtor.mock.results[0].value
      player.trigger('video:waiting')
      expect(r.isBuffering.value).toBe(true)
    })

    it('video:playing 触发后 500ms 内 isBuffering=false', () => {
      vi.useFakeTimers()
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      const player = ArtplayerCtor.mock.results[0].value
      player.trigger('video:waiting')
      expect(r.isBuffering.value).toBe(true)
      player.trigger('video:playing')
      vi.advanceTimersByTime(500)
      expect(r.isBuffering.value).toBe(false)
      vi.useRealTimers()
    })
  })

  describe('emitDanmaku', () => {
    it('plugin 不存在时返回 false', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      expect(r.emitDanmaku('hi', '#fff')).toBe(false)
    })

    it('plugin.emit 抛出时不抛错', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message, logger })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      const player = ArtplayerCtor.mock.results[0].value
      player.plugins.artplayerPluginDanmuku.emit = vi.fn(() => { throw new Error('emit fail') })
      expect(() => r.emitDanmaku('hi', '#fff')).not.toThrow()
      expect(r.emitDanmaku('hi', '#fff')).toBe(false)
    })
  })

  describe('destroyLivePlayer / cleanupLiveRoomPlayer', () => {
    it('未初始化时 destroyLivePlayer 返回 false', () => {
      const r = useLiveRoomPlayer({ room: ref({ streamKey: 'k' }), ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      expect(r.destroyLivePlayer()).toBe(false)
    })

    it('初始化后 destroyLivePlayer 调用 player.destroy()', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      const player = ArtplayerCtor.mock.results[0].value
      r.destroyLivePlayer()
      expect(player.destroy).toHaveBeenCalled()
    })

    it('cleanupLiveRoomPlayer 重置状态', () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.playerRef.value = document.createElement('div')
      r.initPlayer()
      r.bufferStallCount.value = 5
      r.currentQuality.value = '流畅'
      r.cleanupLiveRoomPlayer()
      expect(r.bufferStallCount.value).toBe(0)
      expect(r.currentQuality.value).toBe('自动')
    })
  })

  describe('loadReplays / playReplay / closeReplay', () => {
    it('loadReplays 是异步 no-op，replayList 为空', async () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      await r.loadReplays()
      expect(r.replayList.value).toEqual([])
    })

    it('playReplay url 缺失时返回 false', async () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      await r.playReplay({} as any)
      expect(ArtplayerCtor).not.toHaveBeenCalled()
    })

    it('playReplay 正常时创建 Artplayer', async () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.replayPlayerRef.value = document.createElement('div')
      await r.playReplay({ url: 'http://x.m3u8' } as any)
      expect(ArtplayerCtor).toHaveBeenCalled()
      expect(r.currentReplay.value).toEqual({ url: 'http://x.m3u8' })
    })

    it('closeReplay 销毁 replayPlayer', async () => {
      const room = ref({ streamKey: 'k' })
      const r = useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message })
      r.replayPlayerRef.value = document.createElement('div')
      await r.playReplay({ url: 'http://x.mp4' } as any)
      r.closeReplay()
      expect(r.currentReplay.value).toBe(null)
    })
  })

  // helper
  function useLivePushRoomControlsWith(room: any) {
    return useLiveRoomPlayer({ room, ArtplayerCtor, flvLib: FakeFlvLib as any, DanmukuPlugin: FakeDanmukuPlugin as any, message, logger })
  }
})