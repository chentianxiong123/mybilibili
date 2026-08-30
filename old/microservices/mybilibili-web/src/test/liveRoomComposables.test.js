import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { linkmicApi } from '../api/linkmic.js'
import { isNotificationEnabled, notifyMention } from '../utils/notification.js'
import { useLiveLinkmic } from '../composables/useLiveLinkmic.js'
import { useLiveRoomInteractions } from '../composables/useLiveRoomInteractions.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

vi.mock('../api/linkmic.js', () => ({
  linkmicApi: {
    applyLinkmic: vi.fn(),
    acceptLinkmic: vi.fn(),
    rejectLinkmic: vi.fn(),
    disconnectLinkmic: vi.fn(),
    toggleAudio: vi.fn(),
    toggleVideo: vi.fn(),
    getActiveLinkmics: vi.fn(),
    getPendingApplications: vi.fn(),
    getQueuePosition: vi.fn()
  }
}))

vi.mock('../utils/notification.js', () => ({
  isNotificationEnabled: vi.fn(),
  notifyMention: vi.fn()
}))

const createInteractions = (options = {}) => {
  const room = ref({ roomName: '测试直播间' })
  const me = ref({ id: 7, name: '主播A' })
  const currentUserId = ref(7)
  const sendRoomMessage = options.sendRoomMessage || vi.fn(() => true)
  const emitDanmaku = options.emitDanmaku || vi.fn()

  const composable = useLiveRoomInteractions({
    room,
    me,
    currentUserId,
    sendRoomMessage,
    emitDanmaku,
    ...options
  })

  return {
    room,
    me,
    currentUserId,
    sendRoomMessage,
    emitDanmaku,
    ...composable
  }
}

const createLinkmic = (options = {}) => {
  const room = ref({ userId: 100, streamKey: 'live-100' })
  const isStreamer = ref(false)
  const currentUserId = ref(7)
  const sendRoomMessage = options.sendRoomMessage || vi.fn(() => true)
  const logger = options.logger || { error: vi.fn() }

  const composable = useLiveLinkmic({
    roomId: 'room-1',
    room,
    isStreamer,
    currentUserId,
    sendRoomMessage,
    logger,
    ...options
  })

  return {
    room,
    isStreamer,
    currentUserId,
    sendRoomMessage,
    logger,
    ...composable
  }
}

