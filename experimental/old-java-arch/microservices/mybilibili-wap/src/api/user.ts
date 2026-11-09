import api from './client'

export function normalizeUser(raw: any) {
  const user = raw?.user || raw?.data?.user || raw?.data || raw || {}
  return {
    ...user,
    id: user.id || user.userId || user.mid,
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
  const userStr = localStorage.getItem('user')
  if (!userStr) return null
  try {
    const localUser = normalizeUser(JSON.parse(userStr))
    return localUser.id || null
  } catch (e) {
    return null
  }
}

export async function getMyInfo() {
  try {
    const userStr = localStorage.getItem('user')
    if (!userStr) return { code: '0', data: null }
    const localUser = normalizeUser(JSON.parse(userStr))
    // 兼容可能存在的 id 或 userId 或 user.id 嵌套结构
    const userId = localUser.id
    if (!userId) return { code: '0', data: null }

    const res = await api.get(`/user/${userId}`)
    // 兼容后端 Result 统一封装格式
    const data = normalizeUser(res?.data || res)
    if (data) {
      localStorage.setItem('user', JSON.stringify(data))
    }
    return {
      code: '1',
      data: data
    }
  } catch (e) {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      return { code: '1', data: normalizeUser(JSON.parse(userStr)) }
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
      localStorage.setItem('user', JSON.stringify(data))
    }
    return { code: '1', data }
  } catch (e) {
    return { code: '0', data: null }
  }
}
