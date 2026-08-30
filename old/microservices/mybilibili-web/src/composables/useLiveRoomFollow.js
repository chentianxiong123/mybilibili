import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { sameUserId } from '../utils/userId.js'

const FOLLOWED_ROOMS_KEY = 'followedRooms'

const readFollowedIds = (storage, storageKey) => {
  try {
    const parsed = JSON.parse(storage?.getItem(storageKey) || '[]')
    return Array.isArray(parsed) ? parsed : []
  } catch (e) {
    return []
  }
}

const normalizeCount = value => {
  const count = Number(value || 0)
  return Number.isFinite(count) ? count : 0
}

export function useLiveRoomFollow({
  room,
  storage = typeof localStorage !== 'undefined' ? localStorage : null,
  message = ElMessage,
  storageKey = FOLLOWED_ROOMS_KEY
}) {
  const isFollowing = ref(false)
  const followerCount = ref(0)

  const syncFollowState = (roomData = room.value) => {
    if (!roomData) {
      isFollowing.value = false
      followerCount.value = 0
      return false
    }
    const followed = readFollowedIds(storage, storageKey)
    isFollowing.value = followed.some(id => sameUserId(id, roomData.userId))
    followerCount.value = normalizeCount(roomData.followerCount)
    return true
  }

  const persistFollowedIds = (ids) => {
    storage?.setItem?.(storageKey, JSON.stringify(ids))
  }

  const toggleFollow = () => {
    const roomData = room.value
    if (!roomData?.userId) return false

    const followed = readFollowedIds(storage, storageKey)
    if (isFollowing.value) {
      const updated = followed.filter(id => !sameUserId(id, roomData.userId))
      persistFollowedIds(updated)
      isFollowing.value = false
      followerCount.value = Math.max(0, followerCount.value - 1)
      message?.success?.('已取消关注')
      return true
    }

    const updated = followed.some(id => sameUserId(id, roomData.userId))
      ? followed
      : [...followed, roomData.userId]
    persistFollowedIds(updated)
    isFollowing.value = true
    followerCount.value += 1
    message?.success?.('已关注主播')
    return true
  }

  return {
    isFollowing,
    followerCount,
    syncFollowState,
    toggleFollow
  }
}
