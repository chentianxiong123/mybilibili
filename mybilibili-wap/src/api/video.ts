// 视频详情 API - 复用 mybilibili-web 的接口
// /manuscript/{id} → 稿件详情
// /recommend/related/{id} → 相关推荐
// /comment/list → 评论列表
// /danmaku/video/{id} → 弹幕
import api from './client'

const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || '',
  mid: v.userId,
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0,
  duration: v.duration,
  description: v.description
})

// 稿件详情 - 复用 videoApi.getVideoByManuscriptId() → /manuscript/{id}
export async function getVideoInfo(aId) {
  try {
    const res = await api.get(`/manuscript/${aId}`)
    const data = res?.data || res
    if (!data) return { code: '0', data: null }
    return {
      code: '1',
      data: {
        ...adaptVideo(data),
        mid: data.userId,
        author: data.username || data.nickname || '',
        ctime: data.createTime
      }
    }
  } catch (e) {
    console.error(e)
    return { code: '0', data: null }
  }
}

// 相关推荐 - 复用 recommendApi.getRelatedVideos() → /recommend/related/{id}
export async function getRecommendVides(aId) {
  try {
    const res = await api.get(`/recommend/related/${aId}?size=8`)
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 评论列表 - 复用 commentApi.getComments() → /comment/list
export async function getComments(aId, p = 1) {
  try {
    const res = await api.get(`/comment/list?manuscriptId=${aId}&page=${p}&size=20&sort=time`)
    const list = res?.data || res || []
    return {
      code: '1',
      data: list.map(c => ({
        rpid: c.id,
        content: c.content,
        ctime: c.createTime,
        user: {
          name: c.username || c.nickname || '',
          face: c.avatarUrl || ''
        }
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 弹幕 - 复用 interactionApi.getDanmakus() → /danmaku/video/{id}
export async function getBarrages(cId) {
  try {
    const res = await api.get(`/danmaku/video/${cId}`)
    return {
      code: '1',
      data: (res?.data || res || []).map(b => ({
        text: b.content,
        color: b.color || '#fff',
        time: b.time || 0
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}