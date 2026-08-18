import request from './client'

export const getTicketList = (params) => {
  return request.get('/support/admin/tickets', { params })
}

export const getTicketById = (id) => {
  return request.get(`/support/admin/tickets/${id}`)
}

export const processTicket = (id, adminReply) => {
  return request.put(`/support/admin/tickets/${id}/process`, { adminReply })
}

export const deleteTicket = (id) => {
  return request.delete(`/support/admin/tickets/${id}`)
}
