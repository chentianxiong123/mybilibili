import { describe, it, expect, vi, beforeEach } from 'vitest'

// 捕获 axios.create 返回实例的 request/response 拦截器，直接驱动 handler 做断言
const captured = vi.hoisted(() => ({
  requestHandlers: [] as any[],
  responseHandlers: [] as any[],
}))

const { apiCallable, apiMethods } = vi.hoisted(() => {
  const apiCallable = vi.fn(() => Promise.resolve({ code: 200, data: {} }))
  const apiMethods = {
    get: vi.fn(() => Promise.resolve({ code: 200, data: {} })),
    post: vi.fn(() => Promise.resolve({ code: 200, data: {} })),
    put: vi.fn(() => Promise.resolve({ code: 200, data: {} })),
    delete: vi.fn(() => Promise.resolve({ code: 200, data: {} })),
    request: vi.fn(() => Promise.resolve({ code: 200, data: {} })),
  }
  return { apiCallable, apiMethods }
})

vi.mock('axios', () => ({
  default: {
    create: () => {
      const instance: any = Object.assign(apiCallable, apiMethods, {
        interceptors: {
          request: { use: (h: any) => captured.requestHandlers.push(h) },
          response: { use: (h: any, e: any) => captured.responseHandlers.push({ ok: h, err: e }) },
        },
      })
      return instance
    },
  },
}))

// mock 依赖，避免真实 localStorage / element-plus / fetch
vi.mock('../utils/safeStorage', () => {
  const m = new Map<string, string>()
  return {
    safeStorage: {
      getItem: (k: string) => (m.has(k) ? m.get(k) : null),
      setItem: (k: string, v: string) => { m.set(k, String(v)) },
      removeItem: (k: string) => { m.delete(k) },
    },
  }
})

