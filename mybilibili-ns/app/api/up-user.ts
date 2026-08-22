import api from './client'

export function getUpUserInfo(mId: number) {
  return api.get(`/api/up-user/${mId}`)
}

export function getUpUserVideos(mId: number, page: number = 1) {
  return api.get(`/api/up-user/${mId}/videos?page=${page}`)
}