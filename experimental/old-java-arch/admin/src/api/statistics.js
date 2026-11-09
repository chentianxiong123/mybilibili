import request from './request'

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

export const getVideoPlayStatistics = (startDate, endDate) => {
  return request({
    url: '/statistics/video/play',
    method: 'get',
    params: { startDate, endDate }
  })
}

export const getUserGrowthStatistics = (startDate, endDate) => {
  return request({
    url: '/statistics/user/growth',
    method: 'get',
    params: { startDate, endDate }
  })
}

export const getCommentStatistics = (startDate, endDate) => {
  return request({
    url: '/statistics/comment',
    method: 'get',
    params: { startDate, endDate }
  })
}

export const getHotVideos = (limit = 10) => {
  return request({
    url: '/statistics/video/hot',
    method: 'get',
    params: { limit }
  })
}
