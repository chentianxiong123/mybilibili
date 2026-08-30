import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useLiveRoomRealtime } from '../composables/useLiveRoomRealtime.js'

const createRealtime = (options = {}) => {
  const room = ref(options.room || { streamKey: 'live-100' })
  const me = ref(options.me || { id: 7, name: '观众7' })
  const currentUserId = ref(options.currentUserId ?? 7)
  let wsOptions = null
  const ws = {
    send: vi.fn(() => true),
    close: vi.fn(),
    isOpen: vi.fn(() => true)
  }
  const createWs = vi.fn(options => {
    wsOptions = options
    return ws
  })

  const realtime = useLiveRoomRealtime({
    room,
    me,
    currentUserId,
    createWs: options.createWs || createWs,
    getWsUrl: options.getWsUrl || (() => 'ws://live.example/ws'),
    logger: options.logger || { log: vi.fn() }
  })

  return {
    room,
    me,
    currentUserId,
    ws,
    createWs,
    getWsOptions: () => wsOptions,
    ...realtime
  }
}

describe('useLiveRoomRealtime', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('connects the room socket, sends join on open, and wraps outgoing room messages', () => {
    const realtime = createRealtime()

    expect(realtime.sendRoomMessage({ type: 'chat' })).toBe(false)
    expect(realtime.connectLiveRoomSocket()).toBe(true)

    expect(realtime.createWs).toHaveBeenCalled()
    expect(realtime.getWsOptions().url()).toBe('ws://live.example/ws')

    realtime.getWsOptions().onOpen()
    expect(realtime.ws.send).toHaveBeenCalledWith({
      type: 'join',
      roomCode: 'live-100',
      userId: 7,
      userName: '观众7',
      targetUserId: undefined,
      data: undefined
    })

    expect(realtime.sendRoomMessage({
      type: 'chat',
      targetUserId: 100,
      data: { text: 'hi' }
    })).toBe(true)
    expect(realtime.ws.send).toHaveBeenCalledWith({
      type: 'chat',
      roomCode: 'live-100',
      userId: 7,
      userName: '观众7',
      targetUserId: 100,
      data: { text: 'hi' }
    })

    expect(realtime.cleanupLiveRoomRealtime()).toBe(true)
    expect(realtime.ws.close).toHaveBeenCalled()
  })

  it('routes realtime room messages to attached live handlers', async () => {
    const realtime = createRealtime()
    const handlers = {
      handleLinkmicApplyMessage: vi.fn(),
      handleLinkmicAcceptedMessage: vi.fn(async () => true),
      handleLinkmicRejectedMessage: vi.fn(),
      handleLinkmicDisconnectedMessage: vi.fn(),
      handleHandRaisedMessage: vi.fn(),
      receiveGift: vi.fn(),
      handleOffer: vi.fn(async () => true),
      handleAnswer: vi.fn(async () => true),
      handleIceCandidate: vi.fn(async () => true),
      receiveReaction: vi.fn(),
      receivePinMessage: vi.fn(),
      clearPinnedMessage: vi.fn(),
      receiveDanmaku: vi.fn(),
      handleMuteAudioMessage: vi.fn(async () => true),
      handleMuteVideoMessage: vi.fn(async () => true),
      handlePeerStateMessage: vi.fn()
    }
    realtime.attachRoomMessageHandlers(handlers)

    expect(await realtime.handleRoomMessage({ type: 'linkmic-apply', userId: 1 })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'linkmic-accepted', userId: 2 })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'linkmic-rejected' })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'linkmic-disconnected', userId: 2 })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'hand-raised', userId: 2 })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'gift', data: { emoji: '👏' } })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'offer', userId: 2, data: 'offer-sdp' })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'answer', userId: 2, data: 'answer-sdp' })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'ice-candidate', userId: 2, data: { candidate: 'c' } })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'reaction', data: { emoji: '👍' } })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'pin-message', data: { text: '置顶' } })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'unpin-message' })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'chat', data: { text: '弹幕' } })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'mute-target', targetUserId: 7 })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'mute-video-target', targetUserId: 7 })).toBe(true)
    expect(await realtime.handleRoomMessage({ type: 'peer-state', userId: 2 })).toBe(true)

    expect(handlers.handleLinkmicApplyMessage).toHaveBeenCalledWith({ type: 'linkmic-apply', userId: 1 })
    expect(handlers.handleOffer).toHaveBeenCalledWith(2, 'offer-sdp')
    expect(handlers.handleAnswer).toHaveBeenCalledWith(2, 'answer-sdp')
    expect(handlers.handleIceCandidate).toHaveBeenCalledWith(2, { candidate: 'c' })
    expect(handlers.clearPinnedMessage).toHaveBeenCalled()
  })

  it('updates viewer count and ignores unknown or incomplete messages', async () => {
    const realtime = createRealtime()

    expect(await realtime.handleRoomMessage({ type: 'viewer-count', data: { count: 42 } })).toBe(true)
    expect(realtime.realtimeViewerCount.value).toBe(42)
    expect(await realtime.handleRoomMessage({ type: 'viewer-count', data: {} })).toBe(false)
    expect(await realtime.handleRoomMessage({ type: 'unknown' })).toBe(false)
  })
})
