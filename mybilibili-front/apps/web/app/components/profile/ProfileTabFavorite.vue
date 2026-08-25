<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { interactionApi } from '@/api/client'
import FavoriteFolderList from './FavoriteFolderList.vue'
import FavoriteVideoGrid from './FavoriteVideoGrid.vue'
import FavoriteFolderDialog from './FavoriteFolderDialog.vue'

const props = defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  },
  userInfo: {
    type: Object,
    required: true
  },
  loading: {
    type: Object,
    required: true
  }
})

const router = useRouter()

// 收藏数据
const favorites = ref({
  activeCategory: '默认收藏夹',
  myCollectionsExpanded: true,
  myCollections: [],
  myFavorites: [],
  videos: [],
  sortOptions: ['最新收藏', '最早收藏'],
  activeSort: '最新收藏',
  searchKeyword: ''
})

// 监听收藏排序变化
watch(() => favorites.value.activeSort, () => {
  const activeFolder = favorites.value.myCollections.find(f => f.name === favorites.value.activeCategory)
  if (activeFolder) {
    loadFavoriteFolderVideos(activeFolder.id)
  }
})

// 批量操作状态
const batchMode = ref(false)
const selectedFavorites = ref(new Set())
const batchDeleting = ref(false)

const toggleBatchMode = () => {
  batchMode.value = !batchMode.value
  if (!batchMode.value) {
    selectedFavorites.value.clear()
  }
}

const toggleSelectFavorite = (videoId) => {
  const newSet = new Set(selectedFavorites.value)
  if (newSet.has(videoId)) {
    newSet.delete(videoId)
  } else {
    newSet.add(videoId)
  }
  selectedFavorites.value = newSet
}

const selectAllFavorites = () => {
  if (selectedFavorites.value.size === favorites.value.videos.length) {
    selectedFavorites.value = new Set()
  } else {
    selectedFavorites.value = new Set(favorites.value.videos.map(v => v.id))
  }
}

