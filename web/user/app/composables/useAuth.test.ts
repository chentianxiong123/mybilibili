import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// ===== 在 import useAuth 之前注入 Nuxt 的 auto-import 全局变量 =====
// useAuth.ts 依赖 computed (Nuxt auto-import) 和 useCookie (Nuxt auto-import)
import * as Vue from 'vue'
;(globalThis as any).computed = Vue.computed
;(globalThis as any).ref = Vue.ref
;(globalThis as any).watch = Vue.watch
;(globalThis as any).onMounted = Vue.onMounted
;(globalThis as any).onUnmounted = Vue.onUnmounted
;(globalThis as any).reactive = Vue.reactive
;(globalThis as any).nextTick = Vue.nextTick
;(globalThis as any).useCookie = vi.fn((_key: string, opts: any = {}) => {
  const defaultVal = typeof opts?.default === 'function' ? opts.default() : null
  return Vue.ref(defaultVal)
})

import { useAuth } from './useAuth'

describe('useAuth', () => {
  beforeEach(() => {
    localStorage.clear()
    document.cookie.split(';').forEach(c => {
      document.cookie = c.trim().split('=')[0] + '=; Max-Age=0'
    })
    ;(globalThis as any).useCookie.mockClear()
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('返回字段', () => {
    it('返回 token, refreshToken, user, isLoggedIn', () => {
      const r = useAuth()
      expect(r).toHaveProperty('token')
      expect(r).toHaveProperty('refreshToken')
      expect(r).toHaveProperty('user')
      expect(r).toHaveProperty('isLoggedIn')
    })

    it('调用 useCookie 3 次，分别取 token / refresh_token / user_info', () => {
      useAuth()
      expect((globalThis as any).useCookie).toHaveBeenCalledTimes(3)
      const keys = (globalThis as any).useCookie.mock.calls.map((c: any[]) => c[0])
      expect(keys).toContain('token')
      expect(keys).toContain('refresh_token')
      expect(keys).toContain('user_info')
    })
  })

  describe('cookie 优先路径（useCookie 直接返回值）', () => {
    it('useCookie 返回 token 时该值被使用', () => {
      ;(globalThis as any).useCookie = vi.fn((key: string) => {
        if (key === 'token') return Vue.ref('cookie-tk')
        if (key === 'refresh_token') return Vue.ref('cookie-rt')
        if (key === 'user_info') return Vue.ref(null)
        return Vue.ref(null)
      })
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.token.value).toBe('cookie-tk')
        expect(r.refreshToken.value).toBe('cookie-rt')
        expect(r.isLoggedIn.value).toBe(true)
      })
    })

    it('user_info cookie 解码失败时 user 为 null', () => {
      ;(globalThis as any).useCookie = vi.fn((key: string) => {
        if (key === 'user_info') return Vue.ref('not-json')
        return Vue.ref(null)
      })
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.user.value).toBe(null)
      })
    })

    it('user_info cookie 合法 JSON 时返回解码对象', () => {
      const userObj = { id: 5, username: 'alice' }
      ;(globalThis as any).useCookie = vi.fn((key: string) => {
        if (key === 'user_info') return Vue.ref(encodeURIComponent(JSON.stringify(userObj)))
        return Vue.ref(null)
      })
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.user.value).toEqual(userObj)
      })
    })

    it('isLoggedIn 依赖 token，token 为空时为 false', () => {
      ;(globalThis as any).useCookie = vi.fn(() => Vue.ref(null))
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.isLoggedIn.value).toBe(false)
      })
    })
  })

  describe('响应性', () => {
    it('cookieRef 改变时 token / user 同步更新', () => {
      const cookieTokenRef = Vue.ref<string | null>(null)
      const cookieUserRef = Vue.ref<string | null>(null)
      let i = 0
      ;(globalThis as any).useCookie = vi.fn((key: string) => {
        if (key === 'token') return cookieTokenRef
        if (key === 'user_info') return cookieUserRef
        return Vue.ref(null)
      })
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.token.value).toBe(null)
        expect(r.isLoggedIn.value).toBe(false)

        cookieTokenRef.value = 'late-tk'
        expect(r.token.value).toBe('late-tk')
        expect(r.isLoggedIn.value).toBe(true)

        const u = { id: 1, name: 'u' }
        cookieUserRef.value = encodeURIComponent(JSON.stringify(u))
        expect(r.user.value).toEqual(u)
      })
    })
  })
})