import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { liveApi } from '../api/live.js'

const DEFAULT_CATEGORIES = ['全部', '娱乐', '游戏', '学习', '购物', '赛事', '其他']

export function useLiveRoomList({
  api = liveApi,
  message = ElMessage,
  pollDelay = 10000,
  setIntervalFn = (handler, delay) => setInterval(handler, delay),
  clearIntervalFn = timerId => clearInterval(timerId),
  logger = console
} = {}) {
  const liveRooms = ref([])
  const myRoom = ref(null)
  const loading = ref(false)
  const keyword = ref('')
  const sortBy = ref('hot')
  const categories = DEFAULT_CATEGORIES
  const selectedCategory = ref('全部')
  let pollTimer = null

  const filteredRooms = computed(() => {
    let list = liveRooms.value

    if (myRoom.value) {
      list = list.filter(room => room.userId !== myRoom.value.userId)
    }

    if (selectedCategory.value && selectedCategory.value !== '全部') {
      list = list.filter(room => room.category === selectedCategory.value)
    }

    const kw = keyword.value.trim().toLowerCase()
    if (kw) {
      list = list.filter(room =>
        (room.roomName || '').toLowerCase().includes(kw) ||
        String(room.userId).includes(kw)
      )
    }

    list = [...list]
    if (sortBy.value === 'hot') {
      list.sort((a, b) => (b.viewerCount || 0) - (a.viewerCount || 0))
    } else {
      list.sort((a, b) => (b.id || 0) - (a.id || 0))
    }
    return list
  })

  const totalLive = computed(() => liveRooms.value.length)

  const loadData = async (silent = false) => {
    if (!silent) loading.value = true
    try {
      const [listRes, myRoomRes] = await Promise.allSettled([
        api.getLiveList(),
        api.getMyRoom()
      ])

      if (listRes.status === 'fulfilled' && listRes.value.code === 200) {
        liveRooms.value = listRes.value.data || []
      }

      if (myRoomRes.status === 'fulfilled' && myRoomRes.value.code === 200 && myRoomRes.value.data) {
        myRoom.value = myRoomRes.value.data
      } else {
        myRoom.value = null
      }
    } catch (e) {
      logger?.error?.('获取直播数据失败:', e)
    } finally {
      if (!silent) loading.value = false
    }
  }

  const refreshData = () => {
    const request = loadData()
    message.success('已刷新')
    return request
  }

  const stopPolling = () => {
    if (!pollTimer) return false
    clearIntervalFn(pollTimer)
    pollTimer = null
    return true
  }

  const startPolling = () => {
    stopPolling()
    pollTimer = setIntervalFn(() => loadData(true), pollDelay)
    return pollTimer
  }

  const cleanupLiveRoomList = () => {
    stopPolling()
  }

  return {
    liveRooms,
    myRoom,
    loading,
    keyword,
    sortBy,
    categories,
    selectedCategory,
    filteredRooms,
    totalLive,
    loadData,
    refreshData,
    startPolling,
    stopPolling,
    cleanupLiveRoomList
  }
}
