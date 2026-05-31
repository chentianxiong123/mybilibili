import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { meetingApi } from '../api/meeting.js'

export function useMeetingRoomCatalog({
  api = meetingApi,
  onRoomCreated = async () => {},
  logger = console
} = {}) {
  const loading = ref(false)
  const myRooms = ref([])
  const newRoomName = ref('')
  const reserveLoading = ref(false)
  const reserveForm = reactive({
    roomName: '',
    timeRange: [],
    reason: ''
  })

  const loadMyRooms = async () => {
    try {
      const res = await api.getMyRooms()
      if (res.code === 200) {
        myRooms.value = res.data || []
        return true
      }
      return false
    } catch (e) {
      logger.error?.('获取会议室失败:', e)
      return false
    }
  }

  const createRoom = async () => {
    const roomName = newRoomName.value.trim()
    if (!roomName) {
      ElMessage.warning('请输入会议室名称')
      return false
    }

    loading.value = true
    try {
      const res = await api.createRoom(roomName)
      if (res.code !== 200) {
        ElMessage.error(res.message || '创建失败')
        return false
      }
      ElMessage.success('创建成功')
      newRoomName.value = ''
      await loadMyRooms()
      await onRoomCreated(res.data)
      return true
    } catch (e) {
      ElMessage.error('创建失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const resetReserveForm = () => {
    reserveForm.roomName = ''
    reserveForm.timeRange = []
    reserveForm.reason = ''
  }

  const reserveRoom = async () => {
    const roomName = reserveForm.roomName.trim()
    if (!roomName) {
      ElMessage.warning('请输入会议室名称')
      return false
    }
    if (!Array.isArray(reserveForm.timeRange) || reserveForm.timeRange.length !== 2) {
      ElMessage.warning('请选择预约时间段')
      return false
    }

    reserveLoading.value = true
    try {
      const res = await api.reserveRoom({
        roomName,
        scheduledStart: reserveForm.timeRange[0],
        scheduledEnd: reserveForm.timeRange[1],
        reason: reserveForm.reason.trim()
      })
      if (res.code !== 200) {
        ElMessage.error(res.message || '预约失败')
        return false
      }
      ElMessage.success('预约已提交')
      resetReserveForm()
      await loadMyRooms()
      return true
    } catch (e) {
      ElMessage.error('预约失败')
      return false
    } finally {
      reserveLoading.value = false
    }
  }

  return {
    loading,
    myRooms,
    newRoomName,
    reserveLoading,
    reserveForm,
    loadMyRooms,
    createRoom,
    reserveRoom
  }
}
