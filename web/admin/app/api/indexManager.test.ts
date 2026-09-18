import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { indexManagerApi } from './indexManager'
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

describe('indexManager api', () => {
  it('getStatus GET /search/admin/index/status', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { count: 10 } })
    const res = await indexManagerApi.getStatus()
    expect(requestMock.get).toHaveBeenCalledWith('/search/admin/index/status')
    expect(res.data.count).toBe(10)
  })

  it('validate POST /search/admin/index/validate', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await indexManagerApi.validate()
    expect(requestMock.post).toHaveBeenCalledWith('/search/admin/index/validate')
  })

  it('rebuild POST /search/admin/index/rebuild', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await indexManagerApi.rebuild()
    expect(requestMock.post).toHaveBeenCalledWith('/search/admin/index/rebuild')
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(indexManagerApi.getStatus()).rejects.toThrow('500')
  })
})
