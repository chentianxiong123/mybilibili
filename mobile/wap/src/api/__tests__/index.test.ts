import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() }
}))

vi.mock('../../utils/cache', () => ({
  readCache: vi.fn().mockReturnValue(null),
  writeCache: vi.fn()
}))

import api from '../client'
import { getHomeContent, getCachedHome, getBanners, getCategories, getVideosByCategory, getHotVideos, getRankingArchive, submitFeedback } from '../index'

const mockApi = vi.mocked(api)

describe('home index API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getHomeContent', () => {
    it('并行请求推荐 + 分类 + 热门，返回完整首页数据', async () => {
      mockApi.get
        .mockResolvedValueOnce({ data: [{ id: 1, title: '推荐' }] })  // recommended
        .mockResolvedValueOnce({ data: [{ id: 10, name: '动画' }] })  // category
        .mockResolvedValueOnce({ data: [{ id: 2, title: '热门' }] })  // hot

      const result = await getHomeContent()
      expect(result.code).toBe('1')
      expect(result.data.oneLevelPartitions.length).toBe(1)
      expect(result.data.additionalVideos.length).toBe(1)
      expect(result.data.rankingVideos.length).toBe(1)
    })

    it('部分请求失败 → 降级返回已有数据', async () => {
      mockApi.get
        .mockRejectedValueOnce(new Error('fail'))  // recommended fails
        .mockResolvedValueOnce({ data: [{ id: 10, name: '动画' }] })
        .mockResolvedValueOnce({ data: [] })

      const result = await getHomeContent()
      expect(result.code).toBe('1')
      expect(result.data.additionalVideos).toEqual([])
    })
  })

  describe('getCachedHome', () => {
    it('无缓存 → null', async () => {
      const result = getCachedHome()
      expect(result).toBeNull()
    })
  })

  describe('getBanners', () => {
    it('返回轮播图列表', async () => {
      mockApi.get.mockResolvedValue({
        data: [{ id: 1, title: 'Banner', imageUrl: 'img.jpg', linkUrl: 'http://example.com' }]
      })
      const result = await getBanners()
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
      expect(result.data[0].pic).toBe('img.jpg')
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getBanners()
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getCategories', () => {
    it('返回分类列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1, name: '动画' }] })
      const result = await getCategories()
      expect(result.code).toBe('1')
      expect(result.data[0]).toEqual({ id: 1, name: '动画' })
    })
  })

  describe('getVideosByCategory', () => {
    it('返回分类下的视频列表', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [{ id: 1 }] } })
      const result = await getVideosByCategory(1)
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
    })
  })

  describe('getHotVideos', () => {
    it('返回热门视频', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1 }] })
      const result = await getHotVideos()
      expect(result.code).toBe('1')
    })
  })

  describe('getRankingArchive', () => {
    it('rId > 0 → 分类排行', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1 }] })
      await getRankingArchive({ rId: 1, p: 1, size: 20 })
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/category/1?page=1&size=20')
    })

    it('rId = 0 → 全站热门（offset 分页）', async () => {
      mockApi.get.mockResolvedValue({ data: [] })
      await getRankingArchive({ rId: 0, p: 2, size: 20 })
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/hot?offset=20')
    })
  })

  describe('submitFeedback', () => {
    it('提交反馈', async () => {
      mockApi.post.mockResolvedValue({ data: { id: 1 } })
      const result = await submitFeedback({ type: 'bug', content: '有问题', contact: 'test@test.com' })
      expect(result.code).toBe('1')
      expect(mockApi.post).toHaveBeenCalledWith('/operation/tickets', expect.objectContaining({
        title: 'bug'
      }))
    })

    it('无联系方式 → content 不拼接', async () => {
      mockApi.post.mockResolvedValue({})
      await submitFeedback({ type: 'suggestion', content: '建议', contact: '' })
      const body = mockApi.post.mock.calls[0][1] as any
      expect(body.content).toBe('建议')
    })
  })
})
