import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getUserList, getUserById, updateUserStatus, resetPassword } from './user'
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

describe('user api', () => {
  it('getUserList GET /user/admin/list 带 params', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getUserList({ page: 1, keyword: 'a' })
    expect(requestMock.get).toHaveBeenCalledWith('/user/admin/list', { params: { page: 1, keyword: 'a' } })
  })

  it('getUserById GET /user/admin/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { id: 5 } })
    await getUserById(5)
    expect(requestMock.get).toHaveBeenCalledWith('/user/admin/5')
  })

  it('updateUserStatus PUT 带 status (number)', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await updateUserStatus(7, 1)
    expect(requestMock.put).toHaveBeenCalledWith('/user/admin/7/status', { status: 1 })
  })

  it('resetPassword PUT 带 newPassword', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await resetPassword(7, 'new-pwd')
    expect(requestMock.put).toHaveBeenCalledWith('/user/admin/7/password', { newPassword: 'new-pwd' })
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(getUserById(1)).rejects.toThrow('500')
  })
})
