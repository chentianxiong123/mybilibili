<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Star, MoreFilled } from '@element-plus/icons-vue'
import { interactionApi } from '@/api/client'
import api from '@/api/client'

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

// 收藏列表根据关键词过滤
const filteredFavorites = computed(() => {
  const keyword = (favorites.value.searchKeyword || '').trim().toLowerCase()
  if (!keyword) return favorites.value.videos
  return favorites.value.videos.filter(v => (v.title || '').toLowerCase().includes(keyword))
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

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 加载用户收藏
const loadUserFavorites = async () => {
  if (!props.userId) return

  props.loading.favorites = true
  try {
    const response = await api.get(`/manuscript/user/${props.userId}/collections`)
    if (response.code === 200) {
      favorites.value.videos = response.data
    }
  } catch (error) {
    console.error('获取用户收藏失败:', error)
  } finally {
    props.loading.favorites = false
  }
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

// 获取收藏夹封面
const getFavoriteFolderCover = (videos) => {
  if (!videos || videos.length === 0) {
    return 'https://picsum.photos/id/1025/400/225' // 默认封面
  }
  // 返回最新收藏的视频封面
  return videos[0].coverUrl || videos[0].cover || 'https://picsum.photos/id/1025/400/225'
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
        count: folder.collectCount || 0,
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
  console.log('openCreateFavoriteDialog 函数被调用')
  newFavoriteName.value = ''
  newFavoriteDescription.value = ''
  newFavoriteIsPublic.value = true
  newFavoriteCover.value = ''
  console.log('设置 createFavoriteDialogVisible 为 true')
  createFavoriteDialogVisible.value = true
  console.log('createFavoriteDialogVisible 的值:', createFavoriteDialogVisible.value)
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
  console.log('createFavoriteFolder 函数被调用')
  if (!newFavoriteName.value.trim()) {
    console.log('收藏夹名称为空')
    ElMessage.warning('请输入收藏夹名称')
    return
  }

  console.log('收藏夹名称:', newFavoriteName.value)
  creatingFavorite.value = true
  try {
    console.log('开始调用 interactionApi.createFavoriteFolder')
    const response = await interactionApi.createFavoriteFolder({
      name: newFavoriteName.value
    })
    console.log('interactionApi.createFavoriteFolder 调用成功:', response)
    if (response.code === 200) {
      console.log('创建成功')
      ElMessage.success('创建成功')
      createFavoriteDialogVisible.value = false
      newFavoriteName.value = ''
      newFavoriteDescription.value = ''
      newFavoriteIsPublic.value = true
      newFavoriteCover.value = ''
      // 重新加载收藏夹列表
      console.log('开始调用 loadFavoriteFolders')
      await loadFavoriteFolders()
      console.log('loadFavoriteFolders 调用成功')
    } else {
      console.log('创建失败，response.code:', response.code)
      ElMessage.error('创建失败')
    }
  } catch (error) {
    console.error('创建收藏夹失败:', error)
    ElMessage.error('创建收藏夹失败')
  } finally {
    console.log('finally 块执行')
    creatingFavorite.value = false
  }
}

// 编辑收藏夹对话框
const editFavoriteDialogVisible = ref(false)
const editingFavorite = ref(null)
const editingFavoriteName = ref('')
const updatingFavorite = ref(false)

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
      <div class="favorites-sidebar">
        <!-- 我的创建 -->
        <div class="sidebar-section">
          <div class="section-header">
            <div class="section-title">我的创建</div>
            <div class="section-action" @click="favorites.myCollectionsExpanded = !favorites.myCollectionsExpanded">{{ favorites.myCollectionsExpanded ? '▼' : '▲' }}</div>
          </div>

          <!-- 新建收藏夹按钮 -->
          <div v-if="isOwnSpace && favorites.myCollectionsExpanded" class="new-collection-btn" @click="openCreateFavoriteDialog">
            <div class="new-collection-icon">+</div>
            <div class="new-collection-text">新建收藏夹</div>
          </div>

          <!-- 收藏夹列表 -->
          <div v-if="favorites.myCollectionsExpanded" class="collection-list">
            <div
              v-for="collection in favorites.myCollections"
              :key="collection.name"
              :class="['collection-item', { active: favorites.activeCategory === collection.name }]"
              @click="() => {
                if (batchMode) toggleBatchMode();
                favorites.activeCategory = collection.name;
                loadFavoriteFolderVideos(collection.id);
              }"
            >
              <div class="collection-content">
                <span class="collection-icon">{{ collection.icon }}</span>
                <span class="collection-name">{{ collection.name }}</span>
              </div>
              <div class="collection-actions" @click.stop>
                <el-dropdown trigger="hover" placement="bottom-end">
                  <el-button link class="more-btn">
                    <el-icon><MoreFilled /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openEditFavoriteDialog(collection)">
                        编辑信息
                      </el-dropdown-item>
                      <el-dropdown-item divided @click="deleteFavoriteFolder(collection)">
                        删除
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
              <span class="collection-count">{{ collection.count }}</span>
            </div>

            <!-- 我的收藏内容 -->
            <div
              v-for="favorite in favorites.myFavorites"
              :key="favorite.name"
              :class="['favorite-item', { active: favorites.activeCategory === favorite.name }]"
              @click="favorites.activeCategory = favorite.name"
            >
              <div class="favorite-content">
                <span class="favorite-icon">{{ favorite.icon }}</span>
                <span class="favorite-name">{{ favorite.name }}</span>
              </div>
              <div class="favorite-actions" @click.stop>
                <el-dropdown trigger="hover" placement="bottom-end">
                  <el-button link class="more-btn">
                    <el-icon><MoreFilled /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openEditFavoriteDialog(favorite)">
                        编辑信息
                      </el-dropdown-item>
                      <el-dropdown-item divided @click="deleteFavoriteFolder(favorite)">
                        删除
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
              <span class="favorite-count">{{ favorite.count }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧内容区域 -->
      <div class="favorites-content">
        <!-- 收藏夹头部信息 -->
        <div class="favorite-header">
          <!-- 左侧小封面 -->
          <div class="favorite-header-cover">
            <img loading="lazy" decoding="async" :src="getFavoriteFolderCover(favorites.videos)" :alt="favorites.activeCategory" class="favorite-header-img">
          </div>
          <!-- 右侧信息 -->
          <div class="favorite-header-info">
            <div class="favorite-header-title">{{ favorites.activeCategory }}</div>
            <div class="favorite-header-meta">
              <span class="meta-item">创建者：{{ userInfo.username }}</span>
              <span class="meta-item">{{ (favorites.myCollections.find(c => c.name === favorites.activeCategory) || favorites.myFavorites.find(c => c.name === favorites.activeCategory))?.count || 0 }}个内容</span>
              <span class="meta-item">公开</span>
            </div>
            <div class="favorite-header-actions">
              <button class="action-btn play-all-btn" @click="playAllFavorites">
                <span class="play-icon">▶</span>
                播放全部视频
              </button>
            </div>
          </div>
        </div>

        <!-- 排序和筛选选项 -->
        <div class="sort-filter">
          <div class="batch-operations" v-if="!batchMode">
            <el-button size="small" @click="toggleBatchMode">批量操作</el-button>
          </div>
          <div class="batch-operations active" v-else>
            <el-checkbox
              :indeterminate="selectedFavorites.size > 0 && selectedFavorites.size < favorites.videos.length"
              :model-value="selectedFavorites.size === favorites.videos.length"
              @change="selectAllFavorites"
            >
              全选
            </el-checkbox>
            <span class="batch-selected-count">已选 {{ selectedFavorites.size }} 项</span>
            <el-button
              size="small"
              type="danger"
              :disabled="selectedFavorites.size === 0"
              :loading="batchDeleting"
              @click="batchDeleteFavorites"
            >
              批量删除
            </el-button>
            <el-button size="small" @click="toggleBatchMode">取消</el-button>
          </div>
          <div class="sort-options">
            <select v-model="favorites.activeSort" class="sort-select">
              <option v-for="option in favorites.sortOptions" :key="option" :value="option">{{ option }}</option>
            </select>
          </div>
          <div class="search-box">
            <input
              type="text"
              v-model="favorites.searchKeyword"
              placeholder="输入关键词"
              class="search-input"
            >
            <button class="search-btn">
              <el-icon><Search /></el-icon>
            </button>
          </div>
        </div>

        <!-- 收藏的视频列表 -->
        <div v-if="loading.favorites" class="loading-state">
          <p>加载中...</p>
        </div>
        <div v-else-if="filteredFavorites.length === 0" class="empty-state">
          <p>{{ favorites.searchKeyword ? '没有匹配的视频' : '暂无收藏' }}</p>
        </div>
        <div v-else class="videos-grid">
          <div v-for="video in filteredFavorites" :key="video.id" :class="['video-item', { 'video-item-selected': selectedFavorites.has(video.id), 'batch-mode': batchMode }]" @click="batchMode ? toggleSelectFavorite(video.id) : router.push(`/manuscript/${video.id}`)">
            <div v-if="batchMode" class="video-checkbox" @click.stop>
              <el-checkbox
                :model-value="selectedFavorites.has(video.id)"
                @change="toggleSelectFavorite(video.id)"
              />
            </div>
            <div class="video-cover">
              <img loading="lazy" decoding="async" :src="video.coverUrl || video.cover" :alt="video.title" class="video-cover-img">
              <div class="video-duration">{{ video.duration }}</div>
            </div>
            <div class="video-title">{{ video.title }}</div>
            <div class="video-meta">
              <span class="video-views">{{ video.viewCount }}</span>
              <span class="video-date">{{ video.date }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建收藏夹对话框 -->
    <el-dialog
      v-model="createFavoriteDialogVisible"
      title="收藏夹信息"
      width="400px"
    >
      <!-- 收藏夹封面 -->
      <div class="favorite-cover-section">
        <div class="cover-label">收藏夹封面</div>
        <div class="cover-upload-area">
          <div class="cover-placeholder" @click="$refs.coverInput.click()">
            <el-icon class="cover-icon"><Star /></el-icon>
          </div>
          <input
            ref="coverInput"
            type="file"
            accept="image/*"
            style="display: none"
            @change="handleCoverUpload"
          />
        </div>
      </div>

      <!-- 收藏夹名称 -->
      <div class="favorite-name-section">
        <div class="name-label">*收藏夹名称</div>
        <el-input
          v-model="newFavoriteName"
          placeholder="收藏夹名称"
          maxlength="20"
          show-word-limit
        />
      </div>

      <!-- 简介 -->
      <div class="favorite-description-section">
        <div class="description-label">简介:</div>
        <el-input
          v-model="newFavoriteDescription"
          type="textarea"
          placeholder="可填写简介"
          maxlength="200"
          show-word-limit
          :rows="4"
        />
      </div>

      <!-- 公开收藏夹 -->
      <div class="favorite-public-section">
        <el-checkbox v-model="newFavoriteIsPublic">公开收藏夹</el-checkbox>
      </div>

      <template #footer>
        <el-button @click="createFavoriteDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creatingFavorite" @click="createFavoriteFolder">
          提交
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑收藏夹对话框 -->
    <el-dialog
      v-model="editFavoriteDialogVisible"
      title="编辑收藏夹"
      width="400px"
    >
      <!-- 收藏夹名称 -->
      <div class="favorite-name-section">
        <div class="name-label">*收藏夹名称</div>
        <el-input
          v-model="editingFavoriteName"
          placeholder="收藏夹名称"
          maxlength="20"
          show-word-limit
        />
      </div>

      <template #footer>
        <el-button @click="editFavoriteDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="updatingFavorite" @click="updateFavoriteFolder">
          保存
        </el-button>
      </template>
    </el-dialog>
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

/* 收藏夹头部 */
.favorite-header {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background-color: #fff;
  border-radius: 8px;
}

.favorite-header-cover {
  width: 160px;
  height: 100px;
  border-radius: 4px;
  overflow: hidden;
  flex-shrink: 0;
  background-color: #f0f0f0;
}

.favorite-header-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.favorite-header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
}

.favorite-header-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.favorite-header-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #9499a0;
  margin-bottom: 12px;
}

.meta-item {
  position: relative;
}

.meta-item:not(:last-child)::after {
  content: '·';
  position: absolute;
  right: -10px;
}

.favorite-header-actions {
  display: flex;
  gap: 10px;
}

.favorite-header-actions .play-all-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #00aeec;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.favorite-header-actions .play-all-btn:hover {
  background-color: #0095d9;
}

.play-icon {
  font-size: 12px;
}

.favorites-container {
  display: flex;
  gap: 20px;
}

/* 左侧分类导航 */
.favorites-sidebar {
  width: 220px;
  background-color: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

/* 右侧内容区域 */
.favorites-content {
  flex: 1;
}

/* 侧边栏分组 */
.sidebar-section {
  margin-bottom: 20px;
}

/* 分组标题 */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.section-action {
  cursor: pointer;
  color: #9499a0;
  font-size: 12px;
}

/* 新建收藏夹按钮 */
.new-collection-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-top: 10px;
  margin-bottom: 12px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  border: 1px dashed #e0e0e0;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.new-collection-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.new-collection-icon {
  font-size: 16px;
  color: #9499a0;
}

.new-collection-text {
  font-size: 14px;
}

/* 收藏夹列表 */
.collection-list {
  margin-bottom: 20px;
}

.collection-item, .favorite-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  border-radius: 4px;
  transition: all 0.3s ease;
  white-space: nowrap;
  flex-direction: row;
}

