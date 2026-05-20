<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Microphone, Mute, VideoCamera, VideoPause, UserFilled, CopyDocument, Plus, Monitor, ChatDotRound, Share } from '@element-plus/icons-vue'
import { meetingApi } from '../../api/meeting.js'
import { createReconnectingWS } from '../../utils/reconnectingWs.js'

const router = useRouter()
const loading = ref(false)
const myRooms = ref([])
const joinCode = ref('')
const newRoomName = ref('')

// 当前用户（从 localStorage 读，登录态保证；未登录用临时 ID 兜底）
const getCurrentUser = () => {
  try {
    const u = JSON.parse(localStorage.getItem('user') || '{}')
    if (u && u.id) return { id: Number(u.id), name: u.username || u.nickname || ('用户' + u.id) }
  } catch (e) {}
  // 未登录场景的临时身份
  let tmp = sessionStorage.getItem('meetingTmpId')
  if (!tmp) {
    tmp = String(Date.now() % 1000000)
    sessionStorage.setItem('meetingTmpId', tmp)
  }
  return { id: Number(tmp), name: '访客' + tmp }
}
const me = ref(getCurrentUser())

const localStream = ref(null)
const localVideoRef = ref(null)
const currentRoom = ref(null)

// 远端：peerId(=对端 userId) -> { stream, name }
const remotePeers = reactive({})    // 响应式，驱动 video 网格
const peerConnections = {}          // 非响应式，纯运行时
const pendingIce = {}               // setRemote 之前到的 ice 暂存

const audioEnabled = ref(true)
const videoEnabled = ref(true)
const screenShareEnabled = ref(false)
const virtualBgEnabled = ref(false)   // 虚拟背景/背景模糊
const virtualBgStream = ref(null)    // 处理后的流（含背景模糊）
const roomLocked = ref(false)       // 会议是否被锁定
let ws = null

onMounted(async () => {
  await loadMyRooms()
  window.addEventListener('keydown', handleKeydown)
  statsInterval = setInterval(pollPeerQuality, 5000)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  stopRecording()
  if (statsInterval) clearInterval(statsInterval)
  cleanup()
})

const pollPeerQuality = async () => {
  if (!currentRoom.value) return
  for (const [peerId, pc] of Object.entries(peerConnections)) {
    try {
      const stats = await pc.getStats()
      let rtt = 0, lost = 0, total = 0
      stats.forEach(report => {
        if (report.type === 'candidate-pair' && report.state === 'succeeded') {
          rtt = report.currentRoundTripTime ? Math.round(report.currentRoundTripTime * 1000) : 0
        }
        if (report.type === 'inbound-rtp') {
          lost = report.packetsLost || 0
          total = report.packetsReceived + lost
        }
      })
      const lossRate = total > 0 ? Math.round((lost / total) * 100) : 0
      peerQuality[peerId] = { rtt, lossRate }
    } catch (e) {}
  }
}

const getQualityLevel = (peerId) => {
  const q = peerQuality[peerId]
  if (!q) return 0
  if (q.lossRate > 10 || q.rtt > 500) return 1
  if (q.lossRate > 3 || q.rtt > 200) return 2
  return 3
}

let recordingTimer = null

const startRecording = () => {
  const streams = []
  if (localStream.value) streams.push(localStream.value)
  Object.values(remotePeers).forEach(p => {
    if (p.stream) streams.push(p.stream)
  })
  if (streams.length === 0) {
    ElMessage.warning('没有可录制的媒体流')
    return
  }
  const combined = new MediaStream()
  streams.forEach(s => s.getTracks().forEach(t => combined.addTrack(t)))
  recordedChunks = []
  mediaRecorder = new MediaRecorder(combined, { mimeType: 'video/webm;codecs=vp9' })
  mediaRecorder.ondataavailable = (e) => { if (e.data.size > 0) recordedChunks.push(e.data) }
  mediaRecorder.start(1000)
  isRecording.value = true
  recordingTime.value = 0
  recordingTimer = setInterval(() => { recordingTime.value++ }, 1000)
  ElMessage.success('开始录制')
}

