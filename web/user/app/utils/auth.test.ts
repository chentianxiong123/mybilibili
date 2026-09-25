import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  setAuthSession,
  clearAuthSession,
  getToken,
  getRefreshToken,
  isAccessTokenExpired,
  hasValidAccessToken,
  hasAuthSession,
  decodeJwtPayload,
} from './auth'

// safeStorage 在 happy-dom 中直接走 localStorage，无需额外 mock

function makeJwt(payload: Record<string, any>) {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = btoa(JSON.stringify(payload))
  const sig = 'fake-signature'
  return `${header}.${body}.${sig}`
}

describe('auth utils', () => {
  beforeEach(() => {
    localStorage.clear()
    document.cookie.split(';').forEach(c => {
      document.cookie = c.trim().split('=')[0] + '=; Max-Age=0'
    })
  })

  describe('setAuthSession / getToken / getRefreshToken', () => {
    it('设置 token 后能读取', () => {
      setAuthSession({ token: 'access-123', refreshToken: 'refresh-456' })
      expect(getToken()).toBe('access-123')
      expect(getRefreshToken()).toBe('refresh-456')
    })

    // 阶段 1：token / refresh_token 改由服务端以 HttpOnly 下发，
    // 前端若再用 document.cookie 写一遍会把 HttpOnly 标记冲掉，等于自毁防线。
    it('不再把 token 写进 cookie（HttpOnly 由服务端持有）', () => {
      setAuthSession({ token: 'tok-abc' })
      expect(getToken()).toBe('tok-abc')
      expect(document.cookie).not.toContain('token=tok-abc')
    })

    it('不再把 refreshToken 写进 cookie', () => {
      setAuthSession({ refreshToken: 'ref-xyz' })
      expect(getRefreshToken()).toBe('ref-xyz')
      expect(document.cookie).not.toContain('refresh_token=ref-xyz')
    })

    it('user_info 仍写入 cookie（非凭证，供 useAuth 直接渲染）', () => {
      setAuthSession({ user: { id: 7, nickname: 'n' } })
      expect(document.cookie).toContain('user_info=')
    })

    it('user 信息写入 localStorage', () => {
      const user = { id: 1, username: 'test' }
      setAuthSession({ user })
      const stored = JSON.parse(localStorage.getItem('user') || '{}')
      expect(stored.id).toBe(1)
      expect(stored.username).toBe('test')
    })
  })

  describe('clearAuthSession', () => {
    it('清除后 token 和 refreshToken 为空', () => {
      setAuthSession({ token: 't', refreshToken: 'r' })
      clearAuthSession()
      expect(getToken()).toBe('')
      expect(getRefreshToken()).toBe('')
    })

    // HttpOnly cookie JS 删不掉，清 cookie 由 api/session 的 clearServerSession() 走服务端完成
    it('清除后本地凭证 cookie 不残留', () => {
      setAuthSession({ token: 't', refreshToken: 'r', user: { id: 1 } })
      expect(document.cookie).toContain('user_info=')
      clearAuthSession()
      // happy-dom 删除后可能保留空键名，这里只关心值是否还在
      const m = document.cookie.match(/(?:^|;\s*)user_info=([^;]*)/)
      expect(!m || !m[1]).toBe(true)
      expect(getToken()).toBe('')
      expect(getRefreshToken()).toBe('')
    })
  })

  describe('isAccessTokenExpired', () => {
    it('过期 token 返回 true', () => {
      const token = makeJwt({ exp: Math.floor(Date.now() / 1000) - 100 })
      expect(isAccessTokenExpired(token)).toBe(true)
    })

    it('未过期 token 返回 false', () => {
      const token = makeJwt({ exp: Math.floor(Date.now() / 1000) + 3600 })
      expect(isAccessTokenExpired(token)).toBe(false)
    })

    it('无 exp 的 token 返回 true', () => {
      const token = makeJwt({ sub: 1 })
      expect(isAccessTokenExpired(token)).toBe(true)
    })

    it('空 token 返回 true', () => {
      expect(isAccessTokenExpired('')).toBe(true)
    })
  })

  describe('hasValidAccessToken', () => {
    it('无 token 返回 false', () => {
      expect(hasValidAccessToken()).toBe(false)
    })

    it('有有效 token 返回 true', () => {
      const token = makeJwt({ exp: Math.floor(Date.now() / 1000) + 3600 })
      setAuthSession({ token })
      expect(hasValidAccessToken()).toBe(true)
    })

    it('有过期 token 返回 false', () => {
      const token = makeJwt({ exp: Math.floor(Date.now() / 1000) - 100 })
      setAuthSession({ token })
      expect(hasValidAccessToken()).toBe(false)
    })
  })

  describe('hasAuthSession', () => {
    it('无任何 token 返回 false', () => {
      expect(hasAuthSession()).toBe(false)
    })

    it('有 token 返回 true', () => {
      setAuthSession({ token: 't' })
      expect(hasAuthSession()).toBe(true)
    })

    it('有 refreshToken 返回 true', () => {
      setAuthSession({ refreshToken: 'r' })
      expect(hasAuthSession()).toBe(true)
    })
  })

  describe('decodeJwtPayload', () => {
    it('正常解码 payload', () => {
      const payload = { sub: 42, name: 'test' }
      const token = makeJwt(payload)
      expect(decodeJwtPayload(token)).toEqual(payload)
    })

    it('无效 token 返回 null', () => {
      expect(decodeJwtPayload('not-a-jwt')).toBe(null)
    })

    it('空 token 返回 null', () => {
      expect(decodeJwtPayload('')).toBe(null)
    })
  })
})
