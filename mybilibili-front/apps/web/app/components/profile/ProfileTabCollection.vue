<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { collectionApi } from '@/api/collection.ts'
import ProfileCollectionGrid from './ProfileCollectionGrid.vue'
import ProfileCollectionDetail from './ProfileCollectionDetail.vue'
import ProfileCollectionEditDialog from './ProfileCollectionEditDialog.vue'
import ProfileAddVideoDialog from './ProfileAddVideoDialog.vue'

const props = defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Object,
    required: true
  }
})

const router = useRouter()

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 合集数据
const collections = ref({
  viewType: 'horizontal',
  items: [],
  loading: false
})

// 合集详情数据
const collectionDetail = ref({
  visible: false,
  collectionId: null,
  collection: null,
  manuscripts: [],
  loading: false,
  pagination: {
    currentPage: 1,
    pageSize: 20,
    total: 0
  },
  sortBy: 'default'
})

// 对话框可见性
const createCollectionDialogVisible = ref(false)
const editCollectionDialogVisible = ref(false)
const createCollectionVideoDialogVisible = ref(false)
const addVideoDialogVisible = ref(false)
const currentCollectionId = ref(null)

// 合集 API 返回 snake_case，归一化到 camelCase
const normalizeCollection = (d) => ({
  id: d.id,
  name: d.title || d.name || '',
  title: d.title || d.name || '',
  description: d.description || '',
  coverUrl: d.cover_url || d.coverUrl || d.cover || '',
  isPublic: d.status === 1 || d.isPublic === true || d.is_public === true,
  videoCount: d.manuscript_count || d.videoCount || 0,
  manuscriptCount: d.manuscript_count || d.manuscriptCount || 0,
  viewCount: d.view_count || d.viewCount || 0,
  createTime: d.created_at || d.createTime || '',
  updateTime: d.updated_at || d.updateTime || '',
  updatedAt: d.updated_at || d.updateTime || '',
  userId: d.user_id || d.userId || null,
  videos: []
})

// 稿件归一化
const normalizeManuscript = (m) => ({
  id: m.id,
  title: m.title || '',
  coverUrl: m.cover_url || m.coverUrl || m.cover || '',
  viewCount: m.view_count || m.viewCount || m.view_count || 0,
  uploadTime: m.upload_time || m.uploadTime || m.created_at || m.createTime || '',
  date: formatDate(m.upload_time || m.uploadTime || m.created_at || m.createTime || ''),
  duration: m.duration || '00:00',
  userId: m.user_id || m.userId || 0,
  status: m.status || 0
})

// 加载用户的合集列表
const loadUserCollections = async () => {
  if (!props.userId) return

  collections.value.loading = true
  try {
    const response = await collectionApi.getUserCollections(props.userId, 1, 100)
    if (response.code === 200) {
      const list = (response.data?.list || response.data || []).map(normalizeCollection)
      for (const collection of list) {
        try {
          const videoResponse = await collectionApi.getCollectionManuscripts(collection.id, 1, 10)
          if (videoResponse.code === 200) {
            const videos = (videoResponse.data?.list || videoResponse.data || []).map(normalizeManuscript)
            collection.videos = videos
            if (videos.length > 0 && videos[0].coverUrl) {
              collection.coverUrl = videos[0].coverUrl
            }
          }
        } catch (e) {
          collection.videos = []
        }
      }
      collections.value.items = list
    }
  } catch (error) {
    console.error('获取合集列表失败:', error)
  } finally {
    collections.value.loading = false
  }
}

// 加载合集详情数据
const loadCollectionDetailData = async () => {
  if (!collectionDetail.value.collectionId) return

  collectionDetail.value.loading = true
  try {
    const collectionResponse = await collectionApi.getCollectionById(collectionDetail.value.collectionId)
    if (collectionResponse.code === 200 && collectionResponse.data) {
      collectionDetail.value.collection = normalizeCollection(collectionResponse.data)
    }

    const manuscriptsResponse = await collectionApi.getCollectionManuscripts(
      collectionDetail.value.collectionId,
      collectionDetail.value.pagination.currentPage,
      collectionDetail.value.pagination.pageSize
    )
    if (manuscriptsResponse.code === 200) {
      const manuscripts = (manuscriptsResponse.data?.list || manuscriptsResponse.data || []).map(normalizeManuscript)
      collectionDetail.value.manuscripts = manuscripts
      collectionDetail.value.pagination.total = manuscriptsResponse.data?.total || manuscripts.length
    }
  } catch (error) {
    console.error('获取合集详情失败:', error)
    ElMessage.error('获取合集详情失败')
  } finally {
    collectionDetail.value.loading = false
  }
}

// 视图切换
const goToCollectionDetail = (collectionId) => {
  collectionDetail.value.collectionId = collectionId
  collectionDetail.value.visible = true
  loadCollectionDetailData()
}

