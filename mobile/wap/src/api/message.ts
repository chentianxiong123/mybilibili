import api from './client'
import { isLogin } from '../utils/session'

function hasToken() {
  return isLogin()
}

export async function getConversations() {
  if (!hasToken()) return { code: '0', data: [] }
  try {
    const res = await api.get('/message/conversations')
    const raw = res?.data || res || []
    return {
      code: '1',
      // 后端给的是 snake_case（target_user_id / last_message_content …），
      // Message.vue 全部按 camelCase 取值，不换算会把名字、头像、最后一条消息
      // 全渲染成空，并且 openConversation 退到 item.id（会话 id）当 targetUserId，
      // 点进去打开的是别人的会话 → 空白聊天。
      data: (Array.isArray(raw) ? raw : []).map((c: any) => ({
        ...c,
        userId: c.user_id,
        targetUserId: c.target_user_id,
        targetUserNickname: c.target_user_name,
        targetUsername: c.target_user_name,
        targetUserAvatar: c.target_user_avatar,
        lastMessageContent: c.last_message_content,
        lastMessageTime: c.last_message_time,
        unreadCount: c.unread_count
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getUnreadCounts() {
  if (!hasToken()) return { code: '0', data: {} }
  try {
    const res = await api.get('/message/unread')
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
