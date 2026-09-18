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

import { searchApi } from './search'

describe('search api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('searchVideos 默认参数', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.searchVideos({ keyword: 'cat' })
    expect(mocks.apiGet).toHaveBeenCalledWith(
      '/search/videos?page=1&size=20&sort=relevance&keyword=cat',
    )
  })

  it('searchVideos 自定义 page/size/sort/categoryId/tag', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.searchVideos({
      keyword: 'music',
      page: 2,
      size: 30,
      sort: 'hot',
      categoryId: 5,
      tag: 'v',
    })
    const url = mocks.apiGet.mock.calls[0][0] as string
    expect(url).toContain('page=2')
    expect(url).toContain('size=30')
    expect(url).toContain('sort=hot')
    expect(url).toContain('categoryId=5')
    expect(url).toContain('keyword=music')
    expect(url).toContain('tag=v')
  })

  it('searchVideos 关键字 encodeURIComponent', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.searchVideos({ keyword: 'a b/c' })
    const url = mocks.apiGet.mock.calls[0][0] as string
    expect(url).toContain('keyword=a%20b%2Fc')
  })

  it('searchVideos 无 keyword 不追加', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.searchVideos({})
    const url = mocks.apiGet.mock.calls[0][0] as string
    expect(url).not.toContain('keyword=')
    expect(url).not.toContain('categoryId=')
    expect(url).not.toContain('tag=')
  })

  it('getSearchSuggestions encodeURIComponent keyword', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.getSearchSuggestions('a b')
    expect(mocks.apiGet).toHaveBeenCalledWith('/search/suggest?keyword=a%20b')
  })

  it('getHotSearch', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.getHotSearch()
    expect(mocks.apiGet).toHaveBeenCalledWith('/search/hot')
  })

  it('getSearchHistory', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await searchApi.getSearchHistory()
    expect(mocks.apiGet).toHaveBeenCalledWith('/search/history')
  })

  it('clearSearchHistory', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await searchApi.clearSearchHistory()
    expect(mocks.apiDelete).toHaveBeenCalledWith('/search/history')
  })

  it('deleteSearchHistory encodeURIComponent', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await searchApi.deleteSearchHistory('a b')
    expect(mocks.apiDelete).toHaveBeenCalledWith('/search/history?keyword=a%20b')
  })
})
