<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { VideoCamera, CopyDocument, Cellphone, UserFilled, ChatDotRound, Monitor, Mute, Close, Picture, Upload } from '@element-plus/icons-vue'
import { liveApi } from '../../api/live.js'
import { createReconnectingWS } from '../../utils/reconnectingWs.js'
import { isNotificationEnabled, requestNotificationPermission } from '../../utils/notification.js'

const room = ref(null)
const loading = ref(true)
const rtmpUrl = ref('')
const streamKey = ref('')
const switching = ref(false)
const countdown = ref(0)
let countdownTimer = null

// WebRTC 推流
const sourceType = ref('camera')      // 'camera' | 'screen'
const webrtcStream = ref(null)
const isWebRtcPushing = ref(false)
const videoPreviewRef = ref(null)
let pushPc = null

// 在线观众
const viewers = ref([])             // [{ userId, userName, joinAt }]
const recentChats = ref([])         // 主播看到的最近弹幕 [{ id, from, text }]
const danmakuRate = ref(0)          // 弹幕速率（条/分钟）
const peakViewers = ref(0)
const totalDanmaku = ref(0)
const viewerDanmakuCount = reactive({})  // { [userId]: count }
const leaderboard = computed(() => {
  return Object.entries(viewerDanmakuCount)
    .map(([userId, count]) => {
      const v = viewers.value.find(u => String(u.userId) === String(userId))
      return {
        userId: Number(userId),
        userName: v?.userName || ('用户' + userId),
        count
      }
    })
    .filter(x => x.userId !== me.id) // 排除主播
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

// 观众等级
const viewerLevels = [
  { min: 0,   label: '新人',   color: '#909399', icon: '🌱' },
  { min: 11,  label: '常客',   color: '#67c23a', icon: '🌿' },
  { min: 51,  label: '活跃',   color: '#00a1d6', icon: '🌳' },
  { min: 201, label: '忠实',   color: '#e6a23c', icon: '🏅' },
  { min: 501, label: '贵族',   color: '#f56c6c', icon: '👑' },
]
const getViewerLevel = (userId) => {
  const count = viewerDanmakuCount[String(userId)] || 0
  for (let i = viewerLevels.length - 1; i >= 0; i--) {
    if (count >= viewerLevels[i].min) return viewerLevels[i]
  }
  return viewerLevels[0]
}
let ws = null
let metricsTimer = null

const hostname = window.location.hostname
const isLive = computed(() => room.value?.status === 'live')

const me = (() => {
  try {
    const u = JSON.parse(localStorage.getItem('user') || '{}')
    if (u && u.id) return { id: Number(u.id), name: u.username || u.nickname || ('用户' + u.id) }
  } catch (e) {}
  return { id: 0, name: '主播' }
})()

onMounted(async () => {
  requestNotificationPermission()
  try {
    let res = await liveApi.getMyRoom()
    if (!res || !res.data) {
      res = await liveApi.createRoom('我的直播间')
    }
    if (res && res.code === 200 && res.data) {
      room.value = res.data
      streamKey.value = res.data.streamKey
      rtmpUrl.value = `rtmp://${hostname}/live`
      selectedCategory.value = res.data.category || ''
      connectWs()
      // 定时开播提醒
      if (res.data.scheduledAt) {
        scheduleReminder(res.data.scheduledAt)
      }
      metricsTimer = setInterval(() => {
        const cutoff = Date.now() - 60000
        danmakuRate.value = recentChats.value.filter(m => m.id > cutoff).length
        if (viewers.value.length > peakViewers.value) peakViewers.value = viewers.value.length
        totalDanmaku.value = recentChats.value.length
      }, 1000)
    }
  } catch (e) {
    ElMessage.error('获取直播间信息失败')
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (ws) { ws.close(); ws = null }
  if (metricsTimer) { clearInterval(metricsTimer); metricsTimer = null }
})

const connectWs = () => {
  if (!streamKey.value) return
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = createReconnectingWS({
    url: `${proto}//${location.host}/ws/meeting`,
    onOpen: () => {
      ws.send({
        type: 'join',
        roomCode: streamKey.value,
        userId: me.id,
        userName: me.name + '（主播）'
      })
    },
    onMessage: (msg) => {
      switch (msg.type) {
        case 'room-users':
          viewers.value = (msg.data || []).map(u => ({
            userId: u.userId,
            userName: u.userName || ('用户' + u.userId),
            joinAt: Date.now()
          }))
          break
        case 'user-joined':
          if (msg.userId && msg.userId !== me.id) {
            if (!viewers.value.find(v => v.userId === msg.userId)) {
              viewers.value.push({
                userId: msg.userId,
                userName: msg.userName || ('用户' + msg.userId),
                joinAt: Date.now()
              })
            }
          }
          break
        case 'user-left':
          viewers.value = viewers.value.filter(v => v.userId !== msg.userId)
          break
        case 'chat':
          if (msg.data?.text) {
            recentChats.value.push({
              id: Date.now() + Math.random(),
              from: msg.userName || ('用户' + msg.userId),
              text: msg.data.text
            })
            if (recentChats.value.length > 80) recentChats.value.splice(0, 30)
            // 统计弹幕数
            if (msg.userId) {
              const uid = String(msg.userId)
              viewerDanmakuCount[uid] = (viewerDanmakuCount[uid] || 0) + 1
            }
          }
          break
        case 'viewer-count':
          // 后端会顺便推这个，但我们已经有 viewers.length 更精确，忽略
          break
      }
    }
  })
}

const formatJoinTime = (ts) => {
  const d = Math.floor((Date.now() - ts) / 1000)
  if (d < 60) return d + ' 秒前'
  if (d < 3600) return Math.floor(d / 60) + ' 分钟前'
  return Math.floor(d / 3600) + ' 小时前'
}

const copyText = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制')
  })
}

