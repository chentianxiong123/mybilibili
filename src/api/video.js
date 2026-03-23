import request from './request'

export const getVideoList = (params) => {
  return request({
    url: '/videos',
    method: 'get',
    params
  })
}

export const getVideoById = (id) => {
  return request({
    url: `/videos/${id}`,
    method: 'get'
  })
}

export const getVideoProgress = (videoId) => {
  return request({
    url: `/videos/progress/${videoId}`,
    method: 'get'
  })
}

export const deleteVideo = (id) => {
  return request({
    url: `/videos/${id}`,
    method: 'delete'
  })
}

export const deleteVideos = (ids) => {
  return request({
    url: '/videos/batch',
    method: 'delete',
    data: ids
  })
}
