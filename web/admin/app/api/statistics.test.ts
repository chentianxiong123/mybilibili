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
  getOverviewStatistics, getManuscriptStatusStatistics, getRecentManuscripts
} from './statistics'
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

describe('statistics api', () => {
  it('getOverviewStatistics GET /statistics/overview', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { users: 100 } })
    await getOverviewStatistics()
    expect(requestMock).toHaveBeenCalledWith({ url: '/statistics/overview', method: 'get' })
  })

  it('getManuscriptStatusStatistics GET /statistics/manuscript/status', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getManuscriptStatusStatistics()
    expect(requestMock).toHaveBeenCalledWith({ url: '/statistics/manuscript/status', method: 'get' })
  })

  it('getRecentManuscripts 默认 limit=10', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getRecentManuscripts()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/statistics/manuscript/recent',
      method: 'get',
      params: { limit: 10 },
    })
  })

  it('getRecentManuscripts 自定义 limit', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getRecentManuscripts(20)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/statistics/manuscript/recent',
      method: 'get',
      params: { limit: 20 },
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getOverviewStatistics()).rejects.toThrow('500')
  })
})