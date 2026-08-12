import request from './client'

export const getTicketList = (params) => {
  return request.get('/operation/admin/tickets', { params })
}

export const getTicketById = (id) => {
  return request.get(`/operation/admin/tickets/${id}`)
}

export const processTicket = (id, adminReply) => {
  return request.put(`/operation/admin/tickets/${id}/process`, { adminReply })
}

export const deleteTicket = (id) => {
  return request.delete(`/operation/admin/tickets/${id}`)
}
