<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import flvJs from 'flv.js'
import { liveApi } from '../../api/live.js'

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

onMounted(async () => {
  try {
    console.log('[LiveRoom] 加载直播间, roomId:', roomId)
    const res = await liveApi.getRoom(roomId)
    console.log('[LiveRoom] API响应:', res)
    if (res.code === 200) {
      room.value = res.data
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
    // 等 DOM 更新（loading=false 后 video 才出现）
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
})

onUnmounted(() => {
  if (flvPlayer) {
    flvPlayer.pause()
    flvPlayer.unload()
    flvPlayer.detachMediaElement()
  }
})

const initPlayer = (videoEl) => {
  if (!room.value) {
    console.log('[LiveRoom] initPlayer: room为空，跳过')
    return
  }
  const video = videoEl || videoRef.value
  if (!video) {
    console.error('[LiveRoom] initPlayer: video元素为空')
    return
  }
  // FLV 播放地址
  const flvUrl = `http://${window.location.hostname}:28080/live/${room.value.streamKey}.flv`
  console.log('[LiveRoom] FLV地址:', flvUrl)
  console.log('[LiveRoom] flv.js支持:', flvJs.isSupported())

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
    console.log('[LiveRoom] 播放器已创建并开始播放')
  } else {
    console.error('[LiveRoom] 浏览器不支持flv.js')
    ElMessage.warning('您的浏览器不支持FLV播放，请使用Chrome')
  }
}

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