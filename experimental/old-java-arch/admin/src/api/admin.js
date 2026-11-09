import request from './request'

// 管理员登录
export const adminLogin = (data) => {
  return request({
    url: '/user/login',
    method: 'post',
    data
  })
}

// 管理员注册
export const adminRegister = (data) => {
  return request({
    url: '/user/register',
    method: 'post',
    data
  })
}

// 获取管理员列表
export const getAdminList = () => {
  return request({
    url: '/user/list',
    method: 'get'
  })
}

// 获取管理员详情
export const getAdminById = (id) => {
  return request({
    url: `/user/${id}`,
    method: 'get'
  })
}

// 获取管理员角色
export const getAdminRoles = (id) => {
  return request({
    url: `/user/${id}/roles`,
    method: 'get'
  })
}

// 设置管理员角色
export const setAdminRoles = (id, roleIds) => {
  return request({
    url: `/user/${id}/roles`,
    method: 'put',
    data: roleIds
  })
}