const isEditingRoomName = ref(false)
const editRoomNameInput = ref(null)
const selectedCategory = ref('')
const categories = ['娱乐', '游戏', '学习', '购物', '赛事', '其他']
const startEditRoomName = () => {
  isEditingRoomName.value = true
  nextTick(() => editRoomNameInput.value?.select())
}
const saveRoomName = async () => {
  if (!editRoomNameInput.value) return
  const newName = editRoomNameInput.value.trim()
  if (!newName || newName === room.value.roomName) {
    isEditingRoomName.value = false
    return
  }
  try {
    const res = await liveApi.updateRoom(room.value.id, { roomName: newName })
    if (res.code === 200) {
      room.value.roomName = newName
      ElMessage.success('直播间名称已更新')
    } else {
      ElMessage.error(res.message || '更新失败')
    }
  } catch (e) {
    ElMessage.error('更新失败')
  }
  isEditingRoomName.value = false
}

const saveCategory = async () => {
  if (!room.value) return
  try {
    const res = await liveApi.updateRoom(room.value.id, { category: selectedCategory.value })
    if (res.code === 200) {
      room.value.category = selectedCategory.value
      ElMessage.success('分类已更新')
    } else {
      ElMessage.error(res.message || '更新失败')
    }
  } catch (e) {
    ElMessage.error('更新失败')
  }
}

const showScheduleDialog = ref(false)
const scheduleDate = ref('')
const scheduleTime = ref('')

const openScheduleDialog = async () => {
  if (room.value?.scheduledAt) {
    const d = new Date(room.value.scheduledAt)
    scheduleDate.value = d.toISOString().slice(0, 10)
    scheduleTime.value = d.toTimeSlice().slice(0, 5)
  } else {
    const n = new Date(Date.now() + 3600000)
    scheduleDate.value = n.toISOString().slice(0, 10)
    scheduleTime.value = n.toTimeSlice().slice(0, 5)
  }
  showScheduleDialog.value = true
}

const saveSchedule = async () => {
  if (!room.value || !scheduleDate.value || !scheduleTime.value) return
  const dt = new Date(`${scheduleDate.value}T${scheduleTime.value}`)
  if (dt <= new Date()) {
    ElMessage.warning('定时时间必须大于当前时间')
    return
  }
  try {
    const res = await liveApi.scheduleRoom(room.value.id, dt.getTime())
    if (res.code === 200) {
      room.value.scheduledAt = dt.getTime()
      ElMessage.success('定时开播已设置')
    } else {
      ElMessage.error(res.message || '设置失败')
    }
  } catch (e) {
    ElMessage.error('设置失败')
  }
  showScheduleDialog.value = false
}

