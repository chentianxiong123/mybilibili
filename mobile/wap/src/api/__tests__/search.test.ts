import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn() }
}))

import api from '../client'
import { getHotwords, getSuggests, getSearchResult } from '../search'

const mockApi = vi.mocked(api)

describe('search API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getHotwords', () => {
    it('正常返回关键词列表', async () => {
      mockApi.get.mockResolvedValue({ data: ['AI', '前端', '后端'] })
      const result = await getHotwords()
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(3)
      expect(result.data.map((x: any) => x.keyword)).toEqual(['AI', '前端', '后端'])
    })

    it('过滤空字符串关键词', async () => {
      mockApi.get.mockResolvedValue({ data: ['AI', '', '  ', '后端'] })
      const result = await getHotwords()
      expect(result.code).toBe('1')
      expect(result.data.map((x: any) => x.keyword)).toEqual(['AI', '后端'])
    })

    it('字符串对象混排时拍平为字符串', async () => {
      mockApi.get.mockResolvedValue({ data: ['AI', { keyword: 'test' }, '后端'] })
      const result = await getHotwords()
      expect(result.code).toBe('1')
      expect(result.data.map((x: any) => x.keyword)).toContain('AI')
      expect(result.data.map((x: any) => x.keyword)).toContain('后端')
      expect(result.data.map((x: any) => x.keyword)).toContain('test')
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getHotwords()
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getSuggests', () => {
    it('正常返回', async () => {
      mockApi.get.mockResolvedValue({ data: ['suggest1', 'suggest2'] })
      const result = await getSuggests('keyword')
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(2)
      expect(result.data[0].name).toBe('suggest1')
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getSuggests('test')
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getSearchResult', () => {
    it('视频搜索（默认）→ /search/videos', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [] } })
      await getSearchResult({ keyword: '关键词', page: 1, size: 20 })
      expect(mockApi.get).toHaveBeenCalledWith(
        expect.stringContaining('/search/videos?keyword=')
      )
    })

    it('用户搜索 → /search/users', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [] } })
      await getSearchResult({ keyword: '用户', page: 1, size: 20, searchType: 'upuser' })
      expect(mockApi.get).toHaveBeenCalledWith(
        expect.stringContaining('/search/users?keyword=')
      )
    })

    it('空关键词 → 仍调接口（后端处理）', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [] } })
      await getSearchResult({ keyword: '', page: 1 })
      expect(mockApi.get).toHaveBeenCalled()
    })

    it('分页参数传递', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [] } })
      await getSearchResult({ keyword: 'keyword', page: 3, size: 10 })
      const calledUrl = mockApi.get.mock.calls[0][0] as string
      expect(calledUrl).toContain('page=3')
      expect(calledUrl).toContain('size=10')
    })

    it('排序参数：click → hot', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [] } })
      await getSearchResult({ keyword: 'k', page: 1, order: 'click' })
      const calledUrl = mockApi.get.mock.calls[0][0] as string
      expect(calledUrl).toContain('sort=hot')
    })

    it('排序参数：pubdate → time', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [] } })
      await getSearchResult({ keyword: 'k', page: 1, order: 'pubdate' })
      const calledUrl = mockApi.get.mock.calls[0][0] as string
      expect(calledUrl).toContain('sort=time')
    })
  })
})
