import api from './client'

function hasToken() {
  return !!localStorage.getItem('token')
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
