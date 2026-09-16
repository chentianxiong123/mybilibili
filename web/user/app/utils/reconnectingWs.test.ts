import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createReconnectingWS } from './reconnectingWs'

class FakeWebSocket {
  static instances: FakeWebSocket[] = []
  static OPEN = 1
  static CLOSED = 3
  static CONNECTING = 0
  static CLOSING = 2

  url: string
  readyState = FakeWebSocket.CONNECTING
  sent: (string | ArrayBufferLike | Blob | ArrayBufferView)[] = []
  onopen: ((ev: Event) => any) | null = null
  onmessage: ((ev: { data: any }) => any) | null = null
  onclose: ((ev: CloseEvent) => any) | null = null
  onerror: ((ev: Event) => any) | null = null

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  send(payload: string | ArrayBufferLike | Blob | ArrayBufferView) {
    if (this.readyState !== FakeWebSocket.OPEN) {
      throw new Error('WebSocket is not open')
    }
    this.sent.push(payload)
  }

  close() {
    if (this.readyState === FakeWebSocket.CLOSED) return
    this.readyState = FakeWebSocket.CLOSED
    this.onclose?.({} as CloseEvent)
  }

  // test helpers
  triggerOpen() {
    this.readyState = FakeWebSocket.OPEN
    this.onopen?.({} as Event)
  }
  triggerMessage(data: any) {
    this.onmessage?.({ data })
  }
  triggerClose() {
    this.readyState = FakeWebSocket.CLOSED
    this.onclose?.({} as CloseEvent)
  }
  triggerError() {
    this.onerror?.({} as Event)
  }
}