.collection-content, .favorite-content {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow: hidden;
  justify-content: flex-start;
  flex-direction: row;
}

.collection-count, .favorite-count {
  font-size: 12px;
  color: #9499a0;
  white-space: nowrap;
  flex-shrink: 0;
}

.collection-name, .favorite-name {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  text-align: left;
}

.collection-item:hover, .favorite-item:hover {
  background-color: rgba(0, 174, 236, 0.1);
  color: #00aeec;
}

.collection-item.active, .favorite-item.active {
  background-color: #00aeec;
  color: #fff;
}

.collection-icon, .favorite-icon {
  font-size: 16px;
  min-width: 16px;
  text-align: center;
}

.collection-item.active .collection-count, .favorite-item.active .favorite-count {
  color: rgba(255, 255, 255, 0.8);
}

/* 收藏夹操作按钮 */
.favorite-actions,
.collection-actions {
  opacity: 0;
  transition: opacity 0.3s ease;
  margin-right: 4px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.favorite-item:hover .favorite-actions,
.collection-item:hover .collection-actions {
  opacity: 1;
}

.favorite-actions .more-btn,
.collection-actions .more-btn {
  padding: 2px 4px;
  color: #666;
  font-size: 16px;
}

.favorite-actions .more-btn:hover,
.collection-actions .more-btn:hover {
  background-color: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.favorite-item.active .favorite-actions .more-btn,
.collection-item.active .collection-actions .more-btn {
  color: #fff;
}

.favorite-item.active .favorite-actions .more-btn:hover,
.collection-item.active .collection-actions .more-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

/* 排序和筛选选项 */
.sort-filter {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  font-size: 14px;
  color: #666;
}

.batch-operations {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-operations.active {
  cursor: default;
}

.batch-operations:hover {
  color: #00aeec;
}

.batch-selected-count {
  font-size: 13px;
  color: #666;
  margin-left: 4px;
}

.sort-options {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sort-select {
  padding: 4px 8px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 14px;
  color: #666;
}

/* 搜索框 */
.search-box {
  display: flex;
  align-items: center;
  position: relative;
  width: 200px;
}

.search-box .search-input {
  width: 100%;
  padding: 6px 32px 6px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 16px;
  font-size: 14px;
  box-sizing: border-box;
}

.search-box .search-btn {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
  padding: 4px;
  background-color: transparent;
  color: #9499a0;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.search-box .search-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: #333;
}

/* 收藏的视频列表 */
.favorites-content .videos-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

/* 视频项 */
.video-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  position: relative;
  transition: opacity 0.2s;
}

.video-item.batch-mode {
  cursor: pointer;
}

.video-item.batch-mode .video-cover {
  opacity: 0.9;
}

.video-item-selected {
  border-radius: 8px;
  outline: 2px solid #00aeec;
  outline-offset: 2px;
}

.video-checkbox {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 10;
}

.video-checkbox :deep(.el-checkbox__inner) {
  background-color: #fff;
  border-color: #00aeec;
}

.video-checkbox :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #00aeec;
  border-color: #00aeec;
}

.video-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.video-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.video-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.video-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #999;
}

.video-views {
  font-size: 12px;
  color: #999;
  margin: 0;
}

.video-date {
  font-size: 12px;
  color: #999;
}

/* 加载状态和空状态 */
.loading-state,
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
  color: #9499a0;
  font-size: 14px;
}

/* 收藏夹创建对话框样式 */
.favorite-cover-section {
  margin-bottom: 20px;
}

.cover-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

.cover-upload-area {
  position: relative;
}

.cover-placeholder {
  width: 100px;
  height: 100px;
  background-color: #f0f0f0;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.cover-placeholder:hover {
  background-color: #e0e0e0;
}

.cover-icon {
  font-size: 40px;
  color: #999;
}

.favorite-name-section {
  margin-bottom: 20px;
}

.name-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

.favorite-description-section {
  margin-bottom: 20px;
}

.description-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

.favorite-public-section {
  margin-bottom: 20px;
}
</style>