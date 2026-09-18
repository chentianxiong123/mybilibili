import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getCurrentTask, getQueueInfo, getStatistics, getStreamUrl } from './videoProcess'
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

describe('videoProcess api', () => {
  it('getCurrentTask GET /video/process/admin/current', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await getCurrentTask()
    expect(requestMock.get).toHaveBeenCalledWith('/video/process/admin/current')
  })

  it('getQueueInfo GET /video/process/admin/queue', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { size: 0 } })
    await getQueueInfo()
    expect(requestMock.get).toHaveBeenCalledWith('/video/process/admin/queue')
  })

  it('getStatistics GET /video/process/admin/statistics', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { total: 100 } })
    await getStatistics()
    expect(requestMock.get).toHaveBeenCalledWith('/video/process/admin/statistics')
  })

  it('getStreamUrl 返回 SSE 路径', () => {
    expect(getStreamUrl()).toBe('/api/v1/video/process/admin/stream')
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(getCurrentTask()).rejects.toThrow('500')
  })
})
