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

import { dynamicApi } from './dynamic'

describe('dynamic api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getDynamicList 默认分页参数', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await dynamicApi.getDynamicList()
    expect(mocks.apiGet).toHaveBeenCalledWith('/dynamic/list', { params: { page: 1, size: 10 } })
  })

  it('getDynamicList 自定义分页', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await dynamicApi.getDynamicList(3, 5)
    expect(mocks.apiGet).toHaveBeenCalledWith('/dynamic/list', { params: { page: 3, size: 5 } })
  })

  it('getFollowingDynamics 不带 userId 时不追加', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await dynamicApi.getFollowingDynamics()
    expect(mocks.apiGet).toHaveBeenCalledWith('/dynamic/following', { params: { page: 1, size: 10 } })
  })

  it('getFollowingDynamics 带 userId', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await dynamicApi.getFollowingDynamics(2, 20, 9)
    expect(mocks.apiGet).toHaveBeenCalledWith('/dynamic/following', { params: { page: 2, size: 20, userId: 9 } })
  })

  it('getUserDynamics 拼接 userId 和分页', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await dynamicApi.getUserDynamics(11, 1, 10)
    expect(mocks.apiGet).toHaveBeenCalledWith('/dynamic/user/11', { params: { page: 1, limit: 10 } })
  })

  it('getDynamicById', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { id: 1 } })
    await dynamicApi.getDynamicById(1)
    expect(mocks.apiGet).toHaveBeenCalledWith('/dynamic/1')
  })

  it('publishDynamic 用 multipart', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    const fd = new FormData()
    await dynamicApi.publishDynamic(fd)
    expect(mocks.apiPost).toHaveBeenCalledWith('/dynamic/publish', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  })

  it('deleteDynamic', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await dynamicApi.deleteDynamic(3)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/dynamic/3')
  })

  it('likeDynamic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await dynamicApi.likeDynamic(4)
    expect(mocks.apiPost).toHaveBeenCalledWith('/dynamic/like/4')
  })

  it('unlikeDynamic', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await dynamicApi.unlikeDynamic(4)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/dynamic/like/4')
  })

  it('shareDynamic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await dynamicApi.shareDynamic(5)
    expect(mocks.apiPost).toHaveBeenCalledWith('/dynamic/share/5')
  })
})
