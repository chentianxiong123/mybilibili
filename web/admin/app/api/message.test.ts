import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { messageApi } from './message'
import { safeStorage } from '../utils/safeStorage'
import request from '@/api/client'

const requestMock = request as unknown as ReturnType<typeof vi.fn> & {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

const store = new Map<string, string>()
const localStorageMock = {
  getItem: (k: string) => (store.has(k) ? store.get(k) : null),
  setItem: (k: string, v: string) => { store.set(k, String(v)) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
} as Storage
if (typeof window !== 'undefined') (window as any).localStorage = localStorageMock
;(globalThis as any).localStorage = localStorageMock

beforeEach(() => {
  store.clear()
  requestMock.mockReset()
  requestMock.get.mockReset()
  requestMock.post.mockReset()
  requestMock.put.mockReset()
  requestMock.delete.mockReset()
  safeStorage.setItem('token', 'tok-1')
})

describe('message api - 已登录场景', () => {
  it('getConversations GET /message/conversations', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getConversations()
    expect(requestMock.get).toHaveBeenCalledWith('/message/conversations')
  })

  it('getConversationDetail GET /message/conversations/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await messageApi.getConversationDetail(5)
    expect(requestMock.get).toHaveBeenCalledWith('/message/conversations/5')
  })

  it('deleteConversation DELETE /message/conversations/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await messageApi.deleteConversation(5)
    expect(requestMock.delete).toHaveBeenCalledWith('/message/conversations/5')
  })

  it('getMessages GET /:id/messages 默认 page=1/size=20', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getMessages(5)
    expect(requestMock.get).toHaveBeenCalledWith('/message/conversations/5/messages', { params: { page: 1, size: 20 } })
  })

  it('getMessages 自定义 page/size', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getMessages(5, 3, 50)
    expect(requestMock.get).toHaveBeenCalledWith('/message/conversations/5/messages', { params: { page: 3, size: 50 } })
  })

  it('sendMessage POST /message/send 透传 data', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await messageApi.sendMessage({ to: 'u2', content: 'hi' })
    expect(requestMock.post).toHaveBeenCalledWith('/message/send', { to: 'u2', content: 'hi' })
  })

  it('markAsRead PUT /message/:id/read', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await messageApi.markAsRead(99)
    expect(requestMock.put).toHaveBeenCalledWith('/message/99/read')
  })

  it('batchMarkAsRead PUT /message/batch/read 透传 ids', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await messageApi.batchMarkAsRead([1, 2, 3])
    expect(requestMock.put).toHaveBeenCalledWith('/message/batch/read', { ids: [1, 2, 3] })
  })

  it('getUnreadCounts GET /message/unread/counts', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: {} })
    await messageApi.getUnreadCounts()
    expect(requestMock.get).toHaveBeenCalledWith('/message/unread/counts')
  })

  it('getMessageSettings GET /message/settings', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await messageApi.getMessageSettings()
    expect(requestMock.get).toHaveBeenCalledWith('/message/settings')
  })

  it('updateMessageSettings PUT /message/settings 透传 data', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await messageApi.updateMessageSettings({ replyEnabled: false })
    expect(requestMock.put).toHaveBeenCalledWith('/message/settings', { replyEnabled: false })
  })

  it('deleteMessage DELETE /message/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await messageApi.deleteMessage(11)
    expect(requestMock.delete).toHaveBeenCalledWith('/message/11')
  })

  it('getReplies GET /message/replies 带 params', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getReplies({ page: 1 })
    expect(requestMock.get).toHaveBeenCalledWith('/message/replies', { params: { page: 1 } })
  })

  it('getLikes GET /message/likes 带 params', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getLikes({ page: 1 })
    expect(requestMock.get).toHaveBeenCalledWith('/message/likes', { params: { page: 1 } })
  })

  it('getSystemNotifications GET /message/system 带 params', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await messageApi.getSystemNotifications({ page: 1 })
    expect(requestMock.get).toHaveBeenCalledWith('/message/system', { params: { page: 1 } })
  })

  it('broadcastSystemNotification POST /message/admin/system/broadcast 透传 data', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await messageApi.broadcastSystemNotification({ title: '通知', content: '内容' })
    expect(requestMock.post).toHaveBeenCalledWith('/message/admin/system/broadcast', { title: '通知', content: '内容' })
  })
})

describe('message api - 未登录场景 (authRequired 返回 401)', () => {
  beforeEach(() => {
    store.delete('token')
  })

  it('getConversations 未登录返回 code=401 不调用 http', async () => {
    const res = await messageApi.getConversations()
    expect(res.code).toBe(401)
    expect(res.message).toBe('请先登录')
    expect(requestMock.get).not.toHaveBeenCalled()
  })

  it('getMessages 未登录直接返回 401 envelope 不发请求', async () => {
    const res = await messageApi.getMessages(5)
    expect(res.code).toBe(401)
    expect(requestMock.get).not.toHaveBeenCalled()
  })

  it('sendMessage 未登录返回 401', async () => {
    const res = await messageApi.sendMessage({ to: 'u' })
    expect(res.code).toBe(401)
    expect(requestMock.post).not.toHaveBeenCalled()
  })

  it('getUnreadCounts 未登录返回默认 0', async () => {
    const res = await messageApi.getUnreadCounts()
    expect(res.code).toBe(401)
    expect(requestMock.get).not.toHaveBeenCalled()
  })

  it('broadcastSystemNotification 不受 authRequired 守卫（管理员操作）', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await messageApi.broadcastSystemNotification({ title: 't' })
    expect(requestMock.post).toHaveBeenCalledWith('/message/admin/system/broadcast', { title: 't' })
  })
})
