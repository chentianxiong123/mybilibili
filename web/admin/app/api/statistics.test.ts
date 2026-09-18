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
  getOverviewStatistics, getManuscriptStatusStatistics, getRecentManuscripts,
  getVideoPlayStatistics, getUserGrowthStatistics, getCommentStatistics, getHotVideos
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

  it('getVideoPlayStatistics GET 带 startDate/endDate', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getVideoPlayStatistics('2024-01-01', '2024-01-31')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/statistics/video/play',
      method: 'get',
      params: { startDate: '2024-01-01', endDate: '2024-01-31' },
    })
  })

  it('getUserGrowthStatistics GET 带 startDate/endDate', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getUserGrowthStatistics('2024-01-01', '2024-01-31')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/statistics/user/growth',
      method: 'get',
      params: { startDate: '2024-01-01', endDate: '2024-01-31' },
    })
  })

  it('getCommentStatistics GET 带 startDate/endDate', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getCommentStatistics('2024-01-01', '2024-01-31')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/statistics/comment',
      method: 'get',
      params: { startDate: '2024-01-01', endDate: '2024-01-31' },
    })
  })

  it('getHotVideos 默认 limit=10', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getHotVideos()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/statistics/video/hot',
      method: 'get',
      params: { limit: 10 },
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getOverviewStatistics()).rejects.toThrow('500')
  })
})
