import request from './request'

export function getMeetingRooms(page = 1, size = 10, status = null) {
  const params = { page, size }
  if (status !== null && status !== '') {
    params.status = status
  }
  return request({
    url: '/admin/meeting/rooms',
    method: 'get',
    params
  })
}

export function getPendingMeetingReservations() {
  return request({
    url: '/admin/meeting/pending',
    method: 'get'
  })
}

export function approveMeetingReservation(roomId) {
  return request({
    url: `/admin/meeting/approve/${roomId}`,
    method: 'post'
  })
}

export function rejectMeetingReservation(roomId) {
  return request({
    url: `/admin/meeting/reject/${roomId}`,
    method: 'post'
  })
}

export function forceEndMeeting(roomId) {
  return request({
    url: `/admin/meeting/end/${roomId}`,
    method: 'post'
  })
}
