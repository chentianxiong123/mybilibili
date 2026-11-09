import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { formatScheduledAt, useLivePushRoomControls } from '../composables/useLivePushRoomControls.js'
import { LIVE_ROOM_STATUS } from '../utils/liveMeetingStatus.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn()
  }
}))

const flushAsync = async () => {
  await new Promise(resolve => setTimeout(resolve, 0))
}

const createControls = (options = {}) => {
  const room = options.room || ref({
    id: 1,
    roomName: '原直播间',
    category: '娱乐',
    status: LIVE_ROOM_STATUS.OFFLINE,
    scheduledAt: null
  })
  let nowValue = options.nowValue ?? new Date(2026, 4, 31, 10, 0, 0).getTime()
  let countdownTimerHandler = null
  let reminderTimerHandler = null

  const api = options.api || {
    updateRoom: vi.fn(async () => ({ code: 200 })),
    scheduleRoom: vi.fn(async () => ({ code: 200 })),
    updateRoomStatus: vi.fn(async () => ({ code: 200 }))
  }

  const confirm = options.confirm || vi.fn(async () => true)
  const createNotification = options.createNotification || vi.fn()
  const setIntervalFn = vi.fn(handler => {
    countdownTimerHandler = handler
    return 'countdown-timer'
  })
  const clearIntervalFn = vi.fn()
  const setTimeoutFn = vi.fn(handler => {
    reminderTimerHandler = handler
    return 'reminder-timer'
  })
  const clearTimeoutFn = vi.fn()

  const controls = useLivePushRoomControls({
    room,
    api,
    confirm,
    isNotificationEnabled: options.isNotificationEnabled || (() => true),
    createNotification,
    now: () => nowValue,
    setIntervalFn,
    clearIntervalFn,
    setTimeoutFn,
    clearTimeoutFn,
    nextTickFn: fn => fn()
  })

  return {
    room,
    api,
    confirm,
    createNotification,
    setIntervalFn,
    clearIntervalFn,
    setTimeoutFn,
    clearTimeoutFn,
    getCountdownTimerHandler: () => countdownTimerHandler,
    getReminderTimerHandler: () => reminderTimerHandler,
    tickNow: value => { nowValue = value },
    ...controls
  }
}

