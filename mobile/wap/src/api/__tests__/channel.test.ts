import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn() }
}))

import api from '../client'
import { getRankingPartitions, getRankingRegion } from '../channel'

const mockApi = vi.mocked(api)

describe('channel API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getRankingPartitions', () => {
    it('返回分区列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1, name: '动画' }] })
      const result = await getRankingPartitions()
      expect(result.data).toEqual([{ id: 1, name: '动画' }])
    })
  })

  describe('getRankingRegion', () => {
    it('请求指定分区的排行数据', async () => {
      mockApi.get
        .mockResolvedValueOnce({ data: [{ id: 1 }] })  // hot
        .mockResolvedValueOnce({ data: [{ id: 2 }] })  // category
      const result = await getRankingRegion(1)
      expect(result).toBeDefined()
      expect(mockApi.get).toHaveBeenCalledTimes(2)
    })
  })
})
