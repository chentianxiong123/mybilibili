import request from './request'

// 获取角色列表
export const getRoleList = () => {
  return request({
    url: '/roles',
    method: 'get'
  })
}

// 获取角色详情
export const getRoleById = (id) => {
  return request({
    url: `/roles/${id}`,
    method: 'get'
  })
}

// 添加角色
export const addRole = (data) => {
  return request({
    url: '/roles',
    method: 'post',
    data
  })
}

// 更新角色
export const updateRole = (id, data) => {
  return request({
    url: `/roles/${id}`,
    method: 'put',
    data
  })
}

// 删除角色
export const deleteRole = (id) => {
  return request({
    url: `/roles/${id}`,
    method: 'delete'
  })
}

// 获取角色权限
export const getRolePermissions = (id) => {
  return request({
    url: `/roles/${id}/permissions`,
    method: 'get'
  })
}

// 设置角色权限
export const setRolePermissions = (id, permissionIds) => {
  return request({
    url: `/roles/${id}/permissions`,
    method: 'put',
    data: permissionIds
  })
}

// 获取所有权限
export const getAllPermissions = () => {
  return request({
    url: '/roles/permissions/all',
    method: 'get'
  })
}