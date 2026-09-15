// 搜索 API - 复用 mybilibili-web 的接口
// /search/videos → 搜索视频
// /search/suggest → 搜索建议
// /search/hot → 热搜榜
import api from './client'

const adaptVideo = (v) => ({
  aId: v.manuscript_id || v.id || v.manuscriptId,
  title: v.title,
  pic: v.cover_url || v.coverUrl || v.cover,
  author: v.uploader?.name || v.uploader?.nickname || v.uploader?.username || v.username || v.nickname || v.author || '',
  play: v.view_count || v.viewCount || 0,
  videoReview: v.comment_count || v.danmaku_count || v.commentCount || v.danmakuCount || 0,
  duration: v.duration || ''
})

const adaptUpUser = (u) => ({
  mid: u.mid || u.id,
  name: u.name || u.nickname || u.username || '',
  face: u.face || u.avatar || '',
  sign: u.sign || u.signature || '',
  fans: u.fans || u.follower || u.follower_count || 0,
  videos: u.videos || u.manuscript_count || 0
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
// UP主 搜索走 /search/users（按昵称/用户名模糊匹配）
export async function getSearchResult(params) {
  try {
    const { keyword = '', page = 1, size = 20, order = 'totalrank', searchType = 'all' } = params || {}
    if (searchType === 'upuser') {
      const res = await api.get(`/search/users?keyword=${encodeURIComponent(keyword)}&page=${page}&size=${size}`)
      return {
        code: '1',
        data: (res?.data?.list || []).map(adaptUpUser)
      }
    }
    let sort = 'relevance'
    if (order === 'click') sort = 'hot'
    if (order === 'pubdate') sort = 'time'
    const res = await api.get(`/search/videos?keyword=${encodeURIComponent(keyword)}&page=${page}&size=${size}&sort=${sort}`)
    return {
      code: '1',
      data: ((res?.data?.list) || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 搜索历史 - 纯客户端 localStorage 持久化（与 web 端一致，不依赖后端接口）
const HISTORY_KEY = 'searchHistory'
const HISTORY_MAX = 10

export async function fetchSearchHistory() {
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    const data = raw ? JSON.parse(raw) : []
    return { code: '1', data: Array.isArray(data) ? data : [] }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function pushSearchHistory(keyword) {
  if (!keyword) return { code: '0', data: [] }
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    let list = raw ? JSON.parse(raw) : []
    if (!Array.isArray(list)) list = []
    // 去重后置顶
    list = list.filter(x => x !== keyword)
    list.unshift(keyword)
    if (list.length > HISTORY_MAX) list = list.slice(0, HISTORY_MAX)
    localStorage.setItem(HISTORY_KEY, JSON.stringify(list))
    return { code: '1', data: list }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function clearSearchHistory() {
  try {
    localStorage.removeItem(HISTORY_KEY)
    return { code: '1' }
  } catch (e) {
    return { code: '0' }
  }
}