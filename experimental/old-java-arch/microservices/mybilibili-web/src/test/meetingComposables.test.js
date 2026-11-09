import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeetingChat } from '../composables/useMeetingChat.js'
import { usePeerConnectionMesh } from '../composables/usePeerConnectionMesh.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    info: vi.fn()
  }
}))

const createMeetingChat = (options = {}) => {
  const me = ref({ id: 1, name: '主持人' })
  const participantList = ref([
    { userId: 1, name: '主持人', self: true },
    { userId: 2, name: 'Alice', self: false },
    { userId: 3, name: 'Bob', self: false }
  ])
  const sendMeetingMessage = options.sendMeetingMessage || vi.fn(() => true)
  const chat = useMeetingChat({
    me,
    participantList,
    sendMeetingMessage,
    ...options
  })

  return {
    me,
    participantList,
    sendMeetingMessage,
    ...chat
  }
}

class FakeRTCPeerConnection {
  static instances = []

  constructor(config) {
    this.config = config
    this.tracks = []
    this.senders = []
    this.iceCandidates = []
    this.remoteDescription = null
    this.localDescription = null
    this.closed = false
    this.iceConnectionState = 'new'
    this.connectionState = 'new'
    FakeRTCPeerConnection.instances.push(this)
  }

  addTrack(track, stream) {
    this.tracks.push({ track, stream })
    this.senders.push({
      track,
      replaceTrack: vi.fn(newTrack => {
        this.replacedTrack = newTrack
      })
    })
  }

  async createOffer(options) {
    this.lastOfferOptions = options
    return { type: 'offer', sdp: options?.iceRestart ? 'restart-offer-sdp' : 'offer-sdp' }
  }

  async createAnswer() {
    return { type: 'answer', sdp: 'answer-sdp' }
  }

  async setLocalDescription(description) {
    this.localDescription = description
  }

  async setRemoteDescription(description) {
    this.remoteDescription = description
  }

  async addIceCandidate(candidate) {
    this.iceCandidates.push(candidate)
  }

  getSenders() {
    return this.senders
  }

  close() {
    this.closed = true
  }
}

describe('useMeetingChat', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps @ text when selecting a mention and sends targeted chat', () => {
    const chat = createMeetingChat()

    chat.chatText.value = 'hello @Al'
    chat.handleChatInput()

    expect(chat.showMentionDropdown.value).toBe(true)
    expect(chat.mentionSuggestions.value).toEqual([
      expect.objectContaining({ userId: 2, name: 'Alice' })
    ])

    chat.selectMention({ userId: 2, name: 'Alice' })
    expect(chat.chatText.value).toBe('hello @Alice ')
    expect(chat.mentionSelected.value).toBe(2)

    expect(chat.sendChat()).toBe(true)
    expect(chat.sendMeetingMessage).toHaveBeenCalledWith({
      type: 'chat',
      targetUserId: 2,
      data: { text: 'hello @Alice' }
    })
    expect(chat.chatMessages.value).toEqual([
      expect.objectContaining({
        from: '主持人',
        text: 'hello @Alice',
        self: true,
        mentionId: 2
      })
    ])
    expect(chat.chatText.value).toBe('')
    expect(chat.mentionSelected.value).toBeNull()
  })

  it('does not clear local input when chat delivery fails', () => {
    const chat = createMeetingChat({
      sendMeetingMessage: vi.fn(() => false)
    })
    chat.chatText.value = '发不出去'

    expect(chat.sendChat()).toBe(false)

    expect(chat.chatText.value).toBe('发不出去')
    expect(chat.chatMessages.value).toHaveLength(0)
  })

  it('marks and notifies incoming direct and textual mentions', () => {
    const chat = createMeetingChat()

    chat.receiveChatMessage({
      userId: '2',
      userName: 'Alice',
      targetUserId: '1',
      data: { text: '请看这里' }
    })
    chat.receiveChatMessage({
      userId: 3,
      userName: 'Bob',
      data: { text: '@主持人 轮到你了' }
    })
    expect(chat.receiveChatMessage({
      userId: 4,
      userName: 'Carol',
      data: { text: 123 }
    })).toBe(false)

    expect(chat.chatMessages.value).toEqual([
      expect.objectContaining({ from: 'Alice', text: '请看这里', mentionId: '2' }),
      expect.objectContaining({ from: 'Bob', text: '@主持人 轮到你了', mentionId: 3 })
    ])
    expect(ElMessage.info).toHaveBeenCalledTimes(2)
    expect(ElMessage.info).toHaveBeenCalledWith('Alice 在会议中提到了你')
    expect(ElMessage.info).toHaveBeenCalledWith('Bob 在会议中提到了你')
  })
})

