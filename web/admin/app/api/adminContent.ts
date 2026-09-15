import request from './client'

// ====== 评论区管理（内容审核中心） ======

export const getAdminComments = (params) => {
  return request({
    url: '/moderation/admin/comments',
    method: 'get',
    params
  })
}

export const deleteAdminComment = (type, id) => {
  return request({
    url: `/moderation/admin/comments/${type}/${id}`,
    method: 'delete'
  })
}

export const restoreAdminComment = (type, id) => {
  return request({
    url: `/moderation/admin/comments/${type}/${id}`,
    method: 'put'
  })
}

// ====== 弹幕管理（内容审核中心） ======

export const getAdminDanmaku = (params) => {
  return request({
    url: '/moderation/admin/danmaku',
    method: 'get',
    params
  })
}

export const deleteAdminDanmaku = (id) => {
  return request({
    url: `/moderation/admin/danmaku/${id}`,
    method: 'delete'
  })
}
