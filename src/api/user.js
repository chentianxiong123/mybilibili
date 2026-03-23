import request from './request'

// 获取用户列表
export const getUserList = (params) => {
  return request({
    url: '/users',
    method: 'get',
    params
  })
}

// 获取用户详情
export const getUserById = (id) => {
  return request({
    url: `/users/${id}`,
    method: 'get'
  })
}

// 更新用户状态
export const updateUserStatus = (id, status) => {
  return request({
    url: `/users/${id}/status`,
    method: 'put',
    params: { status }
  })
}

// 重置用户密码
export const resetPassword = (id, newPassword) => {
  return request({
    url: `/users/${id}/password`,
    method: 'put',
    params: { newPassword }
  })
}