<template>
  <el-dialog
    v-model="dialogVisible"
    title="编辑稿件"
    width="720px"
    destroy-on-close
  >
    <div v-loading="editManuscriptLoading">
      <el-alert
        title="保存后稿件会重新进入审核"
        type="warning"
        show-icon
        :closable="false"
        class="edit-manuscript-alert"
      />
      <el-form :model="editManuscriptForm" label-width="84px" class="edit-manuscript-form">
        <el-form-item label="封面">
          <div class="edit-cover-row">
            <el-upload
              action="#"
              :show-file-list="false"
              :auto-upload="false"
              :on-change="handleEditManuscriptCoverChange"
              accept="image/*"
            >
              <div class="edit-cover-box">
                <img loading="lazy" decoding="async"
                  v-if="editManuscriptCoverPreview"
                  :src="editManuscriptCoverPreview"
                  alt="稿件封面"
                />
                <div v-else class="edit-cover-empty">
                  <el-icon><Picture /></el-icon>
                  <span>选择封面</span>
                </div>
              </div>
            </el-upload>
            <div class="edit-cover-tip">支持 JPG、PNG，最大 10MB</div>
          </div>
        </el-form-item>

        <el-form-item label="标题" required>
          <el-input
            v-model="editManuscriptForm.title"
            maxlength="80"
            show-word-limit
            placeholder="请输入稿件标题"
          />
        </el-form-item>

        <el-form-item label="分区" required>
          <el-select
            v-model="editManuscriptForm.categoryId"
            placeholder="请选择分区"
            class="edit-category-select"
          >
            <el-option
              v-for="category in manuscriptCategoryOptions"
              :key="category.value"
              :label="category.label"
              :value="category.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="标签">
          <div class="edit-tags">
            <div class="edit-tag-input-row">
              <el-input
                v-model="editManuscriptTagInput"
                placeholder="输入标签后回车"
                maxlength="20"
                @keydown="handleEditManuscriptTagKeydown"
              >
                <template #append>
                  <el-button
                    @click="addEditManuscriptTag"
                    :disabled="!editManuscriptTagInput || editManuscriptForm.tags.length >= 10"
                  >
                    添加
                  </el-button>
                </template>
              </el-input>
              <span class="edit-tag-count">{{ editManuscriptForm.tags.length }}/10</span>
            </div>
            <div class="edit-tag-list">
              <el-tag
                v-for="tag in editManuscriptForm.tags"
                :key="tag"
                closable
                effect="plain"
                @close="removeEditManuscriptTag(tag)"
              >
                {{ tag }}
              </el-tag>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="简介">
          <el-input
            v-model="editManuscriptForm.description"
            type="textarea"
            :rows="5"
            maxlength="2000"
            show-word-limit
            placeholder="请输入稿件简介"
          />
        </el-form-item>

        <el-form-item label="分P">
          <div class="edit-video-parts">
            <div
              v-for="(part, index) in editManuscriptForm.videos"
              :key="part.id"
              class="edit-video-part"
            >
              <div class="edit-video-order">P{{ index + 1 }}</div>
              <div class="edit-video-main">
                <el-input
                  v-model="part.title"
                  maxlength="80"
                  show-word-limit
                  placeholder="请输入分P标题"
                />
                <div class="edit-video-file-row">
                  <el-upload
                    action="#"
                    :show-file-list="false"
                    :auto-upload="false"
                    :on-change="file => handleEditVideoFileChange(file, index)"
                    accept="video/*"
                  >
                    <el-button size="small" :icon="UploadFilled">替换视频</el-button>
                  </el-upload>
                  <span class="edit-video-file-name">
                    {{ part.replacementFileName || '未替换源视频' }}
                  </span>
                  <el-button
                    v-if="part.replacementFileName"
                    text
                    size="small"
                    type="danger"
                    @click="clearEditVideoReplacement(index)"
                  >
                    移除
                  </el-button>
                </div>
              </div>
              <div class="edit-video-actions">
                <el-button
                  text
                  size="small"
                  :icon="ArrowUp"
                  :disabled="index === 0"
                  @click="moveEditVideoPart(index, -1)"
                />
                <el-button
                  text
                  size="small"
                  :icon="ArrowDown"
                  :disabled="index === editManuscriptForm.videos.length - 1"
                  @click="moveEditVideoPart(index, 1)"
                />
              </div>
            </div>
            <el-empty
              v-if="editManuscriptForm.videos.length === 0"
              description="暂无分P数据"
              :image-size="80"
            />
          </div>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submitEditManuscript" :loading="updatingManuscript">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { manuscriptApi } from '@/api/creator'
