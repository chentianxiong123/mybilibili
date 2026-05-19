<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Monitor, VideoCamera, CopyDocument, Cellphone } from '@element-plus/icons-vue'
import { liveApi } from '../../api/live.js'

const room = ref(null)
const loading = ref(true)
const rtmpUrl = ref('')
const streamKey = ref('')
const previewRef = ref(null)

const isWebRtcPushing = ref(false)
const sourceType = ref('camera')
const screenPreviewStream = ref(null)
let localStream = null
let publisher = null

const hasScreenSelected = computed(() => !!screenPreviewStream.value)

const hostname = window.location.hostname

onMounted(async () => {
  try {
    let res = await liveApi.getMyRoom()
    if (!res || !res.data) {
      res = await liveApi.createRoom('我的直播间')
    }
    if (res && res.code === 200 && res.data) {
      room.value = res.data
      streamKey.value = res.data.streamKey
      rtmpUrl.value = `rtmp://${hostname}/live`
    }
  } catch (e) {
    ElMessage.error('获取直播间信息失败')
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  stopWebRtc()
  stopScreenCapture()
})

const copyText = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制')
  })
}

// 仅捕获屏幕用于预览，不推流
const startScreenCapture = async () => {
  // 如果已有预览流，先释放
  stopScreenCapture()

  try {
    const stream = await navigator.mediaDevices.getDisplayMedia({
      video: true,
      audio: true
    })

    screenPreviewStream.value = stream

    if (previewRef.value) {
      previewRef.value.srcObject = stream
    }

    // 用户点浏览器"停止共享"时自动清理
    const videoTrack = stream.getVideoTracks()[0]
    if (videoTrack) {
      videoTrack.onended = () => {
        stopScreenCapture()
        if (!isWebRtcPushing.value) {
          ElMessage.info('已取消屏幕共享')
        }
      }
    }
  } catch (e) {
    if (e.name === 'NotAllowedError') {
      ElMessage.warning('屏幕共享请求被拒绝')
    } else {
      ElMessage.error('屏幕捕获失败: ' + e.message)
    }
  }
}

const stopScreenCapture = () => {
  if (screenPreviewStream.value) {
    screenPreviewStream.value.getTracks().forEach(t => t.stop())
    screenPreviewStream.value = null
  }
  // 仅限于不在推流中时清空预览，推流中由 stopWebRtc 管理
  if (!isWebRtcPushing.value && previewRef.value) {
    previewRef.value.srcObject = null
  }
}

const getUserMediaStream = async (type) => {
  if (type === 'camera') {
    return navigator.mediaDevices.getUserMedia({
      video: {
        width: { ideal: 1280 },
        height: { ideal: 720 },
        frameRate: { ideal: 30 }
      },
      audio: true
    })
  }
  const stream = screenPreviewStream.value
  if (!stream) throw new Error('未选择屏幕')
  screenPreviewStream.value = null
  return stream
}

const startWebRtc = async (type = 'camera') => {
  if (isWebRtcPushing.value) return

  try {
    const stream = await getUserMediaStream(type)
    localStream = stream

    if (previewRef.value) {
      previewRef.value.srcObject = stream
    }

    if (type === 'screen') {
      const videoTrack = stream.getVideoTracks()[0]
      if (videoTrack) {
        videoTrack.onended = () => {
          stopWebRtc()
          ElMessage.info('屏幕共享已停止')
        }
      }
    }

    const pc = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    })
    publisher = pc

    stream.getTracks().forEach(track => pc.addTrack(track, stream))

    // 设置编码参数：目标 5000kbps，维持分辨率不降
    const sender = pc.getSenders().find(s => s.track?.kind === 'video')
    if (sender) {
      const params = sender.getParameters()
      if (!params.encodings) params.encodings = [{}]
      params.encodings[0] = {
        active: true,
        maxBitrate: 5000000,
        minBitrate: 1000000,
        scaleResolutionDownBy: 1
      }
      sender.setParameters(params)
    }

    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)

    // 使用 SRS 的 codec 参数强制 VP8，绕过 H264 硬件编码器码率 cap
    const whipUrl = `http://${hostname}:1985/rtc/v1/whip/?app=live&stream=${streamKey.value}&codec=vp8`
    console.log('[WHIP] POST to:', whipUrl)
    const resp = await fetch(whipUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/sdp' },
      body: offer.sdp
    })

    if (resp.ok) {
      const answerSdp = await resp.text()
      await pc.setRemoteDescription({ type: 'answer', sdp: answerSdp })
      isWebRtcPushing.value = true
      await liveApi.updateRoomStatus(room.value.id, 'live')
      ElMessage.success('WebRTC推流已开始')
    } else {
      throw new Error('SRS返回错误: ' + resp.status)
    }
  } catch (e) {
    if (type === 'screen') {
      ElMessage.error('屏幕共享推流失败: ' + e.message)
    } else {
      ElMessage.error('摄像头推流失败: ' + e.message)
    }
    stopWebRtc()
  }
}

