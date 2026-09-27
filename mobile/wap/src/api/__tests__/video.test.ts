import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn() }
}))

vi.mock('../../utils/cache', () => ({
  readCache: vi.fn().mockReturnValue(null),
  writeCache: vi.fn()
}))

import api from '../client'
import { getVideoInfo, getRecommendVides, getComments, getBarrages, sendBarrage } from '../video'

const mockApi = vi.mocked(api)

describe('video API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getVideoInfo', () => {
    it('请求视频详情', async () => {
      mockApi.get.mockResolvedValue({
        data: { id: 1, title: '测试视频', videos: [{ playUrl: 'url' }] }
      })
      const result = await getVideoInfo(1)
      expect(result.code).toBe('1')
      expect(result.data.title).toBe('测试视频')
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/1')
    })

    it('空响应 → code 0', async () => {
      mockApi.get.mockResolvedValue(null)
      const result = await getVideoInfo(1)
      expect(result.code).toBe('0')
    })

    it('网络异常 → 返回缓存（如有）', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getVideoInfo(1)
      expect(result.code).toBe('0')
    })
  })

  describe('getRecommendVides', () => {
    it('请求推荐视频', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 2, title: '推荐' }] })
      const result = await getRecommendVides(1)
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
      expect(mockApi.get).toHaveBeenCalledWith('/recommend/related/1?size=8')
    })
  })

  describe('getComments', () => {
    it('请求评论列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1, content: '好看' }] })
      const result = await getComments(1)
      expect(result.code).toBe('1')
      expect(result.data[0].content).toBe('好看')
    })
  })

  describe('getBarrages', () => {
    it('请求弹幕列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ content: '666', color: '#fff', time: 10 }] })
      const result = await getBarrages(123)
      expect(result.code).toBe('1')
      expect(result.data[0].text).toBe('666')
      expect(mockApi.get).toHaveBeenCalledWith('/danmaku/video/123')
    })
  })

  describe('sendBarrage', () => {
    it('发送弹幕', async () => {
      mockApi.post.mockResolvedValue({ data: { id: 1 } })
      const result = await sendBarrage(123, 456, '好看', 10, '#ffffff', 0)
      expect(result.code).toBe('1')
      expect(mockApi.post).toHaveBeenCalledWith('/danmaku/send', {
        video_id: 123,
        manuscript_id: 456,
        content: '好看',
        time: 10,
        color: '#ffffff',
        mode: 0
      })
    })
  })
})
