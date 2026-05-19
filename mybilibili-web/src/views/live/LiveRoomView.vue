<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Phone, PhoneFilled } from '@element-plus/icons-vue'
import flvJs from 'flv.js'
import { liveApi } from '../../api/live.js'
import { linkmicApi } from '../../api/linkmic.js'

const route = useRoute()
const roomId = route.params.roomId
const room = ref(null)
const videoRef = ref(null)
const loading = ref(true)
const isBuffering = ref(false)
let bufferingTimer = null

const showBuffering = () => {
  clearTimeout(bufferingTimer)
  isBuffering.value = true
}
const hideBuffering = () => {
  bufferingTimer = setTimeout(() => { isBuffering.value = false }, 500)
}
const danmakuList = ref([])
const danmakuText = ref('')
let flvPlayer = null

const danmakuColors = ['#fff', '#ff6b81', '#fad34a', '#4fc3f7', '#a78bfa', '#34d399']

// 连麦相关
const myLinkmicId = ref(null)
const activeLinkmics = ref([])
const pendingApplications = ref([])
const linkmicLoading = ref(false)

// peerId(=对端 userId) -> { stream, name }
const linkmicStreams = reactive({})
const linkmicPeerConnections = {}
const pendingIce = {}
let linkmicWs = null

// 当前用户：从 localStorage（登录态可靠）
const getCurrentUser = () => {
  try {
    const u = JSON.parse(localStorage.getItem('user') || '{}')
    if (u && u.id) return { id: Number(u.id), name: u.username || u.nickname || ('用户' + u.id) }
  } catch (e) {}
  return { id: 0, name: '游客' }
}
const me = ref(getCurrentUser())
const currentUserId = ref(me.value.id)
const isStreamer = ref(false)

// 观众端的本地媒体流（连麦后产生）
const localLinkmicStream = ref(null)
const localLinkmicVideoRef = ref(null)

onMounted(async () => {
  try {
    const res = await liveApi.getRoom(roomId)
    if (res.code === 200) {
      room.value = res.data
      isStreamer.value = currentUserId.value === room.value?.userId
    } else {
      ElMessage.error(res.message || '直播间不存在')
    }
  } catch (e) {
    console.error('[LiveRoom] 加载失败:', e)
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
    await nextTick()
    await new Promise(r => setTimeout(r, 100))
    const videoEl = document.querySelector('.live-video')
    if (videoEl) initPlayer(videoEl)
  }

  await loadLinkmicStatus()
  connectLinkmicWebSocket()
})

onUnmounted(() => {
  if (flvPlayer) {
    flvPlayer.pause()
    flvPlayer.unload()
    flvPlayer.detachMediaElement()
  }
  cleanupLinkmic()
})

const initPlayer = (videoEl) => {
  if (!room.value) return
  const video = videoEl || videoRef.value
  if (!video) return
  const flvUrl = `http://${window.location.hostname}:28080/live/${room.value.streamKey}.flv`

  if (flvJs.isSupported()) {
    flvPlayer = flvJs.createPlayer({
      type: 'flv',
      url: flvUrl,
      isLive: true,
      config: {
        enableWorker: true,
        enableStashBuffer: false,
        stashInitialSize: 128,
        autoCleanupSourceBuffer: true,
        autoCleanupMaxDuration: 3,
        autoCleanupMinRecoverDuration: 2,
      }
    })
    flvPlayer.attachMediaElement(video)
    flvPlayer.load()
    flvPlayer.play()
    flvPlayer.on(flvJs.Events.Error, (errType, errDetail, errInfo) => {
      console.error('[LiveRoom] 播放器错误:', errType, errDetail, errInfo)
    })
  }
}

// ============ 连麦核心 ============
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
    console.error('加载连麦状态失败:', e)
  }
}

