import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAdminStore } from './admin'

vi.mock('@/api/admin', () => ({
  adminLogin: vi.fn(),
}))

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
  setActivePinia(createPinia())
  vi.clearAllMocks()
})

describe('admin store', () => {
  describe('初始状态', () => {
    it('无 localStorage 时使用默认空值', () => {
      const s = useAdminStore()
      expect(s.token).toBe('')
      expect(s.userInfo).toBeNull()
      expect(s.role).toBe('')
      expect(s.permissions).toEqual([])
    })

    it('从 localStorage 恢复状态', () => {
      localStorage.setItem('admin_token', 'saved-token')
      localStorage.setItem('admin_user', JSON.stringify({ id: 1, name: 'A' }))
      localStorage.setItem('admin_role', '管理员')
      localStorage.setItem('admin_permissions', JSON.stringify(['video.view']))
      setActivePinia(createPinia())

      const s = useAdminStore()
      expect(s.token).toBe('saved-token')
      expect(s.userInfo).toEqual({ id: 1, name: 'A' })
      expect(s.role).toBe('管理员')
      expect(s.permissions).toEqual(['video.view'])
    })
  })

  describe('login', () => {
    it('成功登录 - code=200', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        code: 200,
        data: {
          token: 'tok-123',
          adminUser: { id: 1, name: 'Admin' },
          role: '管理员',
          permissions: ['video.review'],
        },
      })

      const s = useAdminStore()
      const res = await s.login({ username: 'admin', password: '123' })

      expect(res).toEqual({ success: true })
      expect(s.token).toBe('tok-123')
      expect(s.userInfo).toEqual({ id: 1, name: 'Admin' })
      expect(s.role).toBe('管理员')
      expect(s.permissions).toEqual(['video.review'])
      expect(localStorage.getItem('admin_token')).toBe('tok-123')
      expect(JSON.parse(localStorage.getItem('admin_user')!)).toEqual({ id: 1, name: 'Admin' })
      expect(localStorage.getItem('admin_role')).toBe('管理员')
      expect(JSON.parse(localStorage.getItem('admin_permissions')!)).toEqual(['video.review'])
      expect(localStorage.getItem('admin_id')).toBe('1')
    })

    it('成功登录 - success=true 格式', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        success: true,
        token: 'tok-alt',
        user: { id: 2, name: 'B' },
        data: { role: '编辑', permissions: ['edit'] },
      })

      const s = useAdminStore()
      const res = await s.login({ username: 'editor', password: 'pwd' })

      expect(res).toEqual({ success: true })
      expect(s.token).toBe('tok-alt')
      expect(s.userInfo).toEqual({ id: 2, name: 'B' })
      expect(s.role).toBe('编辑')
      expect(s.permissions).toEqual(['edit'])
    })

    it('登录失败 - 接口返回失败', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        code: 401,
        message: '密码错误',
      })

      const s = useAdminStore()
      const res = await s.login({ username: 'admin', password: 'wrong' })

      expect(res).toEqual({ success: false, message: '密码错误' })
      expect(s.token).toBe('')
      expect(s.userInfo).toBeNull()
    })

    it('登录失败 - 抛异常', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockRejectedValue(new Error('网络异常'))

      const s = useAdminStore()
      const res = await s.login({ username: 'admin', password: '123' })

      expect(res).toEqual({ success: false, message: '网络异常' })
      expect(s.token).toBe('')
    })
  })

  describe('logout', () => {
    it('清空所有状态和 localStorage', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        code: 200,
        data: { token: 'tok', adminUser: { id: 1 }, role: '管理员', permissions: ['a'] },
      })

      const s = useAdminStore()
      await s.login({ username: 'a', password: 'b' })
      expect(s.token).toBe('tok')

      s.logout()
      expect(s.token).toBe('')
      expect(s.userInfo).toBeNull()
      expect(s.role).toBe('')
      expect(s.permissions).toEqual([])
      expect(localStorage.getItem('admin_token')).toBeNull()
      expect(localStorage.getItem('admin_user')).toBeNull()
      expect(localStorage.getItem('admin_role')).toBeNull()
      expect(localStorage.getItem('admin_permissions')).toBeNull()
      expect(localStorage.getItem('admin_id')).toBeNull()
    })
  })

  describe('hasPermission', () => {
    it('permission 为空时返回 true', () => {
      const s = useAdminStore()
      expect(s.hasPermission('')).toBe(true)
      expect(s.hasPermission(null as any)).toBe(true)
      expect(s.hasPermission(undefined as any)).toBe(true)
    })

    it('超级管理员始终返回 true', () => {
      const s = useAdminStore()
      s.role = '超级管理员'
      expect(s.hasPermission('any.permission')).toBe(true)
      expect(s.hasPermission('')).toBe(true)
    })

    it('普通角色检查 permissions 数组', () => {
      const s = useAdminStore()
      s.role = '管理员'
      s.permissions = ['video.review', 'video.edit']

      expect(s.hasPermission('video.review')).toBe(true)
      expect(s.hasPermission('video.edit')).toBe(true)
      expect(s.hasPermission('user.delete')).toBe(false)
    })

    it('permissions 为空时非超管无权限', () => {
      const s = useAdminStore()
      s.role = '管理员'
      s.permissions = []

      expect(s.hasPermission('anything')).toBe(false)
    })
  })
})
