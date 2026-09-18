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

import { getHomeBanners, getCategoryBanners, getBackgroundImage, getUserProfileBackground } from './banner'

describe('banner api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getHomeBanners 调用 GET /banner-images/home', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: ['banner1'] })
    const res = await getHomeBanners()
    expect(mocks.apiGet).toHaveBeenCalledWith('/banner-images/home')
    expect(res).toEqual({ code: 200, data: ['banner1'] })
  })

  it('getCategoryBanners 拼接 categoryId', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await getCategoryBanners(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/banner-images/category/7')
  })

  it('getBackgroundImage 调用 /banner-images/background', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: 'bg.png' })
    await getBackgroundImage()
    expect(mocks.apiGet).toHaveBeenCalledWith('/banner-images/background')
  })

  it('getUserProfileBackground 调用 /banner-images/user-profile', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: 'user-bg.png' })
    await getUserProfileBackground()
    expect(mocks.apiGet).toHaveBeenCalledWith('/banner-images/user-profile')
  })

  it('错误时透传拒绝', async () => {
    mocks.apiGet.mockRejectedValueOnce(new Error('boom'))
    await expect(getHomeBanners()).rejects.toThrow('boom')
  })
})
