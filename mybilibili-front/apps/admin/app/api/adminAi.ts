import { getAdminToken } from '../utils/auth'

const BASE_URL = '/api/v1'

function parseSSEEvent(raw: string) {
  const lines = raw.split('\n')
  let event = ''
  let data = ''
  lines.forEach(line => {
    if (line.startsWith('event:')) event = line.substring(6).trim()
    else if (line.startsWith('data:')) data = line.substring(5).trim()
  })
  if (!event && !data) return null
  return { event: event || 'message', data }
}

function handleSSEEvent({ event, data }: { event: string, data: string }, callbacks: any) {
  const { onData, onDone, onError, onToolCall } = callbacks
  switch (event) {
    case 'data': if (onData) onData(data); break
    case 'tool_call': if (onToolCall) onToolCall(data); break
    case 'done': if (onDone) { try { onDone(JSON.parse(data)) } catch (e) { onDone(data) } }; break
    case 'error': if (onError) onError(data); break
  }
}

function getAuthHeaders() {
  const token = getAdminToken()
  const adminId = localStorage.getItem('admin_id')
  return {
    'Authorization': token ? `Bearer ${token}` : '',
    'X-Admin-Id': adminId || ''
  }
}

export const adminAiApi = {
  sendMessage(content: string, callbacks: any = {}) {
    const { onData, onDone, onError, onToolCall } = callbacks
    const controller = new AbortController()

    fetch(`${BASE_URL}/ai/assistant/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify({ content }),
      signal: controller.signal
    })
    .then(async response => {
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const ct = response.headers.get('content-type') || ''
      if (ct.includes('application/json')) {
        const text = await response.text()
        let data: any = {}
        try { data = JSON.parse(text) } catch (e) { onError?.('响应解析失败'); return }
        if (data.event === 'stream' && Array.isArray(data.parts)) {
          data.parts.forEach((p: string) => onData?.(p))
          const reply = data.parts.join('')
          onDone?.(reply)
        } else {
          onDone?.(data)
        }
        return
      }
      const reader = response.body?.getReader()
      if (!reader) return
      const decoder = new TextDecoder()
      let buffer = ''

      const read = () => {
        return reader.read().then(({ done, value }) => {
          if (done) {
            if (buffer.trim()) {
              let data: any = {}
              try { data = JSON.parse(buffer) } catch { /* ignore */ }
              onDone?.(data)
            }
            return
          }
          buffer += decoder.decode(value, { stream: true })
          const parts = buffer.split('\n\n')
          buffer = parts.pop() || ''
          parts.forEach(part => {
            const event = parseSSEEvent(part)
            if (event) handleSSEEvent(event, { onData, onDone, onError, onToolCall })
          })
          return read()
        }).catch(err => {
          if (err.name !== 'AbortError' && onError) onError(err.message)
        })
      }
      return read()
    })
    .catch(err => {
      if (err.name !== 'AbortError' && onError) onError(err.message)
    })

    return { abort: () => controller.abort() }
  }
}