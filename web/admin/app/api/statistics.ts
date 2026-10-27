import request from './client'

export const getOverviewStatistics = () => {
  return request({
    url: '/statistics/overview',
    method: 'get'
  })
}

export const getManuscriptStatusStatistics = () => {
  return request({
    url: '/statistics/manuscript/status',
    method: 'get'
  })
}

export const getRecentManuscripts = (limit = 10) => {
  return request({
    url: '/statistics/manuscript/recent',
    method: 'get',
    params: { limit }
  })
}
