import request from './request'

export const getTicketList = (params) => {
  return request.get('/ai/admin/tickets', { params })
}

export const getTicketById = (id) => {
  return request.get(`/ai/admin/tickets/${id}`)
}

export const processTicket = (id, adminReply) => {
  return request.put(`/ai/admin/tickets/${id}/process`, { adminReply })
}

export const deleteTicket = (id) => {
  return request.delete(`/ai/admin/tickets/${id}`)
}
