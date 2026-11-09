import request from './request'

// 获取待审核列表
export const getPendingList = (params) => {
  return request({
    url: '/admin/content-review/pending',
    method: 'get',
    params
  })
}

// 获取所有内容
export const getAllContent = (params) => {
  return request({
    url: '/admin/content-review/all',
    method: 'get',
    params
  })
}

// 恢复内容
export const restoreContent = (type, id) => {
  return request({
    url: `/admin/content-review/restore/${type}/${id}`,
    method: 'put'
  })
}

// 删除内容
export const deleteContent = (type, id) => {
  return request({
    url: `/admin/content-review/${type}/${id}`,
    method: 'delete'
  })
}

// 批量处理
export const batchProcess = (data) => {
  return request({
    url: '/admin/content-review/batch',
    method: 'post',
    data
  })
}
