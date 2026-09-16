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

    it('token 同步写入 cookie', () => {
      setAuthSession({ token: 'tok-abc' })
      const cookies = document.cookie
      expect(cookies).toContain('token=tok-abc')
    })

    it('refreshToken 同步写入 cookie', () => {
      setAuthSession({ refreshToken: 'ref-xyz' })
      const cookies = document.cookie
      expect(cookies).toContain('refresh_token=ref-xyz')
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

    it('清除后 cookie 也删除', () => {
      setAuthSession({ token: 't', refreshToken: 'r' })
      clearAuthSession()
      const cookies = document.cookie
      expect(cookies).not.toContain('token=t')
      expect(cookies).not.toContain('refresh_token=r')
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
