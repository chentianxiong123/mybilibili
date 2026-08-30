import { nextTick, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { liveApi } from '../api/live.js'
import { getCurrentUser } from '../utils/currentUser.js'
import { requestNotificationPermission as defaultRequestNotificationPermission } from '../utils/notification.js'
import { copyTextToClipboard } from '../utils/clipboard.js'

const noop = () => {}
const noopAsync = async () => {}

const defaultGetShareUrl = () => (typeof window !== 'undefined' ? window.location.href : '')

export function useLiveRoomSession({
  roomId,
  api = liveApi,
  message = ElMessage,
  getCurrentUserFn = getCurrentUser,
  requestNotificationPermission = defaultRequestNotificationPermission,
  copyText = copyTextToClipboard,
  getShareUrl = defaultGetShareUrl,
  nextTickFn = nextTick,
  logger = console
} = {}) {
  const room = ref(null)
  const loading = ref(true)
  const me = ref(getCurrentUserFn())
  const currentUserId = ref(me.value.id)
  const isStreamer = ref(false)

  const loadRoom = async ({
    syncFollowState = noop,
    initPlayer = noop
  } = {}) => {
    try {
      const res = await api.getRoom(roomId)
      if (res.code === 200) {
        room.value = res.data
        isStreamer.value = currentUserId.value === room.value?.userId
        syncFollowState(room.value)
        return room.value
      }
      message.error(res.message || '直播间不存在')
      return null
    } catch (e) {
      logger?.error?.('[LiveRoom] 加载失败:', e)
      message.error('加载失败')
      return null
    } finally {
      loading.value = false
      await nextTickFn()
      if (room.value) initPlayer()
    }
  }

  const initializeLiveRoom = async ({
    syncFollowState = noop,
    initPlayer = noop,
    loadLinkmicStatus = noopAsync,
    connectLiveRoomSocket = noop
  } = {}) => {
    requestNotificationPermission()
    await loadRoom({ syncFollowState, initPlayer })
    await loadLinkmicStatus()
    connectLiveRoomSocket()
    return room.value
  }

  const shareRoom = () => {
    return copyText(getShareUrl(), {
      successMessage: '直播间链接已复制，可分享给好友！',
      message
    })
  }

  return {
    room,
    loading,
    me,
    currentUserId,
    isStreamer,
    loadRoom,
    initializeLiveRoom,
    shareRoom
  }
}
