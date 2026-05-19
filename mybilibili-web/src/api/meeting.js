import { request } from '../utils/request'

export const meetingApi = {
  createRoom: (roomName) => {
    return request.post('/api/meeting/create', null, { params: { roomName } })
  },
  getRoom: (roomCode) => {
    return request.get(`/api/meeting/room/${roomCode}`)
  },
  getMyRooms: () => {
    return request.get('/api/meeting/my-rooms')
  },
  joinRoom: (roomCode) => {
    return request.post(`/api/meeting/join/${roomCode}`)
  },
  leaveRoom: (roomId) => {
    return request.post(`/api/meeting/leave/${roomId}`)
  },
  getParticipants: (roomId) => {
    return request.get(`/api/meeting/participants/${roomId}`)
  }
}