import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() }
}))

import api from '../client'
import { manuscriptApi, normalizeManuscriptList } from '../manuscript'

const mockApi = vi.mocked(api)

describe('manuscript API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('manuscriptApi', () => {
    it('getMyManuscripts → GET /manuscript/me/list', async () => {
      mockApi.get.mockResolvedValue({ data: { list: [{ id: 1 }] } })
      await manuscriptApi.getMyManuscripts({ page: 2 })
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/me/list', { params: { page: 2 } })
    })

    it('getMyStats → GET /manuscript/me/stats', async () => {
      mockApi.get.mockResolvedValue({ data: { total: 10 } })
      await manuscriptApi.getMyStats()
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/me/stats')
    })

    it('publish → POST /manuscript/:id/publish', async () => {
      mockApi.post.mockResolvedValue({})
      await manuscriptApi.publish(42)
      expect(mockApi.post).toHaveBeenCalledWith('/manuscript/42/publish')
    })

    it('unpublish → POST /manuscript/:id/unpublish', async () => {
      mockApi.post.mockResolvedValue({})
      await manuscriptApi.unpublish(42)
      expect(mockApi.post).toHaveBeenCalledWith('/manuscript/42/unpublish')
    })

    it('remove → DELETE /manuscript/:id', async () => {
      mockApi.delete = vi.fn().mockResolvedValue({})
      await manuscriptApi.remove(42)
      expect(mockApi.delete).toHaveBeenCalledWith('/manuscript/42')
    })
  })

  describe('normalizeManuscriptList', () => {
    it('标准列表格式', () => {
      const res = { data: { list: [{ id: 1 }], total: 10, page: 2, totalPages: 5 } }
      const result = normalizeManuscriptList(res)
      expect(result.list).toEqual([{ id: 1 }])
      expect(result.total).toBe(10)
      expect(result.page).toBe(2)
      expect(result.totalPages).toBe(5)
    })

    it('data 直接是数组', () => {
      const res = { data: [{ id: 1 }, { id: 2 }] }
      const result = normalizeManuscriptList(res)
      expect(result.list).toEqual([{ id: 1 }, { id: 2 }])
      expect(result.total).toBe(2)
      expect(result.page).toBe(1)
    })

    it('空数据 → 默认值', () => {
      const result = normalizeManuscriptList(null)
      expect(result.list).toEqual([])
      expect(result.total).toBe(0)
      expect(result.page).toBe(1)
      expect(result.totalPages).toBe(1)
    })

    it('data 有 list 但无 total → 取 list.length', () => {
      const res = { data: { list: [1, 2, 3] } }
      const result = normalizeManuscriptList(res)
      expect(result.total).toBe(3)
    })
  })
})
