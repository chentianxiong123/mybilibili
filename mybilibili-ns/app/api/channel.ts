import api from './client'

export function getChannelVideos(rId: number, page: number = 1) {
  return api.get(`/api/channel/${rId}/videos?page=${page}`)
}