import api from './index.js'

const authRequired = (data) => (
  localStorage.getItem('token')
    ? null
    : Promise.resolve({ code: 401, message: '请先登录', data })
)

export const messageApi = {
  getConversations: () => authRequired([]) || api.get('/message/conversations'),

  getConversationDetail: (id) => authRequired(null) || api.get(`/message/conversations/${id}`),

  deleteConversation: (id) => authRequired(null) || api.delete(`/message/conversations/${id}`),

  getMessages: (conversationId, page = 1, size = 20) =>
    authRequired([]) || api.get(`/message/conversations/${conversationId}/messages`, { params: { page, size } }),

  sendMessage: (data) => authRequired(null) || api.post('/message/send', data),

  markAsRead: (messageId) => authRequired(null) || api.put(`/message/${messageId}/read`),

  batchMarkAsRead: (ids) => authRequired(null) || api.put('/message/batch/read', { ids }),

  getUnreadCounts: () => authRequired({ private: 0, reply: 0, at: 0, like: 0, system: 0, dynamic: 0 }) || api.get('/message/unread/counts'),

  getMessageSettings: () => authRequired(null) || api.get('/message/settings'),

  updateMessageSettings: (data) => authRequired(null) || api.put('/message/settings', data),

  deleteMessage: (messageId) => authRequired(null) || api.delete(`/message/${messageId}`),

  // 回复我的
  getReplies: (params) => authRequired([]) || api.get('/message/replies', { params }),

  // 收到的赞
  getLikes: (params) => authRequired([]) || api.get('/message/likes', { params }),

  // 系统通知
  getSystemNotifications: (params) => authRequired([]) || api.get('/message/system', { params }),

  // 管理员：全站系统通知广播
  broadcastSystemNotification: (data) => api.post('/message/admin/system/broadcast', data)
}