const backToCollectionsList = () => {
  collectionDetail.value.visible = false
  collectionDetail.value.collectionId = null
  collectionDetail.value.collection = null
  collectionDetail.value.manuscripts = []
}

// 打开对话框
const openCreateCollectionDialog = () => {
  createCollectionDialogVisible.value = true
}

const openEditCollectionDialog = () => {
  editCollectionDialogVisible.value = true
}

const openAddVideoDialog = (collectionId) => {
  currentCollectionId.value = collectionId
  addVideoDialogVisible.value = true
}

const openAddVideoToCollectionDialog = () => {
  openAddVideoDialog(collectionDetail.value.collectionId)
}

// 播放
const playManuscript = (manuscript) => {
  if (manuscript.id) {
    router.push(`/manuscript/${manuscript.id}`)
  }
}

const playCollectionAll = () => {
  if (collectionDetail.value.manuscripts.length > 0) {
    router.push(`/manuscript/${collectionDetail.value.manuscripts[0].id}`)
  } else {
    ElMessage.info('该合集暂无视频')
  }
}

const playCollectionAllFromList = (collection) => {
  if (collection.videos && collection.videos.length > 0) {
    router.push(`/manuscript/${collection.videos[0].id}`)
  } else {
    ElMessage.info('该合集暂无视频')
  }
}

// 删除合集
const deleteCollection = async (collectionId) => {
  try {
    const response = await collectionApi.deleteCollection(collectionId)
    if (response.code === 200) {
      ElMessage.success('删除成功')
      backToCollectionsList()
      loadUserCollections()
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    console.error('删除合集失败:', error)
    ElMessage.error('删除失败')
  }
}

// 从合集中移除视频
const removeVideoFromCollection = async (manuscriptId) => {
  try {
    const response = await collectionApi.removeManuscriptFromCollection(
      collectionDetail.value.collectionId,
      manuscriptId
    )
    if (response.code === 200) {
      ElMessage.success('移除成功')
      loadCollectionDetailData()
    } else {
      ElMessage.error(response.message || '移除失败')
    }
  } catch (error) {
    console.error('移除视频失败:', error)
    ElMessage.error('移除失败')
  }
}

// 详情排序变化
const handleCollectionDetailSortChange = (value) => {
  collectionDetail.value.sortBy = value
  if (value === 'newest') {
    collectionDetail.value.manuscripts.sort((a, b) => new Date(b.uploadTime) - new Date(a.uploadTime))
  } else if (value === 'oldest') {
    collectionDetail.value.manuscripts.sort((a, b) => new Date(a.uploadTime) - new Date(b.uploadTime))
  } else {
    loadCollectionDetailData()
  }
}

// 合集或创建更新后刷新
const handleCollectionChanged = () => {
  loadCollectionDetailData()
  loadUserCollections()
}

const loadData = () => {
  loadUserCollections()
}

onMounted(() => {
  loadData()
})

watch(() => props.userId, () => {
  loadData()
})
</script>

<template>
  <div class="collections-section">
    <!-- 合集详情视图 -->
    <ProfileCollectionDetail
      v-if="collectionDetail.visible"
      :collection-detail="collectionDetail"
      :is-own-space="isOwnSpace"
      @back="backToCollectionsList"
      @edit="openEditCollectionDialog"
      @delete-collection="deleteCollection"
      @add-video="openAddVideoToCollectionDialog"
      @sort-change="handleCollectionDetailSortChange"
      @play-manuscript="playManuscript"
      @play-all="playCollectionAll"
      @remove-video="removeVideoFromCollection"
    />

    <!-- 合集列表视图 -->
    <ProfileCollectionGrid
      v-else
      :collections="collections"
      :is-own-space="isOwnSpace"
      @create-collection="openCreateCollectionDialog"
      @view-change="(type) => { collections.viewType = type }"
      @view-detail="goToCollectionDetail"
      @add-video="openAddVideoDialog"
      @play-all="playCollectionAllFromList"
    />

    <!-- 编辑/新建合集对话框 -->
    <ProfileCollectionEditDialog
      v-model:create-visible="createCollectionDialogVisible"
      v-model:edit-visible="editCollectionDialogVisible"
      v-model:create-video-visible="createCollectionVideoDialogVisible"
      :edit-collection="collectionDetail.collection"
      @created="loadUserCollections"
      @updated="handleCollectionChanged"
    />

    <!-- 添加视频到合集对话框 -->
    <ProfileAddVideoDialog
      v-model:visible="addVideoDialogVisible"
      :collection-id="currentCollectionId"
      @updated="loadUserCollections"
    />
  </div>
</template>

<style scoped>
.collections-section {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  border-radius: 8px;
}
</style>