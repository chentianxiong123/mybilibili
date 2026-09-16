import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'

vi.mock('@/utils/notification.ts', () => ({
  isNotificationEnabled: vi.fn(() => false),
  notifyMention: vi.fn(),
}))

import { isNotificationEnabled, notifyMention } from '@/utils/notification.ts'
import { useLiveRoomInteractions } from './useLiveRoomInteractions'

const isNotifEnabledMock = isNotificationEnabled as unknown as ReturnType<typeof vi.fn>
const notifyMock = notifyMention as unknown as ReturnType<typeof vi.fn>

function makeHarness(opts: any = {}) {
  const room = ref({ roomName: 'R', userId: 1 })
  const me = ref({ id: 99, name: 'me' })
  const currentUserId = ref(99)
  const sendRoomMessage = vi.fn(() => true)
  const emitDanmaku = vi.fn()
  const r = useLiveRoomInteractions({
    room, me, currentUserId,
    sendRoomMessage,
    emitDanmaku,
    ...opts,
  })
  return { ...r, room, me, currentUserId, sendRoomMessage, emitDanmaku }
}

describe('useLiveRoomInteractions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    isNotifEnabledMock.mockReturnValue(false)
  })

  describe('常量', () => {
    it('reactionEmojis / chatEmojis / giftOptions 都是数组', () => {
      const r = makeHarness()
      expect(Array.isArray(r.reactionEmojis)).toBe(true)
      expect(r.reactionEmojis.length).toBeGreaterThan(0)
      expect(Array.isArray(r.chatEmojis)).toBe(true)
      expect(Array.isArray(r.giftOptions)).toBe(true)
      expect(r.giftOptions.length).toBeGreaterThan(0)
    })
  })

  describe('pushDanmaku', () => {
    it('sendDanmaku 加入 danmakuList 并 emitDanmaku', () => {
      const r = makeHarness()
      const before = r.danmakuList.value.length
      r.danmakuText.value = 'hi'
      r.sendDanmaku()
      expect(r.danmakuList.value.length).toBe(before + 1)
      expect(r.danmakuList.value[before].text).toBe('hi')
      expect(r.emitDanmaku).toHaveBeenCalled()
    })

    it('> 200 条时触发 splice(0, 50)', () => {
      const r = makeHarness()
      for (let i = 0; i < 250; i++) {
        r.danmakuText.value = `m${i}`
        r.sendDanmaku()
      }
      expect(r.danmakuList.value.length).toBeLessThanOrEqual(200)
    })
  })

  describe('sendDanmaku', () => {
    it('空文本时不发送', () => {
      const r = makeHarness()
      r.sendDanmaku()
      expect(r.sendRoomMessage).not.toHaveBeenCalled()
    })

    it('非空文本时调用 sendRoomMessage 并 pushDanmaku', () => {
      const r = makeHarness()
      r.danmakuText.value = '  hello  '
      r.sendDanmaku()
      expect(r.sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'chat', data: expect.objectContaining({ text: 'hello' }) }))
      expect(r.danmakuText.value).toBe('')
    })
  })

  describe('sendReaction', () => {
    it('调用 sendRoomMessage 并加入 reactions', () => {
      const r = makeHarness()
      r.sendReaction('❤️')
      expect(r.sendRoomMessage).toHaveBeenCalled()
      expect(r.reactions.value.length).toBeGreaterThanOrEqual(1)
    })

    it('sendRoomMessage 返回 false 时不入 reactions', () => {
      const r = makeHarness({ sendRoomMessage: vi.fn(() => false) })
      r.sendReaction('👍')
      expect(r.reactions.value.length).toBe(0)
    })
  })

  describe('addReaction / receiveReaction', () => {
    it('receiveReaction 推送带 x, y, from 的反应', () => {
      const r = makeHarness()
      r.receiveReaction({ data: { emoji: '❤️', x: 50, y: 80 }, userName: 'bob', userId: 1 })
      expect(r.reactions.value.length).toBe(1)
    })

    it('receiveReaction 无 emoji 时忽略', () => {
      const r = makeHarness()
      r.receiveReaction({ data: {} })
      expect(r.reactions.value.length).toBe(0)
    })

    it('reactions 超过 20 时 shift 最早一条', () => {
      const r = makeHarness()
      for (let i = 0; i < 25; i++) {
        r.receiveReaction({ data: { emoji: '👍' }, userId: i })
      }
      expect(r.reactions.value.length).toBeLessThanOrEqual(20)
    })
  })

  describe('gift 相关', () => {
    it('sendGift 调用 sendRoomMessage 并 pushGift', () => {
      const r = makeHarness()
      r.sendGift({ emoji: '👏', name: '鼓掌', cost: 10 })
      expect(r.sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'gift' }))
      expect(r.showGiftPicker.value).toBe(false)
    })

    it('sendGift 失败时不入 gifts', () => {
      const r = makeHarness({ sendRoomMessage: vi.fn(() => false) })
      r.sendGift({ emoji: '👏', name: '鼓掌', cost: 10 })
      expect(r.gifts.value.length).toBe(0)
    })

    it('receiveGift 无 emoji 时忽略', () => {
      const r = makeHarness()
      r.receiveGift({ data: {} })
      expect(r.gifts.value.length).toBe(0)
    })

    it('receiveGift 正常时入 gifts', () => {
      const r = makeHarness()
      r.receiveGift({ data: { emoji: '🌹', name: '鲜花' }, userName: 'alice' })
      expect(r.gifts.value.length).toBe(1)
    })
  })

  describe('pin / unpin', () => {
    it('pinMessage 调用 sendRoomMessage 并设置 pinnedMessage', () => {
      const r = makeHarness()
      r.pinMessage({ text: 'pinned text', from: 'host' })
      expect(r.sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'pin-message' }))
      expect(r.pinnedMessage.value).toEqual(expect.objectContaining({ text: 'pinned text' }))
    })

    it('unpinMessage 清空 pinnedMessage', () => {
      const r = makeHarness()
      r.pinMessage({ text: 'x' })
      r.unpinMessage()
      expect(r.pinnedMessage.value).toBe(null)
    })

    it('receivePinMessage 无 text 忽略', () => {
      const r = makeHarness()
      r.receivePinMessage({ data: {} })
      expect(r.pinnedMessage.value).toBe(null)
    })

    it('receivePinMessage 正常时设值', () => {
      const r = makeHarness()
      r.receivePinMessage({ data: { text: 'p' }, userName: 'h' })
      expect(r.pinnedMessage.value).toEqual(expect.objectContaining({ text: 'p', from: 'h' }))
    })

    it('clearPinnedMessage 清空', () => {
      const r = makeHarness()
      r.receivePinMessage({ data: { text: 'x' } })
      r.clearPinnedMessage()
      expect(r.pinnedMessage.value).toBe(null)
    })
  })

  describe('receiveDanmaku + 通知', () => {
    it('空 text 时返回 null', () => {
      const r = makeHarness()
      expect(r.receiveDanmaku({ data: {}, userId: 1 })).toBe(null)
    })

    it('非空时加入 danmakuList', () => {
      const r = makeHarness()
      const item = r.receiveDanmaku({ data: { text: 'hi' }, userId: 1, userName: 'alice' })
      expect(item).toBeTruthy()
      expect(item.text).toBe('hi')
    })

    it('提到 me 且 enabled 时 notifyMention 被调用', () => {
      isNotifEnabledMock.mockReturnValue(true)
      const r = makeHarness()
      r.receiveDanmaku({ data: { text: '@me hi' }, userId: 1, userName: 'alice' })
      expect(notifyMock).toHaveBeenCalledWith('alice', 'R', '@me hi')
    })

    it('提到 me 但未启用通知时不 notify', () => {
      isNotifEnabledMock.mockReturnValue(false)
      const r = makeHarness()
      r.receiveDanmaku({ data: { text: '@me hi' }, userId: 1, userName: 'alice' })
      expect(notifyMock).not.toHaveBeenCalled()
    })

    it('自己发的不通知', () => {
      isNotifEnabledMock.mockReturnValue(true)
      const r = makeHarness()
      r.receiveDanmaku({ data: { text: '@me hi' }, userId: 99, userName: 'me' })
      expect(notifyMock).not.toHaveBeenCalled()
    })
  })

  describe('emoji picker', () => {
    it('toggleEmojiPicker 翻转', () => {
      const r = makeHarness()
      expect(r.showEmojiPicker.value).toBe(false)
      r.toggleEmojiPicker()
      expect(r.showEmojiPicker.value).toBe(true)
    })

    it('insertChatEmoji 追加 + 关闭 picker', () => {
      const r = makeHarness()
      r.toggleEmojiPicker()
      r.insertChatEmoji('😀')
      expect(r.danmakuText.value).toContain('😀')
      expect(r.showEmojiPicker.value).toBe(false)
    })
  })

  describe('cleanup', () => {
    it('cleanupLiveRoomInteractions 清空所有状态', () => {
      const r = makeHarness()
      r.sendGift({ emoji: '👏', name: '鼓掌', cost: 10 })
      r.sendReaction('❤️')
      r.showEmojiPicker.value = true
      r.showGiftPicker.value = true
      r.cleanupLiveRoomInteractions()
      expect(r.reactions.value).toEqual([])
      expect(r.gifts.value).toEqual([])
      expect(r.showEmojiPicker.value).toBe(false)
      expect(r.showGiftPicker.value).toBe(false)
    })
  })

  describe('自定义定时器', () => {
    it('setTimeoutFn / clearTimeoutFn 被正确调用（scheduleCleanup）', () => {
      const timers = new Set<any>()
      const setTimeoutFn = vi.fn((cb: () => void, delay: number) => {
        const id = `t-${timers.size}`
        timers.add({ id, cb, delay })
        return id
      })
      const clearTimeoutFn = vi.fn((id: any) => {
        for (const t of timers) if (t.id === id) timers.delete(t)
      })
      const r = makeHarness({ setTimeoutFn, clearTimeoutFn })
      r.sendReaction('❤️')
      expect(setTimeoutFn).toHaveBeenCalled()
      // 至少有一个 timer 注册
      expect(timers.size).toBeGreaterThan(0)
      r.cleanupLiveRoomInteractions()
      expect(clearTimeoutFn).toHaveBeenCalled()
    })
  })
})