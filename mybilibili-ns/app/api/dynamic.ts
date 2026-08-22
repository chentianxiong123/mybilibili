import api from './client'

export function getDynamicFeed(page: number = 1) {
  return api.get(`/api/dynamic/feed?page=${page}`)
}