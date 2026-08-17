import { getToken } from '../utils/auth'
import api from './client'

export const aiSummaryApi = {
  streamSummary(videoId: number, callbacks: Record<string, any> = {}) {
    const { onStart, onData, onDone, onError, onMeta } = callbacks

    const baseURL = window.location.origin
    const url = `${baseURL}/api/ai/summary/stream/${videoId}`

    const token = getToken()

    const controller = new AbortController()

    fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Accept': 'text/event-stream'
      },
      signal: controller.signal
    })
    .then(response => {
      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('未登录或登录已过期')
        }
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      const readStream = () => {
        reader.read().then(({ done, value }) => {
          if (done) {
            if (onDone) onDone('摘要生成完成')
            return
          }

          buffer += decoder.decode(value, { stream: true })

          const lines = buffer.split('\n\n')
          buffer = lines.pop() || ''

          lines.forEach(line => {
            const event = parseSSEEvent(line)
            if (event) {
              handleSSEEvent(event, { onStart, onData, onDone, onError, onMeta })
            }
          })

          readStream()
        }).catch(error => {
          console.error('读取流错误:', error)
          if (onError) onError(error.message || '读取流失败')
        })
      }

      readStream()
    })
    .catch(error => {
      console.error('SSE请求错误:', error)
      if (error.name === 'AbortError') {
        console.log('请求已取消')
      } else {
        if (onError) onError(error.message || '连接失败，请稍后重试')
      }
    })

    return {
      abort: () => controller.abort(),
      close: () => controller.abort()
    }
  },

  getSummary(videoId: number) {
    return api.get(`/ai/summary/${videoId}`)
  },

  checkSummary(videoId: number) {
    return api.get(`/ai/summary/check/${videoId}`)
  }
}

function parseSSEEvent(raw: string) {
  const lines = raw.split('\n')
  let event = ''
  let data = ''

  lines.forEach(line => {
    if (line.startsWith('event:')) {
      event = line.substring(6).trim()
    } else if (line.startsWith('data:')) {
      data = line.substring(5).trim()
    }
  })

  if (!event && !data) return null

  return { event: event || 'message', data }
}

function handleSSEEvent({ event, data }: { event?: string; data?: string }, callbacks: Record<string, any>) {
  const { onStart, onData, onDone, onError, onMeta } = callbacks

  switch (event) {
    case 'start':
      if (onStart) onStart(data)
      break
    case 'data':
      if (onData) {
        try {
          const decoded = decodeURIComponent(escape(atob(data)))
          onData(decoded)
        } catch (e) {
          console.error('解码数据失败:', e)
          onData(data)
        }
      }
      break
    case 'meta':
      if (onMeta) {
        try {
          const meta = JSON.parse(data)
          onMeta(meta)
        } catch (e) {
          console.error('解析元数据失败:', e)
        }
      }
      break
    case 'done':
      if (onDone) onDone(data)
      break
    case 'error':
      if (onError) onError(data)
      break
    default:
      console.log('未知事件类型:', event, data)
  }
}