describe('createReconnectingWS', () => {
  let instances: FakeWebSocket[]
  beforeEach(() => {
    vi.useFakeTimers()
    FakeWebSocket.instances = []
    instances = FakeWebSocket.instances
    vi.stubGlobal('WebSocket', FakeWebSocket)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  describe('基础创建', () => {
    it('立即调用 new WebSocket(url)', () => {
      createReconnectingWS({ url: 'ws://test' })
      expect(FakeWebSocket.instances.length).toBe(1)
      expect(FakeWebSocket.instances[0].url).toBe('ws://test')
    })

    it('url 为函数时调用后再建连', () => {
      const urlFn = vi.fn(() => 'ws://dynamic')
      createReconnectingWS({ url: urlFn })
      expect(urlFn).toHaveBeenCalledTimes(1)
      expect(FakeWebSocket.instances[0].url).toBe('ws://dynamic')
    })

    it('isOpen 在未开时返回 false', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      expect(conn.isOpen()).toBe(false)
    })

    it('isOpen 在打开后返回 true', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      FakeWebSocket.instances[0].triggerOpen()
      expect(conn.isOpen()).toBe(true)
    })
  })

  describe('回调', () => {
    it('onopen 在每次建连成功时调用', () => {
      const onOpen = vi.fn()
      createReconnectingWS({ url: 'ws://t', onOpen })
      FakeWebSocket.instances[0].triggerOpen()
      expect(onOpen).toHaveBeenCalledTimes(1)
    })

    it('onmessage 在收到 JSON 消息时被调用', () => {
      const onMessage = vi.fn()
      createReconnectingWS({ url: 'ws://t', onMessage })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      ws.triggerMessage(JSON.stringify({ type: 'msg', data: 'hi' }))
      expect(onMessage).toHaveBeenCalledWith({ type: 'msg', data: 'hi' })
    })

    it('onmessage 收到无效 JSON 时不抛错', () => {
      const onMessage = vi.fn()
      const logger = { error: vi.fn() }
      createReconnectingWS({ url: 'ws://t', onMessage, logger })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      expect(() => ws.triggerMessage('not-json')).not.toThrow()
      expect(onMessage).not.toHaveBeenCalled()
    })

    it('pong 消息被忽略，不触发 onMessage', () => {
      const onMessage = vi.fn()
      createReconnectingWS({ url: 'ws://t', onMessage })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      ws.triggerMessage(JSON.stringify({ type: 'pong' }))
      expect(onMessage).not.toHaveBeenCalled()
    })

    it('onclose 在关闭时被调用', () => {
      const onClose = vi.fn()
      createReconnectingWS({ url: 'ws://t', onClose })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      ws.triggerClose()
      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('handler 抛错时被 logger.error 捕获，不影响后续', () => {
      const logger = { error: vi.fn() }
      const onMessage = vi.fn(() => {
        throw new Error('boom')
      })
      createReconnectingWS({ url: 'ws://t', onMessage, logger })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      ws.triggerMessage(JSON.stringify({ type: 'x' }))
      expect(logger.error).toHaveBeenCalled()
    })
  })

  describe('send', () => {
    it('已打开时发送对象会 JSON.stringify', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      const ok = conn.send({ type: 'hello', v: 1 })
      expect(ok).toBe(true)
      expect(ws.sent.length).toBe(1)
      expect(ws.sent[0]).toBe(JSON.stringify({ type: 'hello', v: 1 }))
    })

    it('已打开时发送 string 原样发出', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      const ok = conn.send('raw')
      expect(ok).toBe(true)
      expect(ws.sent[0]).toBe('raw')
    })

    it('未连接时返回 false 且不抛错', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      const ok = conn.send({ type: 'x' })
      expect(ok).toBe(false)
      expect(FakeWebSocket.instances[0].sent.length).toBe(0)
    })

    it('send 抛错时 logger.error 被调用并返回 false', () => {
      const logger = { error: vi.fn() }
      const conn = createReconnectingWS({ url: 'ws://t', logger })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      ws.send = () => {
        throw new Error('send fail')
      }
      const ok = conn.send('x')
      expect(ok).toBe(false)
      expect(logger.error).toHaveBeenCalled()
    })
  })

  describe('close', () => {
    it('close 后 ws 被关闭', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      const closeSpy = vi.spyOn(ws, 'close')
      conn.close()
      expect(closeSpy).toHaveBeenCalled()
    })

    it('close 后不再重连', () => {
      createReconnectingWS({ url: 'ws://t' })
      FakeWebSocket.instances[0].triggerClose()
      // schedule a retry - should not happen since not stopped
      vi.advanceTimersByTime(5000)
      // now stop and close
      const conn = createReconnectingWS({ url: 'ws://t' })
      conn.close()
      const before = FakeWebSocket.instances.length
      vi.advanceTimersByTime(30000)
      expect(FakeWebSocket.instances.length).toBe(before)
    })

    it('close 时清理已有重试 timer', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      FakeWebSocket.instances[0].triggerClose()
      // schedule() set retry timer at 1000ms
      vi.advanceTimersByTime(500)
      conn.close()
      // 再推进时间，不应再创建新连接
      const before = FakeWebSocket.instances.length
      vi.advanceTimersByTime(10000)
      expect(FakeWebSocket.instances.length).toBe(before)
    })
  })

  describe('重连 & 指数退避', () => {
    it('首次断开后 1000ms 重连', () => {
      createReconnectingWS({ url: 'ws://t' })
      FakeWebSocket.instances[0].triggerClose()
      vi.advanceTimersByTime(999)
      expect(FakeWebSocket.instances.length).toBe(1)
      vi.advanceTimersByTime(1)
      expect(FakeWebSocket.instances.length).toBe(2)
    })

    it('退避按 1000, 2000, 4000, 8000 增长后被 maxDelay 截断', () => {
      createReconnectingWS({ url: 'ws://t', maxDelay: 5000 })
      // 第一次断开：1000ms 后重连
      FakeWebSocket.instances[0].triggerClose()
      vi.advanceTimersByTime(1000)
      expect(FakeWebSocket.instances.length).toBe(2)
      // 第二次断开：2000ms 后重连
      FakeWebSocket.instances[1].triggerClose()
      vi.advanceTimersByTime(2000)
      expect(FakeWebSocket.instances.length).toBe(3)
      // 第三次断开：4000ms 后重连
      FakeWebSocket.instances[2].triggerClose()
      vi.advanceTimersByTime(4000)
      expect(FakeWebSocket.instances.length).toBe(4)
      // 第四次断开：原本是 8000，被 maxDelay=5000 截断为 5000
      FakeWebSocket.instances[3].triggerClose()
      vi.advanceTimersByTime(4999)
      expect(FakeWebSocket.instances.length).toBe(4)
      vi.advanceTimersByTime(1)
      expect(FakeWebSocket.instances.length).toBe(5)
    })

    it('onopen 成功后 retry 计数器重置', () => {
      createReconnectingWS({ url: 'ws://t', maxDelay: 100000 })
      FakeWebSocket.instances[0].triggerClose()
      vi.advanceTimersByTime(1000)
      FakeWebSocket.instances[1].triggerClose()
      vi.advanceTimersByTime(2000)
      // 第3次断开，但因前一次已 onopen 重置，应在 1000ms 后重连
      FakeWebSocket.instances[2].triggerOpen()
      FakeWebSocket.instances[2].triggerClose()
      vi.advanceTimersByTime(999)
      expect(FakeWebSocket.instances.length).toBe(3)
      vi.advanceTimersByTime(1)
      expect(FakeWebSocket.instances.length).toBe(4)
    })

    it('new WebSocket 抛错时也 schedule 重连', () => {
      class ThrowingWS {
        constructor() {
          throw new Error('boom')
        }
      }
      vi.stubGlobal('WebSocket', ThrowingWS)
      createReconnectingWS({ url: 'ws://t' })
      // 未创建任何实例
      vi.advanceTimersByTime(1000)
      // 第二次调用应再次抛
      vi.advanceTimersByTime(2000)
    })
  })

  describe('心跳', () => {
    it('连接打开后每 25000ms 发 ping', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      expect(ws.sent.length).toBe(0)
      vi.advanceTimersByTime(24999)
      expect(ws.sent.length).toBe(0)
      vi.advanceTimersByTime(1)
      expect(ws.sent.length).toBe(1)
      expect(ws.sent[0]).toBe(JSON.stringify({ type: 'ping' }))
      vi.advanceTimersByTime(25000)
      expect(ws.sent.length).toBe(2)
    })

    it('连接关闭后停止心跳', () => {
      createReconnectingWS({ url: 'ws://t' })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      vi.advanceTimersByTime(25000)
      // open 25000ms 后发了第 1 次 ping
      expect(ws.sent.length).toBe(1)
      ws.triggerClose()
      // schedule 重连在 1000ms 之后
      vi.advanceTimersByTime(1000)
      const ws2 = FakeWebSocket.instances[1]
      // 第1个 ws 已 close，心跳被清理，再推 30s 不会发新 ping
      vi.advanceTimersByTime(30000)
      expect(ws.sent.length).toBe(1)
      // 第二个 ws 没 open，不应发 ping
      expect(ws2.sent.length).toBe(0)
    })

    it('新连接重新开启心跳', () => {
      createReconnectingWS({ url: 'ws://t' })
      const ws1 = FakeWebSocket.instances[0]
      ws1.triggerOpen()
      vi.advanceTimersByTime(25000)
      expect(ws1.sent.length).toBe(1)
      ws1.triggerClose()
      vi.advanceTimersByTime(1000)
      const ws2 = FakeWebSocket.instances[1]
      ws2.triggerOpen()
      vi.advanceTimersByTime(25000)
      expect(ws2.sent.length).toBe(1)
    })
  })

  describe('onerror', () => {
    it('onerror 触发后会调 ws.close()，由 onclose 接管重连', () => {
      createReconnectingWS({ url: 'ws://t' })
      const ws = FakeWebSocket.instances[0]
      ws.triggerOpen()
      ws.triggerError()
      // 应通过 ws.close() -> onclose -> schedule
      vi.advanceTimersByTime(1000)
      expect(FakeWebSocket.instances.length).toBe(2)
    })
  })

  describe('_raw', () => {
    it('返回内部 ws 实例', () => {
      const conn = createReconnectingWS({ url: 'ws://t' })
      expect(conn._raw()).toBe(FakeWebSocket.instances[0])
    })
  })
})