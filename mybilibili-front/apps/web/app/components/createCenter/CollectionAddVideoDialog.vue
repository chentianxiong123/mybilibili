<template>
  <el-dialog
    v-model="dialogVisible"
    title="管理合集稿件"
    width="700px"
    destroy-on-close
  >
    <div class="video-selection-header">
      <span class="selection-tip">勾选稿件添加到合集，取消勾选从合集移除</span>
    </div>

    <div v-if="loading" class="dialog-loading">
      <el-skeleton :rows="5" animated />
    </div>

    <div v-else-if="availableVideos.length === 0" class="dialog-empty">
      <el-empty description="暂无可添加的稿件" />
    </div>

    <div v-else class="video-select-list">
      <el-checkbox-group v-model="selectedVideosProxy">
        <div v-for="video in availableVideos" :key="video.id" class="video-select-item">
          <el-checkbox :label="video.id">
            <div class="video-select-content">
              <img loading="lazy" decoding="async" :src="video.coverUrl || getDefaultCover()" class="video-select-cover" />
              <div class="video-select-info">
                <div class="video-select-title">{{ video.title }}</div>
                <div class="video-select-meta">{{ formatNumber(video.viewCount) }} 播放</div>
              </div>
            </div>
          </el-checkbox>
          <span v-if="selectedVideos.includes(video.id)" class="in-collection-badge">已加入</span>
        </div>
      </el-checkbox-group>
    </div>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="emit('submit')" :loading="loading">更新</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  availableVideos: {
    type: Array,
    default: () => []
  },
  selectedVideos: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'update:selectedVideos', 'submit'])

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const selectedVideosProxy = computed({
  get: () => props.selectedVideos,
  set: (val) => emit('update:selectedVideos', val)
})

const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}
</script>

<style scoped>
.video-select-list {
  max-height: 400px;
  overflow-y: auto;
}

.video-selection-header {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.selection-tip {
  font-size: 13px;
  color: #606266;
}

.dialog-loading,
.dialog-empty {
  padding: 40px 0;
  text-align: center;
}

.video-select-item {
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.video-select-item:last-child {
  border-bottom: none;
}

.video-select-item .el-checkbox {
  flex: 1;
}

.video-select-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.video-select-cover {
  width: 80px;
  height: 45px;
  object-fit: cover;
  border-radius: 4px;
}

.in-collection-badge {
  background: linear-gradient(135deg, #005980, #1890ff);
  color: #ffffff;
  padding: 6px 14px;
  border-radius: 4px;
  font-size: 12px;
  flex-shrink: 0;
  margin-left: 8px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.video-select-info {
  flex: 1;
}

.video-select-title {
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
}

.video-select-meta {
  font-size: 12px;
  color: #909399;
}
</style>