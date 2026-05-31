import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createReconnectingWS } from '../utils/reconnectingWs.js'

const HEARTBEAT_INTERVAL = 25000

function installFakeWebSocket() {
  const sockets = []

  class FakeWebSocket {
    static CONNECTING = 0
    static OPEN = 1
    static CLOSING = 2
    static CLOSED = 3

    constructor(url) {
      this.url = url
      this.readyState = FakeWebSocket.CONNECTING
      this.send = vi.fn()
      this.close = vi.fn(() => {
        if (this.readyState === FakeWebSocket.CLOSED) return
        this.readyState = FakeWebSocket.CLOSED
        this.onclose && this.onclose({})
      })
      sockets.push(this)
    }

    open() {
      this.readyState = FakeWebSocket.OPEN
      this.onopen && this.onopen({})
    }

    receive(data) {
      this.onmessage && this.onmessage({ data })
    }

    fail() {
      this.onerror && this.onerror({})
    }
  }

  globalThis.WebSocket = FakeWebSocket
  return sockets
}

describe('createReconnectingWS', () => {
  let originalWebSocket
  let sockets

  beforeEach(() => {
    vi.useFakeTimers()
    originalWebSocket = globalThis.WebSocket
    sockets = installFakeWebSocket()
  })

  afterEach(() => {
    globalThis.WebSocket = originalWebSocket
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('opens a socket, sends payloads, and keeps the heartbeat alive', () => {
    const onOpen = vi.fn()
    const conn = createReconnectingWS({
      url: () => 'ws://meeting.example/ws',
      onOpen
    })

    expect(sockets).toHaveLength(1)
    expect(conn.isOpen()).toBe(false)

    sockets[0].open()

    expect(onOpen).toHaveBeenCalled()
    expect(conn.isOpen()).toBe(true)
    expect(conn.send({ type: 'chat', text: 'hello' })).toBe(true)
    expect(conn.send('raw-message')).toBe(true)
    expect(sockets[0].send).toHaveBeenNthCalledWith(1, '{"type":"chat","text":"hello"}')
    expect(sockets[0].send).toHaveBeenNthCalledWith(2, 'raw-message')

    vi.advanceTimersByTime(HEARTBEAT_INTERVAL)

    expect(sockets[0].send).toHaveBeenNthCalledWith(3, '{"type":"ping"}')
  })

  it('parses messages, ignores pong and invalid payloads, and protects callbacks', () => {
    const logger = { error: vi.fn() }
    const onMessage = vi.fn(() => {
      throw new Error('consumer failed')
    })
    createReconnectingWS({
      url: 'ws://meeting.example/ws',
      onMessage,
      logger
    })
    sockets[0].open()

    sockets[0].receive('{"type":"chat","text":"hello"}')
    sockets[0].receive('{"type":"pong"}')
    sockets[0].receive('not-json')

    expect(onMessage).toHaveBeenCalledTimes(1)
    expect(onMessage).toHaveBeenCalledWith({ type: 'chat', text: 'hello' })
    expect(logger.error).toHaveBeenCalledTimes(1)
  })

  it('reconnects after close or error and stops reconnecting after manual close', () => {
    const onClose = vi.fn()
    const conn = createReconnectingWS({
      url: () => `ws://meeting.example/ws-${sockets.length + 1}`,
      onClose,
      maxDelay: 2000
    })

    sockets[0].open()
    sockets[0].close()

    expect(onClose).toHaveBeenCalledTimes(1)
    expect(sockets).toHaveLength(1)

    vi.advanceTimersByTime(1000)

    expect(sockets).toHaveLength(2)
    expect(sockets[1].url).toBe('ws://meeting.example/ws-2')

    sockets[1].fail()
    expect(sockets[1].close).toHaveBeenCalled()

    vi.advanceTimersByTime(2000)

    expect(sockets).toHaveLength(3)

    conn.close()
    vi.advanceTimersByTime(2000)

    expect(conn._raw()).toBeNull()
    expect(sockets).toHaveLength(3)
  })
})
