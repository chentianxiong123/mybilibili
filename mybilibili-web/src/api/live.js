import api from './index.js'

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
  }
}