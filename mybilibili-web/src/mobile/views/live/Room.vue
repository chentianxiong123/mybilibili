<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { getRoomInfo, getDanMuConfig } from '../../api/live'
import { formatTenThousand } from '../../utils/format'

const route = useRoute()
const roomId = route.params.roomId

const room = ref<any>(null)
const anchor = ref<any>(null)
const loading = ref(true)
const onlineNum = ref(0)
const danmakuList = ref([])
const danmakuText = ref('')
const isFollowing = ref(false)
const followerCount = ref(0)
const playerRef = ref(null)
let art = null

onMounted(async () => {
  try {
    const res = await getRoomInfo(roomId)
    if (res.code === '1') {
      room.value = res.data
      followerCount.value = res.data.followerCount || 0
      if (res.data.uId) {
        anchor.value = res.data.anchor || null
      }
      initPlayer()
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
    await initDanmaku()
  }
})

onUnmounted(() => {
  if (art) { try { art.destroy() } catch (e) {} }
})

const initPlayer = () => {
  // 使用已有的 LiveRoomView 的播放器逻辑
  // 这里用简单的 video 标签演示
  const container = playerRef.value
  if (!container || !room.value) return
  // 简单播放器 - 如果有流地址就用 video
  const video = document.createElement('video')
  video.style.cssText = 'width:100%;height:100%;object-fit:contain;background:#000;'
  video.controls = true
  video.playsInline = true
  video.muted = true
  video.autoplay = true

  // 如果有播放地址
  if (room.value.playUrl) {
    video.src = room.value.playUrl
  } else {
    // 显示封面
    video.poster = room.value.cover || ''
  }

  container.appendChild(video)
}

const initDanmaku = async () => {
  try {
    const res = await getDanMuConfig(roomId)
    if (res.code === '1') {
      const { host, port } = res.data
      const wsUrl = `wss://${host}/sub`
      connectBarrageWS(wsUrl)
    }
  } catch (e) {
    console.error('弹幕连接失败:', e)
  }
}

const connectBarrageWS = (url) => {
  try {
    const ws = new WebSocket(url)
    ws.onopen = () => {
      ws.send(JSON.stringify({ roomId }))
    }
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.cmd === 'DANMU_MSG') {
          const text = data.info?.[1]
          const color = '#' + Number(data.info?.[0]?.[3] || 0).toString(16).padStart(6, '0')
          if (text) danmakuList.value.push({ text, color, from: '弹幕' })
          if (danmakuList.value.length > 100) danmakuList.value.shift()
        }
        if (data.online) {
          onlineNum.value = data.online
        }
      } catch (e) {}
    }
    ws.onerror = () => {}
    ws.onclose = () => {
      setTimeout(() => connectBarrageWS(url), 3000)
    }
  } catch (e) {}
}

const toggleFollow = () => {
  isFollowing.value = !isFollowing.value
  followerCount.value += isFollowing.value ? 1 : -1
}

const sendDanmaku = () => {
  const text = danmakuText.value.trim()
  if (!text) return
  danmakuList.value.push({ text, color: '#fff', from: '我' })
  danmakuText.value = ''
}
</script>

<template>
  <div class="live-room-page">
    <template v-if="loading">
      <div class="loading">加载中...</div>
    </template>
    <template v-else-if="room">
      <div class="room-layout">
        <div class="video-section">
          <div class="player-container" ref="playerRef"></div>
          <div class="video-info">
            <div class="room-title">{{ room.title }}</div>
            <div class="room-meta">
              <span class="live-status">直播中</span>
              <span>{{ formatTenThousand(onlineNum || room.onlineNum) }}人气</span>
              <span>{{ formatTenThousand(followerCount) }}粉丝</span>
              <span
                :class="['follow-btn', { following: isFollowing }]"
                @click="toggleFollow"
              >
                {{ isFollowing ? '已关注' : '+关注' }}
              </span>
            </div>
          </div>
        </div>
        <div class="chat-section">
          <div class="chat-header">
            <span>互动聊天</span>
          </div>
          <div class="chat-messages">
            <div v-for="(d, i) in danmakuList" :key="i" class="chat-msg">
              <span class="chat-from">{{ d.from }}: </span>
              <span>{{ d.text }}</span>
            </div>
          </div>
          <div class="chat-input">
            <input
              v-model="danmakuText"
              placeholder="发个弹幕~"
              @keydown.enter="sendDanmaku"
            />
            <button @click="sendDanmaku">发送</button>
          </div>
        </div>
      </div>
    </template>
    <div v-else class="empty">直播间不存在</div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.live-room-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0a0a0a;
}

.room-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.video-section {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.player-container {
  flex: 1;
  background: #000;
}

.video-info {
  background: rgba(0,0,0,0.8);
  padding: 10px 12px;
  color: #fff;
}

.room-title { font-size: 15px; font-weight: 600; margin-bottom: 6px; }

.room-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: #ccc;
}

.live-status {
  background: $theme-pink;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.follow-btn {
  padding: 2px 8px;
  border-radius: 10px;
  background: $theme-pink;
  font-size: 11px;
  cursor: pointer;

  &.following { background: #666; }
}

.chat-section {
  width: 280px;
  background: #fff;
  display: flex;
  flex-direction: column;
  border-left: 1px solid $border-color;
}

.chat-header {
  padding: 12px;
  font-weight: 600;
  font-size: 14px;
  border-bottom: 1px solid $border-color;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.chat-msg {
  margin-bottom: 6px;
  font-size: 13px;
  line-height: 1.4;
}

.chat-from { color: #00a1d6; }

.chat-input {
  display: flex;
  gap: 8px;
  padding: 10px;
  border-top: 1px solid $border-color;

  input {
    flex: 1;
    border: 1px solid $border-color;
    border-radius: 14px;
    padding: 4px 12px;
    font-size: 13px;
    outline: none;
  }

  button {
    background: $theme-pink;
    color: #fff;
    border: none;
    border-radius: 14px;
    padding: 4px 12px;
    font-size: 13px;
    cursor: pointer;
  }
}

.loading, .empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
}
</style>