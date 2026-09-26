import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAdminStore } from './admin'

vi.mock('@/api/admin', () => ({
  adminLogin: vi.fn(),
}))

vi.mock('@/api/session', () => ({
  clearServerSession: vi.fn(() => Promise.resolve()),
  default: vi.fn(() => Promise.resolve()),
}))

const routerPushSpy = vi.fn(() => Promise.resolve())
vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: routerPushSpy })
  }
})

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
  routerPushSpy.mockClear()
})

describe('admin store', () => {
  describe('初始状态', () => {
    it('无 localStorage 时使用默认空值', () => {
      const s = useAdminStore()
      expect(store.has('admin_token')).toBe(false)
      expect(s.userInfo).toBeNull()
      expect(s.role).toBe('')
      expect(s.permissions).toEqual([])
    })

    it('从 localStorage 恢复状态', () => {
      localStorage.setItem('admin_token', 'saved-token') // 升级前的明文遗留
      localStorage.setItem('admin_user', JSON.stringify({ id: 1, name: 'A' }))
      localStorage.setItem('admin_role', '管理员')
      localStorage.setItem('admin_permissions', JSON.stringify(['video.view']))
      setActivePinia(createPinia())

      const s = useAdminStore()
      // 展示/授权信息照常恢复，凭证不进 store 也不进 localStorage
      expect(s.userInfo).toEqual({ id: 1, name: 'A' })
      expect(s.role).toBe('管理员')
      expect(s.permissions).toEqual(['video.view'])
      // store 上已彻底没有 token 这一栏，遗留副本也不会被读出来
      expect('token' in s).toBe(false)
      expect(s).not.toHaveProperty('token')
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
      expect(store.has('admin_token')).toBe(false)
      expect(s.userInfo).toEqual({ id: 1, name: 'Admin' })
      expect(s.role).toBe('管理员')
      expect(s.permissions).toEqual(['video.review'])
      expect(localStorage.getItem('admin_token')).toBeNull()
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
      expect(store.has('admin_token')).toBe(false)
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
      expect(store.has('admin_token')).toBe(false)
      expect(s.userInfo).toBeNull()
    })

    it('登录失败 - 抛异常', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockRejectedValue(new Error('网络异常'))

      const s = useAdminStore()
      const res = await s.login({ username: 'admin', password: '123' })

      expect(res).toEqual({ success: false, message: '网络异常' })
      expect(store.has('admin_token')).toBe(false)
    })
  })

  describe('logout', () => {
    it('清空所有状态和 localStorage，并作废服务端 HttpOnly cookie', async () => {
      const { clearServerSession } = await import('@/api/session')
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        code: 200,
        data: { token: 'tok', adminUser: { id: 1 }, role: '管理员', permissions: ['a'] },
      })

      const s = useAdminStore()
      await s.login({ username: 'a', password: 'b' })
      vi.mocked(clearServerSession).mockClear()
      await s.logout()
      expect(clearServerSession).toHaveBeenCalledTimes(1)
    })

    it('清空所有状态和 localStorage', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        code: 200,
        data: { token: 'tok', adminUser: { id: 1 }, role: '管理员', permissions: ['a'] },
      })

      const s = useAdminStore()
      await s.login({ username: 'a', password: 'b' })
      expect(store.has('admin_token')).toBe(false)

      await s.logout()
      expect(store.has('admin_token')).toBe(false)
      expect(s.userInfo).toBeNull()
      expect(s.role).toBe('')
      expect(s.permissions).toEqual([])
      expect(localStorage.getItem('admin_token')).toBeNull()
      expect(localStorage.getItem('admin_user')).toBeNull()
      expect(localStorage.getItem('admin_role')).toBeNull()
      expect(localStorage.getItem('admin_permissions')).toBeNull()
      expect(localStorage.getItem('admin_id')).toBeNull()
    })

    it('logout 后自动跳转 /login（修复 logout 200 但页面不动的 UX bug）', async () => {
      const { adminLogin } = await import('@/api/admin')
      vi.mocked(adminLogin).mockResolvedValue({
        code: 200,
        data: { token: 'tok', adminUser: { id: 1 }, role: '管理员', permissions: ['a'] },
      })

      const s = useAdminStore()
      await s.login({ username: 'a', password: 'b' })
      await s.logout()
      expect(routerPushSpy).toHaveBeenCalledWith('/login')
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

// ====== 补充：嵌套对象 / 错误状态 / 状态隔离 ======

describe('admin store 补充 - 嵌套用户信息', () => {
  beforeEach(() => {
    store.clear()
    setActivePinia(createPinia())
    vi.clearAllMocks()
    routerPushSpy.mockClear()
  })

  it('登录响应含嵌套 adminUser 时正确提取 adminId', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      code: 200,
      data: {
        token: 'tok',
        adminUser: { id: 88, name: 'N', profile: { dept: 'tech' } },
        role: '管理员',
        permissions: ['video.review'],
      },
    })

    const s = useAdminStore()
    await s.login({ username: 'u', password: 'p' })

    expect(s.userInfo).toEqual({ id: 88, name: 'N', profile: { dept: 'tech' } })
    expect(localStorage.getItem('admin_id')).toBe('88')
  })

  it('登录响应无 role 时使用默认值 "管理员"', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      code: 200,
      data: {
        token: 'tok',
        adminUser: { id: 1 },
        permissions: [],
      },
    })

    const s = useAdminStore()
    await s.login({ username: 'u', password: 'p' })
    expect(s.role).toBe('管理员')
    expect(s.permissions).toEqual([])
  })

  it('登录响应无 adminUser 也无 user 时回退到 loginData.username', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      code: 200,
      data: { token: 'tok', permissions: [] },
    })

    const s = useAdminStore()
    await s.login({ username: 'fallback', password: 'p' })
    expect(s.userInfo).toEqual({ id: null, username: 'fallback' })
  })

  it('success=true 但无 token 时 token 为 undefined', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      success: true,
      user: { id: 1 },
      data: { permissions: [] },
    })

    const s = useAdminStore()
    const res = await s.login({ username: 'u', password: 'p' })
    expect(res).toEqual({ success: true })
    expect(store.has('admin_token')).toBe(false)
  })
})

