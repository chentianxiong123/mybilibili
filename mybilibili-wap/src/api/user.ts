import api from './client'
import { getLocalUser } from '../utils/session'
import storage, { K } from '../utils/storage_layer'

export function normalizeUser(raw: any) {
  const user = raw?.user || raw?.data?.user || raw?.data || raw || {}
  return {
    ...user,
    id: user.id || user.userId || user.mid || user.user_id,
    username: user.username || user.name || '',
    nickname: user.nickname || user.username || user.name || '',
    avatar: user.avatar || user.avatarUrl || user.face || '',
    bio: user.bio || user.sign || '',
    signature: user.signature || user.bio || user.sign || '',
    followerCount: user.followerCount ?? user.follower ?? user.followers ?? user.fans ?? 0,
    followingCount: user.followingCount ?? user.following ?? user.attentions ?? 0,
    dynamicCount: user.dynamicCount ?? user.dynamics ?? 0,
    manuscriptCount: user.manuscriptCount ?? user.videos ?? user.videoCount ?? 0,
    coinCount: user.coinCount ?? user.coins ?? 0,
    level: user.level || 1
  }
}

export function getLocalUserId() {
  const localUser = getLocalUser()
  if (!localUser) return null
  try {
    return normalizeUser(localUser).id || null
  } catch (e) {
    return null
  }
}

export async function getMyInfo() {
  try {
    const localUser = getLocalUser()
    if (!localUser) return { code: '0', data: null }
    const normalized = normalizeUser(localUser)
    // 兼容可能存在的 id 或 userId 或 user.id 嵌套结构
    const userId = normalized.id
    if (!userId) return { code: '0', data: null }

    const res = await api.get(`/user/${userId}`)
    // 兼容后端 Result 统一封装格式
    const data = normalizeUser(res?.data || res)
    if (data) {
      storage.set(K.user, data)
    }
    return {
      code: '1',
      data: data
    }
  } catch (e) {
    const local = getLocalUser()
    if (local) {
      return { code: '1', data: normalizeUser(local) }
    }
    return { code: '0', data: null }
  }
}

export async function getFollowingList(userId: number) {
  try {
    const res = await api.get(`/user/${userId}/following`)
    const list = res?.data || res || []
    return { code: '1', data: Array.isArray(list) ? list.map(normalizeUser) : [] }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getFollowerList(userId: number) {
  try {
    const res = await api.get(`/user/${userId}/followers`)
    const list = res?.data || res || []
    return { code: '1', data: Array.isArray(list) ? list.map(normalizeUser) : [] }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function updateMyInfo(userId: number, payload: Record<string, any>) {
  try {
    const res = await api.put(`/user/${userId}`, payload)
    const data = normalizeUser(res?.data || res)
    if (data) {
      storage.set(K.user, data)
    }
    return { code: '1', data }
  } catch (e) {
    return { code: '0', data: null }
  }
}
