import request from './request'

export const getFeedbackList = (params) => {
  return request.get('/ai/admin/feedback', { params })
}

export const getFeedbackById = (id) => {
  return request.get(`/ai/admin/feedback/${id}`)
}

export const processFeedback = (id, adminReply) => {
  return request.put(`/ai/admin/feedback/${id}/process`, { adminReply })
}

export const deleteFeedback = (id) => {
  return request.delete(`/ai/admin/feedback/${id}`)
}

export const submitFeedback = (data) => {
  return request.post('/ai/feedback', data)
}