const stopRecording = () => {
  if (!mediaRecorder || mediaRecorder.state === 'inactive') return
  mediaRecorder.stop()
  clearInterval(recordingTimer)
  isRecording.value = false
  const blob = new Blob(recordedChunks, { type: 'video/webm' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `会议录制_${new Date().toLocaleString()}.webm`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('录制已保存')
}

const formatRecordingTime = (s) => {
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

const formatDuration = (s) => {
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

const loadMyRooms = async () => {
  try {
    const res = await meetingApi.getMyRooms()
    if (res.code === 200) {
      myRooms.value = res.data || []
    }
  } catch (e) {
    console.error('获取会议室失败:', e)
  }
}

const createRoom = async () => {
  if (!newRoomName.value.trim()) {
    ElMessage.warning('请输入会议室名称')
    return
  }
  loading.value = true
  try {
    const res = await meetingApi.createRoom(newRoomName.value)
    if (res.code === 200) {
      ElMessage.success('创建成功')
      newRoomName.value = ''
      await loadMyRooms()
      await joinRoomByCode(res.data.roomCode)
    }
  } catch (e) {
    ElMessage.error('创建失败')
  } finally {
    loading.value = false
  }
}

const joinRoomByCode = async (code) => {
  const roomCode = code || joinCode.value.trim()
  if (!roomCode) {
    ElMessage.warning('请输入会议室邀请码')
    return
  }
  loading.value = true
  try {
    const res = await meetingApi.getRoom(roomCode)
    if (res.code !== 200) {
      ElMessage.error(res.message || '会议室不存在')
      return
    }
    currentRoom.value = res.data

    // 非主持人先发加入请求，等主持人批准后再进
    if (!isHost.value) {
      waitingForHost.value = true
      ws = createReconnectingWS({
        url: `${proto}//${location.host}/ws/meeting`,
        onOpen: () => {
          ws.send({
            type: 'join-request',
            roomCode,
            userId: me.value.id,
            userName: me.value.name
          })
        },
        onMessage: async (msg) => {
          await handleSignaling(msg)
        },
        onClose: () => console.log('[Meeting WS] 断开（将自动重连）')
      })
      ElMessage.info('正在等待主持人批准...')
      loading.value = false
      return
    }

    const joinRes = await meetingApi.joinRoom(roomCode)
    if (joinRes.code !== 200) {
      ElMessage.error(joinRes.message || '加入失败')
      return
    }

    await startLocalStream()
    connectWebSocket(roomCode)
    meetingDuration.value = 0
    meetingTimer = setInterval(() => { meetingDuration.value++ }, 1000)
    ElMessage.success('加入成功')
  } catch (e) {
    console.error(e)
    ElMessage.error('加入失败')
  } finally {
    loading.value = false
  }
}

const startLocalStream = async () => {
  try {
    localStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    await nextTick()
    if (localVideoRef.value) {
      localVideoRef.value.srcObject = localStream.value
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('无法获取摄像头/麦克风')
  }
}

const connectWebSocket = (roomCode) => {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = createReconnectingWS({
    url: `${proto}//${location.host}/ws/meeting`,
    onOpen: () => {
      // 每次（含重连）建连后都重新加入房间
      ws.send({
        type: 'join',
        roomCode,
        userId: me.value.id,
        userName: me.value.name
      })
    },
    onMessage: async (msg) => {
      await handleSignaling(msg)
    },
    onClose: () => console.log('[Meeting WS] 断开（将自动重连）')
  })
}

const handleSignaling = async (msg) => {
  switch (msg.type) {
    case 'room-users':
      if (!isHost.value) {
        // 非主持人：收到房间成员说明已获准进入
        waitingForHost.value = false
      }
      for (const u of (msg.data || [])) {
        ensurePeerEntry(u.userId, u.userName)
        await createPeerConnection(u.userId, true)
      }
      broadcastSelfState()
      break
    case 'join-request':
      // 主持人收到加入申请
      if (isHost.value) {
        if (!waitingUsers.value.find(w => w.userId === msg.userId)) {
          waitingUsers.value.push({ userId: msg.userId, userName: msg.userName, ts: Date.now() })
          ElMessage.info(`${msg.userName || ('用户' + msg.userId)} 请求加入会议`)
        }
      }
      break
    case 'admit-user':
      // 收到批准，可以进入会议了
      if (msg.targetUserId === me.value.id) {
        waitingForHost.value = false
        await startLocalStream()
        connectWebSocket(currentRoom.value.roomCode)
        ElMessage.success('主持人已允许加入')
      }
      break
    case 'user-joined':
      // 我是已有成员，等新人发 offer 过来即可（不主动发，避免双方都发 offer）
      ensurePeerEntry(msg.userId, msg.userName)
      // 给新人广播一下自己的状态
      broadcastSelfState()
      break
    case 'user-left':
      removePeerConnection(msg.userId)
      break
    case 'offer':
      await handleOffer(msg.userId, msg.data)
      break
    case 'answer':
      await handleAnswer(msg.userId, msg.data)
      break
    case 'ice-candidate':
      await handleIceCandidate(msg.userId, msg.data)
      break
    case 'peer-state':
      if (msg.data && remotePeers[msg.userId]) {
        // 记录音频开关变化用于发言检测
        if (msg.data.audioEnabled !== undefined) {
          updateAudioActivity(msg.userId, msg.data.audioEnabled)
        }
        Object.assign(remotePeers[msg.userId], msg.data)
      } else if (msg.data) {
        remotePeers[msg.userId] = { name: msg.userName || ('用户' + msg.userId), ...msg.data }
      }
      break
    case 'spotlight':
      if (msg.data?.targetUserId) {
        spotlightId.value = msg.data.active ? msg.data.targetUserId : null
      }
      break
    case 'lock-room':
      if (!isHost.value) {
        // 非主持人收到锁定状态更新
        roomLocked.value = msg.data?.locked === true
        if (roomLocked.value) {
          ElMessage.warning('会议已被主持人锁定，请等待')
        } else {
          ElMessage.success('会议已解锁')
        }
      }
      break
    case 'transfer-host':
      if (msg.targetUserId === me.value.id) {
        isHost.value = true
        ElMessage.success('你已成为新的主持人')
      } else if (msg.userId === me.value.id) {
        isHost.value = false
      }
      break
    case 'kick':
      if (msg.targetUserId === me.value.id || msg.data?.targetUserId === me.value.id) {
        ElMessage.warning('你已被主持人移出会议')
        leaveRoom()
      }
      break
    case 'mute-target':
      if (msg.targetUserId === me.value.id || msg.data?.targetUserId === me.value.id) {
        if (audioEnabled.value) toggleAudio()
        ElMessage.info('主持人已将你静音')
      }
      break
    case 'mute-all':
      if (msg.userId !== me.value.id && audioEnabled.value) {
        toggleAudio()
        ElMessage.info('主持人已全员静音')
      }
      break
    case 'chat':
      if (msg.data?.text) {
        const isMentioned = msg.targetUserId === me.value.id || (msg.data.text && msg.data.text.includes('@' + me.value.name))
        chatMessages.value.push({
          id: Date.now() + Math.random(),
          from: msg.userName || ('用户' + msg.userId),
          text: msg.data.text,
          self: false,
          mentionId: msg.targetUserId === me.value.id ? msg.userId : (isMentioned ? msg.userId : null)
        })
        if (chatMessages.value.length > 200) chatMessages.value.splice(0, 50)
        if (isMentioned && !msg.data?.text.includes('@' + me.value.name)) {
          ElMessage.info(`${msg.userName} 在会议中提到了你`)
        }
      }
      break
    case 'hand-raised':
      // 更新对应 peer 的 handRaised 状态
      if (remotePeers[msg.userId]) {
        remotePeers[msg.userId].handRaised = !!msg.data?.raised
      }
      // 主持人收到举手通知
      if (isHost.value && msg.data?.raised) {
        ElMessage.info(`${msg.userName || ('用户' + msg.userId)} 举手了`)
      }
      break
    case 'meeting-ended':
      ElMessage.warning('主持人已结束会议')
      cleanup()
      currentRoom.value = null
      await loadMyRooms()
      break
  }
}

// 占位（避免 ontrack 来时还没有条目导致 UI 不显示昵称）
const ensurePeerEntry = (peerId, name) => {
  if (!remotePeers[peerId]) {
    remotePeers[peerId] = {
      name: name || ('用户' + peerId),
      audioEnabled: true,
      videoEnabled: true,
      screenShareEnabled: false,
      handRaised: false
    }
  } else if (name && !remotePeers[peerId].name) {
    remotePeers[peerId].name = name
  }
}

const broadcastSelfState = () => {
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'peer-state',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    userName: me.value.name,
    data: {
      audioEnabled: audioEnabled.value,
      videoEnabled: videoEnabled.value,
      screenShareEnabled: screenShareEnabled.value
    }
  })
}

const createPeerConnection = async (peerId, isInitiator) => {
  if (peerConnections[peerId]) return peerConnections[peerId]

  const pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  })
  peerConnections[peerId] = pc

  if (localStream.value) {
    localStream.value.getTracks().forEach(track => {
      pc.addTrack(track, localStream.value)
    })
  }

  pc.ontrack = (event) => {
    ensurePeerEntry(peerId)
    remotePeers[peerId].stream = event.streams[0]
  }

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      sendSignaling('ice-candidate', peerId, event.candidate)
    }
  }

  pc.onconnectionstatechange = () => {
    if (['failed', 'closed', 'disconnected'].includes(pc.connectionState)) {
      console.log('[Meeting] 连接状态', peerId, pc.connectionState)
    }
  }

  pc.oniceconnectionstatechange = async () => {
    if (pc.iceConnectionState === 'failed') {
      console.warn('[Meeting] ICE failed, 尝试 restart', peerId)
      try {
        const offer = await pc.createOffer({ iceRestart: true })
        await pc.setLocalDescription(offer)
        sendSignaling('offer', peerId, offer.sdp)
      } catch (e) {
        console.error('[Meeting] ICE restart 失败', e)
      }
    }
  }

  if (isInitiator) {
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignaling('offer', peerId, offer.sdp)
  }
  return pc
}

