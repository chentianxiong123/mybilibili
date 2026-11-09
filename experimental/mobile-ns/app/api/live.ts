import api from './client'

export function getLiveIndexData() {
  return api.get('/api/live/index')
}

export function getLiveList(category?: string) {
  return api.get(`/api/live/list${category ? `?category=${category}` : ''}`)
}

export function getRoomInfo(roomId: string | number) {
  return api.get(`/api/live/room/${roomId}`)
}

export function getDanMuConfig(roomId: string | number) {
  return api.get(`/api/live/danmu/${roomId}/config`)
}

export function getLiveAreas() {
  return api.get('/api/live/areas')
}