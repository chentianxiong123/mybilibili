<template>
  <div class="collection-management">
    <CollectionDetail
      v-if="collectionDetail.visible"
      :collection="collectionDetail.collection"
      :manuscripts="collectionDetail.manuscripts"
      :loading="collectionDetail.loading"
      :sort-by="collectionDetail.sortBy"
      :pagination="collectionDetail.pagination"
      @back="backToCollectionsList"
      @edit="openEditCollectionDialog"
      @play-all="playCollectionAll"
      @add-video="openAddVideoToCollectionDialog"
      @sort-change="handleCollectionDetailSortChange"
      @remove-video="handleRemoveVideoFromCollection"
      @delete="handleDeleteCollection"
      @play-video="playManuscript"
    />

    <CollectionList
      v-else
      :items="myCollections.items"
      :loading="myCollections.loading"
      v-model:view-type="myCollections.viewType"
      @open-create="openCreateCollectionDialog"
      @go-to-detail="goToCollectionDetail"
      @add-video="openAddVideoDialog"
      @play-all="playCollectionAllFromList"
      @play-video="playManuscript"
    />

    <CollectionCreateDialog
      v-model:visible="createCollectionDialogVisible"
      mode="create"
      :form="createCollectionForm"
      :submitting="creatingCollection"
      @submit="handleCreateCollection"
    />

    <CollectionCreateDialog
      v-model:visible="editCollectionDialogVisible"
      mode="edit"
      :form="editCollectionForm"
      :submitting="updatingCollection"
      @submit="handleUpdateCollection"
    />

    <CollectionAddVideoDialog
      v-model:visible="addVideoDialogVisible"
      v-model:selected-videos="selectedVideos"
      :available-videos="availableVideos"
      :loading="addingVideo"
      @submit="handleAddVideoToCollection"
    />
  </div>
</template>

<script setup>
import { safeStorage } from '@/utils/safeStorage'
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { manuscriptApi, collectionApi } from '@/api/creator'
import { useUserStore } from '@/stores/user'
import CollectionList from './CollectionList.vue'
import CollectionDetail from './CollectionDetail.vue'
import CollectionCreateDialog from './CollectionCreateDialog.vue'
import CollectionAddVideoDialog from './CollectionAddVideoDialog.vue'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()
const userStore = useUserStore()

const getCurrentUserId = () => {
  const userStr = safeStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user?.id
    } catch (e) {
      return userStore.userInfo.id
    }
  }
  return userStore.userInfo.id
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const myCollections = ref({
  viewType: 'horizontal',
  items: [],
  loading: false
})

const collectionDetail = ref({
  visible: false,
  collectionId: null,
  collection: null,
  manuscripts: [],
  loading: false,
  sortBy: 'default',
  pagination: {
    page: 1,
    size: 20,
    total: 0
  }
})

