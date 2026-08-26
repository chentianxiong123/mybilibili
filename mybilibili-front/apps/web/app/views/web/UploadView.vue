<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { categoryApi } from '@/api/client'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, VideoPlay, Document } from '@element-plus/icons-vue'
import { useChunkedManuscriptUpload } from '@/composables/useChunkedManuscriptUpload.ts'
import { hasAuthSession } from '@/utils/auth.ts'
import ManuscriptForm from '@/components/ManuscriptForm.vue'
import CoverUploader from '@/components/CoverUploader.vue'
import VideoPartUploader from '@/components/VideoPartUploader.vue'
import UploadProgressBar from '@/components/UploadProgressBar.vue'
import { saveDraft as saveLocalDraft, getDraft, listDrafts } from '@/utils/drafts'

const router = useRouter()

const uploadForm = reactive({
  title: '',
  categoryId: null,
  tags: [],
  description: '',
  coverFile: null,
  type: 'original'
})

const videoParts = ref([])
const categories = ref([])
const currentDraftId = ref('')
const draftCount = ref(0)

const loadCategories = async () => {
  try {
    const res = await categoryApi.getCategoryList()
    const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : [])
    if (list.length) {
      categories.value = list.map(category => ({
        value: category.id,
        label: category.name
      }))
    } else {
      ElMessage.warning('获取分区列表失败，使用默认分区')
      categories.value = [
        { value: 1, label: '动画' },
        { value: 2, label: '音乐' },
        { value: 3, label: '舞蹈' },
        { value: 4, label: '游戏' },
        { value: 5, label: '知识' },
        { value: 6, label: '资讯' },
        { value: 7, label: '美食' },
        { value: 8, label: '生活' }
      ]
    }
  } catch (error) {
    console.error('获取分区列表失败:', error)
    ElMessage.warning('获取分区列表失败，使用默认分区')
    categories.value = [
      { value: 1, label: '动画' },
      { value: 2, label: '音乐' },
      { value: 3, label: '舞蹈' },
      { value: 4, label: '游戏' },
      { value: 5, label: '知识' },
      { value: 6, label: '资讯' },
      { value: 7, label: '美食' },
      { value: 8, label: '生活' }
    ]
  }
}

const uploadRules = {
  title: [
    { required: true, message: '请输入稿件标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  categoryId: [
    { required: true, message: '请选择稿件分区', trigger: 'change' }
  ]
}

const manuscriptFormRef = ref()
const coverPreview = ref('')
const showUploadDialog = ref(false)
const isSubmittingRequest = ref(false)
const uploadProgress = ref(0)
const currentUploadingPart = ref('')

const {
  stage: uploadStage,
  stageLabel: uploadStageLabel,
  percentage: chunkedPercentage,
  uploadedBytes,
  totalBytes,
  speed: uploadSpeed,
  etaSeconds,
  partProgress: chunkPartProgress,
  error: uploadError,
  isUploading: isChunkedUploading,
  isFinished: isChunkedFinished,
  start: startChunkedUpload,
  cancel: cancelChunkedUpload,
  UPLOAD_STAGES
} = useChunkedManuscriptUpload()

const checkLoginStatus = () => {
  if (!hasAuthSession()) {
    ElMessage.warning('请先登录后再上传稿件')
    return false
  }
  return true
}

const handleCoverChange = (file) => {
  uploadForm.coverFile = file
}

const previewUpdate = (preview) => {
  coverPreview.value = preview
}

const handleAddVideo = (newPart) => {
  const part = {
    ...newPart,
    sortOrder: videoParts.value.length
  }
  videoParts.value.push(part)
  ElMessage.success(`已添加分P: ${newPart.title}`)
}

const handleRemovePart = (index) => {
  videoParts.value.splice(index, 1)
  videoParts.value.forEach((part, idx) => {
    part.sortOrder = idx
  })
}

const handleMovePartUp = (index) => {
  if (index === 0) return
  const temp = videoParts.value[index]
  videoParts.value[index] = videoParts.value[index - 1]
  videoParts.value[index - 1] = temp
  videoParts.value.forEach((part, idx) => {
    part.sortOrder = idx
  })
}

const handleMovePartDown = (index) => {
  if (index === videoParts.value.length - 1) return
  const temp = videoParts.value[index]
  videoParts.value[index] = videoParts.value[index + 1]
  videoParts.value[index + 1] = temp
  videoParts.value.forEach((part, idx) => {
    part.sortOrder = idx
  })
}

const handleSubmit = () => {
  if (!checkLoginStatus()) return
  if (videoParts.value.length === 0) {
    ElMessage.warning('请至少添加一个视频分P')
    return
  }

  manuscriptFormRef.value.validate((valid) => {
    if (!valid) return false

    if (!uploadForm.coverFile) {
      ElMessage.error('请上传稿件封面')
      return
    }

    const invalidParts = videoParts.value.filter(part => !part.file)
    if (invalidParts.length > 0) {
      ElMessage.error('部分视频文件缺失，请重新上传')
      return
    }

    const manuscriptData = {
      title: uploadForm.title,
      description: uploadForm.description,
      cover: uploadForm.coverFile,
      categoryId: uploadForm.categoryId,
      tags: uploadForm.tags,
      videos: videoParts.value.map((part, index) => ({
        file: part.file,
        title: part.title,
        sortOrder: index,
        durationSeconds: part.duration || 0
      }))
    }

    showUploadDialog.value = true
    isSubmittingRequest.value = true

    startChunkedUpload(manuscriptData)
      .then(response => {
        isSubmittingRequest.value = false
        ElMessage.success('投稿成功，已进入审核/处理中队列')
        router.push({
          path: '/create-center/content-articles',
          query: { status: 'processing' }
        })
      })
      .catch(error => {
        console.error('上传错误:', error)
        if (error.message !== 'cancelled') {
          ElMessage.error(uploadError.value || error.message || '上传失败')
        }
        isSubmittingRequest.value = false
      })
  })
}

const handleCancelUpload = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要取消上传吗？已上传的分片将被清除。',
      '提示',
      { confirmButtonText: '确定取消', cancelButtonText: '继续上传', type: 'warning' }
    )
    await cancelChunkedUpload()
    isSubmittingRequest.value = false
    showUploadDialog.value = false
    ElMessage.info('上传已取消')
  } catch {}
}