describe('useLiveRoomInteractions', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    isNotificationEnabled.mockReturnValue(false)
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('receives danmaku, emits it to the player, and sends mention notifications', () => {
    isNotificationEnabled.mockReturnValue(true)
    const live = createInteractions()

    const item = live.receiveDanmaku({
      type: 'chat',
      userId: 22,
      userName: '观众22',
      data: {
        text: '@主播A 来了',
        color: '#fad34a'
      }
    })

    expect(item.text).toBe('@主播A 来了')
    expect(live.danmakuList.value).toHaveLength(1)
    expect(live.emitDanmaku).toHaveBeenCalledWith('@主播A 来了', '#fad34a')
    expect(notifyMention).toHaveBeenCalledWith('观众22', '测试直播间', '@主播A 来了')

    expect(live.receiveDanmaku({
      userId: '7',
      userName: '主播A',
      data: { text: '@主播A 自己发的' }
    })).toEqual(expect.objectContaining({ text: '@主播A 自己发的' }))
    expect(notifyMention).toHaveBeenCalledTimes(1)

    expect(live.receiveDanmaku({
      userId: 23,
      data: { text: 123 }
    })).toBeNull()
    expect(live.danmakuList.value).toHaveLength(2)
  })

  it('keeps local danmaku visible even when websocket send fails', () => {
    const live = createInteractions({
      sendRoomMessage: vi.fn(() => false)
    })

    live.danmakuText.value = '离线弹幕'
    live.sendDanmaku()

    expect(live.sendRoomMessage).toHaveBeenCalledWith({
      type: 'chat',
      data: expect.objectContaining({ text: '离线弹幕' })
    })
    expect(live.danmakuList.value).toEqual([
      expect.objectContaining({ text: '离线弹幕', from: '主播A' })
    ])
    expect(live.danmakuText.value).toBe('')
  })

  it('only shows a sent reaction after the websocket accepts the message', () => {
    const failed = createInteractions({
      sendRoomMessage: vi.fn(() => false)
    })

    failed.sendReaction('👍')
    expect(failed.reactions.value).toHaveLength(0)

    const live = createInteractions()
    live.sendReaction('👍')

    expect(live.sendRoomMessage).toHaveBeenCalledWith({
      type: 'reaction',
      data: expect.objectContaining({ emoji: '👍' })
    })
    expect(live.reactions.value).toEqual([
      expect.objectContaining({ emoji: '👍', from: '主播A' })
    ])
  })

  it('sends gifts, pins messages, and mirrors successful local state', () => {
    const live = createInteractions()
    live.showGiftPicker.value = true

    live.sendGift({ emoji: '👏', name: '鼓掌', cost: 10 })
    live.pinMessage({ text: '重要消息', from: '观众' })
    live.unpinMessage()

    expect(live.sendRoomMessage).toHaveBeenNthCalledWith(1, {
      type: 'gift',
      data: { emoji: '👏', name: '鼓掌', cost: 10 }
    })
    expect(live.gifts.value).toEqual([
      expect.objectContaining({ emoji: '👏', name: '鼓掌', from: '主播A' })
    ])
    expect(live.showGiftPicker.value).toBe(false)
    expect(ElMessage.success).toHaveBeenCalledWith('赠送了 鼓掌')
    expect(live.sendRoomMessage).toHaveBeenNthCalledWith(2, {
      type: 'pin-message',
      data: { text: '重要消息' }
    })
    expect(live.sendRoomMessage).toHaveBeenNthCalledWith(3, { type: 'unpin-message' })
    expect(live.pinnedMessage.value).toBeNull()
  })

  it('cleans up delayed reaction and gift timers', () => {
    const live = createInteractions()

    live.sendReaction('👍')
    live.sendGift({ emoji: '👏', name: '鼓掌', cost: 10 })
    expect(vi.getTimerCount()).toBeGreaterThan(0)

    live.cleanupLiveRoomInteractions()
    expect(live.reactions.value).toEqual([])
    expect(live.gifts.value).toEqual([])
    expect(vi.getTimerCount()).toBe(0)

    vi.advanceTimersByTime(3000)
    expect(live.reactions.value).toEqual([])
    expect(live.gifts.value).toEqual([])
  })
})

