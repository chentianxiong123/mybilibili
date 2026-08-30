// UP主 API - 复用 mybilibili-web 的接口
// /user/{id} → 用户信息
// /manuscript/user/{id} → 用户稿件
import api from './client'
import { normalizeUser } from './user'

const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || '',
  mid: v.userId,
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0,
  duration: v.duration,
  uploadTime: v.uploadTime || v.createdAt || v.updatedAt,
  visibility: v.visibility || v.status
})

// 用户信息 - 复用 userApi.getUserById() → /user/{id}
export async function getUserInfo(mId) {
  try {
    const res = await api.get(`/user/${mId}`)
    const data = normalizeUser(res?.data || res)
    if (!data) return { code: '0', data: null }
    return {
      code: '1',
      data: {
        mid: data.id,
        name: data.nickname || data.username || '',
        face: data.avatar || '',
        level: data.level || 1,
        sign: data.bio || '',
        signature: data.signature || '',
        announcement: data.announcement || '',
        birthdate: data.birthdate || '',
        experience: data.experience || 0,
        coinCount: data.coinCount || 0,
        follower: data.followerCount ?? 0,
        following: data.followingCount ?? 0,
        dynamicCount: data.dynamicCount || 0,
        manuscriptCount: data.manuscriptCount || 0,
        totalViewCount: data.totalViewCount || 0,
        totalLikeCount: data.totalLikeCount || 0,
        pinnedVideo: data.pinnedVideo || null,
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
