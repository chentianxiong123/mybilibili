import { ref } from 'vue'
import { createReconnectingWS as defaultCreateReconnectingWS } from '../utils/reconnectingWs.js'

const defaultWsUrl = () => {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/ws/meeting`
}

export function useLiveRoomRealtime({
  room,
  me,
  currentUserId,
  createWs = defaultCreateReconnectingWS,
  getWsUrl = defaultWsUrl,
  logger = console
}) {
  const realtimeViewerCount = ref(0)
  const handlers = {}
  let ws = null

  const attachRoomMessageHandlers = (nextHandlers = {}) => {
    Object.assign(handlers, nextHandlers)
    return handlers
  }

  const sendRoomMessage = ({ type, targetUserId, data }) => {
    if (!ws || !ws.isOpen()) return false
    return ws.send({
      type,
      roomCode: room.value?.streamKey,
      userId: currentUserId.value,
      userName: me.value.name,
      targetUserId,
      data
    })
  }

  const callHandler = (name, ...args) => {
    if (typeof handlers[name] !== 'function') return false
    handlers[name](...args)
    return true
  }

  const callAsyncHandler = async (name, ...args) => {
    if (typeof handlers[name] !== 'function') return false
    await handlers[name](...args)
    return true
  }

  const handleRoomMessage = async (msg) => {
    switch (msg.type) {
      case 'linkmic-apply':
        return callHandler('handleLinkmicApplyMessage', msg)
      case 'linkmic-accepted':
        return callAsyncHandler('handleLinkmicAcceptedMessage', msg)
      case 'linkmic-rejected':
        return callHandler('handleLinkmicRejectedMessage', msg)
      case 'linkmic-disconnected':
        return callHandler('handleLinkmicDisconnectedMessage', msg)
      case 'hand-raised':
        return callHandler('handleHandRaisedMessage', msg)
      case 'gift':
        return callHandler('receiveGift', msg)
      case 'offer':
        return callAsyncHandler('handleOffer', msg.userId, msg.data)
      case 'answer':
        return callAsyncHandler('handleAnswer', msg.userId, msg.data)
      case 'ice-candidate':
        return callAsyncHandler('handleIceCandidate', msg.userId, msg.data)
      case 'reaction':
        return callHandler('receiveReaction', msg)
      case 'pin-message':
        return callHandler('receivePinMessage', msg)
      case 'unpin-message':
        return callHandler('clearPinnedMessage')
      case 'chat':
        return callHandler('receiveDanmaku', msg)
      case 'mute-target':
        return callAsyncHandler('handleMuteAudioMessage', msg)
      case 'mute-video-target':
        return callAsyncHandler('handleMuteVideoMessage', msg)
      case 'peer-state':
        return callHandler('handlePeerStateMessage', msg)
      case 'viewer-count':
        if (msg.data?.count === undefined) return false
        realtimeViewerCount.value = msg.data.count
        return true
      default:
        return false
    }
  }

  const cleanupLiveRoomRealtime = () => {
    if (!ws) return false
    try { ws.close() } catch (e) {}
    ws = null
    return true
  }

  const connectLiveRoomSocket = () => {
    cleanupLiveRoomRealtime()
    ws = createWs({
      url: getWsUrl,
      onOpen: () => {
        if (room.value?.streamKey && currentUserId.value) {
          sendRoomMessage({ type: 'join' })
        }
      },
      onMessage: handleRoomMessage,
      onClose: () => logger?.log?.('[LiveRoom WS] 断开（将自动重连）'),
      logger
    })
    return true
  }

  return {
    realtimeViewerCount,
    attachRoomMessageHandlers,
    sendRoomMessage,
    handleRoomMessage,
    connectLiveRoomSocket,
    cleanupLiveRoomRealtime
  }
}
