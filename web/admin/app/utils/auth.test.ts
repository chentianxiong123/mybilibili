import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearAuthSession,
  clearAdminSession,
  decodeJwtPayload,
  getAdminPermissions,
  getAdminRole,
  getAdminToken,
  getAdminUser,
  getCurrentUserId,
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

// ====== 补充：token 刷新、登出边界、错误边界 ======

describe('auth 补充 - token 刷新判断', () => {
  it('getCurrentUserId 优先从 stored user.id 取值', () => {
    setAuthSession({ user: { id: 99, name: 'u' } })
    expect(getCurrentUserId()).toBe(99)
  })

  it('getCurrentUserId 无 user 时回退到 JWT payload.sub', () => {
    setAuthSession({ token: makeJWT({ sub: '77', exp: Math.floor(Date.now() / 1000) + 3600 }) })
    expect(getCurrentUserId()).toBe('77')
  })

  it('getCurrentUserId 无 user 无 JWT 时返回 null', () => {
    setAuthSession({ token: 'no-payload' })
    expect(getCurrentUserId()).toBeNull()
  })

  it('isAccessTokenExpired 支持 leeway 提前判定为过期', () => {
    const soon = makeJWT({ exp: Math.floor(Date.now() / 1000) + 30 })
    expect(isAccessTokenExpired(soon, 0)).toBe(false)
    expect(isAccessTokenExpired(soon, 60_000)).toBe(true)
  })

  it('isAccessTokenExpired 无 exp 字段视为过期', () => {
    expect(isAccessTokenExpired(makeJWT({ sub: '1' }))).toBe(true)
  })

  it('hasValidAccessToken 接受 leeway', () => {
    setAuthSession({ token: makeJWT({ exp: Math.floor(Date.now() / 1000) + 30 }) })
    expect(hasValidAccessToken()).toBe(true)
  })
})

describe('auth 补充 - 登出/会话隔离边界', () => {
  it('clearAuthSession 不影响 admin session', () => {
    setAdminSession({ token: 'admin-1', role: '管理员', permissions: ['a'] })
    setAuthSession({ token: 'user-1', refreshToken: 'r' })
    clearAuthSession()
    expect(getToken()).toBe('')
    expect(getAdminToken()).toBe('admin-1')
    expect(getAdminRole()).toBe('管理员')
  })

  it('clearAdminSession 不影响 user session', () => {
    setAuthSession({ token: 'u-1', refreshToken: 'r' })
    setAdminSession({ token: 'admin-1' })
    clearAdminSession()
    expect(getAdminToken()).toBe('')
    expect(getToken()).toBe('u-1')
    expect(getRefreshToken()).toBe('r')
  })

  it('setAuthSession 不传 user 时不会清空已存在 user', () => {
    setAuthSession({ user: { id: 1, name: 'A' } })
    setAuthSession({ token: 'new-token' })
    expect(getStoredUser()).toEqual({ id: 1, name: 'A' })
    expect(getToken()).toBe('new-token')
  })

  it('setAuthSession 不传 refreshToken 时不会清空已有值', () => {
    setAuthSession({ token: 'a', refreshToken: 'r1' })
    setAuthSession({ token: 'b' })
    expect(getRefreshToken()).toBe('r1')
    expect(getToken()).toBe('b')
  })

  it('hasAuthSession 在仅有 refreshToken 时返回 true', () => {
    expect(hasAuthSession()).toBe(false)
    setAuthSession({ refreshToken: 'r-only' })
    expect(hasAuthSession()).toBe(true)
  })
})

describe('auth 补充 - 错误边界', () => {
  it('decodeJwtPayload 单段字符串返回 null', () => {
    expect(decodeJwtPayload('one-part')).toBeNull()
  })

  it('decodeJwtPayload payload 不是合法 JSON 时返回 null', () => {
    const b64 = (s: string) => btoa(s).replace(/=/g, '')
    const header = b64('{}')
    const body = b64('{not-json')
    expect(decodeJwtPayload(`${header}.${body}.sig`)).toBeNull()
  })

  it('getAdminUser 解析坏 JSON 返回 null', () => {
    localStorage.setItem('admin_user', '{oops')
    expect(getAdminUser()).toBeNull()
  })

  it('getStoredUser 没有 user 时返回 null', () => {
    expect(getStoredUser()).toBeNull()
  })

  it('setAdminSession 只传 token 时 user/role/permissions 默认值', () => {
    setAdminSession({ token: 'tok' })
    expect(getAdminUser()).toBeNull()
    expect(getAdminRole()).toBe('')
    expect(getAdminPermissions()).toEqual([])
  })

  it('decodeJwtPayload 默认参数读取 getToken()', () => {
    setAuthSession({ token: unexpiredJWT })
    const p = decodeJwtPayload()
    expect(p?.sub).toBe('42')
  })

  it('decodeJwtPayload token 为空字符串返回 null', () => {
    expect(decodeJwtPayload('')).toBeNull()
  })

  it('admin permissions 重复 key 覆盖而非追加', () => {
    setAdminSession({ token: 't', permissions: ['a', 'b'] })
    setAdminSession({ token: 't', permissions: ['c'] })
    expect(getAdminPermissions()).toEqual(['c'])
  })
})