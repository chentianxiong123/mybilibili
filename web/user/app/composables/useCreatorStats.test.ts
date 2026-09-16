import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

vi.mock('@/api/creator', () => ({
  statsApi: {
    getOverview: vi.fn(),
    getTrend: vi.fn(),
    getRanking: vi.fn(),
    getFansTrend: vi.fn(),
    getManuscriptTrend: vi.fn(),
  }
}))

import { statsApi } from '@/api/creator'
import { useCreatorStats } from './useCreatorStats'

const mockStatsApi = statsApi as unknown as {
  getOverview: ReturnType<typeof vi.fn>
  getTrend: ReturnType<typeof vi.fn>
  getRanking: ReturnType<typeof vi.fn>
  getFansTrend: ReturnType<typeof vi.fn>
  getManuscriptTrend: ReturnType<typeof vi.fn>
}

describe('useCreatorStats', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('loadOverview', () => {
    it('成功后把 res.data 合并进 overview', async () => {
      mockStatsApi.getOverview.mockResolvedValueOnce({
        code: 200,
        data: { totalViews: 1000, totalLikes: 200, updateTime: '2025-01-01' }
      })
      const { loadOverview, overview } = useCreatorStats()
      await loadOverview()
      expect(overview.value.totalViews).toBe(1000)
      expect(overview.value.totalLikes).toBe(200)
      expect(overview.value.updateTime).toBe('2025-01-01')
    })

    it('非 200 响应时 overview 保持初值', async () => {
      mockStatsApi.getOverview.mockResolvedValueOnce({ code: 500, data: null })
      const { loadOverview, overview } = useCreatorStats()
      await loadOverview()
      expect(overview.value.totalViews).toBe(0)
    })

    it('loading.overview 在执行期间为 true', async () => {
      let resolveFn: any
      mockStatsApi.getOverview.mockReturnValueOnce(new Promise((r) => { resolveFn = r }))
      const { loadOverview, loading } = useCreatorStats()
      const p = loadOverview()
      expect(loading.overview).toBe(true)
      resolveFn({ code: 200, data: {} })
      await p
      expect(loading.overview).toBe(false)
    })

    it('rejected 时 console.error 被调用，loading 仍会复位', async () => {
      mockStatsApi.getOverview.mockRejectedValueOnce(new Error('boom'))
      const { loadOverview, loading } = useCreatorStats()
      await loadOverview()
      expect(loading.overview).toBe(false)
      expect(console.error).toHaveBeenCalled()
    })

    it('并发调用直接返回，不重复触发 API', async () => {
      let resolveFn: any
      mockStatsApi.getOverview.mockReturnValueOnce(new Promise((r) => { resolveFn = r }))
      const { loadOverview } = useCreatorStats()
      const p1 = loadOverview()
      await loadOverview()
      expect(mockStatsApi.getOverview).toHaveBeenCalledTimes(1)
      resolveFn({ code: 200, data: {} })
      await p1
    })
  })

  describe('loadTrend', () => {
    it('默认 days = 7 传给 statsApi.getTrend', async () => {
      mockStatsApi.getTrend.mockResolvedValueOnce({ code: 200, data: { dates: ['d1'] } })
      const { loadTrend } = useCreatorStats()
      await loadTrend()
      expect(mockStatsApi.getTrend).toHaveBeenCalledWith(7)
    })

    it('传入 days 时透传', async () => {
      mockStatsApi.getTrend.mockResolvedValueOnce({ code: 200, data: { dates: [] } })
      const { loadTrend } = useCreatorStats()
      await loadTrend(30)
      expect(mockStatsApi.getTrend).toHaveBeenCalledWith(30)
    })

    it('成功后赋值 trendData', async () => {
      mockStatsApi.getTrend.mockResolvedValueOnce({
        code: 200,
        data: { dates: ['2025-01-01', '2025-01-02'], views: [100, 200] }
      })
      const { loadTrend, trendData } = useCreatorStats()
      await loadTrend()
      expect(trendData.value.dates).toEqual(['2025-01-01', '2025-01-02'])
      expect(trendData.value.views).toEqual([100, 200])
    })

    it('非 200 时 trendData 保持空', async () => {
      mockStatsApi.getTrend.mockResolvedValueOnce({ code: 500 })
      const { loadTrend, trendData } = useCreatorStats()
      await loadTrend()
      expect(trendData.value.dates).toEqual([])
    })
  })

  describe('loadRanking', () => {
    it('默认参数为 views, 10', async () => {
      mockStatsApi.getRanking.mockResolvedValueOnce({ code: 200, data: { list: [] } })
      const { loadRanking } = useCreatorStats()
      await loadRanking()
      expect(mockStatsApi.getRanking).toHaveBeenCalledWith('views', 10)
    })

    it('成功后 rankingList 来自 res.data.list', async () => {
      mockStatsApi.getRanking.mockResolvedValueOnce({
        code: 200,
        data: { list: [{ id: 1, title: 'top' }] }
      })
      const { loadRanking, rankingList } = useCreatorStats()
      await loadRanking('likes', 5)
      expect(rankingList.value).toEqual([{ id: 1, title: 'top' }])
      expect(mockStatsApi.getRanking).toHaveBeenCalledWith('likes', 5)
    })

    it('缺少 data.list 时 rankingList 为空', async () => {
      mockStatsApi.getRanking.mockResolvedValueOnce({ code: 200, data: {} })
      const { loadRanking, rankingList } = useCreatorStats()
      await loadRanking()
      expect(rankingList.value).toEqual([])
    })
  })

  describe('loadFansTrend', () => {
    it('默认 days = 30', async () => {
      mockStatsApi.getFansTrend.mockResolvedValueOnce({ code: 200, data: { dates: [] } })
      const { loadFansTrend } = useCreatorStats()
      await loadFansTrend()
      expect(mockStatsApi.getFansTrend).toHaveBeenCalledWith(30)
    })

    it('成功后赋值 fansTrend', async () => {
      mockStatsApi.getFansTrend.mockResolvedValueOnce({
        code: 200,
        data: { dates: ['x'], newFollowers: [3], totalFollowers: [10], currentFollowers: 10 }
      })
      const { loadFansTrend, fansTrend } = useCreatorStats()
      await loadFansTrend()
      expect(fansTrend.value.currentFollowers).toBe(10)
    })
  })

  describe('loadManuscriptTrend', () => {
    it('成功后赋值 manuscriptTrend', async () => {
      mockStatsApi.getManuscriptTrend.mockResolvedValueOnce({
        code: 200,
        data: { dates: ['d'], views: [100], titles: ['t'] }
      })
      const { loadManuscriptTrend, manuscriptTrend } = useCreatorStats()
      await loadManuscriptTrend()
      expect(manuscriptTrend.value.titles).toEqual(['t'])
    })

    it('非 200 保持初值', async () => {
      mockStatsApi.getManuscriptTrend.mockResolvedValueOnce({ code: 500 })
      const { loadManuscriptTrend, manuscriptTrend } = useCreatorStats()
      await loadManuscriptTrend()
      expect(manuscriptTrend.value.titles).toEqual([])
    })

    it('并发调用会被忽略', async () => {
      let resolveFn: any
      mockStatsApi.getManuscriptTrend.mockReturnValueOnce(new Promise((r) => { resolveFn = r }))
      const { loadManuscriptTrend } = useCreatorStats()
      const p = loadManuscriptTrend()
      await loadManuscriptTrend()
      expect(mockStatsApi.getManuscriptTrend).toHaveBeenCalledTimes(1)
      resolveFn({ code: 200, data: {} })
      await p
    })
  })

  describe('loading 互斥', () => {
    it('trend 与 manuscriptTrend 共用 loading.trend，互斥', async () => {
      let resolveTrend: any
      mockStatsApi.getTrend.mockReturnValueOnce(new Promise((r) => { resolveTrend = r }))
      const { loadTrend, loadManuscriptTrend, loading } = useCreatorStats()
      const p1 = loadTrend(7)
      expect(loading.trend).toBe(true)
      await loadManuscriptTrend()
      expect(mockStatsApi.getManuscriptTrend).not.toHaveBeenCalled()
      resolveTrend({ code: 200, data: {} })
      await p1
      expect(loading.trend).toBe(false)
    })
  })
})