import api from './client'

export function search(keyword: string, page: number = 1) {
  return api.get(`/api/search?keyword=${encodeURIComponent(keyword)}&page=${page}`)
}

export function getHotwords() {
  return api.get('/api/search/hotwords')
}

export function getHotRank() {
  return api.get('/api/search/hot-rank')
}