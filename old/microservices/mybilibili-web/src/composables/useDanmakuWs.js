import { ref, onUnmounted } from 'vue'
import { createReconnectingWS } from '../utils/reconnectingWs.js'

function getWsBaseUrl() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}`
}

export function useDanmakuWs() {
  let ws = null
  const isConnected = ref(false)
  let onDanmakuCallback = null

  function connect(videoId) {
    disconnect()
    if (!videoId) return

    const token = localStorage.getItem('token')
    const params = new URLSearchParams({ videoId: videoId.toString() })
    if (token) params.set('token', token)

    ws = createReconnectingWS({
      url: () => `${getWsBaseUrl()}/ws/danmaku?${params.toString()}`,
      onOpen: () => {
        isConnected.value = true
      },
      onClose: () => {
        isConnected.value = false
      },
      onMessage: (msg) => {
        if (msg.type === 'danmaku' && onDanmakuCallback) {
          onDanmakuCallback(msg)
        }
      }
    })
  }

  function disconnect() {
    if (ws) {
      ws.close()
      ws = null
    }
    isConnected.value = false
  }

  function onDanmaku(cb) {
    onDanmakuCallback = cb
  }

  function sendDanmaku(content, time, color, mode) {
    if (!ws || !ws.isOpen()) return false
    return ws.send({
      type: 'danmaku',
      content,
      time,
      color: color || '#ffffff',
      mode: mode || 0
    })
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    connect,
    disconnect,
    onDanmaku,
    sendDanmaku
  }
}
