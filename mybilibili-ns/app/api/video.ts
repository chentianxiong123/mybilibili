import api from './client'

export function getVideoInfo(aId: number) {
  return api.get(`/api/videos/${aId}`)
}

export function getRecommendVides(aId: number) {
  return api.get(`/api/videos/${aId}/recommend`)
}

export function getComments(aId: number, page: number = 1) {
  return api.get(`/api/comments/${aId}?page=${page}`)
}

export function getBarrages(aId: number) {
  return api.get(`/api/barrages/${aId}`)
}