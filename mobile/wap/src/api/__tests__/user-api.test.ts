import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../../utils/session', () => ({
  getLocalUser: vi.fn()
}))

vi.mock('../../utils/storage_layer', () => ({
  default: { get: vi.fn(), set: vi.fn(), remove: vi.fn() },
  K: { user: 'user' }
}))

vi.mock('../client', () => ({
  default: { get: vi.fn(), put: vi.fn() }
}))

import api from '../client'
import { getLocalUser } from '../../utils/session'
import storage from '../../utils/storage_layer'
import { getMyInfo, getFollowingList, getFollowerList, updateMyInfo, getLocalUserId } from '../user'

const mockApi = vi.mocked(api)
const mockGetLocalUser = vi.mocked(getLocalUser)
const mockStorage = vi.mocked(storage)

describe('user API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getLocalUserId', () => {
    it('有本地用户 → 返回 id', () => {
      mockGetLocalUser.mockReturnValue({ id: 42 })
      expect(getLocalUserId()).toBe(42)
    })

    it('无本地用户 → null', () => {
      mockGetLocalUser.mockReturnValue(null)
      expect(getLocalUserId()).toBeNull()
    })
  })

  describe('getMyInfo', () => {
    it('无本地用户 → 返回空', async () => {
      mockGetLocalUser.mockReturnValue(null)
      const result = await getMyInfo()
      expect(result).toEqual({ code: '0', data: null })
    })

    it('有本地用户 + 服务端成功 → 更新本地', async () => {
      mockGetLocalUser.mockReturnValue({ id: 1, name: 'admin' })
      mockApi.get.mockResolvedValue({ data: { id: 1, username: 'admin' } })
      const result = await getMyInfo()
      expect(result.code).toBe('1')
      expect(result.data).toBeDefined()
      expect(mockStorage.set).toHaveBeenCalled()
    })

    it('网络异常 → 返回本地数据', async () => {
      mockGetLocalUser.mockReturnValue({ id: 1, name: 'admin' })
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getMyInfo()
      expect(result.code).toBe('1')
      expect(result.data).toBeDefined()
    })
  })

  describe('getFollowingList', () => {
    it('返回关注列表（经过 normalizeUser）', async () => {
      mockApi.get.mockResolvedValue({
        data: [{ id: 2, username: 'user2' }, { id: 3, name: 'user3' }]
      })
      const result = await getFollowingList(1)
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(2)
      expect(result.data[0].id).toBe(2)
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getFollowingList(1)
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getFollowerList', () => {
    it('返回粉丝列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 5 }] })
      const result = await getFollowerList(1)
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
    })
  })

  describe('updateMyInfo', () => {
    it('更新成功 → 返回更新后的用户数据', async () => {
      mockApi.put.mockResolvedValue({ data: { id: 1, nickname: '新名字' } })
      const result = await updateMyInfo(1, { nickname: '新名字' })
      expect(result.code).toBe('1')
      expect(mockStorage.set).toHaveBeenCalled()
    })

    it('更新失败 → 返回 code 0', async () => {
      mockApi.put.mockRejectedValue(new Error('fail'))
      const result = await updateMyInfo(1, { nickname: 'x' })
      expect(result.code).toBe('0')
    })
  })
})
