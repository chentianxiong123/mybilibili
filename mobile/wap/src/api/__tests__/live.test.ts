import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn() }
}))

import api from '../client'
import { getAreas, getLiveIndexData, getLiveListData, getRoomInfo, getDanMuConfig } from '../live'

const mockApi = vi.mocked(api)

describe('live API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getAreas', () => {
    it('从直播列表中提取去重的分区信息', async () => {
      mockApi.get.mockResolvedValue({
        data: [
          { areaId: 1, areaName: '游戏', parentAreaId: 0 },
          { areaId: 1, areaName: '游戏', parentAreaId: 0 },
          { areaId: 2, areaName: '娱乐', parentAreaId: 0 }
        ]
      })
      const result = await getAreas()
      expect(result.code).toBe('1')
      expect(result.data.length).toBe(2)
      expect(result.data[0]).toEqual({ id: 1, name: '游戏', parentId: 0 })
    })

    it('网络异常 → 空数组', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getAreas()
      expect(result).toEqual({ code: '0', data: [] })
    })
  })

  describe('getLiveIndexData', () => {
    it('返回直播首页数据（bannerList + itemList）', async () => {
      mockApi.get.mockResolvedValue({
        data: [{ id: 1, roomName: '直播间1', status: 'live', viewerCount: 100 }]
      })
      const result = await getLiveIndexData()
      expect(result.code).toBe('1')
      expect(result.data.bannerList).toEqual([])
      expect(result.data.itemList.length).toBe(1)
      expect(result.data.itemList[0].areaName).toBe('热门直播')
      expect(result.data.itemList[0].lives.length).toBe(1)
      expect(result.data.itemList[0].lives[0].roomId).toBe(1)
    })

    it('网络异常 → 空数据', async () => {
      mockApi.get.mockRejectedValue(new Error('network'))
      const result = await getLiveIndexData()
      expect(result.code).toBe('0')
      expect(result.data.itemList).toEqual([])
    })
  })

  describe('getLiveListData', () => {
    it('返回直播列表', async () => {
      mockApi.get.mockResolvedValue({
        data: [{ id: 1, roomName: '直播间1' }, { id: 2, roomName: '直播间2' }]
      })
      const result = await getLiveListData()
      expect(result.code).toBe('1')
      expect(result.data.list.length).toBe(2)
      expect(result.data.list[0].roomId).toBe(1)
    })
  })

  describe('getRoomInfo', () => {
    it('返回直播间详情', async () => {
      mockApi.get.mockResolvedValue({
        data: { id: 1, roomName: '测试直播间', userId: 100, followerCount: 500 }
      })
      const result = await getRoomInfo(1)
      expect(result.code).toBe('1')
      expect(result.data.roomId).toBe(1)
      expect(result.data.uId).toBe(100)
      expect(result.data.followerCount).toBe(500)
    })

    it('无数据 → code 0', async () => {
      mockApi.get.mockResolvedValue(null)
      const result = await getRoomInfo(1)
      expect(result.code).toBe('0')
    })
  })

  describe('getDanMuConfig', () => {
    it('返回弹幕配置（同步）', () => {
      const result = getDanMuConfig(1)
      expect(result.code).toBe('1')
      expect(result.data.host).toBeDefined()
    })
  })
})
