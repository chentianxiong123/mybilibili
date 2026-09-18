import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { adminLogin, adminRegister, getAdminList, getAdminById, getAdminRoles, setAdminRoles, updateAdmin } from './admin'
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

describe('admin api', () => {
  it('adminLogin POST /admin/login 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { token: 't' } })
    const data = { username: 'u', password: 'p' }
    await adminLogin(data)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/login', method: 'post', data })
  })

  it('adminRegister POST /admin/register', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 1 } })
    const data = { username: 'new', password: 'pwd' }
    await adminRegister(data)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/register', method: 'post', data })
  })

  it('getAdminList GET /admin/list', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAdminList()
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/list', method: 'get' })
  })

  it('getAdminById GET /admin/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 5 } })
    await getAdminById(5)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/5', method: 'get' })
  })

  it('getAdminRoles GET /admin/:id/roles', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAdminRoles(3)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/3/roles', method: 'get' })
  })

  it('setAdminRoles PUT /admin/:id/roles 带 role_ids 字段', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await setAdminRoles(3, [1, 2, 3])
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/3/roles',
      method: 'put',
      data: { role_ids: [1, 2, 3] },
    })
  })

  it('updateAdmin PUT /admin/:id 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    const data = { username: 'x', email: 'x@y.z' }
    await updateAdmin(9, data)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/9', method: 'put', data })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(adminLogin({ username: 'a', password: 'b' })).rejects.toThrow('500')
  })
})
