import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getConversations, getUnreadCounts, getNotifications } from '../message'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn() }
}))

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn()
}))

import api from '../client'
import { isLogin } from '../../utils/session'

const mockApi = vi.mocked(api)
const mockIsLogin = vi.mocked(isLogin)

describe('message API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockIsLogin.mockReturnValue(true)
  })

  describe('getConversations', () => {
    it('未登录 → 空数组', async () => {
      mockIsLogin.mockReturnValue(false)
      const result = await getConversations()
      expect(result).toEqual({ code: '0', data: [] })
      expect(mockApi.get).not.toHaveBeenCalled()
    })

    it('返回数据 snake_case → camelCase 换算', async () => {
      mockApi.get.mockResolvedValue({
        data: [
          {
            user_id: 1,
            target_user_id: 2,
            target_user_name: '张三',
            target_user_avatar: 'avatar.jpg',
            last_message_content: '你好',
            last_message_time: '2025-01-01',
            unread_count: 3
          }
        ]
      })
      const result = await getConversations()
      expect(result.code).toBe('1')
      const item = result.data[0]
      expect(item.userId).toBe(1)
      expect(item.targetUserId).toBe(2)
      expect(item.targetUserNickname).toBe('张三')
      expect(item.targetUserAvatar).toBe('avatar.jpg')
      expect(item.lastMessageContent).toBe('你好')
      expect(item.lastMessageTime).toBe('2025-01-01')
      expect(item.unreadCount).toBe(3)
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getConversations()
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getUnreadCounts', () => {
    it('未登录 → 空对象', async () => {
      mockIsLogin.mockReturnValue(false)
      const result = await getUnreadCounts()
      expect(result).toEqual({ code: '0', data: {} })
    })

    it('正常返回', async () => {
      mockApi.get.mockResolvedValue({ data: { chat: 5, reply: 2 } })
      const result = await getUnreadCounts()
      expect(result.code).toBe('1')
      expect(result.data).toEqual({ chat: 5, reply: 2 })
    })
  })

  describe('getNotifications', () => {
    it('未登录 → 空数组', async () => {
      mockIsLogin.mockReturnValue(false)
      const result = await getNotifications('reply')
      expect(result).toEqual({ code: '0', data: [] })
    })

    it('reply 类型 → /message/replies', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1 }] })
      await getNotifications('reply')
      expect(mockApi.get).toHaveBeenCalledWith('/message/replies')
    })

    it('at 类型 → /message/at', async () => {
      mockApi.get.mockResolvedValue({ data: [] })
      await getNotifications('at')
      expect(mockApi.get).toHaveBeenCalledWith('/message/at')
    })

    it('like 类型 → /message/likes', async () => {
      mockApi.get.mockResolvedValue({ data: [] })
      await getNotifications('like')
      expect(mockApi.get).toHaveBeenCalledWith('/message/likes')
    })

    it('system 类型 → /message/system', async () => {
      mockApi.get.mockResolvedValue({ data: [] })
      await getNotifications('system')
      expect(mockApi.get).toHaveBeenCalledWith('/message/system')
    })

    it('未知类型 → 空数组', async () => {
      const result = await getNotifications('unknown')
      expect(result).toEqual({ code: '0', data: [] })
    })
  })
})
