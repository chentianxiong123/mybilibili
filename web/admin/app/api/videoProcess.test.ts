import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

// happy-dom 18 未把 localStorage 挂到 window，手动 polyfill（与 auth.test.ts 保持一致）
const store = new Map<string, string>()
const localStorageMock = {
  getItem: (k: string) => (store.has(k) ? store.get(k) : null),
  setItem: (k: string, v: string) => { store.set(k, String(v)) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
}
if (typeof window !== 'undefined') (window as any).localStorage = localStorageMock
;(globalThis as any).localStorage = localStorageMock

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
  store.clear()
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

  it('getStreamUrl 无凭证时保持纯路径', () => {
    expect(getStreamUrl()).toBe('/api/v1/video/process/admin/stream')
  })

  it('getStreamUrl 用 access_token 携带凭证（EventSource 无法设置请求头）', () => {
    localStorageMock.setItem('admin_token', 'tok-abc')
    expect(getStreamUrl()).toBe('/api/v1/video/process/admin/stream?access_token=tok-abc')
  })

  it('getStreamUrl 对特殊字符做 URL 编码', () => {
    localStorageMock.setItem('admin_token', 'a+b/c=d')
    expect(getStreamUrl()).toBe('/api/v1/video/process/admin/stream?access_token=a%2Bb%2Fc%3Dd')
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(getCurrentTask()).rejects.toThrow('500')
  })
})
