import { describe, it, expect, beforeEach, vi } from 'vitest'

vi.mock('../storage_layer', () => ({
  default: {
    get: vi.fn(),
    set: vi.fn(),
    remove: vi.fn()
  },
  K: {
    token: 'token',
    refreshToken: 'refresh_token',
    user: 'user'
  }
}))

vi.mock('../../api/client', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn()
  }
}))

import {
  getLocalUser,
  getLocalUserId,
  isLogin,
  saveSession,
  clearSession,
  logout,
  tryRefresh,
  bootstrapSession
} from '../session'
import storage from '../storage_layer'
import api from '../../api/client'

const mockStorage = vi.mocked(storage)
const mockApi = vi.mocked(api)

describe('session', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getLocalUser', () => {
    it('有用户数据 → 返回清洗后的用户（token 等敏感字段被移除）', () => {
      const user = { id: 1, name: 'admin', token: 'secret' }
      mockStorage.get.mockReturnValue(user)
      const result = getLocalUser()
      expect(result).toEqual({ id: 1, name: 'admin' })
      expect(mockStorage.get).toHaveBeenCalledWith('user')
    })

    it('无用户数据 → null', () => {
      mockStorage.get.mockReturnValue(null)
      expect(getLocalUser()).toBeNull()
    })
  })

  describe('getLocalUserId', () => {
    it('有用户且有 id → 返回 id', () => {
      mockStorage.get.mockReturnValue({ id: 42 })
      expect(getLocalUserId()).toBe(42)
    })

    it('有 userId 字段', () => {
      mockStorage.get.mockReturnValue({ userId: 99 })
      expect(getLocalUserId()).toBe(99)
    })

    it('无用户 → 0', () => {
      mockStorage.get.mockReturnValue(null)
      expect(getLocalUserId()).toBe(0)
    })
  })

  describe('isLogin', () => {
    it('有用户 → true', () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      expect(isLogin()).toBe(true)
    })

    it('无用户 → false', () => {
      mockStorage.get.mockReturnValue(null)
      expect(isLogin()).toBe(false)
    })
  })

  describe('saveSession', () => {
    it('只保存清洗后的 user，不保存 token/refreshToken', () => {
      saveSession('token123', 'refresh456', { id: 1, name: 'u', token: 's' })
      expect(mockStorage.set).toHaveBeenCalledWith('user', { id: 1, name: 'u' })
      // 不应保存 token
      expect(mockStorage.set).not.toHaveBeenCalledWith('token', expect.anything())
    })
  })

  describe('clearSession', () => {
    it('清除 user + token + refreshToken', () => {
      clearSession()
      expect(mockStorage.remove).toHaveBeenCalledWith('user')
      expect(mockStorage.remove).toHaveBeenCalledWith('token')
      expect(mockStorage.remove).toHaveBeenCalledWith('refresh_token')
    })
  })

  describe('logout', () => {
    it('调 /user/logout 后清除本地', async () => {
      mockApi.post.mockResolvedValue({})
      await logout()
      expect(mockApi.post).toHaveBeenCalledWith('/user/logout', {})
      expect(mockStorage.remove).toHaveBeenCalledWith('user')
    })

    it('网络失败也能清除本地', async () => {
      mockApi.post.mockRejectedValue(new Error('network'))
      await logout()
      expect(mockStorage.remove).toHaveBeenCalledWith('user')
    })
  })

  describe('tryRefresh', () => {
    it('未登录 → 直接返回 false', async () => {
      mockStorage.get.mockReturnValue(null)
      const result = await tryRefresh()
      expect(result).toBe(false)
    })

    it('已登录 → 调 /user/token/refresh', async () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      mockApi.post.mockResolvedValue({ code: 200 })
      const result = await tryRefresh()
      expect(result).toBe(true)
      expect(mockApi.post).toHaveBeenCalledWith('/user/token/refresh', {})
    })

    it('refresh 失败 → 返回 false', async () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      mockApi.post.mockRejectedValue(new Error('fail'))
      const result = await tryRefresh()
      expect(result).toBe(false)
    })

    it('并发调用只发一次请求', async () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      let callCount = 0
      mockApi.post.mockImplementation(() => {
        callCount++
        return new Promise(resolve => setTimeout(() => resolve({ code: 200 }), 100))
      })

      const [r1, r2, r3] = await Promise.all([
        tryRefresh(),
        tryRefresh(),
        tryRefresh()
      ])
      expect(callCount).toBe(1)
      expect(r1).toBe(true)
      expect(r2).toBe(true)
      expect(r3).toBe(true)
    })
  })

  describe('bootstrapSession', () => {
    it('无登录态 → false', async () => {
      mockStorage.get.mockReturnValue(null)
      const result = await bootstrapSession()
      expect(result).toBe(false)
    })

    it('有登录态 + /user/me 成功 → true', async () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      mockApi.get.mockResolvedValue({ data: { id: 1 } })
      const result = await bootstrapSession()
      expect(result).toBe(true)
    })

    it('有登录态 + 401 + refresh 成功 → true', async () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      mockApi.get
        .mockRejectedValueOnce({ response: { status: 401 } })
      mockApi.post.mockResolvedValue({ code: 200 })
      const result = await bootstrapSession()
      expect(result).toBe(true)
    })

    it('网络异常且有本地用户 → true（离线可用）', async () => {
      mockStorage.get.mockReturnValue({ id: 1 })
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await bootstrapSession()
      expect(result).toBe(true)
    })

    it('网络异常且无本地用户 → false', async () => {
      // mockStorage.get 返回 null（isLogin = false）或返回 null（getLocalUser）
      mockStorage.get.mockReturnValue(null)
      const result = await bootstrapSession()
      expect(result).toBe(false)
    })
  })
})
