import request from './request'

export const getReportList = (params) => {
  return request({
    url: '/admin/report/list',
    method: 'get',
    params
  })
}

export const processReport = (id, data) => {
  return request({
    url: `/admin/report/process/${id}`,
    method: 'put',
    data
  })
}
