// 简易 WebSocket 自动重连封装
// 用法:
//   const conn = createReconnectingWS({
//     url: () => `ws://...`,
//     onMessage: (data) => {...},
//     onOpen: () => {...},   // 每次（含重连）建连成功都会调用，用于重新 join 房间
//     onClose: () => {...},
//     maxDelay: 10000,
//     logger: console
//   })
//   conn.send(payloadObj)   // 自动 JSON.stringify
//   conn.close()            // 主动关闭，不再重连

const BASE_RETRY_DELAY = 1000
const HEARTBEAT_INTERVAL = 25000

function safeCall(handler, logger, ...args) {
  if (typeof handler !== 'function') return
  try {
    handler(...args)
  } catch (error) {
    logger?.error?.(error)
  }
}

function parsePayload(payload) {
  try {
    return JSON.parse(payload)
  } catch {
    return null
  }
}

function serializePayload(payload) {
  return typeof payload === 'string' ? payload : JSON.stringify(payload)
}

export function createReconnectingWS({ url, onMessage, onOpen, onClose, maxDelay = 10000, logger = console }) {
  let ws = null
  let stopped = false
  let retry = 0
  let retryTimer = null
  let pingTimer = null

  const clearPingTimer = () => {
    clearInterval(pingTimer)
    pingTimer = null
  }

  const schedule = () => {
    if (stopped) return
    clearTimeout(retryTimer)
    const delay = Math.min(maxDelay, BASE_RETRY_DELAY * Math.pow(2, retry))
    retry += 1
    retryTimer = setTimeout(connect, delay)
  }

  const startHeartbeat = () => {
    clearPingTimer()
    pingTimer = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        try {
          ws.send(JSON.stringify({ type: 'ping' }))
        } catch (error) {
          logger?.error?.(error)
        }
      }
    }, HEARTBEAT_INTERVAL)
  }

  const connect = () => {
    if (stopped) return
    try {
      ws = new WebSocket(typeof url === 'function' ? url() : url)
    } catch {
      schedule()
      return
    }

    ws.onopen = () => {
      retry = 0
      safeCall(onOpen, logger)
      startHeartbeat()
    }

    ws.onmessage = (event) => {
      const msg = parsePayload(event.data)
      if (!msg || msg.type === 'pong') return
      safeCall(onMessage, logger, msg)
    }

    ws.onclose = () => {
      clearPingTimer()
      safeCall(onClose, logger)
      schedule()
    }

    ws.onerror = () => {
      // 让 onclose 统一处理重连
      try {
        ws && ws.close()
      } catch (error) {
        logger?.error?.(error)
      }
    }
  }

  const send = (payload) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) return false
    try {
      ws.send(serializePayload(payload))
      return true
    } catch (error) {
      logger?.error?.(error)
      return false
    }
  }

  const close = () => {
    stopped = true
    clearTimeout(retryTimer)
    clearPingTimer()
    if (ws) {
      try {
        ws.close()
      } catch (error) {
        logger?.error?.(error)
      }
      ws = null
    }
  }

  const isOpen = () => ws && ws.readyState === WebSocket.OPEN

  connect()

  return { send, close, isOpen, _raw: () => ws }
}
