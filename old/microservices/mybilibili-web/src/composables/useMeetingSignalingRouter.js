import { ElMessage } from 'element-plus'
import { sameUserId } from '../utils/userId.js'
const PEER_STATE_FIELDS = ['audioEnabled', 'videoEnabled', 'screenShareEnabled', 'handRaised']

const applyPeerState = (peer, state) => {
  for (const field of PEER_STATE_FIELDS) {
    if (state[field] !== undefined) {
      peer[field] = state[field]
    }
  }
}

export function useMeetingSignalingRouter({
  me,
  currentRoom,
  isHost,
  waitingForHost,
  remotePeers,
  audioEnabled,
  spotlightId,
  ensurePeerEntry,
  createPeerConnection,
  broadcastSelfState,
  removePeerConnection,
  handleOffer,
  handleAnswer,
  handleIceCandidate,
  updateAudioActivity,
  toggleAudio,
  enterApprovedRoom,
  cancelWaitingForHost,
  leaveRoom,
  cleanupRoomSession,
  cleanupRoomArtifacts,
  loadMyRooms,
  receiveChatMessage,
  handleJoinRequestMessage,
  handleHandRaisedMessage,
  handleLockRoomMessage,
  handleTransferHostMessage
}) {
  const isDirectSelfTarget = (msg) => sameUserId(msg.targetUserId, me.value.id)

  const isSelfTarget = (msg) => {
    return isDirectSelfTarget(msg) || sameUserId(msg.data?.targetUserId, me.value.id)
  }

  const handlePeerStateMessage = (msg) => {
    if (!msg.data) return false
    if (remotePeers[msg.userId]) {
      if (msg.data.audioEnabled !== undefined) {
        updateAudioActivity(msg.userId, msg.data.audioEnabled)
      }
      if (msg.userName) {
        remotePeers[msg.userId].name = msg.userName
      }
      applyPeerState(remotePeers[msg.userId], msg.data)
      return true
    }
    const nextPeer = {
      name: msg.userName || (`用户${msg.userId}`),
    }
    applyPeerState(nextPeer, msg.data)
    remotePeers[msg.userId] = nextPeer
    return true
  }

  const handleSelfMuteMessage = (msg, message) => {
    if (!isSelfTarget(msg)) {
      return false
    }
    if (audioEnabled.value) toggleAudio()
    ElMessage.info(message)
    return true
  }

  const handleSignaling = async (msg) => {
    switch (msg.type) {
      case 'room-users':
        if (!isHost.value) {
          waitingForHost.value = false
        }
        for (const user of (msg.data || [])) {
          ensurePeerEntry(user.userId, user.userName)
          await createPeerConnection(user.userId, true)
        }
        broadcastSelfState()
        return true
      case 'join-request':
        return handleJoinRequestMessage(msg)
      case 'admit-user':
        if (isDirectSelfTarget(msg)) {
          return enterApprovedRoom()
        }
        return false
      case 'reject-user':
        if (isDirectSelfTarget(msg)) {
          ElMessage.warning('主持人已拒绝你的加入申请')
          cancelWaitingForHost()
          return true
        }
        return false
      case 'user-joined':
        ensurePeerEntry(msg.userId, msg.userName)
        broadcastSelfState()
        return true
      case 'user-left':
        removePeerConnection(msg.userId)
        return true
      case 'offer':
        await handleOffer(msg.userId, msg.data)
        return true
      case 'answer':
        await handleAnswer(msg.userId, msg.data)
        return true
      case 'ice-candidate':
        await handleIceCandidate(msg.userId, msg.data)
        return true
      case 'peer-state':
        return handlePeerStateMessage(msg)
      case 'spotlight':
        if (msg.data?.targetUserId) {
          spotlightId.value = msg.data.active ? msg.data.targetUserId : null
          return true
        }
        return false
      case 'lock-room':
        return handleLockRoomMessage(msg)
      case 'transfer-host':
        return handleTransferHostMessage(msg)
      case 'kick':
        if (isSelfTarget(msg)) {
          ElMessage.warning('你已被主持人移出会议')
          leaveRoom()
          return true
        }
        return false
      case 'mute-target':
        return handleSelfMuteMessage(msg, '主持人已将你静音')
      case 'mute-all':
        if (!sameUserId(msg.userId, me.value.id) && audioEnabled.value) {
          toggleAudio()
          ElMessage.info('主持人已全员静音')
          return true
        }
        return false
      case 'chat':
        receiveChatMessage(msg)
        return true
      case 'hand-raised':
        return handleHandRaisedMessage(msg)
      case 'meeting-ended':
        ElMessage.warning('主持人已结束会议')
        cleanupRoomSession()
        cleanupRoomArtifacts()
        currentRoom.value = null
        await loadMyRooms()
        return true
      default:
        return false
    }
  }

  return {
    handlePeerStateMessage,
    handleSignaling
  }
}
