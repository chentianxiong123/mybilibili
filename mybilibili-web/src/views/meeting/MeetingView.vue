<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Microphone, MicrophoneOff, Video, VideoOff, Users, CopyDocument, Plus, Monitor } from '@element-plus/icons-vue'
import { meetingApi } from '../../api/meeting.js'

const router = useRouter()
const loading = ref(false)
const myRooms = ref([])
const joinCode = ref('')
const createDialogVisible = ref(false)
const newRoomName = ref('')

const localStream = ref(null)
const localVideoRef = ref(null)
const participants = ref([])
let ws = null
let peerConnections = {}
const remoteStreams = ref({})

const audioEnabled = ref(true)
const videoEnabled = ref(true)
const currentRoom = ref(null)
const screenShareEnabled = ref(false)

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
      createDialogVisible.value = false
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
    ElMessage.error('加入失败')
  } finally {
    loading.value = false
  }
}

const startLocalStream = async () => {
  try {
    localStream.value = await navigator.mediaDevices.getUserMedia({
      video: true,
      audio: true
    })
    if (localVideoRef.value) {
      localVideoRef.value.srcObject = localStream.value
    }
  } catch (e) {
    ElMessage.error('无法获取摄像头/麦克风')
  }
}

const connectWebSocket = (roomCode) => {
  const userId = parseInt(document.cookie.match(/userId=(\d+)/)?.[1] || '0') || Date.now()
  ws = new WebSocket(`ws://${window.location.host}/ws/meeting`)

  ws.onopen = () => {
    ws.send(JSON.stringify({
      type: 'join',
      roomCode,
      userId,
      userName: '用户' + userId
    }))
  }

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data)
    handleSignaling(msg)
  }

  ws.onclose = () => {
    console.log('[Meeting] WebSocket 断开')
  }
}

const handleSignaling = async (msg) => {
  switch (msg.type) {
    case 'user-joined':
      await createPeerConnection(msg.userId, true)
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
    case 'toggle-audio':
    case 'toggle-video':
    case 'toggle-screen':
      break
  }
}

const createPeerConnection = async (peerId, isInitiator) => {
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
    remoteStreams.value[peerId] = event.streams[0]
    const videoEl = document.getElementById('video-' + peerId)
    if (videoEl) {
      videoEl.srcObject = event.streams[0]
    }
  }

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      sendSignaling('ice-candidate', peerId, event.candidate)
    }
  }

  if (isInitiator) {
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignaling('offer', peerId, offer.sdp)
  }
}

const handleOffer = async (peerId, sdp) => {
  await createPeerConnection(peerId, false)
  await peerConnections[peerId].setRemoteDescription({ type: 'offer', sdp })
  const answer = await peerConnections[peerId].createAnswer()
  await peerConnections[peerId].setLocalDescription(answer)
  sendSignaling('answer', peerId, answer.sdp)
}

const handleAnswer = async (peerId, sdp) => {
  if (peerConnections[peerId]) {
    await peerConnections[peerId].setRemoteDescription({ type: 'answer', sdp })
  }
}

const handleIceCandidate = async (peerId, candidate) => {
  if (peerConnections[peerId]) {
    await peerConnections[peerId].addIceCandidate(candidate)
  }
}

const sendSignaling = (type, peerId, data) => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    const userId = parseInt(document.cookie.match(/userId=(\d+)/)?.[1] || '0') || Date.now()
    ws.send(JSON.stringify({
      type,
      roomCode: currentRoom.value?.roomCode,
      userId,
      userName: '用户' + userId,
      data
    }))
  }
}

const toggleAudio = () => {
  if (localStream.value) {
    const audioTrack = localStream.value.getAudioTracks()[0]
    if (audioTrack) {
      audioTrack.enabled = !audioTrack.enabled
      audioEnabled.value = audioTrack.enabled
    }
  }
}

const toggleVideo = () => {
  if (localStream.value) {
    const videoTrack = localStream.value.getVideoTracks()[0]
    if (videoTrack) {
      videoTrack.enabled = !videoTrack.enabled
      videoEnabled.value = videoTrack.enabled
    }
  }
}

const startScreenShare = async () => {
  try {
    const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: false })
    const screenTrack = screenStream.getVideoTracks()[0]

    // 替换所有 peer connection 中的视频轨道
    Object.values(peerConnections).forEach(pc => {
      const senders = pc.getSenders()
      const videoSender = senders.find(s => s.track?.kind === 'video')
      if (videoSender) {
        videoSender.replaceTrack(screenTrack)
      }
    })

    screenShareEnabled.value = true
    screenTrack.onended = () => {
      stopScreenShare()
    }
  } catch (e) {
    ElMessage.error('屏幕共享失败')
  }
}

const stopScreenShare = () => {
  if (localStream.value) {
    const videoTrack = localStream.value.getVideoTracks()[0]
    Object.values(peerConnections).forEach(pc => {
      const senders = pc.getSenders()
      const videoSender = senders.find(s => s.track?.kind === 'video')
      if (videoSender && videoTrack) {
        videoSender.replaceTrack(videoTrack)
      }
    })
  }
  screenShareEnabled.value = false
}

const cleanup = () => {
  if (localStream.value) {
    localStream.value.getTracks().forEach(t => t.stop())
    localStream.value = null
  }
  Object.values(peerConnections).forEach(pc => pc.close())
  peerConnections = {}
  remoteStreams.value = {}
  if (ws) {
    ws.close()
    ws = null
  }
}

const copyRoomCode = (code) => {
  navigator.clipboard.writeText(code).then(() => {
    ElMessage.success('邀请码已复制')
  })
}

const leaveRoom = async () => {
  if (currentRoom.value) {
    await meetingApi.leaveRoom(currentRoom.value.id)
  }
  cleanup()
  currentRoom.value = null
  router.push('/meeting')
}

const remotePeerIds = computed(() => Object.keys(peerConnections))
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
          <video :id="'video-' + peerId" autoplay playsinline class="video-element" />
          <div class="video-label">参与者 {{ remotePeerIds.indexOf(peerId) + 1 }}</div>
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