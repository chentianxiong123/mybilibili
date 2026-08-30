import request from './request'

// 获取安全设置
export const getSecuritySettings = () => {
  return request({
    url: '/admin/security-settings',
    method: 'get'
  })
}

// 更新安全设置
export const updateSecuritySettings = (data) => {
  return request({
    url: '/admin/security-settings',
    method: 'put',
    data
  })
}

// 登录日志API
export const adminLoginLogApi = {
  getLoginLogs: (params) => {
    return request({
      url: '/admin/login-logs/list',
      method: 'get',
      params
    })
  },
  getUserLoginLogs: (userId, page, size) => {
    return request({
      url: `/admin/login-logs/user/${userId}`,
      method: 'get',
      params: { page, size }
    })
  }
}