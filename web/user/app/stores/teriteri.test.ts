import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useTeriteriStore } from './teriteri'

vi.mock('element-plus', () => ({
  ElMessage: { error: vi.fn(), success: vi.fn() }
}))

vi.mock('axios', () => ({
  default: { get: vi.fn().mockResolvedValue({ data: { code: 200, data: {} } }) }
}))

vi.mock('@/teriteri-src/network/request', () => ({
  get: vi.fn().mockResolvedValue({ data: { data: { reply: 0, at: 0, love: 0, system: 0, whisper: 0, dynamic: 0 } } })
}))

describe('useTeriteriStore', () => {
  let store: ReturnType<typeof useTeriteriStore>

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useTeriteriStore()
  })

  describe('initData', () => {
    it('resets state to defaults', () => {
      store.isLogin = true
      store.user = { uid: 1 }
      store.msgUnread = [5, 4, 3, 2, 1, 0]
      store.attitudeToVideo = { 1: 1 }
      store.favorites = [1]
      store.likeComment = [1]
      store.dislikeComment = [1]

      store.initData()

      expect(store.isLogin).toBe(false)
      expect(store.user).toEqual({})
      expect(store.msgUnread).toEqual([0, 0, 0, 0, 0, 0])
      expect(store.attitudeToVideo).toEqual({})
      expect(store.favorites).toEqual([])
      expect(store.likeComment).toEqual([])
      expect(store.dislikeComment).toEqual([])
    })
  })

  describe('simple setters', () => {
    it('updateIsLogin', () => {
      store.updateIsLogin(true)
      expect(store.isLogin).toBe(true)
    })

    it('updateUser', () => {
      const u = { uid: 1, name: 'test' }
      store.updateUser(u)
      expect(store.user).toStrictEqual(u)
    })

    it('updateChannels', () => {
      const c = [1, 2, 3]
      store.updateChannels(c)
      expect(store.channels).toStrictEqual(c)
    })

    it('updateCarousels', () => {
      const c = ['a', 'b']
      store.updateCarousels(c)
      expect(store.carousels).toStrictEqual(c)
    })

    it('updateAttitudeToVideo', () => {
      const a = { 1: -1 }
      store.updateAttitudeToVideo(a)
      expect(store.attitudeToVideo).toStrictEqual(a)
    })

    it('updateLikeComment', () => {
      store.updateLikeComment([1, 2])
      expect(store.likeComment).toEqual([1, 2])
    })

    it('updateDislikeComment', () => {
      store.updateDislikeComment([3])
      expect(store.dislikeComment).toEqual([3])
    })

    it('updateFavorites', () => {
      store.updateFavorites([10, 20])
      expect(store.favorites).toEqual([10, 20])
    })

    it('updateUserFavList defaults to [] if falsy', () => {
      store.updateUserFavList(null as any)
      expect(store.userFavList).toEqual([])
    })

    it('updateUserFavList sets value', () => {
      store.updateUserFavList([1, 2])
      expect(store.userFavList).toEqual([1, 2])
    })

    it('updateTrendings', () => {
      store.updateTrendings([1])
      expect(store.trendings).toEqual([1])
    })

    it('updateMatchingCount', () => {
      store.updateMatchingCount([3, 4])
      expect(store.matchingCount).toEqual([3, 4])
    })
  })

  describe('updateChatList', () => {
    it('pushes items to chatList', () => {
      store.updateChatList([{ chat: { id: 1 } }])
      expect(store.chatList).toHaveLength(1)
      store.updateChatList([{ chat: { id: 2 } }])
      expect(store.chatList).toHaveLength(2)
    })
  })

  describe('handleWsClose', () => {
    it('resets login-related state', () => {
      store.isLogin = true
      store.user = { uid: 1 }
      store.msgUnread = [5, 4, 3, 2, 1, 0]
      store.attitudeToVideo = { 1: 1 }
      store.favorites = [1]
      store.likeComment = [1]
      store.dislikeComment = [1]

      store.handleWsClose()

      expect(store.isLogin).toBe(false)
      expect(store.user).toEqual({})
      expect(store.msgUnread).toEqual([0, 0, 0, 0, 0, 0])
      expect(store.attitudeToVideo).toEqual({})
      expect(store.favorites).toEqual([])
      expect(store.likeComment).toEqual([])
      expect(store.dislikeComment).toEqual([])
    })
  })

  describe('handleWsMessage', () => {
    const makeEvent = (data: any) => ({ data: JSON.stringify(data) }) as MessageEvent

    describe('type error', () => {
      it('calls initData and removes token on login expired', async () => {
        const { ElMessage } = await import('element-plus')
        store.isLogin = true
        store.user = { uid: 1 }
        const removeItemSpy = vi.spyOn(Storage.prototype, 'removeItem')

        store.handleWsMessage(makeEvent({ type: 'error', data: '登录已过期' }))

        expect(store.isLogin).toBe(false)
        expect(store.user).toEqual({})
        expect(removeItemSpy).toHaveBeenCalledWith('teri_token')
        expect(ElMessage.error).toHaveBeenCalledWith('登录已过期')
        removeItemSpy.mockRestore()
      })

      it('shows error for other messages', async () => {
        const { ElMessage } = await import('element-plus')
        store.handleWsMessage(makeEvent({ type: 'error', data: 'something wrong' }))
        expect(ElMessage.error).toHaveBeenCalledWith('something wrong')
      })
    })

    describe('type reply', () => {
      it('接收 increments msgUnread[0]', () => {
        store.handleWsMessage(makeEvent({ type: 'reply', data: { type: '接收' } }))
        expect(store.msgUnread[0]).toBe(1)
        store.handleWsMessage(makeEvent({ type: 'reply', data: { type: '接收' } }))
        expect(store.msgUnread[0]).toBe(2)
      })

      it('全部已读 resets msgUnread[0]', () => {
        store.msgUnread[0] = 5
        store.handleWsMessage(makeEvent({ type: 'reply', data: { type: '全部已读' } }))
        expect(store.msgUnread[0]).toBe(0)
      })
    })

    describe('type at', () => {
      it('接收 increments msgUnread[1]', () => {
        store.handleWsMessage(makeEvent({ type: 'at', data: { type: '接收' } }))
        expect(store.msgUnread[1]).toBe(1)
      })

      it('全部已读 resets msgUnread[1]', () => {
        store.msgUnread[1] = 3
        store.handleWsMessage(makeEvent({ type: 'at', data: { type: '全部已读' } }))
        expect(store.msgUnread[1]).toBe(0)
      })
    })

    describe('type love', () => {
      it('接收 increments msgUnread[2]', () => {
        store.handleWsMessage(makeEvent({ type: 'love', data: { type: '接收' } }))
        expect(store.msgUnread[2]).toBe(1)
      })

      it('全部已读 resets msgUnread[2]', () => {
        store.msgUnread[2] = 2
        store.handleWsMessage(makeEvent({ type: 'love', data: { type: '全部已读' } }))
        expect(store.msgUnread[2]).toBe(0)
      })
    })

    describe('type system', () => {
      it('接收 increments msgUnread[3]', () => {
        store.handleWsMessage(makeEvent({ type: 'system', data: { type: '接收' } }))
        expect(store.msgUnread[3]).toBe(1)
      })

      it('全部已读 resets msgUnread[3]', () => {
        store.msgUnread[3] = 4
        store.handleWsMessage(makeEvent({ type: 'system', data: { type: '全部已读' } }))
        expect(store.msgUnread[3]).toBe(0)
      })
    })

    describe('type dynamic', () => {
      it('接收 increments msgUnread[5]', () => {
        store.handleWsMessage(makeEvent({ type: 'dynamic', content: { type: '接收' } }))
        expect(store.msgUnread[5]).toBe(1)
      })

      it('全部已读 resets msgUnread[5]', () => {
        store.msgUnread[5] = 6
        store.handleWsMessage(makeEvent({ type: 'dynamic', content: { type: '全部已读' } }))
        expect(store.msgUnread[5]).toBe(0)
      })
    })

    describe('type whisper', () => {
      it('全部已读 resets msgUnread[4] and all chat unread', () => {
        store.chatList = [
          { chat: { id: 1, unread: 3 }, user: { uid: 10 }, detail: { more: true, list: [] } },
          { chat: { id: 2, unread: 5 }, user: { uid: 20 }, detail: { more: true, list: [] } }
        ]
        store.msgUnread[4] = 8
        store.handleWsMessage(makeEvent({ type: 'whisper', data: { type: '全部已读' } }))
        expect(store.msgUnread[4]).toBe(0)
        expect(store.chatList[0].chat.unread).toBe(0)
        expect(store.chatList[1].chat.unread).toBe(0)
      })

      it('已读 reduces msgUnread[4] and marks specific chat as read', () => {
        store.chatList = [
          { chat: { id: 1, unread: 3 }, user: { uid: 10 }, detail: { more: true, list: [] } }
        ]
        store.msgUnread[4] = 5
        store.handleWsMessage(makeEvent({ type: 'whisper', data: { type: '已读', id: 1, count: 3 } }))
        expect(store.msgUnread[4]).toBe(2)
        expect(store.chatList[0].chat.unread).toBe(0)
      })

      it('已读 does not go below 0', () => {
        store.msgUnread[4] = 1
        store.handleWsMessage(makeEvent({ type: 'whisper', data: { type: '已读', id: 999, count: 5 } }))
        expect(store.msgUnread[4]).toBe(0)
      })

      it('移除 deletes chat and resets chatId if matching', () => {
        store.chatList = [
          { chat: { id: 1, unread: 2 }, user: { uid: 10 }, detail: { more: true, list: [] } }
        ]
        store.chatId = 10
        store.msgUnread[4] = 3
        store.handleWsMessage(makeEvent({ type: 'whisper', data: { type: '移除', id: 1, count: 2 } }))
        expect(store.chatList).toHaveLength(0)
        expect(store.chatId).toBe(-1)
        expect(store.msgUnread[4]).toBe(1)
      })

      it('移除 does not reset chatId if user uid does not match', () => {
        store.chatList = [
          { chat: { id: 1, unread: 2 }, user: { uid: 10 }, detail: { more: true, list: [] } }
        ]
        store.chatId = 99
        store.handleWsMessage(makeEvent({ type: 'whisper', data: { type: '移除', id: 1, count: 2 } }))
        expect(store.chatId).toBe(99)
      })

      it('移除 does nothing if chat not found', () => {
        store.chatList = []
        store.handleWsMessage(makeEvent({ type: 'whisper', data: { type: '移除', id: 999, count: 1 } }))
        expect(store.chatList).toHaveLength(0)
      })

      it('接收 adds new chat from other user', () => {
        store.user = { uid: 1 }
        const detail = { userId: 2, anotherId: 1, id: 100 }
        const chat = { id: 10, userId: 2, latestTime: '2024-01-02T00:00:00Z' }
        const user = { uid: 2, name: 'other' }
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '接收', chat, detail, user, online: false }
        }))
        expect(store.msgUnread[4]).toBe(1)
        expect(store.chatList).toHaveLength(1)
        expect(store.chatList[0].detail.list).toHaveLength(1)
      })

      it('接收 from other user does not increment unread if online', () => {
        store.user = { uid: 1 }
        const detail = { userId: 2, anotherId: 1, id: 100 }
        const chat = { id: 10, userId: 2, latestTime: '2024-01-02T00:00:00Z' }
        const user = { uid: 2, name: 'other' }
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '接收', chat, detail, user, online: true }
        }))
        expect(store.msgUnread[4]).toBe(0)
      })

      it('接收 from self updates existing chat', () => {
        store.user = { uid: 1 }
        store.isChatPage = true
        store.chatList = [
          {
            chat: { id: 10, userId: 2, latestTime: '2024-01-01T00:00:00Z' },
            user: { uid: 2 },
            detail: { more: true, list: [] }
          }
        ]
        const detail = { userId: 1, anotherId: 2, id: 200 }
        const chat = { id: 10, userId: 2, latestTime: '2024-01-03T00:00:00Z' }
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '接收', chat, detail, user: { uid: 1 }, online: true }
        }))
        expect(store.chatList[0].detail.list).toHaveLength(1)
        expect(store.chatList[0].chat.latestTime).toBe('2024-01-03T00:00:00Z')
      })

      it('接收 from self does not push if not on chat page', () => {
        store.user = { uid: 1 }
        store.isChatPage = false
        store.chatList = [
          {
            chat: { id: 10, userId: 2, latestTime: '2024-01-01T00:00:00Z' },
            user: { uid: 2 },
            detail: { more: true, list: [] }
          }
        ]
        const detail = { userId: 1, anotherId: 2, id: 200 }
        const chat = { id: 10, userId: 2, latestTime: '2024-01-03T00:00:00Z' }
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '接收', chat, detail, user: { uid: 1 }, online: true }
        }))
        expect(store.chatList[0].detail.list).toHaveLength(0)
      })

      it('撤回 marks message as withdrawn (sender side)', () => {
        store.user = { uid: 1 }
        store.chatList = [
          {
            chat: { id: 10, userId: 2 },
            user: { uid: 2 },
            detail: { more: true, list: [{ id: 100, withdraw: 0 }] }
          }
        ]
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '撤回', id: 100, sendId: 1, acceptId: 2 }
        }))
        expect(store.chatList[0].detail.list[0].withdraw).toBe(1)
      })

      it('撤回 marks message as withdrawn (receiver side)', () => {
        store.user = { uid: 1 }
        store.chatList = [
          {
            chat: { id: 10, userId: 2 },
            user: { uid: 2 },
            detail: { more: true, list: [{ id: 200, withdraw: 0 }] }
          }
        ]
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '撤回', id: 200, sendId: 2, acceptId: 1 }
        }))
        expect(store.chatList[0].detail.list[0].withdraw).toBe(1)
      })

      it('撤回 does nothing if chat not found', () => {
        store.user = { uid: 1 }
        store.chatList = []
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '撤回', id: 100, sendId: 2, acceptId: 1 }
        }))
        expect(store.chatList).toHaveLength(0)
      })

      it('撤回 does nothing if message not found', () => {
        store.user = { uid: 1 }
        store.chatList = [
          {
            chat: { id: 10, userId: 2 },
            user: { uid: 2 },
            detail: { more: true, list: [{ id: 999, withdraw: 0 }] }
          }
        ]
        store.handleWsMessage(makeEvent({
          type: 'whisper',
          data: { type: '撤回', id: 100, sendId: 2, acceptId: 1 }
        }))
        expect(store.chatList[0].detail.list[0].withdraw).toBe(0)
      })
    })
  })
})
