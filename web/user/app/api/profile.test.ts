import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  safeStorage: {
    getItem: vi.fn(),
  },
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), { get: mocks.apiGet }),
}))

vi.mock('@/utils/safeStorage', () => ({
  safeStorage: mocks.safeStorage,
}))

import { profileApi } from './profile'

describe('profileApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getMyProfile 未登录时 reject', async () => {
    mocks.safeStorage.getItem.mockReturnValue(null)
    await expect(profileApi.getMyProfile()).rejects.toThrow('未登录')
    expect(mocks.apiGet).not.toHaveBeenCalled()
  })

  it('getMyProfile 已登录时取 user.id 调接口', async () => {
    mocks.safeStorage.getItem.mockReturnValue(JSON.stringify({ id: 42 }))
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { id: 42 } })
    const res = await profileApi.getMyProfile()
    expect(mocks.apiGet).toHaveBeenCalledWith('/profile/42')
    expect(res).toEqual({ code: 200, data: { id: 42 } })
  })

  it('getProfile 直接传 userId', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { id: 7 } })
    const res = await profileApi.getProfile(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/profile/7')
    expect(res).toEqual({ code: 200, data: { id: 7 } })
  })

  it('getMyProfile storage 非法 JSON 时抛出同步异常', () => {
    mocks.safeStorage.getItem.mockReturnValue('not json')
    expect(() => profileApi.getMyProfile()).toThrow(SyntaxError)
  })
})
