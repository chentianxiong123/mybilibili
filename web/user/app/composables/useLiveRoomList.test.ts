import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

vi.mock('@/api/live.ts', () => ({
  liveApi: {
    getLiveList: vi.fn(),
    getMyRoom: vi.fn(),
  },
}))

import { liveApi } from '@/api/live.ts'
import { useLiveRoomList } from './useLiveRoomList'

const api = liveApi as unknown as {
  getLiveList: ReturnType<typeof vi.fn>
  getMyRoom: ReturnType<typeof vi.fn>
}

describe('useLiveRoomList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('loadData', () => {
    it('成功时同时请求 getLiveList 和 getMyRoom', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 200, data: [{ id: 1, userId: 10, viewerCount: 5 }] })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: { userId: 99, roomName: 'myroom' } })
      const r = useLiveRoomList()
      await r.loadData()
      expect(api.getLiveList).toHaveBeenCalledTimes(1)
      expect(api.getMyRoom).toHaveBeenCalledTimes(1)
      expect(r.liveRooms.value).toHaveLength(1)
      expect(r.myRoom.value).toEqual({ userId: 99, roomName: 'myroom' })
    })

    it('silent=true 时 loading 不被设为 true', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 200, data: [] })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const r = useLiveRoomList()
      expect(r.loading.value).toBe(false)
      await r.loadData(true)
      expect(r.loading.value).toBe(false)
    })

    it('非 silent 时 loading 在执行后回到 false', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 200, data: [] })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const r = useLiveRoomList()
      await r.loadData()
      expect(r.loading.value).toBe(false)
    })

    it('listRes.code 非 200 时 liveRooms 保持空', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 500, data: null })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const r = useLiveRoomList()
      await r.loadData()
      expect(r.liveRooms.value).toEqual([])
    })

    it('myRoomRes 失败时 myRoom 被置为 null', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 200, data: [] })
      api.getMyRoom.mockResolvedValueOnce({ code: 500, data: null })
      const r = useLiveRoomList()
      await r.loadData()
      expect(r.myRoom.value).toBe(null)
    })

    it('任意一个请求 rejected 时整个 loadData 不抛错', async () => {
      api.getLiveList.mockRejectedValueOnce(new Error('boom'))
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const r = useLiveRoomList()
      await expect(r.loadData()).resolves.toBeUndefined()
    })
  })

  describe('filteredRooms', () => {
    const seed = async () => {
      api.getLiveList.mockResolvedValueOnce({
        code: 200,
        data: [
          { id: 1, userId: 10, roomName: 'fun', category: '游戏', viewerCount: 100 },
          { id: 2, userId: 20, roomName: 'learn', category: '学习', viewerCount: 50 },
          { id: 3, userId: 30, roomName: 'mixed', category: '游戏', viewerCount: 200 },
          { id: 4, userId: 40, roomName: 'all-cat', category: '娱乐', viewerCount: 10 },
        ]
      })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const r = useLiveRoomList()
      await r.loadData()
      return r
    }

    it('默认按 hot 排序（viewCount 降序）', async () => {
      const r = await seed()
      const ids = r.filteredRooms.value.map(x => x.id)
      expect(ids).toEqual([3, 1, 2, 4])
    })

    it('sortBy=id 时按 id 降序', async () => {
      const r = await seed()
      r.sortBy.value = 'id'
      const ids = r.filteredRooms.value.map(x => x.id)
      expect(ids).toEqual([4, 3, 2, 1])
    })

    it('selectedCategory 过滤生效', async () => {
      const r = await seed()
      r.selectedCategory.value = '游戏'
      const ids = r.filteredRooms.value.map(x => x.id)
      expect(ids).toEqual([3, 1])
    })

    it('myRoom 存在时排除自己的直播间', async () => {
      api.getLiveList.mockResolvedValueOnce({
        code: 200,
        data: [
          { id: 1, userId: 10, roomName: 'other' },
          { id: 2, userId: 99, roomName: 'mine' },
        ]
      })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: { userId: 99 } })
      const r = useLiveRoomList()
      await r.loadData()
      const ids = r.filteredRooms.value.map(x => x.id)
      expect(ids).toEqual([1])
    })

    it('keyword 命中 roomName（不区分大小写）', async () => {
      const r = await seed()
      r.keyword.value = 'FUN'
      const ids = r.filteredRooms.value.map(x => x.id)
      expect(ids).toEqual([1])
    })

    it('keyword 命中 userId 字符串', async () => {
      const r = await seed()
      r.keyword.value = '20'
      const ids = r.filteredRooms.value.map(x => x.id)
      expect(ids).toEqual([2])
    })

    it('全选 全部 + 空 keyword 时返回所有房间', async () => {
      const r = await seed()
      expect(r.filteredRooms.value).toHaveLength(4)
    })

    it('selectedCategory 为 空串时不过滤', async () => {
      const r = await seed()
      r.selectedCategory.value = ''
      expect(r.filteredRooms.value).toHaveLength(4)
    })
  })

  describe('refreshData', () => {
    it('调用 message.success 并触发 loadData', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 200, data: [] })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const message = { success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn() }
      const r = useLiveRoomList({ message })
      await r.refreshData()
      expect(message.success).toHaveBeenCalledWith('已刷新')
      expect(api.getLiveList).toHaveBeenCalledTimes(1)
    })
  })

  describe('startPolling / stopPolling', () => {
    it('startPolling 注册定时器', () => {
      const setIntervalFn = vi.fn(() => 'timer-id' as any)
      const clearIntervalFn = vi.fn()
      const r = useLiveRoomList({ setIntervalFn, clearIntervalFn })
      const tid = r.startPolling()
      expect(setIntervalFn).toHaveBeenCalledTimes(1)
      expect(tid).toBe('timer-id')
    })

    it('startPolling 重复调用会先 stop 再注册', () => {
      const setIntervalFn = vi.fn(() => 'tid')
      const clearIntervalFn = vi.fn()
      const r = useLiveRoomList({ setIntervalFn, clearIntervalFn })
      r.startPolling()
      r.startPolling()
      expect(clearIntervalFn).toHaveBeenCalledTimes(1)
      expect(setIntervalFn).toHaveBeenCalledTimes(2)
    })

    it('stopPolling 清理定时器并返回 true', () => {
      const setIntervalFn = vi.fn(() => 'tid')
      const clearIntervalFn = vi.fn()
      const r = useLiveRoomList({ setIntervalFn, clearIntervalFn })
      r.startPolling()
      const ok = r.stopPolling()
      expect(ok).toBe(true)
      expect(clearIntervalFn).toHaveBeenCalledWith('tid')
    })

    it('stopPolling 第二次调用返回 false', () => {
      const setIntervalFn = vi.fn(() => 'tid')
      const r = useLiveRoomList({ setIntervalFn })
      r.startPolling()
      r.stopPolling()
      expect(r.stopPolling()).toBe(false)
    })

    it('cleanupLiveRoomList 等价于 stopPolling', () => {
      const setIntervalFn = vi.fn(() => 'tid')
      const clearIntervalFn = vi.fn()
      const r = useLiveRoomList({ setIntervalFn, clearIntervalFn })
      r.startPolling()
      r.cleanupLiveRoomList()
      expect(clearIntervalFn).toHaveBeenCalledWith('tid')
    })

    it('pollDelay 默认 10000ms', () => {
      const setIntervalFn = vi.fn(() => 'tid')
      const r = useLiveRoomList({ setIntervalFn })
      r.startPolling()
      expect(setIntervalFn).toHaveBeenCalledWith(expect.any(Function), 10000)
    })

    it('custom pollDelay 生效', () => {
      const setIntervalFn = vi.fn(() => 'tid')
      const r = useLiveRoomList({ setIntervalFn, pollDelay: 5000 })
      r.startPolling()
      expect(setIntervalFn).toHaveBeenCalledWith(expect.any(Function), 5000)
    })
  })

  describe('default API / categories', () => {
    it('categories 包含 全部 / 娱乐 / 游戏 / 学习 / 购物 / 赛事 / 其他', () => {
      const r = useLiveRoomList()
      expect(r.categories).toEqual(['全部', '娱乐', '游戏', '学习', '购物', '赛事', '其他'])
    })

    it('selectedCategory 默认 全部', () => {
      const r = useLiveRoomList()
      expect(r.selectedCategory.value).toBe('全部')
    })

    it('totalLive 等于 liveRooms.length', async () => {
      api.getLiveList.mockResolvedValueOnce({ code: 200, data: [{}, {}, {}] })
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      const r = useLiveRoomList()
      await r.loadData()
      expect(r.totalLive.value).toBe(3)
    })
  })
})