describe('usePeerConnectionMesh', () => {
  beforeEach(() => {
    FakeRTCPeerConnection.instances = []
    vi.stubGlobal('RTCPeerConnection', FakeRTCPeerConnection)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('creates initiator peer connections, binds local tracks, and forwards local ICE', async () => {
    const videoTrack = { kind: 'video' }
    const audioTrack = { kind: 'audio' }
    const stream = { getTracks: () => [videoTrack, audioTrack] }
    const localStream = ref(stream)
    const remotePeers = {}
    const ensurePeerEntry = vi.fn(peerId => {
      remotePeers[peerId] = remotePeers[peerId] || {}
    })
    const sendSignal = vi.fn()

    const mesh = usePeerConnectionMesh({
      localStream,
      remotePeers,
      ensurePeerEntry,
      sendSignal
    })

    const pc = await mesh.createPeerConnection('peer-1', true)
    pc.onicecandidate({ candidate: { candidate: 'local-ice' } })
    pc.ontrack({ streams: ['remote-stream'] })

    expect(pc.tracks).toEqual([
      { track: videoTrack, stream },
      { track: audioTrack, stream }
    ])
    expect(sendSignal).toHaveBeenCalledWith('offer', 'peer-1', 'offer-sdp')
    expect(sendSignal).toHaveBeenCalledWith('ice-candidate', 'peer-1', { candidate: 'local-ice' })
    expect(ensurePeerEntry).toHaveBeenCalledWith('peer-1')
    expect(remotePeers['peer-1'].stream).toBe('remote-stream')
  })

  it('queues ICE until an offer creates a peer connection and then sends an answer', async () => {
    const localStream = ref(null)
    const remotePeers = { 'peer-1': { old: true } }
    const sendSignal = vi.fn()

    const mesh = usePeerConnectionMesh({
      localStream,
      remotePeers,
      ensurePeerEntry: vi.fn(peerId => {
        remotePeers[peerId] = remotePeers[peerId] || {}
      }),
      sendSignal,
      resetRemoteOnOffer: true
    })

    await mesh.handleIceCandidate('peer-1', { candidate: 'early-ice' })
    await mesh.handleOffer('peer-1', 'offer-sdp')

    const pc = mesh.peerConnections['peer-1']
    expect(pc.remoteDescription).toEqual({ type: 'offer', sdp: 'offer-sdp' })
    expect(pc.iceCandidates).toEqual([{ candidate: 'early-ice' }])
    expect(remotePeers['peer-1']).toBeUndefined()
    expect(sendSignal).toHaveBeenCalledWith('answer', 'peer-1', 'answer-sdp')
  })

  it('restarts ICE only when allowed and replaces video tracks across peers', async () => {
    const oldVideo = { kind: 'video' }
    const newVideo = { kind: 'video' }
    const localStream = ref({ getTracks: () => [oldVideo] })
    const sendSignal = vi.fn()
    const logger = {
      warn: vi.fn(),
      error: vi.fn()
    }
    const mesh = usePeerConnectionMesh({
      localStream,
      remotePeers: {},
      ensurePeerEntry: vi.fn(),
      sendSignal,
      canRestartIce: () => true,
      logger
    })

    const pc = await mesh.createPeerConnection('peer-1', false)
    pc.iceConnectionState = 'failed'
    await pc.oniceconnectionstatechange()
    pc.connectionState = 'disconnected'
    pc.onconnectionstatechange()
    mesh.replaceVideoTrack(newVideo)

    expect(pc.lastOfferOptions).toEqual({ iceRestart: true })
    expect(sendSignal).toHaveBeenCalledWith('offer', 'peer-1', 'restart-offer-sdp')
    expect(pc.replacedTrack).toBe(newVideo)
    expect(logger.warn).toHaveBeenCalledWith('[WebRTC] ICE failed, 尝试 restart', 'peer-1')
    expect(logger.warn).toHaveBeenCalledWith('[WebRTC] 连接状态', 'peer-1', 'disconnected')

    mesh.removePeerConnection('peer-1')
    expect(pc.closed).toBe(true)
    expect(mesh.peerConnections['peer-1']).toBeUndefined()
  })
})
