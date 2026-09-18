import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getReportList, processReport } from './report'
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

describe('report api', () => {
  it('getReportList GET /moderation/admin/report/list 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getReportList({ page: 1, status: 'pending' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/report/list',
      method: 'get',
      params: { page: 1, status: 'pending' },
    })
  })

  it('processReport PUT /moderation/admin/report/process/:id 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    const data = { action: 'resolve', remark: '已处理' }
    await processReport(11, data)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/report/process/11',
      method: 'put',
      data,
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getReportList({})).rejects.toThrow('500')
  })
})
