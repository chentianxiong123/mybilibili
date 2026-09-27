import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn() }
}))

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn()
}))

import api from '../client'
import { isLogin } from '../../utils/session'
import { getCollectedVideos, getFavoriteFolders, getFavoriteFolderVideos } from '../favorite'

const mockApi = vi.mocked(api)
const mockIsLogin = vi.mocked(isLogin)

describe('favorite API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockIsLogin.mockReturnValue(true)
  })

  describe('getCollectedVideos', () => {
    it('未登录 → 空数组', async () => {
      mockIsLogin.mockReturnValue(false)
      const result = await getCollectedVideos()
      expect(result).toEqual({ code: '0', data: [] })
    })

    it('已登录 → 返回收藏列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1, title: '视频1' }] })
      const result = await getCollectedVideos()
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/user/collections')
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getCollectedVideos()
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getFavoriteFolders', () => {
    it('未登录 → 空数组', async () => {
      mockIsLogin.mockReturnValue(false)
      const result = await getFavoriteFolders()
      expect(result).toEqual({ code: '0', data: [] })
    })

    it('已登录 → 返回收藏夹列表', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 1, name: '默认收藏夹' }] })
      const result = await getFavoriteFolders()
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(1)
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/favorite/folders')
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('fail'))
      const result = await getFavoriteFolders()
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getFavoriteFolderVideos', () => {
    it('未登录 → 空数组', async () => {
      mockIsLogin.mockReturnValue(false)
      const result = await getFavoriteFolderVideos(1)
      expect(result).toEqual({ code: '0', data: [] })
    })

    it('已登录 → 返回收藏夹内视频', async () => {
      mockApi.get.mockResolvedValue({ data: { records: [{ id: 10 }] } })
      const result = await getFavoriteFolderVideos(1)
      expect(result.code).toBe('1')
      expect(result.data).toEqual([{ id: 10 }])
      expect(mockApi.get).toHaveBeenCalledWith('/manuscript/favorite/folders/1/videos')
    })

    it('data 直接是数组时也能处理', async () => {
      mockApi.get.mockResolvedValue({ data: [{ id: 20 }] })
      const result = await getFavoriteFolderVideos(2)
      expect(result.code).toBe('1')
      expect(result.data).toEqual([{ id: 20 }])
    })
  })
})
