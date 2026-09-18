import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getAuditLogs, getAuditLogDetail } from './audit'
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

describe('audit api', () => {
  it('getAuditLogs GET /admin/audit-logs/list 带 params 透传', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { records: [], total: 0 } })
    const res = await getAuditLogs({ page: 1, size: 20 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/audit-logs/list',
      method: 'get',
      params: { page: 1, size: 20 },
    })
    expect(res).toEqual({ code: 200, data: { records: [], total: 0 } })
  })

  it('getAuditLogs 无 params 时仍能调用', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAuditLogs()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/audit-logs/list',
      method: 'get',
      params: undefined,
    })
  })

  it('getAuditLogDetail GET /admin/audit-logs/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 7, action: 'login' } })
    const res = await getAuditLogDetail(7)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/audit-logs/7',
      method: 'get',
    })
    expect(res.data.id).toBe(7)
  })

  it('错误响应透传', async () => {
    const err = new Error('network')
    requestMock.mockRejectedValue(err)
    await expect(getAuditLogs({ page: 1 })).rejects.toBe(err)
  })

  it('不同 id 的 URL 拼接正确', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await getAuditLogDetail(123)
    expect(requestMock).toHaveBeenCalledWith(expect.objectContaining({ url: '/admin/audit-logs/123' }))
  })
})
