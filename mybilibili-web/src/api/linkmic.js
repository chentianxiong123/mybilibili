import { request } from '../utils/request'

export const linkmicApi = {
  applyLinkmic: (roomId) => {
    return request.post(`/api/live/linkmic/apply/${roomId}`)
  },
  acceptLinkmic: (linkmicId) => {
    return request.post(`/api/live/linkmic/accept/${linkmicId}`)
  },
  rejectLinkmic: (linkmicId) => {
    return request.post(`/api/live/linkmic/reject/${linkmicId}`)
  },
  disconnectLinkmic: (linkmicId) => {
    return request.post(`/api/live/linkmic/disconnect/${linkmicId}`)
  },
  toggleAudio: (linkmicId, enabled) => {
    return request.post(`/api/live/linkmic/toggle-audio/${linkmicId}`, null, { params: { enabled } })
  },
  toggleVideo: (linkmicId, enabled) => {
    return request.post(`/api/live/linkmic/toggle-video/${linkmicId}`, null, { params: { enabled } })
  },
  getActiveLinkmics: (roomId) => {
    return request.get(`/api/live/linkmic/active/${roomId}`)
  },
  getPendingApplications: (roomId) => {
    return request.get(`/api/live/linkmic/pending/${roomId}`)
  }
}