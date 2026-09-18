import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPut: vi.fn(),
  apiPost: vi.fn(),
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

import { userPrivacyApi } from './userPrivacy'

describe('userPrivacyApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getPrivacySettings 成功时返回 api 结果', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { publicCollection: true } })
    const res = await userPrivacyApi.getPrivacySettings()
    expect(mocks.apiGet).toHaveBeenCalledWith('/user/privacy/settings')
    expect(res).toEqual({ code: 200, data: { publicCollection: true } })
  })

  it('getPrivacySettings 404 时回退默认设置', async () => {
    const err: any = new Error('not found')
    err.response = { status: 404 }
    mocks.apiGet.mockRejectedValueOnce(err)
    const res = await userPrivacyApi.getPrivacySettings()
    expect(res.code).toBe(200)
    expect(res.data).toEqual({
      publicCollection: true,
      publicBirthdayTags: false,
      publicCoinVideos: false,
      publicLikeVideos: false,
      publicFollowingList: false,
      publicFollowersList: false,
      tags: [],
    })
  })

  it('getPrivacySettings 非 404 错误时透传 reject', async () => {
    const err: any = new Error('boom')
    err.response = { status: 500 }
    mocks.apiGet.mockRejectedValueOnce(err)
    await expect(userPrivacyApi.getPrivacySettings()).rejects.toBe(err)
  })

  it('updatePrivacySettings PUT body', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    const payload = { publicCollection: false }
    await userPrivacyApi.updatePrivacySettings(payload)
    expect(mocks.apiPut).toHaveBeenCalledWith('/user/privacy/settings', payload)
  })

  it('getUserTags', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await userPrivacyApi.getUserTags()
    expect(mocks.apiGet).toHaveBeenCalledWith('/user/privacy/tags')
  })

  it('addUserTag 用 query params', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await userPrivacyApi.addUserTag('x')
    expect(mocks.apiPost).toHaveBeenCalledWith('/user/privacy/tags', null, { params: { tagName: 'x' } })
  })

  it('removeUserTag 用 query params', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await userPrivacyApi.removeUserTag('y')
    expect(mocks.apiDelete).toHaveBeenCalledWith('/user/privacy/tags', { params: { tagName: 'y' } })
  })
})
