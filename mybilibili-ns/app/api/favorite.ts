import api from './client'

export function getFavorites(page: number = 1) {
  return api.get(`/api/favorite/list?page=${page}`)
}