import { ref, onMounted, onUnmounted } from 'vue'
import { createReconnectingWS } from '../utils/reconnectingWs.js'
import { ElNotification } from 'element-plus'

const notificationWs = ref(null)
const unreadCounts = ref({
  private: 0,
  reply: 0,
  at: 0,
  like: 0,
  system: 0,
  total: 0
})
const isConnected = ref(false)
let onNotificationCallbacks = []

function getWsBaseUrl() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}`
}

function connect() {
  const token = localStorage.getItem('token')
  if (!token) return

  if (notificationWs.value) {
    notificationWs.value.close()
  }

  notificationWs.value = createReconnectingWS({
    url: () => `${getWsBaseUrl()}/ws/notification?token=${encodeURIComponent(localStorage.getItem('token'))}`,
    onOpen: () => {
      isConnected.value = true
    },
    onClose: () => {
      isConnected.value = false
    },
    onMessage: (msg) => {
      if (msg.type === 'unread_count') {
        unreadCounts.value = { ...unreadCounts.value, ...msg.counts }
      } else if (msg.type === 'notification') {
        handleNotification(msg)
      }
    }
  })
}

function disconnect() {
  if (notificationWs.value) {
    notificationWs.value.close()
    notificationWs.value = null
  }
  isConnected.value = false
}

function handleNotification(msg) {
  const typeLabels = {
    private: '新私信',
    reply: '新回复',
    like: '新点赞',
    system: '系统通知',
    at: '有人@你'
  }
  const title = typeLabels[msg.notificationType] || '新通知'
  ElNotification({
    title,
    message: msg.content?.length > 50 ? msg.content.substring(0, 50) + '...' : msg.content,
    type: 'info',
    duration: 4000,
    position: 'top-right'
  })
  onNotificationCallbacks.forEach(cb => {
    try { cb(msg) } catch (e) { console.error(e) }
  })
}

export function useNotificationWs() {
  return {
    unreadCounts,
    isConnected,
    connect,
    disconnect,
    onNotification(cb) {
      onNotificationCallbacks.push(cb)
      return () => {
        onNotificationCallbacks = onNotificationCallbacks.filter(c => c !== cb)
      }
    },
    updateCounts(counts) {
      unreadCounts.value = { ...unreadCounts.value, ...counts }
    }
  }
}
