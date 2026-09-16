import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock API client
vi.mock('@/api/client', () => ({
  userApi: {
    login: vi.fn(),
    register: vi.fn(),
    getUserById: vi.fn(),
    updateUser: vi.fn(),
  },
}))

// Mock auth utils
vi.mock('@/utils/auth', () => ({
  setAuthSession: vi.fn(),
  clearAuthSession: vi.fn(),
  getToken: vi.fn(() => ''),
  getRefreshToken: vi.fn(() => ''),
  getStoredUser: vi.fn(() => null),
  getCurrentUserId: vi.fn(() => null),
}))

import { useUserStore } from './user'
import { userApi } from '@/api/client'
import { setAuthSession, clearAuthSession } from '@/utils/auth'

describe('user store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('初始状态', () => {
    it('默认未登录', () => {
      const store = useUserStore()
      expect(store.isLoggedIn).toBe(false)
      expect(store.token).toBe('')
      expect(store.userInfo.username).toBe('')
    })

    it('默认用户信息字段正确', () => {
      const store = useUserStore()
      expect(store.userInfo.level).toBe(1)
      expect(store.userInfo.gender).toBe(0)
      expect(store.userInfo.followingCount).toBe(0)
    })
  })

  describe('login', () => {
    it('登录成功更新状态', async () => {
      const mockResponse = {
        code: 200,
        data: {
          token: 'access-token',
          refresh_token: 'refresh-token',
          user: { id: 1, username: 'testuser', nickname: 'Test' },
        },
      }
      vi.mocked(userApi.login).mockResolvedValue(mockResponse)

      const store = useUserStore()
      const result = await store.login({ username: 'testuser', password: '123456' })

      expect(result.success).toBe(true)
      expect(store.isLoggedIn).toBe(true)
      expect(store.token).toBe('access-token')
      expect(store.userInfo.username).toBe('testuser')
      expect(setAuthSession).toHaveBeenCalledWith(expect.objectContaining({
        token: 'access-token',
        refreshToken: 'refresh-token',
      }))
    })

    it('登录失败返回错误信息', async () => {
      vi.mocked(userApi.login).mockResolvedValue({
        code: 401,
        message: '密码错误',
        data: null,
      })

      const store = useUserStore()
      const result = await store.login({ username: 'testuser', password: 'wrong' })

      expect(result.success).toBe(false)
      expect(result.message).toBe('密码错误')
      expect(store.isLoggedIn).toBe(false)
    })

    it('登录异常返回失败', async () => {
      vi.mocked(userApi.login).mockRejectedValue(new Error('网络错误'))

      const store = useUserStore()
      const result = await store.login({ username: 'testuser', password: '123' })

      expect(result.success).toBe(false)
    })
  })

  describe('register', () => {
    it('注册成功', async () => {
      vi.mocked(userApi.register).mockResolvedValue({
        code: 200,
        message: '注册成功',
        data: null,
      })

      const store = useUserStore()
      const result = await store.register({
        username: 'newuser',
        password: 'Test1234',
      })

      expect(result.success).toBe(true)
      expect(userApi.register).toHaveBeenCalledWith(expect.objectContaining({
        username: 'newuser',
      }))
    })

    it('注册失败返回错误', async () => {
      vi.mocked(userApi.register).mockResolvedValue({
        code: 409,
        message: '用户名已存在',
        data: null,
      })

      const store = useUserStore()
      const result = await store.register({
        username: 'existing',
        password: 'Test1234',
      })

      expect(result.success).toBe(false)
      expect(result.message).toContain('已存在')
    })
  })

  describe('logout', () => {
    it('退出后状态清空', () => {
      const store = useUserStore()
      // 先模拟登录状态
      store.token = 'some-token'
      store.refreshToken = 'some-refresh'
      store.isLoggedIn = true
      store.userInfo.username = 'testuser'

      store.logout()

      expect(store.isLoggedIn).toBe(false)
      expect(store.token).toBe('')
      expect(store.refreshToken).toBe('')
      expect(store.userInfo.username).toBe('')
      expect(clearAuthSession).toHaveBeenCalled()
    })
  })

  describe('setToken / clearToken', () => {
    it('setToken 写入 token 并调用 setAuthSession', () => {
      const store = useUserStore()
      store.setToken('new-token', 'new-refresh')

      expect(store.token).toBe('new-token')
      expect(store.refreshToken).toBe('new-refresh')
      expect(setAuthSession).toHaveBeenCalledWith({
        token: 'new-token',
        refreshToken: 'new-refresh',
      })
    })

    it('clearToken 清空 token', () => {
      const store = useUserStore()
      store.token = 't'
      store.refreshToken = 'r'
      store.clearToken()

      expect(store.token).toBe('')
      expect(store.refreshToken).toBe('')
      expect(clearAuthSession).toHaveBeenCalled()
    })
  })

  describe('getters', () => {
    it('getUserId 返回用户 ID', () => {
      const store = useUserStore()
      store.userInfo.id = 42
      expect(store.getUserId).toBe(42)
    })

    it('getIsLoggedIn 返回登录状态', () => {
      const store = useUserStore()
      expect(store.getIsLoggedIn).toBe(false)
      store.isLoggedIn = true
      expect(store.getIsLoggedIn).toBe(true)
    })
  })
})
