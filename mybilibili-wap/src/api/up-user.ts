// UP主 API - 复用 mybilibili-web 的接口
// /user/{id} → 用户信息
// /manuscript/user/{id} → 用户稿件
import api from './client'

const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || '',
  mid: v.userId,
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0,
  duration: v.duration
})

// 用户信息 - 复用 userApi.getUserById() → /user/{id}
export async function getUserInfo(mId) {
  try {
    const res = await api.get(`/user/${mId}`)
    const data = res?.data || res
    if (!data) return { code: '0', data: null }
    return {
      code: '1',
      data: {
        mid: data.id,
        name: data.username || data.nickname || '',
        face: data.avatarUrl || '',
        level: data.level || 1,
        sign: data.bio || '',
        follower: data.followerCount || 0,
        following: data.followingCount || 0,
        sex: data.gender || '保密'
      }
    }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 用户稿件 - 复用 videoApi.getVideosByUserId() → /manuscript/user/{id}
export async function getUserVideos(mId, p = 1, size = 20) {
  try {
    const res = await api.get(`/manuscript/user/${mId}?page=${p}&size=${size}`)
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}