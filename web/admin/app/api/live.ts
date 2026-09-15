import api from './client'

// ====== Admin 直播管理 ======
export function getLiveRooms(page = 1, size = 10, status = '') {
  return api.get('/live/admin/rooms', { params: { page, size, status } })
}
export function getLiveRoom(id: number) {
  return api.get(`/live/admin/rooms/${id}`)
}
export function updateLiveRoomStatus(id: number, status: string) {
  return api.put(`/live/admin/rooms/${id}/status`, { status })
}
export function getLiveStats() {
  return api.get('/live/admin/stats')
}

export const liveApi = {
  createRoom(roomName) {
    return api.post('/live/room/create', { roomName })
  },
  getMyRoom() {
    return api.get('/live/room/my')
  },
  getLiveList() {
    return api.get('/live/room/list')
  },
  getRoom(id) {
    return api.get(`/live/room/${id}`)
  },
  updateRoomStatus(id, status) {
    return api.put(`/live/room/${id}/status`, { status })
  },
  updateRoom(id, data) {
    return api.put(`/live/room/${id}`, data)
  },
  uploadCover(roomId, file) {
    const form = new FormData()
    form.append('file', file)
    form.append('roomId', roomId)
    return api.post('/live/room/cover', form, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
  scheduleRoom(id, scheduledAt) {
    return api.put(`/live/room/${id}/schedule`, { scheduledAt })
  }
}