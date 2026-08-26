<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, Delete, ArrowUp, ArrowDown, VideoPlay } from '@element-plus/icons-vue'

const props = defineProps({
  videoParts: { type: Array, required: true }
})

const emit = defineEmits(['add-video', 'remove-part', 'move-up', 'move-down'])

const videoUploadRef = ref(null)

const getVideoDuration = (file) => {
  return new Promise((resolve, reject) => {
    const video = document.createElement('video')
    video.preload = 'metadata'
    video.onloadedmetadata = () => {
      URL.revokeObjectURL(video.src)
      resolve(Math.floor(video.duration))
    }
    video.onerror = (error) => reject(error)
    video.src = URL.createObjectURL(file)
  })
}

const handleVideoUpload = async (file) => {
  if (!file.raw) {
    ElMessage.error('视频文件读取失败，请重新选择')
    return false
  }

  const isVideo = file.raw.type.startsWith('video/')
  const isLt4G = file.raw.size / 1024 / 1024 / 1024 < 4

  if (!isVideo) {
    ElMessage.error('只能上传视频文件!')
    return false
  }
  if (!isLt4G) {
    ElMessage.error('视频大小不能超过 4GB!')
    return false
  }

  let duration = 0
  try {
    duration = await getVideoDuration(file.raw)
  } catch (error) {
    console.error('获取视频时长失败:', error)
  }

  const newPart = {
    id: Date.now(),
    file: file.raw,
    title: file.name.replace(/\.[^/.]+$/, ''),
    size: file.raw.size,
    duration
  }

  emit('add-video', newPart)
  return false
}

const removeVideoPart = (index) => {
  ElMessageBox.confirm(
    `确定要删除分P "${props.videoParts[index].title}" 吗？`,
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    emit('remove-part', index)
    ElMessage.success('分P已删除')
  }).catch(() => {})
}

const movePartUp = (index) => {
  if (index === 0) return
  emit('move-up', index)
}

const movePartDown = (index) => {
  if (index === props.videoParts.length - 1) return
  emit('move-down', index)
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const triggerVideoUpload = () => {
  if (videoUploadRef.value) {
    videoUploadRef.value.$el.querySelector('input').click()
  }
}

defineExpose({
  triggerVideoUpload,
  videoUploadRef
})
</script>

<template>
  <div class="video-parts-section">
    <div class="section-header">
      <h3 class="section-title">
        <el-icon><VideoPlay /></el-icon>
        视频分P
        <span class="part-count">({{ videoParts.length }}个视频)</span>
      </h3>
      <el-upload
        ref="videoUploadRef"
        class="video-upload-trigger"
        action="#"
        :on-change="handleVideoUpload"
        :auto-upload="false"
        accept="video/*"
        :show-file-list="false"
        multiple
      >
        <el-button type="primary" :icon="UploadFilled">
          添加视频
        </el-button>
      </el-upload>
    </div>

    <!-- 视频分P列表 -->
    <div v-if="videoParts.length > 0" class="video-parts-list">
      <div
        v-for="(part, index) in videoParts"
        :key="part.id"
        class="video-part-item"
      >
        <div class="part-number">P{{ index + 1 }}</div>
        <div class="part-info">
          <el-input
            v-model="part.title"
            placeholder="请输入分P标题"
            maxlength="100"
            class="part-title-input"
          ></el-input>
          <div class="part-meta">
            <span class="part-size">{{ part.file ? formatFileSize(part.size) : '' }}</span>
            <span v-if="part.file" class="part-filename">{{ part.file.name }}</span>
            <span v-else class="part-filename missing-file">文件已丢失，请重新选择视频</span>
          </div>
        </div>
        <div class="part-actions">
          <el-button
            type="primary"
            text
            :icon="ArrowUp"
            :disabled="index === 0"
            @click="movePartUp(index)"
            title="上移"
          ></el-button>
          <el-button
            type="primary"
            text
            :icon="ArrowDown"
            :disabled="index === videoParts.length - 1"
            @click="movePartDown(index)"
            title="下移"
          ></el-button>
          <el-button
            type="danger"
            text
            :icon="Delete"
            @click="removeVideoPart(index)"
            title="删除"
          ></el-button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="video-empty-state">
      <el-icon class="empty-icon"><UploadFilled /></el-icon>
      <p class="empty-text">还没有添加视频</p>
      <p class="empty-subtext">点击上方"添加视频"按钮或拖拽视频文件到此处</p>
      <el-upload
        class="video-upload-area"
        action="#"
        :on-change="handleVideoUpload"
        :auto-upload="false"
        accept="video/*"
        :show-file-list="false"
        drag
        multiple
      >
        <div class="upload-drag-content">
          <el-icon class="drag-icon"><UploadFilled /></el-icon>
          <div class="drag-text">将视频文件拖到此处，或点击上传</div>
        </div>
      </el-upload>
    </div>
  </div>
</template>

<style scoped>
.video-parts-section {
  background-color: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title .el-icon {
  color: #409eff;
}

.part-count {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
}

.video-parts-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.video-part-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  transition: all 0.3s;
}

.video-part-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.part-number {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #409eff 0%, #1677ff 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
}

.part-info {
  flex: 1;
  min-width: 0;
}

.part-title-input {
  width: 100%;
  margin-bottom: 8px;
}

.part-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: #909399;
}

.part-size {
  background-color: #e6f7ff;
  color: #1890ff;
  padding: 2px 8px;
  border-radius: 4px;
}

.part-filename {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.part-actions {
  display: flex;
  gap: 4px;
}

.video-empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 64px;
  color: #dcdfe6;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
  color: #606266;
  margin: 0 0 8px 0;
}

.empty-subtext {
  font-size: 13px;
  color: #909399;
  margin: 0 0 24px 0;
}

.video-upload-area {
  max-width: 500px;
  margin: 0 auto;
}

.upload-drag-content {
  padding: 40px;
}

.drag-icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 16px;
}

.drag-text {
  font-size: 14px;
  color: #606266;
}
</style>