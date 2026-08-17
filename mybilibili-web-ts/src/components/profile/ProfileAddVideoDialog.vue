<script setup>
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, ArrowDown, View } from '@element-plus/icons-vue'
import { collectionApi } from '@/api/collection.ts'
import { videoApi } from '@/api/client'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  collectionId: {
    type: [String, Number],
    default: null
  }
})

const emit = defineEmits(['update:visible', 'updated'])

const visibleModel = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
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

const userVideos = ref([])
const selectedVideos = ref([])
const addVideoLoading = ref(false)
const addVideoSearchKeyword = ref('')
const addVideoSortBy = ref('newest')

// 初始化
watch(() => props.visible, (val) => {
  if (val) {
    addVideoSearchKeyword.value = ''
    addVideoSortBy.value = 'newest'
    selectedVideos.value = []
    loadUserVideosForSelection()
  }
})

// 加载用户视频供选择
const loadUserVideosForSelection = async () => {
  addVideoLoading.value = true
  try {
    // 获取当前合集已有的稿件ID列表
    const collectionResponse = await collectionApi.getCollectionManuscripts(props.collectionId, 1, 100)
    if (collectionResponse.code === 200) {
      const manuscripts = collectionResponse.data || []
      selectedVideos.value = manuscripts.map(m => m.id)
    }

    // 获取用户所有视频（稿件）
    const currentUserId = JSON.parse(localStorage.getItem('user') || '{}')?.id
    const response = await videoApi.getVideosByUserId(currentUserId, 'latest', 3)
    if (response.code === 200) {
      userVideos.value = (response.data || [])
        .map(video => ({
          ...video,
          date: formatDate(video.uploadTime)
        }))
    }
  } catch (error) {
    console.error('获取数据失败:', error)
  } finally {
    addVideoLoading.value = false
  }
}

// 搜索视频
const handleSearchVideos = () => {
  if (!addVideoSearchKeyword.value.trim()) {
    loadUserVideosForSelection()
    return
  }

  const keyword = addVideoSearchKeyword.value.toLowerCase()
  loadUserVideosForSelection().then(() => {
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

// 添加选中的视频到合集
const handleAddVideosToCollection = async () => {
  try {
    // 获取当前合集中已有的稿件
    const collectionResponse = await collectionApi.getCollectionManuscripts(props.collectionId, 1, 100)
    let existingManuscriptIds = []
    if (collectionResponse.code === 200) {
      existingManuscriptIds = (collectionResponse.data || []).map(m => m.id)
    }

    // 计算需要添加和移除的稿件
    const toAdd = selectedVideos.value.filter(id => !existingManuscriptIds.includes(id))
    const toRemove = existingManuscriptIds.filter(id => !selectedVideos.value.includes(id))

    // 执行添加操作
    const addPromises = toAdd.map((manuscriptId, index) => {
      return collectionApi.addManuscriptToCollection(
        props.collectionId,
        manuscriptId,
        existingManuscriptIds.length + index
      )
    })

    // 执行移除操作
    const removePromises = toRemove.map(manuscriptId => {
      return collectionApi.removeManuscriptFromCollection(
        props.collectionId,
        manuscriptId
      )
    })

    // 等待所有操作完成
    await Promise.all([...addPromises, ...removePromises])

    ElMessage.success('更新成功')
    emit('update:visible', false)
    emit('updated')
  } catch (error) {
    console.error('更新合集视频失败:', error)
    ElMessage.error('更新失败: ' + (error.response?.data?.message || error.message))
  }
}
</script>

<template>
  <el-dialog
    v-model="visibleModel"
    title="管理合集视频"
    width="700px"
    destroy-on-close
  >
    <div class="video-selection-header">
      <span class="selection-title">勾选视频添加到合集，取消勾选从合集移除</span>
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
      <el-checkbox-group v-model="selectedVideos">
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
          <span v-if="selectedVideos.includes(video.id)" class="in-collection-tag">已加入</span>
        </div>
      </el-checkbox-group>
    </div>

    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button
        type="primary"
        @click="handleAddVideosToCollection"
      >
        更新
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
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

.in-collection-tag {
  background: linear-gradient(135deg, #005980, #1890ff);
  color: #ffffff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  flex-shrink: 0;
  margin-left: 8px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
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