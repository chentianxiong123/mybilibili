// 首页 API - 复用 mybilibili-web 的接口
// /manuscript/recommended → 推荐视频
// /recommend/hot → 热门视频
// /category → 分类列表
// /banner-images/home → 轮播图
import api from './client'

// 视频字段统一适配（兼容 VideoRecommendVO、ManuscriptVO、ManuscriptDocument）
const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id || v.videoId,
  title: v.title || '无标题',
  pic: v.coverUrl || v.cover || '',
  author: v.userName || v.uploader?.name || v.username || v.nickname || v.author || 'UP主',
  mid: v.userId,
  play: v.viewCount || v.play || 0,
  videoReview: v.commentCount || v.videoReview || v.danmakuCount || 0,
  duration: v.duration || v.durationSeconds,
  description: v.description || ''
})

// 首页推荐 + 排行榜 + 分区
export async function getHomeContent() {
  try {
    // 并行请求，不相互阻塞
    const [recRes, catRes] = await Promise.all([
      api.get('/manuscript/recommended').catch(() => ({ code: 500, data: [] })),
      api.get('/category').catch(() => ({ code: 500, data: [] }))
    ])

    // 热门视频独立请求，不阻塞主内容加载
    const hotRes = await api.get('/manuscript/hot').catch(() => ({ code: 500, data: [] }))
    const hotRaw = hotRes?.data || hotRes || []
    const hotList = Array.isArray(hotRaw) ? hotRaw : (hotRaw?.list || [])
    console.log('[getHomeContent] hotRes:', hotRes, 'hotList:', hotList)

    return {
      code: '1',
      data: {
        oneLevelPartitions: (catRes?.data || catRes || []).map?.(c => ({ id: c.id, name: c.name })) || [],
        additionalVideos: (recRes?.data || recRes || []).map(adaptVideo),
        rankingVideos: hotList.map(adaptVideo)
      }
    }
  } catch (e) {
    console.error('getHomeContent error:', e)
    return { code: '0', data: {} }
  }
}

// 轮播图 - 复用 getHomeBanners() → /banner-images/home
export async function getBanners() {
  try {
    const res = await api.get('/banner-images/home')
    const rawData = res?.data || res
    const list = Array.isArray(rawData) ? rawData : (rawData?.list || [])
    return {
      code: '1',
      data: list.map(b => ({
        id: b.id,
        name: b.title || '',
        pic: b.imageUrl,
        url: b.linkUrl || '#'
      }))
    }
  } catch (e) {
    console.error('getBanners error:', e)
    return { code: '0', data: [] }
  }
}

// 历史记录 - 适配后端 WatchHistoryController → /watch-history
export const historyApi = {
  getHistoryList: (page = 1, size = 20) => api.get(`/watch-history?page=${page}&size=${size}`),
  clearHistory: () => api.delete('/watch-history'),
  deleteHistoryItem: (id) => api.delete(`/watch-history/${id}`)
}

// 分类列表
export async function getCategories() {
  try {
    const res = await api.get('/category')
    return {
      code: '1',
      data: (res?.data || res || []).map(c => ({ id: c.id, name: c.name }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 分类视频
export async function getVideosByCategory(categoryId) {
  try {
    const res = await api.get(`/manuscript/category/${categoryId}`)
    const rawData = res?.data || res
    const list = rawData?.list || (Array.isArray(rawData) ? rawData : [])
    return {
      code: '1',
      data: list.map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 视频列表
export async function getVideoList(page = 1, size = 20) {
  try {
    const res = await api.get(`/manuscript/list?page=${page}&size=${size}`)
    const rawData = res?.data || res
    const list = rawData?.list || (Array.isArray(rawData) ? rawData : [])
    return {
      code: '1',
      data: list.map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 热门视频（复用 manuscriptApi.getHotManuscripts → /manuscript/hot）
export async function getHotVideos(categoryId, size = 10) {
  try {
    const res = await api.get('/manuscript/hot')
    const rawData = res?.data || res || []
    const list = Array.isArray(rawData) ? rawData : (rawData?.list || [])
    return {
      code: '1',
      data: list.map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 频道分类列表
export { getCategories as getRankingPartitions }

// 分类热门视频
export async function getRankingRegion(rId) {
  try {
    const [hotRes, catRes] = await Promise.all([
      api.get('/manuscript/hot').catch(() => ({ code: 500, data: [] })),
      rId ? api.get(`/manuscript/category/${rId}`).catch(() => ({ code: 500, data: null })) : Promise.resolve({ code: 200, data: null })
    ])
    const hotList = hotRes?.data || hotRes || []

    const rawCatData = catRes?.data || catRes
    const catList = rawCatData?.list || (Array.isArray(rawCatData) ? rawCatData : [])

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

// 排行榜分区
export async function getRankings() {
  return getCategories()
}

// 排行榜视频
export async function getRankingArchive(params) {
  try {
    const { rId = 0, p = 1, size = 20 } = params || {}
    let res
    if (rId) {
      res = await api.get(`/manuscript/category/${rId}?page=${p}&size=${size}`)
    } else {
      res = await api.get(`/manuscript/list?page=${p}&size=${size}`)
    }
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}