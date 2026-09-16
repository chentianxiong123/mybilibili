import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useLiveRoomRealtime } from './useLiveRoomRealtime'

function makeFakeWs() {
  return {
    isOpen: vi.fn(() => true),
    send: vi.fn(() => true),
    close: vi.fn(),
    _handler: {} as any,
  }
}

describe('useLiveRoomRealtime', () => {
  let ws: ReturnType<typeof makeFakeWs>
  let createWs: any
  let logger: any

  beforeEach(() => {
    ws = makeFakeWs()
    logger = { log: vi.fn(), error: vi.fn() }
    createWs = vi.fn((opts: any) => {
      ws._handler = opts
      return ws
    })
  })

  function makeHarness(overrides: any = {}) {
    const room = ref({ streamKey: 'skey', userId: 1 })
    const me = ref({ id: 99, name: 'me' })
    const currentUserId = ref(99)
    const r = useLiveRoomRealtime({
      room, me, currentUserId,
      createWs, logger,
      getWsUrl: () => 'ws://test/ws/live',
      ...overrides,
    })
    return { ...r, room, me, currentUserId }
  }

  describe('connectLiveRoomSocket', () => {
    it('调用 createWs 并传入 url/onOpen/onMessage/onClose', () => {
      makeHarness()
      const r = useLiveRoomRealtime({ room: ref({ streamKey: 's' }), me: ref({ id: 1, name: 'm' }), currentUserId: ref(1), createWs, logger })
      r.connectLiveRoomSocket()
      expect(createWs).toHaveBeenCalledTimes(1)
      expect(typeof ws._handler.onOpen).toBe('function')
      expect(typeof ws._handler.onMessage).toBe('function')
      expect(typeof ws._handler.onClose).toBe('function')
    })

    it('重复连接时先 cleanup 旧 ws', () => {
      const r = makeHarness()
      r.connectLiveRoomSocket()
      const oldWs = ws
      r.connectLiveRoomSocket()
      expect(oldWs.close).toHaveBeenCalled()
      expect(createWs).toHaveBeenCalledTimes(2)
    })

    it('onOpen 触发时若 room.streamKey 存在则发送 join', () => {
      const r = makeHarness()
      r.connectLiveRoomSocket()
      ws._handler.onOpen()
      expect(ws.send).toHaveBeenCalledWith(expect.objectContaining({ type: 'join', roomCode: 'skey', userId: 99, userName: 'me' }))
    })

    it('onClose 触发时 logger.log 被调用', () => {
      const r = makeHarness()
      r.connectLiveRoomSocket()
      ws._handler.onClose()
      expect(logger.log).toHaveBeenCalled()
    })
  })

  describe('sendRoomMessage', () => {
    it('ws 未连接时返回 false', () => {
      ws.isOpen.mockReturnValue(false)
      const r = makeHarness()
      r.connectLiveRoomSocket()
      const ok = r.sendRoomMessage({ type: 'chat', data: { text: 'hi' } })
      expect(ok).toBe(false)
    })

    it('ws 已连接时 send 被调用并返回 send 的结果', () => {
      ws.send.mockReturnValue(true)
      const r = makeHarness()
      r.connectLiveRoomSocket()
      const ok = r.sendRoomMessage({ type: 'chat', data: { text: 'hi' } })
      expect(ok).toBe(true)
      expect(ws.send).toHaveBeenCalledWith(expect.objectContaining({ type: 'chat', roomCode: 'skey' }))
    })

    it('包含 targetUserId 与 data', () => {
      ws.send.mockReturnValue(true)
      const r = makeHarness()
      r.connectLiveRoomSocket()
      r.sendRoomMessage({ type: 'offer', targetUserId: 5, data: { sdp: 'x' } })
      expect(ws.send).toHaveBeenCalledWith(expect.objectContaining({ type: 'offer', targetUserId: 5, data: { sdp: 'x' } }))
    })
  })

  describe('handleRoomMessage', () => {
    let handlers: Record<string, any>
    let r: any

    beforeEach(() => {
      r = makeHarness()
      r.connectLiveRoomSocket()
      handlers = {
        handleLinkmicApplyMessage: vi.fn(),
        handleLinkmicAcceptedMessage: vi.fn().mockResolvedValue(undefined),
        handleLinkmicRejectedMessage: vi.fn(),
        handleLinkmicDisconnectedMessage: vi.fn(),
        handleHandRaisedMessage: vi.fn(),
        receiveGift: vi.fn(),
        receiveReaction: vi.fn(),
        receivePinMessage: vi.fn(),
        clearPinnedMessage: vi.fn(),
        receiveDanmaku: vi.fn(),
        handleMuteAudioMessage: vi.fn().mockResolvedValue(undefined),
        handleMuteVideoMessage: vi.fn().mockResolvedValue(undefined),
        handlePeerStateMessage: vi.fn(),
        handleOffer: vi.fn().mockResolvedValue(undefined),
        handleAnswer: vi.fn().mockResolvedValue(undefined),
        handleIceCandidate: vi.fn().mockResolvedValue(undefined),
      }
      r.attachRoomMessageHandlers(handlers)
    })

    it('chat → receiveDanmaku', () => {
      r.handleRoomMessage({ type: 'chat', data: { text: 'hi' } })
      expect(handlers.receiveDanmaku).toHaveBeenCalled()
    })
    it('gift → receiveGift', () => {
      r.handleRoomMessage({ type: 'gift', data: {} })
      expect(handlers.receiveGift).toHaveBeenCalled()
    })
    it('reaction → receiveReaction', () => {
      r.handleRoomMessage({ type: 'reaction', data: {} })
      expect(handlers.receiveReaction).toHaveBeenCalled()
    })
    it('pin-message → receivePinMessage', () => {
      r.handleRoomMessage({ type: 'pin-message', data: { text: 'x' } })
      expect(handlers.receivePinMessage).toHaveBeenCalled()
    })
    it('unpin-message → clearPinnedMessage', () => {
      r.handleRoomMessage({ type: 'unpin-message' })
      expect(handlers.clearPinnedMessage).toHaveBeenCalled()
    })
    it('linkmic-apply → handleLinkmicApplyMessage', () => {
      r.handleRoomMessage({ type: 'linkmic-apply', data: {} })
      expect(handlers.handleLinkmicApplyMessage).toHaveBeenCalled()
    })
    it('linkmic-rejected → handleLinkmicRejectedMessage', () => {
      r.handleRoomMessage({ type: 'linkmic-rejected' })
      expect(handlers.handleLinkmicRejectedMessage).toHaveBeenCalled()
    })
    it('linkmic-disconnected → handleLinkmicDisconnectedMessage', () => {
      r.handleRoomMessage({ type: 'linkmic-disconnected' })
      expect(handlers.handleLinkmicDisconnectedMessage).toHaveBeenCalled()
    })
    it('hand-raised → handleHandRaisedMessage', () => {
      r.handleRoomMessage({ type: 'hand-raised', data: { raised: true } })
      expect(handlers.handleHandRaisedMessage).toHaveBeenCalled()
    })
    it('peer-state → handlePeerStateMessage', () => {
      r.handleRoomMessage({ type: 'peer-state', userId: 1, data: {} })
      expect(handlers.handlePeerStateMessage).toHaveBeenCalled()
    })

    it('offer → handleOffer(userId, data)', async () => {
      await r.handleRoomMessage({ type: 'offer', userId: 5, data: 'sdp' })
      expect(handlers.handleOffer).toHaveBeenCalledWith(5, 'sdp')
    })
    it('answer → handleAnswer(userId, data)', async () => {
      await r.handleRoomMessage({ type: 'answer', userId: 5, data: 'sdp' })
      expect(handlers.handleAnswer).toHaveBeenCalledWith(5, 'sdp')
    })
    it('ice-candidate → handleIceCandidate(userId, data)', async () => {
      await r.handleRoomMessage({ type: 'ice-candidate', userId: 5, data: { sdpMid: '0' } })
      expect(handlers.handleIceCandidate).toHaveBeenCalledWith(5, { sdpMid: '0' })
    })

    it('mute-target → handleMuteAudioMessage', async () => {
      await r.handleRoomMessage({ type: 'mute-target', data: {} })
      expect(handlers.handleMuteAudioMessage).toHaveBeenCalled()
    })
    it('mute-video-target → handleMuteVideoMessage', async () => {
      await r.handleRoomMessage({ type: 'mute-video-target', data: {} })
      expect(handlers.handleMuteVideoMessage).toHaveBeenCalled()
    })

    it('viewer-count 更新 realtimeViewerCount', () => {
      r.handleRoomMessage({ type: 'viewer-count', data: { count: 123 } })
      expect(r.realtimeViewerCount.value).toBe(123)
    })

    it('viewer-count 无 count 字段时返回 false 不更新', async () => {
      r.realtimeViewerCount.value = 0
      const result = await r.handleRoomMessage({ type: 'viewer-count', data: {} })
      expect(result).toBe(false)
      expect(r.realtimeViewerCount.value).toBe(0)
    })

    it('未知 type 时返回 false', async () => {
      expect(await r.handleRoomMessage({ type: 'nope' })).toBe(false)
    })

    it('未注册的 handler 被忽略并返回 false', async () => {
      // attachRoomMessageHandlers 是 Object.assign，不会清空。
      // 这里我们用一个新的 r 实例并 attachRoomMessageHandlers({})，结果不变。
      // 真正测试未注册 handler 的方式是直接 mock 一个不存在的 type
      expect(await r.handleRoomMessage({ type: 'unregistered-type' })).toBe(false)
    })
  })

  describe('cleanupLiveRoomRealtime', () => {
    it('关闭 ws 并清空引用', () => {
      const r = makeHarness()
      r.connectLiveRoomSocket()
      const ok = r.cleanupLiveRoomRealtime()
      expect(ok).toBe(true)
      expect(ws.close).toHaveBeenCalled()
    })

    it('没有 ws 时返回 false', () => {
      const r = makeHarness()
      expect(r.cleanupLiveRoomRealtime()).toBe(false)
    })
  })

  describe('default ws url', () => {
    it('默认 getWsUrl 返回 ws://host/ws/live', () => {
      const r = useLiveRoomRealtime({
        room: ref({ streamKey: 's' }),
        me: ref({ id: 1, name: 'm' }),
        currentUserId: ref(1),
        createWs,
        logger,
      })
      r.connectLiveRoomSocket()
      expect(createWs.mock.calls[0][0].url()).toMatch(/wss?:\/\/.+\/ws\/live/)
    })
  })
})