// 推流中切换模式 → 自动停止旧推流
watch(sourceType, (newType) => {
  if (isWebRtcPushing.value) {
    stopWebRtc()
  }
  // 切到屏幕模式时不清旧预览，切离屏幕模式时清理预览
  if (newType !== 'screen') {
    stopScreenCapture()
  }
})

const stopWebRtc = () => {
  if (publisher) {
    publisher.close()
    publisher = null
  }
  if (localStream) {
    localStream.getTracks().forEach(t => t.stop())
    localStream = null
  }
  if (previewRef.value) {
    previewRef.value.srcObject = null
  }
  isWebRtcPushing.value = false
}
</script>

<template>
  <div class="push-page">
    <div v-if="loading" v-loading="true" class="loading-state" />

    <template v-else-if="room">
      <!-- 顶部横幅 -->
      <div class="page-banner">
        <div class="banner-left">
          <h2 class="page-title">直播控制台</h2>
          <span class="room-tag">
            <VideoCamera style="width: 14px; height: 14px;" />
            {{ room.roomName || '我的直播间' }}
          </span>
        </div>
        <div class="banner-right">
          <span class="status-badge" :class="{ 'is-live': isWebRtcPushing }">
            <span class="status-dot" />
            {{ isWebRtcPushing ? '直播中' : '未开播' }}
          </span>
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

        <!-- WebRTC 推流 -->
        <el-card class="config-card webrtc-card" shadow="never">
          <template #header>
            <div class="card-header">
              <div class="card-header-left">
                <VideoCamera style="width: 18px; height: 18px;" />
                <span>浏览器直接推流</span>
              </div>
              <el-tag size="small" type="success" effect="plain">轻量</el-tag>
            </div>
          </template>

          <div class="webrtc-body">
            <!-- 模式切换 -->
            <el-radio-group v-model="sourceType" class="source-toggle" :disabled="isWebRtcPushing">
              <el-radio-button value="camera">
                <VideoCamera style="width: 14px; height: 14px; margin-right: 4px;" />
                摄像头
              </el-radio-button>
              <el-radio-button value="screen">
                <Monitor style="width: 14px; height: 14px; margin-right: 4px;" />
                屏幕共享
              </el-radio-button>
            </el-radio-group>

            <!-- 预览区域 -->
            <div class="preview-area" :class="{ 'screen-mode': sourceType === 'screen' && !isWebRtcPushing && !hasScreenSelected }">
              <div class="live-indicator" v-if="isWebRtcPushing">
                <span class="live-dot" />
                <span>LIVE</span>
              </div>

              <video
                ref="previewRef"
                autoplay
                muted
                class="preview-video"
                :class="{ 'is-pushing': isWebRtcPushing }"
              />

              <!-- 屏幕模式：未选择时的遮罩 -->
              <div v-if="sourceType === 'screen' && !isWebRtcPushing && !hasScreenSelected" class="screen-overlay">
                <div class="screen-overlay-icon">
                  <Monitor style="width: 48px; height: 48px;" />
                </div>
                <p class="screen-overlay-text">选择要共享的屏幕、窗口或标签页</p>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="webrtc-actions">
              <template v-if="!isWebRtcPushing">
                <template v-if="sourceType === 'camera'">
                  <el-button type="primary" size="large" round @click="startWebRtc('camera')">
                    <VideoCamera style="width: 16px; height: 16px; margin-right: 4px;" />
                    开始推流
                  </el-button>
                </template>
                <template v-else>
                  <el-button
                    v-if="!hasScreenSelected"
                    type="primary"
                    size="large"
                    round
                    @click="startScreenCapture"
                  >
                    <Monitor style="width: 16px; height: 16px; margin-right: 4px;" />
                    选择屏幕
                  </el-button>
                  <div v-else class="screen-btn-group">
                    <el-button type="primary" size="large" round @click="startWebRtc('screen')">
                      <VideoCamera style="width: 16px; height: 16px; margin-right: 4px;" />
                      开始推流
                    </el-button>
                    <el-button size="large" round @click="startScreenCapture">重新选择</el-button>
                  </div>
                </template>
              </template>
              <template v-else>
                <el-button type="danger" size="large" round @click="stopWebRtc">
                  停止推流
                </el-button>
                <div class="push-status">
                  <span class="source-badge">{{ sourceType === 'camera' ? '摄像头' : '屏幕共享' }}</span>
                  <span class="push-info">推流中 · {{ streamKey }}</span>
                </div>
              </template>
            </div>

            <!-- 提示文字 -->
            <p v-if="sourceType === 'screen' && !hasScreenSelected && !isWebRtcPushing" class="hint-text">
              浏览器将弹出安全选择框，系统级的隐私保护，不会被任何页面拦截
            </p>
          </div>
        </el-card>
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

