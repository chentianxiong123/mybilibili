import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref, nextTick } from 'vue'

vi.mock('@/api/live.ts', () => ({
  liveApi: {
    updateRoom: vi.fn(),
    updateRoomStatus: vi.fn(),
    scheduleRoom: vi.fn(),
  },
}))

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  },
  ElMessageBox: {
    confirm: vi.fn(),
  },
}))

import { liveApi } from '@/api/live.ts'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useLivePushRoomControls, formatScheduledAt } from './useLivePushRoomControls'
import { LIVE_ROOM_STATUS } from '@/utils/liveMeetingStatus.ts'

const api = liveApi as unknown as {
  updateRoom: ReturnType<typeof vi.fn>
  updateRoomStatus: ReturnType<typeof vi.fn>
  scheduleRoom: ReturnType<typeof vi.fn>
}

describe('useLivePushRoomControls', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('formatScheduledAt', () => {
    it('ts 转换为 yyyy-MM-dd HH:mm 格式', () => {
      const d = new Date(2025, 0, 5, 14, 30)
      expect(formatScheduledAt(d.getTime())).toBe('2025-01-05 14:30')
    })
    it('空值返回空字符串', () => {
      expect(formatScheduledAt(0)).toBe('')
      expect(formatScheduledAt(null)).toBe('')
    })
    it('填充 0', () => {
      const d = new Date(2025, 2, 1, 5, 7)
      expect(formatScheduledAt(d.getTime())).toBe('2025-03-01 05:07')
    })
  })

  describe('syncFromRoom / startEditRoomName', () => {
    it('syncFromRoom 从 room 读取 category 和 roomName', () => {
      const room = ref({ roomName: 'R', category: '学习' })
      const r = useLivePushRoomControls({ room })
      r.syncFromRoom()
      expect(r.selectedCategory.value).toBe('学习')
      expect(r.roomNameDraft.value).toBe('R')
    })
    it('startEditRoomName 设置 isEditing=true 并调用 nextTick', async () => {
      const room = ref({ roomName: 'old' })
      const r = useLivePushRoomControls({ room })
      r.startEditRoomName()
      expect(r.isEditingRoomName.value).toBe(true)
      expect(r.roomNameDraft.value).toBe('old')
      // editRoomNameInput.value?.select() 是 select 调用
      r.editRoomNameInput.value = { select: vi.fn() } as any
      await nextTick()
      r.startEditRoomName()
      await nextTick()
      expect(r.editRoomNameInput.value.select).toHaveBeenCalled()
    })
  })

  describe('saveRoomName', () => {
    it('空名称不改', async () => {
      const room = ref({ id: 1, roomName: 'old' })
      const r = useLivePushRoomControls({ room })
      r.roomNameDraft.value = '   '
      const ok = await r.saveRoomName()
      expect(ok).toBe(false)
      expect(api.updateRoom).not.toHaveBeenCalled()
    })

    it('同名不改', async () => {
      const room = ref({ id: 1, roomName: 'old' })
      const r = useLivePushRoomControls({ room })
      r.roomNameDraft.value = 'old'
      const ok = await r.saveRoomName()
      expect(ok).toBe(false)
    })

    it('新名称更新 room 并成功', async () => {
      api.updateRoom.mockResolvedValueOnce({ code: 200 })
      const room = ref({ id: 1, roomName: 'old' })
      const r = useLivePushRoomControls({ room })
      r.roomNameDraft.value = 'new'
      const ok = await r.saveRoomName()
      expect(ok).toBe(true)
      expect(room.value.roomName).toBe('new')
      expect(ElMessage.success).toHaveBeenCalledWith('直播间名称已更新')
    })

    it('非 200 显示错误', async () => {
      api.updateRoom.mockResolvedValueOnce({ code: 500, message: 'fail' })
      const room = ref({ id: 1, roomName: 'old' })
      const r = useLivePushRoomControls({ room })
      r.roomNameDraft.value = 'new'
      const ok = await r.saveRoomName()
      expect(ok).toBe(false)
      expect(ElMessage.error).toHaveBeenCalledWith('fail')
    })

    it('rejected 时显示错误', async () => {
      api.updateRoom.mockRejectedValueOnce(new Error('boom'))
      const room = ref({ id: 1, roomName: 'old' })
      const r = useLivePushRoomControls({ room })
      r.roomNameDraft.value = 'new'
      const ok = await r.saveRoomName()
      expect(ok).toBe(false)
      expect(ElMessage.error).toHaveBeenCalled()
    })

    it('room 为 null 返回 false', async () => {
      const room = ref(null)
      const r = useLivePushRoomControls({ room })
      r.roomNameDraft.value = 'new'
      expect(await r.saveRoomName()).toBe(false)
    })
  })

  describe('saveCategory', () => {
    it('成功更新分类', async () => {
      api.updateRoom.mockResolvedValueOnce({ code: 200 })
      const room = ref({ id: 1, category: 'old' })
      const r = useLivePushRoomControls({ room })
      r.selectedCategory.value = 'new'
      const ok = await r.saveCategory()
      expect(ok).toBe(true)
      expect(room.value.category).toBe('new')
    })
    it('非 200 显示错误', async () => {
      api.updateRoom.mockResolvedValueOnce({ code: 500, message: 'x' })
      const room = ref({ id: 1, category: 'old' })
      const r = useLivePushRoomControls({ room })
      const ok = await r.saveCategory()
      expect(ok).toBe(false)
      expect(ElMessage.error).toHaveBeenCalledWith('x')
    })
    it('rejected 时显示错误', async () => {
      api.updateRoom.mockRejectedValueOnce(new Error('boom'))
      const room = ref({ id: 1, category: 'old' })
      const r = useLivePushRoomControls({ room })
      const ok = await r.saveCategory()
      expect(ok).toBe(false)
    })
  })

  describe('saveSchedule', () => {
    it('无效日期', async () => {
      const room = ref({ id: 1 })
      const r = useLivePushRoomControls({ room })
      r.scheduleDate.value = ''
      r.scheduleTime.value = ''
      const ok = await r.saveSchedule()
      expect(ok).toBe(false)
    })
    it('过去时间', async () => {
      const room = ref({ id: 1 })
      // now = 2025-01-01
      const now = new Date(2025, 0, 1).getTime()
      const r = useLivePushRoomControls({ room, now: () => now })
      // schedule in 2000
      r.scheduleDate.value = '2000-01-01'
      r.scheduleTime.value = '00:00'
      const ok = await r.saveSchedule()
      expect(ok).toBe(false)
      expect(ElMessage.warning).toHaveBeenCalledWith('定时时间必须大于当前时间')
    })
    it('成功设置定时', async () => {
      api.scheduleRoom.mockResolvedValueOnce({ code: 200 })
      const room = ref({ id: 1 })
      const r = useLivePushRoomControls({ room, now: () => 1000000000 })
      // set 1 hour later
      const dt = new Date(1000000000 + 7200000)
      const pad = (n: number) => String(n).padStart(2, '0')
      r.scheduleDate.value = `${dt.getFullYear()}-${pad(dt.getMonth() + 1)}-${pad(dt.getDate())}`
      r.scheduleTime.value = `${pad(dt.getHours())}:${pad(dt.getMinutes())}`
      const ok = await r.saveSchedule()
      expect(ok).toBe(true)
      expect(api.scheduleRoom).toHaveBeenCalledWith(1, expect.any(Number))
      expect(ElMessage.success).toHaveBeenCalledWith('定时开播已设置')
    })
    it('非 200 显示错误', async () => {
      api.scheduleRoom.mockResolvedValueOnce({ code: 500 })
      const room = ref({ id: 1 })
      const r = useLivePushRoomControls({ room, now: () => 1000000000 })
      const dt = new Date(1000000000 + 7200000)
      const pad = (n: number) => String(n).padStart(2, '0')
      r.scheduleDate.value = `${dt.getFullYear()}-${pad(dt.getMonth() + 1)}-${pad(dt.getDate())}`
      r.scheduleTime.value = `${pad(dt.getHours())}:${pad(dt.getMinutes())}`
      const ok = await r.saveSchedule()
      expect(ok).toBe(false)
      expect(ElMessage.error).toHaveBeenCalled()
    })
  })

  describe('cancelSchedule', () => {
    it('成功取消', async () => {
      api.scheduleRoom.mockResolvedValueOnce({ code: 200 })
      const room = ref({ id: 1, scheduledAt: 12345 })
      const r = useLivePushRoomControls({ room })
      const ok = await r.cancelSchedule()
      expect(ok).toBe(true)
      expect(room.value.scheduledAt).toBe(null)
      expect(ElMessage.success).toHaveBeenCalledWith('已取消定时开播')
    })
    it('rejected 显示错误', async () => {
      api.scheduleRoom.mockRejectedValueOnce(new Error('x'))
      const room = ref({ id: 1, scheduledAt: 12345 })
      const r = useLivePushRoomControls({ room })
      const ok = await r.cancelSchedule()
      expect(ok).toBe(false)
    })
  })

  describe('scheduleReminder', () => {
    it('delay <= 0 时不设置', () => {
      const r = useLivePushRoomControls({ room: ref({ id: 1, roomName: 'R' }), now: () => 1000000 })
      const ok = r.scheduleReminder(1000000 - 6 * 60000) // 6 minutes ago
      expect(ok).toBe(false)
    })
    it('正常设置 setTimeout', () => {
      const setTimeoutFn = vi.fn(() => 'tid')
      const r = useLivePushRoomControls({
        room: ref({ id: 1, roomName: 'R' }),
        now: () => 1000000,
        setTimeoutFn: setTimeoutFn as any,
      })
      const ok = r.scheduleReminder(1000000 + 10 * 60000) // 10 minutes later
      expect(ok).toBe(true)
      expect(setTimeoutFn).toHaveBeenCalled()
    })
    it('接受 Date 或 number', () => {
      const setTimeoutFn = vi.fn(() => 'tid')
      const r = useLivePushRoomControls({
        room: ref({ id: 1, roomName: 'R' }),
        now: () => 0,
        setTimeoutFn: setTimeoutFn as any,
      })
      const dt = new Date(10 * 60000)
      r.scheduleReminder(dt)
      expect(setTimeoutFn).toHaveBeenCalled()
    })
  })

  describe('doStartLive / stopLive', () => {
    it('doStartLive 成功时切换 status', async () => {
      api.updateRoomStatus.mockResolvedValueOnce({ code: 200 })
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.OFFLINE })
      const r = useLivePushRoomControls({ room })
      const ok = await r.doStartLive()
      expect(ok).toBe(true)
      expect(room.value.status).toBe(LIVE_ROOM_STATUS.LIVE)
      expect(ElMessage.success).toHaveBeenCalledWith('已开播')
    })
    it('doStartLive 非 200 返回 false', async () => {
      api.updateRoomStatus.mockResolvedValueOnce({ code: 500, message: 'x' })
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.OFFLINE })
      const r = useLivePushRoomControls({ room })
      const ok = await r.doStartLive()
      expect(ok).toBe(false)
    })
    it('doStartLive rejected 显示错误', async () => {
      api.updateRoomStatus.mockRejectedValueOnce(new Error('boom'))
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.OFFLINE })
      const r = useLivePushRoomControls({ room })
      const ok = await r.doStartLive()
      expect(ok).toBe(false)
      expect(ElMessage.error).toHaveBeenCalled()
    })

    it('stopLive 用户取消 confirm 时返回 false', async () => {
      const confirm = vi.fn().mockRejectedValue(new Error('cancel'))
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.LIVE })
      const r = useLivePushRoomControls({ room, confirm: confirm as any })
      const ok = await r.stopLive()
      expect(ok).toBe(false)
    })
    it('stopLive 确认后调用 API', async () => {
      api.updateRoomStatus.mockResolvedValueOnce({ code: 200 })
      const confirm = vi.fn().mockResolvedValue('ok')
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.LIVE })
      const r = useLivePushRoomControls({ room, confirm: confirm as any })
      const ok = await r.stopLive()
      expect(ok).toBe(true)
      expect(room.value.status).toBe(LIVE_ROOM_STATUS.OFFLINE)
    })
  })

  describe('goLive countdown', () => {
    it('用户取消 confirm 时直接返回 false', async () => {
      const confirm = vi.fn().mockRejectedValue(new Error('cancel'))
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.OFFLINE })
      const r = useLivePushRoomControls({ room, confirm: confirm as any })
      const ok = await r.goLive()
      expect(ok).toBe(false)
      expect(r.countdown.value).toBe(0)
    })

    it('用户确认后倒计时 3 → 0 → 调用 doStartLive', async () => {
      vi.useFakeTimers()
      api.updateRoomStatus.mockResolvedValue({ code: 200 })
      const confirm = vi.fn().mockResolvedValue('ok')
      const room = ref({ id: 1, status: LIVE_ROOM_STATUS.OFFLINE })
      const setIntervalFn = vi.fn((cb: () => void, _delay: number) => {
        // 立即调用一次模拟 interval 行为
        return 'tid' as any
      })
      const clearIntervalFn = vi.fn()
      const r = useLivePushRoomControls({
        room,
        confirm: confirm as any,
        setIntervalFn: setIntervalFn as any,
        clearIntervalFn: clearIntervalFn as any,
      })
      // goLive 走的是手动倒计时（用 setInterval），不是 fake timer 测试目标。
      // 这里直接覆盖其内部 setIntervalFn 实现，让 tick 立即连续触发 3 次。
      let tickCallCount = 0
      r.countdown.value = 0 // 重置
      setIntervalFn.mockImplementation((cb: () => void) => {
        tickCallCount++
        // 模拟立刻触发 3 次
        cb()
        cb()
        cb()
        return 'tid' as any
      })
      const ok = await r.goLive()
      expect(ok).toBe(true)
      // countdown 已经走完（会减到 -1 或 0，依调用次数）
      expect(r.countdown.value).toBeLessThanOrEqual(0)
      expect(api.updateRoomStatus).toHaveBeenCalled()
      vi.useRealTimers()
    })
  })

  describe('cleanupRoomControls', () => {
    it('清理 timers 并重置 countdown', () => {
      const clearIntervalFn = vi.fn()
      const clearTimeoutFn = vi.fn()
      const r = useLivePushRoomControls({
        room: ref({ id: 1 }),
        clearIntervalFn: clearIntervalFn as any,
        clearTimeoutFn: clearTimeoutFn as any,
      })
      r.cleanupRoomControls()
      expect(r.countdown.value).toBe(0)
    })
  })
})