import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { liveApi } from '../api/live.js'
import { getCurrentUser } from '../utils/currentUser.js'
import { copyTextToClipboard } from '../utils/clipboard.js'
import { LIVE_ROOM_STATUS, isLiveRoomStatus } from '../utils/liveMeetingStatus.js'

const DEFAULT_ROOM_NAME = '我的直播间'

const defaultGetHostname = () => (typeof window !== 'undefined' ? window.location.hostname : 'localhost')

export function useLivePushRoomSetup({
  api = liveApi,
  message = ElMessage,
  getHostname = defaultGetHostname,
  getCurrentUserFn = getCurrentUser,
  copyText = copyTextToClipboard,
  now = () => Date.now(),
  logger = console
} = {}) {
  const room = ref(null)
  const loading = ref(true)
  const rtmpUrl = ref('')
  const streamKey = ref('')
  const me = getCurrentUserFn({ fallbackName: '主播' })

  const isLive = computed(() => isLiveRoomStatus(room.value?.status))

  const syncRoomState = (roomData) => {
    room.value = roomData || null
    streamKey.value = roomData?.streamKey || ''
    rtmpUrl.value = roomData ? `rtmp://${getHostname()}/live` : ''
    return room.value
  }

  const loadRoom = async () => {
    loading.value = true
    try {
      let res = await api.getMyRoom()
      if (!res || !res.data) {
        res = await api.createRoom(DEFAULT_ROOM_NAME)
      }
      if (res && res.code === 200 && res.data) {
        syncRoomState(res.data)
        return room.value
      }
      syncRoomState(null)
      return null
    } catch (e) {
      logger?.error?.('获取直播间信息失败:', e)
      message.error('获取直播间信息失败')
      syncRoomState(null)
      return null
    } finally {
      loading.value = false
    }
  }

  const handleWebRtcPublished = async () => {
    if (!room.value) return false
    if (!isLiveRoomStatus(room.value.status)) {
      try {
        const res = await api.updateRoomStatus(room.value.id, LIVE_ROOM_STATUS.LIVE)
        if (res.code === 200) {
          room.value.status = LIVE_ROOM_STATUS.LIVE
          return true
        }
      } catch (e) {
        logger?.error?.('更新直播间状态失败:', e)
      }
    }
    room.value.status = LIVE_ROOM_STATUS.LIVE
    return true
  }

  const formatJoinTime = (ts) => {
    const seconds = Math.floor((now() - ts) / 1000)
    if (seconds < 60) return `${seconds} 秒前`
    if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
    return `${Math.floor(seconds / 3600)} 小时前`
  }

  const copyField = (text) => {
    return copyText(text, {
      successMessage: '已复制',
      message
    })
  }

  const beforeCoverUpload = (file) => {
    const isImg = file.type?.startsWith('image/')
    const isLt2M = file.size / 1024 / 1024 < 2
    if (!isImg) {
      message.error('只能上传图片文件')
      return false
    }
    if (!isLt2M) {
      message.error('图片大小不能超过 2MB')
      return false
    }
    return true
  }

  const uploadCover = async ({ file }) => {
    if (!room.value) return false
    try {
      const res = await api.uploadCover(room.value.id, file)
      if (res.code === 200 && res.data) {
        room.value.coverUrl = res.data
        message.success('封面上传成功')
        return true
      }
      message.error(res.message || '上传失败')
      return false
    } catch (e) {
      logger?.error?.('上传封面失败:', e)
      message.error('封面上传失败')
      return false
    }
  }

  return {
    room,
    loading,
    rtmpUrl,
    streamKey,
    me,
    isLive,
    loadRoom,
    syncRoomState,
    handleWebRtcPublished,
    formatJoinTime,
    copyField,
    beforeCoverUpload,
    uploadCover
  }
}
