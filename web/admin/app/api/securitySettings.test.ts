import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getSecuritySettings, updateSecuritySettings, adminLoginLogApi } from './securitySettings'
import request from '@/api/client'

const requestMock = request as unknown as ReturnType<typeof vi.fn> & {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  requestMock.mockReset()
  requestMock.get.mockReset()
  requestMock.post.mockReset()
  requestMock.put.mockReset()
  requestMock.delete.mockReset()
})

describe('securitySettings api', () => {
  it('getSecuritySettings GET /admin/security-settings', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { pwdMinLen: 8 } })
    const res = await getSecuritySettings()
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/security-settings', method: 'get' })
    expect(res.data.pwdMinLen).toBe(8)
  })

  it('updateSecuritySettings PUT 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateSecuritySettings({ pwdMinLen: 12 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/security-settings',
      method: 'put',
      data: { pwdMinLen: 12 },
    })
  })

  it('adminLoginLogApi.getLoginLogs GET 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await adminLoginLogApi.getLoginLogs({ page: 1, size: 20 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/login-logs/list',
      method: 'get',
      params: { page: 1, size: 20 },
    })
  })

  it('adminLoginLogApi.getUserLoginLogs GET /user/:id 带 page/size', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await adminLoginLogApi.getUserLoginLogs(7, 2, 50)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/login-logs/user/7',
      method: 'get',
      params: { page: 2, size: 50 },
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getSecuritySettings()).rejects.toThrow('500')
  })
})
