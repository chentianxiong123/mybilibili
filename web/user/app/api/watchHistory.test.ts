import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
  apiPut: vi.fn(),
  apiDelete: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: mocks.apiPut,
    delete: mocks.apiDelete,
  }),
}))

import { watchHistoryApi } from './watchHistory'

describe('watchHistory api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getWatchHistory 默认分页', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await watchHistoryApi.getWatchHistory()
    expect(mocks.apiGet).toHaveBeenCalledWith('/watch-history?page=1&size=20')
  })

  it('getWatchHistory 自定义', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await watchHistoryApi.getWatchHistory(2, 50)
    expect(mocks.apiGet).toHaveBeenCalledWith('/watch-history?page=2&size=50')
  })

  it('recordWatchHistory 把秒数映射到字段', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await watchHistoryApi.recordWatchHistory(11, 100, 22, 200)
    expect(mocks.apiPost).toHaveBeenCalledWith('/watch-history', null, {
      params: { videoId: 11, progressSeconds: 100, videoDuration: 200 },
    })
  })

  it('recordWatchHistory progress 缺省回退 0', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await watchHistoryApi.recordWatchHistory(11, undefined, undefined, undefined)
    expect(mocks.apiPost).toHaveBeenCalledWith('/watch-history', null, {
      params: { videoId: 11, progressSeconds: 0, videoDuration: 0 },
    })
  })

  it('clearWatchHistory', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await watchHistoryApi.clearWatchHistory()
    expect(mocks.apiDelete).toHaveBeenCalledWith('/watch-history')
  })

  it('deleteWatchHistory', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await watchHistoryApi.deleteWatchHistory(7)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/watch-history/7')
  })

  it('batchDeleteWatchHistory POST ids', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await watchHistoryApi.batchDeleteWatchHistory([1, 2, 3])
    expect(mocks.apiPost).toHaveBeenCalledWith('/watch-history/batch-delete', { ids: [1, 2, 3] })
  })

  it('getWatchHistoryStats', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await watchHistoryApi.getWatchHistoryStats()
    expect(mocks.apiGet).toHaveBeenCalledWith('/watch-history/stats')
  })

  it('getContinueWatching 默认 size=10', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await watchHistoryApi.getContinueWatching()
    expect(mocks.apiGet).toHaveBeenCalledWith('/watch-history/continue-watching?size=10')
  })

  it('getTodayWatchTime', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: 0 })
    await watchHistoryApi.getTodayWatchTime()
    expect(mocks.apiGet).toHaveBeenCalledWith('/watch-history/today-time')
  })

  it('checkWatchStatus', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { watched: true } })
    await watchHistoryApi.checkWatchStatus(5)
    expect(mocks.apiGet).toHaveBeenCalledWith('/watch-history/check/5')
  })

  it('updateProgress PUT body', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await watchHistoryApi.updateProgress(5, 60)
    expect(mocks.apiPut).toHaveBeenCalledWith('/watch-history/progress', { videoId: 5, progress: 60 })
  })
})
