/**
 * Admin AI Assistant API — SSE流式通信
 */

const BASE_URL = window.location.origin

function parseSSEEvent(raw) {
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

function handleSSEEvent({ event, data }, callbacks) {
  const { onData, onDone, onError, onToolCall } = callbacks
  switch (event) {
    case 'data':  if (onData) onData(data); break
    case 'tool_call': if (onToolCall) onToolCall(data); break
    case 'done':  if (onDone) { try { onDone(JSON.parse(data)) } catch (e) { onDone(data) } }; break
    case 'error': if (onError) onError(data); break
  }
}

function getAuthHeaders() {
  const token = localStorage.getItem('admin_token')
  const adminId = localStorage.getItem('admin_id')
  return {
    'Authorization': token ? `Bearer ${token}` : '',
    'X-Admin-Id': adminId || ''
  }
}

export const adminAiApi = {
  sendMessage(content, callbacks = {}) {
    const { onData, onDone, onError, onToolCall } = callbacks
    const controller = new AbortController()

    fetch(`${BASE_URL}/api/ai/admin/assistant/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify({ content }),
      signal: controller.signal
    })
    .then(response => {
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      const read = () => {
        return reader.read().then(({ done, value }) => {
          if (done) return
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
