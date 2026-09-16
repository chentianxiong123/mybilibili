import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearAuthSession,
  clearAdminSession,
  decodeJwtPayload,
  getAdminPermissions,
  getAdminRole,
  getAdminToken,
  getAdminUser,
  getRefreshToken,
  getStoredUser,
  getToken,
  hasAdminSession,
  hasAuthSession,
  hasValidAccessToken,
  isAccessTokenExpired,
  setAdminSession,
  setAuthSession,
} from '../utils/auth'

// happy-dom 18 未把 localStorage 挂到 window，手动 polyfill
const store = new Map<string, string>()
const localStorageMock = {
  getItem: (k: string) => (store.has(k) ? store.get(k) : null),
  setItem: (k: string, v: string) => { store.set(k, String(v)) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
} as Storage
if (typeof window !== 'undefined') (window as any).localStorage = localStorageMock
;(globalThis as any).localStorage = localStorageMock

// 生成一个可解码的 JWT（header.payload.signature），payload 可选 exp。
function makeJWT(payload: Record<string, unknown>): string {
  const b64 = (s: string) => btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  const header = b64(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = b64(JSON.stringify(payload))
  return `${header}.${body}.signature`
}

const unexpiredJWT = makeJWT({ sub: '42', exp: Math.floor(Date.now() / 1000) + 3600 })
const expiredJWT = makeJWT({ sub: '42', exp: Math.floor(Date.now() / 1000) - 3600 })

beforeEach(() => {
  store.clear()
  clearAuthSession()
  clearAdminSession()
})

describe('auth 会话管理', () => {
  it('setAuthSession 写入 token + cookie 后可读回', () => {
    setAuthSession({ token: unexpiredJWT, refreshToken: 'refresh-1', user: { id: 42, name: '管理员' } })
    expect(getToken()).toBe(unexpiredJWT)
    expect(getRefreshToken()).toBe('refresh-1')
    expect(getStoredUser()).toEqual({ id: 42, name: '管理员' })
  })

  it('clearAuthSession 清除所有认证信息', () => {
    setAuthSession({ token: 't', refreshToken: 'r' })
    clearAuthSession()
    expect(getToken()).toBe('')
    expect(getRefreshToken()).toBe('')
    expect(hasAuthSession()).toBe(false)
  })

  it('hasAuthSession 存在性判断', () => {
    expect(hasAuthSession()).toBe(false)
    setAuthSession({ token: 't' })
    expect(hasAuthSession()).toBe(true)
  })

  it('isAccessTokenExpired 判断过期', () => {
    expect(isAccessTokenExpired(expiredJWT)).toBe(true)
    expect(isAccessTokenExpired(unexpiredJWT)).toBe(false)
  })

  it('hasValidAccessToken 有效性判断', () => {
    expect(hasValidAccessToken()).toBe(false)
    setAuthSession({ token: expiredJWT })
    expect(hasValidAccessToken()).toBe(false)
    setAuthSession({ token: unexpiredJWT })
    expect(hasValidAccessToken()).toBe(true)
  })

  it('decodeJwtPayload 解码 payload', () => {
    const payload = decodeJwtPayload(unexpiredJWT)
    expect(payload).not.toBeNull()
    expect(payload!.sub).toBe('42')
    expect(decodeJwtPayload('')).toBeNull()
    expect(decodeJwtPayload('not-a-jwt')).toBeNull()
  })
})

describe('admin 会话管理', () => {
  it('setAdminSession 写入后可读回', () => {
    setAdminSession({ token: 'admin-token', user: { id: 1 }, role: 'SUPER', permissions: ['video.review'] })
    expect(getAdminToken()).toBe('admin-token')
    expect(getAdminUser()).toEqual({ id: 1 })
    expect(getAdminRole()).toBe('SUPER')
    expect(getAdminPermissions()).toEqual(['video.review'])
  })

  it('clearAdminSession 清除所有', () => {
    setAdminSession({ token: 'admin-token', role: 'SUPER' })
    clearAdminSession()
    expect(getAdminToken()).toBe('')
    expect(getAdminRole()).toBe('')
    expect(getAdminPermissions()).toEqual([])
    expect(hasAdminSession()).toBe(false)
  })

  it('hasAdminSession 存在性判断', () => {
    expect(hasAdminSession()).toBe(false)
    setAdminSession({ token: 'admin-token' })
    expect(hasAdminSession()).toBe(true)
  })

  it('getAdminPermissions 容错坏 JSON', () => {
    localStorage.setItem('admin_permissions', 'invalid{json')
    expect(getAdminPermissions()).toEqual([])
  })
})