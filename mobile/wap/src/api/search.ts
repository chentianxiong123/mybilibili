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
// 后端可能返回 keyword 为空串的条目，兜底成整个对象会把模板渲染成 JSON 文本，
// 这里统一拍平成非空字符串再交给视图。
const toKeyword = (item) => {
  if (typeof item === 'string') return item.trim()
  if (item && typeof item === 'object') {
    const k = item.keyword
    return typeof k === 'string' ? k.trim() : ''
  }
  return ''
}

export async function getHotwords() {
  try {
    const res = await api.get('/search/hot')
    const data = res?.data || res || []
    return {
      code: '1',
      data: data.map(item => ({ keyword: toKeyword(item) })).filter(x => x.keyword)
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
      data: data.map(item => {
        const name = toKeyword(item)
        return { name, value: name }
      }).filter(x => x.name)
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
