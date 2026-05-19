<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Microphone, MicrophoneOff, Video, VideoOff, Users, CopyDocument, Plus, Monitor } from '@element-plus/icons-vue'
import { meetingApi } from '../../api/meeting.js'

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
let ws = null

onMounted(async () => {
  await loadMyRooms()
})

onUnmounted(() => {
  cleanup()
})

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

    const joinRes = await meetingApi.joinRoom(roomCode)
    if (joinRes.code !== 200) {
      ElMessage.error(joinRes.message || '加入失败')
      return
    }

    await startLocalStream()
    connectWebSocket(roomCode)
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
  ws = new WebSocket(`${proto}//${location.host}/ws/meeting`)

  ws.onopen = () => {
    ws.send(JSON.stringify({
      type: 'join',
      roomCode,
      userId: me.value.id,
      userName: me.value.name
    }))
  }

  ws.onmessage = async (event) => {
    let msg
    try { msg = JSON.parse(event.data) } catch { return }
    await handleSignaling(msg)
  }

  ws.onclose = () => console.log('[Meeting WS] 断开')
  ws.onerror = (e) => console.error('[Meeting WS] 错误', e)
}

const handleSignaling = async (msg) => {
  switch (msg.type) {
    case 'room-users':
      // 我是新人，对每个已有成员发起 offer
      for (const u of (msg.data || [])) {
        await createPeerConnection(u.userId, true)
      }
      break
    case 'user-joined':
      // 我是已有成员，等新人发 offer 过来即可（不主动发，避免双方都发 offer）
      console.log('[Meeting] 新成员加入', msg.userId)
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
  }
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
    remotePeers[peerId] = { stream: event.streams[0], name: '参与者' + peerId }
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

  if (isInitiator) {
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignaling('offer', peerId, offer.sdp)
  }
  return pc
}

const handleOffer = async (peerId, sdp) => {
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
  if (!ws || ws.readyState !== WebSocket.OPEN) return
  ws.send(JSON.stringify({
    type,
    roomCode: currentRoom.value?.roomCode,
    userId: me.value.id,
    userName: me.value.name,
    targetUserId,
    data
  }))
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

const toggleAudio = () => {
  if (!localStream.value) return
  const track = localStream.value.getAudioTracks()[0]
  if (track) {
    track.enabled = !track.enabled
    audioEnabled.value = track.enabled
  }
}

const toggleVideo = () => {
  if (!localStream.value) return
  const track = localStream.value.getVideoTracks()[0]
  if (track) {
    track.enabled = !track.enabled
    videoEnabled.value = track.enabled
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
}

const cleanup = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    try {
      ws.send(JSON.stringify({
        type: 'leave',
        roomCode: currentRoom.value?.roomCode,
        userId: me.value.id
      }))
    } catch (e) {}
  }
  if (ws) { ws.close(); ws = null }
  Object.keys(peerConnections).forEach(removePeerConnection)
  if (localStream.value) {
    localStream.value.getTracks().forEach(t => t.stop())
    localStream.value = null
  }
}

const copyRoomCode = (code) => {
  navigator.clipboard.writeText(code).then(() => {
    ElMessage.success('邀请码已复制')
  })
}

const leaveRoom = async () => {
  if (currentRoom.value) {
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
</script>

<template>
  <div class="meeting-page">
    <!-- 会议中视图 -->
    <div v-if="currentRoom" class="meeting-active">
      <div class="meeting-header">
        <div class="header-left">
          <h2>{{ currentRoom.roomName }}</h2>
          <span class="room-code" @click="copyRoomCode(currentRoom.roomCode)">
            <CopyDocument style="width: 12px; height: 12px;" />
            {{ currentRoom.roomCode }}
          </span>
        </div>
        <div class="header-right">
          <el-button type="danger" @click="leaveRoom">离开会议</el-button>
        </div>
      </div>

      <div class="video-grid">
        <!-- 本地视频 -->
        <div class="video-card local">
          <video ref="localVideoRef" autoplay muted playsinline class="video-element" />
          <div class="video-label">我</div>
          <div class="video-controls">
            <el-button :type="audioEnabled ? '' : 'danger'" circle @click="toggleAudio">
              <el-icon><Microphone v-if="audioEnabled" /><MicrophoneOff v-else /></el-icon>
            </el-button>
            <el-button :type="videoEnabled ? '' : 'danger'" circle @click="toggleVideo">
              <el-icon><Video v-if="videoEnabled" /><VideoOff v-else /></el-icon>
            </el-button>
            <el-button :type="screenShareEnabled ? 'primary' : ''" circle @click="startScreenShare">
              <el-icon><Monitor /></el-icon>
            </el-button>
          </div>
        </div>

        <!-- 远端视频 -->
        <div v-for="peerId in remotePeerIds" :key="peerId" class="video-card">
          <video :ref="el => setRemoteVideo(peerId, el)" autoplay playsinline class="video-element" />
          <div class="video-label">{{ remotePeers[peerId]?.name || ('参与者 ' + (remotePeerIds.indexOf(peerId) + 1)) }}</div>
        </div>
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
              <Users style="width: 18px; height: 18px;" />
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
              <el-icon><Users /></el-icon>
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
.room-code {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #9499a0;
}
</style>