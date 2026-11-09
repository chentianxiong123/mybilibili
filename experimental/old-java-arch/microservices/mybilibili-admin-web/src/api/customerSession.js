import request from './request'

// 获取待处理会话列表
export const getPendingSessions = () => {
  return request({
    url: '/ai/admin/customer/sessions',
    method: 'get'
  })
}

// 获取会话消息历史
export const getSessionMessages = (sessionId) => {
  return request({
    url: `/ai/admin/customer/sessions/${sessionId}/messages`,
    method: 'get'
  })
}

// 客服发送消息
export const sendReply = (sessionId, adminId, content) => {
  return request({
    url: `/ai/admin/customer/sessions/${sessionId}/reply`,
    method: 'post',
    data: { adminId, content }
  })
}

// 标记会话为已处理
export const resolveSession = (sessionId) => {
  return request({
    url: `/ai/admin/customer/sessions/${sessionId}/resolve`,
    method: 'post'
  })
}

// 获取待处理会话数量
export const getPendingCount = () => {
  return request({
    url: '/ai/admin/customer/sessions/pending/count',
    method: 'get'
  })
}