const batchDeleteFavorites = async () => {
  if (selectedFavorites.value.size === 0) {
    ElMessage.warning('请选择要删除的视频')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedFavorites.value.size} 个视频吗？`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchDeleting.value = true
    const activeFolder = favorites.value.myCollections.find(c => c.name === favorites.value.activeCategory)
    if (!activeFolder) return

    const deletePromises = []
    for (const videoId of selectedFavorites.value) {
      deletePromises.push(
        interactionApi.removeVideoFromFavoriteFolder(activeFolder.id, videoId)
          .catch(err => console.error(`删除视频 ${videoId} 失败:`, err))
      )
    }

    await Promise.all(deletePromises)
    ElMessage.success(`成功删除 ${selectedFavorites.value.size} 个视频`)
    selectedFavorites.value.clear()
    await loadFavoriteFolderVideos(activeFolder.id)
    await loadFavoriteFolders()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  } finally {
    batchDeleting.value = false
  }
}

// 新建收藏夹对话框
const createFavoriteDialogVisible = ref(false)
const newFavoriteName = ref('')
const newFavoriteDescription = ref('')
const newFavoriteIsPublic = ref(true)
const newFavoriteCover = ref('')
const creatingFavorite = ref(false)

// 编辑收藏夹对话框
const editFavoriteDialogVisible = ref(false)
const editingFavorite = ref(null)
const editingFavoriteName = ref('')
const updatingFavorite = ref(false)

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 加载收藏夹视频列表
const loadFavoriteFolderVideos = async (folderId) => {
  props.loading.favorites = true
  try {
    const sortOrder = favorites.value.activeSort === '最早收藏' ? 'asc' : 'desc'
    const response = await interactionApi.getFavoriteFolderVideos(folderId, 1, 1000, sortOrder)
    if (response.code === 200) {
      // 处理视频数据，添加date字段
      const videos = response.data.map(video => {
        return {
          ...video,
          date: formatDate(video.createdAt || video.collectTime)
        }
      })
      favorites.value.videos = videos
    }
  } catch (error) {
    console.error('获取收藏夹视频失败:', error)
  } finally {
    props.loading.favorites = false
  }
}

// 加载收藏夹列表
const loadFavoriteFolders = async () => {
  try {
    const response = await interactionApi.getFavoriteFolders()
    if (response.code === 200) {
      // 获取用户创建的收藏夹
      const userFolders = (response.data || []).map(folder => ({
        id: folder.id,
        name: folder.name,
        count: folder.video_count || 0,
        icon: '📁'
      }))

      // 检查是否存在默认收藏夹
      const hasDefaultFolder = userFolders.some(folder => folder.name === '默认收藏夹')

      // 如果不存在默认收藏夹，添加一个
      if (!hasDefaultFolder) {
        userFolders.unshift({
          id: 0,
          name: '默认收藏夹',
          count: 0,
          icon: '📁'
        })
      } else {
        // 如果存在默认收藏夹，将其移到最前面
        const defaultFolderIndex = userFolders.findIndex(folder => folder.name === '默认收藏夹')
        if (defaultFolderIndex > 0) {
          const defaultFolder = userFolders.splice(defaultFolderIndex, 1)[0]
          userFolders.unshift(defaultFolder)
        }
      }

      favorites.value.myCollections = userFolders
      // 清空 myFavorites，因为默认收藏夹已经包含在 myCollections 中
      favorites.value.myFavorites = []

      // 根据当前选中的收藏夹加载视频列表
      const activeFolder = userFolders.find(folder => folder.name === favorites.value.activeCategory)
      if (activeFolder) {
        await loadFavoriteFolderVideos(activeFolder.id)
      }
    }
  } catch (error) {
    console.error('加载收藏夹列表失败:', error)
  }
}

// 打开新建收藏夹对话框
const openCreateFavoriteDialog = () => {
  newFavoriteName.value = ''
  newFavoriteDescription.value = ''
  newFavoriteIsPublic.value = true
  newFavoriteCover.value = ''
  createFavoriteDialogVisible.value = true
}

// 处理封面上传
const handleCoverUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    // 这里可以添加文件上传逻辑，目前只是简单处理
    const reader = new FileReader()
    reader.onload = (e) => {
      newFavoriteCover.value = e.target.result
    }
    reader.readAsDataURL(file)
  }
}

// 创建收藏夹
const createFavoriteFolder = async () => {
  if (!newFavoriteName.value.trim()) {
    ElMessage.warning('请输入收藏夹名称')
    return
  }

  creatingFavorite.value = true
  try {
    const response = await interactionApi.createFavoriteFolder({
      name: newFavoriteName.value
    })
    if (response.code === 200) {
      ElMessage.success('创建成功')
      createFavoriteDialogVisible.value = false
      newFavoriteName.value = ''
      newFavoriteDescription.value = ''
      newFavoriteIsPublic.value = true
      newFavoriteCover.value = ''
      // 重新加载收藏夹列表
      await loadFavoriteFolders()
    } else {
      ElMessage.error('创建失败')
    }
  } catch (error) {
    console.error('创建收藏夹失败:', error)
    ElMessage.error('创建收藏夹失败')
  } finally {
    creatingFavorite.value = false
  }
}

// 打开编辑收藏夹对话框
const openEditFavoriteDialog = (favorite) => {
  editingFavorite.value = favorite
  editingFavoriteName.value = favorite.name
  editFavoriteDialogVisible.value = true
}

// 更新收藏夹
const updateFavoriteFolder = async () => {
  if (!editingFavoriteName.value.trim()) {
    ElMessage.warning('请输入收藏夹名称')
    return
  }

  updatingFavorite.value = true
  try {
    const response = await interactionApi.updateFavoriteFolder(editingFavorite.value.id, editingFavoriteName.value)
    if (response.code === 200) {
      ElMessage.success('更新成功')
      editFavoriteDialogVisible.value = false
      await loadFavoriteFolders()
    } else {
      ElMessage.error(response.message || '更新失败')
    }
  } catch (error) {
    console.error('更新收藏夹失败:', error)
    ElMessage.error('更新收藏夹失败')
  } finally {
    updatingFavorite.value = false
  }
}

// 删除收藏夹
const deleteFavoriteFolder = async (favorite) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除收藏夹"${favorite.name}"吗？删除后无法恢复。`,
      '删除收藏夹',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await interactionApi.deleteFavoriteFolder(favorite.id)
    if (response.code === 200) {
      ElMessage.success('删除成功')
      await loadFavoriteFolders()
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除收藏夹失败:', error)
      ElMessage.error('删除收藏夹失败')
    }
  }
}

