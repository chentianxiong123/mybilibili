// 首页 API - 复用 mybilibili-web 的接口
// /manuscript/recommended → 推荐视频
// /recommend/hot → 热门视频
// /category → 分类列表
// /banner-images/home → 轮播图
import api from './client'

// 字段适配：后端字段 → 移动端期望的字段
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

// 首页推荐 + 排行榜 + 分区
export async function getHomeContent() {
  try {
    const recRes = await api.get('/manuscript/recommended').catch(() => ({ code: 500, data: [] }))
    const hotRes = await api.get('/recommend/hot?size=30').catch(() => ({ code: 500, data: [] }))
    const catRes = await api.get('/category').catch(() => ({ code: 500, data: [] }))

    return {
      code: '1',
      data: {
        oneLevelPartitions: (catRes?.data || catRes || []).map?.(c => ({ id: c.id, name: c.name })) || [],
        additionalVideos: (recRes?.data || recRes || []).map(adaptVideo),
        rankingVideos: (hotRes?.data || hotRes || []).map(adaptVideo)
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
    const list = res?.data || res || []
    return {
      code: '1',
      data: list.map(b => ({
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
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 视频列表
export async function getVideoList(page = 1, size = 20) {
  try {
    const res = await api.get(`/manuscript/list?page=${page}&size=${size}`)
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 热门视频
export async function getHotVideos(categoryId, size = 10) {
  try {
    let url = `/recommend/hot?size=${size}`
    if (categoryId) url += `&categoryId=${categoryId}`
    const res = await api.get(url)
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptVideo)
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
      api.get(`/recommend/hot?size=30${rId ? `&categoryId=${rId}` : ''}`).catch(() => ({ code: 500, data: [] })),
      rId ? api.get(`/manuscript/category/${rId}`).catch(() => ({ code: 500, data: [] })) : Promise.resolve({ code: 200, data: [] })
    ])
    const hotList = hotRes?.data || hotRes || []
    const catList = catRes?.data || catRes || []
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