const handleOffer = async (peerId, sdp) => {
  // 收到 offer 说明对端发起新连接，旧 PC 替换掉避免状态错乱
  if (peerConnections[peerId]) {
    peerConnections[peerId].close()
    delete peerConnections[peerId]
  }
  const pc = await createPeerConnection(peerId, false)
  await pc.setRemoteDescription({ type: 'offer', sdp })
  // flush 暂存 ice
  if (pendingIce[peerId]) {
    for (const c of pendingIce[peerId]) await pc.addIceCandidate(c)
    delete pendingIce[peerId]
  }
  const answer = await pc.createAnswer()
  await pc.setLocalDescription(answer)
  sendSignaling('answer', peerId, answer.sdp)
}

const handleAnswer = async (peerId, sdp) => {
  const pc = peerConnections[peerId]
  if (!pc) return
  await pc.setRemoteDescription({ type: 'answer', sdp })
  if (pendingIce[peerId]) {
    for (const c of pendingIce[peerId]) await pc.addIceCandidate(c)
    delete pendingIce[peerId]
  }
}

const handleIceCandidate = async (peerId, candidate) => {
  const pc = peerConnections[peerId]
  if (pc && pc.remoteDescription) {
    try { await pc.addIceCandidate(candidate) } catch (e) { console.warn('addIceCandidate', e) }
  } else {
    (pendingIce[peerId] = pendingIce[peerId] || []).push(candidate)
  }
}

const sendSignaling = (type, targetUserId, data) => {
  if (!ws || !ws.isOpen()) return
  ws.send({
    type,
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    userName: me.value.name,
    targetUserId,
    data
  })
}

const removePeerConnection = (peerId) => {
  const pc = peerConnections[peerId]
  if (pc) {
    pc.close()
    delete peerConnections[peerId]
  }
  delete remotePeers[peerId]
  delete pendingIce[peerId]
}

const toggleVideo = () => {
  if (!localStream.value) return
  const track = localStream.value.getVideoTracks()[0]
  if (track) {
    track.enabled = !track.enabled
    videoEnabled.value = track.enabled
    broadcastSelfState()
  }
}

// 虚拟背景 / 背景模糊
// 使用 Canvas captureStream 实现背景模糊（轻量方案，无需外部 ML 库）
const virtualBgVideo = ref(null)
let virtualBgRafId = null

const applyVirtualBg = async () => {
  if (!localStream.value) return
  const videoTrack = localStream.value.getVideoTracks()[0]
  if (!videoTrack) return
  try {
    // 创建用于捕获摄像头并叠加模糊背景的 canvas
    const inputCanvas = document.createElement('canvas')
    inputCanvas.width = 640; inputCanvas.height = 480
    const inputCtx = inputCanvas.getContext('2d')
    const blurCanvas = document.createElement('canvas')
    blurCanvas.width = 640; blurCanvas.height = 480
    const blurCtx = blurCanvas.getContext('2d')
    blurCtx.filter = 'blur(14px)'
    blurCtx.drawImage(inputCanvas, 0, 0, 640, 480)

    // 隐藏式 video 元素连接到本地轨道
    const hiddenVideo = document.createElement('video')
    hiddenVideo.srcObject = localStream.value
    hiddenVideo.autoplay = true
    hiddenVideo.muted = true
    hiddenVideo.play()

    const outputStream = new MediaStream()

    // 用 MediaRecorder 从 canvas 编码，再从 blob 构建流（简化处理）
    // 这里直接用 requestVideoFrameCallback 把每帧画到 canvas 并加模糊
    let lastTime = 0
    const processFrame = (timestamp) => {
      if (!virtualBgEnabled.value) return
      if (hiddenVideo.readyState >= 2) {
        inputCtx.drawImage(hiddenVideo, 0, 0, 640, 480)
        // 取原始画面中心区域加模糊作为背景
        blurCtx.drawImage(inputCanvas, 0, 0, 640, 480)
        // 画主图覆盖中心（人像区域），留出模糊边框形成"抠图"效果
        inputCtx.drawImage(blurCanvas, 0, 0, 640, 480)
        // 中心矩形保留清晰（前景区）
        const cx = 320, cy = 240, cw = 200, ch = 300
        inputCtx.clearRect(cx - cw/2, cy - ch/2, cw, ch)
        inputCtx.drawImage(hiddenVideo, cx - cw/2, cy - ch/2, cw, ch, cx - cw/2, cy - ch/2, cw, ch)
      }
      virtualBgRafId = requestAnimationFrame(processFrame)
    }
    virtualBgRafId = requestVideoFrameCallback(processFrame)

    // 用 CaptureStream 收集 canvas 输出
    const canvasStream = inputCanvas.captureStream(30)
    const canvasTrack = canvasStream.getVideoTracks()[0]
    outputStream.addTrack(canvasTrack)
    virtualBgStream.value = outputStream

    // 替换本地预览
    if (localVideoRef.value) localVideoRef.value.srcObject = outputStream
    // 替换所有 peer 的视频轨道
    Object.values(peerConnections).forEach(pc => {
      const sender = pc.getSenders().find(s => s.track && s.track.kind === 'video')
      if (sender) sender.replaceTrack(canvasTrack)
    })
    virtualBgEnabled.value = true
    ElMessage.success('已开启背景模糊')
  } catch (e) {
    console.error(e)
    ElMessage.error('虚拟背景开启失败')
  }
}

const stopVirtualBg = () => {
  virtualBgEnabled.value = false
  if (virtualBgRafId) {
    cancelAnimationFrame(virtualBgRafId)
    virtualBgRafId = null
  }
  if (virtualBgStream.value) {
    virtualBgStream.value.getTracks().forEach(t => t.stop())
    virtualBgStream.value = null
  }
  // 恢复原始摄像头轨道
  if (localStream.value) {
    const camTrack = localStream.value.getVideoTracks()[0]
    Object.values(peerConnections).forEach(pc => {
      const sender = pc.getSenders().find(s => s.track && s.track.kind === 'video')
      if (sender && camTrack) sender.replaceTrack(camTrack)
    })
    if (localVideoRef.value) localVideoRef.value.srcObject = localStream.value
  }
}

const toggleVirtualBg = () => {
  if (virtualBgEnabled.value) {
    stopVirtualBg()
  } else {
    applyVirtualBg()
  }
}

const startScreenShare = async () => {
  if (screenShareEnabled.value) return stopScreenShare()
  try {
    const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: false })
    const screenTrack = screenStream.getVideoTracks()[0]
    Object.values(peerConnections).forEach(pc => {
      const sender = pc.getSenders().find(s => s.track && s.track.kind === 'video')
      if (sender) sender.replaceTrack(screenTrack)
    })
    if (localVideoRef.value) localVideoRef.value.srcObject = screenStream
    screenShareEnabled.value = true
    broadcastSelfState()
    screenTrack.onended = () => stopScreenShare()
  } catch (e) {
    if (e.name !== 'NotAllowedError') ElMessage.error('屏幕共享失败')
  }
}

