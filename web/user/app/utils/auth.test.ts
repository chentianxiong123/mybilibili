import { describe, it, expect, beforeEach } from 'vitest'
import {
  setAuthSession,
  clearAuthSession,
  hasAuthSession,
  getCurrentUserId,
  getSessionUser,
  getStoredUser,
  scrubCredentials,
} from './auth'

// safeStorage 在 happy-dom 中直接走 localStorage，无需额外 mock

describe('auth utils', () => {
  beforeEach(() => {
    localStorage.clear()
    // 部分用例带 Path=/ 写 cookie，删的时候也必须带同样的 Path 才删得掉
    document.cookie.split(';').forEach(c => {
      const name = c.trim().split('=')[0]
      document.cookie = `${name}=; Path=/; Max-Age=0`
      document.cookie = `${name}=; Max-Age=0`
    })
  })

  describe('setAuthSession', () => {
    it('丢弃凭证参数——它们只允许存在于 HttpOnly cookie', () => {
      setAuthSession({ token: 'access-123', refreshToken: 'refresh-456', user: { id: 1 } })
      const dump = JSON.stringify(localStorage)
      expect(dump).not.toContain('access-123')
      expect(dump).not.toContain('refresh-456')
      expect(document.cookie).not.toContain('access-123')
    })

    it('user 信息写入 localStorage', () => {
      const user = { id: 1, username: 'test' }
      setAuthSession({ user })
      const stored = JSON.parse(localStorage.getItem('user') || '{}')
      expect(stored.id).toBe(1)
      expect(stored.username).toBe('test')
    })

    it('不再由客户端写 user_info cookie（该 cookie 归服务端管）', () => {
      setAuthSession({ user: { id: 7, nickname: 'n' } })
      expect(document.cookie).not.toContain('user_info=')
    })
  })

  describe('clearAuthSession', () => {
    it('清空本地可见状态与遗留凭证副本', () => {
      setAuthSession({ user: { id: 1 } })
      // 升级前遗留的可窃取明文
      localStorage.setItem('teri_token', 'legacy')
      localStorage.setItem('teri_refresh_token', 'legacy-r')
      localStorage.setItem('token', 'legacy-std')

      clearAuthSession()

      expect(getStoredUser()).toBe(null)
      expect(localStorage.getItem('teri_token')).toBe(null)
      expect(localStorage.getItem('teri_refresh_token')).toBe(null)
      expect(localStorage.getItem('token')).toBe(null)
      expect(localStorage.getItem('admin_token')).toBe(null)
      expect(hasAuthSession()).toBe(false)
    })

    it('HttpOnly cookie JS 删不掉，清 cookie 走 clearServerSession', () => {
      document.cookie = `user_info=${encodeURIComponent(JSON.stringify({ id: 1 }))}; Path=/`
      clearAuthSession()
      const m = document.cookie.match(/(?:^|;\s*)user_info=([^;]*)/)
      expect(!m || !m[1]).toBe(true)
    })
  })

  describe('hasAuthSession / getSessionUser / getCurrentUserId', () => {
    it('无任何信号返回 false', () => {
      expect(hasAuthSession()).toBe(false)
      expect(getSessionUser()).toBe(null)
      expect(getCurrentUserId()).toBe(null)
    })

    it('读服务端下发的 user_info cookie', () => {
      document.cookie = `user_info=${encodeURIComponent(JSON.stringify({ id: 4, nickname: 'string' }))}; Path=/`
      expect(hasAuthSession()).toBe(true)
      expect(getSessionUser()?.nickname).toBe('string')
      expect(getCurrentUserId()).toBe(4)
    })

    it('user_info 里 + 视为空格（服务端用 url.QueryEscape 写入）', () => {
      const raw = encodeURIComponent(JSON.stringify({ id: 1, nickname: 'a b' })).replace(/%20/g, '+')
      document.cookie = `user_info=${raw}; Path=/`
      expect(getSessionUser()?.nickname).toBe('a b')
    })

    it('localStorage 有 user 也算登录（SSR/首屏兜底）', () => {
      setAuthSession({ user: { id: 9 } })
      expect(hasAuthSession()).toBe(true)
      expect(getCurrentUserId()).toBe(9)
    })

    it('user_info 优先于 localStorage', () => {
      setAuthSession({ user: { id: 9 } })
      document.cookie = `user_info=${encodeURIComponent(JSON.stringify({ id: 3 }))}; Path=/`
      expect(getCurrentUserId()).toBe(3)
    })
  })

  describe('scrubCredentials：登录响应里的 token 不能跟着 user 进可读存储', () => {
    it('登录接口把 token/refresh_token 混在 data 里时被洗掉', () => {
      const loginPayload = {
        id: 6,
        nickname: '管理员',
        token: 'eyJ-should-not-be-here',
        refresh_token: 'nor-this',
        user: { id: 6, token: 'nested-should-go' }
      }
      const scrubbed = scrubCredentials(loginPayload)
      expect(scrubbed).toEqual({ id: 6, nickname: '管理员', user: { id: 6 } })
      expect(JSON.stringify(scrubbed)).not.toContain('should-not-be-here')
      expect(JSON.stringify(scrubbed)).not.toContain('nested-should-go')
    })

    it('setAuthSession 收到带 token 的 user 时不会写出去', () => {
      setAuthSession({ user: { id: 1, token: 'leak-1', refresh_token: 'leak-2' } })
      const raw = localStorage.getItem('user') || ''
      expect(raw).not.toContain('leak-1')
      expect(raw).not.toContain('leak-2')
      expect(getStoredUser()).toEqual({ id: 1 })
    })

    it('读路径也清洗：升级前落盘的副本被覆盖读出', () => {
      localStorage.setItem('user', JSON.stringify({ id: 1, token: 'legacy-leak' }))
      expect(getStoredUser()).toEqual({ id: 1 })
    })

    it('不误伤正常业务字段', () => {
      expect(scrubCredentials({ id: 1, level: 3, pointCount: 100, avatar: 'a.png' }))
        .toEqual({ id: 1, level: 3, pointCount: 100, avatar: 'a.png' })
      expect(scrubCredentials(null)).toBeNull()
      expect(scrubCredentials('x')).toBe('x')
      expect(scrubCredentials([{ token: 'a' }, { id: 1 }])).toEqual([{}, { id: 1 }])
    })
  })

  describe('凭证不落盘（XSS 防线）', () => {
    it('任意调用都不会把 token 写进 localStorage / document.cookie', () => {
      setAuthSession({ token: 'leak-me', refreshToken: 'leak-me-too', user: { id: 1 } })
      expect(localStorage.getItem('token')).toBe(null)
      expect(localStorage.getItem('teri_token')).toBe(null)
      expect(localStorage.getItem('refreshToken')).toBe(null)
      expect(localStorage.getItem('teri_refresh_token')).toBe(null)
      expect(document.cookie).not.toContain('leak-me')
      expect(document.cookie).not.toContain('leak-me-too')
    })
  })
})
