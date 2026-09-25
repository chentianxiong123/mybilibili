import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// ===== 在 import useAuth 之前注入 Nuxt 的 auto-import 全局变量 =====
// useAuth.ts 依赖 computed / useCookie（Nuxt auto-import）
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
    it('返回 user 与 isLoggedIn，不再暴露任何凭证字段', () => {
      const r = useAuth()
      expect(r).toHaveProperty('user')
      expect(r).toHaveProperty('isLoggedIn')
      // token / refresh_token 是 HttpOnly，SSR 一旦经 useCookie 读出来
      // 就会被序列化进页面 payload，等于把 HttpOnly 又拆了
      expect(r).not.toHaveProperty('token')
      expect(r).not.toHaveProperty('refreshToken')
    })

    it('只 useCookie 一次，且只读 user_info', () => {
      useAuth()
      expect((globalThis as any).useCookie).toHaveBeenCalledTimes(1)
      expect((globalThis as any).useCookie.mock.calls[0][0]).toBe('user_info')
    })
  })

  describe('cookie 路径', () => {
    it('user_info 解码失败时 user 为 null', () => {
      ;(globalThis as any).useCookie = vi.fn(() => Vue.ref('not-json'))
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.user.value).toBe(null)
      })
    })

    it('user_info 合法 JSON 时返回解码对象，isLoggedIn 为 true', () => {
      const userObj = { id: 5, username: 'alice' }
      ;(globalThis as any).useCookie = vi.fn((key: string) => {
        if (key === 'user_info') return Vue.ref(encodeURIComponent(JSON.stringify(userObj)))
        return Vue.ref(null)
      })
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.user.value).toEqual(userObj)
        expect(r.isLoggedIn.value).toBe(true)
      })
    })

    it('user 为空时 isLoggedIn 为 false', () => {
      ;(globalThis as any).useCookie = vi.fn(() => Vue.ref(null))
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.isLoggedIn.value).toBe(false)
      })
    })
  })

  describe('响应性', () => {
    it('user_info 变化时 user / isLoggedIn 同步更新', () => {
      const cookieUserRef = Vue.ref<string | null>(null)
      ;(globalThis as any).useCookie = vi.fn(() => cookieUserRef)
      vi.resetModules()
      return import('./useAuth').then(({ useAuth: freshAuth }) => {
        const r = freshAuth()
        expect(r.user.value).toBe(null)
        expect(r.isLoggedIn.value).toBe(false)

        const u = { id: 1, name: 'u' }
        cookieUserRef.value = encodeURIComponent(JSON.stringify(u))
        expect(r.user.value).toEqual(u)
        expect(r.isLoggedIn.value).toBe(true)
      })
    })
  })
})