const stopScreenShare = () => {
  if (!localStream.value) return
  const cameraTrack = localStream.value.getVideoTracks()[0]
  Object.values(peerConnections).forEach(pc => {
    const sender = pc.getSenders().find(s => s.track && s.track.kind === 'video')
    if (sender && cameraTrack) sender.replaceTrack(cameraTrack)
  })
  if (localVideoRef.value) localVideoRef.value.srcObject = localStream.value
  screenShareEnabled.value = false
  broadcastSelfState()
}

const cleanup = () => {
  stopRecording()
  if (meetingTimer) { clearInterval(meetingTimer); meetingTimer = null }
  if (ws && ws.isOpen()) {
    try {
      ws.send({
        type: 'leave',
        roomCode: currentRoom.value?.roomCode,
        userId: me.value.id
      })
    } catch (e) {}
  }
  if (ws) { ws.close(); ws = null }
  Object.keys(peerConnections).forEach(removePeerConnection)
  if (localStream.value) {
    localStream.value.getTracks().forEach(t => t.stop())
    localStream.value = null
  }
}

const shareRoom = () => {
  const url = `${location.origin}/meeting?code=${currentRoom.value.roomCode}`
  navigator.clipboard.writeText(url).then(() => {
    ElMessage.success('邀请链接已复制，可直接发送给朋友加入会议')
  })
}

const copyRoomCode = (code) => {
  navigator.clipboard.writeText(code).then(() => {
    ElMessage.success('邀请码已复制')
  })
}

const leaveRoom = async () => {
  if (currentRoom.value && isHost.value) {
    // 主持人确认是否结束整场会议
    let endMeeting = false
    try {
      await ElMessageBox.confirm('结束会议会让所有人退出，是否结束？否则你只是自己退出。', '主持人离开', {
        confirmButtonText: '结束会议',
        cancelButtonText: '仅自己离开',
        type: 'warning'
      })
      endMeeting = true
    } catch (e) {
      endMeeting = false
    }
    if (endMeeting) {
      try { await meetingApi.endRoom(currentRoom.value.id) } catch (e) {}
      // 广播给房间内所有人，让他们退出
      if (ws && ws.isOpen()) {
        ws.send({
          type: 'meeting-ended',
          roomCode: currentRoom.value.roomCode,
          userId: me.value.id,
          data: {}
        })
      }
    } else {
      try { await meetingApi.leaveRoom(currentRoom.value.id) } catch (e) {}
    }
  } else if (currentRoom.value) {
    try { await meetingApi.leaveRoom(currentRoom.value.id) } catch (e) {}
  }
  cleanup()
  currentRoom.value = null
  await loadMyRooms()
}

// video 元素 ref 回调：把对应 peer 的 stream 绑上去
const setRemoteVideo = (peerId, el) => {
  if (!el) return
  const peer = remotePeers[peerId]
  if (peer && peer.stream && el.srcObject !== peer.stream) {
    el.srcObject = peer.stream
  }
}

const remotePeerIds = computed(() => Object.keys(remotePeers))
const isHost = computed(() => currentRoom.value && me.value.id === currentRoom.value.creatorId)
const sidebarOpen = ref(true)
const chatOpen = ref(false)
const chatMessages = ref([])
const chatText = ref('')
const showEmojiPicker = ref(false)
const emojis = ['😀', '😂', '👍', '👏', '❤️', '🔥', '🤔', '😅', '🙌', '😎', '🤝', '👀', '🎉', '💯', '😱', '🤯']
const controlsVisible = ref(true)
const activeSpeakerId = ref(null)
const handRaised = ref(false)
const isRecording = ref(false)
const recordingTime = ref(0)
const waitingUsers = ref([])  // 待审批的加入申请
const waitingForHost = ref(false)  // 加入时等待主持人审批
const editingName = ref(false)
const editNameInput = ref(null)
const showMentionDropdown = ref(false)
const mentionQuery = ref('')
const mentionSelected = ref(null)
const meetingDuration = ref(0)
const spotlightId = ref(null)
let hideControlsTimer = null
let meetingTimer = null
let qualityTimer = null
const peerQuality = reactive({})  // peerId -> { rtt, lost }
let statsInterval = null
let speakerTimer = null
let mediaRecorder = null
let recordedChunks = []

// ============ 发言检测（音频活动 > 2s 判定为发言者） ============
const trackAudioStart = {}  // peerId -> timestamp when audio enabled

const updateAudioActivity = (peerId, isEnabled) => {
  if (isEnabled) {
    if (!trackAudioStart[peerId]) trackAudioStart[peerId] = Date.now()
    const elapsed = Date.now() - trackAudioStart[peerId]
    if (elapsed > 2000 && activeSpeakerId.value !== peerId) {
      activeSpeakerId.value = peerId
      resetSpeakerTimer()
    }
  } else {
    delete trackAudioStart[peerId]
    if (activeSpeakerId.value === peerId) {
      activeSpeakerId.value = null
    }
  }
}

const resetSpeakerTimer = () => {
  if (speakerTimer) clearTimeout(speakerTimer)
  speakerTimer = setTimeout(() => {
    activeSpeakerId.value = null
  }, 3000)
}

// 在 toggleAudio 时也触发发言检测
const toggleAudio = () => {
  if (!localStream.value) return
  const track = localStream.value.getAudioTracks()[0]
  if (track) {
    track.enabled = !track.enabled
    audioEnabled.value = track.enabled
    broadcastSelfState()
    updateAudioActivity('local', track.enabled)
  }
}
const participantList = computed(() => {
  const list = [{
    userId: me.value.id,
    name: me.value.name + '（我）',
    audioEnabled: audioEnabled.value,
    videoEnabled: videoEnabled.value,
    screenShareEnabled: screenShareEnabled.value,
    isHost: currentRoom.value && me.value.id === currentRoom.value.creatorId,
    self: true
  }]
  for (const id of Object.keys(remotePeers)) {
    const p = remotePeers[id]
    list.push({
      userId: Number(id),
      name: p.name || ('用户' + id),
      audioEnabled: p.audioEnabled !== false,
      videoEnabled: p.videoEnabled !== false,
      screenShareEnabled: !!p.screenShareEnabled,
      handRaised: !!p.handRaised,
      isHost: currentRoom.value && Number(id) === currentRoom.value.creatorId,
      self: false
    })
  }
  return list
})

const totalHandRaised = computed(() => {
  return participantList.value.filter(p => p.handRaised && !p.self).length
})

// ============ 主持人举手控制 ============
const approveHandRaise = (userId) => {
  lowerHand(userId)
  ElMessage.success('已批准发言')
}

const lowerHand = (userId) => {
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'hand-raised',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    targetUserId: userId,
    data: { raised: false }
  })
  if (remotePeers[userId]) remotePeers[userId].handRaised = false
}

