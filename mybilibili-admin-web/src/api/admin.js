import request from './request'

export const adminLogin = (data) => {
  return request({
    url: '/admin/login',
    method: 'post',
    data
  })
}

export const adminRegister = (data) => {
  return request({
    url: '/admin/register',
    method: 'post',
    data
  })
}

export const getAdminList = () => {
  return request({
    url: '/admin/list',
    method: 'get'
  })
}

export const getAdminById = (id) => {
  return request({
    url: `/admin/${id}`,
    method: 'get'
  })
}

export const getAdminRoles = (id) => {
  return request({
    url: `/admin/${id}/roles`,
    method: 'get'
  })
}

export const setAdminRoles = (id, roleIds) => {
  return request({
    url: `/admin/${id}/roles`,
    method: 'put',
    data: roleIds
  })
}