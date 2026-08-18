<script setup>
import { computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, CircleCheck } from '@element-plus/icons-vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  stage: { type: String, default: '' },
  stageLabel: { type: String, default: '' },
  percentage: { type: Number, default: 0 },
  uploadedBytes: { type: Number, default: 0 },
  totalBytes: { type: Number, default: 0 },
  speed: { type: Number, default: 0 },
  etaSeconds: { type: Number, default: -1 },
  partProgress: { type: Array, default: () => [] },
  error: { type: String, default: '' },
  isUploading: { type: Boolean, default: false },
  isFinished: { type: Boolean, default: false },
  isSubmitting: { type: Boolean, default: false },
  UPLOAD_STAGES: { type: Object, default: () => ({}) }
})

const emit = defineEmits(['close', 'cancel'])

const dialogVisible = computed({
  get: () => props.show,
  set: (val) => { if (!val) emit('close') }
})

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatSpeed = (bytesPerSec) => {
  if (bytesPerSec <= 0) return '--'
  if (bytesPerSec < 1024) return bytesPerSec + ' B/s'
  if (bytesPerSec < 1024 * 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s'
  if (bytesPerSec < 1024 * 1024 * 1024) return (bytesPerSec / (1024 * 1024)).toFixed(1) + ' MB/s'
  return (bytesPerSec / (1024 * 1024 * 1024)).toFixed(2) + ' GB/s'
}

const formatEta = (seconds) => {
  if (seconds < 0 || !seconds) return '--'
  if (seconds < 60) return seconds + '秒'
  if (seconds < 3600) return Math.floor(seconds / 60) + '分' + (seconds % 60) + '秒'
  return Math.floor(seconds / 3600) + '时' + Math.floor((seconds % 3600) / 60) + '分'
}

const closeUploadDialog = () => {
  emit('close')
}

const handleCancelUpload = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要取消上传吗？已上传的分片将被清除。',
      '提示',
      { confirmButtonText: '确定取消', cancelButtonText: '继续上传', type: 'warning' }
    )
    emit('cancel')
    ElMessage.info('上传已取消')
  } catch {}
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    title="稿件上传进度"
    width="560px"
    class="upload-progress-dialog"
    :close-on-click-modal="!isSubmitting"
    :close-on-press-escape="!isSubmitting"
  >
    <div class="progress-content">
      <el-progress
        :percentage="percentage"
        :stroke-width="20"
        :status="stage === UPLOAD_STAGES.COMPLETED ? 'success' : stage === UPLOAD_STAGES.FAILED ? 'exception' : ''"
        class="upload-progress-bar"
      ></el-progress>
      <div class="upload-status">
        <el-icon v-if="!isFinished" class="status-icon loading"><Loading /></el-icon>
        <el-icon v-else-if="stage === UPLOAD_STAGES.COMPLETED" class="status-icon success"><CircleCheck /></el-icon>
        <el-icon v-else class="status-icon" style="color: #f56c6c"><CircleCheck /></el-icon>
        <span class="status-text">{{ stageLabel }}</span>
      </div>
      <div v-if="isUploading" class="upload-stats">
        <span class="stat-item">{{ formatFileSize(uploadedBytes) }} / {{ formatFileSize(totalBytes) }}</span>
        <span class="stat-item">{{ formatSpeed(speed) }}</span>
        <span class="stat-item" v-if="etaSeconds > 0">剩余 {{ formatEta(etaSeconds) }}</span>
      </div>
      <div v-if="isUploading && partProgress.length > 1" class="part-progress-list">
        <div v-for="(pp, idx) in partProgress" :key="idx" class="part-progress-item">
          <span class="part-progress-label">{{ pp.title }}</span>
          <el-progress
            :percentage="pp.total > 0 ? Math.round((pp.uploaded / pp.total) * 100) : 0"
            :stroke-width="6"
            :show-text="false"
            class="part-mini-progress"
          />
          <span class="part-progress-count">{{ pp.uploaded }}/{{ pp.total }}</span>
        </div>
      </div>
      <div v-if="stage === UPLOAD_STAGES.COMPLETED" class="upload-hint">
        稿件已提交，正在等待审核/转码处理
      </div>
      <div v-if="error" class="upload-error">
        {{ error }}
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button
          v-if="isUploading"
          type="danger"
          @click="handleCancelUpload"
        >取消上传</el-button>
        <el-button
          v-if="isFinished"
          @click="closeUploadDialog"
        >关闭</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.progress-content {
  padding: 20px;
}

.upload-progress-bar {
  margin-bottom: 20px;
}

.upload-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.status-icon {
  font-size: 20px;
}

.status-icon.loading {
  color: #409eff;
  animation: rotating 2s linear infinite;
}

.status-icon.success {
  color: #67c23a;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.status-text {
  font-size: 14px;
  color: #606266;
}

.upload-hint {
  text-align: center;
  font-size: 13px;
  color: #909399;
}

.dialog-footer {
  display: flex;
  justify-content: center;
}

.upload-stats {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #909399;
}

.stat-item {
  white-space: nowrap;
}

.part-progress-list {
  margin-top: 16px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 8px;
}

.part-progress-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.part-progress-item:last-child {
  margin-bottom: 0;
}

.part-progress-label {
  width: 80px;
  font-size: 12px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
}

.part-mini-progress {
  flex: 1;
}

.part-progress-count {
  width: 48px;
  font-size: 11px;
  color: #909399;
  text-align: right;
  flex-shrink: 0;
}

.upload-error {
  text-align: center;
  font-size: 13px;
  color: #f56c6c;
  margin-top: 8px;
}
</style>