const admitUser = (userId) => {
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'admit-user',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    targetUserId: userId
  })
  waitingUsers.value = waitingUsers.value.filter(w => w.userId !== userId)
}

const rejectUser = (userId) => {
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'reject-user',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    targetUserId: userId
  })
  waitingUsers.value = waitingUsers.value.filter(w => w.userId !== userId)
}

const spotlightUser = (userId) => {
  spotlightId.value = spotlightId.value === userId ? null : userId
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'spotlight',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    data: { targetUserId: userId, active: spotlightId.value === userId }
  })
}

const toggleLockRoom = () => {
  roomLocked.value = !roomLocked.value
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'lock-room',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    data: { locked: roomLocked.value }
  })
  ElMessage.success(roomLocked.value ? '会议已锁定' : '会议已解锁')
}

const transferHost = async (userId) => {
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'transfer-host',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    targetUserId: userId
  })
  isHost.value = false
  ElMessage.success('已将主持人转让')
}

const kickUser = (userId) => {
  if (!isHost.value) return
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'kick',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    targetUserId: userId,
    data: { targetUserId: userId }
  })
  removePeerConnection(userId)
}

const muteUser = (userId) => {
  if (!isHost.value) return
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'mute-target',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    targetUserId: userId,
    data: { targetUserId: userId }
  })
}

const muteAll = () => {
  if (!isHost.value) return
  if (!ws || !ws.isOpen()) return
  ws.send({
    type: 'mute-all',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    data: {}
  })
  ElMessage.success('已请求全员静音')
}

const screenSharerId = computed(() => {
  const entry = Object.entries(remotePeers).find(([, p]) => p.screenShareEnabled)
  return entry ? entry[0] : null
})

const showControls = () => {
  controlsVisible.value = true
  if (hideControlsTimer) clearTimeout(hideControlsTimer)
  hideControlsTimer = setTimeout(() => { controlsVisible.value = false }, 3000)
}

const exitFullscreen = () => {
  if (document.fullscreenElement) document.exitFullscreen()
}

const handleKeydown = (e) => {
  if (!currentRoom.value) return
  // 忽略 input/textarea 内的按键
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return
  switch (e.key) {
    case 'Escape':
      exitFullscreen()
      break
    case 'm':
    case 'M':
      toggleAudio()
      break
    case 'v':
    case 'V':
      toggleVideo()
      break
    case 's':
    case 'S':
      startScreenShare()
      break
    case 'h':
    case 'H':
      toggleHandRaise()
      break
  }
}

const handleVideoDoubleClick = (peerId) => {
  togglePiP(peerId)
}

const togglePiP = async (peerId) => {
  let videoEl = null
  if (peerId === 'local') {
    videoEl = localVideoRef.value
  } else {
    const idx = remotePeerIds.indexOf(peerId)
    if (idx >= 0) {
      const cards = document.querySelectorAll('.video-card video')
      videoEl = cards[idx + 1]
    }
  }
  if (!videoEl) return
  try {
    if (document.pictureInPictureElement) {
      await document.exitPictureInPicture()
    } else if (document.pictureInPictureEnabled) {
      await videoEl.requestPictureInPicture()
    } else {
      ElMessage.warning('当前浏览器不支持画中画')
    }
  } catch (e) {
    console.warn('PiP error:', e)
  }
}

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

const insertEmoji = (emoji) => {
  chatText.value += emoji
  showEmojiPicker.value = false
}

const sendChat = () => {
  const text = chatText.value.trim()
  if (!text) return
  if (!ws || !ws.isOpen()) return
  const mentionId = mentionSelected.value
  ws.send({
    type: 'chat',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    userName: me.value.name,
    targetUserId: mentionId,
    data: { text }
  })
  // 自己也立即看见
  chatMessages.value.push({
    id: Date.now(),
    from: me.value.name,
    text,
    self: true,
    mentionId
  })
  chatText.value = ''
  mentionSelected.value = null
  showMentionDropdown.value = false
}

const handleChatInput = (e) => {
  const val = chatText.value
  const atIdx = val.lastIndexOf('@')
  if (atIdx >= 0 && atIdx === val.length - 1) {
    showMentionDropdown.value = true
    mentionQuery.value = ''
  } else if (atIdx >= 0) {
    const query = val.slice(atIdx + 1)
    mentionQuery.value = query
    showMentionDropdown.value = true
  } else {
    showMentionDropdown.value = false
  }
}

const selectMention = (p) => {
  const atIdx = chatText.value.lastIndexOf('@')
  chatText.value = chatText.value.slice(0, atIdx) + p.name + ' '
  mentionSelected.value = p.userId
  showMentionDropdown.value = false
}

const mentionSuggestions = computed(() => {
  if (!showMentionDropdown.value) return []
  const q = mentionQuery.value.toLowerCase()
  return participantList.value.filter(p => p.name.toLowerCase().includes(q) && !p.self)
})

const startEditName = () => {
  editingName.value = true
  nextTick(() => editNameInput.value?.select())
}
const saveName = () => {
  const newName = editNameInput.value?.trim()
  if (!newName) { editingName.value = false; return }
  me.value.name = newName
  broadcastSelfState()
  ElMessage.success('昵称已更新')
  editingName.value = false
}

const toggleHandRaise = () => {
  if (!ws || !ws.isOpen()) return
  handRaised.value = !handRaised.value
  ws.send({
    type: 'hand-raised',
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    userName: me.value.name,
    data: { raised: handRaised.value }
  })
  ElMessage.info(handRaised.value ? '已举手，请等待主持人允许' : '已放下手')
}
</script>

