import api from './index.js'

export const meetingApi = {
  createRoom(roomName) {
    return api.post('/meeting/create', null, { params: { roomName } })
  },
  getRoom(roomCode) {
    return api.get(`/meeting/room/${roomCode}`)
  },
  getMyRooms() {
    return api.get('/meeting/my-rooms')
  },
  joinRoom(roomCode) {
    return api.post(`/meeting/join/${roomCode}`)
  },
  leaveRoom(roomId) {
    return api.post(`/meeting/leave/${roomId}`)
  },
  getParticipants(roomId) {
    return api.get(`/meeting/participants/${roomId}`)
  }
}
