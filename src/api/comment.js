import request from './request'

// 获取评论列表
export const getCommentList = (params) => {
  return request({
    url: '/comments',
    method: 'get',
    params
  })
}

// 获取评论详情
export const getCommentById = (id) => {
  return request({
    url: `/comments/${id}`,
    method: 'get'
  })
}

// 删除评论
export const deleteComment = (id) => {
  return request({
    url: `/comments/${id}`,
    method: 'delete'
  })
}

// 更新评论状态
export const updateCommentStatus = (id, status) => {
  return request({
    url: `/comments/${id}/status`,
    method: 'put',
    params: { status }
  })
}