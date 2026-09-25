<script setup>
import { safeStorage } from '@/utils/safeStorage'
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Search, ArrowDown, View } from '@element-plus/icons-vue'
import { collectionApi } from '@/api/collection.ts'
import { videoApi } from '@/api/client'

const props = defineProps({
  createVisible: {
    type: Boolean,
    default: false
  },
  editVisible: {
    type: Boolean,
    default: false
  },
  createVideoVisible: {
    type: Boolean,
    default: false
  },
  editCollection: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:createVisible', 'update:editVisible', 'update:createVideoVisible', 'created', 'updated'])

const createModel = computed({
  get: () => props.createVisible,
  set: (val) => emit('update:createVisible', val)
})
const editModel = computed({
  get: () => props.editVisible,
  set: (val) => emit('update:editVisible', val)
})
const createVideoModel = computed({
  get: () => props.createVideoVisible,
  set: (val) => emit('update:createVideoVisible', val)
})

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 获取默认封面
const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}

// 新建合集表单
const createCollectionForm = ref({
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const creatingCollection = ref(false)

// 编辑合集表单
const editCollectionForm = ref({
  id: null,
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const updatingCollection = ref(false)

// 新建合集选择视频
const createCollectionSelectedVideos = ref([])
const userVideos = ref([])
const addVideoSearchKeyword = ref('')
const addVideoSortBy = ref('newest')
const addVideoLoading = ref(false)

// 初始化创建表单
watch(() => props.createVisible, (val) => {
  if (val) {
    createCollectionForm.value = {
      name: '',
      description: '',
      cover: null,
      coverUrl: '',
      isPublic: true
    }
    createCollectionSelectedVideos.value = []
  }
})

// 初始化编辑表单
watch(() => props.editVisible, (val) => {
  if (val && props.editCollection) {
    editCollectionForm.value = {
      id: props.editCollection.id,
      name: props.editCollection.title || '',
      description: props.editCollection.description || '',
      cover: null,
      coverUrl: props.editCollection.coverUrl || '',
      isPublic: props.editCollection.status === 1
    }
  }
})

// 处理编辑合集封面上传
const handleEditCollectionCoverChange = (file) => {
  const isJPG = file.raw.type === 'image/jpeg'
  const isPNG = file.raw.type === 'image/png'
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isJPG && !isPNG) {
    ElMessage.error('封面图片只能是 JPG 或 PNG 格式!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('封面图片大小不能超过 2MB!')
    return false
  }

  editCollectionForm.value.cover = file.raw
  editCollectionForm.value.coverUrl = URL.createObjectURL(file.raw)
  return false
}

// 更新合集
const handleUpdateCollection = async () => {
  if (!editCollectionForm.value.name.trim()) {
    ElMessage.warning('请输入合集名称')
    return
  }

  updatingCollection.value = true
  try {
    const response = await collectionApi.updateCollection(editCollectionForm.value.id, {
      name: editCollectionForm.value.name,
      description: editCollectionForm.value.description,
      cover: editCollectionForm.value.cover,
      isPublic: editCollectionForm.value.isPublic
    })

    if (response.code === 200) {
      ElMessage.success('更新成功')
      emit('update:editVisible', false)
      emit('updated')
    } else {
      ElMessage.error(response.message || '更新失败')
    }
  } catch (error) {
    console.error('更新合集失败:', error)
    ElMessage.error('更新失败')
  } finally {
    updatingCollection.value = false
  }
}

// 处理合集封面上传
const handleCollectionCoverChange = (file) => {
  const isJPG = file.raw.type === 'image/jpeg'
  const isPNG = file.raw.type === 'image/png'
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isJPG && !isPNG) {
    ElMessage.error('封面图片只能是 JPG 或 PNG 格式!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('封面图片大小不能超过 2MB!')
    return false
  }

  createCollectionForm.value.cover = file.raw
  createCollectionForm.value.coverUrl = URL.createObjectURL(file.raw)
  return false
}

// 创建合集
const handleCreateCollection = async () => {
  if (!createCollectionForm.value.name.trim()) {
    ElMessage.warning('请输入合集名称')
    return
  }

  creatingCollection.value = true
  try {
    const response = await collectionApi.createCollection({
      name: createCollectionForm.value.name,
      description: createCollectionForm.value.description,
      cover: createCollectionForm.value.cover,
      isPublic: createCollectionForm.value.isPublic,
      manuscriptIds: createCollectionSelectedVideos.value
    })

    if (response.code === 200 && response.data?.id) {
      ElMessage.success('创建成功')
      emit('update:createVisible', false)
      createCollectionSelectedVideos.value = []
      emit('created')
    } else {
      ElMessage.error(response.message || '创建失败')
    }
  } catch (error) {
    console.error('创建合集失败:', error)
    ElMessage.error('创建合集失败')
  } finally {
    creatingCollection.value = false
  }
}

// 加载用户视频供新建合集选择
const loadUserVideosForCreateCollection = async () => {
  addVideoLoading.value = true
  try {
    const currentUserId = JSON.parse(safeStorage.getItem('user') || '{}')?.id
    const response = await videoApi.getVideosByUserId(currentUserId, 'latest', 3)
    if (response.code === 200) {
      userVideos.value = (response.data || []).map(video => ({
        ...video,
        date: formatDate(video.uploadTime)
      }))
    }
  } catch (error) {
    console.error('获取用户视频失败:', error)
  } finally {
    addVideoLoading.value = false
  }
}

// 打开新建合集选择视频对话框
const openCreateCollectionVideoDialog = async () => {
  if (!createCollectionForm.value.name.trim()) {
    ElMessage.warning('请先输入合集名称')
    return
  }
  createCollectionSelectedVideos.value = []
  addVideoSearchKeyword.value = ''
  addVideoSortBy.value = 'newest'
  await loadUserVideosForCreateCollection()
  emit('update:createVideoVisible', true)
}

// 搜索视频
const handleSearchVideos = () => {
  if (!addVideoSearchKeyword.value.trim()) {
    loadUserVideosForCreateCollection()
    return
  }

  const keyword = addVideoSearchKeyword.value.toLowerCase()
  loadUserVideosForCreateCollection().then(() => {
    userVideos.value = userVideos.value.filter(v =>
      v.title.toLowerCase().includes(keyword)
    )
  })
}

// 处理排序变化
const handleVideoSortChange = (sortType) => {
  addVideoSortBy.value = sortType
  if (sortType === 'newest') {
    userVideos.value.sort((a, b) => new Date(b.uploadTime) - new Date(a.uploadTime))
  } else if (sortType === 'oldest') {
    userVideos.value.sort((a, b) => new Date(a.uploadTime) - new Date(b.uploadTime))
  }
}
</script>

<template>
  <!-- 编辑合集对话框 -->
  <el-dialog
    v-model="editModel"
    title="编辑视频列表"
    width="500px"
    destroy-on-close
  >
    <el-form label-position="top">
      <el-form-item label="视频列表标题" required>
        <el-input
          v-model="editCollectionForm.name"
          placeholder="请输入视频列表标题"
          maxlength="10"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="视频列表简介">
        <el-input
          v-model="editCollectionForm.description"
          type="textarea"
          :rows="3"
          placeholder="请输入视频列表简介（选填）"
          maxlength="256"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="合集封面">
        <el-upload
          class="collection-cover-uploader"
          action="#"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleEditCollectionCoverChange"
          accept="image/jpeg,image/png"
        >
          <div v-if="editCollectionForm.coverUrl" class="cover-preview">
            <img loading="lazy" decoding="async" :src="editCollectionForm.coverUrl" />
          </div>
          <div v-else class="cover-placeholder">
            <el-icon :size="32"><Plus /></el-icon>
            <span>点击上传封面</span>
            <span class="cover-hint">支持 JPG、PNG 格式，最大 2MB</span>
          </div>
        </el-upload>
      </el-form-item>

      <el-form-item>
        <el-checkbox v-model="editCollectionForm.isPublic">
          公开合集
        </el-checkbox>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:editVisible', false)">取消</el-button>
      <el-button
        type="primary"
        :loading="updatingCollection"
        @click="handleUpdateCollection"
      >
        确定
      </el-button>
    </template>
  </el-dialog>

  <!-- 新建合集对话框 -->
  <el-dialog
    v-model="createModel"
    title="新建合集"
    width="500px"
    destroy-on-close
  >
    <el-form label-position="top">
      <el-form-item label="合集名称" required>
        <el-input
          v-model="createCollectionForm.name"
          placeholder="请输入合集名称"
          maxlength="50"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="合集描述">
        <el-input
          v-model="createCollectionForm.description"
          type="textarea"
          :rows="3"
          placeholder="请输入合集描述（选填）"
          maxlength="200"
          show-word-limit
        />
      </el-form-item>

      <el-form-item>
        <el-checkbox v-model="createCollectionForm.isPublic">
          公开合集
        </el-checkbox>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:createVisible', false)">取消</el-button>
      <el-button
        type="primary"
        :loading="creatingCollection"
        @click="openCreateCollectionVideoDialog"
      >
        下一步：选择视频
      </el-button>
    </template>
  </el-dialog>

  <!-- 新建合集 - 选择视频对话框 -->
  <el-dialog
    v-model="createVideoModel"
    title="选择视频"
    width="700px"
    destroy-on-close
  >
    <div class="video-selection-header">
      <span class="selection-title">你还可以添加 <strong>1000</strong> 个视频</span>
    </div>

    <div class="video-search-bar">
      <el-input
        v-model="addVideoSearchKeyword"
        placeholder="搜索我上传的视频"
        :prefix-icon="Search"
        clearable
        @keyup.enter="handleSearchVideos"
        @clear="handleSearchVideos"
      />
      <el-button type="primary" @click="handleSearchVideos">搜索</el-button>
    </div>

    <div class="video-list-header">
      <span class="header-left">我的视频列表</span>
      <el-dropdown @command="handleVideoSortChange">
        <span class="header-right">
          {{ addVideoSortBy === 'newest' ? '最新发布' : '最早发布' }}
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="newest">最新发布</el-dropdown-item>
            <el-dropdown-item command="oldest">最早发布</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div v-if="addVideoLoading" class="dialog-loading">
      <el-skeleton :rows="5" animated />
    </div>

    <div v-else-if="userVideos.length === 0" class="dialog-empty">
      <el-empty description="暂无视频" />
    </div>

    <div v-else class="video-selection-list">
      <el-checkbox-group v-model="createCollectionSelectedVideos">
        <div
          v-for="video in userVideos"
          :key="video.id"
          class="video-selection-item"
        >
          <el-checkbox :value="video.id">
            <div class="checkbox-content">
              <img loading="lazy" decoding="async"
                :src="video.coverUrl || getDefaultCover()"
                :alt="video.title"
                class="checkbox-cover"
              />
              <div class="checkbox-info">
                <span class="checkbox-title" :title="video.title">{{ video.title }}</span>
                <div class="checkbox-meta">
                  <span class="meta-views">
                    <el-icon><View /></el-icon>
                    {{ video.viewCount || 0 }}
                  </span>
                  <span class="meta-date">{{ video.date }}</span>
                </div>
              </div>
            </div>
          </el-checkbox>
        </div>
      </el-checkbox-group>
    </div>

    <template #footer>
      <el-button @click="emit('update:createVideoVisible', false)">取消</el-button>
      <el-button
        type="primary"
        :loading="creatingCollection"
        @click="handleCreateCollection"
      >
        确定 ({{ createCollectionSelectedVideos.length }})
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
/* 封面上传样式 */
.collection-cover-uploader {
  width: 100%;
}

.collection-cover-uploader .cover-preview {
  width: 100%;
  height: 160px;
  border-radius: 8px;
  overflow: hidden;
}

.collection-cover-uploader .cover-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-cover-uploader .cover-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 160px;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  color: #909399;
  cursor: pointer;
  transition: all 0.3s ease;
}

.collection-cover-uploader .cover-placeholder:hover {
  border-color: #00aeec;
  color: #00aeec;
}

.collection-cover-uploader .cover-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

/* 视频选择对话框样式 */
.video-selection-header {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.selection-title {
  font-size: 14px;
  color: #666;
}

.selection-title strong {
  color: #00aeec;
  font-size: 16px;
}

.video-search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.video-search-bar .el-input {
  flex: 1;
}

.video-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.header-left {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.header-right {
  font-size: 14px;
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}

.dialog-loading {
  padding: 40px 0;
}

.dialog-empty {
  padding: 40px 0;
}

.video-selection-list {
  max-height: 400px;
  overflow-y: auto;
}

.video-selection-item {
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.video-selection-item:last-child {
  border-bottom: none;
}

.video-selection-item .el-checkbox {
  width: 100%;
  flex: 1;
}

.video-selection-item .el-checkbox__label {
  flex: 1;
  padding-left: 12px;
}

.checkbox-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.checkbox-cover {
  width: 120px;
  height: 68px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}

.checkbox-info {
  flex: 1;
  min-width: 0;
}

.checkbox-title {
  font-size: 14px;
  color: #333;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 8px;
}

.checkbox-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: #999;
}

.meta-views {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>