const createCollectionDialogVisible = ref(false)
const createCollectionForm = ref({
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const creatingCollection = ref(false)

const editCollectionDialogVisible = ref(false)
const editCollectionForm = ref({
  id: null,
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const updatingCollection = ref(false)

const addVideoDialogVisible = ref(false)
const addVideoCollectionId = ref(null)
const availableVideos = ref([])
const selectedVideos = ref([])
const addingVideo = ref(false)

// 合集 API 返回 snake_case，归一化到 camelCase
const normalizeCollection = (d) => ({
  id: d.id,
  title: d.title || '',
  name: d.title || d.name || '',
  description: d.description || '',
  coverUrl: d.cover_url || d.coverUrl || '',
  isPublic: d.status === 1 || d.isPublic === true,
  status: d.status,
  videoCount: d.manuscript_count || d.videoCount || 0,
  viewCount: d.view_count || d.viewCount || 0,
  createTime: d.created_at || d.createTime || '',
  updateTime: d.updated_at || d.updateTime || '',
  createdAt: d.created_at || d.createTime || '',
  updatedAt: d.updated_at || d.updateTime || '',
  userId: d.user_id || d.userId || null,
  userName: d.user_name || d.userName || '',
  userAvatar: d.user_avatar || d.userAvatar || ''
})

const loadUserCollections = async () => {
  const userId = getCurrentUserId()
  if (!userId) return

  myCollections.value.loading = true
  try {
    const response = await collectionApi.getUserCollections(userId, 1, 100)
    if (response.code === 200) {
      const rawList = Array.isArray(response.data) ? response.data : (response.data?.list || [])
      const list = rawList.map(normalizeCollection)
      for (const collection of list) {
        try {
          const videoResponse = await collectionApi.getCollectionManuscripts(collection.id, 1, 10)
          if (videoResponse.code === 200) {
            const rawVideos = Array.isArray(videoResponse.data) ? videoResponse.data : (videoResponse.data?.list || [])
            collection.videos = rawVideos.map(video => ({
              ...video,
              coverUrl: video.cover_url || video.coverUrl || '',
              uploadTime: video.upload_time || video.uploadTime,
              date: formatDate(video.upload_time || video.uploadTime)
            }))
            if (!collection.coverUrl && collection.videos.length > 0 && collection.videos[0].coverUrl) {
              collection.coverUrl = collection.videos[0].coverUrl
            }
          }
        } catch (e) {
          console.error('获取合集稿件失败:', e)
          collection.videos = []
        }
      }
      myCollections.value.items = list
    }
  } catch (error) {
    console.error('获取合集列表失败:', error)
  } finally {
    myCollections.value.loading = false
  }
}

const goToCollectionDetail = async (collectionId) => {
  collectionDetail.value.visible = true
  collectionDetail.value.collectionId = collectionId
  collectionDetail.value.loading = true

  try {
    const response = await collectionApi.getCollectionById(collectionId)
    if (response.code === 200) {
      collectionDetail.value.collection = normalizeCollection(response.data)
    }

    const videoResponse = await collectionApi.getCollectionManuscripts(collectionId, 1, 20)
    if (videoResponse.code === 200) {
      const rawVideos = Array.isArray(videoResponse.data) ? videoResponse.data : (videoResponse.data?.list || [])
      collectionDetail.value.manuscripts = rawVideos.map(video => ({
        ...video,
        coverUrl: video.cover_url || video.coverUrl || '',
        uploadTime: video.upload_time || video.uploadTime
      }))
      collectionDetail.value.pagination.total = videoResponse.data?.total ?? rawVideos.length
    }
  } catch (error) {
    console.error('获取合集详情失败:', error)
  } finally {
    collectionDetail.value.loading = false
  }
}

const backToCollectionsList = () => {
  collectionDetail.value.visible = false
  collectionDetail.value.collectionId = null
  collectionDetail.value.collection = null
  collectionDetail.value.manuscripts = []
}

const openCreateCollectionDialog = () => {
  createCollectionForm.value = {
    name: '',
    description: '',
    cover: null,
    coverUrl: '',
    isPublic: true
  }
  createCollectionDialogVisible.value = true
}

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
      isPublic: createCollectionForm.value.isPublic
    })
    if (response.code === 200) {
      ElMessage.success('创建成功')
      createCollectionDialogVisible.value = false
      await loadUserCollections()
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

const openEditCollectionDialog = () => {
  if (!collectionDetail.value.collection) return

  editCollectionForm.value = {
    id: collectionDetail.value.collectionId,
    name: collectionDetail.value.collection.title || '',
    description: collectionDetail.value.collection.description || '',
    cover: null,
    coverUrl: collectionDetail.value.collection.coverUrl || '',
    isPublic: collectionDetail.value.collection.status === 1
  }
  editCollectionDialogVisible.value = true
}

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
      editCollectionDialogVisible.value = false
      await goToCollectionDetail(editCollectionForm.value.id)
      await loadUserCollections()
    } else {
      ElMessage.error(response.message || '更新失败')
    }
  } catch (error) {
    console.error('更新合集失败:', error)
    ElMessage.error('更新合集失败')
  } finally {
    updatingCollection.value = false
  }
}

