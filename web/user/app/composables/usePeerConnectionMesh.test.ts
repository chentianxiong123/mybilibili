import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { usePeerConnectionMesh } from './usePeerConnectionMesh'

class FakeRTCPeerConnection {
  static instances: FakeRTCPeerConnection[] = []
  localDescription: any = null
  remoteDescription: any = null
  connectionState = 'new'
  iceConnectionState = 'new'
  iceGatheringState = 'new'
  ontrack: any = null
  onicecandidate: any = null
  onconnectionstatechange: any = null
  oniceconnectionstatechange: any = null
  addedTracks: any[] = []
  getSendersResult: any[] = []
  remoteStreams: any[] = []

  constructor(public cfg_: any) {
    FakeRTCPeerConnection.instances.push(this)
  }
  async createOffer(opts?: any) {
    return { type: 'offer', sdp: 'v=0\r\no=- ...' + JSON.stringify(opts || {}) }
  }
  async createAnswer() {
    return { type: 'answer', sdp: 'v=0\r\no=- answer' }
  }
  async setLocalDescription(desc: any) {
    this.localDescription = desc
  }
  async setRemoteDescription(desc: any) {
    this.remoteDescription = desc
  }
  addTrack(track: any, stream: any) {
    this.addedTracks.push({ track, stream })
  }
  addIceCandidate(c: any) {
    if (!this.remoteDescription) throw new Error('no remote desc yet')
    return Promise.resolve()
  }
  getSenders() {
    return this.getSendersResult
  }
  close() {
    this.connectionState = 'closed'
  }
}

beforeEach(() => {
  FakeRTCPeerConnection.instances = []
  vi.stubGlobal('RTCPeerConnection', FakeRTCPeerConnection)
})

