import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() }
}))

import api from '../client'
import { dynamicApi } from '../dynamic'

const mockApi = vi.mocked(api)

describe('dynamic API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getDynamicList → GET /dynamic/list', async () => {
    mockApi.get.mockResolvedValue({ data: { list: [] } })
    await dynamicApi.getDynamicList(1, 10)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/list', { params: { page: 1, size: 10 } })
  })

  it('getFollowingDynamics → GET /dynamic/following', async () => {
    mockApi.get.mockResolvedValue({ data: [] })
    await dynamicApi.getFollowingDynamics(1, 10, 42)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/following', { params: { page: 1, size: 10, userId: 42 } })
  })

  it('getFollowingDynamics 无 userId → 不传 userId', async () => {
    mockApi.get.mockResolvedValue({ data: [] })
    await dynamicApi.getFollowingDynamics(1, 10)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/following', { params: { page: 1, size: 10 } })
  })

  it('getUserDynamics → GET /dynamic/user/:id', async () => {
    mockApi.get.mockResolvedValue({ data: [] })
    await dynamicApi.getUserDynamics(42, 2, 5)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/user/42', { params: { page: 2, limit: 5 } })
  })

  it('getDynamicById → GET /dynamic/:id', async () => {
    mockApi.get.mockResolvedValue({ data: { id: 1 } })
    await dynamicApi.getDynamicById(1)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/1')
  })

  it('publishDynamic → POST /dynamic/publish with FormData', async () => {
    mockApi.post.mockResolvedValue({})
    const fd = new FormData()
    await dynamicApi.publishDynamic(fd)
    expect(mockApi.post).toHaveBeenCalledWith('/dynamic/publish', fd, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  })

  it('deleteDynamic → DELETE /dynamic/:id', async () => {
    mockApi.delete = vi.fn().mockResolvedValue({})
    await dynamicApi.deleteDynamic(1)
    expect(mockApi.delete).toHaveBeenCalledWith('/dynamic/1')
  })

  it('likeDynamic → POST /dynamic/like/:id', async () => {
    mockApi.post.mockResolvedValue({})
    await dynamicApi.likeDynamic(1)
    expect(mockApi.post).toHaveBeenCalledWith('/dynamic/like/1')
  })

  it('unlikeDynamic → DELETE /dynamic/like/:id', async () => {
    mockApi.delete = vi.fn().mockResolvedValue({})
    await dynamicApi.unlikeDynamic(1)
    expect(mockApi.delete).toHaveBeenCalledWith('/dynamic/like/1')
  })

  it('checkLikeStatus → GET /dynamic/like/status/:id', async () => {
    mockApi.get.mockResolvedValue({ data: { liked: true } })
    await dynamicApi.checkLikeStatus(1)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/like/status/1')
  })

  it('shareDynamic → POST /dynamic/share/:id', async () => {
    mockApi.post.mockResolvedValue({})
    await dynamicApi.shareDynamic(1)
    expect(mockApi.post).toHaveBeenCalledWith('/dynamic/share/1')
  })

  it('getComments → GET /dynamic/comment/list', async () => {
    mockApi.get.mockResolvedValue({ data: [] })
    await dynamicApi.getComments(1, 2, 10)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/comment/list', {
      params: { dynamicId: 1, page: 2, size: 10 }
    })
  })

  it('getReplies → GET /dynamic/comment/replies', async () => {
    mockApi.get.mockResolvedValue({ data: [] })
    await dynamicApi.getReplies(1, 2, 20)
    expect(mockApi.get).toHaveBeenCalledWith('/dynamic/comment/replies', {
      params: { commentId: 1, page: 2, size: 20 }
    })
  })

  it('addComment → POST /dynamic/comment/add', async () => {
    mockApi.post.mockResolvedValue({})
    await dynamicApi.addComment(1, '好看')
    expect(mockApi.post).toHaveBeenCalledWith('/dynamic/comment/add', null, {
      params: { dynamicId: 1, content: '好看' }
    })
  })

  it('addComment 有 parentId 和 replyUserId → 带上', async () => {
    mockApi.post.mockResolvedValue({})
    await dynamicApi.addComment(1, '回复', 10, 20)
    expect(mockApi.post).toHaveBeenCalledWith('/dynamic/comment/add', null, {
      params: { dynamicId: 1, content: '回复', parentId: 10, replyUserId: 20 }
    })
  })

  it('deleteComment → DELETE /dynamic/comment/delete/:id', async () => {
    mockApi.delete = vi.fn().mockResolvedValue({})
    await dynamicApi.deleteComment(1)
    expect(mockApi.delete).toHaveBeenCalledWith('/dynamic/comment/delete/1')
  })

  it('likeComment → POST /dynamic/comment/like/:id', async () => {
    mockApi.post.mockResolvedValue({})
    await dynamicApi.likeComment(1)
    expect(mockApi.post).toHaveBeenCalledWith('/dynamic/comment/like/1')
  })

  it('unlikeComment → DELETE /dynamic/comment/like/:id', async () => {
    mockApi.delete = vi.fn().mockResolvedValue({})
    await dynamicApi.unlikeComment(1)
    expect(mockApi.delete).toHaveBeenCalledWith('/dynamic/comment/like/1')
  })
})