const handleDeleteCollection = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要删除这个合集吗？删除后无法恢复。',
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await collectionApi.deleteCollection(collectionDetail.value.collectionId)
    if (response.code === 200) {
      ElMessage.success('删除成功')
      backToCollectionsList()
      await loadUserCollections()
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除合集失败:', error)
      ElMessage.error('删除合集失败')
    }
  }
}

const handleCollectionDetailSortChange = (sort) => {
  collectionDetail.value.sortBy = sort
}

const openAddVideoToCollectionDialog = async () => {
  await openAddVideoDialog(collectionDetail.value.collectionId)
}

const openAddVideoDialog = async (collectionId) => {
  addVideoCollectionId.value = collectionId
  selectedVideos.value = []
  addingVideo.value = true

  try {
    const [manuscriptsRes, videosRes] = await Promise.all([
      collectionApi.getCollectionManuscripts(collectionId, 1, 100),
      manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    ])

    if (videosRes.code === 200) {
      const data = videosRes.data || {}
      const list = data.list || data.records || data.items || data.data || []
      availableVideos.value = list.map(v => ({
        ...v,
        id: v.manuscriptId || v.id,
        coverUrl: v.coverUrl || v.cover_url,
        viewCount: v.viewCount || v.view_count || 0,
        title: v.title || '无标题'
      }))
    } else {
      availableVideos.value = []
    }

    if (manuscriptsRes.code === 200) {
      const manuscripts = manuscriptsRes.data || []
      selectedVideos.value = manuscripts.map(m => m.manuscriptId || m.id)
    }
  } catch (error) {
    console.error('获取数据失败:', error)
    availableVideos.value = []
  } finally {
    addingVideo.value = false
  }

  addVideoDialogVisible.value = true
}

const handleAddVideoToCollection = async () => {
  if (selectedVideos.value.length === 0) {
    ElMessage.warning('请选择要添加的视频')
    return
  }

  addingVideo.value = true
  try {
    for (const videoId of selectedVideos.value) {
      await collectionApi.addManuscriptToCollection(addVideoCollectionId.value, videoId, 0)
    }
    ElMessage.success('添加成功')
    addVideoDialogVisible.value = false

    if (collectionDetail.value.visible) {
      await goToCollectionDetail(collectionDetail.value.collectionId)
    }
    await loadUserCollections()
  } catch (error) {
    console.error('添加视频失败:', error)
    ElMessage.error('添加视频失败')
  } finally {
    addingVideo.value = false
  }
}

const handleRemoveVideoFromCollection = async (manuscript) => {
  try {
    await ElMessageBox.confirm(
      '确定要从合集中移除这个视频吗？',
      '移除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await collectionApi.removeManuscriptFromCollection(
      collectionDetail.value.collectionId,
      manuscript.id
    )
    if (response.code === 200) {
      ElMessage.success('移除成功')
      await goToCollectionDetail(collectionDetail.value.collectionId)
      await loadUserCollections()
    } else {
      ElMessage.error(response.message || '移除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('移除视频失败:', error)
      ElMessage.error('移除视频失败')
    }
  }
}

const playManuscript = (manuscript) => {
  router.push(`/manuscript/${manuscript.id}`)
}

const playCollectionAll = () => {
  if (collectionDetail.value.manuscripts.length > 0) {
    router.push(`/manuscript/${collectionDetail.value.manuscripts[0].id}`)
  }
}

const playCollectionAllFromList = (collection) => {
  if (collection.videos && collection.videos.length > 0) {
    router.push(`/manuscript/${collection.videos[0].id}`)
  }
}

watch(
  () => props.active,
  (newVal) => {
    if (newVal) {
      loadUserCollections()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.collection-management {
  margin-top: 20px;
}
</style>