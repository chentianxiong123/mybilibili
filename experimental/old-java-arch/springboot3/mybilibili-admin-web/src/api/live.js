import request from './request'

export function getLiveRooms(page = 1, size = 10, status = '') {
  return request({
    url: '/admin/live/rooms',
    method: 'get',
    params: { page, size, status }
  })
}

export function getLiveRoom(id) {
  return request({ url: `/admin/live/rooms/${id}`, method: 'get' })
}

export function updateLiveRoomStatus(id, status) {
  return request({ url: `/admin/live/rooms/${id}/status`, method: 'put', data: { status } })
}

export function getLiveStats() {
  return request({ url: '/admin/live/stats', method: 'get' })
}
