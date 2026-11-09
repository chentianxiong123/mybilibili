import { beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeetingSignalingRouter } from '../composables/useMeetingSignalingRouter.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    warning: vi.fn(),
    info: vi.fn()
  }
}))

const createRouter = (options = {}) => {
  const me = options.me || ref({ id: 1, name: '我' })
  const currentRoom = options.currentRoom || ref({ id: 1, roomCode: 'ROOM01' })
  const hostUserId = options.hostUserId || ref(1)
  const isHost = options.isHost || ref(false)
  const waitingForHost = options.waitingForHost || ref(true)
  const remotePeers = reactive(options.remotePeers || {
    2: { name: '旧名字', audioEnabled: true, videoEnabled: true }
  })
  const audioEnabled = options.audioEnabled || ref(true)
  const spotlightId = options.spotlightId || ref(null)

  const deps = {
    me,
    currentRoom,
    hostUserId,
    isHost,
    waitingForHost,
    remotePeers,
    audioEnabled,
    spotlightId,
    ensurePeerEntry: vi.fn((peerId, name) => {
      if (!remotePeers[peerId]) remotePeers[peerId] = { name }
    }),
    createPeerConnection: vi.fn(async () => true),
    broadcastSelfState: vi.fn(),
    removePeerConnection: vi.fn(),
    handleOffer: vi.fn(async () => true),
    handleAnswer: vi.fn(async () => true),
    handleIceCandidate: vi.fn(async () => true),
    updateAudioActivity: vi.fn(),
    toggleAudio: vi.fn(() => {
      audioEnabled.value = !audioEnabled.value
      return true
    }),
    enterApprovedRoom: vi.fn(async () => true),
    cancelWaitingForHost: vi.fn(),
    leaveRoom: vi.fn(),
    cleanupRoomSession: vi.fn(),
    cleanupRoomArtifacts: vi.fn(() => {
      hostUserId.value = null
    }),
    loadMyRooms: vi.fn(async () => true),
    receiveChatMessage: vi.fn(),
    handleJoinRequestMessage: vi.fn(() => true),
    handleHandRaisedMessage: vi.fn(() => true),
    handleLockRoomMessage: vi.fn(() => true),
    handleTransferHostMessage: vi.fn(() => true),
    ...options.deps
  }

  return {
    ...deps,
    ...useMeetingSignalingRouter(deps)
  }
}

