import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getAdminComments, deleteAdminComment, restoreAdminComment, getAdminDanmaku, deleteAdminDanmaku } from './adminContent'
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

describe('adminContent api', () => {
  it('getAdminComments GET /moderation/admin/comments 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { records: [], total: 0 } })
    await getAdminComments({ page: 1, size: 20 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/comments',
      method: 'get',
      params: { page: 1, size: 20 },
    })
  })

  it('deleteAdminComment DELETE /moderation/admin/comments/:type/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteAdminComment('VIDEO', 7)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/comments/VIDEO/7',
      method: 'delete',
    })
  })

  it('restoreAdminComment PUT /moderation/admin/comments/:type/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await restoreAdminComment('DYNAMIC', 9)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/comments/DYNAMIC/9',
      method: 'put',
    })
  })

  it('getAdminDanmaku GET /moderation/admin/danmaku', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAdminDanmaku({ videoId: 1 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/danmaku',
      method: 'get',
      params: { videoId: 1 },
    })
  })

  it('deleteAdminDanmaku DELETE /moderation/admin/danmaku/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteAdminDanmaku(11)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/danmaku/11',
      method: 'delete',
    })
  })
})
