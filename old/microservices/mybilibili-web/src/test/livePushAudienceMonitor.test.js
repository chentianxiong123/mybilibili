import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useLivePushAudienceMonitor } from '../composables/useLivePushAudienceMonitor.js'

const createMonitor = (options = {}) => {
  const streamKey = ref(options.streamKey ?? 'stream-1')
  const me = options.me || { id: 1, name: '主播' }
  let wsOptions = null
  const ws = {
    send: vi.fn(),
    close: vi.fn()
  }
  const createWs = vi.fn(options => {
    wsOptions = options
    return ws
  })
  let timerHandler = null
  const setTimer = vi.fn(handler => {
    timerHandler = handler
    return 'metrics-timer'
  })
  const clearTimer = vi.fn()
  let nowValue = options.nowValue || 1000

  const monitor = useLivePushAudienceMonitor({
    streamKey,
    me,
    createWs,
    getWsUrl: () => 'ws://example/ws/meeting',
    now: () => nowValue,
    random: () => 0.25,
    setTimer,
    clearTimer
  })

  return {
    streamKey,
    me,
    ws,
    createWs,
    setTimer,
    clearTimer,
    tickNow: value => { nowValue = value },
    runTimer: () => timerHandler?.(),
    getWsOptions: () => wsOptions,
    ...monitor
  }
}

describe('useLivePushAudienceMonitor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('connects to the audience websocket and announces the streamer', () => {
    const monitor = createMonitor()

    expect(monitor.connectAudienceWs()).toBe(true)

    expect(monitor.createWs).toHaveBeenCalledWith(expect.objectContaining({
      url: 'ws://example/ws/meeting'
    }))
    monitor.getWsOptions().onOpen()
    expect(monitor.ws.send).toHaveBeenCalledWith({
      type: 'join',
      roomCode: 'stream-1',
      userId: 1,
      userName: '主播（主播）'
    })
    expect(monitor.setTimer).toHaveBeenCalledWith(expect.any(Function), 1000)
  })

  it('tracks room users, joins, leaves, and avoids duplicate viewers', () => {
    const monitor = createMonitor()
    monitor.connectAudienceWs()
    const { onMessage } = monitor.getWsOptions()

    onMessage({
      type: 'room-users',
      data: [
        { userId: 2, userName: 'Alice' },
        { userId: 3 }
      ]
    })
    expect(monitor.viewers.value).toEqual([
      { userId: 2, userName: 'Alice', joinAt: 1000 },
      { userId: 3, userName: '用户3', joinAt: 1000 }
    ])

    monitor.tickNow(2000)
    onMessage({ type: 'user-joined', userId: '2', userName: 'Alice' })
    onMessage({ type: 'user-joined', userId: 4, userName: 'Bob' })
    onMessage({ type: 'user-joined', userId: '1', userName: '主播' })

    expect(monitor.viewers.value).toEqual([
      { userId: 2, userName: 'Alice', joinAt: 1000 },
      { userId: 3, userName: '用户3', joinAt: 1000 },
      { userId: 4, userName: 'Bob', joinAt: 2000 }
    ])

    onMessage({ type: 'user-left', userId: '3' })
    expect(monitor.viewers.value.map(v => v.userId)).toEqual([2, 4])
  })

  it('tracks chat activity, metrics, levels, and leaderboard', () => {
    const monitor = createMonitor()
    monitor.connectAudienceWs()
    const { onMessage } = monitor.getWsOptions()
    onMessage({
      type: 'room-users',
      data: [
        { userId: 2, userName: 'Alice' },
        { userId: 3, userName: 'Bob' }
      ]
    })

    onMessage({ type: 'chat', userId: 2, userName: 'Alice', data: { text: '第一条' } })
    for (let i = 0; i < 11; i++) {
      onMessage({ type: 'chat', userId: 3, userName: 'Bob', data: { text: `刷屏${i}` } })
    }
    onMessage({ type: 'chat', userId: 1, userName: '主播', data: { text: '自己的消息' } })

    monitor.runTimer()

    expect(monitor.recentChats.value).toHaveLength(13)
    expect(monitor.danmakuRate.value).toBe(13)
    expect(monitor.peakViewers.value).toBe(2)
    expect(monitor.totalDanmaku.value).toBe(13)
    expect(monitor.viewerDanmakuCount['2']).toBe(1)
    expect(monitor.viewerDanmakuCount['3']).toBe(11)
    expect(monitor.leaderboard.value).toEqual([
      { userId: 3, userName: 'Bob', count: 11 },
      { userId: 2, userName: 'Alice', count: 1 }
    ])
    expect(monitor.getViewerLevel(3).label).toBe('常客')
  })

  it('does not connect without a stream key and cleans up state', () => {
    const monitor = createMonitor({ streamKey: '' })

    expect(monitor.connectAudienceWs()).toBe(false)
    expect(monitor.createWs).not.toHaveBeenCalled()

    monitor.streamKey.value = 'stream-2'
    monitor.connectAudienceWs()
    monitor.getWsOptions().onMessage({ type: 'room-users', data: [{ userId: 2, userName: 'Alice' }] })
    monitor.getWsOptions().onMessage({ type: 'chat', userId: 2, data: { text: 'hi' } })
    monitor.runTimer()

    monitor.cleanupAudienceMonitor()

    expect(monitor.ws.close).toHaveBeenCalled()
    expect(monitor.clearTimer).toHaveBeenCalledWith('metrics-timer')
    expect(monitor.viewers.value).toEqual([])
    expect(monitor.recentChats.value).toEqual([])
    expect(monitor.danmakuRate.value).toBe(0)
    expect(monitor.peakViewers.value).toBe(0)
    expect(monitor.totalDanmaku.value).toBe(0)
    expect(monitor.viewerDanmakuCount['2']).toBeUndefined()
  })
})
