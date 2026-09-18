import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getAiUsageOverview, getAiUsageFeatures, getAiUsageDaily } from './aiUsage'
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

describe('aiUsage api', () => {
  it('getAiUsageOverview GET /ai/usage/overview', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { total: 100 } })
    await getAiUsageOverview()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/usage/overview',
      method: 'get',
    })
  })

  it('getAiUsageFeatures GET /ai/usage/features', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAiUsageFeatures()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/usage/features',
      method: 'get',
    })
  })

  it('getAiUsageDaily 默认 days=7', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAiUsageDaily()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/usage/daily',
      method: 'get',
      params: { days: 7 },
    })
  })

  it('getAiUsageDaily 自定义 days', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAiUsageDaily(30)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/usage/daily',
      method: 'get',
      params: { days: 30 },
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('boom'))
    await expect(getAiUsageOverview()).rejects.toThrow('boom')
  })
})