<template>
  <div class="meeting-page">
    <!-- 非主持人等待审批 -->
    <div v-if="waitingForHost" class="waiting-room">
      <div class="waiting-card">
        <div class="waiting-icon">⏳</div>
        <h3>等待主持人批准</h3>
        <p>请耐心等待，主持人允许后将自动加入会议</p>
        <el-button type="danger" @click="waitingForHost = false; currentRoom = null">取消等待</el-button>
      </div>
    </div>

    <!-- 主持人审批面板 -->
    <div v-if="isHost && waitingUsers.length > 0" class="waiting-approve-bar">
      <div class="approve-title">加入申请 ({{ waitingUsers.length }})</div>
      <div v-for="w in waitingUsers" :key="w.userId" class="waiting-item">
        <span>{{ w.userName || ('用户' + w.userId) }}</span>
        <el-button size="small" type="success" @click="admitUser(w.userId)">允许</el-button>
        <el-button size="small" type="danger" @click="rejectUser(w.userId)">拒绝</el-button>
      </div>
    </div>

    <!-- 会议中视图 -->
    <div v-if="currentRoom && !waitingForHost" class="meeting-active">
      <div class="meeting-header">
        <div class="header-left">
          <h2>{{ currentRoom.roomName }}</h2>
          <span class="room-code" @click="copyRoomCode(currentRoom.roomCode)">
            <CopyDocument style="width: 12px; height: 12px;" />
            {{ currentRoom.roomCode }}
          </span>
          <el-button size="small" @click="shareRoom" style="font-size:12px;">
            <el-icon><Share /></el-icon>
            邀请加入
          </el-button>
          <el-tag v-if="isHost" type="warning" size="small" effect="dark">主持人</el-tag>
          <el-tag type="info" size="small">{{ formatDuration(meetingDuration) }}</el-tag>
        </div>
        <div class="header-right">
          <el-button v-if="isHost" type="warning" plain @click="muteAll">全员静音</el-button>
          <el-button @click="chatOpen = !chatOpen">
            <el-icon><ChatDotRound /></el-icon>
            <span style="margin-left:4px">聊天</span>
          </el-button>
          <el-button @click="sidebarOpen = !sidebarOpen">
            <el-icon><UserFilled /></el-icon>
            <span style="margin-left:4px">{{ participantList.length }}</span>
          </el-button>
          <el-button v-if="!isRecording" type="info" plain @click="startRecording">
            <el-icon><VideoCamera /></el-icon>
            录制
          </el-button>
          <el-button v-else type="danger" plain @click="stopRecording">
            <span class="rec-dot" />{{ formatRecordingTime(recordingTime) }}
          </el-button>
          <el-button type="danger" @click="leaveRoom">离开会议</el-button>
        </div>
      </div>

      <div class="meeting-body" @mousemove="showControls" :class="{ 'speaker-mode': activeSpeakerId && !screenSharerId }">
        <div class="video-grid" :class="{ 'has-sharer': screenSharerId, 'has-speaker': activeSpeakerId && !screenSharerId, 'has-spotlight': spotlightId && !screenSharerId }">
          <!-- 本地视频 -->
          <div class="video-card local" :class="{ 'main-sharer': screenSharerId === 'local', 'active-speaker': activeSpeakerId === 'local', 'spotlighted': spotlightId === 'local' }" @dblclick="handleVideoDoubleClick('local')">
            <video ref="localVideoRef" autoplay muted playsinline class="video-element" />
            <div class="video-label" :class="{ 'controls-hidden': !controlsVisible }">我</div>
            <div class="video-controls" :class="{ 'controls-hidden': !controlsVisible }">
              <el-button :type="audioEnabled ? '' : 'danger'" circle @click="toggleAudio">
                <el-icon><Microphone v-if="audioEnabled" /><Mute v-else /></el-icon>
              </el-button>
              <el-button :type="videoEnabled ? '' : 'danger'" circle @click="toggleVideo">
                <el-icon><VideoCamera v-if="videoEnabled" /><VideoPause v-else /></el-icon>
              </el-button>
              <el-button :type="screenShareEnabled ? 'primary' : ''" circle @click="startScreenShare">
                <el-icon><Monitor /></el-icon>
              </el-button>
              <el-button :type="virtualBgEnabled ? 'warning' : ''" circle @click="toggleVirtualBg" title="虚拟背景/背景模糊">
                🎨
              </el-button>
              <el-button circle @click="togglePiP('local')" title="画中画">
                <el-icon><VideoCamera /></el-icon>
              </el-button>
              <el-button :type="handRaised ? 'warning' : ''" circle @click="toggleHandRaise" title="举手">
                ✋
              </el-button>
              <el-button v-if="isHost" :type="roomLocked ? 'danger' : ''" circle @click="toggleLockRoom" :title="roomLocked ? '解锁会议' : '锁定会议'">
                🔒
              </el-button>
            </div>
          </div>
          <!-- 举手计数 -->
          <div v-if="totalHandRaised > 0" class="hand-count-badge">
            ✋ {{ totalHandRaised }} 人举手
          </div>

          <!-- 远端视频 -->
          <div v-for="peerId in remotePeerIds" :key="peerId" class="video-card" :class="{ 'main-sharer': screenSharerId === peerId, 'active-speaker': activeSpeakerId === peerId, 'spotlighted': spotlightId === peerId }" @dblclick="handleVideoDoubleClick(peerId)">
            <video :ref="el => setRemoteVideo(peerId, el)" autoplay playsinline class="video-element" />
            <div class="video-label" :class="{ 'controls-hidden': !controlsVisible }">{{ remotePeers[peerId]?.name || ('参与者 ' + (remotePeerIds.indexOf(peerId) + 1)) }}</div>
            <div v-if="remotePeers[peerId]" class="video-status" :class="{ 'controls-hidden': !controlsVisible }">
              <el-icon v-if="!remotePeers[peerId].audioEnabled" class="status-off"><Mute /></el-icon>
              <el-icon v-if="!remotePeers[peerId].videoEnabled" class="status-off"><VideoPause /></el-icon>
              <el-icon v-if="remotePeers[peerId].screenShareEnabled" class="status-on"><Monitor /></el-icon>
            </div>
          </div>
        </div>

        <!-- 参与者侧栏 -->
        <transition name="slide-x">
          <aside v-if="sidebarOpen" class="participant-sidebar">
            <div class="sidebar-header">
              参与者 ({{ participantList.length }})
              <span v-if="!editingName" class="my-name" @click="startEditName" title="点击修改昵称">
                我: {{ me.name }} ✏️
              </span>
              <input v-else ref="editNameInput" :value="me.name" class="name-input" @blur="saveName" @keydown.enter="saveName" />
            </div>
            <div class="participant-list">
              <div v-for="p in participantList" :key="p.userId" class="participant-item">
                <div class="p-info">
                  <span class="p-name">{{ p.name }}</span>
                  <el-tag v-if="p.isHost" type="warning" size="small">主持</el-tag>
                  <el-tag v-if="p.handRaised" type="info" size="small">✋ 举手</el-tag>
                </div>
                <div class="p-status">
                  <div :class="['quality-bars', `q${getQualityLevel(p.userId)}`]" :title="`延迟${peerQuality[p.userId]?.rtt || '?'}ms 丢包${peerQuality[p.userId]?.lossRate || 0}%`">
                    <span /><span /><span />
                  </div>
                  <el-icon :class="p.audioEnabled ? 'on' : 'off'" :title="p.audioEnabled ? '麦克风开' : '已静音'">
                    <Microphone v-if="p.audioEnabled" /><Mute v-else />
                  </el-icon>
                  <el-icon :class="p.videoEnabled ? 'on' : 'off'" :title="p.videoEnabled ? '摄像头开' : '摄像头关'">
                    <VideoCamera v-if="p.videoEnabled" /><VideoPause v-else />
                  </el-icon>
                  <el-icon v-if="p.screenShareEnabled" class="on" title="屏幕共享中"><Monitor /></el-icon>
                </div>
                <div v-if="isHost && !p.self" class="p-actions">
                  <el-button v-if="p.handRaised" size="small" type="success" @click="approveHandRaise(p.userId)">批准发言</el-button>
                  <el-button v-if="p.handRaised" size="small" link type="warning" @click="lowerHand(p.userId)">放下</el-button>
                  <el-button size="small" :type="spotlightId === p.userId ? 'warning' : ''" link @click="spotlightUser(p.userId)">聚焦</el-button>
                  <el-button size="small" link @click="muteUser(p.userId)" :disabled="!p.audioEnabled">静音</el-button>
                  <el-button size="small" link type="warning" @click="transferHost(p.userId)">转交主持</el-button>
                  <el-button size="small" link type="danger" @click="kickUser(p.userId)">移出</el-button>
                </div>
              </div>
            </div>
          </aside>
        </transition>

        <!-- 聊天侧栏 -->
        <transition name="slide-x">
          <aside v-if="chatOpen" class="chat-sidebar">
            <div class="sidebar-header">会议聊天</div>
            <div class="chat-messages">
              <div v-if="chatMessages.length === 0" class="chat-empty">暂无消息</div>
              <div v-for="m in chatMessages" :key="m.id" :class="['chat-msg', { self: m.self, mentioned: m.mentionId }]">
                <div class="chat-from">
                  {{ m.from }}
                  <el-tag v-if="m.mentionId" size="small" type="warning">@我</el-tag>
                </div>
                <div class="chat-text">{{ m.text }}</div>
              </div>
            </div>
            <!-- @提及下拉 -->
            <div v-if="showMentionDropdown && mentionSuggestions.length > 0" class="mention-dropdown">
              <div v-for="p in mentionSuggestions" :key="p.userId" class="mention-item" @click="selectMention(p)">
                {{ p.name }}<el-tag v-if="p.isHost" size="small" type="warning" style="margin-left:4px">主持</el-tag>
              </div>
            </div>
            <div class="chat-input">
              <div class="emoji-wrap">
                <el-button circle @click="toggleEmojiPicker" title="表情">😀</el-button>
                <transition name="fade">
                  <div v-if="showEmojiPicker" class="emoji-picker">
                    <span v-for="e in emojis" :key="e" class="emoji-item" @click="insertEmoji(e)">{{ e }}</span>
                  </div>
                </transition>
              </div>
              <el-input v-model="chatText" placeholder="说点什么..." @keydown.enter="sendChat" @input="handleChatInput" />
              <el-button type="primary" @click="sendChat">发送</el-button>
            </div>
          </aside>
        </transition>
      </div>

      <!-- 空槽位提示 -->
      <div v-if="remotePeerIds.length === 0" class="waiting-hint">
        等待其他人加入...
      </div>
    </div>

    <!-- 会议室列表视图 -->
    <div v-else class="meeting-list">
      <div class="page-banner">
        <h2>视频会议</h2>
      </div>

      <div class="meeting-grid">
        <!-- 创建会议室 -->
        <el-card class="create-card" shadow="never">
          <template #header>
            <div class="card-header">
              <Plus style="width: 18px; height: 18px;" />
              <span>创建会议室</span>
            </div>
          </template>
          <div class="create-form">
            <el-input v-model="newRoomName" placeholder="会议室名称" />
            <el-button type="primary" style="width: 100%; margin-top: 12px;" @click="createRoom" :loading="loading">
              创建并加入
            </el-button>
          </div>
        </el-card>

        <!-- 加入会议室 -->
        <el-card class="join-card" shadow="never">
          <template #header>
            <div class="card-header">
              <UserFilled style="width: 18px; height: 18px;" />
              <span>加入会议</span>
            </div>
          </template>
          <div class="join-form">
            <el-input v-model="joinCode" placeholder="输入邀请码" maxlength="6" />
            <el-button type="success" style="width: 100%; margin-top: 12px;" @click="joinRoomByCode()" :loading="loading">
              加入
            </el-button>
          </div>
        </el-card>

        <!-- 我的会议室 -->
        <el-card class="rooms-card" shadow="never">
          <template #header>
            <div class="card-header">
              <el-icon><UserFilled /></el-icon>
              <span>我的会议室</span>
            </div>
          </template>
          <div class="rooms-list" v-if="myRooms.length > 0">
            <div v-for="room in myRooms" :key="room.id" class="room-item" @click="joinRoomByCode(room.roomCode)">
              <div class="room-info">
                <span class="room-name">{{ room.roomName }}</span>
                <span class="room-code" @click.stop="copyRoomCode(room.roomCode)">
                  <CopyDocument style="width: 12px; height: 12px;" />
                  {{ room.roomCode }}
                </span>
              </div>
              <el-tag :type="room.status === 1 ? 'success' : 'default'" size="small">
                {{ room.status === 1 ? '进行中' : room.status === 2 ? '已结束' : '未开始' }}
              </el-tag>
            </div>
          </div>
          <el-empty v-else description="暂无会议室" />
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.meeting-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 24px;
  min-height: calc(100vh - 64px);
  background: #f6f7f8;
}

