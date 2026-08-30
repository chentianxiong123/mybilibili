/**
 * 浏览器 Web Notification 工具
 * 用于后台运行时的新消息提醒
 */

let permission = 'default'

export const requestNotificationPermission = async () => {
  if (!('Notification' in window)) return false
  if (Notification.permission === 'granted') {
    permission = 'granted'
    return true
  }
  if (Notification.permission !== 'denied') {
    const res = await Notification.requestPermission()
    permission = res
    return res === 'granted'
  }
  return false
}

export const isNotificationSupported = () => 'Notification' in window

export const isNotificationEnabled = () => Notification.permission === 'granted'

/**
 * 显示浏览器通知
 * @param {string} title - 通知标题
 * @param {object} options - Notification options: body, icon, tag, data
 */
export const showNotification = (title, options = {}) => {
  if (!isNotificationEnabled()) return null
  const notif = new Notification(title, {
    icon: '/favicon.ico',
    badge: '/favicon.ico',
    ...options
  })
  notif.onclick = () => {
    window.focus()
    notif.close()
    if (options.data?.url) {
      window.location.href = options.data.url
    }
  }
  setTimeout(() => notif.close(), 8000)
  return notif
}

/**
 * 会议邀请通知
 */
export const notifyMeetingInvite = (meetingName, fromUser, roomCode) => {
  return showNotification(`会议邀请：${meetingName}`, {
    body: `${fromUser} 邀请你加入会议`,
    tag: 'meeting-invite',
    data: { url: `/meeting/${roomCode}` }
  })
}

/**
 * 主播开播通知
 */
export const notifyStreamLive = (roomName, streamerName) => {
  return showNotification(`${streamerName} 开播了！`, {
    body: `${roomName} 正在进行直播，点击进入`,
    tag: 'stream-live',
    data: { url: `/live/${roomName}` }
  })
}

/**
 * 被@提及通知
 */
export const notifyMention = (fromUser, roomName, text) => {
  return showNotification(`${fromUser} 在直播间提及你`, {
    body: text.length > 60 ? text.slice(0, 60) + '...' : text,
    tag: 'mention',
    data: { url: `/live/${roomName}` }
  })
}