describe('useLivePushRoomControls', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('syncs editable room state and enters edit mode', () => {
    const controls = createControls()

    controls.syncFromRoom()

    expect(controls.selectedCategory.value).toBe('娱乐')
    expect(controls.roomNameDraft.value).toBe('原直播间')

    const select = vi.fn()
    controls.editRoomNameInput.value = { select }
    controls.startEditRoomName()

    expect(controls.isEditingRoomName.value).toBe(true)
    expect(select).toHaveBeenCalled()
  })

  it('saves room name and category changes while ignoring blank names', async () => {
    const controls = createControls()

    controls.roomNameDraft.value = '  新直播间  '
    expect(await controls.saveRoomName()).toBe(true)
    expect(controls.api.updateRoom).toHaveBeenCalledWith(1, { roomName: '新直播间' })
    expect(controls.room.value.roomName).toBe('新直播间')
    expect(ElMessage.success).toHaveBeenCalledWith('直播间名称已更新')

    controls.roomNameDraft.value = '   '
    controls.isEditingRoomName.value = true
    expect(await controls.saveRoomName()).toBe(false)
    expect(controls.api.updateRoom).toHaveBeenCalledTimes(1)
    expect(controls.roomNameDraft.value).toBe('新直播间')

    controls.selectedCategory.value = '学习'
    expect(await controls.saveCategory()).toBe(true)
    expect(controls.api.updateRoom).toHaveBeenCalledWith(1, { category: '学习' })
    expect(controls.room.value.category).toBe('学习')
    expect(ElMessage.success).toHaveBeenCalledWith('分类已更新')
  })

  it('formats and pre-fills schedule dialog values from existing and default times', () => {
    const controls = createControls({
      room: ref({
        id: 1,
        roomName: '原直播间',
        category: '娱乐',
        status: LIVE_ROOM_STATUS.OFFLINE,
        scheduledAt: new Date(2026, 4, 31, 8, 5, 0).getTime()
      })
    })

    controls.openScheduleDialog()

    expect(controls.scheduleDate.value).toBe('2026-05-31')
    expect(controls.scheduleTime.value).toBe('08:05')
    expect(controls.showScheduleDialog.value).toBe(true)
    expect(formatScheduledAt(controls.room.value.scheduledAt)).toBe('2026-05-31 08:05')

    controls.room.value.scheduledAt = null
    controls.openScheduleDialog()

    expect(controls.scheduleDate.value).toBe('2026-05-31')
    expect(controls.scheduleTime.value).toBe('11:00')
  })

  it('schedules reminders, fires notifications, and clears timers on cleanup', async () => {
    const controls = createControls({
      createNotification: vi.fn(),
      isNotificationEnabled: () => true
    })

    controls.scheduleReminder(new Date(2026, 4, 31, 10, 10, 0).getTime())

    expect(controls.setTimeoutFn).toHaveBeenCalledWith(expect.any(Function), 300000)

    controls.getReminderTimerHandler()()

    expect(ElMessage.warning).toHaveBeenCalledWith(expect.objectContaining({
      message: '距离开播还有 5 分钟：原直播间',
      duration: 0
    }))
    expect(controls.createNotification).toHaveBeenCalledWith('直播提醒', {
      body: '原直播间 即将开播！'
    })

    controls.goLive()
    await flushAsync()

    expect(controls.confirm).toHaveBeenCalled()
    expect(controls.countdown.value).toBe(2)
    expect(controls.setIntervalFn).toHaveBeenCalledWith(expect.any(Function), 1000)

    controls.scheduleReminder(new Date(2026, 4, 31, 10, 15, 0).getTime())
    controls.cleanupRoomControls()

    expect(controls.clearIntervalFn).toHaveBeenCalledWith('countdown-timer')
    expect(controls.clearTimeoutFn).toHaveBeenCalledWith('reminder-timer')
    expect(controls.countdown.value).toBe(0)
  })

  it('rejects past schedules, saves future schedules, cancels them, and starts or stops live', async () => {
    const controls = createControls({
      confirm: vi.fn(async () => true)
    })

    controls.scheduleDate.value = '2026-05-31'
    controls.scheduleTime.value = '09:00'
    expect(await controls.saveSchedule()).toBe(false)
    expect(ElMessage.warning).toHaveBeenCalledWith('定时时间必须大于当前时间')
    expect(controls.api.scheduleRoom).not.toHaveBeenCalled()

    controls.scheduleDate.value = '2026-05-31'
    controls.scheduleTime.value = '12:00'
    expect(await controls.saveSchedule()).toBe(true)
    expect(controls.api.scheduleRoom).toHaveBeenCalledWith(1, new Date(2026, 4, 31, 12, 0, 0).getTime())
    expect(controls.room.value.scheduledAt).toBe(new Date(2026, 4, 31, 12, 0, 0).getTime())
    expect(controls.showScheduleDialog.value).toBe(false)

    expect(await controls.cancelSchedule()).toBe(true)
    expect(controls.api.scheduleRoom).toHaveBeenLastCalledWith(1, null)
    expect(controls.room.value.scheduledAt).toBeNull()
    expect(controls.clearTimeoutFn).toHaveBeenCalledWith('reminder-timer')

    const livePromise = controls.goLive()
    await flushAsync()
    controls.getCountdownTimerHandler()()
    controls.getCountdownTimerHandler()()
    await livePromise

    expect(controls.api.updateRoomStatus).toHaveBeenCalledWith(1, LIVE_ROOM_STATUS.LIVE)
    expect(controls.room.value.status).toBe(LIVE_ROOM_STATUS.LIVE)
    expect(ElMessage.success).toHaveBeenCalledWith('已开播')

    expect(await controls.stopLive()).toBe(true)
    expect(controls.api.updateRoomStatus).toHaveBeenLastCalledWith(1, LIVE_ROOM_STATUS.OFFLINE)
    expect(controls.room.value.status).toBe(LIVE_ROOM_STATUS.OFFLINE)
    expect(ElMessage.success).toHaveBeenCalledWith('已下播')
  })
})
