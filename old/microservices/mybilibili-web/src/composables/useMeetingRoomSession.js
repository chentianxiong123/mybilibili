import { ElMessage, ElMessageBox } from 'element-plus'
import { meetingApi as defaultMeetingApi } from '../api/meeting.js'
import { createReconnectingWS as defaultCreateReconnectingWS } from '../utils/reconnectingWs.js'
import { getMeetingJoinBlockMessage as defaultJoinBlockMessage } from '../utils/liveMeetingStatus.js'

const defaultWsUrl = () => {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/ws/meeting`
}

export function useMeetingRoomSession({
  me,
  joinCode,
  currentRoom,
  hostUserId,
  loading,
  waitingForHost,
  startLocalStream,
  restartMeetingTimer,
  startQualityPolling,
  loadMyRooms,
  onMessage,
  api = defaultMeetingApi,
  createWs = defaultCreateReconnectingWS,
  getWsUrl = defaultWsUrl,
  getJoinBlockMessage = defaultJoinBlockMessage,
  confirm = (...args) => ElMessageBox.confirm(...args),
  logger = console
}) {
  let ws = null

  const normalizeRoomCode = (code) => {
    const roomCode = code || joinCode.value.trim()
    return roomCode ? roomCode.trim() : ''
  }

  const closeSocket = () => {
    if (!ws) return false
    const socket = ws
    ws = null
    try {
      socket.close()
      return true
    } catch (e) {
      return false
    }
  }

  const sendSocketMessage = (payload) => {
    if (!ws || !ws.isOpen()) return false
    try {
      ws.send(payload)
      return true
    } catch (e) {
      return false
    }
  }

  const openMeetingSocket = (roomCode, messageType) => {
    closeSocket()
    ws = createWs({
      url: getWsUrl,
      onOpen: () => {
        sendSocketMessage({
          type: messageType,
          roomCode,
          userId: me.value.id,
          userName: me.value.name
        })
      },
      onMessage
    })
    return true
  }

  const connectWaitingForHost = (roomCode) => openMeetingSocket(roomCode, 'join-request')

  const connectRoom = (roomCode) => openMeetingSocket(roomCode, 'join')

  const sendMeetingMessage = ({ type, targetUserId, data }) => {
    return sendSocketMessage({
      type,
      roomCode: currentRoom.value?.roomCode,
      userId: me.value.id,
      userName: me.value.name,
      targetUserId,
      data
    })
  }

  const sendSignaling = (type, targetUserId, data) => {
    return sendMeetingMessage({ type, targetUserId, data })
  }

  const resetRoomState = () => {
    waitingForHost.value = false
    currentRoom.value = null
    hostUserId.value = null
  }

  const joinRoomByCode = async (code) => {
    const roomCode = normalizeRoomCode(code)
    if (!roomCode) {
      ElMessage.warning('请输入会议室邀请码')
      return false
    }
    loading.value = true
    try {
      const res = await api.getRoom(roomCode)
      if (res.code !== 200) {
        ElMessage.error(res.message || '会议室不存在')
        return false
      }
      const roomData = res.data
      const isRoomHost = me.value.id === roomData.creatorId
      const blockMessage = getJoinBlockMessage(roomData.status)
      if (blockMessage) {
        ElMessage.warning(blockMessage)
        return false
      }

      currentRoom.value = roomData
      hostUserId.value = roomData.creatorId

      if (!isRoomHost) {
        waitingForHost.value = true
        connectWaitingForHost(roomCode)
        ElMessage.info('正在等待主持人批准...')
        return true
      }

      const joinRes = await api.joinRoom(roomCode)
      if (joinRes.code !== 200) {
        ElMessage.error(joinRes.message || '加入失败')
        return false
      }

      await startLocalStream()
      connectRoom(roomCode)
      restartMeetingTimer()
      startQualityPolling()
      ElMessage.success('加入成功')
      return true
    } catch (e) {
      logger.error?.(e)
      ElMessage.error('加入失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const enterApprovedRoom = async () => {
    if (!currentRoom.value?.roomCode) {
      cancelWaitingForHost()
      return false
    }
    try {
      const joinRes = await api.joinRoom(currentRoom.value.roomCode)
      if (joinRes.code !== 200) {
        ElMessage.error(joinRes.message || '加入失败')
        cancelWaitingForHost()
        return false
      }
      waitingForHost.value = false
      await startLocalStream()
      connectRoom(currentRoom.value.roomCode)
      restartMeetingTimer()
      startQualityPolling()
      ElMessage.success('主持人已允许加入')
      return true
    } catch (e) {
      ElMessage.error('加入失败')
      cancelWaitingForHost()
      return false
    }
  }

  const cancelWaitingForHost = () => {
    cleanupRoomSession({ notifyLeave: false })
    resetRoomState()
    return true
  }

  const leaveRoom = async () => {
    if (!currentRoom.value) return false
    const roomId = currentRoom.value.id
    const roomCode = currentRoom.value.roomCode
    if (hostUserId.value && me.value.id === hostUserId.value) {
      let endMeeting = false
      try {
        await confirm('结束会议会让所有人退出，是否结束？否则你只是自己退出。', '主持人离开', {
          confirmButtonText: '结束会议',
          cancelButtonText: '仅自己离开',
          type: 'warning'
        })
        endMeeting = true
      } catch (e) {
        endMeeting = false
      }
      if (endMeeting) {
        const ended = await endRoomOnServer(roomId)
        if (ended) {
          sendSocketMessage({
            type: 'meeting-ended',
            roomCode,
            userId: me.value.id,
            data: {}
          })
        }
      } else {
        await leaveRoomOnServer(roomId)
      }
    } else {
      await leaveRoomOnServer(roomId)
    }

    cleanupRoomSession()
    resetRoomState()
    await loadMyRooms()
    return true
  }

  const cleanupRoomSession = ({ notifyLeave = true } = {}) => {
    if (notifyLeave) {
      sendSocketMessage({
        type: 'leave',
        roomCode: currentRoom.value?.roomCode,
        userId: me.value.id
      })
    }
    closeSocket()
    waitingForHost.value = false
    return true
  }

  const leaveRoomOnServer = async (roomId) => {
    try {
      const res = await api.leaveRoom(roomId)
      return res?.code === 200
    } catch (e) {
      return false
    }
  }

  const endRoomOnServer = async (roomId) => {
    try {
      const res = await api.endRoom(roomId)
      if (res?.code === 200) {
        return true
      }
      ElMessage.error(res?.message || '结束会议失败')
      return false
    } catch (e) {
      ElMessage.error('结束会议失败')
      return false
    }
  }

  return {
    connectWaitingForHost,
    connectRoom,
    sendMeetingMessage,
    sendSignaling,
    joinRoomByCode,
    enterApprovedRoom,
    cancelWaitingForHost,
    leaveRoom,
    cleanupRoomSession
  }
}
