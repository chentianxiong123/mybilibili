<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Microphone, MicrophoneOff, Video, VideoOff, Phone, PhoneFilled } from '@element-plus/icons-vue'
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
const isLinkmicVisible = ref(false)
const myLinkmicId = ref(null)  // 我的连麦ID
const activeLinkmics = ref([])
const pendingApplications = ref([])
const linkmicLoading = ref(false)

// 连麦者视频相关（用于主播端显示连麦者画面）
const linkmicStreams = ref({})
let linkmicWs = null

// 当前用户信息（从cookie模拟，实际应该从store获取）
const currentUserId = ref(parseInt(document.cookie.match(/userId=(\d+)/)?.[1] || '0') || 0)
const isStreamer = ref(false)

onMounted(async () => {
  try {
    console.log('[LiveRoom] 加载直播间, roomId:', roomId)
    const res = await liveApi.getRoom(roomId)
    console.log('[LiveRoom] API响应:', res)
    if (res.code === 200) {
      room.value = res.data
      isStreamer.value = currentUserId.value === room.value?.userId
      console.log('[LiveRoom] 房间数据:', res.data, 'streamKey:', res.data?.streamKey)
    } else {
      console.error('[LiveRoom] 非200响应:', res)
      ElMessage.error(res.message || '直播间不存在')
    }
  } catch (e) {
    console.error('[LiveRoom] 加载失败, 错误:', e)
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
    await nextTick()
    await new Promise(r => setTimeout(r, 100))
    const videoEl = document.querySelector('.live-video')
    console.log('[LiveRoom] video元素:', videoEl)
    if (videoEl) {
      initPlayer(videoEl)
    } else {
      console.error('[LiveRoom] video元素不存在')
    }
  }

  // 加载连麦状态
  await loadLinkmicStatus()
  connectLinkmicWebSocket()
})

onUnmounted(() => {
  if (flvPlayer) {
    flvPlayer.pause()
    flvPlayer.unload()
    flvPlayer.detachMediaElement()
  }
  disconnectLinkmic()
})

const initPlayer = (videoEl) => {
  if (!room.value) return
  const video = videoEl || videoRef.value
  if (!video) return
  const flvUrl = `http://${window.location.hostname}:28080/live/${room.value.streamKey}.flv`
  console.log('[LiveRoom] FLV地址:', flvUrl)

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

// ============ 连麦功能 ============
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
  const wsUrl = `ws://${window.location.host}/ws/meeting`
  linkmicWs = new WebSocket(wsUrl)

  linkmicWs.onopen = () => {
    console.log('[Linkmic WS] 连接成功')
  }

  linkmicWs.onmessage = async (event) => {
    const msg = JSON.parse(event.data)
    console.log('[Linkmic WS] 收到消息:', msg.type)

    switch (msg.type) {
      case 'linkmic-apply':
        if (isStreamer.value) {
          pendingApplications.value.push(msg.data)
        }
        break
      case 'linkmic-accepted':
        myLinkmicId.value = msg.data.linkmicId
        ElMessage.success('连麦已接通')
        break
      case 'linkmic-rejected':
        ElMessage.warning('连麦被拒绝')
        break
      case 'linkmic-disconnected':
        removeLinkmicStream(msg.data.viewerId)
        break
      case 'offer':
      case 'answer':
      case 'ice-candidate':
        await handleLinkmicSignaling(msg)
        break
    }
  }

  linkmicWs.onclose = () => {
    console.log('[Linkmic WS] 断开')
  }
}

// 观众申请连麦
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

      if (linkmicWs && linkmicWs.readyState === WebSocket.OPEN) {
        linkmicWs.send(JSON.stringify({
          type: 'linkmic-apply',
          roomCode: room.value?.streamKey,
          userId: currentUserId.value,
          data: res.data
        }))
      }
    } else {
      ElMessage.error(res.message || '申请失败')
    }
  } catch (e) {
    ElMessage.error('申请连麦失败')
  } finally {
    linkmicLoading.value = false
  }
}

// 主播同意连麦
const acceptLinkmic = async (linkmic) => {
  try {
    await linkmicApi.acceptLinkmic(linkmic.id)
    linkmic.status = 1

    if (linkmicWs && linkmicWs.readyState === WebSocket.OPEN) {
      linkmicWs.send(JSON.stringify({
        type: 'linkmic-accepted',
        roomCode: room.value?.streamKey,
        userId: linkmic.viewerId,
        data: { linkmicId: linkmic.id }
      }))
    }

    activeLinkmics.value.push(linkmic)
    pendingApplications.value = pendingApplications.value.filter(a => a.id !== linkmic.id)
    ElMessage.success('已同意连麦')
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

// 主播拒绝连麦
const rejectLinkmic = async (linkmic) => {
  try {
    await linkmicApi.rejectLinkmic(linkmic.id)
    pendingApplications.value = pendingApplications.value.filter(a => a.id !== linkmic.id)

    if (linkmicWs && linkmicWs.readyState === WebSocket.OPEN) {
      linkmicWs.send(JSON.stringify({
        type: 'linkmic-rejected',
        roomCode: room.value?.streamKey,
        userId: linkmic.viewerId,
        data: {}
      }))
    }
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

// 断开连麦
const disconnectLinkmic = async (linkmic) => {
  try {
    await linkmicApi.disconnectLinkmic(linkmic.id)
    removeLinkmicStream(linkmic.viewerId)
    activeLinkmics.value = activeLinkmics.value.filter(l => l.id !== linkmic.id)

    if (linkmicWs && linkmicWs.readyState === WebSocket.OPEN) {
      linkmicWs.send(JSON.stringify({
        type: 'linkmic-disconnected',
        roomCode: room.value?.streamKey,
        userId: linkmic.viewerId,
        data: { viewerId: linkmic.viewerId }
      }))
    }
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

// WebRTC 信令处理（简化版，实际需要更完整的逻辑）
let linkmicPeerConnections = {}

const handleLinkmicSignaling = async (msg) => {
  // 这里需要完整的 WebRTC 连接逻辑
  // 由于连麦涉及到主播和观众之间的 P2P 连接，需要更复杂的实现
  console.log('[Linkmic] 信令消息:', msg.type, msg.userId)
}

const removeLinkmicStream = (viewerId) => {
  if (linkmicPeerConnections[viewerId]) {
    linkmicPeerConnections[viewerId].close()
    delete linkmicPeerConnections[viewerId]
  }
  delete linkmicStreams.value[viewerId]
}

// ============ 弹幕功能 ============
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

            <!-- 连麦视频区域（主播端显示连麦者） -->
            <div v-if="activeLinkmics.length > 0" class="linkmic-videos">
              <div v-for="linkmic in activeLinkmics" :key="linkmic.id" class="linkmic-video-card">
                <video :id="'linkmic-' + linkmic.viewerId" autoplay playsinline />
                <div class="linkmic-name">{{ linkmic.viewerName }}</div>
                <div class="linkmic-actions">
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