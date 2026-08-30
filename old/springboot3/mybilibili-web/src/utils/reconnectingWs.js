// 简易 WebSocket 自动重连封装
// 用法:
//   const conn = createReconnectingWS({
//     url: () => `ws://...`,
//     onMessage: (data) => {...},
//     onOpen: () => {...},   // 每次（含重连）建连成功都会调用，用于重新 join 房间
//     onClose: () => {...},
//     maxDelay: 10000
//   })
//   conn.send(payloadObj)   // 自动 JSON.stringify
//   conn.close()            // 主动关闭，不再重连

export function createReconnectingWS({ url, onMessage, onOpen, onClose, maxDelay = 10000 }) {
  let ws = null
  let stopped = false
  let retry = 0
  let retryTimer = null
  let pingTimer = null

  const connect = () => {
    if (stopped) return
    try {
      ws = new WebSocket(typeof url === 'function' ? url() : url)
    } catch (e) {
      schedule()
      return
    }

    ws.onopen = () => {
      retry = 0
      try { onOpen && onOpen() } catch (e) { console.error(e) }
      // 每 25s 发一次 ping 防止反代/防火墙静默断连
      clearInterval(pingTimer)
      pingTimer = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          try { ws.send(JSON.stringify({ type: 'ping' })) } catch (e) {}
        }
      }, 25000)
    }

    ws.onmessage = (event) => {
      let msg
      try { msg = JSON.parse(event.data) } catch { return }
      if (msg && msg.type === 'pong') return
      try { onMessage && onMessage(msg) } catch (e) { console.error(e) }
    }

    ws.onclose = () => {
      clearInterval(pingTimer)
      try { onClose && onClose() } catch (e) {}
      schedule()
    }

    ws.onerror = () => {
      // 让 onclose 处理重连
      try { ws && ws.close() } catch (e) {}
    }
  }

  const schedule = () => {
    if (stopped) return
    clearTimeout(retryTimer)
    const delay = Math.min(maxDelay, 1000 * Math.pow(2, retry))
    retry++
    retryTimer = setTimeout(connect, delay)
  }

  const send = (obj) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      try { ws.send(typeof obj === 'string' ? obj : JSON.stringify(obj)); return true } catch (e) {}
    }
    return false
  }

  const close = () => {
    stopped = true
    clearTimeout(retryTimer)
    clearInterval(pingTimer)
    if (ws) {
      try { ws.close() } catch (e) {}
      ws = null
    }
  }

  const isOpen = () => ws && ws.readyState === WebSocket.OPEN

  connect()

  return { send, close, isOpen, _raw: () => ws }
}
