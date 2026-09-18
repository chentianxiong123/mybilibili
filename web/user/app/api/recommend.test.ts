import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: vi.fn(),
    delete: vi.fn(),
  }),
}))

import { recommendApi } from './recommend'

describe('recommend api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getRelatedVideos 默认 size=8', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getRelatedVideos(101)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/related/101?size=8')
  })

  it('getRelatedVideos 自定义 size', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getRelatedVideos(101, 5)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/related/101?size=5')
  })

  it('getHotVideos 无分类时不追加 categoryId', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getHotVideos()
    const url = mocks.apiGet.mock.calls[0][0] as string
    expect(url).toBe('/recommend/hot?size=10')
    expect(url).not.toContain('categoryId')
  })

  it('getHotVideos 带分类', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getHotVideos(3, 4)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/hot?size=4&categoryId=3')
  })

  it('getRecommendedVideos', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getRecommendedVideos(15)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/for-you?size=15')
  })

  it('getBehaviorBasedRecommendations', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getBehaviorBasedRecommendations(20)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/behavior?size=20')
  })

  it('getCollaborativeRecommendations', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getCollaborativeRecommendations(8)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/collaborative?size=8')
  })

  it('getContentBasedRecommendations', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getContentBasedRecommendations(7, 6)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/content/7?size=6')
  })

  it('getGuessYouLike 默认 size=12', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getGuessYouLike()
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/guess-you-like?size=12')
  })

  it('getRealtimeHot', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getRealtimeHot(5)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/realtime-hot?size=5')
  })

  it('getCategoryHot', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await recommendApi.getCategoryHot(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/recommend/category-hot?size=7')
  })
})