/* 会议中 */
.meeting-active {
  height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
}
.meeting-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 20px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.header-left h2 {
  margin: 0;
  font-size: 20px;
}
.room-code {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  background: #e3e5e7;
  border-radius: 6px;
  font-size: 13px;
  color: #61666d;
  cursor: pointer;
}
.room-code:hover {
  background: #dcdfe6;
}

.video-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  align-content: start;
}
.meeting-body {
  flex: 1;
  display: flex;
  gap: 16px;
  overflow: hidden;
}
.meeting-body .video-grid {
  flex: 1;
  overflow-y: auto;
}
.video-status {
  position: absolute;
  top: 12px;
  left: 12px;
  display: flex;
  gap: 6px;
  padding: 4px 8px;
  background: rgba(0,0,0,0.5);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
}
.video-status .status-off { color: #ff5252; }
.video-status .status-on { color: #4fc3f7; }
.video-card.spotlighted {
  border: 3px solid #e6a23c;
  box-shadow: 0 0 20px rgba(230, 162, 60, 0.5);
}
.video-grid.has-spotlight {
  display: grid;
  grid-template-columns: 1fr 200px;
  gap: 16px;
}
.video-grid.has-spotlight .video-card.spotlighted {
  grid-column: 1;
  aspect-ratio: 16/9;
}
.video-grid.has-spotlight .video-card:not(.spotlighted) {
  grid-column: 2;
  aspect-ratio: 16/9;
  max-height: 160px;
  align-self: start;
}

/* 参与者侧栏 */
.participant-sidebar {
  width: 280px;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
}
.sidebar-header {
  padding: 14px 16px;
  font-weight: 600;
  border-bottom: 1px solid #e3e5e7;
  font-size: 14px;
  color: #18191c;
}
.participant-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}
.participant-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  gap: 8px;
  border-bottom: 1px solid #f6f7f8;
  flex-wrap: wrap;
}
.p-info {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}
.p-name {
  font-size: 13px;
  color: #18191c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.p-status {
  display: flex;
  gap: 6px;
  font-size: 14px;
}
.p-status .on { color: #67c23a; }
.p-status .off { color: #c0c4cc; }
.p-actions {
  width: 100%;
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}
.slide-x-enter-active, .slide-x-leave-active { transition: transform .25s ease, opacity .25s ease; }
.slide-x-enter-from, .slide-x-leave-to { transform: translateX(20px); opacity: 0; }

/* 聊天侧栏 */
.chat-sidebar {
  width: 300px;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.chat-empty {
  text-align: center;
  color: #9499a0;
  font-size: 13px;
  padding-top: 40px;
}
.chat-msg {
  background: #f6f7f8;
  padding: 8px 12px;
  border-radius: 8px;
  max-width: 80%;
  word-break: break-word;
}
.chat-msg.self {
  align-self: flex-end;
  background: #e6f4ff;
}
.chat-from {
  font-size: 11px;
  color: #9499a0;
  margin-bottom: 2px;
}
.chat-text {
  font-size: 13px;
  color: #18191c;
}
.chat-input {
  display: flex;
  gap: 6px;
  padding: 10px;
  border-top: 1px solid #e3e5e7;
}
.chat-input { display: flex; gap: 6px; padding: 10px; border-top: 1px solid #e3e5e7; align-items: center; }
.chat-input .el-input { flex: 1; }
.emoji-wrap { position: relative; }
.emoji-picker {
  position: absolute;
  bottom: 40px;
  left: 0;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 8px;
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
  width: 240px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 10;
}
.emoji-item { font-size: 20px; cursor: pointer; padding: 4px; border-radius: 4px; }
.emoji-item:hover { background: #f6f7f8; }
.fade-enter-active, .fade-leave-active { transition: opacity .2s; }
/* @提及 */
.mention-dropdown {
  border-top: 1px solid #e3e5e7;
  max-height: 120px;
  overflow-y: auto;
}
.mention-item {
  padding: 6px 12px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.mention-item:hover { background: #f6f7f8; }
.chat-msg.mentioned .chat-text { color: #e6a23c; font-weight: 500; }
.hand-count-badge {
  position: absolute;
  bottom: 16px;
  left: 16px;
  padding: 4px 12px;
  background: rgba(240, 64, 64, 0.9);
  color: #fff;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  z-index: 10;
}
/* 昵称编辑 */
.sidebar-header .my-name {
  font-size: 12px;
  color: #00a1d6;
  cursor: pointer;
  margin-left: 8px;
  font-weight: 400;
}
.sidebar-header .name-input {
  border: 1px solid #00a1d6;
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 12px;
  width: 100px;
  margin-left: 8px;
}
/* 连接质量指示 */
.quality-bars {
  display: inline-flex;
  align-items: flex-end;
  gap: 1px;
  height: 12px;
}
.quality-bars span {
  width: 3px;
  border-radius: 1px;
  background: #c0c4cc;
}
.quality-bars span:nth-child(1) { height: 4px; }
.quality-bars span:nth-child(2) { height: 7px; }
.quality-bars span:nth-child(3) { height: 10px; }
.quality-bars.q1 span { background: #f04040; }
.quality-bars.q2 span { background: #e6a23c; }
.quality-bars.q3 span { background: #67c23a; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.header-right {
  display: flex;
  gap: 8px;
  align-items: center;
}
.rec-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  background: #f04040;
  border-radius: 50%;
  margin-right: 4px;
  animation: rec-blink 1s infinite;
}
@keyframes rec-blink { 0%,100%{opacity:1} 50%{opacity:0.3} }

/* 等待室 */
.waiting-room {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60vh;
}
.waiting-card {
  text-align: center;
  padding: 48px;
  background: #fff;
  border-radius: 16px;
  max-width: 400px;
}
.waiting-icon { font-size: 64px; margin-bottom: 16px; }
.waiting-card h3 { margin: 0 0 12px; font-size: 20px; color: #18191c; }
.waiting-card p { margin: 0 0 24px; color: #9499a0; font-size: 14px; }

/* 待审批加入栏 */
.waiting-approve-bar {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}
.approve-title { font-size: 14px; font-weight: 600; margin-bottom: 12px; color: #18191c; }
.waiting-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  border-bottom: 1px solid #f6f7f8;
}
.waiting-item:last-child { border-bottom: none; }
.waiting-item span { flex: 1; font-size: 13px; }

.video-card {
  position: relative;
  background: #1a1a1a;
  border-radius: 12px;
  overflow: hidden;
  aspect-ratio: 16/9;
}
.video-card.local {
  border: 2px solid #00a1d6;
}
.video-element {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.video-label {
  position: absolute;
  bottom: 12px;
  left: 12px;
  padding: 4px 12px;
  background: rgba(0,0,0,0.6);
  color: #fff;
  border-radius: 4px;
  font-size: 13px;
}
.video-controls {
  position: absolute;
  bottom: 12px;
  right: 12px;
  display: flex;
  gap: 8px;
}

.waiting-hint {
  text-align: center;
  color: #9499a0;
  padding: 40px;
  font-size: 14px;
}

/* 会议室列表 */
.meeting-list .page-banner {
  margin-bottom: 28px;
}
.meeting-list .page-banner h2 {
  font-size: 24px;
  font-weight: 700;
  color: #18191c;
  margin: 0;
}
.meeting-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
@media (max-width: 768px) {
  .meeting-grid {
    grid-template-columns: 1fr;
  }
}
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #18191c;
}

.rooms-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.room-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f6f7f8;
  border-radius: 8px;
  cursor: pointer;
}
.room-item:hover {
  background: #e3e5e7;
}
.room-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.room-name {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}
.controls-hidden {
  opacity: 0 !important;
  transition: opacity 0.4s ease;
}
.video-card:hover .video-label,
.video-card:hover .video-controls,
.video-card:hover .video-status {
  opacity: 1;
}
.video-label.controls-hidden,
.video-controls.controls-hidden,
.video-status.controls-hidden {
  opacity: 0;
}
.video-status { transition: opacity 0.4s ease; }
.video-label { transition: opacity 0.4s ease; }
.video-controls { transition: opacity 0.4s ease; }

/* 屏幕共享主画面布局 */
.video-grid.has-sharer {
  display: grid;
  grid-template-columns: 1fr 220px;
  grid-template-rows: 1fr;
  gap: 16px;
}
.video-grid.has-sharer .video-card.main-sharer {
  grid-column: 1;
  grid-row: 1;
  aspect-ratio: 16/9;
  max-height: none;
}
.video-grid.has-sharer .video-card:not(.main-sharer) {
  grid-column: 2;
  grid-row: 1;
  aspect-ratio: 16/9;
  max-height: 200px;
  align-self: start;
}

/* 发言者视图布局（与屏幕共享互斥） */
.video-grid.has-speaker {
  display: grid;
  grid-template-columns: 1fr 200px;
  grid-template-rows: 1fr;
  gap: 16px;
}
.video-grid.has-speaker .video-card.active-speaker {
  grid-column: 1;
  grid-row: 1;
  aspect-ratio: 16/9;
}
.video-grid.has-speaker .video-card:not(.active-speaker) {
  grid-column: 2;
  grid-row: 1;
  aspect-ratio: 16/9;
  max-height: 180px;
  align-self: start;
}
</style>