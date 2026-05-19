<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoCamera, CopyDocument, Cellphone } from '@element-plus/icons-vue'
import { liveApi } from '../../api/live.js'

const room = ref(null)
const loading = ref(true)
const rtmpUrl = ref('')
const streamKey = ref('')

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

const copyText = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制')
  })
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
</style>