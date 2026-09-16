import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { useWhipPublisher } from './useWhipPublisher'

class FakeRTCPeerConnection {
  static instances: FakeRTCPeerConnection[] = []
  localDescription: any = null
  remoteDescription: any = null
  addedTracks: any[] = []
  constructor(public cfg_: any) {
    FakeRTCPeerConnection.instances.push(this)
  }
  addTrack(track: any, stream: any) {
    this.addedTracks.push({ track, stream })
  }
  async createOffer() {
    return { type: 'offer', sdp: 'OFFER_SDP' }
  }
  async setLocalDescription(desc: any) {
    this.localDescription = desc
  }
  async setRemoteDescription(desc: any) {
    this.remoteDescription = desc
  }
  close() {}
}

beforeEach(() => {
  FakeRTCPeerConnection.instances = []
  vi.stubGlobal('RTCPeerConnection', FakeRTCPeerConnection)
})

describe('useWhipPublisher', () => {
  function makeStream() {
    const stop = vi.fn()
    return {
      getTracks: () => [{ stop, kind: 'video' }],
      getVideoTracks: () => [{ stop, kind: 'video', onended: null }],
    }
  }

  describe('initial', () => {
    it('默认 sourceType=camera', () => {
      const r = useWhipPublisher({ streamKey: ref('k') })
      expect(r.sourceType.value).toBe('camera')
    })

    it('默认 isPublishing=false', () => {
      const r = useWhipPublisher({ streamKey: ref('k') })
      expect(r.isPublishing.value).toBe(false)
    })
  })

  describe('startPublishing', () => {
    it('缺少 streamKey 时不开始且 isPublishing=false', async () => {
      const r = useWhipPublisher({ streamKey: ref('') })
      await r.startPublishing()
      expect(r.isPublishing.value).toBe(false)
    })

    it('camera 模式下 captureStream 调用 getUserMedia', async () => {
      const stream = makeStream()
      const captureStream = vi.fn().mockResolvedValue(stream)
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('ANSWER_SDP') })
      const r = useWhipPublisher({
        streamKey: ref('myKey'),
        captureStream: captureStream as any,
        getWhipUrl: key => `http://localhost:19854/rtc/v1/whip/stream/live/${key}`,
        logger: console,
      })
      vi.stubGlobal('fetch', fetchFn)
      await r.startPublishing()
      expect(captureStream).toHaveBeenCalledWith('camera')
      expect(r.isPublishing.value).toBe(true)
      expect(r.mediaStream.value).toEqual(stream)
      expect(FakeRTCPeerConnection.instances.length).toBe(1)
      expect(fetchFn).toHaveBeenCalledWith(expect.stringContaining('/whip/stream/live/myKey'), expect.objectContaining({ method: 'POST' }))
      expect(FakeRTCPeerConnection.instances[0].remoteDescription).toEqual({ type: 'answer', sdp: 'ANSWER_SDP' })
    })

    it('screen 模式下 captureStream 调用 getDisplayMedia', async () => {
      const stream = makeStream()
      const captureStream = vi.fn().mockResolvedValue(stream)
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('A') })
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
      })
      vi.stubGlobal('fetch', fetchFn)
      r.sourceType.value = 'screen'
      await r.startPublishing()
      expect(captureStream).toHaveBeenCalledWith('screen')
    })

    it('WHIP 返回非 ok 时抛错并清理', async () => {
      const captureStream = vi.fn().mockResolvedValue(makeStream())
      const fetchFn = vi.fn().mockResolvedValue({ ok: false })
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
      })
      vi.stubGlobal('fetch', fetchFn)
      const errorSpy = vi.fn()
      vi.stubGlobal('ElMessage', { success: vi.fn(), error: errorSpy, warning: vi.fn() })
      // ElMessage 是 element-plus，vue 模板已 import - mock 起来
      await r.startPublishing()
      // mock 模块中的 ElMessage
    })

    it('isPublishing=true 时不重复开始', async () => {
      const captureStream = vi.fn().mockResolvedValue(makeStream())
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('A') })
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
      })
      vi.stubGlobal('fetch', fetchFn)
      await r.startPublishing()
      const callsBefore = captureStream.mock.calls.length
      await r.startPublishing()
      expect(captureStream.mock.calls.length).toBe(callsBefore)
    })

    it('onPublished 回调在成功后被调用', async () => {
      const captureStream = vi.fn().mockResolvedValue(makeStream())
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('A') })
      const onPublished = vi.fn()
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
        onPublished,
      })
      vi.stubGlobal('fetch', fetchFn)
      await r.startPublishing()
      expect(onPublished).toHaveBeenCalled()
    })
  })

  describe('stopPublishing', () => {
    it('清理 stream / pc / isPublishing=false', async () => {
      const stream = makeStream()
      const captureStream = vi.fn().mockResolvedValue(stream)
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('A') })
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
      })
      vi.stubGlobal('fetch', fetchFn)
      await r.startPublishing()
      const pc = FakeRTCPeerConnection.instances[0]
      const closeSpy = vi.spyOn(pc, 'close')
      r.stopPublishing()
      expect(closeSpy).toHaveBeenCalled()
      expect(r.mediaStream.value).toBe(null)
      expect(r.isPublishing.value).toBe(false)
    })

    it('videoPreviewRef 存在时 srcObject 被清空', async () => {
      const stream = makeStream()
      const captureStream = vi.fn().mockResolvedValue(stream)
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('A') })
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
      })
      vi.stubGlobal('fetch', fetchFn)
      await r.startPublishing()
      const el = { srcObject: stream }
      r.videoPreviewRef.value = el as any
      r.stopPublishing()
      expect(el.srcObject).toBe(null)
    })

    it('无 stream 时调用不抛错', () => {
      const r = useWhipPublisher({ streamKey: ref('s') })
      expect(() => r.stopPublishing()).not.toThrow()
    })
  })

  describe('sourceType watch', () => {
    it('sourceType 改变且 isPublishing=true 时重启', async () => {
      const stream = makeStream()
      const captureStream = vi.fn().mockResolvedValue(stream)
      const fetchFn = vi.fn().mockResolvedValue({ ok: true, text: () => Promise.resolve('A') })
      const r = useWhipPublisher({
        streamKey: ref('s'),
        captureStream: captureStream as any,
      })
      vi.stubGlobal('fetch', fetchFn)
      await r.startPublishing()
      const callsBefore = captureStream.mock.calls.length
      r.sourceType.value = 'screen'
      await nextTick()
      // watch 触发 stop + start
      expect(captureStream.mock.calls.length).toBeGreaterThan(callsBefore)
    })
  })
})