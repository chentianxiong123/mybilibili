import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import {
  getRoleList, getRoleById, addRole, updateRole, deleteRole,
  getRolePermissions, setRolePermissions, getRoleTemplates, applyRoleTemplate, getAllPermissions
} from './role'
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

describe('role api', () => {
  it('getRoleList GET /admin/roles', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getRoleList()
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/roles', method: 'get' })
  })

  it('getRoleById GET /admin/roles/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 3 } })
    await getRoleById(3)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/roles/3', method: 'get' })
  })

  it('addRole POST /admin/roles 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 1 } })
    await addRole({ name: '编辑', code: 'editor' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/roles',
      method: 'post',
      data: { name: '编辑', code: 'editor' },
    })
  })

  it('updateRole PUT /admin/roles/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateRole(3, { name: '改名' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/roles/3',
      method: 'put',
      data: { name: '改名' },
    })
  })

  it('deleteRole DELETE /admin/roles/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteRole(4)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/roles/4', method: 'delete' })
  })

  it('getRolePermissions GET /admin/roles/:id/permissions', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getRolePermissions(5)
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/roles/5/permissions', method: 'get' })
  })

  it('setRolePermissions PUT 带 permission_ids', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await setRolePermissions(5, [1, 2, 3])
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/roles/5/permissions',
      method: 'put',
      data: { permission_ids: [1, 2, 3] },
    })
  })

  it('getRoleTemplates GET /admin/roles/templates', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getRoleTemplates()
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/roles/templates', method: 'get' })
  })

  it('applyRoleTemplate PUT /admin/roles/:id/template/:code', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await applyRoleTemplate(5, 'editor')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/roles/5/template/editor',
      method: 'put',
    })
  })

  it('getAllPermissions GET /admin/permissions', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAllPermissions()
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/permissions', method: 'get' })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getRoleList()).rejects.toThrow('500')
  })
})
