import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { clearServerSession } from './session'

describe('clearServerSession', () => {
  const fetchMock = vi.fn(() => Promise.resolve({ ok: true }))

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.unstubAllGlobals()
  })

  it('POST /api/v1/admin/logout 携带 same-origin 凭证', async () => {
    await clearServerSession()
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/logout', {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' }
    })
  })

  it('支持自定义路径', async () => {
    await clearServerSession('/custom/logout')
    expect(fetchMock).toHaveBeenCalledWith('/custom/logout', expect.objectContaining({ method: 'POST' }))
  })

  it('网络失败静默吞掉（登出幂等）', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('fetch failed'))
    await expect(clearServerSession()).resolves.toBeUndefined()
  })

  it('服务端非 2xx 也静默（不抛错）', async () => {
    fetchMock.mockResolvedValueOnce({ ok: false, status: 500 })
    await expect(clearServerSession()).resolves.toBeUndefined()
  })

  it('无 fetch 环境直接返回（不抛错）', async () => {
    vi.stubGlobal('fetch', undefined)
    await expect(clearServerSession()).resolves.toBeUndefined()
  })
})