import { categoryApi } from '@/api/client'
import { toWebP } from '@/utils/toWebP'
import { Picture, UploadFilled, ArrowUp, ArrowDown } from '@element-plus/icons-vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  manuscriptId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['update:visible', 'saved'])

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const manuscriptCategoryOptions = ref([])
const editManuscriptLoading = ref(false)
const updatingManuscript = ref(false)
const editManuscriptTagInput = ref('')
const editManuscriptCoverPreview = ref('')
const editManuscriptForm = reactive({
  id: null,
  title: '',
  description: '',
  categoryId: null,
  tags: [],
  videos: [],
  cover: null
})

const loadManuscriptCategories = async () => {
  if (manuscriptCategoryOptions.value.length > 0) {
    return
  }
  const response = await categoryApi.getCategoryList()
  const list = Array.isArray(response) ? response : (Array.isArray(response?.data) ? response.data : [])
  if (!list.length) {
    throw new Error('获取分区列表失败')
  }
  manuscriptCategoryOptions.value = list.map(category => ({
    value: category.id,
    label: category.name
  }))
}

const editArticle = async (id) => {
  editManuscriptLoading.value = true
  editManuscriptTagInput.value = ''
  try {
    await loadManuscriptCategories()
    const response = await manuscriptApi.getMyManuscriptById(id)
    if (response.code !== 200 || !response.data) {
      throw new Error(response.message || '获取稿件详情失败')
    }
    const manuscript = response.data
    editManuscriptForm.id = manuscript.id
    editManuscriptForm.title = manuscript.title || ''
    editManuscriptForm.description = manuscript.description || ''
    editManuscriptForm.categoryId = manuscript.categoryId || null
    editManuscriptForm.tags = Array.isArray(manuscript.tags) ? [...manuscript.tags] : []
    editManuscriptForm.videos = Array.isArray(manuscript.videos)
      ? manuscript.videos
        .map((video, index) => ({
          id: video.id,
          title: video.title || `P${index + 1}`,
          videoOrder: video.videoOrder ?? index,
          durationSeconds: video.durationSeconds || 0,
          replacementFile: null,
          replacementFileName: ''
        }))
        .sort((a, b) => a.videoOrder - b.videoOrder)
      : []
    normalizeEditVideoOrders()
    editManuscriptForm.cover = null
    editManuscriptCoverPreview.value = manuscript.coverUrl || ''
  } catch (error) {
    console.error('获取稿件详情失败:', error)
    ElMessage.error(error.message || '获取稿件详情失败')
    dialogVisible.value = false
  } finally {
    editManuscriptLoading.value = false
  }
}

const handleEditManuscriptCoverChange = async (file) => {
  if (!file.raw) {
    ElMessage.error('封面文件读取失败')
    return false
  }
  const isImage = file.raw.type.startsWith('image/')
  const isLt10M = file.raw.size / 1024 / 1024 < 10

  if (!isImage) {
    ElMessage.error('封面只能是图片格式')
    return false
  }
  if (!isLt10M) {
    ElMessage.error('封面大小不能超过 10MB')
    return false
  }

  editManuscriptForm.cover = await toWebP(file.raw)
  editManuscriptCoverPreview.value = URL.createObjectURL(editManuscriptForm.cover)
  return false
}

const getEditVideoDuration = (file) => {
  return new Promise((resolve) => {
    const video = document.createElement('video')
    video.preload = 'metadata'
    video.onloadedmetadata = () => {
      URL.revokeObjectURL(video.src)
      resolve(Math.floor(video.duration || 0))
    }
    video.onerror = () => resolve(0)
    video.src = URL.createObjectURL(file)
  })
}

const normalizeEditVideoOrders = () => {
  editManuscriptForm.videos.forEach((video, index) => {
    video.videoOrder = index
  })
}

const moveEditVideoPart = (index, direction) => {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= editManuscriptForm.videos.length) {
    return
  }
  const current = editManuscriptForm.videos[index]
  editManuscriptForm.videos[index] = editManuscriptForm.videos[targetIndex]
  editManuscriptForm.videos[targetIndex] = current
  normalizeEditVideoOrders()
}

