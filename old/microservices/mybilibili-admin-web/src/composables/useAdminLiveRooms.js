import { ref } from 'vue'
import { getLiveRooms, getLiveStats, updateLiveRoomStatus } from '../api/live'
import {
  getNextLiveStatus,
  isLiveRoomStatus,
  LIVE_ROOM_STATUS
} from '../utils/liveMeetingStatus'
import { normalizePagedResult, runConfirmedAction } from '../utils/adminPage'

const DEFAULT_STATS = Object.freeze({
  totalRooms: 0,
  onlineRooms: 0,
  totalViewers: 0
})

export function useAdminLiveRooms({
  fetchRooms = getLiveRooms,
  fetchStats = getLiveStats,
  updateStatus = updateLiveRoomStatus,
  normalizePage = normalizePagedResult,
  runAction = runConfirmedAction,
  logger = console
} = {}) {
  const loading = ref(false)
  const rooms = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const statusFilter = ref('')
  const stats = ref({ ...DEFAULT_STATS })

  const statusOptions = [
    { label: '全部', value: '' },
    { label: '直播中', value: LIVE_ROOM_STATUS.LIVE },
    { label: '离线', value: LIVE_ROOM_STATUS.OFFLINE }
  ]

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

  const loadStats = async () => {
    try {
      const res = await fetchStats()
      if (res.code === 200) {
        stats.value = { ...DEFAULT_STATS, ...(res.data || {}) }
        return true
      }
      return false
    } catch (e) {
      logger?.error?.(e)
      return false
    }
  }

  const handlePageChange = async (nextPage) => {
    page.value = nextPage
    return loadRooms()
  }

  const handleStatusFilterChange = async () => {
    page.value = 1
    return loadRooms()
  }

  const toggleRoomStatus = async (row) => {
    const newStatus = getNextLiveStatus(row.status)
    const action = isLiveRoomStatus(newStatus) ? '开播' : '下播'
    return runAction({
      message: `确定${action}「${row.roomName}」？`,
      action: () => updateStatus(row.id, newStatus),
      successMessage: `${action}成功`,
      onSuccess: async () => {
        await loadRooms()
        await loadStats()
      }
    })
  }

  const initialize = async () => {
    await loadRooms()
    await loadStats()
  }

  return {
    loading,
    rooms,
    total,
    page,
    pageSize,
    statusFilter,
    stats,
    statusOptions,
    loadRooms,
    loadStats,
    handlePageChange,
    handleStatusFilterChange,
    toggleRoomStatus,
    initialize
  }
}
