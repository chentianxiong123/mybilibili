<script setup>
import { ref } from 'vue'
import { ChatDotRound, Picture, Link } from '@element-plus/icons-vue'
import EmojiPopover from '@/components/EmojiPopover.vue'
import VideoSelectDialog from '@/components/VideoSelectDialog.vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user.ts'

const emit = defineEmits(['publish'])
const userStore = useUserStore()

const dynamicContent = ref('')
const selectedImages = ref([])
const imagePreviewUrls = ref([])
const showEmojiPicker = ref(false)
const emojiBtnRef = ref(null)
const refVideoId = ref(null)
const refVideoInfo = ref(null)
const showVideoSelectDialog = ref(false)

const handleImageSelect = (event) => {
  const files = Array.from(event.target.files)
  if (files.length + selectedImages.value.length > 9) {
    ElMessage.warning('最多只能上传9张图片')
    return
  }
  files.forEach(file => {
    if (!file.type.startsWith('image/')) {
      ElMessage.error('只能上传图片文件')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      ElMessage.error('图片大小不能超过5MB')
      return
    }
    selectedImages.value.push(file)
    imagePreviewUrls.value.push(URL.createObjectURL(file))
  })
  event.target.value = ''
}

const removeImage = (index) => {
  URL.revokeObjectURL(imagePreviewUrls.value[index])
  selectedImages.value.splice(index, 1)
  imagePreviewUrls.value.splice(index, 1)
}

const handlePublish = () => {
  if (!dynamicContent.value.trim() && selectedImages.value.length === 0 && !refVideoId.value) {
    ElMessage.warning('请输入动态内容或添加图片')
    return
  }
  emit('publish', {
    content: dynamicContent.value,
    images: selectedImages.value,
    refVideoId: refVideoId.value
  })
}

const openVideoRefDialog = () => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  showVideoSelectDialog.value = true
}

const handleVideoSelect = (video) => {
  refVideoId.value = video.id
  refVideoInfo.value = {
    id: video.id,
    title: video.title,
    cover: video.coverUrl
  }
  showVideoSelectDialog.value = false
  ElMessage.success('已选择视频：' + video.title)
}

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

const insertEmoji = (emoji) => {
  dynamicContent.value += emoji
}

const clearVideoRef = () => {
  refVideoId.value = null
  refVideoInfo.value = null
}

const reset = () => {
  dynamicContent.value = ''
  selectedImages.value = []
  imagePreviewUrls.value = []
  refVideoId.value = null
  refVideoInfo.value = null
}

defineExpose({ reset })
</script>

<template>
  <div class="publish-box">
    <div class="publish-input-wrapper">
      <textarea
        v-model="dynamicContent"
        class="publish-input"
        placeholder="有什么想和大家分享的？"
        rows="3"
      ></textarea>
    </div>

    <div class="image-preview" v-if="imagePreviewUrls.length > 0">
      <div v-for="(url, index) in imagePreviewUrls" :key="index" class="preview-item">
        <img loading="lazy" decoding="async" :src="url" alt="预览">
        <button class="remove-btn" @click="removeImage(index)">×</button>
      </div>
    </div>

    <div class="video-ref-preview" v-if="refVideoInfo">
      <div class="video-ref-card">
        <span>引用视频：{{ refVideoInfo.title }}</span>
        <button class="clear-btn" @click="clearVideoRef">×</button>
      </div>
    </div>

    <div class="publish-toolbar">
      <div class="toolbar-left">
        <button
          ref="emojiBtnRef"
          class="tool-btn"
          :class="{ active: showEmojiPicker }"
          title="表情"
          @click="toggleEmojiPicker"
        >
          <el-icon><ChatDotRound /></el-icon>
        </button>
        <EmojiPopover
          v-model:visible="showEmojiPicker"
          :trigger-ref="emojiBtnRef"
          @select="insertEmoji"
        />
        <label class="tool-btn" title="图片">
          <el-icon><Picture /></el-icon>
          <input type="file" accept="image/*" multiple hidden @change="handleImageSelect">
        </label>
        <button class="tool-btn" title="引用视频" @click="openVideoRefDialog">
          <el-icon><Link /></el-icon>
        </button>
      </div>
      <div class="toolbar-right">
        <span class="word-count">{{ dynamicContent.length }}/200</span>
        <button class="publish-btn" :disabled="!dynamicContent.trim() && selectedImages.length === 0 && !refVideoId" @click="handlePublish">
          发布
        </button>
      </div>
    </div>
  </div>

  <VideoSelectDialog
    v-model:visible="showVideoSelectDialog"
    :user-id="userStore.userInfo?.id"
    @select="handleVideoSelect"
  />
</template>

<style scoped>
.publish-box {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.publish-input {
  width: 100%;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 12px;
  font-size: 14px;
  resize: none;
  outline: none;
  transition: border-color 0.3s;
  font-family: inherit;
  box-sizing: border-box;
}

.publish-input:focus {
  border-color: #00aeec;
}

.image-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.preview-item {
  position: relative;
  width: 80px;
  height: 80px;
}

.preview-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 6px;
}

.remove-btn {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #ff2442;
  color: #fff;
  border: none;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-ref-preview {
  margin-top: 12px;
}

.video-ref-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f6f7f8;
  border-radius: 6px;
  font-size: 13px;
}

.clear-btn {
  background: none;
  border: none;
  color: #9499a0;
  cursor: pointer;
  font-size: 16px;
}

.publish-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.toolbar-left {
  display: flex;
  gap: 8px;
}

.tool-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9499a0;
  transition: all 0.3s;
}

.tool-btn:hover {
  background: #f1f2f3;
  color: #00aeec;
}

.tool-btn.active {
  color: #00aeec;
  background: #e6f7ff;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.word-count {
  font-size: 12px;
  color: #9499a0;
}

.publish-btn {
  padding: 8px 24px;
  background: #00aeec;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.publish-btn:hover:not(:disabled) {
  background: #00a0d8;
}

.publish-btn:disabled {
  background: #e3e5e7;
  cursor: not-allowed;
}
</style>