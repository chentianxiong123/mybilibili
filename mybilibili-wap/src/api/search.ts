// 搜索 API - 复用 mybilibili-web 的接口
// /search/videos → 搜索视频
// /search/suggest → 搜索建议
// /search/hot → 热搜榜
import api from './client'

const adaptVideo = (v) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || '',
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0
})

// 热搜榜 - 复用 searchApi.getHotSearch() → /search/hot
export async function getHotwords() {
  try {
    const res = await api.get('/search/hot')
    const data = res?.data || res || []
    return {
      code: '1',
      data: data.map(item => ({ keyword: item.keyword || item }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 搜索建议 - 复用 searchApi.getSearchSuggestions() → /search/suggest
export async function getSuggests(keyword) {
  try {
    const res = await api.get(`/search/suggest?keyword=${encodeURIComponent(keyword)}`)
    const data = res?.data || res || []
    return {
      code: '1',
      data: data.map(item => ({
        name: item.keyword || item,
        value: item.keyword || item
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 搜索结果 - 复用 searchApi.searchVideos() → /search/videos
export async function getSearchResult(params) {
  try {
    const { keyword = '', page = 1, size = 20, order = 'totalrank' } = params || {}
    let sort = 'relevance'
    if (order === 'click') sort = 'hot'
    if (order === 'pubdate') sort = 'time'
    const res = await api.get(`/search/videos?keyword=${encodeURIComponent(keyword)}&page=${page}&size=${size}&sort=${sort}`)
    return {
      code: '1',
      data: ((res?.data?.content) || res || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}