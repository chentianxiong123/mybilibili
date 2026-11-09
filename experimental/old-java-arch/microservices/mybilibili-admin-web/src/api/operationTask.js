import request from './request'

export const getOperationTasks = (params) => {
  return request({
    url: '/admin/operation-tasks/list',
    method: 'get',
    params
  })
}

export const getOperationTaskDetail = (id) => {
  return request({
    url: `/admin/operation-tasks/${id}`,
    method: 'get'
  })
}