describe('useLiveLinkmic', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads active linkmics and streamer-only pending applications', async () => {
    linkmicApi.getActiveLinkmics.mockResolvedValue({
      code: 200,
      data: [{ id: 1, viewerId: 10 }]
    })
    linkmicApi.getPendingApplications.mockResolvedValue({
      code: 200,
      data: [{ id: 2, viewerId: 11 }]
    })
    const live = createLinkmic()
    live.isStreamer.value = true

    await live.loadLinkmicStatus()

    expect(linkmicApi.getActiveLinkmics).toHaveBeenCalledWith('room-1')
    expect(linkmicApi.getPendingApplications).toHaveBeenCalledWith('room-1')
    expect(live.activeLinkmics.value).toEqual([{ id: 1, viewerId: 10 }])
    expect(live.pendingApplications.value).toEqual([{ id: 2, viewerId: 11 }])
  })

  it('logs load failures without breaking the caller', async () => {
    const logger = { error: vi.fn() }
    linkmicApi.getActiveLinkmics.mockRejectedValue(new Error('network down'))
    const live = createLinkmic({ logger })

    await expect(live.loadLinkmicStatus()).resolves.toBeUndefined()

    expect(logger.error).toHaveBeenCalledWith('加载连麦状态失败:', expect.any(Error))
  })

  it('deduplicates incoming applications and marks hand raises for streamers', () => {
    const live = createLinkmic()
    live.isStreamer.value = true

    const msg = {
      userId: 11,
      userName: '观众11',
      data: { id: 2, viewerId: 11, viewerName: '观众11' }
    }

    live.handleLinkmicApplyMessage(msg)
    live.handleLinkmicApplyMessage(msg)
    live.handleHandRaisedMessage({
      userId: 11,
      userName: '观众11',
      data: { raised: true }
    })

    expect(live.pendingApplications.value).toHaveLength(1)
    expect(live.pendingApplications.value[0].handRaised).toBe(true)
    expect(ElMessage.info).toHaveBeenCalledWith('观众11 举手了')
  })

  it('accepts a linkmic request and notifies the target viewer', async () => {
    linkmicApi.acceptLinkmic.mockResolvedValue({ code: 200 })
    const live = createLinkmic()
    const linkmic = { id: 3, viewerId: 12, viewerName: '观众12', status: 0 }
    live.pendingApplications.value = [linkmic]

    await live.acceptLinkmic(linkmic)

    expect(linkmicApi.acceptLinkmic).toHaveBeenCalledWith(3)
    expect(linkmic.status).toBe(1)
    expect(live.activeLinkmics.value).toEqual([linkmic])
    expect(live.pendingApplications.value).toEqual([])
    expect(live.sendRoomMessage).toHaveBeenCalledWith({
      type: 'linkmic-accepted',
      targetUserId: 12,
      data: { linkmicId: 3 }
    })
  })

  it('resets viewer state after rejection and remote disconnect', () => {
    const live = createLinkmic()
    live.myLinkmicId.value = 4
    live.myQueuePosition.value = 2
    live.viewerHandRaised.value = true

    live.handleLinkmicRejectedMessage()

    expect(live.myLinkmicId.value).toBeNull()
    expect(live.myQueuePosition.value).toBe(0)
    expect(live.viewerHandRaised.value).toBe(false)

    live.myLinkmicId.value = 5
    live.myQueuePosition.value = 1
    live.viewerHandRaised.value = true
    live.handleLinkmicDisconnectedMessage({
      userId: 100,
      data: { viewerId: 7 }
    })

    expect(live.myLinkmicId.value).toBeNull()
    expect(live.myQueuePosition.value).toBe(0)
    expect(live.viewerHandRaised.value).toBe(false)
    expect(ElMessage.info).toHaveBeenCalledWith('连麦已断开')
  })

  it('merges only supported peer state fields and accepts string target ids for mute messages', async () => {
    const live = createLinkmic({
      currentUserId: ref(7)
    })

    expect(live.handlePeerStateMessage({
      userId: 12,
      userName: '观众12',
      data: {
        audioEnabled: false,
        videoEnabled: true,
        unsafeField: 'ignored'
      }
    })).toBe(true)
    expect(live.linkmicStreams[12]).toEqual({
      name: '观众12',
      stream: null,
      audioEnabled: false,
      videoEnabled: true
    })
    expect(live.linkmicStreams[12]).not.toHaveProperty('unsafeField')

    await expect(live.handleMuteAudioMessage({
      targetUserId: '7'
    })).resolves.toBe(true)
    expect(ElMessage.info).toHaveBeenCalledWith('主播已将你静音')
  })

  it('toggles hand raise only when the room message is sent', () => {
    const blocked = createLinkmic({
      sendRoomMessage: vi.fn(() => false)
    })

    blocked.toggleViewerHandRaise()
    expect(blocked.viewerHandRaised.value).toBe(false)

    const live = createLinkmic()
    live.toggleViewerHandRaise()

    expect(live.viewerHandRaised.value).toBe(true)
    expect(live.sendRoomMessage).toHaveBeenCalledWith({
      type: 'hand-raised',
      targetUserId: 100,
      data: { raised: true }
    })
  })
})
