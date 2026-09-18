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

import {
  creatorApi,
  manuscriptApi,
  collectionApi as creatorCollectionApi,
  followApi,
  statsApi,
} from './creator'

describe('creator api - creatorApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getMyFollowers', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await creatorApi.getMyFollowers()
    expect(mocks.apiGet).toHaveBeenCalledWith('/follow/me/followers')
  })

  it('getMyFollowing', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await creatorApi.getMyFollowing()
    expect(mocks.apiGet).toHaveBeenCalledWith('/follow/me/following')
  })

  it('getFans / getFollowing', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await creatorApi.getFans(5)
    await creatorApi.getFollowing(5)
    expect(mocks.apiGet).toHaveBeenNthCalledWith(1, '/user/5/followers')
    expect(mocks.apiGet).toHaveBeenNthCalledWith(2, '/user/5/following')
  })

  it('checkFollow', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { following: true } })
    await creatorApi.checkFollow(6)
    expect(mocks.apiGet).toHaveBeenCalledWith('/follow/check/6')
  })

  it('getComments 传 params', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await creatorApi.getComments({ page: 1, size: 10 })
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/comments', { params: { page: 1, size: 10 } })
  })

  it('deleteComment / deleteReply', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await creatorApi.deleteComment(1)
    await creatorApi.deleteReply(2)
    expect(mocks.apiDelete).toHaveBeenNthCalledWith(1, '/creator/comments/1')
    expect(mocks.apiDelete).toHaveBeenNthCalledWith(2, '/creator/comments/reply/2')
  })

  it('replyComment 用 null body + query params', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await creatorApi.replyComment(1, 'hi', 9)
    expect(mocks.apiPost).toHaveBeenCalledWith(
      '/creator/comments/1/reply',
      null,
      { params: { content: 'hi', replyToUserId: 9 } },
    )
  })

  it('getDanmakuList / deleteDanmaku', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await creatorApi.getDanmakuList({ page: 1 })
    await creatorApi.deleteDanmaku(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/danmaku/list', { params: { page: 1 } })
    expect(mocks.apiDelete).toHaveBeenCalledWith('/creator/danmaku/7')
  })
})

describe('creator api - manuscriptApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getMyManuscripts 传 params', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await manuscriptApi.getMyManuscripts({ status: 1 })
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/me/list', { params: { status: 1 } })
  })

  it('getMyStats', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await manuscriptApi.getMyStats()
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/me/stats')
  })

  it('getUserManuscripts 拼接 userId', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await manuscriptApi.getUserManuscripts(8, { page: 1 })
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/user/8', { params: { page: 1 } })
  })

  it('getManuscriptById / getMyManuscriptById', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await manuscriptApi.getManuscriptById(9)
    await manuscriptApi.getMyManuscriptById(9)
    expect(mocks.apiGet).toHaveBeenNthCalledWith(1, '/manuscript/9')
    expect(mocks.apiGet).toHaveBeenNthCalledWith(2, '/manuscript/9')
  })

  it('updateManuscript 用 multipart', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 'v.mp4')
    await manuscriptApi.updateManuscript(1, {
      title: 'T',
      description: 'D',
      categoryId: 2,
      cover: file,
      tags: ['a', 'b'],
      videos: [{ id: 1, title: 'P1', file: file, durationSeconds: 10 }],
    })
    const [url, form, cfg] = mocks.apiPut.mock.calls[0]
    expect(url).toBe('/manuscript/1')
    expect(form.get('title')).toBe('T')
    expect(form.get('description')).toBe('D')
    expect(form.get('categoryId')).toBe('2')
    expect(form.get('cover')).toBe(file)
    expect(form.getAll('tags')).toEqual(['a', 'b'])
    expect(form.get('videos[0].id')).toBe('1')
    expect(form.get('videos[0].file')).toBe(file)
    expect(cfg).toEqual({ headers: { 'Content-Type': 'multipart/form-data' } })
  })

  it('deleteManuscript / unpublishManuscript / publishManuscript', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.deleteManuscript(1)
    await manuscriptApi.unpublishManuscript(1)
    await manuscriptApi.publishManuscript(1)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/manuscript/1')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(1, '/manuscript/1/unpublish')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(2, '/manuscript/1/publish')
  })
})

describe('creator api - collectionApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getUserCollections 默认 size=100', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await creatorCollectionApi.getUserCollections(1)
    expect(mocks.apiGet).toHaveBeenCalledWith('/collection/user/1?page=1&size=100')
  })

  it('getCollectionManuscripts 默认 size=20', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await creatorCollectionApi.getCollectionManuscripts(1)
    expect(mocks.apiGet).toHaveBeenCalledWith('/collection/1/manuscripts?page=1&size=20')
  })

  it('createCollection 走 multipart 并设置 isPublic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await creatorCollectionApi.createCollection({ name: 'n', description: 'd', isPublic: false })
    const [url, , cfg] = mocks.apiPost.mock.calls[0]
    expect(url).toBe('/collection')
    expect(cfg).toEqual({ headers: { 'Content-Type': 'multipart/form-data' } })
  })

  it('updateCollection 走 multipart', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await creatorCollectionApi.updateCollection(1, { name: 'n' })
    expect(mocks.apiPut).toHaveBeenCalledWith(
      '/collection/1',
      expect.any(FormData),
      { headers: { 'Content-Type': 'multipart/form-data' } },
    )
  })

  it('deleteCollection', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await creatorCollectionApi.deleteCollection(1)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/collection/1')
  })

  it('addManuscriptToCollection 携带 order', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await creatorCollectionApi.addManuscriptToCollection(1, 2, 5)
    expect(mocks.apiPost).toHaveBeenCalledWith('/collection/1/manuscript/2?order=5')
  })

  it('removeManuscriptFromCollection', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await creatorCollectionApi.removeManuscriptFromCollection(1, 2)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/collection/1/manuscript/2')
  })
})

describe('creator api - followApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('follow / unfollow / checkFollow', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { following: true } })
    await followApi.follow(1)
    await followApi.unfollow(1)
    await followApi.checkFollow(1)
    expect(mocks.apiPost).toHaveBeenCalledWith('/follow/1')
    expect(mocks.apiDelete).toHaveBeenCalledWith('/follow/1')
    expect(mocks.apiGet).toHaveBeenCalledWith('/follow/check/1')
  })
})

describe('creator api - statsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getOverview', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await statsApi.getOverview()
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/overview')
  })

  it('getTrend 默认 days=7', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await statsApi.getTrend()
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/trend?days=7')
  })

  it('getRanking 默认参数', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await statsApi.getRanking()
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/ranking?sortBy=views&limit=10')
  })

  it('getLatestComments / getFansRanking / getFansTrend / getManuscriptTrend', async () => {
    mocks.apiGet.mockResolvedValue({ code: 200, data: [] })
    await statsApi.getLatestComments(3)
    await statsApi.getFansRanking('like', 5)
    await statsApi.getFansTrend(14)
    await statsApi.getManuscriptTrend()
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/latest-comments?limit=3')
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/fans-ranking?type=like&limit=5')
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/fans-trend?days=14')
    expect(mocks.apiGet).toHaveBeenCalledWith('/creator/stats/manuscript-trend')
  })
})