/* ===== 顶部横幅 ===== */
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
}
.banner-right {
  display: flex;
  align-items: center;
}
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  background: #e3e5e7;
  color: #9499a0;
  transition: all 0.3s;
}
.status-badge.is-live {
  background: #fef0f0;
  color: #f04040;
}
.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
.status-badge.is-live .status-dot {
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* ===== 卡片通用 ===== */
.content-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.config-card {
  border: 1px solid #e3e5e7;
  border-radius: 12px;
  overflow: hidden;
}
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

/* ===== OBS 卡片 ===== */
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

/* ===== WebRTC 卡片 ===== */
.webrtc-body {
  padding: 24px 20px 20px;
}
.source-toggle {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}
.source-toggle :deep(.el-radio-button__inner) {
  padding: 8px 28px;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
}

/* 预览区域 */
.preview-area {
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  background: #000;
  margin-bottom: 16px;
  max-width: 560px;
  margin-left: auto;
  margin-right: auto;
}
.preview-area.screen-mode {
  min-height: 260px;
}
.preview-video {
  display: block;
  width: 100%;
  max-height: 380px;
  object-fit: contain;
  transition: opacity 0.3s;
}
.preview-video.is-pushing {
  border: 2px solid #f04040;
  border-radius: 12px;
}

/* LIVE 标记 */
.live-indicator {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 10;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: #f04040;
  color: #fff;
  font-size: 13px;
  font-weight: 700;
  border-radius: 4px;
  letter-spacing: 0.5px;
}
.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #fff;
  animation: pulse 1.5s infinite;
}

/* 屏幕选择遮罩 */
.screen-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  gap: 12px;
}
.screen-overlay-icon {
  opacity: 0.6;
}
.screen-overlay-text {
  font-size: 14px;
  opacity: 0.8;
  margin: 0;
}

/* 操作按钮 */
.webrtc-actions {
  text-align: center;
}
.screen-btn-group {
  display: flex;
  justify-content: center;
  gap: 12px;
}
.push-status {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
.source-badge {
  display: inline-block;
  padding: 3px 10px;
  background: #eef0f2;
  border-radius: 4px;
  font-size: 12px;
  color: #61666d;
}
.push-info {
  font-size: 12px;
  color: #9499a0;
  font-family: 'SF Mono', 'Fira Code', monospace;
}

.hint-text {
  margin: 12px 0 0;
  font-size: 12px;
  color: #9499a0;
  text-align: center;
}

/* ===== 底部 ===== */
.page-footer {
  margin-top: 20px;
}
.page-footer :deep(.el-alert) {
  border-radius: 8px;
}

/* ===== 加载 ===== */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60vh;
}
</style>