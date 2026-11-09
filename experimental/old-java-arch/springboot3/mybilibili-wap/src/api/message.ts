import api from './client'

export async function getConversations() {
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
