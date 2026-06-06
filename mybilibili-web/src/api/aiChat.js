/**
 * AI客服 API — 支持RESTful和SSE流式通信
 */

const BASE_URL = window.location.origin
const CUSTOMER_CONVERSATION_ID = 'customer-service'

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
  const { onStart, onData, onDone, onError, onTransfer } = callbacks
  switch (event) {
    case 'start': if (onStart) onStart(data); break
    case 'data':  if (onData) onData(data); break
    case 'transfer': if (onTransfer) onTransfer(data); break
    case 'done':  if (onDone) { try { onDone(JSON.parse(data)) } catch (e) { onDone(data) } }; break
    case 'error': if (onError) onError(data); break
  }
}

function getAuthHeaders() {
  const token = localStorage.getItem('token')
  const user = getCurrentUser()
  return {
    'Authorization': token ? `Bearer ${token}` : '',
    'X-User-Id': user?.id ? String(user.id) : ''
  }
}

function getCurrentUser() {
  try {
    return JSON.parse(localStorage.getItem('user') || '{}')
  } catch (e) {
    return {}
  }
}

function normalizeMessage(message) {
  const role = String(message.role || '').toLowerCase()
  return {
    id: message.id || `${role || 'message'}-${message.createdAt || Date.now()}`,
    role: role === 'user' ? 'user' : role === 'system' ? 'system' : 'assistant',
    content: message.content || '',
    createdAt: message.createdAt || new Date().toISOString()
  }
}

export const aiChatApi = {
  getConversations() {
    const user = getCurrentUser()
    if (!user?.id) {
      return Promise.resolve({ code: 401, message: '请先登录', data: [] })
    }
    return fetch(`${BASE_URL}/api/ai/customer/history/${user.id}`, {
      headers: getAuthHeaders()
    })
      .then(r => r.json())
      .then(res => {
        const messages = res.data || []
        return {
          code: res.code,
          message: res.message,
          data: messages.length > 0 ? [{
            id: CUSTOMER_CONVERSATION_ID,
            title: 'AI客服对话',
            updatedAt: messages[messages.length - 1]?.createdAt
          }] : []
        }
      })
  },

  createConversation() {
    const user = getCurrentUser()
    if (!user?.id) {
      return Promise.resolve({ code: 401, message: '请先登录', data: null })
    }
    return Promise.resolve({
      code: 200,
      data: {
        id: CUSTOMER_CONVERSATION_ID,
        title: 'AI客服对话',
        createdAt: new Date().toISOString()
      }
    })
  },

  getMessages(conversationId) {
    const user = getCurrentUser()
    if (!user?.id) {
      return Promise.resolve({ code: 401, message: '请先登录', data: [] })
    }
    return fetch(`${BASE_URL}/api/ai/customer/history/${user.id}`, {
      headers: getAuthHeaders()
    })
      .then(r => r.json())
      .then(res => ({
        code: res.code,
        message: res.message,
        data: (res.data || []).map(normalizeMessage)
      }))
  },

  sendMessage(conversationId, content, callbacks = {}) {
    const { onStart, onData, onDone, onError, onTransfer } = callbacks
    const controller = new AbortController()
    const user = getCurrentUser()

    if (!user?.id) {
      if (onError) onError('请先登录')
      return { abort: () => controller.abort() }
    }

    fetch(`${BASE_URL}/api/ai/customer/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify({ userId: user.id, content }),
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
            if (event) handleSSEEvent(event, { onStart, onData, onDone, onError, onTransfer })
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
  },

  transferToHuman(reason = '用户主动转人工') {
    const user = getCurrentUser()
    if (!user?.id) {
      return Promise.resolve({ code: 401, message: '请先登录', data: null })
    }
    return fetch(`${BASE_URL}/api/ai/customer/transfer`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify({ userId: user.id, reason })
    }).then(r => r.json())
  }
}
