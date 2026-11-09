import { nextTick, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { linkmicApi } from '../api/linkmic.js'
import { usePeerConnectionMesh } from './usePeerConnectionMesh.js'
import { sameUserId } from '../utils/userId.js'
const LINKMIC_PEER_STATE_FIELDS = ['audioEnabled', 'videoEnabled']

const applyLinkmicPeerState = (peer, state = {}) => {
  for (const field of LINKMIC_PEER_STATE_FIELDS) {
    if (state[field] !== undefined) {
      peer[field] = state[field]
    }
  }
}

export function useLiveLinkmic({
  roomId,
  room,
  isStreamer,
  currentUserId,
  sendRoomMessage,
  logger = console
}) {
  const myLinkmicId = ref(null)
  const myQueuePosition = ref(0)
  const activeLinkmics = ref([])
  const pendingApplications = ref([])
  const linkmicLoading = ref(false)
  const linkmicStreams = reactive({})
  const localLinkmicStream = ref(null)
  const localLinkmicVideoRef = ref(null)
  const localLinkmicAudioEnabled = ref(true)
  const localLinkmicVideoEnabled = ref(true)
  const viewerHandRaised = ref(false)

  const ensureLinkmicEntry = (peerId, name) => {
    if (!linkmicStreams[peerId]) {
      linkmicStreams[peerId] = {
        name: name || ('用户' + peerId),
        stream: null
      }
    } else if (name && !linkmicStreams[peerId].name) {
      linkmicStreams[peerId].name = name
    }
  }

  const sendSignaling = (type, targetUserId, data) => {
    return sendRoomMessage({ type, targetUserId, data })
  }

  const sendBroadcast = (type, data) => {
    return sendRoomMessage({ type, data })
  }

  const isCurrentUserTarget = (msg) => {
    return sameUserId(msg.targetUserId, currentUserId.value) || sameUserId(msg.data?.targetUserId, currentUserId.value)
  }

  const {
    peerConnections,
    createPeerConnection,
    handleOffer,
    handleAnswer,
    handleIceCandidate,
    removePeerConnection,
    closeAllPeerConnections
  } = usePeerConnectionMesh({
    localStream: localLinkmicStream,
    remotePeers: linkmicStreams,
    ensurePeerEntry: ensureLinkmicEntry,
    sendSignal: sendSignaling,
    canRestartIce: () => !!localLinkmicStream.value,
    resetRemoteOnOffer: true,
    logPrefix: '[Linkmic]'
  })

  const stopLocalLinkmicStream = () => {
    if (!localLinkmicStream.value) return
    localLinkmicStream.value.getTracks().forEach(track => track.stop())
    localLinkmicStream.value = null
    if (localLinkmicVideoRef.value) {
      localLinkmicVideoRef.value.srcObject = null
    }
  }

  const removeLinkmicStream = (viewerId) => {
    if (!viewerId) return
    removePeerConnection(viewerId)
  }

  const loadLinkmicStatus = async () => {
    try {
      const res = await linkmicApi.getActiveLinkmics(roomId)
      if (res.code === 200) {
        activeLinkmics.value = res.data || []
      }
      if (isStreamer.value) {
        const pendingRes = await linkmicApi.getPendingApplications(roomId)
        if (pendingRes.code === 200) {
          pendingApplications.value = pendingRes.data || []
        }
      }
    } catch (e) {
      logger.error?.('加载连麦状态失败:', e)
    }
  }

  const viewerStartConnection = async (streamerId) => {
    if (!streamerId) return
    try {
      if (!localLinkmicStream.value) {
        localLinkmicStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
        await nextTick()
        if (localLinkmicVideoRef.value) {
          localLinkmicVideoRef.value.srcObject = localLinkmicStream.value
        }
      }
      if (peerConnections[streamerId]) {
        removePeerConnection(streamerId)
      }
      await createPeerConnection(streamerId, true)
    } catch (e) {
      logger.error?.('[Linkmic] viewer start failed', e)
      ElMessage.error('无法获取摄像头/麦克风')
    }
  }

  const cleanupLinkmic = () => {
    closeAllPeerConnections()
    stopLocalLinkmicStream()
  }

  const setRemoteVideo = (peerId, el) => {
    if (!el) return
    const peer = linkmicStreams[peerId]
    if (peer && peer.stream && el.srcObject !== peer.stream) {
      el.srcObject = peer.stream
    }
  }

  const applyLinkmic = async () => {
    if (!currentUserId.value) {
      ElMessage.warning('请先登录')
      return
    }
    linkmicLoading.value = true
    try {
      const res = await linkmicApi.applyLinkmic(roomId)
      if (res.code === 200) {
        myLinkmicId.value = res.data.id
        try {
          const qres = await linkmicApi.getQueuePosition(roomId)
          if (qres.code === 200) myQueuePosition.value = qres.data || 0
        } catch (e) {}
        if (myQueuePosition.value > 0) {
          ElMessage.success('连麦申请已发送，当前排队第 ' + myQueuePosition.value + ' 位')
        } else {
          ElMessage.success('连麦申请已发送，请等待主播同意')
        }
        sendBroadcast('linkmic-apply', res.data)
      } else {
        ElMessage.error(res.message || '申请失败')
      }
    } catch (e) {
      ElMessage.error('申请连麦失败')
    } finally {
      linkmicLoading.value = false
    }
  }

  const acceptLinkmic = async (linkmic) => {
    try {
      await linkmicApi.acceptLinkmic(linkmic.id)
      linkmic.status = 1
      if (!activeLinkmics.value.find(item => item.id === linkmic.id)) {
        activeLinkmics.value.push(linkmic)
      }
      pendingApplications.value = pendingApplications.value.filter(item => item.id !== linkmic.id)
      sendSignaling('linkmic-accepted', linkmic.viewerId, { linkmicId: linkmic.id })
      ElMessage.success('已同意连麦')
    } catch (e) {
      ElMessage.error('操作失败')
    }
  }

  const rejectLinkmic = async (linkmic) => {
    try {
      await linkmicApi.rejectLinkmic(linkmic.id)
      pendingApplications.value = pendingApplications.value.filter(item => item.id !== linkmic.id)
      sendSignaling('linkmic-rejected', linkmic.viewerId, {})
    } catch (e) {
      ElMessage.error('操作失败')
    }
  }

  const disconnectLinkmic = async (linkmic) => {
    try {
      await linkmicApi.disconnectLinkmic(linkmic.id)
      removeLinkmicStream(linkmic.viewerId)
      activeLinkmics.value = activeLinkmics.value.filter(item => item.id !== linkmic.id)
      sendSignaling('linkmic-disconnected', linkmic.viewerId, { viewerId: linkmic.viewerId })
    } catch (e) {
      ElMessage.error('操作失败')
    }
  }

  const viewerHangup = async () => {
    if (!myLinkmicId.value) return
    try {
      await linkmicApi.disconnectLinkmic(myLinkmicId.value)
    } catch (e) {}
    if (room.value?.userId) {
      sendSignaling('linkmic-disconnected', room.value.userId, { viewerId: currentUserId.value })
      removePeerConnection(room.value.userId)
    }
    myLinkmicId.value = null
    myQueuePosition.value = 0
    viewerHandRaised.value = false
    stopLocalLinkmicStream()
    ElMessage.info('已断开连麦')
  }

  const toggleViewerHandRaise = () => {
    const raised = !viewerHandRaised.value
    if (!sendSignaling('hand-raised', room.value?.userId, { raised })) return
    viewerHandRaised.value = raised
    ElMessage.info(viewerHandRaised.value ? '已举手，请等待主播接受' : '已放下手')
  }

  const toggleLocalAudio = async () => {
    if (!localLinkmicStream.value) return
    const track = localLinkmicStream.value.getAudioTracks()[0]
    if (!track) return
    track.enabled = !track.enabled
    localLinkmicAudioEnabled.value = track.enabled
    sendBroadcast('peer-state', {
      audioEnabled: localLinkmicAudioEnabled.value,
      videoEnabled: localLinkmicVideoEnabled.value
    })
    if (myLinkmicId.value) {
      try { await linkmicApi.toggleAudio(myLinkmicId.value, track.enabled) } catch (e) {}
    }
  }

  const toggleLocalVideo = async () => {
    if (!localLinkmicStream.value) return
    const track = localLinkmicStream.value.getVideoTracks()[0]
    if (!track) return
    track.enabled = !track.enabled
    localLinkmicVideoEnabled.value = track.enabled
    sendBroadcast('peer-state', {
      audioEnabled: localLinkmicAudioEnabled.value,
      videoEnabled: localLinkmicVideoEnabled.value
    })
    if (myLinkmicId.value) {
      try { await linkmicApi.toggleVideo(myLinkmicId.value, track.enabled) } catch (e) {}
    }
  }

  const muteLinkmicAudio = (linkmic) => {
    sendSignaling('mute-target', linkmic.viewerId, { targetUserId: linkmic.viewerId })
    ElMessage.success('已请求 ' + linkmic.viewerName + ' 静音')
  }

  const muteLinkmicVideo = (linkmic) => {
    sendSignaling('mute-video-target', linkmic.viewerId, { targetUserId: linkmic.viewerId })
    ElMessage.success('已请求 ' + linkmic.viewerName + ' 关摄像头')
  }

  const handleLinkmicApplyMessage = (msg) => {
    if (!isStreamer.value || !msg.data) return
    if (!pendingApplications.value.find(item => item.id === msg.data.id)) {
      pendingApplications.value.push(msg.data)
    }
  }

  const handleLinkmicAcceptedMessage = async (msg) => {
    myLinkmicId.value = msg.data?.linkmicId
    myQueuePosition.value = 0
    ElMessage.success('连麦已接通，正在建立连接...')
    await viewerStartConnection(msg.userId)
  }

  const handleLinkmicRejectedMessage = () => {
    ElMessage.warning('连麦被拒绝')
    myLinkmicId.value = null
    myQueuePosition.value = 0
    viewerHandRaised.value = false
  }

  const handleLinkmicDisconnectedMessage = (msg) => {
    const viewerId = msg.data?.viewerId
    const disconnectedMe = !isStreamer.value && (sameUserId(viewerId, currentUserId.value) || isCurrentUserTarget(msg))
    const peerId = disconnectedMe ? msg.userId : (viewerId || msg.userId)
    removeLinkmicStream(peerId)
    if (!disconnectedMe) return
    myLinkmicId.value = null
    myQueuePosition.value = 0
    viewerHandRaised.value = false
    stopLocalLinkmicStream()
    ElMessage.info('连麦已断开')
  }

  const handleHandRaisedMessage = (msg) => {
    if (!isStreamer.value || !msg.data) return
    const viewerId = msg.userId
    const raised = !!msg.data.raised
    const app = pendingApplications.value.find(item => item.viewerId === viewerId)
    if (app) app.handRaised = raised
    if (raised) ElMessage.info(`${msg.userName || ('观众' + viewerId)} 举手了`)
  }

  const handleRemoteMediaToggle = async (msg, enabledRef, toggle, message) => {
    if (!isCurrentUserTarget(msg)) return false
    if (enabledRef.value) {
      await toggle()
      ElMessage.info(message)
    }
    return true
  }

  const handleMuteAudioMessage = async (msg) => {
    return handleRemoteMediaToggle(msg, localLinkmicAudioEnabled, toggleLocalAudio, '主播已将你静音')
  }

  const handleMuteVideoMessage = async (msg) => {
    return handleRemoteMediaToggle(msg, localLinkmicVideoEnabled, toggleLocalVideo, '主播已关闭你的摄像头')
  }

  const handlePeerStateMessage = (msg) => {
    if (!msg.userId || !msg.data) return false
    ensureLinkmicEntry(msg.userId, msg.userName)
    applyLinkmicPeerState(linkmicStreams[msg.userId], msg.data)
    return true
  }

  return {
    myLinkmicId,
    myQueuePosition,
    activeLinkmics,
    pendingApplications,
    linkmicLoading,
    linkmicStreams,
    localLinkmicVideoRef,
    localLinkmicAudioEnabled,
    localLinkmicVideoEnabled,
    viewerHandRaised,
    loadLinkmicStatus,
    cleanupLinkmic,
    setRemoteVideo,
    applyLinkmic,
    acceptLinkmic,
    rejectLinkmic,
    disconnectLinkmic,
    viewerHangup,
    toggleViewerHandRaise,
    toggleLocalAudio,
    toggleLocalVideo,
    muteLinkmicAudio,
    muteLinkmicVideo,
    handleLinkmicApplyMessage,
    handleLinkmicAcceptedMessage,
    handleLinkmicRejectedMessage,
    handleLinkmicDisconnectedMessage,
    handleHandRaisedMessage,
    handleMuteAudioMessage,
    handleMuteVideoMessage,
    handlePeerStateMessage,
    handleOffer,
    handleAnswer,
    handleIceCandidate
  }
}
