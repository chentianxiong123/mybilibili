import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearAuthSession,
  clearAdminSession,
  getAdminPermissions,
  getAdminRole,
  getAdminUser,
  getCurrentUserId,
  getStoredUser,
  hasAdminSession,
  hasAuthSession,
  setAdminSession,
  setAuthSession,
  scrubCredentials,
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

beforeEach(() => {
  store.clear()
  clearAuthSession()
  clearAdminSession()
})

// JS 层只碰展示信息，凭证一律 HttpOnly cookie
const jwtLike = 'eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI0MiJ9.sig'

describe('auth 会话管理（只留展示信息）', () => {
  it('setAuthSession 丢弃 token / refreshToken，只保存 user', () => {
    setAuthSession({ token: jwtLike, refreshToken: 'refresh-1', user: { id: 42, name: '管理员' } })
    expect(getStoredUser()).toEqual({ id: 42, name: '管理员' })
    expect(store.has('token')).toBe(false)
    expect(store.has('refreshToken')).toBe(false)
    expect(JSON.stringify([...store.values()])).not.toContain(jwtLike)
    expect(JSON.stringify([...store.values()])).not.toContain('refresh-1')
  })

  it('setAuthSession 即使只传凭证也什么都不留', () => {
    setAuthSession({ token: jwtLike, refreshToken: 'refresh-1' })
    expect(store.size).toBe(0)
    expect(hasAuthSession()).toBe(false)
  })

  it('hasAuthSession 由 user 回答，不由凭证回答', () => {
    expect(hasAuthSession()).toBe(false)
    setAuthSession({ token: jwtLike })
    expect(hasAuthSession()).toBe(false)
    setAuthSession({ user: { id: 1 } })
    expect(hasAuthSession()).toBe(true)
  })

  it('setAuthSession 不传 user 时不会清空已存在 user', () => {
    setAuthSession({ user: { id: 1, name: 'A' } })
    setAuthSession({ token: 'whatever' })
    expect(getStoredUser()).toEqual({ id: 1, name: 'A' })
  })

  it('clearAuthSession 清展示信息并顺手清掉升级前的明文凭证明文', () => {
    localStorage.setItem('token', jwtLike)
    localStorage.setItem('refreshToken', 'r')
    localStorage.setItem('teri_token', 't')
    setAuthSession({ user: { id: 1 } })
    clearAuthSession()
    expect(getStoredUser()).toBeNull()
    expect(hasAuthSession()).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('refreshToken')).toBeNull()
    expect(localStorage.getItem('teri_token')).toBeNull()
  })

  it('clearAdminSession 不影响 user session', () => {
    setAuthSession({ user: { id: 1 } })
    setAdminSession({ user: { id: 1 }, role: '管理员' })
    clearAdminSession()
    expect(hasAdminSession()).toBe(false)
    expect(hasAuthSession()).toBe(true)
    expect(getAdminRole()).toBe('')
  })

  it('clearAuthSession 不影响 admin session', () => {
    setAdminSession({ user: { id: 1 }, role: '超级管理员', permissions: ['a'] })
    setAuthSession({ user: { id: 2 } })
    clearAuthSession()
    expect(hasAuthSession()).toBe(false)
    expect(hasAdminSession()).toBe(true)
    expect(getAdminRole()).toBe('超级管理员')
    expect(getAdminPermissions()).toEqual(['a'])
  })
})

describe('getCurrentUserId', () => {
  it('从 stored user.id 取值', () => {
    setAuthSession({ user: { id: 99, name: 'u' } })
    expect(getCurrentUserId()).toBe(99)
  })

  it('没有 user 时返回 null（不再回退去解 JWT——token 在 HttpOnly cookie 里读不到）', () => {
    setAuthSession({ token: jwtLike })
    expect(getCurrentUserId()).toBeNull()
  })

  it('user.id 缺失时返回 null', () => {
    setAuthSession({ user: { name: 'noid' } })
    expect(getCurrentUserId()).toBeNull()
  })
})

