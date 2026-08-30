import api from './index.js'

export const linkmicApi = {
  applyLinkmic(roomId) {
    return api.post(`/live/linkmic/apply/${roomId}`)
  },
  acceptLinkmic(linkmicId) {
    return api.post(`/live/linkmic/accept/${linkmicId}`)
  },
  rejectLinkmic(linkmicId) {
    return api.post(`/live/linkmic/reject/${linkmicId}`)
  },
  disconnectLinkmic(linkmicId) {
    return api.post(`/live/linkmic/disconnect/${linkmicId}`)
  },
  toggleAudio(linkmicId, enabled) {
    return api.post(`/live/linkmic/toggle-audio/${linkmicId}`, null, { params: { enabled } })
  },
  toggleVideo(linkmicId, enabled) {
    return api.post(`/live/linkmic/toggle-video/${linkmicId}`, null, { params: { enabled } })
  },
  getActiveLinkmics(roomId) {
    return api.get(`/live/linkmic/active/${roomId}`)
  },
  getPendingApplications(roomId) {
    return api.get(`/live/linkmic/pending/${roomId}`)
  },
  getQueuePosition(roomId) {
    return api.get(`/live/linkmic/queue-position/${roomId}`)
  }
}