const saveDraft = () => {
  if (!uploadForm.title) {
    ElMessage.warning('请先填写稿件标题再保存草稿')
    return
  }
  const videoPartsMeta = videoParts.value.map((part, index) => ({
    title: part.title || `P${index + 1}`,
    sortOrder: index,
    // File 对象无法序列化，仅保存元信息；继续编辑时需重新选择文件
    hasLocalFile: !!part.file
  }))
  const saved = saveLocalDraft({
    id: currentDraftId.value || undefined,
    title: uploadForm.title,
    categoryId: uploadForm.categoryId,
    tags: uploadForm.tags,
    description: uploadForm.description,
    type: uploadForm.type,
    coverPreview: coverPreview.value || undefined,
    videoParts: videoPartsMeta,
    hasLocalVideoFiles: videoParts.value.some((p) => p.file)
  })
  if (saved) {
    currentDraftId.value = saved.id
    ElMessage.success('草稿已保存到本地，可在草稿箱中继续编辑')
  } else {
    ElMessage.error('草稿保存失败')
  }
}

const goDraftsBox = () => {
  router.push('/create-center/drafts')
}

const cancelUpload = () => {
  if (videoParts.value.length > 0 || uploadForm.title || uploadForm.description) {
    ElMessageBox.confirm(
      '确定要取消上传吗？已填写的内容将不会保存。',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      router.go(-1)
    }).catch(() => {})
  } else {
    router.go(-1)
  }
}

onMounted(() => {
  loadCategories()
  draftCount.value = listDrafts().length
  // 从 URL 的 draftId 恢复草稿
  const params = new URLSearchParams(window.location.search)
  const draftId = params.get('draftId')
  if (draftId) {
    const draft = getDraft(draftId)
    if (draft) {
      currentDraftId.value = draft.id
      uploadForm.title = draft.title || ''
      uploadForm.categoryId = draft.categoryId
      uploadForm.tags = Array.isArray(draft.tags) ? draft.tags : []
      uploadForm.description = draft.description || ''
      uploadForm.type = draft.type || 'original'
      coverPreview.value = draft.coverPreview || ''
      // 恢复分P标题（文件需重新选择）
      videoParts.value = (draft.videoParts || []).map((p, i) => ({
        id: 'dp_' + Date.now().toString(36) + '_' + i,
        title: p.title || `P${i + 1}`,
        sortOrder: i,
        file: null
      }))
      ElMessage.success('已恢复草稿，请重新选择视频文件后发布')
    } else {
      ElMessage.warning('草稿不存在或已删除')
    }
  }
})
</script>

