import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() }
}))

import api from '../client'
import { postComment, replyComment, likeComment, unlikeComment, getReplies } from '../comment'

const mockApi = vi.mocked(api)

describe('comment API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('postComment', () => {
    it('发表评论 → POST /comment/add (x-www-form-urlencoded)', async () => {
      mockApi.post.mockResolvedValue({ data: { id: 1 } })
      const result = await postComment(42, '好看')
      expect(result.code).toBe('1')
      expect(mockApi.post).toHaveBeenCalledWith(
        '/comment/add',
        'manuscriptId=42&content=%E5%A5%BD%E7%9C%8B',
        { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
      )
    })

    it('网络异常 → code 0', async () => {
      mockApi.post.mockRejectedValue(new Error('network'))
      const result = await postComment(1, 'test')
      expect(result.code).toBe('0')
    })
  })

  describe('replyComment', () => {
    it('回复评论 → POST /comment/reply', async () => {
      mockApi.post.mockResolvedValue({})
      await replyComment(10, '回复内容')
      expect(mockApi.post).toHaveBeenCalledWith(
        '/comment/reply',
        'commentId=10&content=%E5%9B%9E%E5%A4%8D%E5%86%85%E5%AE%B9',
        { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
      )
    })

    it('带 replyToUserId → 追加到 body', async () => {
      mockApi.post.mockResolvedValue({})
      await replyComment(10, '回复', 20)
      const body = mockApi.post.mock.calls[0][1] as string
      expect(body).toContain('replyToUserId=20')
    })
  })

  describe('likeComment', () => {
    it('点赞 → POST /comment/:id/like', async () => {
      mockApi.post.mockResolvedValue({})
      await likeComment(5)
      expect(mockApi.post).toHaveBeenCalledWith('/comment/5/like')
    })
  })

  describe('unlikeComment', () => {
    it('取消点赞 → DELETE /comment/:id/like', async () => {
      mockApi.delete = vi.fn().mockResolvedValue({})
      await unlikeComment(5)
      expect(mockApi.delete).toHaveBeenCalledWith('/comment/5/like')
    })
  })

  describe('getReplies', () => {
    it('获取回复列表 → GET /comment/:id/replies', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1 }] })
      const result = await getReplies(5, 2, 10)
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
      expect(mockApi.get).toHaveBeenCalledWith('/comment/5/replies?page=2&size=10')
    })

    it('默认分页参数', async () => {
      mockApi.get.mockResolvedValue({ data: [] })
      await getReplies(5)
      expect(mockApi.get).toHaveBeenCalledWith('/comment/5/replies?page=1&size=7')
    })
  })
})
