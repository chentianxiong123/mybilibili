import { ElMessage } from 'element-plus'
import { sameUserId } from '../utils/userId.js'

export function useMeetingHostControls({
  me,
  isHost,
  remotePeers,
  waitingUsers,
  roomLocked,
  spotlightId,
  hostUserId,
  sendMeetingMessage,
  removePeerConnection
}) {
  const lowerHand = (userId) => {
    if (!isHost.value) return false
    if (!sendMeetingMessage({ type: 'hand-raised', targetUserId: userId, data: { raised: false } })) return false
    if (remotePeers[userId]) remotePeers[userId].handRaised = false
    return true
  }

  const approveHandRaise = (userId) => {
    if (!lowerHand(userId)) return false
    ElMessage.success('已批准发言')
    return true
  }

  const admitUser = (userId) => {
    if (!isHost.value) return false
    if (!sendMeetingMessage({ type: 'admit-user', targetUserId: userId })) return false
    waitingUsers.value = waitingUsers.value.filter(w => !sameUserId(w.userId, userId))
    return true
  }

  const rejectUser = (userId) => {
    if (!isHost.value) return false
    if (!sendMeetingMessage({ type: 'reject-user', targetUserId: userId })) return false
    waitingUsers.value = waitingUsers.value.filter(w => !sameUserId(w.userId, userId))
    return true
  }

  const spotlightUser = (userId) => {
    if (!isHost.value) return false
    const nextSpotlight = sameUserId(spotlightId.value, userId) ? null : userId
    if (!sendMeetingMessage({
      type: 'spotlight',
      data: { targetUserId: userId, active: sameUserId(nextSpotlight, userId) }
    })) return false
    spotlightId.value = nextSpotlight
    return true
  }

  const toggleLockRoom = () => {
    if (!isHost.value) return false
    const nextLocked = !roomLocked.value
    if (!sendMeetingMessage({ type: 'lock-room', data: { locked: nextLocked } })) return false
    roomLocked.value = nextLocked
    ElMessage.success(roomLocked.value ? '会议已锁定' : '会议已解锁')
    return true
  }

  const transferHost = (userId) => {
    if (!isHost.value) return false
    if (!sendMeetingMessage({ type: 'transfer-host', targetUserId: userId })) return false
    hostUserId.value = userId
    ElMessage.success('已将主持人转让')
    return true
  }

  const kickUser = (userId) => {
    if (!isHost.value) return false
    if (!sendMeetingMessage({ type: 'kick', targetUserId: userId, data: { targetUserId: userId } })) return false
    removePeerConnection(userId)
    return true
  }

  const muteUser = (userId) => {
    if (!isHost.value) return false
    return sendMeetingMessage({ type: 'mute-target', targetUserId: userId, data: { targetUserId: userId } })
  }

  const muteAll = () => {
    if (!isHost.value) return false
    if (!sendMeetingMessage({ type: 'mute-all', data: {} })) return false
    ElMessage.success('已请求全员静音')
    return true
  }

  const handleJoinRequestMessage = (msg) => {
    if (!isHost.value) return false
    if (!msg?.userId) return false
    if (!waitingUsers.value.find(w => sameUserId(w.userId, msg.userId))) {
      waitingUsers.value.push({ userId: msg.userId, userName: msg.userName, ts: Date.now() })
      ElMessage.info(`${msg.userName || ('用户' + msg.userId)} 请求加入会议`)
    }
    return true
  }

  const handleHandRaisedMessage = (msg) => {
    if (!msg?.userId || !msg.data) return false
    if (remotePeers[msg.userId]) {
      remotePeers[msg.userId].handRaised = !!msg.data.raised
    }
    if (isHost.value && msg.data.raised) {
      ElMessage.info(`${msg.userName || ('用户' + msg.userId)} 举手了`)
    }
    return true
  }

  const handleLockRoomMessage = (msg) => {
    if (isHost.value) return false
    roomLocked.value = msg?.data?.locked === true
    if (roomLocked.value) {
      ElMessage.warning('会议已被主持人锁定，请等待')
    } else {
      ElMessage.success('会议已解锁')
    }
    return true
  }

  const handleTransferHostMessage = (msg) => {
    if (!msg?.targetUserId) return false
    hostUserId.value = msg.targetUserId
    if (sameUserId(msg.targetUserId, me.value.id)) {
      ElMessage.success('你已成为新的主持人')
    }
    return true
  }

  return {
    approveHandRaise,
    lowerHand,
    admitUser,
    rejectUser,
    spotlightUser,
    toggleLockRoom,
    transferHost,
    kickUser,
    muteUser,
    muteAll,
    handleJoinRequestMessage,
    handleHandRaisedMessage,
    handleLockRoomMessage,
    handleTransferHostMessage
  }
}