describe('admin store 补充 - 错误状态', () => {
  beforeEach(() => {
    store.clear()
    setActivePinia(createPinia())
    vi.clearAllMocks()
    routerPushSpy.mockClear()
  })

  it('错误响应无 message 时使用默认 "登录失败"', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      code: 401,
    })

    const s = useAdminStore()
    const res = await s.login({ username: 'u', password: 'p' })
    expect(res).toEqual({ success: false, message: '登录失败' })
  })

  it('异常无 message 字段时使用默认 "登录失败"', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockRejectedValue({})

    const s = useAdminStore()
    const res = await s.login({ username: 'u', password: 'p' })
    expect(res).toEqual({ success: false, message: '登录失败' })
  })

  it('登录失败时不会写入 localStorage', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({ code: 401, message: 'no' })

    const s = useAdminStore()
    await s.login({ username: 'u', password: 'p' })

    expect(localStorage.getItem('admin_token')).toBeNull()
    expect(localStorage.getItem('admin_user')).toBeNull()
    expect(localStorage.getItem('admin_role')).toBeNull()
    expect(localStorage.getItem('admin_permissions')).toBeNull()
    expect(localStorage.getItem('admin_id')).toBeNull()
  })

  it('登录异常时不会写入 localStorage', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockRejectedValue(new Error('网络断开'))

    const s = useAdminStore()
    await s.login({ username: 'u', password: 'p' })

    expect(localStorage.getItem('admin_token')).toBeNull()
    expect(localStorage.getItem('admin_id')).toBeNull()
  })
})

describe('admin store 补充 - logout 状态隔离', () => {
  beforeEach(() => {
    store.clear()
    setActivePinia(createPinia())
    vi.clearAllMocks()
    routerPushSpy.mockClear()
  })

  it('logout 后再次 logout 是幂等的（不抛错）', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      code: 200,
      data: { token: 't', adminUser: { id: 1 }, role: '管理员', permissions: ['a'] },
    })

    const s = useAdminStore()
    await s.login({ username: 'u', password: 'p' })
    await s.logout()
    expect(() => { void s.logout() }).not.toThrow()
    expect(store.has('admin_token')).toBe(false)
    expect(s.userInfo).toBeNull()
  })

  it('未登录时 logout 不抛错', () => {
    const s = useAdminStore()
    expect(() => { void s.logout() }).not.toThrow()
    expect(store.has('admin_token')).toBe(false)
  })

  it('logout 后再次登录可重新建立完整状态', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin)
      .mockResolvedValueOnce({
        code: 200,
        data: { token: 'tok1', adminUser: { id: 1 }, role: '管理员', permissions: ['a'] },
      })
      .mockResolvedValueOnce({
        code: 200,
        data: { token: 'tok2', adminUser: { id: 2 }, role: '编辑', permissions: ['b'] },
      })

    const s = useAdminStore()
    await s.login({ username: 'u1', password: 'p' })
    await s.logout()
    const res = await s.login({ username: 'u2', password: 'p' })
    expect(res).toEqual({ success: true })
    expect(store.has('admin_token')).toBe(false)
    expect(s.role).toBe('编辑')
    expect(s.permissions).toEqual(['b'])
    expect(localStorage.getItem('admin_token')).toBeNull()
    expect(localStorage.getItem('admin_id')).toBe('2')
  })
})

describe('admin store 补充 - hasPermission 边界', () => {
  beforeEach(() => {
    store.clear()
    setActivePinia(createPinia())
    vi.clearAllMocks()
    routerPushSpy.mockClear()
  })

  it('permissions 数组包含 falsy 值不影响判断', () => {
    const s = useAdminStore()
    s.role = '管理员'
    s.permissions = ['a', '', null as any, undefined as any]
    expect(s.hasPermission('a')).toBe(true)
    expect(s.hasPermission('b')).toBe(false)
  })

  it('adminUser.id 为 0 时因 falsy 不写入 admin_id', async () => {
    const { adminLogin } = await import('@/api/admin')
    vi.mocked(adminLogin).mockResolvedValue({
      code: 200,
      data: {
        token: 'tok',
        adminUser: { id: 0 },
        role: '管理员',
        permissions: [],
      },
    })

    const s = useAdminStore()
    await s.login({ username: 'u', password: 'p' })
    expect(localStorage.getItem('admin_id')).toBeNull()
  })
})
