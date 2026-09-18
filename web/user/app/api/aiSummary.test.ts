import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  fetch: vi.fn(),
  auth: {
    getToken: vi.fn(),
  },
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), { get: mocks.apiGet }),
}))

vi.mock('@/utils/auth', () => ({
  getToken: mocks.auth.getToken,
}))

global.fetch = mocks.fetch

import { aiSummaryApi } from './aiSummary'

function buildSSEResponse(events: string[], ok = true) {
  const encoder = new TextEncoder()
  const stream = new ReadableStream({
    start(controller) {
      for (const e of events) controller.enqueue(encoder.encode(e))
      controller.close()
    },
  })
  return {
    ok,
    status: ok ? 200 : 500,
    statusText: ok ? 'OK' : 'ERR',
    body: stream,
    headers: new Headers({ 'content-type': 'text/event-stream' }),
  } as unknown as Response
}

describe('aiSummaryApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.auth.getToken.mockReturnValue(null)
  })

  describe('REST', () => {
    it('getSummary', async () => {
      mocks.apiGet.mockResolvedValueOnce({ code: 200, data: 'summary' })
      const res = await aiSummaryApi.getSummary(7)
      expect(mocks.apiGet).toHaveBeenCalledWith('/ai/summary/7')
      expect(res).toEqual({ code: 200, data: 'summary' })
    })

    it('checkSummary', async () => {
      mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { has: true } })
      await aiSummaryApi.checkSummary(8)
      expect(mocks.apiGet).toHaveBeenCalledWith('/ai/summary/check/8')
    })
  })

  describe('streamSummary', () => {
    it('401 时调用 onError "未登录或登录已过期"', async () => {
      mocks.fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        body: null,
        headers: new Headers(),
      } as unknown as Response)
      const onError = vi.fn()
      aiSummaryApi.streamSummary(9, { onError })
      await new Promise(r => setTimeout(r, 0))
      expect(onError).toHaveBeenCalledWith(expect.stringContaining('未登录'))
    })

    it('无 token 时 Authorization 为空', async () => {
      mocks.fetch.mockResolvedValueOnce(buildSSEResponse(['']))
      aiSummaryApi.streamSummary(10)
      const headers = (mocks.fetch.mock.calls[0][1] as RequestInit).headers as Record<string, string>
      expect(headers['Authorization']).toBe('')
      expect(headers['Accept']).toBe('text/event-stream')
    })

    it('有 token 时 Authorization 携带 Bearer', async () => {
      mocks.auth.getToken.mockReturnValue('t123')
      mocks.fetch.mockResolvedValueOnce(buildSSEResponse(['']))
      aiSummaryApi.streamSummary(11)
      const headers = (mocks.fetch.mock.calls[0][1] as RequestInit).headers as Record<string, string>
      expect(headers['Authorization']).toBe('Bearer t123')
    })

    it('解析 data 事件: base64 编码', async () => {
      // 'hi' base64 = 'aGk='
      mocks.fetch.mockResolvedValueOnce(
        buildSSEResponse(['event:data\ndata:aGk=\n\n']),
      )
      const onData = vi.fn()
      aiSummaryApi.streamSummary(12, { onData })
      await new Promise(r => setTimeout(r, 0))
      await new Promise(r => setTimeout(r, 0))
      expect(onData).toHaveBeenCalledWith('hi')
    })

    it('解析 meta 事件 JSON', async () => {
      mocks.fetch.mockResolvedValueOnce(
        buildSSEResponse(['event:meta\ndata:{"k":1}\n\n']),
      )
      const onMeta = vi.fn()
      aiSummaryApi.streamSummary(13, { onMeta })
      await new Promise(r => setTimeout(r, 0))
      await new Promise(r => setTimeout(r, 0))
      expect(onMeta).toHaveBeenCalledWith({ k: 1 })
    })

    it('解析 done 事件', async () => {
      mocks.fetch.mockResolvedValueOnce(
        buildSSEResponse(['event:done\ndata:END\n\n']),
      )
      const onDone = vi.fn()
      aiSummaryApi.streamSummary(14, { onDone })
      await new Promise(r => setTimeout(r, 0))
      await new Promise(r => setTimeout(r, 0))
      expect(onDone).toHaveBeenCalledWith('END')
    })

    it('解析 error 事件', async () => {
      mocks.fetch.mockResolvedValueOnce(
        buildSSEResponse(['event:error\ndata:oops\n\n']),
      )
      const onError = vi.fn()
      aiSummaryApi.streamSummary(15, { onError })
      await new Promise(r => setTimeout(r, 0))
      await new Promise(r => setTimeout(r, 0))
      expect(onError).toHaveBeenCalledWith('oops')
    })

    it('返回 abort/close 句柄', () => {
      mocks.fetch.mockResolvedValueOnce(buildSSEResponse(['']))
      const handle = aiSummaryApi.streamSummary(16)
      expect(typeof handle.abort).toBe('function')
      expect(typeof handle.close).toBe('function')
    })

    it('fetch reject 时 onError 收到消息', async () => {
      mocks.fetch.mockRejectedValueOnce(new Error('fail'))
      const onError = vi.fn()
      aiSummaryApi.streamSummary(17, { onError })
      await new Promise(r => setTimeout(r, 0))
      expect(onError).toHaveBeenCalledWith('fail')
    })
  })
})
