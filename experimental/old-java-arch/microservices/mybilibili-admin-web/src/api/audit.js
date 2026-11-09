import request from './request'

export const getAuditLogs = (params) => {
  return request({
    url: '/admin/audit-logs/list',
    method: 'get',
    params
  })
}

export const getAuditLogDetail = (id) => {
  return request({
    url: `/admin/audit-logs/${id}`,
    method: 'get'
  })
}