describe('usePeerConnectionMesh', () => {
  describe('createPeerConnection', () => {
    it('isInitiator=true 时自动 createOffer + setLocalDescription + sendSignal offer', async () => {
      const sendSignal = vi.fn()
      const localStream = ref(null)
      const remotePeers: any = {}
      const ensurePeerEntry = vi.fn()
      const r = usePeerConnectionMesh({ localStream, remotePeers, ensurePeerEntry, sendSignal })
      const pc = await r.createPeerConnection('p1', true)
      expect(pc).toBeTruthy()
      expect(FakeRTCPeerConnection.instances.length).toBe(1)
      expect(pc.localDescription?.type).toBe('offer')
      expect(sendSignal).toHaveBeenCalledWith('offer', 'p1', expect.stringContaining('v=0'))
    })

    it('isInitiator=false 时不主动 offer', async () => {
      const sendSignal = vi.fn()
      const localStream = ref(null)
      const r = usePeerConnectionMesh({ localStream, remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal })
      await r.createPeerConnection('p2', false)
      expect(sendSignal).not.toHaveBeenCalled()
    })

    it('已有 pc 时直接返回已有实例', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal })
      const pc1 = await r.createPeerConnection('p1', true)
      const pc2 = await r.createPeerConnection('p1', true)
      expect(pc1).toBe(pc2)
      expect(FakeRTCPeerConnection.instances.length).toBe(1)
    })

    it('localStream 存在时把每个 track 添加到 pc', async () => {
      const track1 = { kind: 'video' }
      const track2 = { kind: 'audio' }
      const stream: any = { getTracks: () => [track1, track2] }
      const localStream = ref(stream)
      const r = usePeerConnectionMesh({ localStream, remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      const pc = await r.createPeerConnection('p1', true)
      expect(pc.addedTracks.length).toBe(2)
    })

    it('onicecandidate 有 candidate 时 sendSignal ice-candidate', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal })
      await r.createPeerConnection('p1', false)
      const pc = FakeRTCPeerConnection.instances[0]
      pc.onicecandidate?.({ candidate: { sdpMid: '0', candidate: '...' } })
      expect(sendSignal).toHaveBeenCalledWith('ice-candidate', 'p1', expect.objectContaining({ sdpMid: '0' }))
    })

    it('onicecandidate 无 candidate 时不调用 sendSignal', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal })
      await r.createPeerConnection('p1', false)
      FakeRTCPeerConnection.instances[0].onicecandidate?.({ candidate: null })
      expect(sendSignal).not.toHaveBeenCalled()
    })

    it('ontrack 时把 streams[0] 写到 remotePeers[peerId]', async () => {
      const remotePeers: any = {}
      const stream = { id: 'remote-1' }
      const ensurePeerEntry = vi.fn((id: string) => {
        if (!remotePeers[id]) remotePeers[id] = { stream: null }
      })
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers, ensurePeerEntry, sendSignal: vi.fn() })
      await r.createPeerConnection('p1', false)
      const pc = FakeRTCPeerConnection.instances[0]
      pc.ontrack?.({ streams: [stream] })
      expect(ensurePeerEntry).toHaveBeenCalledWith('p1')
      expect(remotePeers.p1.stream).toBe(stream)
    })
  })

  describe('handleOffer', () => {
    it('setRemoteDescription(offer) + createAnswer + sendSignal answer', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal })
      await r.handleOffer('p1', 'remote-sdp')
      expect(sendSignal).toHaveBeenCalledWith('answer', 'p1', expect.stringContaining('answer'))
      const pc = FakeRTCPeerConnection.instances[0]
      expect(pc.remoteDescription).toEqual({ type: 'offer', sdp: 'remote-sdp' })
    })

    it('resetRemoteOnOffer=true 时删除 remotePeers[peerId]', async () => {
      const remotePeers: any = { p1: { stream: 'old' } }
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers, ensurePeerEntry: vi.fn(), sendSignal: vi.fn(), resetRemoteOnOffer: true })
      await r.handleOffer('p1', 'sdp')
      // 'old' 已被 delete，pc 还未触发 ontrack 写回，所以条目此时不存在
      expect(remotePeers.p1?.stream).not.toBe('old')
    })

    it('已有 pc 时先 close 再删除再创建', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal })
      await r.createPeerConnection('p1', false)
      const pc1 = FakeRTCPeerConnection.instances[0]
      const closeSpy = vi.spyOn(pc1, 'close')
      await r.handleOffer('p1', 'sdp')
      expect(closeSpy).toHaveBeenCalled()
      expect(FakeRTCPeerConnection.instances.length).toBe(2)
    })
  })

  describe('handleAnswer', () => {
    it('setRemoteDescription(answer) 写入 pc', async () => {
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', true)
      const pc = FakeRTCPeerConnection.instances[0]
      await r.handleAnswer('p1', 'answer-sdp')
      expect(pc.remoteDescription).toEqual({ type: 'answer', sdp: 'answer-sdp' })
    })

    it('pc 不存在时静默返回', async () => {
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await expect(r.handleAnswer('nonexistent', 'sdp')).resolves.toBeUndefined()
    })
  })

  describe('handleIceCandidate', () => {
    it('pc 已设 remoteDescription 时直接 addIceCandidate', async () => {
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', true)
      const pc = FakeRTCPeerConnection.instances[0]
      // 模拟远端描述存在
      pc.remoteDescription = { type: 'offer', sdp: 'x' }
      const addSpy = vi.spyOn(pc, 'addIceCandidate')
      await r.handleIceCandidate('p1', { sdpMid: '0', candidate: '...' })
      expect(addSpy).toHaveBeenCalled()
    })

    it('pc 不存在或无 remoteDescription 时缓存到 pendingIce', async () => {
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', false) // no remoteDesc
      const pc = FakeRTCPeerConnection.instances[0]
      const addSpy = vi.spyOn(pc, 'addIceCandidate')
      await r.handleIceCandidate('p1', { sdpMid: '0' })
      expect(addSpy).not.toHaveBeenCalled()
      // 后续 handleAnswer 会 flush
      await r.handleAnswer('p1', 'x') // 这个是 answer 流，不会 flush... 看源码，handleAnswer 也有 flush
      expect(addSpy).toHaveBeenCalled()
    })

    it('pc 不存在时缓存', async () => {
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await expect(r.handleIceCandidate('none', { sdpMid: '0' })).resolves.toBeUndefined()
    })
  })

  describe('removePeerConnection', () => {
    it('close pc 并清理', async () => {
      const remotePeers: any = { p1: { stream: 'x' } }
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', true)
      const pc = FakeRTCPeerConnection.instances[0]
      const closeSpy = vi.spyOn(pc, 'close')
      r.removePeerConnection('p1')
      expect(closeSpy).toHaveBeenCalled()
      expect(remotePeers.p1).toBeUndefined()
    })

    it('pc 不存在时只清 remotePeers', () => {
      const remotePeers: any = { p1: { stream: 'x' } }
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      r.removePeerConnection('p1')
      expect(remotePeers.p1).toBeUndefined()
    })
  })

  describe('closeAllPeerConnections', () => {
    it('关闭所有 pc', async () => {
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', true)
      await r.createPeerConnection('p2', true)
      const close1 = vi.spyOn(FakeRTCPeerConnection.instances[0], 'close')
      const close2 = vi.spyOn(FakeRTCPeerConnection.instances[1], 'close')
      r.closeAllPeerConnections()
      expect(close1).toHaveBeenCalled()
      expect(close2).toHaveBeenCalled()
    })
  })

  describe('replaceVideoTrack', () => {
    it('替换每个 pc 中 video sender 的 track', async () => {
      const newTrack = { kind: 'video', id: 'new' }
      const sender: any = { track: { kind: 'video', id: 'old' }, replaceTrack: vi.fn(function (this: any, t: any) { this.track = t }) }
      FakeRTCPeerConnection.prototype.getSenders = function() { return [sender] }
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', true)
      r.replaceVideoTrack(newTrack as any)
      expect(sender.track).toBe(newTrack)
    })

    it('newTrack 为 null 时不替换', async () => {
      const sender: any = { track: { kind: 'video' }, replaceTrack: vi.fn() }
      FakeRTCPeerConnection.prototype.getSenders = function() { return [sender] }
      const r = usePeerConnectionMesh({ localStream: ref(null), remotePeers: {}, ensurePeerEntry: vi.fn(), sendSignal: vi.fn() })
      await r.createPeerConnection('p1', true)
      r.replaceVideoTrack(null as any)
      expect(sender.replaceTrack).not.toHaveBeenCalled()
    })
  })

  describe('ICE restart', () => {
    it('iceConnectionState=failed + canRestartIce=true 时 createOffer(iceRestart)', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({
        localStream: ref(null),
        remotePeers: {},
        ensurePeerEntry: vi.fn(),
        sendSignal,
        canRestartIce: () => true,
      })
      await r.createPeerConnection('p1', true)
      const pc = FakeRTCPeerConnection.instances[0]
      // 清空之前的 offer sendSignal 调用计数
      sendSignal.mockClear()
      pc.iceConnectionState = 'failed'
      pc.oniceconnectionstatechange?.()
      await new Promise(r => setTimeout(r, 0))
      expect(sendSignal).toHaveBeenCalledWith('offer', 'p1', expect.stringContaining('iceRestart'))
    })

    it('canRestartIce=false 时不 restart', async () => {
      const sendSignal = vi.fn()
      const r = usePeerConnectionMesh({
        localStream: ref(null),
        remotePeers: {},
        ensurePeerEntry: vi.fn(),
        sendSignal,
        canRestartIce: () => false,
      })
      await r.createPeerConnection('p1', true)
      sendSignal.mockClear()
      FakeRTCPeerConnection.instances[0].iceConnectionState = 'failed'
      FakeRTCPeerConnection.instances[0].oniceconnectionstatechange?.()
      await new Promise(r => setTimeout(r, 0))
      expect(sendSignal).not.toHaveBeenCalled()
    })
  })
})