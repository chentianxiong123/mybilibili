import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 热搜词
export async function getHotwords() {
  const res = await axios.get(`${API_BASE}/search/hotword`)
  return res.data
}

// 搜索建议
export async function getSuggests(keyword: string) {
  const res = await axios.get(`${API_BASE}/search/suggest`, { params: { w: keyword } })
  return res.data
}

// 搜索结果
export async function getSearchResult(params: {
  keyword: string
  page?: number
  size?: number
  searchType?: string
  order?: string
}) {
  const res = await axios.post(`${API_BASE}/search`, params)
  return res.data
}