const handleEditVideoFileChange = async (file, index) => {
  const part = editManuscriptForm.videos[index]
  if (!part || !file.raw) {
    ElMessage.error('视频文件读取失败')
    return false
  }
  const isVideo = file.raw.type.startsWith('video/')
  const isLt4G = file.raw.size / 1024 / 1024 / 1024 < 4

  if (!isVideo) {
    ElMessage.error('只能选择视频文件')
    return false
  }
  if (!isLt4G) {
    ElMessage.error('视频大小不能超过 4GB')
    return false
  }

  const durationSeconds = await getEditVideoDuration(file.raw)
  part.replacementFile = file.raw
  part.replacementFileName = file.name
  part.durationSeconds = durationSeconds
  ElMessage.success(`已选择替换文件：${file.name}`)
  return false
}

const clearEditVideoReplacement = (index) => {
  const part = editManuscriptForm.videos[index]
  if (!part) return
  part.replacementFile = null
  part.replacementFileName = ''
}

const addEditManuscriptTag = () => {
  const tag = editManuscriptTagInput.value.trim()
  if (!tag) return
  if (editManuscriptForm.tags.length >= 10) {
    ElMessage.warning('最多只能添加10个标签')
    return
  }
  if (editManuscriptForm.tags.includes(tag)) {
    ElMessage.warning('标签已存在')
    return
  }
  editManuscriptForm.tags.push(tag)
  editManuscriptTagInput.value = ''
}

const removeEditManuscriptTag = (tag) => {
  editManuscriptForm.tags = editManuscriptForm.tags.filter(item => item !== tag)
}

const handleEditManuscriptTagKeydown = (event) => {
  if (event.key === 'Enter') {
    event.preventDefault()
    addEditManuscriptTag()
  }
}

const submitEditManuscript = async () => {
  if (!editManuscriptForm.title.trim()) {
    ElMessage.warning('请输入稿件标题')
    return
  }
  if (!editManuscriptForm.categoryId) {
    ElMessage.warning('请选择稿件分区')
    return
  }

  updatingManuscript.value = true
  try {
    const response = await manuscriptApi.updateManuscript(editManuscriptForm.id, {
      title: editManuscriptForm.title.trim(),
      description: editManuscriptForm.description,
      categoryId: editManuscriptForm.categoryId,
      tags: editManuscriptForm.tags,
      videos: editManuscriptForm.videos.map(video => ({
        id: video.id,
        title: video.title,
        videoOrder: video.videoOrder,
        durationSeconds: video.durationSeconds,
        file: video.replacementFile
      })),
      cover: editManuscriptForm.cover
    })
    if (response.code === 200) {
      ElMessage.success('稿件已更新，等待重新审核')
      dialogVisible.value = false
      emit('saved')
    } else {
      ElMessage.error(response.message || '更新稿件失败')
    }
  } catch (error) {
    console.error('更新稿件失败:', error)
    ElMessage.error(error.message || '更新稿件失败')
  } finally {
    updatingManuscript.value = false
  }
}

// 打开对话框时加载稿件数据
watch(
  () => [props.visible, props.manuscriptId],
  ([visible, manuscriptId]) => {
    if (visible && manuscriptId) {
      editArticle(manuscriptId)
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.edit-manuscript-alert {
  margin-bottom: 18px;
}

.edit-manuscript-form {
  padding-right: 8px;
}

.edit-cover-row {
  display: flex;
  align-items: flex-end;
  gap: 14px;
}

.edit-cover-box {
  width: 192px;
  height: 108px;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  background: #f5f7fa;
}

.edit-cover-box img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.edit-cover-empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #909399;
  font-size: 13px;
}

.edit-cover-empty .el-icon {
  font-size: 24px;
}

.edit-cover-tip {
  color: #909399;
  font-size: 12px;
  line-height: 20px;
}

.edit-category-select {
  width: 100%;
}

.edit-tags {
  width: 100%;
}

.edit-tag-input-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.edit-tag-count {
  flex-shrink: 0;
  color: #909399;
  font-size: 12px;
}

.edit-tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.edit-video-parts {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.edit-video-part {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) 72px;
  gap: 10px;
  align-items: center;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fafafa;
}

.edit-video-order {
  width: 40px;
  height: 32px;
  border-radius: 4px;
  background: #e6f7ff;
  color: #1677a8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 13px;
}

.edit-video-main {
  min-width: 0;
}

.edit-video-file-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
  min-height: 28px;
}

.edit-video-file-name {
  min-width: 0;
  flex: 1;
  color: #909399;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.edit-video-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}
</style>