const connectLinkmicWebSocket = () => {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  linkmicWs = new WebSocket(`${proto}//${location.host}/ws/meeting`)

  linkmicWs.onopen = () => {
    // 用 streamKey 作为 roomCode 加入房间，便于主播/观众间路由
    if (room.value?.streamKey && currentUserId.value) {
      linkmicWs.send(JSON.stringify({
        type: 'join',
        roomCode: room.value.streamKey,
        userId: currentUserId.value,
        userName: me.value.name
      }))
    }
  }

  linkmicWs.onmessage = async (event) => {
    let msg
    try { msg = JSON.parse(event.data) } catch { return }

    switch (msg.type) {
      case 'linkmic-apply':
        if (isStreamer.value && msg.data) {
          // 避免重复
          if (!pendingApplications.value.find(a => a.id === msg.data.id)) {
            pendingApplications.value.push(msg.data)
          }
        }
        break
      case 'linkmic-accepted':
        // 观众收到：表示主播同意了，开始建立 P2P
        myLinkmicId.value = msg.data?.linkmicId
        ElMessage.success('连麦已接通，正在建立连接...')
        await viewerStartConnection(msg.userId /* 主播 userId */)
        break
      case 'linkmic-rejected':
        ElMessage.warning('连麦被拒绝')
        myLinkmicId.value = null
        break
      case 'linkmic-disconnected':
        removeLinkmicStream(msg.data?.viewerId || msg.userId)
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

  linkmicWs.onclose = () => console.log('[Linkmic WS] 断开')
  linkmicWs.onerror = (e) => console.error('[Linkmic WS]', e)
}

// ============ WebRTC P2P ============
const createPC = (peerId) => {
  if (linkmicPeerConnections[peerId]) return linkmicPeerConnections[peerId]
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  })
  linkmicPeerConnections[peerId] = pc

  pc.ontrack = (event) => {
    linkmicStreams[peerId] = { stream: event.streams[0] }
  }

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      sendSignaling('ice-candidate', peerId, event.candidate)
    }
  }

  pc.onconnectionstatechange = () => {
    console.log('[Linkmic] PC状态', peerId, pc.connectionState)
  }

  return pc
}

// 观众侧：主播同意后，由观众发起 offer（观众有摄像头要推给主播）
const viewerStartConnection = async (streamerId) => {
  try {
    if (!localLinkmicStream.value) {
      localLinkmicStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      await nextTick()
      if (localLinkmicVideoRef.value) {
        localLinkmicVideoRef.value.srcObject = localLinkmicStream.value
      }
    }
    const pc = createPC(streamerId)
    localLinkmicStream.value.getTracks().forEach(t => pc.addTrack(t, localLinkmicStream.value))
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignaling('offer', streamerId, offer.sdp)
  } catch (e) {
    console.error('[Linkmic] viewer start failed', e)
    ElMessage.error('无法获取摄像头/麦克风')
  }
}

const handleOffer = async (peerId, sdp) => {
  // 主播侧收到观众的 offer
  const pc = createPC(peerId)
  await pc.setRemoteDescription({ type: 'offer', sdp })
  if (pendingIce[peerId]) {
    for (const c of pendingIce[peerId]) await pc.addIceCandidate(c)
    delete pendingIce[peerId]
  }
  const answer = await pc.createAnswer()
  await pc.setLocalDescription(answer)
  sendSignaling('answer', peerId, answer.sdp)
}

const handleAnswer = async (peerId, sdp) => {
  const pc = linkmicPeerConnections[peerId]
  if (!pc) return
  await pc.setRemoteDescription({ type: 'answer', sdp })
  if (pendingIce[peerId]) {
    for (const c of pendingIce[peerId]) await pc.addIceCandidate(c)
    delete pendingIce[peerId]
  }
}

const handleIceCandidate = async (peerId, candidate) => {
  const pc = linkmicPeerConnections[peerId]
  if (pc && pc.remoteDescription) {
    try { await pc.addIceCandidate(candidate) } catch (e) { console.warn(e) }
  } else {
    (pendingIce[peerId] = pendingIce[peerId] || []).push(candidate)
  }
}

const sendSignaling = (type, targetUserId, data) => {
  if (!linkmicWs || linkmicWs.readyState !== WebSocket.OPEN) return
  linkmicWs.send(JSON.stringify({
    type,
    roomCode: room.value?.streamKey,
    userId: currentUserId.value,
    userName: me.value.name,
    targetUserId,
    data
  }))
}

const sendBroadcast = (type, data) => {
  if (!linkmicWs || linkmicWs.readyState !== WebSocket.OPEN) return
  linkmicWs.send(JSON.stringify({
    type,
    roomCode: room.value?.streamKey,
    userId: currentUserId.value,
    userName: me.value.name,
    data
  }))
}

