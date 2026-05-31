import { nextTick, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { liveApi } from '../api/live.js'
import { LIVE_ROOM_STATUS } from '../utils/liveMeetingStatus.js'
import { isNotificationEnabled as defaultNotificationEnabled } from '../utils/notification.js'

const DEFAULT_CATEGORIES = ['娱乐', '游戏', '学习', '购物', '赛事', '其他']

const pad = n => String(n).padStart(2, '0')

const toDateFields = (value) => {
  const d = new Date(value)
  return {
    date: `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`,
    time: `${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
}

const toDateField = value => value instanceof Date ? toDateFields(value).date : String(value)

const toTimeField = value => value instanceof Date ? toDateFields(value).time : String(value)

export function formatScheduledAt(ts) {
  if (!ts) return ''
  const fields = toDateFields(ts)
  return `${fields.date} ${fields.time}`
}

export function useLivePushRoomControls({
  room,
  api = liveApi,
  confirm = (...args) => ElMessageBox.confirm(...args),
  isNotificationEnabled = defaultNotificationEnabled,
  createNotification = (title, options) => new Notification(title, options),
  now = () => Date.now(),
  setIntervalFn = (handler, delay) => setInterval(handler, delay),
  clearIntervalFn = timerId => clearInterval(timerId),
  setTimeoutFn = (handler, delay) => setTimeout(handler, delay),
  clearTimeoutFn = timerId => clearTimeout(timerId),
  nextTickFn = nextTick
}) {
  const switching = ref(false)
  const countdown = ref(0)
  let countdownTimer = null
  let scheduleNotifyTimer = null

  const isEditingRoomName = ref(false)
  const roomNameDraft = ref('')
  const editRoomNameInput = ref(null)
  const selectedCategory = ref('')
  const categories = DEFAULT_CATEGORIES

  const showScheduleDialog = ref(false)
  const scheduleDate = ref('')
  const scheduleTime = ref('')

  const clearCountdownTimer = () => {
    if (!countdownTimer) return
    clearIntervalFn(countdownTimer)
    countdownTimer = null
  }

  const clearScheduleReminder = () => {
    if (!scheduleNotifyTimer) return
    clearTimeoutFn(scheduleNotifyTimer)
    scheduleNotifyTimer = null
  }

  const syncFromRoom = () => {
    selectedCategory.value = room.value?.category || ''
    roomNameDraft.value = room.value?.roomName || ''
  }

  const startEditRoomName = () => {
    roomNameDraft.value = room.value?.roomName || ''
    isEditingRoomName.value = true
    nextTickFn(() => editRoomNameInput.value?.select())
  }

  const saveRoomName = async () => {
    if (!room.value) return false
    const newName = roomNameDraft.value.trim()
    if (!newName || newName === room.value.roomName) {
      roomNameDraft.value = room.value.roomName || ''
      isEditingRoomName.value = false
      return false
    }
    try {
      const res = await api.updateRoom(room.value.id, { roomName: newName })
      if (res.code === 200) {
        room.value.roomName = newName
        roomNameDraft.value = newName
        ElMessage.success('直播间名称已更新')
        isEditingRoomName.value = false
        return true
      }
      ElMessage.error(res.message || '更新失败')
      return false
    } catch (e) {
      ElMessage.error('更新失败')
      return false
    } finally {
      isEditingRoomName.value = false
    }
  }

  const saveCategory = async () => {
    if (!room.value) return false
    try {
      const res = await api.updateRoom(room.value.id, { category: selectedCategory.value })
      if (res.code === 200) {
        room.value.category = selectedCategory.value
        ElMessage.success('分类已更新')
        return true
      }
      ElMessage.error(res.message || '更新失败')
      return false
    } catch (e) {
      ElMessage.error('更新失败')
      return false
    }
  }

  const openScheduleDialog = () => {
    const source = room.value?.scheduledAt || now() + 3600000
    const fields = toDateFields(source)
    scheduleDate.value = fields.date
    scheduleTime.value = fields.time
    showScheduleDialog.value = true
  }

  const saveSchedule = async () => {
    if (!room.value || !scheduleDate.value || !scheduleTime.value) return false
    const dt = new Date(`${toDateField(scheduleDate.value)}T${toTimeField(scheduleTime.value)}`)
    if (Number.isNaN(dt.getTime())) {
      ElMessage.warning('请选择有效的定时时间')
      return false
    }
    if (dt <= new Date(now())) {
      ElMessage.warning('定时时间必须大于当前时间')
      return false
    }
    try {
      const scheduledAt = dt.getTime()
      const res = await api.scheduleRoom(room.value.id, scheduledAt)
      if (res.code === 200) {
        room.value.scheduledAt = scheduledAt
        ElMessage.success('定时开播已设置')
        scheduleReminder(scheduledAt)
        showScheduleDialog.value = false
        return true
      }
      ElMessage.error(res.message || '设置失败')
      return false
    } catch (e) {
      ElMessage.error('设置失败')
      return false
    } finally {
      showScheduleDialog.value = false
    }
  }

  const cancelSchedule = async () => {
    if (!room.value) return false
    try {
      const res = await api.scheduleRoom(room.value.id, null)
      if (res.code === 200) {
        room.value.scheduledAt = null
        clearScheduleReminder()
        ElMessage.success('已取消定时开播')
        showScheduleDialog.value = false
        return true
      }
      ElMessage.error(res.message || '取消失败')
      return false
    } catch (e) {
      ElMessage.error('取消失败')
      return false
    } finally {
      showScheduleDialog.value = false
    }
  }

  const scheduleReminder = (ts) => {
    clearScheduleReminder()
    const ms = typeof ts === 'number' ? ts : new Date(ts).getTime()
    const delay = ms - 5 * 60000 - now()
    if (delay <= 0) return false
    scheduleNotifyTimer = setTimeoutFn(() => {
      ElMessage.warning({
        message: `距离开播还有 5 分钟：${room.value?.roomName || '我的直播间'}`,
        duration: 0
      })
      if (isNotificationEnabled()) {
        createNotification('直播提醒', { body: `${room.value?.roomName || '直播间'} 即将开播！` })
      }
    }, delay)
    return true
  }

  const doStartLive = async () => {
    if (!room.value) return false
    switching.value = true
    try {
      const res = await api.updateRoomStatus(room.value.id, LIVE_ROOM_STATUS.LIVE)
      if (res.code === 200) {
        room.value.status = LIVE_ROOM_STATUS.LIVE
        ElMessage.success('已开播')
        return true
      }
      ElMessage.error(res.message || '开播失败')
      return false
    } catch (e) {
      ElMessage.error('开播失败')
      return false
    } finally {
      switching.value = false
    }
  }

  const goLive = async () => {
    if (!room.value) return false
    countdown.value = 3
    const confirmStart = await confirm(
      `<div style="text-align:center;font-size:48px;margin-bottom:8px;">${countdown.value}</div><div style="font-size:14px;color:#9499a0;">即将开始直播...</div>`,
      '确定要开播吗？',
      {
        confirmButtonText: '立即开播',
        cancelButtonText: '取消',
        type: 'warning',
        showCancelButton: true,
        dangerouslyUseHTMLString: true,
        closeOnClickModal: false
      }
    ).catch(() => false)
    if (!confirmStart) {
      countdown.value = 0
      return false
    }

    return new Promise(resolve => {
      const tick = () => {
        countdown.value--
        if (countdown.value <= 0) {
          clearCountdownTimer()
          resolve(doStartLive())
        }
      }
      clearCountdownTimer()
      countdownTimer = setIntervalFn(tick, 1000)
      tick()
    })
  }

  const stopLive = async () => {
    if (!room.value) return false
    try {
      await confirm('确定要结束本场直播吗？观众将无法继续观看。', '结束直播', {
        confirmButtonText: '结束',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch (e) {
      return false
    }
    switching.value = true
    try {
      const res = await api.updateRoomStatus(room.value.id, LIVE_ROOM_STATUS.OFFLINE)
      if (res.code === 200) {
        room.value.status = LIVE_ROOM_STATUS.OFFLINE
        ElMessage.success('已下播')
        return true
      }
      ElMessage.error(res.message || '下播失败')
      return false
    } catch (e) {
      ElMessage.error('下播失败')
      return false
    } finally {
      switching.value = false
    }
  }

  const cleanupRoomControls = () => {
    clearCountdownTimer()
    clearScheduleReminder()
    countdown.value = 0
  }

  return {
    switching,
    countdown,
    isEditingRoomName,
    roomNameDraft,
    editRoomNameInput,
    selectedCategory,
    categories,
    showScheduleDialog,
    scheduleDate,
    scheduleTime,
    syncFromRoom,
    startEditRoomName,
    saveRoomName,
    saveCategory,
    openScheduleDialog,
    saveSchedule,
    cancelSchedule,
    formatScheduledAt,
    scheduleReminder,
    goLive,
    doStartLive,
    stopLive,
    cleanupRoomControls
  }
}
