import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  safeStorage: {
    getItem: vi.fn(),
  },
  fetch: vi.fn(),
}))

vi.mock('@/utils/safeStorage', () => ({
  safeStorage: mocks.safeStorage,
}))

global.fetch = mocks.fetch

import { adminAiApi } from './adminAi'

function buildSSEResponse(events: string[]) {
  const encoder = new TextEncoder()
  const stream = new ReadableStream({
    start(controller) {
      for (const e of events) {
        controller.enqueue(encoder.encode(e))
      }
      controller.close()
    },
  })
  return {
    ok: true,
    body: stream,
    headers: new Headers({ 'content-type': 'text/event-stream' }),
  } as unknown as Response
}

describe('adminAiApi.sendMessage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.safeStorage.getItem.mockReturnValue(null)
  })

  it('凭证走 HttpOnly cookie，不设 Authorization，也不自塞身份头', async () => {
    mocks.fetch.mockResolvedValueOnce(buildSSEResponse(['']))
    const handlers = { onData: vi.fn(), onDone: vi.fn(), onError: vi.fn(), onToolCall: vi.fn() }
    await adminAiApi.sendMessage('hello', handlers)
    expect(mocks.fetch).toHaveBeenCalled()
    const opts = mocks.fetch.mock.calls[0][1] as RequestInit
    expect(opts.method).toBe('POST')
    const headers = opts.headers as Record<string, string>
    // 能被 JS 读出来塞进 Authorization 的东西，XSS 也能读——必须为空
    expect(headers['Authorization']).toBeUndefined()
    expect(headers['X-Admin-Id']).toBeUndefined()
    expect(headers['X-User-Id']).toBeUndefined()
    expect(opts.body).toBe(JSON.stringify({ content: 'hello' }))
  })

  it('即使 localStorage 有 admin_token 也不带进请求头', async () => {
    mocks.safeStorage.getItem.mockImplementation((k: string) => {
      if (k === 'admin_token') return 'admin_t'
      return null
    })
    mocks.fetch.mockResolvedValueOnce(buildSSEResponse(['']))
    await adminAiApi.sendMessage('hi')
    const opts = mocks.fetch.mock.calls[0][1] as RequestInit
    const headers = opts.headers as Record<string, string>
    expect(headers['Authorization']).toBeUndefined()
    expect(headers['X-Admin-Id']).toBeUndefined()
  })

  it('解析 SSE event: / data: 并触发 onData', async () => {
    mocks.fetch.mockResolvedValueOnce(
      buildSSEResponse(['event:data\ndata:hello\n\n']),
    )
    const handlers = { onData: vi.fn(), onDone: vi.fn() }
    await adminAiApi.sendMessage('x', handlers)
    await new Promise(r => setTimeout(r, 0))
    await new Promise(r => setTimeout(r, 0))
    expect(handlers.onData).toHaveBeenCalledWith('hello')
  })

  it('解析 SSE tool_call', async () => {
    mocks.fetch.mockResolvedValueOnce(
      buildSSEResponse(['event:tool_call\ndata:{"name":"a"}\n\n']),
    )
    const handlers = { onToolCall: vi.fn() }
    await adminAiApi.sendMessage('x', handlers)
    await new Promise(r => setTimeout(r, 0))
    expect(handlers.onToolCall).toHaveBeenCalledWith('{"name":"a"}')
  })

  it('解析 SSE done JSON 触发 onDone 解析', async () => {
    mocks.fetch.mockResolvedValueOnce(
      buildSSEResponse(['event:done\ndata:{"ok":1}\n\n']),
    )
    const handlers = { onDone: vi.fn() }
    await adminAiApi.sendMessage('x', handlers)
    await new Promise(r => setTimeout(r, 0))
    expect(handlers.onDone).toHaveBeenCalledWith({ ok: 1 })
  })

  it('done 非 JSON 时回退为原始字符串', async () => {
    mocks.fetch.mockResolvedValueOnce(
      buildSSEResponse(['event:done\ndata:raw\n\n']),
    )
    const handlers = { onDone: vi.fn() }
    await adminAiApi.sendMessage('x', handlers)
    await new Promise(r => setTimeout(r, 0))
    expect(handlers.onDone).toHaveBeenCalledWith('raw')
  })

  it('HTTP 错误时调用 onError', async () => {
    mocks.fetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      body: null,
      headers: new Headers(),
    } as unknown as Response)
    const handlers = { onError: vi.fn() }
    await adminAiApi.sendMessage('x', handlers)
    await new Promise(r => setTimeout(r, 0))
    expect(handlers.onError).toHaveBeenCalledWith(expect.stringContaining('HTTP'))
  })

  it('fetch reject 时调用 onError', async () => {
    mocks.fetch.mockRejectedValueOnce(new Error('network down'))
    const handlers = { onError: vi.fn() }
    await adminAiApi.sendMessage('x', handlers)
    await new Promise(r => setTimeout(r, 0))
    expect(handlers.onError).toHaveBeenCalledWith('network down')
  })

  it('返回 abort 句柄', () => {
    mocks.fetch.mockResolvedValueOnce(buildSSEResponse(['']))
    const handle = adminAiApi.sendMessage('x')
    expect(typeof handle.abort).toBe('function')
  })
})
