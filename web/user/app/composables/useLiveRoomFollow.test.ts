import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useLiveRoomFollow } from './useLiveRoomFollow'

const STORAGE_KEY = 'followedRooms'

describe('useLiveRoomFollow', () => {
  beforeEach(() => {
    localStorage.clear()
    document.cookie.split(';').forEach(c => {
      document.cookie = c.trim().split('=')[0] + '=; Max-Age=0'
    })
  })

  describe('syncFollowState', () => {
    it('roomData 为 null 时 isFollowing=false 且 followerCount=0', () => {
      const room = ref(null)
      const r = useLiveRoomFollow({ room })
      r.syncFollowState(null)
      expect(r.isFollowing.value).toBe(false)
      expect(r.followerCount.value).toBe(0)
    })

    it('localStorage 已关注该 userId 时 isFollowing=true', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([100]))
      const room = ref({ userId: 100, followerCount: 42 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.isFollowing.value).toBe(true)
      expect(r.followerCount.value).toBe(42)
    })

    it('localStorage 未关注时 isFollowing=false', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([100]))
      const room = ref({ userId: 200, followerCount: 10 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.isFollowing.value).toBe(false)
    })

    it('localStorage 数据非法 JSON 时视作空数组', () => {
      localStorage.setItem(STORAGE_KEY, 'not-json')
      const room = ref({ userId: 200, followerCount: 10 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.isFollowing.value).toBe(false)
    })

    it('localStorage 是 string-of-array-of-number 时仍能匹配', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(['100']))
      const room = ref({ userId: 100, followerCount: 5 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.isFollowing.value).toBe(true)
    })

    it('followerCount 非数字时被规范化为 0', () => {
      const room = ref({ userId: 1, followerCount: 'abc' })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.followerCount.value).toBe(0)
    })

    it('无参调用时从 room.value 读取', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([1]))
      const room = ref({ userId: 1, followerCount: 3 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.isFollowing.value).toBe(true)
    })

    it('返回 true 表示成功同步', () => {
      const room = ref({ userId: 1, followerCount: 1 })
      const r = useLiveRoomFollow({ room })
      expect(r.syncFollowState()).toBe(true)
      expect(r.syncFollowState(null)).toBe(false)
    })
  })

  describe('toggleFollow', () => {
    it('未关注时点击关注写入 localStorage 并 +1', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([]))
      const room = ref({ userId: 7, followerCount: 3 })
      const message = { success: vi.fn(), error: vi.fn() }
      const r = useLiveRoomFollow({ room, message })
      r.syncFollowState() // 先同步初始 followerCount
      const ok = r.toggleFollow()
      expect(ok).toBe(true)
      expect(r.isFollowing.value).toBe(true)
      expect(r.followerCount.value).toBe(4)
      expect(message.success).toHaveBeenCalledWith('已关注主播')
      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]')
      expect(stored).toEqual([7])
    })

    it('已关注时点击取消关注 -1 并清出数组', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([7]))
      const room = ref({ userId: 7, followerCount: 3 })
      const message = { success: vi.fn(), error: vi.fn() }
      const r = useLiveRoomFollow({ room, message })
      r.syncFollowState()
      const ok = r.toggleFollow()
      expect(ok).toBe(true)
      expect(r.isFollowing.value).toBe(false)
      expect(r.followerCount.value).toBe(2)
      expect(message.success).toHaveBeenCalledWith('已取消关注')
      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]')
      expect(stored).toEqual([])
    })

    it('followerCount 减到 0 时不会变成负数', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([7]))
      const room = ref({ userId: 7, followerCount: 0 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      r.toggleFollow()
      expect(r.followerCount.value).toBe(0)
    })

    it('已关注列表中已存在该 userId 时不再追加', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([7]))
      const room = ref({ userId: 7, followerCount: 0 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      r.toggleFollow() // 取消
      r.toggleFollow() // 再关注
      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]')
      expect(stored).toEqual([7])
    })

    it('room.value 为 null 或无 userId 时返回 false', () => {
      const r1 = useLiveRoomFollow({ room: ref(null) })
      expect(r1.toggleFollow()).toBe(false)
      const r2 = useLiveRoomFollow({ room: ref({}) })
      expect(r2.toggleFollow()).toBe(false)
    })

    it('storage 为 null 时不抛错', () => {
      const r = useLiveRoomFollow({ room: ref({ userId: 7, followerCount: 0 }), storage: null as any })
      const ok = r.toggleFollow()
      expect(ok).toBe(true)
      expect(r.isFollowing.value).toBe(true)
    })

    it('message 为 null 时不抛错', () => {
      const r = useLiveRoomFollow({ room: ref({ userId: 7, followerCount: 0 }), message: null as any })
      expect(() => r.toggleFollow()).not.toThrow()
    })

    it('custom storageKey 生效', () => {
      localStorage.setItem('myKey', JSON.stringify([]))
      const r = useLiveRoomFollow({ room: ref({ userId: 9, followerCount: 0 }), storageKey: 'myKey' })
      r.toggleFollow()
      expect(localStorage.getItem('myKey')).toBe(JSON.stringify([9]))
    })

    it('同一 userId 多种形式（string/number）作为 key 时能匹配', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(['100']))
      const room = ref({ userId: 100, followerCount: 1 })
      const r = useLiveRoomFollow({ room })
      r.syncFollowState()
      expect(r.isFollowing.value).toBe(true)
    })
  })
})