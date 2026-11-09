import { ref } from 'vue'
import {
  approveMeetingReservation,
  forceEndMeeting,
  getMeetingRooms,
  getPendingMeetingReservations,
  rejectMeetingReservation
} from '../api/meeting'
import { normalizePagedResult, runConfirmedAction } from '../utils/adminPage'

export function useAdminMeetings({
  fetchRooms = getMeetingRooms,
  fetchPending = getPendingMeetingReservations,
  approveReservation = approveMeetingReservation,
  rejectReservation = rejectMeetingReservation,
  endMeeting = forceEndMeeting,
  normalizePage = normalizePagedResult,
  runAction = runConfirmedAction,
  logger = console
} = {}) {
  const loading = ref(false)
  const rooms = ref([])
  const pendingReservations = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const activeTab = ref('list')
  const statusFilter = ref(null)
  const detailDialog = ref(false)
  const currentRoom = ref(null)

  const loadRooms = async () => {
    loading.value = true
    try {
      const res = await fetchRooms(page.value, pageSize.value, statusFilter.value)
      if (res.code === 200) {
        const pageResult = normalizePage(res.data)
        rooms.value = pageResult.records
        total.value = pageResult.total
        return true
      }
      return false
    } catch (e) {
      logger?.error?.(e)
      return false
    } finally {
      loading.value = false
    }
  }

  const loadPending = async () => {
    try {
      const res = await fetchPending()
      pendingReservations.value = res.data || []
      return res.code === 200
    } catch (e) {
      logger?.error?.(e)
      return false
    }
  }

  const refreshRooms = async () => {
    page.value = 1
    return loadRooms()
  }

  const refreshAfterDecision = async () => {
    await loadPending()
    await loadRooms()
  }

  const runReservationDecision = (row, label, successMessage, requestAction) => {
    return runAction({
      message: `确认${label}「${row.roomName}」的预约？`,
      title: '审批',
      action: () => requestAction(row.id),
      successMessage,
      onSuccess: refreshAfterDecision
    })
  }

  const approve = (row) => runReservationDecision(row, '通过', '已通过', approveReservation)

  const reject = (row) => runReservationDecision(row, '拒绝', '已拒绝', rejectReservation)

  const forceEnd = (row) => {
    return runAction({
      message: `确认强制结束「${row.roomName}」？`,
      title: '警告',
      action: () => endMeeting(row.id),
      successMessage: '已强制结束',
      onSuccess: loadRooms
    })
  }

  const viewDetail = (row) => {
    currentRoom.value = row
    detailDialog.value = true
  }

  const initialize = async () => {
    await loadRooms()
    await loadPending()
  }

  return {
    loading,
    rooms,
    pendingReservations,
    total,
    page,
    pageSize,
    activeTab,
    statusFilter,
    detailDialog,
    currentRoom,
    loadRooms,
    loadPending,
    refreshRooms,
    approve,
    reject,
    forceEnd,
    viewDetail,
    initialize
  }
}
