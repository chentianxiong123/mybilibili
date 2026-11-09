import request from './request'

export const getUserList = (params) => {
  return request({
    url: '/user/admin/list',
    method: 'get',
    params
  })
}

export const getUserById = (id) => {
  return request({
    url: `/user/admin/${id}`,
    method: 'get'
  })
}

export const updateUserStatus = (id, status) => {
  return request({
    url: `/user/admin/${id}/status`,
    method: 'put',
    params: { status }
  })
}

export const resetPassword = (id, newPassword) => {
  return request({
    url: `/user/admin/${id}/password`,
    method: 'put',
    params: { newPassword }
  })
}
