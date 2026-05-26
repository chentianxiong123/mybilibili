import { userApi } from '../../api/index.js'

const handleRes = (res) => {
  if (res && res.code === 200) return res.data || {}
  return {}
}

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

export async function getUserInfo(mId) {
  try {
    const res = await userApi.getUserById(mId)
    const data = handleRes(res)
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

export async function getUserVideos(mId, p = 1, size = 20) {
  try {
    const { videoApi } = await import('../../api/index.js')
    const res = await videoApi.getVideosByUserId(mId)
    return {
      code: '1',
      data: (handleRes(res) || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}