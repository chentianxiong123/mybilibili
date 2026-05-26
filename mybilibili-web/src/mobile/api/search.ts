import { searchApi } from '../../api/search.js'

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

export async function getHotwords() {
  try {
    const res = await searchApi.getHotSearch()
    const data = handleRes(res) || []
    return {
      code: '1',
      data: data.map(item => ({ keyword: item.keyword || item }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getSuggests(keyword) {
  try {
    const res = await searchApi.getSearchSuggestions(keyword)
    return {
      code: '1',
      data: (handleRes(res) || []).map(item => ({
        name: item.keyword || item,
        value: item.keyword || item
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getSearchResult(params) {
  try {
    const { keyword = '', page = 1, size = 20, searchType = 'all', order = 'totalrank' } = params || {}
    let sort = 'relevance'
    if (order === 'click') sort = 'hot'
    if (order === 'pubdate') sort = 'time'
    const res = await searchApi.searchVideos({ keyword, page, size, sort })
    return {
      code: '1',
      data: (handleRes(res) || []).map(adaptVideo)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}