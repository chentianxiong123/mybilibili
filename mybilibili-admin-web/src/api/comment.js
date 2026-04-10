import request from './request'

export const getCommentList = (params) => {
  return request({
    url: '/comment/admin/list',
    method: 'get',
    params
  })
}

export const getCommentById = (id) => {
  return request({
    url: `/comment/admin/${id}`,
    method: 'get'
  })
}

export const deleteComment = (id) => {
  return request({
    url: `/comment/admin/${id}`,
    method: 'delete'
  })
}

export const updateCommentStatus = (id, status) => {
  return request({
    url: `/comment/admin/${id}/status`,
    method: 'put',
    params: { status }
  })
}