const cancelSchedule = async () => {
  try {
    const res = await liveApi.scheduleRoom(room.value.id, null)
    if (res.code === 200) {
      room.value.scheduledAt = null
      ElMessage.success('已取消定时开播')
    }
  } catch (e) {
    ElMessage.error('取消失败')
  }
  showScheduleDialog.value = false
}

const formatScheduledAt = (ts) => {
  if (!ts) return ''
  const d = new Date(ts)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

let scheduleNotifyTimer = null

const scheduleReminder = (ts) => {
  clearTimeout(scheduleNotifyTimer)
  const ms = typeof ts === 'number' ? ts : new Date(ts).getTime()
  const notifyAt = ms - 5 * 60000  // 开播前5分钟
  const delay = notifyAt - Date.now()
  if (delay > 0) {
    scheduleNotifyTimer = setTimeout(() => {
      ElMessage.warning({
        message: `距离开播还有 5 分钟：${room.value?.roomName || '我的直播间'}`,
        duration: 0
      })
      if (isNotificationEnabled()) {
        new Notification('直播提醒', { body: `${room.value?.roomName || '直播间'} 即将开播！` })
      }
    }, delay)
  }
}

const goLive = async () => {
  if (!room.value) return
  // 3秒倒计时确认
  countdown.value = 3
  const confirmStart = await ElMessageBox.confirm(
    `<div style="text-align:center;font-size:48px;margin-bottom:8px;">${countdown.value}</div><div style="font-size:14px;color:#9499a0;">即将开始直播...</div>`,
    '确定要开播吗？',
    {
      confirmButtonText: '立即开播',
      cancelButtonText: '取消',
      type: 'warning',
      showCancelButton: true,
      dangerouslyUseHTMLString: true,
      closeOnClickModal: false
    }
  ).catch(() => false)
  if (!confirmStart) return

  // 倒计时动画
  return new Promise((resolve) => {
    const tick = () => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(countdownTimer)
        resolve(doStartLive())
      }
    }
    countdownTimer = setInterval(tick, 1000)
    tick()
  })
}

const doStartLive = async () => {
  if (!room.value) return
  switching.value = true
  try {
    const res = await liveApi.updateRoomStatus(room.value.id, 'live')
    if (res.code === 200) {
      room.value.status = 'live'
      ElMessage.success('已开播')
    } else {
      ElMessage.error(res.message || '开播失败')
    }
  } catch (e) {
    ElMessage.error('开播失败')
  } finally {
    switching.value = false
  }
}

const stopLive = async () => {
  if (!room.value) return
  try {
    await ElMessageBox.confirm('确定要结束本场直播吗？观众将无法继续观看。', '结束直播', {
      confirmButtonText: '结束',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch (e) {
    return
  }
  switching.value = true
  try {
    const res = await liveApi.updateRoomStatus(room.value.id, 'offline')
    if (res.code === 200) {
      room.value.status = 'offline'
      ElMessage.success('已下播')
    } else {
      ElMessage.error(res.message || '下播失败')
    }
  } catch (e) {
    ElMessage.error('下播失败')
  } finally {
    switching.value = false
  }
}

// ---- WebRTC 推流 ----
const getUserMediaStream = async (type) => {
  if (type === 'screen') {
    return navigator.mediaDevices.getDisplayMedia({ video: true, audio: false })
  }
  return navigator.mediaDevices.getUserMedia({ video: true, audio: false })
}

const startWebRtc = async () => {
  if (isWebRtcPushing.value) return
  if (!room.value) return
  try {
    isWebRtcPushing.value = true
    const stream = await getUserMediaStream(sourceType.value)
    webrtcStream.value = stream
    await nextTick()
    if (videoPreviewRef.value) {
      videoPreviewRef.value.srcObject = stream
    }
    // WHIP 推流到 SRS
    pushPc = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    })
    stream.getTracks().forEach(track => pushPc.addTrack(track, stream))
    const offer = await pushPc.createOffer()
    await pushPc.setLocalDescription(offer)
    const whipUrl = `http://${location.hostname}:19854/rtc/v1/whip/stream/live/${streamKey.value}`
    const res = await fetch(whipUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/sdp' },
      body: offer.sdp
    })
    if (!res.ok) throw new Error('WHIP push failed')
    const answer = await res.text()
    await pushPc.setRemoteDescription({ type: 'answer', sdp: answer })
    // 监听屏幕共享停止
    if (sourceType.value === 'screen') {
      stream.getVideoTracks()[0].onended = () => stopWebRtc()
    }
    // 自动开始直播
    if (room.value.status !== 'live') {
      await liveApi.updateRoomStatus(room.value.id, 'live')
      room.value.status = 'live'
    }
    ElMessage.success(sourceType.value === 'screen' ? '屏幕共享推流中' : '摄像头推流中')
  } catch (e) {
    console.error(e)
    ElMessage.error(e.message || '推流失败，请确保 SRS 已启动')
    stopWebRtc()
  }
}

