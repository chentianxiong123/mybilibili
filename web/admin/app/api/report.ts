import request from './client'

export const getReportList = (params) => {
  return request({
    url: '/moderation/admin/report/list',
    method: 'get',
    params
  })
}

export const processReport = (id, data) => {
  return request({
    url: `/moderation/admin/report/process/${id}`,
    method: 'put',
    data
  })
}