// 选择收藏夹
const selectFavoriteFolder = (collection) => {
  if (batchMode.value) {
    toggleBatchMode()
  }
  favorites.value.activeCategory = collection.name
  loadFavoriteFolderVideos(collection.id)
}

// 播放收藏视频
const playAllFavorites = () => {
  if (favorites.value.videos.length > 0) {
    router.push(`/manuscript/${favorites.value.videos[0].id}`)
  } else {
    ElMessage.info('暂无收藏视频')
  }
}

const loadData = () => {
  if (!props.userId) return
  // 注意：loadFavoriteFolders 内部会调用 loadFavoriteFolderVideos 加载当前选中收藏夹的视频
  loadFavoriteFolders()
}

onMounted(() => {
  loadData()
})

watch(() => props.userId, () => {
  loadData()
})
</script>

<template>
  <div class="favorites-section">
    <div class="favorites-container">
      <!-- 左侧分类导航 -->
      <FavoriteFolderList
        :collections="favorites.myCollections"
        :my-favorites="favorites.myFavorites"
        :active-category="favorites.activeCategory"
        :my-collections-expanded="favorites.myCollectionsExpanded"
        :is-own-space="isOwnSpace"
        @toggle-expand="favorites.myCollectionsExpanded = !favorites.myCollectionsExpanded"
        @open-create="openCreateFavoriteDialog"
        @select-folder="selectFavoriteFolder"
        @open-edit="openEditFavoriteDialog"
        @delete-folder="deleteFavoriteFolder"
      />

      <!-- 右侧内容区域 -->
      <FavoriteVideoGrid
        :videos="favorites.videos"
        :loading="loading.favorites"
        :batch-mode="batchMode"
        :selected-favorites="selectedFavorites"
        :active-category="favorites.activeCategory"
        :active-sort="favorites.activeSort"
        :sort-options="favorites.sortOptions"
        :search-keyword="favorites.searchKeyword"
        :user-info="userInfo"
        :collections="favorites.myCollections"
        :my-favorites="favorites.myFavorites"
        @toggle-batch="toggleBatchMode"
        @toggle-select="toggleSelectFavorite"
        @select-all="selectAllFavorites"
        @batch-delete="batchDeleteFavorites"
        @play-all="playAllFavorites"
        @update:active-sort="value => favorites.activeSort = value"
        @update:search-keyword="value => favorites.searchKeyword = value"
        @navigate="id => router.push(`/manuscript/${id}`)"
      />
    </div>

    <!-- 新建/编辑收藏夹对话框 -->
    <FavoriteFolderDialog
      v-model:create-visible="createFavoriteDialogVisible"
      v-model:edit-visible="editFavoriteDialogVisible"
      v-model:new-favorite-name="newFavoriteName"
      v-model:new-favorite-description="newFavoriteDescription"
      v-model:new-favorite-is-public="newFavoriteIsPublic"
      v-model:editing-favorite-name="editingFavoriteName"
      :creating-favorite="creatingFavorite"
      :updating-favorite="updatingFavorite"
      @cover-upload="handleCoverUpload"
      @create="createFavoriteFolder"
      @update="updateFavoriteFolder"
    />
  </div>
</template>

<style scoped>
/* 收藏页面样式 */
.favorites-section {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.favorites-container {
  display: flex;
  gap: 20px;
}
</style>