/**
 * AI客服 API — 支持RESTful和SSE流式通信
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
  const { onStart, onData, onDone, onError } = callbacks
  switch (event) {
    case 'start': if (onStart) onStart(data); break
    case 'data':  if (onData) onData(data); break
    case 'done':  if (onDone) { try { onDone(JSON.parse(data)) } catch (e) { onDone(data) } }; break
    case 'error': if (onError) onError(data); break
  }
}

function getAuthHeaders() {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  let userId = ''
  try { userId = userStr ? JSON.parse(userStr).id || '' : '' } catch (e) {}
  return {
    'Authorization': token ? `Bearer ${token}` : '',
    'X-User-Id': String(userId)
  }
}

export const aiChatApi = {
  getConversations() {
    return fetch(`${BASE_URL}/api/ai/chat/conversations`, {
      headers: getAuthHeaders()
    }).then(r => r.json())
  },

  createConversation() {
    return fetch(`${BASE_URL}/api/ai/chat/conversations`, {
      method: 'POST',
      headers: getAuthHeaders()
    }).then(r => r.json())
  },

  deleteConversation(id) {
    return fetch(`${BASE_URL}/api/ai/chat/conversations/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    }).then(r => r.json())
  },

  getMessages(conversationId) {
    return fetch(`${BASE_URL}/api/ai/chat/conversations/${conversationId}/messages`, {
      headers: getAuthHeaders()
    }).then(r => r.json())
  },

  sendMessage(conversationId, content, callbacks = {}) {
    const { onStart, onData, onDone, onError } = callbacks
    const controller = new AbortController()

    fetch(`${BASE_URL}/api/ai/chat/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify({ conversationId, content }),
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
            if (event) handleSSEEvent(event, { onStart, onData, onDone, onError })
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