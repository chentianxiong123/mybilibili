import { copyTextToClipboard } from '../utils/clipboard.js'

const noop = () => {}
const noopAsync = async () => false

const defaultGetOrigin = () => (typeof location !== 'undefined' ? location.origin : '')

export function useMeetingSessionLifecycle({
  currentRoom,
  hostUserId,
  cleanupRecording = noop,
  cleanupMeetingSessionUi = noop,
  closeAllPeerConnections = noop,
  cleanupMedia = noop,
  cleanupPresenceSignals = noop,
  cleanupRoomSession = noop,
  leaveMeetingSession = noopAsync,
  copyText = copyTextToClipboard,
  getOrigin = defaultGetOrigin
} = {}) {
  const cleanupRoomArtifacts = () => {
    cleanupRecording()
    cleanupMeetingSessionUi()
    closeAllPeerConnections()
    cleanupMedia()
    cleanupPresenceSignals()
    if (hostUserId) hostUserId.value = null
    return true
  }

  const leaveRoom = async () => {
    const left = await leaveMeetingSession()
    if (left) cleanupRoomArtifacts()
    return left === true
  }

  const cleanup = () => {
    cleanupRoomArtifacts()
    cleanupRoomSession()
    return true
  }

  const shareRoom = () => {
    const roomCode = currentRoom.value?.roomCode
    if (!roomCode) return false
    return copyText(`${getOrigin()}/meeting?code=${roomCode}`, {
      successMessage: '邀请链接已复制，可直接发送给朋友加入会议'
    })
  }

  const copyRoomCode = (code) => {
    return copyText(code, {
      successMessage: '邀请码已复制'
    })
  }

  return {
    cleanupRoomArtifacts,
    leaveRoom,
    cleanup,
    shareRoom,
    copyRoomCode
  }
}
