import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import {
  getPendingSessions, getSessionMessages, sendReply, resolveSession, getPendingCount
} from './customerSession'
import request from '@/api/client'

const requestMock = request as unknown as ReturnType<typeof vi.fn> & {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  requestMock.mockReset()
  requestMock.get.mockReset()
  requestMock.post.mockReset()
  requestMock.put.mockReset()
  requestMock.delete.mockReset()
})

describe('customerSession api', () => {
  it('getPendingSessions GET /ai/customer/sessions', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getPendingSessions()
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/customer/sessions', method: 'get' })
  })

  it('getSessionMessages GET /ai/admin/customer/sessions/:id/messages', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getSessionMessages(7)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/customer/sessions/7/messages',
      method: 'get',
    })
  })

  it('sendReply POST /ai/admin/customer/sessions/:id/reply 透传 adminId/content', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await sendReply(7, 100, '你好')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/customer/sessions/7/reply',
      method: 'post',
      data: { adminId: 100, content: '你好' },
    })
  })

  it('resolveSession POST /ai/admin/customer/sessions/:id/resolve', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await resolveSession(8)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/customer/sessions/8/resolve',
      method: 'post',
    })
  })

  it('getPendingCount GET /ai/customer/sessions/pending/count', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { count: 3 } })
    const res = await getPendingCount()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/customer/sessions/pending/count',
      method: 'get',
    })
    expect(res.data.count).toBe(3)
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('401'))
    await expect(getPendingSessions()).rejects.toThrow('401')
  })
})
