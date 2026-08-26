import api from './client'
import { getToken } from '../utils/session'
import { isLogin as hasToken } from '../utils/session'

function hasToken() {
  return !!getToken()
}

export async function getConversations() {
  if (!hasToken()) return { code: '0', data: [] }
  try {
    const res = await api.get('/message/conversations')
    return {
      code: '1',
      data: res?.data || res || []
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getUnreadCounts() {
  if (!hasToken()) return { code: '0', data: {} }
  try {
    const res = await api.get('/message/unread/')
    return {
      code: '1',
      data: res?.data || res || {}
    }
  } catch (e) {
    return { code: '0', data: {} }
  }
}

// 消息通知列表：reply(回复) / at(@) / like(点赞) / system(系统)
export async function getNotifications(type) {
  if (!hasToken()) return { code: '0', data: [] }
  const pathMap = {
    reply: '/message/replies',
    at: '/message/at',
    like: '/message/likes',
    system: '/message/system'
  }
  const path = pathMap[type]
  if (!path) return { code: '0', data: [] }
  try {
    const res = await api.get(path)
    return {
      code: '1',
      data: res?.data || res || []
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}
