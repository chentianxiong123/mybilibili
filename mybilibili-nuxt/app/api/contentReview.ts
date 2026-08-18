import request from './client'

export const getPendingList = (params) => {
  return request({
    url: '/moderation/admin/pending',
    method: 'get',
    params
  })
}

export const getAllContent = (params) => {
  return request({
    url: '/moderation/admin/all',
    method: 'get',
    params
  })
}

export const restoreContent = (type, id) => {
  return request({
    url: `/moderation/admin/restore/${type}/${id}`,
    method: 'put'
  })
}

export const deleteContent = (type, id) => {
  return request({
    url: `/moderation/admin/${type}/${id}`,
    method: 'delete'
  })
}

export const batchProcess = (data) => {
  return request({
    url: '/moderation/admin/batch',
    method: 'post',
    data
  })
}
