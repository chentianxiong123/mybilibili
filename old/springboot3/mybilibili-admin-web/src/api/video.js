import request from './request'

export const getVideoList = (params) => {
  return request({
    url: '/video/admin/list',
    method: 'get',
    params
  })
}

export const getVideoById = (id) => {
  return request({
    url: `/video/admin/${id}`,
    method: 'get'
  })
}

export const getVideoProgress = (videoId) => {
  return request({
    url: `/video/admin/progress/${videoId}`,
    method: 'get'
  })
}

export const deleteVideo = (id) => {
  return request({
    url: `/video/admin/${id}`,
    method: 'delete'
  })
}

export const deleteVideos = (ids) => {
  return request({
    url: '/video/admin/batch',
    method: 'delete',
    data: ids
  })
}
