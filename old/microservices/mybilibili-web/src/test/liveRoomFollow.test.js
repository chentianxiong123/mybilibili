import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useLiveRoomFollow } from '../composables/useLiveRoomFollow.js'

const createStorage = (initial = {}) => {
  const values = { ...initial }
  return {
    getItem: vi.fn(key => values[key] ?? null),
    setItem: vi.fn((key, value) => {
      values[key] = value
    }),
    values
  }
}

const createFollow = (options = {}) => {
  const room = ref(options.room || { userId: 100, followerCount: 3 })
  const storage = options.storage || createStorage()
  const message = options.message || { success: vi.fn() }
  const follow = useLiveRoomFollow({
    room,
    storage,
    message,
    storageKey: options.storageKey
  })
  return {
    room,
    storage,
    message,
    ...follow
  }
}

describe('useLiveRoomFollow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('syncs follower count and existing follow state from storage', () => {
    const storage = createStorage({
      followedRooms: JSON.stringify(['100', 200])
    })
    const follow = createFollow({ storage })

    expect(follow.syncFollowState()).toBe(true)

    expect(follow.isFollowing.value).toBe(true)
    expect(follow.followerCount.value).toBe(3)
    expect(storage.getItem).toHaveBeenCalledWith('followedRooms')
  })

  it('adds a followed streamer without duplicating ids', () => {
    const storage = createStorage({
      followedRooms: JSON.stringify([200])
    })
    const follow = createFollow({ storage })
    follow.syncFollowState()

    expect(follow.toggleFollow()).toBe(true)

    expect(follow.isFollowing.value).toBe(true)
    expect(follow.followerCount.value).toBe(4)
    expect(storage.setItem).toHaveBeenCalledWith('followedRooms', JSON.stringify([200, 100]))
    expect(follow.message.success).toHaveBeenCalledWith('已关注主播')
  })

  it('removes a followed streamer and keeps count non-negative', () => {
    const storage = createStorage({
      followedRooms: JSON.stringify([100, 200])
    })
    const follow = createFollow({
      room: { userId: 100, followerCount: 0 },
      storage
    })
    follow.syncFollowState()

    expect(follow.toggleFollow()).toBe(true)

    expect(follow.isFollowing.value).toBe(false)
    expect(follow.followerCount.value).toBe(0)
    expect(storage.setItem).toHaveBeenCalledWith('followedRooms', JSON.stringify([200]))
    expect(follow.message.success).toHaveBeenCalledWith('已取消关注')
  })

  it('handles invalid storage payloads and missing rooms defensively', () => {
    const storage = createStorage({
      followedRooms: 'not-json'
    })
    const follow = createFollow({ storage })

    expect(follow.syncFollowState(null)).toBe(false)
    expect(follow.isFollowing.value).toBe(false)
    expect(follow.followerCount.value).toBe(0)

    expect(follow.syncFollowState()).toBe(true)
    expect(follow.isFollowing.value).toBe(false)
    expect(follow.toggleFollow()).toBe(true)
    expect(storage.setItem).toHaveBeenCalledWith('followedRooms', JSON.stringify([100]))
  })

  it('normalizes invalid follower counts to zero', () => {
    const follow = createFollow({
      room: { userId: 100, followerCount: 'not-a-number' }
    })

    expect(follow.syncFollowState()).toBe(true)
    expect(follow.followerCount.value).toBe(0)
  })
})
