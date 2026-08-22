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
    const res = await api.get('/message/unread/counts')
    return {
      code: '1',
      data: res?.data || res || {}
    }
  } catch (e) {
    return { code: '0', data: {} }
  }
}