<template>
  <div class="upload-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">上传稿件</h1>
      <p class="page-subtitle">支持多P视频上传，一个稿件可包含多个视频</p>
    </div>

    <!-- 信息卡片 -->
    <div class="info-cards">
      <div class="info-card">
        <el-icon class="info-card-icon"><VideoPlay /></el-icon>
        <div class="info-card-content">
          <h3 class="info-card-title">视频大小</h3>
          <p class="info-card-desc">单个视频4G以内，时长2小时以内</p>
        </div>
      </div>
      <div class="info-card">
        <el-icon class="info-card-icon"><Document /></el-icon>
        <div class="info-card-content">
          <h3 class="info-card-title">视频格式</h3>
          <p class="info-card-desc">推荐MP4/MOV/MKV格式，转码更快</p>
        </div>
      </div>
      <div class="info-card">
        <el-icon class="info-card-icon"><UploadFilled /></el-icon>
        <div class="info-card-content">
          <h3 class="info-card-title">多P支持</h3>
          <p class="info-card-desc">一个稿件支持多个视频分P</p>
        </div>
      </div>
    </div>

    <!-- 表单区域 -->
    <div class="form-area">
      <div class="cover-wrapper">
        <label class="cover-label">封面</label>
        <CoverUploader
          :cover-preview="coverPreview"
          @cover-change="handleCoverChange"
          @update:cover-preview="previewUpdate"
        />
      </div>
      <ManuscriptForm
        ref="manuscriptFormRef"
        :form="uploadForm"
        :categories="categories"
        :rules="uploadRules"
      />
    </div>

    <!-- 视频分P -->
    <VideoPartUploader
      :video-parts="videoParts"
      @add-video="handleAddVideo"
      @remove-part="handleRemovePart"
      @move-up="handleMovePartUp"
      @move-down="handleMovePartDown"
    />

    <!-- 底部操作按钮 -->
    <div class="form-actions">
      <el-button @click="cancelUpload" :disabled="isSubmittingRequest" size="large">取消</el-button>
      <el-button @click="saveDraft" :disabled="isSubmittingRequest" size="large">存草稿</el-button>
      <el-button @click="goDraftsBox" :disabled="isSubmittingRequest" size="large">草稿箱{{ draftCount ? ` (${draftCount})` : '' }}</el-button>
      <el-button
        type="primary"
        @click="handleSubmit"
        :loading="isSubmittingRequest"
        :disabled="isSubmittingRequest || videoParts.length === 0"
        size="large"
      >
        {{ isSubmittingRequest ? '上传中...' : '立即投稿' }}
      </el-button>
    </div>

    <!-- 上传进度对话框 -->
    <UploadProgressBar
      :show="showUploadDialog"
      :stage="uploadStage"
      :stage-label="uploadStageLabel"
      :percentage="chunkedPercentage"
      :uploaded-bytes="uploadedBytes"
      :total-bytes="totalBytes"
      :speed="uploadSpeed"
      :eta-seconds="etaSeconds"
      :part-progress="chunkPartProgress"
      :error="uploadError"
      :is-uploading="isChunkedUploading"
      :is-finished="isChunkedFinished"
      :is-submitting="isSubmittingRequest"
      :upload_stages="UPLOAD_STAGES"
      @close="showUploadDialog = false"
      @cancel="handleCancelUpload"
    />
  </div>
</template>

<style scoped>
.upload-container {
  width: 100%;
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.page-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.info-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.info-card {
  flex: 1;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.info-card:nth-child(2) {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  box-shadow: 0 4px 12px rgba(240, 147, 251, 0.3);
}

.info-card:nth-child(3) {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  box-shadow: 0 4px 12px rgba(79, 172, 254, 0.3);
}

.info-card-icon {
  font-size: 32px;
  opacity: 0.9;
}

.info-card-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 4px 0;
}

.info-card-desc {
  font-size: 13px;
  opacity: 0.9;
  margin: 0;
}

.form-area {
  background-color: #fff;
  border-radius: 12px;
  padding: 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.cover-wrapper {
  display: flex;
  align-items: flex-start;
  margin-bottom: 22px;
}

.cover-label {
  width: 100px;
  padding-right: 12px;
  font-size: 14px;
  color: #606266;
  text-align: right;
  flex-shrink: 0;
  line-height: 32px;
}

.form-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding: 24px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  position: sticky;
  bottom: 20px;
}


</style>