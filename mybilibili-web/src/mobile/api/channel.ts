import { categoryApi } from '../../api/index.js'
import { videoApi } from '../../api/index.js'
import { recommendApi } from '../../api/recommend.js'

const handleRes = (res) => {
  if (res && res.code === 200) return res.data || {}
  return {}
}

const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || '',
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0
})

export async function getRankingPartitions() {
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

export async function getRankingRegion(rId) {
  try {
    const [hotRes, catRes] = await Promise.all([
      recommendApi.getHotVideos(rId, 30),
      rId ? videoApi.getVideosByCategoryId(rId) : Promise.resolve({ code: 200, data: [] })
    ])
    const hotList = handleRes(hotRes) || []
    const catList = handleRes(catRes) || []
    return {
      code: '1',
      data: {
        hotVideos: hotList.slice(0, 4).map(adaptVideo),
        partitions: [{
          id: rId,
          name: '当前分类',
          videos: catList.map(adaptVideo)
        }]
      }
    }
  } catch (e) {
    return { code: '0', data: { hotVideos: [], partitions: [] } }
  }
}

export async function getPartitions() {
  return getRankingPartitions()
}