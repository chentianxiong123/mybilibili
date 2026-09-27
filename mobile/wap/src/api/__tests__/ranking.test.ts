import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn() }
}))

import api from '../client'
import { getRankings, getRankingArchive } from '../ranking'

const mockApi = vi.mocked(api)

describe('ranking API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getRankings → GET /category', async () => {
    mockApi.get.mockResolvedValue({ data: [{ id: 1, name: '动画' }] })
    const result = await getRankings()
    expect(result.data).toEqual([{ id: 1, name: '动画' }])
    expect(mockApi.get).toHaveBeenCalledWith('/category')
  })

  it('getRankingArchive → GET /manuscript/hot (rId=0 全站)', async () => {
    mockApi.get.mockResolvedValue({ data: { list: [] } })
    await getRankingArchive({ rId: 0, p: 1, size: 20 })
    expect(mockApi.get).toHaveBeenCalledWith(
      expect.stringContaining('/manuscript/hot?')
    )
  })

  it('getRankingArchive → 带 offset 分页', async () => {
    mockApi.get.mockResolvedValue({ data: { list: [] } })
    await getRankingArchive({ rId: 0, p: 3, size: 20 })
    const url = mockApi.get.mock.calls[0][0] as string
    expect(url).toContain('offset=')
    // 全站排行用 offset 分页，URL 不含 size 参数（后端固定每页 20 条）
    expect(url).toBe('/manuscript/hot?offset=40')
  })
})
