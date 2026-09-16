import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useLivePushAudienceMonitor } from './useLivePushAudienceMonitor'

function makeFakeWs() {
  return {
    isOpen: () => true,
    send: vi.fn(() => true),
    close: vi.fn(),
    _handler: {} as any,
  }
}

describe('useLivePushAudienceMonitor', () => {
  let ws: ReturnType<typeof makeFakeWs>
  let createWs: any

  beforeEach(() => {
    ws = makeFakeWs()
    createWs = vi.fn((opts: any) => {
      ws._handler = opts
      return ws
    })
  })

  function makeHarness(overrides: any = {}) {
    const streamKey = ref('skey')
    const me = { id: 99, name: 'host' }
    const r = useLivePushAudienceMonitor({
      streamKey,
      me,
      createWs,
      getWsUrl: () => 'ws://test/ws/live',
      now: () => 1000000,
      random: () => 0.1,
      setTimer: vi.fn(() => 'timer-id' as any) as any,
      clearTimer: vi.fn(),
      ...overrides,
    })
    return { ...r, streamKey, me }
  }

  describe('connectAudienceWs', () => {
    it('streamKey 为空时返回 false 不创建 ws', () => {
      const r = makeHarness({ streamKey: ref('') })
      const ok = r.connectAudienceWs()
      expect(ok).toBe(false)
      expect(createWs).not.toHaveBeenCalled()
    })

    it('正常连接后调用 createWs 并设 metricsTimer', () => {
      const r = makeHarness()
      const ok = r.connectAudienceWs()
      expect(ok).toBe(true)
      expect(createWs).toHaveBeenCalledTimes(1)
    })

    it('重复连接时关闭旧的 ws', () => {
      const r = makeHarness()
      r.connectAudienceWs()
      const oldWs = ws
      r.connectAudienceWs()
      expect(oldWs.close).toHaveBeenCalled()
    })

    it('onOpen 时 send join 消息含 userName（主播）', () => {
      const r = makeHarness()
      r.connectAudienceWs()
      ws._handler.onOpen()
      expect(ws.send).toHaveBeenCalledWith(expect.objectContaining({
        type: 'join',
        roomCode: 'skey',
        userId: 99,
        userName: 'host（主播）',
      }))
    })

    it('metricsTimer 在连接时 setTimer 被调用', () => {
      const setTimer = vi.fn(() => 'tid')
      const r = makeHarness({ setTimer })
      r.connectAudienceWs()
      expect(setTimer).toHaveBeenCalled()
    })

    it('metricsTimer 第二次连接时不会重复创建', () => {
      const setTimer = vi.fn(() => 'tid')
      const r = makeHarness({ setTimer })
      r.connectAudienceWs()
      r.connectAudienceWs()
      // setTimer 调用次数：第一次调用 1 次，第二次因为 metricsTimer 已存在不会触发 setTimer
      // 但 connectAudienceWs 内部先 close 旧 ws
      expect(setTimer.mock.calls.length).toBe(1)
    })
  })

  describe('消息分发', () => {
    let r: any
    beforeEach(() => {
      r = makeHarness()
      r.connectAudienceWs()
    })

    it('room-users 替换 viewers 列表', () => {
      ws._handler.onMessage({
        type: 'room-users',
        data: [{ userId: 1, userName: 'alice' }, { userId: 2 }]
      })
      expect(r.viewers.value).toHaveLength(2)
      expect(r.viewers.value[0].userName).toBe('alice')
      expect(r.viewers.value[1].userName).toBe('用户2') // 默认 userName 拼接
    })

    it('user-joined 时新增 viewer，但过滤掉自己', () => {
      ws._handler.onMessage({ type: 'user-joined', userId: 1, userName: 'bob' })
      expect(r.viewers.value).toHaveLength(1)
      ws._handler.onMessage({ type: 'user-joined', userId: 99, userName: 'host' })
      expect(r.viewers.value).toHaveLength(1) // 99 是自己
    })

    it('user-joined 重复 userId 不重复添加', () => {
      ws._handler.onMessage({ type: 'user-joined', userId: 1, userName: 'bob' })
      ws._handler.onMessage({ type: 'user-joined', userId: 1, userName: 'bob2' })
      expect(r.viewers.value).toHaveLength(1)
    })

    it('user-left 移除 viewer', () => {
      ws._handler.onMessage({ type: 'user-joined', userId: 1, userName: 'bob' })
      ws._handler.onMessage({ type: 'user-left', userId: 1 })
      expect(r.viewers.value).toHaveLength(0)
    })

    it('chat 消息增加 recentChats 和 viewerDanmakuCount', () => {
      ws._handler.onMessage({
        type: 'chat',
        userId: 5,
        userName: 'alice',
        data: { text: 'hi' }
      })
      expect(r.recentChats.value).toHaveLength(1)
      expect(r.viewerDanmakuCount['5']).toBe(1)
    })

    it('chat 消息超过 80 时 splice(0, 30)', () => {
      for (let i = 0; i < 90; i++) {
        ws._handler.onMessage({
          type: 'chat',
          userId: i,
          userName: `u${i}`,
          data: { text: 'm' + i }
        })
      }
      expect(r.recentChats.value.length).toBeLessThanOrEqual(80)
    })

    it('chat 无 text 时不入 recentChats', () => {
      ws._handler.onMessage({ type: 'chat', userId: 5, data: {} })
      expect(r.recentChats.value).toHaveLength(0)
    })
  })

  describe('leaderboard / getViewerLevel', () => {
    it('leaderboard 按 count 降序，取前 5，过滤掉自己', () => {
      const r = makeHarness({ me: { id: 1, name: 'host' } })
      r.connectAudienceWs()
      // 5 个非自己观众各自弹幕
      for (let i = 2; i <= 6; i++) {
        for (let j = 0; j < i; j++) {
          r.viewerDanmakuCount[String(i)] = (r.viewerDanmakuCount[String(i)] || 0) + 1
        }
        r.viewers.value.push({ userId: i, userName: `u${i}` })
      }
      // user 1 is host (should be filtered)
      r.viewerDanmakuCount['1'] = 999
      r.viewers.value.push({ userId: 1, userName: 'host' })
      expect(r.leaderboard.value.length).toBeLessThanOrEqual(5)
      expect(r.leaderboard.value[0].userId).toBeGreaterThan(1)
    })

    it('viewerLevels 是预定义等级', () => {
      const r = makeHarness()
      expect(Array.isArray(r.viewerLevels)).toBe(true)
      expect(r.viewerLevels.length).toBeGreaterThan(0)
    })

    it('getViewerLevel 返回正确等级', () => {
      const r = makeHarness()
      r.viewerDanmakuCount['1'] = 5
      const lv = r.getViewerLevel(1)
      expect(lv.label).toBe('新人')
      r.viewerDanmakuCount['2'] = 20
      expect(r.getViewerLevel(2).label).toBe('常客')
      r.viewerDanmakuCount['3'] = 100
      expect(r.getViewerLevel(3).label).toBe('活跃')
      r.viewerDanmakuCount['4'] = 300
      expect(r.getViewerLevel(4).label).toBe('忠实')
      r.viewerDanmakuCount['5'] = 600
      expect(r.getViewerLevel(5).label).toBe('贵族')
    })

    it('getViewerLevel 对无数据的 userId 返回新人', () => {
      const r = makeHarness()
      expect(r.getViewerLevel(999).label).toBe('新人')
    })
  })

  describe('updateMetrics', () => {
    it('60 秒内的弹幕计入 danmakuRate', () => {
      const r = makeHarness({ now: () => 1000000 })
      r.connectAudienceWs()
      // 时间 t = 999500
      vi.spyOn(Date, 'now').mockReturnValue(999500)
      ws._handler.onMessage({ type: 'chat', userId: 5, data: { text: 'a' } })
      r.updateMetrics()
      expect(r.danmakuRate.value).toBe(1)
    })

    it('peakViewers 在 viewers 增长时更新', () => {
      const r = makeHarness()
      r.connectAudienceWs()
      ws._handler.onMessage({ type: 'user-joined', userId: 1 })
      ws._handler.onMessage({ type: 'user-joined', userId: 2 })
      r.updateMetrics()
      expect(r.peakViewers.value).toBe(2)
    })

    it('totalDanmaku 等于 recentChats.length', () => {
      const r = makeHarness()
      r.connectAudienceWs()
      ws._handler.onMessage({ type: 'chat', userId: 5, data: { text: 'a' } })
      ws._handler.onMessage({ type: 'chat', userId: 6, data: { text: 'b' } })
      r.updateMetrics()
      expect(r.totalDanmaku.value).toBe(2)
    })
  })

  describe('cleanupAudienceMonitor', () => {
    it('关闭 ws、清理 viewers、danmakuRate、peakViewers、totalDanmaku', () => {
      const clearTimer = vi.fn()
      const r = makeHarness({ clearTimer })
      r.connectAudienceWs()
      ws._handler.onMessage({ type: 'user-joined', userId: 1, userName: 'a' })
      ws._handler.onMessage({ type: 'chat', userId: 5, data: { text: 'a' } })
      r.peakViewers.value = 5
      r.updateMetrics()
      r.cleanupAudienceMonitor()
      expect(ws.close).toHaveBeenCalled()
      expect(r.viewers.value).toEqual([])
      expect(r.recentChats.value).toEqual([])
      expect(r.danmakuRate.value).toBe(0)
      expect(r.peakViewers.value).toBe(0)
      expect(r.totalDanmaku.value).toBe(0)
      expect(clearTimer).toHaveBeenCalled()
    })
  })
})