describe('useMeetingSignalingRouter', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('routes room users into peer connections and broadcasts local state', async () => {
    const router = createRouter()

    expect(await router.handleSignaling({
      type: 'room-users',
      data: [
        { userId: 2, userName: 'Alice' },
        { userId: 3, userName: 'Bob' }
      ]
    })).toBe(true)

    expect(router.waitingForHost.value).toBe(false)
    expect(router.ensurePeerEntry).toHaveBeenCalledWith(2, 'Alice')
    expect(router.ensurePeerEntry).toHaveBeenCalledWith(3, 'Bob')
    expect(router.createPeerConnection).toHaveBeenCalledWith(2, true)
    expect(router.createPeerConnection).toHaveBeenCalledWith(3, true)
    expect(router.broadcastSelfState).toHaveBeenCalled()
  })

  it('updates existing and new peer state messages', () => {
    const router = createRouter()

    expect(router.handlePeerStateMessage({
      type: 'peer-state',
      userId: 2,
      userName: 'Alice',
      data: { audioEnabled: false, screenShareEnabled: true, unsafeField: 'ignored' }
    })).toBe(true)

    expect(router.remotePeers[2]).toEqual(expect.objectContaining({
      name: 'Alice',
      audioEnabled: false,
      screenShareEnabled: true
    }))
    expect(router.remotePeers[2]).not.toHaveProperty('unsafeField')
    expect(router.updateAudioActivity).toHaveBeenCalledWith(2, false)

    expect(router.handlePeerStateMessage({
      type: 'peer-state',
      userId: 4,
      userName: 'Carol',
      data: { videoEnabled: false, unsafeField: 'ignored' }
    })).toBe(true)
    expect(router.remotePeers[4]).toEqual({
      name: 'Carol',
      videoEnabled: false
    })
  })

  it('delegates host, chat, hand, and WebRTC signaling messages', async () => {
    const router = createRouter()

    expect(await router.handleSignaling({ type: 'join-request', userId: 5 })).toBe(true)
    expect(router.handleJoinRequestMessage).toHaveBeenCalledWith({ type: 'join-request', userId: 5 })

    expect(await router.handleSignaling({ type: 'offer', userId: 2, data: 'offer-sdp' })).toBe(true)
    expect(await router.handleSignaling({ type: 'answer', userId: 2, data: 'answer-sdp' })).toBe(true)
    expect(await router.handleSignaling({ type: 'ice-candidate', userId: 2, data: { candidate: 'c' } })).toBe(true)
    expect(router.handleOffer).toHaveBeenCalledWith(2, 'offer-sdp')
    expect(router.handleAnswer).toHaveBeenCalledWith(2, 'answer-sdp')
    expect(router.handleIceCandidate).toHaveBeenCalledWith(2, { candidate: 'c' })

    expect(await router.handleSignaling({ type: 'chat', userId: 2, data: { text: 'hi' } })).toBe(true)
    expect(router.receiveChatMessage).toHaveBeenCalled()
    expect(await router.handleSignaling({ type: 'hand-raised', userId: 2, data: { raised: true } })).toBe(true)
    expect(router.handleHandRaisedMessage).toHaveBeenCalled()
    expect(await router.handleSignaling({ type: 'lock-room', data: { locked: true } })).toBe(true)
    expect(router.handleLockRoomMessage).toHaveBeenCalled()
    expect(await router.handleSignaling({ type: 'transfer-host', targetUserId: 2 })).toBe(true)
    expect(router.handleTransferHostMessage).toHaveBeenCalled()
  })

  it('handles self-targeted admission, rejection, kick, and mute messages with id coercion', async () => {
    const router = createRouter({ me: ref({ id: 1, name: '我' }) })

    expect(await router.handleSignaling({ type: 'admit-user', targetUserId: '1' })).toBe(true)
    expect(router.enterApprovedRoom).toHaveBeenCalled()

    expect(await router.handleSignaling({ type: 'reject-user', targetUserId: '1' })).toBe(true)
    expect(ElMessage.warning).toHaveBeenCalledWith('主持人已拒绝你的加入申请')
    expect(router.cancelWaitingForHost).toHaveBeenCalled()

    expect(await router.handleSignaling({ type: 'mute-target', data: { targetUserId: '1' } })).toBe(true)
    expect(router.toggleAudio).toHaveBeenCalledTimes(1)
    expect(ElMessage.info).toHaveBeenCalledWith('主持人已将你静音')

    router.audioEnabled.value = true
    expect(await router.handleSignaling({ type: 'mute-all', userId: '1' })).toBe(false)
    expect(router.toggleAudio).toHaveBeenCalledTimes(1)

    expect(await router.handleSignaling({ type: 'mute-all', userId: 2 })).toBe(true)
    expect(router.toggleAudio).toHaveBeenCalledTimes(2)
    expect(ElMessage.info).toHaveBeenCalledWith('主持人已全员静音')

    expect(await router.handleSignaling({ type: 'kick', data: { targetUserId: '1' } })).toBe(true)
    expect(ElMessage.warning).toHaveBeenCalledWith('你已被主持人移出会议')
    expect(router.leaveRoom).toHaveBeenCalled()
  })

  it('cleans up local meeting state when the host ends the meeting', async () => {
    const router = createRouter()

    expect(await router.handleSignaling({ type: 'meeting-ended' })).toBe(true)

    expect(ElMessage.warning).toHaveBeenCalledWith('主持人已结束会议')
    expect(router.cleanupRoomSession).toHaveBeenCalled()
    expect(router.cleanupRoomArtifacts).toHaveBeenCalled()
    expect(router.currentRoom.value).toBeNull()
    expect(router.hostUserId.value).toBeNull()
    expect(router.loadMyRooms).toHaveBeenCalled()
  })
})
