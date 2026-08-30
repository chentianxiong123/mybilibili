/**
 * Customer Service AI API - SSE流式通信
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
  const { onData, onDone, onError } = callbacks
  switch (event) {
    case 'data':  if (onData) onData(data); break
    case 'done':  if (onDone) { try { onDone(JSON.parse(data)) } catch (e) { onDone(data) } }; break
    case 'error': if (onError) onError(data); break
  }
}

function getCustomerHeaders() {
  const userId = localStorage.getItem('user_id')
  return {
    'X-User-Id': userId || ''
  }
}

export const customerServiceApi = {
  /**
   * 发送消息（SSE流式）
   * @param {string|number} userId
   * @param {string} content
   * @param {object} callbacks - { onData, onDone, onError }
   * @returns {{ abort: function }}
   */
  sendMessage(userId, content, callbacks = {}) {
    const { onData, onDone, onError } = callbacks
    const controller = new AbortController()

    fetch(`${BASE_URL}/api/ai/customer/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-Id': String(userId)
      },
      body: JSON.stringify({ userId, content }),
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
            if (event) handleSSEEvent(event, { onData, onDone, onError })
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

  /**
   * 转人工客服
   * @param {string|number} userId
   * @param {string|number} sessionId
   * @param {string} reason
   * @returns {Promise}
   */
  transferToHuman(userId, sessionId, reason) {
    return fetch(`${BASE_URL}/api/ai/customer/transfer`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-Id': String(userId)
      },
      body: JSON.stringify({ userId, sessionId, reason })
    })
    .then(response => {
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      return response.json()
    })
    .then(data => {
      if (data.code === 200 || data.code === 0) {
        return data
      }
      throw new Error(data.message || '转人工失败')
    })
  },

  /**
   * 获取历史消息
   * @param {string|number} userId
   * @returns {Promise<Array>}
   */
  getHistory(userId) {
    return fetch(`${BASE_URL}/api/ai/customer/history/${userId}`, {
      method: 'GET',
      headers: {
        'X-User-Id': String(userId)
      }
    })
    .then(response => {
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      return response.json()
    })
    .then(data => {
      if (data.code === 200 || data.code === 0) {
        return data.data || []
      }
      throw new Error(data.message || '获取历史消息失败')
    })
  }
}