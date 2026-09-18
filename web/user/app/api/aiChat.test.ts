import { describe, it, expect, vi, beforeEach } from 'vitest'

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

import { aiChatApi } from './aiChat'

function buildJsonResponse(body: any, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: new Headers({ 'content-type': 'application/json' }),
    json: () => Promise.resolve(body),
    text: () => Promise.resolve(JSON.stringify(body)),
  } as unknown as Response
}

describe('aiChatApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.safeStorage.getItem.mockReturnValue(null)
  })

  describe('getConversations', () => {
    it('未登录时返回 code 401 不调 fetch', async () => {
      const res = await aiChatApi.getConversations()
      expect(res).toEqual({ code: 401, message: '请先登录', data: [] })
      expect(mocks.fetch).not.toHaveBeenCalled()
    })

    it('登录后空 data 返回空数组', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce(buildJsonResponse({ code: 200, data: [] }))
      const res = await aiChatApi.getConversations()
      expect(res.data).toEqual([])
    })

    it('登录后有 messages 返回单条固定 conversation', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce(
        buildJsonResponse({
          code: 200,
          data: [
            { id: 'm1', content: 'a', createdAt: '2024-01-01T00:00:00Z', role: 'user' },
            { id: 'm2', content: 'b', createdAt: '2024-01-02T00:00:00Z', role: 'assistant' },
          ],
        }),
      )
      const res = await aiChatApi.getConversations()
      expect(res.data).toHaveLength(1)
      expect(res.data[0].id).toBe('customer-service')
      expect(res.data[0].title).toBe('AI客服对话')
      expect(res.data[0].updatedAt).toBe('2024-01-02T00:00:00Z')
    })

    it('非 JSON 响应时抛错', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'text/html' }),
        text: () => Promise.resolve('<html>x</html>'),
      } as unknown as Response)
      await expect(aiChatApi.getConversations()).rejects.toThrow(/non-JSON/)
    })
  })

  describe('createConversation', () => {
    it('未登录返回 401', async () => {
      const res = await aiChatApi.createConversation()
      expect(res.code).toBe(401)
    })

    it('登录返回固定 ID 的 conversation', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 2 }) : 't',
      )
      const res = await aiChatApi.createConversation()
      expect(res.code).toBe(200)
      expect(res.data.id).toBe('customer-service')
      expect(res.data.title).toBe('AI客服对话')
    })
  })

  describe('getMessages', () => {
    it('未登录返回 401', async () => {
      const res = await aiChatApi.getMessages(1)
      expect(res).toEqual({ code: 401, message: '请先登录', data: [] })
    })

    it('登录后 normalize message role/content', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce(
        buildJsonResponse({
          code: 200,
          data: [
            { id: 'm1', role: 'user', content: 'hi', createdAt: '2024-01-01T00:00:00Z' },
            { id: 'm2', role: 'assistant', content: 'hello' },
          ],
        }),
      )
      const res = await aiChatApi.getMessages(1)
      expect(res.data).toHaveLength(2)
      expect(res.data[0].role).toBe('user')
      expect(res.data[1].role).toBe('assistant')
      expect(typeof res.data[1].id).toBe('string')
    })

    it('role 异常时归为 assistant', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce(
        buildJsonResponse({
          code: 200,
          data: [{ role: 'robot', content: 'x' }],
        }),
      )
      const res = await aiChatApi.getMessages(1)
      expect(res.data[0].role).toBe('assistant')
    })
  })

  describe('sendMessage', () => {
    it('未登录触发 onError 并返回 abort 句柄', async () => {
      const onError = vi.fn()
      const handle = aiChatApi.sendMessage(1, 'hi', { onError })
      expect(typeof handle.abort).toBe('function')
      expect(onError).toHaveBeenCalledWith('请先登录')
      expect(mocks.fetch).not.toHaveBeenCalled()
    })
  })

  describe('transferToHuman', () => {
    it('未登录返回 401', async () => {
      const res = await aiChatApi.transferToHuman()
      expect(res).toEqual({ code: 401, message: '请先登录', data: null })
    })

    it('登录调用 transfer 接口', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce(buildJsonResponse({ code: 200, data: { ok: true } }))
      const res = await aiChatApi.transferToHuman('need help')
      expect(res.code).toBe(200)
      expect(mocks.fetch).toHaveBeenCalled()
      const opts = mocks.fetch.mock.calls[0][1] as RequestInit
      expect(opts.body).toBe(JSON.stringify({ userId: 1, reason: 'need help' }))
    })

    it('HTTP 错误抛出', async () => {
      mocks.safeStorage.getItem.mockImplementation((k: string) =>
        k === 'user' ? JSON.stringify({ id: 1 }) : 't',
      )
      mocks.fetch.mockResolvedValueOnce(buildJsonResponse({}, 500))
      await expect(aiChatApi.transferToHuman()).rejects.toThrow(/HTTP/)
    })
  })
})