vi.mock('element-plus', () => ({
  ElMessage: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('./session', () => ({
  clearServerSession: vi.fn(() => Promise.resolve()),
  default: vi.fn(() => Promise.resolve()),
}))

import request, {
  getUserList, updateUserStatus, resetPassword,
  adminLoginLogApi, videoApi, commentApi, interactionApi, historyApi,
} from './client'
import { hasAuthSession, hasAdminSession, clearAuthSession, clearAdminSession } from '../utils/auth'

let reqOkHandler: (r: any) => any
let resOkHandler: (r: any) => any
let resErrHandler: (e: any) => any
reqOkHandler = captured.requestHandlers[0]
resOkHandler = captured.responseHandlers[0].ok
resErrHandler = captured.responseHandlers[0].err

beforeEach(() => {
  vi.clearAllMocks()
  clearAuthSession()
  clearAdminSession()
})

describe('client.ts 响应归一化（response 拦截器）', () => {
  const runOk = (raw: any) => resOkHandler({ config: { url: '/x', method: 'get' }, data: raw })

  it('信封结构 {code,data} 原样返回', () => {
    const out = runOk({ code: 200, data: { list: [1] } })
    expect(out).toEqual({ code: 200, data: { list: [1] } })
  })

  it('裸数组 → {code:200,data:[...]}', () => {
    expect(runOk([1, 2])).toEqual({ code: 200, data: [1, 2] })
  })

  it('null → {code:200,data:null}', () => {
    expect(runOk(null)).toEqual({ code: 200, data: null })
  })

  it('undefined → {code:200,data:null}', () => {
    expect(runOk(undefined)).toEqual({ code: 200, data: null })
  })

  it('裸对象 → {code:200,data:{...}}', () => {
    expect(runOk({ foo: 'bar' })).toEqual({ code: 200, data: { foo: 'bar' } })
  })

  it('GET 命中 CACHE_PATHS 时写入缓存（第二次走 adapter）', () => {
    // 手动触发请求拦截器：非图片 GET 且命中缓存路径时不应立即拦截（首次）
    const cfg1: any = { url: '/category', method: 'get', params: undefined }
    reqOkHandler(cfg1)
    expect(cfg1.adapter).toBeUndefined()

    // 触发响应成功写入缓存
    resOkHandler({ config: { url: '/category', method: 'get', params: undefined }, data: { code: 200, data: [1] } })

    // 第二次请求 → adapter 被注入（走缓存）
    const cfg2: any = { url: '/category', method: 'get', params: undefined }
    reqOkHandler(cfg2)
    expect(typeof cfg2.adapter).toBe('function')
  })
})

describe('client.ts 401 处理（错误拦截器）', () => {
  const makeErr = (over = {}) => ({
    config: { url: '/admin/x', method: 'get' },
    response: { status: 401, data: { message: 'unauthorized' } },
    ...over,
  })

  it('非 401 错误 → ElMessage 提示并 reject', async () => {
    const e = { config: { url: '/x', method: 'get' }, response: { status: 500, data: {} } }
    await expect(resErrHandler(e)).rejects.toBe(e)
  })

  it('401 且无 admin session → 走用户 401 分支，无 admin session 直接 reject', async () => {
    const e = makeErr()
    await expect(resErrHandler(e)).rejects.toBe(e)
  })

  it('401 + admin session + 刷新成功 → 重试原请求并 resolve', async () => {
    const { setAdminSession } = await import('../utils/auth')
    setAdminSession({ user: { id: 1, name: 'A' }, role: 'admin', permissions: [] })
    const retried: any = { code: 200, data: { list: [] } }
    apiCallable.mockResolvedValueOnce(retried)
    apiMethods.post.mockResolvedValueOnce({ code: 200, data: { token: 'new' } })

    const out = await resErrHandler(makeErr())
    expect(apiMethods.post).toHaveBeenCalledWith('/admin/token/refresh', {})
    expect(out).toBe(retried)
  })

  it('401 + admin session + 刷新失败 → 清 session、调服务端登出、跳 /admin/login', async () => {
    const { setAdminSession, hasAdminSession } = await import('../utils/auth')
    setAdminSession({ user: { id: 1, name: 'A' }, role: 'admin', permissions: [] })
    apiMethods.post.mockRejectedValueOnce(new Error('refresh failed'))

    await expect(resErrHandler(makeErr())).rejects.toMatchObject({ response: { status: 401 } })
    expect(hasAdminSession()).toBe(false)
  })

  it('401 + admin session + 已重试过(_adminRefreshed) → 直接作废', async () => {
    const { setAdminSession, hasAdminSession } = await import('../utils/auth')
    setAdminSession({ user: { id: 1, name: 'A' }, role: 'admin', permissions: [] })
    const e = makeErr({ config: { url: '/admin/x', method: 'get', _adminRefreshed: true } })

    await expect(resErrHandler(e)).rejects.toBe(e)
    expect(hasAdminSession()).toBe(false)
  })

  it('401 + admin session + 自身是 refresh 接口 → 直接作废（不递归刷新）', async () => {
    // 这个 case 修了一个 deadlock：refresh 自己 401 时原代码会把自己
    // 入 failedQueue，api.post() 的 Promise 永远 pending，
    // endAdminSession / processQueue / isRefreshing=false 全部不会触发，
    // 用户卡死在原页、isRefreshing 锁死 → 后续请求全排队等死。
    const { setAdminSession, hasAdminSession } = await import('../utils/auth')
    setAdminSession({ user: { id: 1, name: 'A' }, role: 'admin', permissions: [] })
    const e = makeErr({ config: { url: '/admin/token/refresh', method: 'post' } })

    await expect(resErrHandler(e)).rejects.toBe(e)
    expect(hasAdminSession()).toBe(false)
    expect(apiMethods.post).not.toHaveBeenCalledWith('/admin/token/refresh', {})
  })
})

describe('client.ts 导出 API 方法', () => {
  it('getUserList → GET /user/admin/list 携带 params', () => {
    getUserList({ page: 1, size: 20 })
    expect(apiMethods.get).toHaveBeenCalledWith('/user/admin/list', { params: { page: 1, size: 20 } })
  })

  it('updateUserStatus → PUT /user/admin/{id}/status', () => {
    updateUserStatus(5, 'blocked')
    expect(apiMethods.put).toHaveBeenCalledWith('/user/admin/5/status', { status: 'blocked' })
  })

  it('resetPassword → PUT /user/admin/{id}/password', () => {
    resetPassword(5, 'newpass')
    expect(apiMethods.put).toHaveBeenCalledWith('/user/admin/5/password', { newPassword: 'newpass' })
  })

  it('adminLoginLogApi.getLoginLogs → GET /admin/login-logs/list', () => {
    adminLoginLogApi.getLoginLogs({ page: 1 })
    expect(apiMethods.get).toHaveBeenCalledWith('/admin/login-logs/list', { params: { page: 1 } })
  })

  it('videoApi.getVideoById → GET /manuscript/{id}', () => {
    videoApi.getVideoById(9)
    expect(apiMethods.get).toHaveBeenCalledWith('/manuscript/9')
  })

  it('commentApi.getComments → GET /comment/list (非 DYNAMIC)', () => {
    commentApi.getComments('VIDEO', 3, 2, 10, 'time')
    expect(apiMethods.get).toHaveBeenCalledWith('/comment/list?manuscriptId=3&page=2&size=10&sort=time')
  })

  it('commentApi.getComments → GET /dynamic/comment/list (DYNAMIC)', () => {
    commentApi.getComments('DYNAMIC', 7, 1, 20, 'time')
    expect(apiMethods.get).toHaveBeenCalledWith('/dynamic/comment/list?dynamicId=7&page=1&size=20&sort=time')
  })

  it('interactionApi.likeManuscript → POST /manuscript/{id}/like', () => {
    interactionApi.likeManuscript(3, true)
    expect(apiMethods.post).toHaveBeenCalledWith('/manuscript/3/like')
  })

  it('interactionApi.likeManuscript → DELETE /manuscript/{id}/like (取消)', () => {
    interactionApi.likeManuscript(3, false)
    expect(apiMethods.delete).toHaveBeenCalledWith('/manuscript/3/like')
  })

  it('historyApi.getHistoryList 未登录 → resolve {code:401}', async () => {
    const res = await historyApi.getHistoryList()
    expect(res.code).toBe(401)
  })
})