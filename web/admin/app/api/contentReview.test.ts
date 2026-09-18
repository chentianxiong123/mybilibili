import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getPendingList, getAllContent, restoreContent, deleteContent, batchProcess } from './contentReview'
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

describe('contentReview api', () => {
  it('getPendingList GET /moderation/admin/pending 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { records: [], total: 0 } })
    await getPendingList({ page: 1, size: 10 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/pending',
      method: 'get',
      params: { page: 1, size: 10 },
    })
  })

  it('getAllContent GET /moderation/admin/all 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAllContent({ status: 'approved' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/all',
      method: 'get',
      params: { status: 'approved' },
    })
  })

  it('restoreContent PUT /moderation/admin/restore/:type/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await restoreContent('VIDEO', 7)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/restore/VIDEO/7',
      method: 'put',
    })
  })

  it('deleteContent DELETE /moderation/admin/:type/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteContent('COMMENT', 9)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/COMMENT/9',
      method: 'delete',
    })
  })

  it('batchProcess POST /moderation/admin/batch 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    const data = { ids: [1, 2, 3], action: 'approve' }
    await batchProcess(data)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/batch',
      method: 'post',
      data,
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getPendingList({})).rejects.toThrow('500')
  })
})
