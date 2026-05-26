import { recommendApi } from '../../api/recommend.js'
import { videoApi } from '../../api/index.js'
import { categoryApi } from '../../api/index.js'

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
  videoReview: v.danmakuCount || 0
})

export async function getRankings(rId) {
  try {
    const res = await categoryApi.getCategoryList()
    return {
      code: '1',
      data: (handleRes(res) || []).map(c => ({ id: c.id, name: c.name }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getRankingArchive(params) {
  try {
    const { rId = 0, p = 1, size = 20 } = params || {}
    let res
    if (rId) {
      res = await videoApi.getVideosByCategoryId(rId)
    } else {
      res = await videoApi.getVideoList(p, size)
    }
    return {
      code: '1',
      data: (handleRes(res) || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}