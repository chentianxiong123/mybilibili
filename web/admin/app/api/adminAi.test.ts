import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { adminAiApi } from './adminAi'

const store = new Map<string, string>()
const localStorageMock = {
  getItem: (k: string) => (store.has(k) ? store.get(k) : null),
  setItem: (k: string, v: string) => { store.set(k, String(v)) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
} as Storage
if (typeof window !== 'undefined') (window as any).localStorage = localStorageMock
;(globalThis as any).localStorage = localStorageMock

const fetchMock = vi.fn()
;(globalThis as any).fetch = fetchMock

beforeEach(() => {
  store.clear()
  fetchMock.mockReset()
})

describe('adminAiApi.sendMessage', () => {
  it('POST /ai/assistant/send 带 content 与 auth 头', async () => {
    store.set('admin_token', 'tok-1')
    store.set('admin_id', '42')
    fetchMock.mockResolvedValue({
      ok: true,
      headers: { get: (k: string) => k === 'content-type' ? 'application/json' : null },
      text: async () => JSON.stringify({ event: 'done', data: 'ok' }),
    })

    const onDone = vi.fn()
    adminAiApi.sendMessage('hello', { onDone })

    await new Promise(r => setTimeout(r, 10))
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/ai/assistant/send',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
          'Authorization': 'Bearer tok-1',
        }),
        body: JSON.stringify({ content: 'hello' }),
      }),
    )
  })

  it('非 ok 状态调用 onError', async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 500, headers: { get: () => null } })
    const onError = vi.fn()
    adminAiApi.sendMessage('hi', { onError })
    await new Promise(r => setTimeout(r, 10))
    expect(onError).toHaveBeenCalled()
  })

  it('JSON 响应 event=stream 拼接 parts 后调用 onDone', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      text: async () => JSON.stringify({ event: 'stream', parts: ['a', 'b', 'c'] }),
    })
    const onData = vi.fn()
    const onDone = vi.fn()
    adminAiApi.sendMessage('x', { onData, onDone })
    await new Promise(r => setTimeout(r, 10))
    expect(onData).toHaveBeenCalledTimes(3)
    expect(onData).toHaveBeenNthCalledWith(1, 'a')
    expect(onData).toHaveBeenNthCalledWith(2, 'b')
    expect(onData).toHaveBeenNthCalledWith(3, 'c')
    expect(onDone).toHaveBeenCalledWith('abc')
  })

  it('JSON 解析失败调用 onError 响应解析失败', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      text: async () => 'not json',
    })
    const onError = vi.fn()
    adminAiApi.sendMessage('x', { onError })
    await new Promise(r => setTimeout(r, 10))
    expect(onError).toHaveBeenCalledWith('响应解析失败')
  })

  it('abort 调用 controller.abort', () => {
    fetchMock.mockResolvedValue({ ok: true, headers: { get: () => 'application/json' }, text: async () => '{}' })
    const handle = adminAiApi.sendMessage('hi')
    expect(() => handle.abort()).not.toThrow()
  })

  it('无 token 时 Authorization 为空字符串', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      text: async () => JSON.stringify({ event: 'done' }),
    })
    adminAiApi.sendMessage('x')
    await new Promise(r => setTimeout(r, 10))
    const headers = fetchMock.mock.calls[0][1].headers
    expect(headers['Authorization']).toBe('')
    // 管理员身份由后端从验签 token 推导，客户端不得自塞 X-Admin-Id
    expect(headers['X-Admin-Id']).toBeUndefined()
  })
})
