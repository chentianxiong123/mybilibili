import api from './index.ts'

// ====== Admin 会议管理 ======
export function getMeetingRooms(page = 1, size = 10, status: string | null = null) {
  const params: Record<string, any> = { page, size }
  if (status !== null && status !== '') params.status = status
  return api.get('/admin/meeting/rooms', { params })
}
export function getPendingMeetingReservations() {
  return api.get('/admin/meeting/pending')
}
export function approveMeetingReservation(id: number) {
  return api.post(`/admin/meeting/approve/${id}`)
}
export function rejectMeetingReservation(id: number) {
  return api.post(`/admin/meeting/reject/${id}`)
}
export function forceEndMeeting(id: number) {
  return api.post(`/admin/meeting/force-end/${id}`)
}

export const meetingApi = {
  createRoom(roomName) {
    return api.post('/meeting/create', null, { params: { roomName } })
  },
  reserveRoom(data) {
    return api.post('/meeting/reserve', data)
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
  endRoom(roomId) {
    return api.post(`/meeting/end/${roomId}`)
  },
  getParticipants(roomId) {
    return api.get(`/meeting/participants/${roomId}`)
  }
}
