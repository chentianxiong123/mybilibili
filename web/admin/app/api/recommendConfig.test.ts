import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { recommendConfigApi } from './recommendConfig'
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

describe('recommendConfig api', () => {
  it('getConfig GET /search/admin/recommend-config', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { topN: 10 } })
    const res = await recommendConfigApi.getConfig()
    expect(requestMock.get).toHaveBeenCalledWith('/search/admin/recommend-config')
    expect(res.data.topN).toBe(10)
  })

  it('updateConfig PUT /search/admin/recommend-config 透传 data', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await recommendConfigApi.updateConfig({ topN: 20 })
    expect(requestMock.put).toHaveBeenCalledWith('/search/admin/recommend-config', { topN: 20 })
  })

  it('resetConfig POST /search/admin/recommend-config/reset', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await recommendConfigApi.resetConfig()
    expect(requestMock.post).toHaveBeenCalledWith('/search/admin/recommend-config/reset')
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(recommendConfigApi.getConfig()).rejects.toThrow('500')
  })
})