// ============ 业务动作 ============
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
      ElMessage.success('连麦申请已发送，请等待主播同意')
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
    activeLinkmics.value.push(linkmic)
    pendingApplications.value = pendingApplications.value.filter(a => a.id !== linkmic.id)
    // 定向告诉这个观众："你被同意了"，观众收到后会发 offer
    sendSignaling('linkmic-accepted', linkmic.viewerId, { linkmicId: linkmic.id })
    ElMessage.success('已同意连麦')
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const rejectLinkmic = async (linkmic) => {
  try {
    await linkmicApi.rejectLinkmic(linkmic.id)
    pendingApplications.value = pendingApplications.value.filter(a => a.id !== linkmic.id)
    sendSignaling('linkmic-rejected', linkmic.viewerId, {})
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const disconnectLinkmic = async (linkmic) => {
  try {
    await linkmicApi.disconnectLinkmic(linkmic.id)
    removeLinkmicStream(linkmic.viewerId)
    activeLinkmics.value = activeLinkmics.value.filter(l => l.id !== linkmic.id)
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
  }
  myLinkmicId.value = null
  if (localLinkmicStream.value) {
    localLinkmicStream.value.getTracks().forEach(t => t.stop())
    localLinkmicStream.value = null
  }
  // 关掉跟主播的 PC
  if (room.value?.userId) {
    removePC(room.value.userId)
  }
  ElMessage.info('已断开连麦')
}

const removePC = (peerId) => {
  const pc = linkmicPeerConnections[peerId]
  if (pc) { pc.close(); delete linkmicPeerConnections[peerId] }
  delete linkmicStreams[peerId]
  delete pendingIce[peerId]
}

const removeLinkmicStream = (viewerId) => {
  removePC(viewerId)
}

const cleanupLinkmic = () => {
  Object.keys(linkmicPeerConnections).forEach(removePC)
  if (localLinkmicStream.value) {
    localLinkmicStream.value.getTracks().forEach(t => t.stop())
    localLinkmicStream.value = null
  }
  if (linkmicWs) {
    try { linkmicWs.close() } catch (e) {}
    linkmicWs = null
  }
}

const setRemoteVideo = (peerId, el) => {
  if (!el) return
  const peer = linkmicStreams[peerId]
  if (peer && peer.stream && el.srcObject !== peer.stream) {
    el.srcObject = peer.stream
  }
}

// ============ 弹幕 ============
const sendDanmaku = () => {
  const text = danmakuText.value.trim()
  if (!text) return
  danmakuList.value.push({
    id: Date.now(),
    text,
    color: danmakuColors[Math.floor(Math.random() * danmakuColors.length)]
  })
  danmakuText.value = ''
}
</script>

<template>
  <div class="live-room-page">
    <template v-if="loading">
      <div class="loading-state">
        <span class="buffering-icon" />
        <span style="color:#fff;font-size:14px;">加载中...</span>
      </div>
    </template>
    <template v-else-if="room">
      <div class="room-layout">
        <div class="video-section">
          <div class="video-wrapper" :class="{ buffering: isBuffering }">
            <video ref="videoRef" autoplay muted controls class="live-video" @waiting="showBuffering" @playing="hideBuffering"></video>
            <transition name="fade">
              <div v-if="isBuffering" class="buffering-overlay">
                <span class="buffering-icon" />
                <span>缓冲中...</span>
              </div>
            </transition>
            <!-- 弹幕层 -->
            <div class="danmaku-overlay">
              <div
                v-for="d in danmakuList"
                :key="d.id"
                class="danmaku-item"
                :style="{ color: d.color }"
              >
                {{ d.text }}
              </div>
            </div>
            <div class="video-info">
              <div class="room-title">{{ room.roomName }}</div>
            <div class="room-meta">
                <span class="live-status-dot" />
                <span>直播中</span>
                <span class="viewer">{{ room.viewerCount || 0 }} 人观看</span>
              </div>
            </div>

            <!-- 连麦视频区域（显示其他连麦者） -->
            <div v-if="activeLinkmics.length > 0 || (myLinkmicId && !isStreamer)" class="linkmic-videos">
              <!-- 观众端本地预览 -->
              <div v-if="!isStreamer && myLinkmicId" class="linkmic-video-card">
                <video ref="localLinkmicVideoRef" autoplay playsinline muted />
                <div class="linkmic-name">我（连麦中）</div>
              </div>
              <!-- 远端连麦者画面 -->
              <div v-for="linkmic in activeLinkmics" :key="linkmic.id" class="linkmic-video-card">
                <video :ref="el => setRemoteVideo(linkmic.viewerId, el)" autoplay playsinline />
                <div class="linkmic-name">{{ linkmic.viewerName }}</div>
                <div v-if="isStreamer" class="linkmic-actions">
                  <el-button size="small" type="danger" @click="disconnectLinkmic(linkmic)">
                    断开
                  </el-button>
                </div>
              </div>
            </div>
          </div>

          <!-- 观众端连麦按钮 -->
          <div v-if="!isStreamer && !myLinkmicId" class="linkmic-apply-bar">
            <el-button type="primary" size="large" round @click="applyLinkmic" :loading="linkmicLoading">
              <el-icon><Phone /></el-icon>
              申请连麦
            </el-button>
          </div>
          <div v-else-if="!isStreamer && myLinkmicId" class="linkmic-apply-bar">
            <el-button type="danger" size="large" round @click="viewerHangup">
              <el-icon><PhoneFilled /></el-icon>
              挂断连麦
            </el-button>
          </div>

          <!-- 主播端连麦管理 -->
          <div v-if="isStreamer && pendingApplications.length > 0" class="linkmic-pending-bar">
            <div class="pending-title">连麦申请 ({{ pendingApplications.length }})</div>
            <div class="pending-list">
              <div v-for="app in pendingApplications" :key="app.id" class="pending-item">
                <span>{{ app.viewerName }}</span>
                <el-button size="small" type="success" @click="acceptLinkmic(app)">同意</el-button>
                <el-button size="small" type="danger" @click="rejectLinkmic(app)">拒绝</el-button>
              </div>
            </div>
          </div>
        </div>
        <div class="chat-section">
          <div class="chat-header">互动聊天</div>
          <div class="chat-messages">
            <div v-for="d in danmakuList" :key="d.id" class="chat-msg">
              <span class="chat-name">观众: </span>
              <span>{{ d.text }}</span>
            </div>
          </div>
          <div class="chat-input">
            <el-input
              v-model="danmakuText"
              placeholder="发个弹幕~"
              @keydown.enter="sendDanmaku"
            />
            <el-button type="primary" @click="sendDanmaku">发送</el-button>
          </div>
        </div>
      </div>
    </template>
    <div v-else class="empty-state">
      <el-empty description="直播间不存在" />
    </div>
  </div>
</template>

<style scoped>
.live-room-page { height: calc(100vh - 64px); background: #0a0a0a; }
.room-layout { display: flex; height: 100%; }
.video-section { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.video-wrapper { flex: 1; position: relative; background: #000; }
.video-wrapper.buffering .live-video { opacity: 0.5; }
.buffering-overlay {
  position: absolute; inset: 0;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  color: #fff; gap: 12px; font-size: 14px;
  pointer-events: none;
}
.buffering-icon {
  width: 36px; height: 36px; border: 3px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
@keyframes spin { to { transform: rotate(360deg); } }
.live-video { width: 100%; height: 100%; object-fit: contain; }
.danmaku-overlay {
  position: absolute; top: 0; left: 0; right: 0; bottom: 80px;
  pointer-events: none; overflow: hidden; padding: 20px;
}
.danmaku-item {
  font-size: 16px; font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
  margin-bottom: 8px; animation: danmakuFade 5s linear forwards;
  white-space: nowrap;
}
@keyframes danmakuFade {
  0% { transform: translateX(100%); opacity: 0; }
  5% { opacity: 1; }
  90% { opacity: 1; }
  100% { transform: translateX(-100%); opacity: 0; }
}
.video-info {
  position: absolute; bottom: 0; left: 0; right: 0;
  background: linear-gradient(transparent, rgba(0,0,0,0.8));
  padding: 40px 20px 16px;
}
.room-title { font-size: 18px; font-weight: 600; color: #fff; margin-bottom: 8px; }
.room-meta { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #ccc; }
.live-status-dot { width: 8px; height: 8px; background: #f04040; border-radius: 50%; animation: pulse 1.5s infinite; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.viewer { color: #9499a0; }

/* 连麦视频 */
.linkmic-videos {
  position: absolute;
  top: 60px;
  right: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  z-index: 10;
}
.linkmic-video-card {
  width: 160px;
  height: 120px;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}
.linkmic-video-card video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.linkmic-name {
  position: absolute;
  bottom: 4px;
  left: 8px;
  font-size: 12px;
  color: #fff;
}
.linkmic-actions {
  position: absolute;
  top: 4px;
  right: 4px;
}

/* 连麦申请栏 */
.linkmic-apply-bar {
  position: absolute;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.linkmic-pending-bar {
  background: rgba(0,0,0,0.8);
  padding: 12px 16px;
  border-top: 1px solid #333;
}
.pending-title {
  color: #fff;
  font-size: 14px;
  margin-bottom: 8px;
}
.pending-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pending-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 13px;
}

.chat-section { width: 320px; background: #fff; display: flex; flex-direction: column; border-left: 1px solid #e3e5e7; }
.chat-header { padding: 16px; font-weight: 600; border-bottom: 1px solid #e3e5e7; }
.chat-messages { flex: 1; overflow-y: auto; padding: 12px; }
.chat-msg { margin-bottom: 8px; font-size: 14px; }
.chat-name { color: #00a1d6; }
.chat-input { display: flex; gap: 8px; padding: 12px; border-top: 1px solid #e3e5e7; }
.chat-input .el-input { flex: 1; }
.loading-state { display: flex; align-items: center; justify-content: center; height: 60vh; }
.empty-state { display: flex; align-items: center; justify-content: center; height: 60vh; background: #0a0a0a; }
</style>