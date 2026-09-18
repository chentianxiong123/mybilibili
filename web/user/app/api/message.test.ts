import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
  apiPut: vi.fn(),
  apiDelete: vi.fn(),
  safeStorage: {
    getItem: vi.fn(),
  },
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: mocks.apiPut,
    delete: mocks.apiDelete,
  }),
}))

vi.mock('@/utils/safeStorage', () => ({
  safeStorage: mocks.safeStorage,
}))

import { messageApi } from './message'

describe('messageApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('未登录时 getConversations 直接返回 401', async () => {
    mocks.safeStorage.getItem.mockReturnValue(null)
    const res = await messageApi.getConversations()
    expect(res).toEqual({ code: 401, message: '请先登录', data: [] })
    expect(mocks.apiGet).not.toHaveBeenCalled()
  })

  it('已登录时 getConversations 调 /message/conversations', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await messageApi.getConversations()
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/conversations')
  })

  it('未登录时 getConversationDetail 返回 null data', async () => {
    mocks.safeStorage.getItem.mockReturnValue(null)
    const res = await messageApi.getConversationDetail(1)
    expect(res).toEqual({ code: 401, message: '请先登录', data: null })
  })

  it('已登录时 getConversationDetail', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await messageApi.getConversationDetail(2)
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/conversations/2')
  })

  it('deleteConversation 登录态', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await messageApi.deleteConversation(3)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/message/conversations/3')
  })

  it('getMessages 默认分页', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await messageApi.getMessages(1)
    expect(mocks.apiGet).toHaveBeenCalledWith(
      '/message/conversations/1/messages',
      { params: { page: 1, size: 20 } },
    )
  })

  it('sendMessage 调 /message/send', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await messageApi.sendMessage({ conversationId: 1, content: 'hi' })
    expect(mocks.apiPost).toHaveBeenCalledWith('/message/send', { conversationId: 1, content: 'hi' })
  })

  it('markAsRead / batchMarkAsRead', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiPut.mockResolvedValue({ code: 200 })
    await messageApi.markAsRead(5)
    await messageApi.batchMarkAsRead([1, 2, 3])
    expect(mocks.apiPut).toHaveBeenNthCalledWith(1, '/message/5/read')
    expect(mocks.apiPut).toHaveBeenNthCalledWith(2, '/message/batch/read', { ids: [1, 2, 3] })
  })

  it('getUnreadCounts 已登录', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await messageApi.getUnreadCounts()
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/unread/counts')
  })

  it('getUnreadCounts 未登录', async () => {
    mocks.safeStorage.getItem.mockReturnValue(null)
    const res = await messageApi.getUnreadCounts()
    expect(res.data).toEqual({ private: 0, reply: 0, at: 0, like: 0, system: 0, dynamic: 0 })
  })

  it('getMessageSettings / updateMessageSettings', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await messageApi.getMessageSettings()
    await messageApi.updateMessageSettings({ muted: true })
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/settings')
    expect(mocks.apiPut).toHaveBeenCalledWith('/message/settings', { muted: true })
  })

  it('deleteMessage', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await messageApi.deleteMessage(7)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/message/7')
  })

  it('getReplies / getLikes / getSystemNotifications 传 params', async () => {
    mocks.safeStorage.getItem.mockReturnValue('t')
    mocks.apiGet.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getReplies({ page: 1 })
    await messageApi.getLikes({ page: 1 })
    await messageApi.getSystemNotifications({ page: 1 })
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/replies', { params: { page: 1 } })
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/likes', { params: { page: 1 } })
    expect(mocks.apiGet).toHaveBeenCalledWith('/message/system', { params: { page: 1 } })
  })

  it('broadcastSystemNotification 不需要登录', async () => {
    mocks.safeStorage.getItem.mockReturnValue(null)
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await messageApi.broadcastSystemNotification({ content: 'hi' })
    expect(mocks.apiPost).toHaveBeenCalledWith('/message/admin/system/broadcast', { content: 'hi' })
  })
})