const stopWebRtc = () => {
  if (videoPreviewRef.value) {
    videoPreviewRef.value.srcObject = null
  }
  if (pushPc) {
    pushPc.close()
    pushPc = null
  }
  if (webrtcStream.value) {
    webrtcStream.value.getTracks().forEach(t => t.stop())
    webrtcStream.value = null
  }
  isWebRtcPushing.value = false
}

watch(sourceType, (newType) => {
  if (isWebRtcPushing.value) {
    stopWebRtc()
    startWebRtc()
  }
})

const beforeCoverUpload = (file) => {
  const isImg = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isImg) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB')
    return false
  }
  return true
}

const doUploadCover = async ({ file }) => {
  try {
    const res = await liveApi.uploadCover(room.value.id, file)
    if (res.code === 200 && res.data) {
      room.value.coverUrl = res.data
      ElMessage.success('封面上传成功')
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  } catch (e) {
    ElMessage.error('封面上传失败')
  }
}

onUnmounted(() => {
  stopWebRtc()
  if (ws) { ws.close(); ws = null }
  if (metricsTimer) { clearInterval(metricsTimer); metricsTimer = null }
})
</script>

<template>
  <div class="push-page">
    <div v-if="loading" v-loading="true" class="loading-state" />

    <template v-else-if="room">
      <!-- 开播倒计时浮层 -->
      <div v-if="countdown > 0" class="countdown-overlay">
        <div class="countdown-ring">
          <span class="countdown-number">{{ countdown }}</span>
        </div>
        <p>即将开始直播...</p>
      </div>

      <!-- 顶部横幅 -->
      <div class="page-banner">
        <div class="banner-left">
          <h2 class="page-title">直播控制台</h2>
          <span class="room-tag" v-if="!isEditingRoomName" @click="startEditRoomName" title="点击修改直播间名称">
            <VideoCamera style="width: 14px; height: 14px;" />
            {{ room.roomName || '我的直播间' }}
          </span>
          <input
            v-else
            ref="editRoomNameInput"
            v-model="room.roomName"
            class="room-name-input"
            @blur="saveRoomName"
            @keydown.enter="saveRoomName"
          />
          <el-tag v-if="isLive" type="danger" effect="dark" size="small">
            <span class="live-dot" />直播中
          </el-tag>
          <el-tag v-else type="info" size="small">未开播</el-tag>
          <el-select v-model="selectedCategory" size="small" placeholder="选择分类" style="width:100px" @change="saveCategory">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
          <div v-if="isLive" class="floating-metrics">
            <span class="metric-item">
              <el-icon><UserFilled /></el-icon>
              {{ viewers.length }}
            </span>
            <span class="metric-divider" />
            <span class="metric-item">
              <el-icon><ChatDotRound /></el-icon>
              {{ danmakuRate }}/分
            </span>
            <span class="metric-divider" />
            <span class="metric-item stat">
              峰值 {{ peakViewers }}
            </span>
            <span class="metric-divider" />
            <span class="metric-item stat">
              弹幕 {{ totalDanmaku }}
            </span>
            <span v-if="leaderboard.length > 0" class="metric-divider" />
            <span v-if="leaderboard.length > 0" class="metric-item leaderboard-btn" title="弹幕活跃榜">
              <span class="trophy-icon">🏆</span>
              <span class="leaderboard-preview">
                <span v-for="(v, i) in leaderboard.slice(0, 3)" :key="v.userId" class="leaderboard-entry">
                  <span class="rank">{{ i + 1 }}</span>
                  <span class="name">{{ v.userName.slice(0, 4) }}</span>
                  <span class="count">{{ v.count }}</span>
                </span>
              </span>
            </span>
          </div>
        </div>
        <div class="banner-right">
          <el-button v-if="!isLive && !room?.scheduledAt" type="primary" plain @click="openScheduleDialog" title="设置定时开播">
            🕐 定时开播
          </el-button>
          <el-tag v-if="!isLive && room?.scheduledAt" type="warning" size="small" style="cursor:pointer" @click="openScheduleDialog">
            🕐 {{ formatScheduledAt(room.scheduledAt) }}
          </el-tag>
          <el-button v-if="!isLive" type="danger" @click="goLive" :loading="switching">
            开始直播
          </el-button>
          <el-button v-else type="warning" @click="stopLive" :loading="switching">
            结束直播
          </el-button>
        </div>
      </div>

      <div class="content-grid">
        <!-- OBS 推流 -->
        <el-card class="config-card obs-card" shadow="never">
          <template #header>
            <div class="card-header">
              <div class="card-header-left">
                <Cellphone style="width: 18px; height: 18px;" />
                <span>OBS / 第三方软件推流</span>
              </div>
              <el-tag size="small" type="warning" effect="plain">推荐</el-tag>
            </div>
          </template>

          <div class="obs-body">
            <div class="config-item" v-for="(item, idx) in [
              { label: '推流地址', value: rtmpUrl },
              { label: '流密钥（Stream Key）', value: streamKey },
              { label: '完整推流地址', value: rtmpUrl + '/' + streamKey },
            ]" :key="idx">
              <div class="item-label">{{ item.label }}</div>
              <div class="item-value-row">
                <el-input :model-value="item.value" readonly class="item-input">
                  <template #append>
                    <el-button @click="copyText(item.value)">
                      <el-icon><CopyDocument /></el-icon>
                      复制
                    </el-button>
                  </template>
                </el-input>
              </div>
            </div>

            <div class="obs-steps">
              <div class="step-title">配置步骤</div>
              <ol class="step-list">
                <li>打开 OBS Studio → <strong>设置</strong> → <strong>直播</strong></li>
                <li>服务选择 <strong>"自定义..."</strong></li>
                <li>将上方 <strong>推流地址</strong> 和 <strong>流密钥</strong> 分别填入对应输入框</li>
                <li>点击 <strong>"开始推流"</strong></li>
              </ol>
            </div>
          </div>
        </el-card>

        <!-- 浏览器推流（WebRTC） -->
        <el-card class="config-card webrtc-card" shadow="never">
          <template #header>
            <div class="card-header">
              <div class="card-header-left">
                <VideoCamera style="width: 18px; height: 18px;" />
                <span>浏览器推流</span>
              </div>
              <el-tag size="small" type="success" effect="plain">无需 OBS</el-tag>
            </div>
          </template>
          <div class="webrtc-body">
            <!-- 模式切换 -->
            <div class="source-toggle">
              <el-radio-group v-model="sourceType" size="default">
                <el-radio-button value="camera">
                  <el-icon><VideoCamera /></el-icon> 摄像头
                </el-radio-button>
                <el-radio-button value="screen">
                  <el-icon><Monitor /></el-icon> 屏幕共享
                </el-radio-button>
              </el-radio-group>
            </div>

            <!-- 视频预览 -->
            <div class="video-preview-wrap">
              <video ref="videoPreviewRef" class="video-preview" autoplay muted playsinline />
              <div v-if="!isWebRtcPushing" class="preview-placeholder">
                <el-icon><VideoCamera /></el-icon>
                <span>{{ sourceType === 'screen' ? '点击下方按钮开始屏幕共享' : '点击下方按钮预览摄像头' }}</span>
              </div>
              <div v-if="isWebRtcPushing" class="pushing-badge">
                <span class="live-dot" />推流中
              </div>
            </div>

            <!-- 推流按钮 -->
            <div class="webrtc-actions">
              <el-button v-if="!isWebRtcPushing" type="primary" @click="startWebRtc" :disabled="!room">
                <el-icon><VideoCamera /></el-icon>
                {{ sourceType === 'screen' ? '开始屏幕共享' : '开始推流' }}
              </el-button>
              <el-button v-else type="danger" @click="stopWebRtc">
                <el-icon><Close /></el-icon>
                停止推流
              </el-button>
            </div>

            <div class="webrtc-hint">
              <el-alert type="info" :closable="false" show-icon>
                <template #title>
                  浏览器推流使用 WebRTC 直推 SRS，适合不方便使用 OBS 的场景
                </template>
              </el-alert>
            </div>
          </div>
        </el-card>

        <!-- 直播间封面 -->
        <el-card class="config-card cover-card" shadow="never">
          <template #header>
            <div class="card-header">
              <div class="card-header-left">
                <Picture style="width: 18px; height: 18px;" />
                <span>直播间封面</span>
              </div>
            </div>
          </template>
          <div class="cover-body">
            <div class="cover-preview">
              <img v-if="room.coverUrl" :src="room.coverUrl" alt="cover" />
              <div v-else class="cover-placeholder">
                <el-icon><Picture /></el-icon>
                <span>暂未上传封面</span>
              </div>
            </div>
            <div class="cover-actions">
              <el-upload
                :show-file-list="false"
                :before-upload="beforeCoverUpload"
                :http-request="doUploadCover"
                accept="image/*"
              >
                <el-button type="primary" plain size="small">
                  <el-icon><Upload /></el-icon>
                  上传封面
                </el-button>
              </el-upload>
              <span class="cover-hint">建议尺寸 1280×720，JPG/PNG，不超过 2MB</span>
            </div>
          </div>
        </el-card>

        <!-- 右侧监控面板 -->
        <div class="monitor-column">
          <el-card class="monitor-card" shadow="never">
            <template #header>
              <div class="card-header">
                <div class="card-header-left">
                  <UserFilled style="width: 16px; height: 16px;" />
                  <span>在线观众</span>
                </div>
                <el-tag type="success" effect="plain" size="small">{{ viewers.length }}</el-tag>
              </div>
            </template>
            <div class="viewer-list">
              <div v-if="viewers.length === 0" class="monitor-empty">
                {{ isLive ? '等待观众进入...' : '开播后才会有观众' }}
              </div>
              <div v-for="v in viewers" :key="v.userId" class="viewer-item">
                <div class="viewer-avatar">{{ (v.userName || '?').charAt(0).toUpperCase() }}</div>
                <div class="viewer-info">
                  <div class="viewer-name">
                    {{ v.userName }}
                    <span v-if="viewerDanmakuCount[String(v.userId)]" class="level-badge" :style="{ background: getViewerLevel(v.userId).color }">
                      {{ getViewerLevel(v.userId).icon }}{{ getViewerLevel(v.userId).label }}
                    </span>
                  </div>
                  <div class="viewer-time">加入于 {{ formatJoinTime(v.joinAt) }} · {{ viewerDanmakuCount[String(v.userId)] || 0 }} 条弹幕</div>
                </div>
              </div>
            </div>
          </el-card>

          <el-card class="monitor-card" shadow="never">
            <template #header>
              <div class="card-header">
                <div class="card-header-left">
                  <ChatDotRound style="width: 16px; height: 16px;" />
                  <span>实时弹幕</span>
                </div>
                <el-tag type="info" effect="plain" size="small">最近 {{ recentChats.length }}</el-tag>
              </div>
            </template>
            <div class="chat-monitor">
              <div v-if="recentChats.length === 0" class="monitor-empty">
                暂无观众发言
              </div>
              <div v-for="m in recentChats.slice().reverse()" :key="m.id" class="chat-item">
                <span class="chat-from">{{ m.from }}：</span>
                <span class="chat-text">{{ m.text }}</span>
              </div>
            </div>
          </el-card>
        </div>
      </div>

      <!-- 底部提示 -->
      <div class="page-footer">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>
            推流前请确保 SRS 媒体服务器已启动，否则推流将失败
          </template>
        </el-alert>
      </div>
    </template>

    <!-- 定时开播对话框 -->
    <el-dialog v-model="showScheduleDialog" title="定时开播" width="380px">
      <div class="schedule-form">
        <p class="schedule-tip">设置开播时间，到时将提醒您开播</p>
        <el-date-picker
          v-model="scheduleDate"
          type="date"
          placeholder="选择日期"
          style="width:100%;margin-bottom:12px"
          :disabled-date="d => d < new Date()"
        />
        <el-time-picker
          v-model="scheduleTime"
          placeholder="选择时间"
          style="width:100%"
          format="HH:mm"
          value-format="HH:mm"
        />
      </div>
      <template #footer>
        <el-button @click="showScheduleDialog = false">取消</el-button>
        <el-button v-if="room?.scheduledAt" type="danger" plain @click="cancelSchedule">取消定时</el-button>
        <el-button type="primary" @click="saveSchedule">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.push-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 24px;
  min-height: calc(100vh - 64px);
  background: #f6f7f8;
}

