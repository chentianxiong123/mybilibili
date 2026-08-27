<script setup>
import { ElMessage } from 'element-plus'
import { UploadFilled, Plus } from '@element-plus/icons-vue'
import { toWebP } from '@/utils/toWebP'

defineProps({
  coverPreview: { type: String, default: '' }
})

const emit = defineEmits(['cover-change', 'update:coverPreview'])

const handleCoverUpload = async (file) => {
  const isImage = file.raw.type.startsWith('image/')
  const isLt10M = file.raw.size / 1024 / 1024 < 10

  if (!isImage) {
    ElMessage.error('封面只能是图片格式!')
    return false
  }
  if (!isLt10M) {
    ElMessage.error('封面大小不能超过 10MB!')
    return false
  }

  emit('cover-change', await toWebP(file.raw))
  const reader = new FileReader()
  reader.onload = (e) => {
    emit('update:coverPreview', e.target.result)
  }
  reader.readAsDataURL(file.raw)
  return false
}
</script>

<template>
  <div class="cover-upload-section">
    <el-upload
      class="cover-uploader"
      action="#"
      :on-change="handleCoverUpload"
      :auto-upload="false"
      accept="image/*"
      :show-file-list="false"
    >
      <div v-if="coverPreview" class="cover-preview">
        <img loading="lazy" decoding="async" :src="coverPreview" alt="封面预览">
        <div class="cover-overlay">
          <el-icon><UploadFilled /></el-icon>
          <span>更换封面</span>
        </div>
      </div>
      <div v-else class="cover-placeholder">
        <el-icon class="placeholder-icon"><Plus /></el-icon>
        <span>点击上传封面</span>
      </div>
    </el-upload>
    <div class="cover-tip">
      <p>建议尺寸：16:9（1920×1080）</p>
      <p>最大 10MB，支持 JPG、PNG 格式</p>
      <p class="cover-tip-highlight">清晰的封面能吸引更多观众</p>
    </div>
  </div>
</template>

<style scoped>
.cover-upload-section {
  display: flex;
  align-items: flex-start;
  gap: 20px;
}

.cover-uploader {
  width: 240px;
  height: 135px;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  cursor: pointer;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f7fa;
  transition: all 0.3s;
}

.cover-uploader:hover {
  border-color: #409eff;
  background-color: #ecf5ff;
}

.cover-preview {
  width: 100%;
  height: 100%;
  position: relative;
}

.cover-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #fff;
  opacity: 0;
  transition: opacity 0.3s;
}

.cover-preview:hover .cover-overlay {
  opacity: 1;
}

.cover-overlay .el-icon {
  font-size: 24px;
  margin-bottom: 4px;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.placeholder-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.cover-tip {
  font-size: 13px;
  color: #606266;
  line-height: 1.8;
}

.cover-tip p {
  margin: 0;
}

.cover-tip-highlight {
  color: #409eff;
  font-weight: 500;
}
</style>