describe('admin 会话管理', () => {
  it('setAdminSession 写入后可读回', () => {
    setAdminSession({ user: { id: 1 }, role: 'SUPER', permissions: ['video.review'] })
    expect(getAdminUser()).toEqual({ id: 1 })
    expect(getAdminRole()).toBe('SUPER')
    expect(getAdminPermissions()).toEqual(['video.review'])
  })

  it('setAdminSession 丢弃 token，不把凭证写进 localStorage', () => {
    setAdminSession({ token: 'admin-token', user: { id: 1 } })
    expect(store.has('admin_token')).toBe(false)
    expect(JSON.stringify([...store.values()])).not.toContain('admin-token')
    expect(getAdminUser()).toEqual({ id: 1 })
  })

  it('clearAdminSession 清除所有并清掉升级前的 admin_token 副本', () => {
    localStorage.setItem('admin_token', 'legacy')
    setAdminSession({ user: { id: 1 }, role: 'SUPER', permissions: ['video.review'] })
    clearAdminSession()
    expect(getAdminUser()).toBeNull()
    expect(getAdminRole()).toBe('')
    expect(getAdminPermissions()).toEqual([])
    expect(hasAdminSession()).toBe(false)
    expect(localStorage.getItem('admin_token')).toBeNull()
  })

  it('hasAdminSession 由 admin_user 回答', () => {
    expect(hasAdminSession()).toBe(false)
    setAdminSession({ token: 'admin-token' })
    expect(hasAdminSession()).toBe(false)
    setAdminSession({ user: { id: 1 } })
    expect(hasAdminSession()).toBe(true)
  })

  it('getAdminPermissions 容错坏 JSON', () => {
    localStorage.setItem('admin_permissions', 'invalid{json')
    expect(getAdminPermissions()).toEqual([])
  })

  it('admin permissions 重复 key 覆盖而非追加', () => {
    setAdminSession({ user: { id: 1 }, permissions: ['a', 'b'] })
    setAdminSession({ user: { id: 1 }, permissions: ['c'] })
    expect(getAdminPermissions()).toEqual(['c'])
  })

  it('setAdminSession 只传 user 时 role/permissions 默认值', () => {
    setAdminSession({ user: { id: 1 } })
    expect(getAdminRole()).toBe('')
    expect(getAdminPermissions()).toEqual([])
  })
})

describe('scrubCredentials：登录响应里的 token 不能跟着展示信息进 localStorage', () => {
  it('setAdminSession 收到带 token 的 user 时不会写出去', () => {
    setAdminSession({ token: 'admin-tok', user: { id: 1, token: 'nested-tok' }, role: '管理员' })
    const raw = localStorage.getItem('admin_user') || ''
    expect(raw).not.toContain('nested-tok')
    expect(raw).not.toContain('admin-tok')
    expect(getAdminUser()).toEqual({ id: 1 })
    expect(store.has('admin_token')).toBe(false)
  })

  it('读路径也清洗：升级前落盘的副本被覆盖读出', () => {
    localStorage.setItem('admin_user', JSON.stringify({ id: 1, token: 'legacy-leak' }))
    expect(getAdminUser()).toEqual({ id: 1 })
  })

  it('不误伤正常业务字段', () => {
    expect(scrubCredentials({ id: 1, role: '管理员', permissions: ['a'] }))
      .toEqual({ id: 1, role: '管理员', permissions: ['a'] })
    expect(scrubCredentials(null)).toBeNull()
  })
})

describe('错误边界', () => {
  it('getAdminUser 解析坏 JSON 返回 null', () => {
    localStorage.setItem('admin_user', '{oops')
    expect(getAdminUser()).toBeNull()
  })

  it('getStoredUser 没有 user 时返回 null', () => {
    expect(getStoredUser()).toBeNull()
  })

  it('getStoredUser 解析坏 JSON 返回 null', () => {
    localStorage.setItem('user', '{oops')
    expect(getStoredUser()).toBeNull()
  })
})