.page-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}
.banner-left {
  display: flex;
  align-items: center;
  gap: 14px;
}
.page-title {
  font-size: 24px;
  font-weight: 700;
  color: #18191c;
  margin: 0;
  letter-spacing: -0.3px;
}
.room-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  background: #eef0f2;
  border-radius: 6px;
  font-size: 13px;
  color: #61666d;
  cursor: pointer;
}
.room-tag:hover { background: #dcdfe6; }
.room-name-input {
  border: 1px solid #00a1d6;
  border-radius: 6px;
  padding: 3px 8px;
  font-size: 13px;
  background: #fff;
  color: #18191c;
  outline: none;
  max-width: 200px;
}
.countdown-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.countdown-ring {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  border: 6px solid #fb7299;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  animation: countdown-pulse 1s infinite;
}
@keyframes countdown-pulse { 0%,100%{transform:scale(1)} 50%{transform:scale(1.05)} }
.countdown-number { font-size: 56px; font-weight: 700; color: #fb7299; }
.countdown-overlay p { color: #fff; font-size: 16px; margin: 0; }
.live-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #fff;
  border-radius: 50%;
  margin-right: 4px;
  animation: live-blink 1.5s infinite;
  vertical-align: middle;
}
.floating-metrics {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 3px 10px;
  background: rgba(240, 64, 64, 0.1);
  border: 1px solid rgba(240, 64, 64, 0.2);
  border-radius: 20px;
}
.metric-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  color: #f04040;
  font-weight: 600;
}
.metric-item .el-icon { font-size: 13px; }
.metric-item.stat { font-size: 11px; }
.metric-divider {
  width: 1px;
  height: 12px;
  background: rgba(240, 64, 64, 0.3);
}
@keyframes live-blink { 0%,100%{opacity:1} 50%{opacity:0.3} }
.leaderboard-btn {
  cursor: default;
  position: relative;
}
.trophy-icon { font-size: 13px; }
.leaderboard-preview {
  display: none;
  position: absolute;
  top: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 10px 14px;
  min-width: 160px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 100;
}
.leaderboard-btn:hover .leaderboard-preview { display: flex; flex-direction: column; gap: 6px; }
.leaderboard-entry {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.leaderboard-entry .rank {
  width: 16px; height: 16px;
  background: #fb7299;
  color: #fff;
  border-radius: 50%;
  font-size: 10px;
  display: flex; align-items: center; justify-content: center;
  font-weight: 600;
}
.leaderboard-entry .name { flex: 1; color: #18191c; font-weight: 500; }
.leaderboard-entry .count { color: #fb7299; font-weight: 600; }

.cover-card :deep(.el-card__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #eef0f2;
  background: #fafbfc;
}
.cover-card :deep(.el-card__body) { padding: 0; }
.cover-body { padding: 20px; display: flex; flex-direction: column; gap: 14px; }
.cover-preview {
  width: 100%;
  aspect-ratio: 16/9;
  border-radius: 8px;
  overflow: hidden;
  background: #f6f7f8;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cover-preview img { width: 100%; height: 100%; object-fit: cover; }
.cover-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #c0c4cc;
  font-size: 13px;
}
.cover-placeholder .el-icon { font-size: 36px; }
.cover-actions { display: flex; flex-direction: column; gap: 8px; }
.cover-hint { font-size: 11px; color: #9499a0; }

.content-grid {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 20px;
  align-items: start;
}
@media (max-width: 900px) {
  .content-grid { grid-template-columns: 1fr; }
}
.config-card {
  border: 1px solid #e3e5e7;
  border-radius: 12px;
  overflow: hidden;
}

/* 监控面板（右栏） */
.monitor-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.monitor-card {
  border: 1px solid #e3e5e7;
  border-radius: 12px;
  overflow: hidden;
}
.monitor-empty {
  padding: 24px 12px;
  text-align: center;
  color: #9499a0;
  font-size: 13px;
}
.viewer-list {
  max-height: 320px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.viewer-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 4px;
  border-bottom: 1px solid #f6f7f8;
}
.viewer-item:last-child { border-bottom: none; }
.viewer-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #fb7299, #ff9eb1);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}
.viewer-info { flex: 1; min-width: 0; }
.viewer-name {
  font-size: 13px;
  color: #18191c;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.viewer-time {
  font-size: 11px;
  color: #9499a0;
}
.level-badge {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  color: #fff;
  font-weight: 600;
  vertical-align: middle;
  margin-left: 4px;
}

.chat-monitor {
  max-height: 320px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.chat-item {
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px dashed #f6f7f8;
}
.chat-item:last-child { border-bottom: none; }
.chat-from { color: #00a1d6; font-weight: 500; }
.chat-text { color: #18191c; }
.config-card :deep(.el-card__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #eef0f2;
  background: #fafbfc;
}
.config-card :deep(.el-card__body) {
  padding: 0;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.card-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #18191c;
}

.obs-body {
  padding: 20px;
}
.config-item {
  margin-bottom: 16px;
}
.config-item:last-of-type {
  margin-bottom: 20px;
}
.item-label {
  font-size: 13px;
  font-weight: 500;
  color: #61666d;
  margin-bottom: 6px;
}
.item-value-row {
  display: flex;
  gap: 8px;
}
.item-input :deep(.el-input__inner) {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 13px;
  color: #18191c;
  background: #f6f7f8;
}
.item-input :deep(.el-input-group__append) {
  padding: 0;
  background: #fff;
}
.item-input :deep(.el-input-group__append .el-button) {
  margin: 0;
  border: none;
  border-radius: 0;
  color: #00a1d6;
}
.item-input :deep(.el-input-group__append .el-button:hover) {
  color: #00b3e5;
  background: #f6f7f8;
}

.obs-steps {
  margin-top: 8px;
  padding: 16px;
  background: #f6f7f8;
  border-radius: 8px;
}
.step-title {
  font-size: 13px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 10px;
}
.step-list {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  color: #61666d;
  line-height: 1.8;
}
.step-list li strong {
  color: #18191c;
}

.schedule-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.schedule-tip { font-size: 13px; color: #9499a0; margin: 0 0 4px; }

.page-footer {
  margin-top: 20px;
}
.page-footer :deep(.el-alert) {
  border-radius: 8px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60vh;
}

/* 浏览器推流 */
.webrtc-card :deep(.el-card__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #eef0f2;
  background: #fafbfc;
}
.webrtc-card :deep(.el-card__body) {
  padding: 0;
}
.webrtc-body {
  padding: 20px;
}
.source-toggle {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}
.source-toggle .el-radio-button__inner {
  display: flex;
  align-items: center;
  gap: 4px;
}
.video-preview-wrap {
  position: relative;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  aspect-ratio: 16/9;
  margin-bottom: 16px;
}
.video-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.preview-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #606266;
  font-size: 13px;
}
.preview-placeholder .el-icon {
  font-size: 40px;
  color: #c0c4cc;
}
.pushing-badge {
  position: absolute;
  top: 10px;
  left: 10px;
  background: rgba(240, 64, 64, 0.85);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.pushing-badge .live-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #fff;
  border-radius: 50%;
  animation: live-blink 1.5s infinite;
}
.webrtc-actions {
  display: flex;
  justify-content: center;
  margin-bottom: 14px;
}
.webrtc-actions .el-button {
  min-width: 140px;
}
.webrtc-hint :deep(.el-alert) {
  border-radius: 6px;
}
.webrtc-hint :deep(.el-alert__body) {
  padding: 6px 10px;
}
</style>