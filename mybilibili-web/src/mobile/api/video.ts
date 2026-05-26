import { videoApi } from '../../api/index.js'
import { recommendApi } from '../../api/recommend.js'
import { interactionApi } from '../../api/index.js'

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
  duration: v.duration,
  description: v.description
})

const adaptUser = (u) => ({
  name: u.username || u.nickname || '',
  face: u.avatarUrl || ''
})

export async function getVideoInfo(aId) {
  try {
    const res = await videoApi.getVideoByManuscriptId(aId)
    const data = handleRes(res)
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

export async function getRecommendVides(aId) {
  try {
    const res = await recommendApi.getRelatedVideos(aId, 8)
    return {
      code: '1',
      data: (handleRes(res) || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getComments(aId, p = 1) {
  try {
    const { commentApi } = await import('../../api/index.js')
    const res = await commentApi.getComments('MANUSCRIPT', aId, p, 20)
    const list = handleRes(res) || []
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

export async function getBarrages(cId) {
  try {
    const res = await interactionApi.getDanmakus(cId)
    return {
      code: '1',
      data: (handleRes(res) || []).map(b => ({
        text: b.content,
        color: b.color || '#fff',
        time: b.time || 0
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}