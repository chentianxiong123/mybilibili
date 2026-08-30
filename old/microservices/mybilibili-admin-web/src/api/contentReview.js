import request from './request'

export const getPendingList = (params) => {
  return request({
    url: '/admin/content-review/pending',
    method: 'get',
    params
  })
}

export const getAllContent = (params) => {
  return request({
    url: '/admin/content-review/all',
    method: 'get',
    params
  })
}

export const restoreContent = (type, id) => {
  return request({
    url: `/admin/content-review/restore/${type}/${id}`,
    method: 'put'
  })
}

export const deleteContent = (type, id) => {
  return request({
    url: `/admin/content-review/${type}/${id}`,
    method: 'delete'
  })
}

export const batchProcess = (data) => {
  return request({
    url: '/admin/content-review/batch',
    method: 'post',
    data
  })
}
