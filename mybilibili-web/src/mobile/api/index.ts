// 首页 API - 使用项目已有 API
import { getHomeBanners } from '../../api/banner.js'
import { recommendApi } from '../../api/recommend.js'
import { categoryApi } from '../../api/category.js'

// 统一处理响应码：项目 code=200，旧代码期望 code='1'
const handleRes = (res) => {
  if (res && res.code === 200) return res.data || {}
  return {}
}

// 字段适配：项目字段 → 移动端期望的字段
const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || v.author || '',
  mid: v.userId,
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0,
  duration: v.duration,
  description: v.description
})

// 首页推荐 + 排行榜视频
export async function getHomeContent() {
  try {
    const [recRes, hotRes, catRes] = await Promise.all([
      recommendApi.getRecommendedVideos(20),
      recommendApi.getHotVideos(null, 30),
      categoryApi.getCategoryList()
    ])

    return {
      code: '1',
      data: {
        oneLevelPartitions: handleRes(catRes).map?.(c => ({ id: c.id, name: c.name })) || [],
        additionalVideos: (handleRes(recRes) || []).map(adaptVideo),
        rankingVideos: (handleRes(hotRes) || []).map(adaptVideo)
      }
    }
  } catch (e) {
    console.error('getHomeContent error:', e)
    return { code: '0', data: {} }
  }
}

// 轮播图
export async function getBanners() {
  try {
    const res = await getHomeBanners()
    return {
      code: '1',
      data: (handleRes(res) || []).map(b => ({
        id: b.id,
        name: b.title || '',
        pic: b.imageUrl || b.url,
        url: b.link || b.targetUrl || '#'
      }))
    }
  } catch (e) {
    console.error('getBanners error:', e)
    return { code: '0', data: [] }
  }
}