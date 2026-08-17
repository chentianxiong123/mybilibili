<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, ArrowRight, Edit, Plus, VideoPlay, MoreFilled, Clock, Grid, List, View, ArrowDown, Search, Delete } from '@element-plus/icons-vue'
import { videoApi } from '@/api/client'
import { collectionApi } from '@/api/collection.ts'

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

// 格式化数字
const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

// 获取默认封面
const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}

// 合集数据
const collections = ref({
  viewType: 'horizontal',
  items: [],
  loading: false
})

// 新建合集对话框
const createCollectionDialogVisible = ref(false)
const createCollectionForm = ref({
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const creatingCollection = ref(false)

// 编辑合集对话框
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

// 添加视频到合集对话框
const addVideoDialogVisible = ref(false)
const addVideoSearchKeyword = ref('')
const addVideoSortBy = ref('newest')
const userVideos = ref([])
const selectedVideos = ref([])
const addVideoLoading = ref(false)
const currentCollectionId = ref(null)

// 新建合集选择视频对话框
const createCollectionVideoDialogVisible = ref(false)
const createCollectionSelectedVideos = ref([])

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

// 加载用户的合集列表
const loadUserCollections = async () => {
  if (!props.userId) return

  collections.value.loading = true
  try {
    console.log('开始获取合集列表，用户ID:', props.userId)
    const response = await collectionApi.getUserCollections(props.userId, 1, 100)
    console.log('获取合集列表响应:', response)
    if (response.code === 200) {
      const list = response.data || []
      console.log('合集列表:', list)
      // 为每个合集加载视频（仅用于显示，不用于计数）
      for (const collection of list) {
        console.log('处理合集:', collection.id, collection.title)
        try {
          const videoResponse = await collectionApi.getCollectionManuscripts(collection.id, 1, 10)
          console.log('合集', collection.id, '的稿件响应:', videoResponse)
          if (videoResponse.code === 200) {
            collection.videos = (videoResponse.data || []).map(video => ({
              ...video,
              date: formatDate(video.uploadTime)
            }))
            // 始终用第一个视频的封面作为合集封面
            if (collection.videos.length > 0 && collection.videos[0].coverUrl) {
              collection.coverUrl = collection.videos[0].coverUrl
            }
            console.log('合集', collection.id, '的视频列表:', collection.videos)
          }
        } catch (e) {
          console.error('获取合集', collection.id, '的稿件失败:', e)
          collection.videos = []
        }
      }
      collections.value.items = list
      console.log('最终合集数据:', collections.value.items)
    }
  } catch (error) {
    console.error('获取合集列表失败:', error)
  } finally {
    collections.value.loading = false
  }
}

// 打开新建合集对话框
const openCreateCollectionDialog = () => {
  createCollectionForm.value = {
    name: '',
    description: '',
    cover: null,
    coverUrl: '',
    isPublic: true
  }
  createCollectionSelectedVideos.value = []
  createCollectionDialogVisible.value = true
}

// 打开编辑合集对话框
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
      editCollectionDialogVisible.value = false
      // 刷新合集详情和列表
      loadCollectionDetailData()
      loadUserCollections()
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
      createCollectionDialogVisible.value = false
      createCollectionSelectedVideos.value = []
      loadUserCollections()
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

// 打开添加视频对话框
const openAddVideoDialog = async (collectionId) => {
  currentCollectionId.value = collectionId
  addVideoSearchKeyword.value = ''
  addVideoSortBy.value = 'newest'
  selectedVideos.value = []
  addVideoDialogVisible.value = true
  await loadUserVideosForSelection()
}

// 加载用户视频供选择
const loadUserVideosForSelection = async () => {
  addVideoLoading.value = true
  try {
    // 获取当前合集已有的稿件ID列表
    const collectionResponse = await collectionApi.getCollectionManuscripts(currentCollectionId.value, 1, 100)
    if (collectionResponse.code === 200) {
      const manuscripts = collectionResponse.data || []
      selectedVideos.value = manuscripts.map(m => m.id)
      console.log('当前合集中的稿件IDs:', selectedVideos.value)
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
      console.log('用户所有视频:', userVideos.value)
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
  console.log('开始更新合集视频，当前合集ID:', currentCollectionId.value)
  console.log('选中的视频IDs:', selectedVideos.value)

  try {
    // 获取当前合集中已有的稿件
    const collectionResponse = await collectionApi.getCollectionManuscripts(currentCollectionId.value, 1, 100)
    let existingManuscriptIds = []
    if (collectionResponse.code === 200) {
      existingManuscriptIds = (collectionResponse.data || []).map(m => m.id)
    }
    console.log('当前合集中已有的稿件IDs:', existingManuscriptIds)

    // 计算需要添加和移除的稿件
    const toAdd = selectedVideos.value.filter(id => !existingManuscriptIds.includes(id))
    const toRemove = existingManuscriptIds.filter(id => !selectedVideos.value.includes(id))

    console.log('需要添加的稿件:', toAdd)
    console.log('需要移除的稿件:', toRemove)

    // 执行添加操作
    const addPromises = toAdd.map((manuscriptId, index) => {
      return collectionApi.addManuscriptToCollection(
        currentCollectionId.value,
        manuscriptId,
        existingManuscriptIds.length + index
      )
    })

    // 执行移除操作
    const removePromises = toRemove.map(manuscriptId => {
      return collectionApi.removeManuscriptFromCollection(
        currentCollectionId.value,
        manuscriptId
      )
    })

    // 等待所有操作完成
    await Promise.all([...addPromises, ...removePromises])

    console.log('更新合集视频成功')
    ElMessage.success('更新成功')
    addVideoDialogVisible.value = false
    loadUserCollections()
  } catch (error) {
    console.error('更新合集视频失败:', error)
    ElMessage.error('更新失败: ' + (error.response?.data?.message || error.message))
  }
}

// 在当前页面内显示合集详情
const goToCollectionDetail = (collectionId) => {
  collectionDetail.value.collectionId = collectionId
  collectionDetail.value.visible = true
  loadCollectionDetailData()
}

// 返回合集列表
const backToCollectionsList = () => {
  collectionDetail.value.visible = false
  collectionDetail.value.collectionId = null
  collectionDetail.value.collection = null
  collectionDetail.value.manuscripts = []
}

// 加载合集详情数据
const loadCollectionDetailData = async () => {
  if (!collectionDetail.value.collectionId) return

  collectionDetail.value.loading = true
  try {
    // 加载合集信息
    const collectionResponse = await collectionApi.getCollectionById(collectionDetail.value.collectionId)
    if (collectionResponse.code === 200) {
      collectionDetail.value.collection = collectionResponse.data
    }

    // 加载稿件列表
    const manuscriptsResponse = await collectionApi.getCollectionManuscripts(
      collectionDetail.value.collectionId,
      collectionDetail.value.pagination.currentPage,
      collectionDetail.value.pagination.pageSize
    )
    if (manuscriptsResponse.code === 200) {
      const manuscripts = manuscriptsResponse.data || []
      collectionDetail.value.manuscripts = manuscripts
      collectionDetail.value.pagination.total = manuscripts.length
    }
  } catch (error) {
    console.error('获取合集详情失败:', error)
    ElMessage.error('获取合集详情失败')
  } finally {
    collectionDetail.value.loading = false
  }
}

// 处理合集详情分页变化
const handleCollectionDetailPageChange = (page) => {
  collectionDetail.value.pagination.currentPage = page
  loadCollectionDetailData()
}

// 处理合集详情每页数量变化
const handleCollectionDetailSizeChange = (size) => {
  collectionDetail.value.pagination.pageSize = size
  collectionDetail.value.pagination.currentPage = 1
  loadCollectionDetailData()
}

// 处理合集详情排序变化
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

// 播放稿件
const playManuscript = (manuscript) => {
  if (manuscript.id) {
    router.push(`/manuscript/${manuscript.id}`)
  }
}

// 播放合集全部视频
const playCollectionAll = () => {
  if (collectionDetail.value.manuscripts.length > 0) {
    router.push(`/manuscript/${collectionDetail.value.manuscripts[0].id}`)
  } else {
    ElMessage.info('该合集暂无视频')
  }
}

// 从合集列表播放全部视频
const playCollectionAllFromList = (collection) => {
  if (collection.videos && collection.videos.length > 0) {
    router.push(`/manuscript/${collection.videos[0].id}`)
  } else {
    ElMessage.info('该合集暂无视频')
  }
}

// 打开添加视频到合集对话框
const openAddVideoToCollectionDialog = () => {
  openAddVideoDialog(collectionDetail.value.collectionId)
}

// 处理合集操作命令
const handleCollectionCommand = (command) => {
  if (command === 'delete') {
    ElMessageBox.confirm('确定要删除这个视频列表吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      deleteCollection(collectionDetail.value.collectionId)
    }).catch(() => {})
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

// 处理视频操作命令
const handleVideoCommand = (command, manuscript) => {
  if (command === 'remove') {
    ElMessageBox.confirm('确定要从列表中移除这个视频吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      removeVideoFromCollection(manuscript.id)
    }).catch(() => {})
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
  createCollectionVideoDialogVisible.value = true
}

// 加载用户视频供新建合集选择
const loadUserVideosForCreateCollection = async () => {
  addVideoLoading.value = true
  try {
    const currentUserId = JSON.parse(localStorage.getItem('user') || '{}')?.id
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
    <div v-if="collectionDetail.visible" class="collection-detail-view">
      <!-- 返回按钮和标题 -->
      <div class="collection-detail-header">
        <div class="header-left">
          <el-button text :icon="ArrowLeft" @click="backToCollectionsList">
            返回
          </el-button>
          <span class="header-title">我的合集和视频列表</span>
          <el-icon><ArrowRight /></el-icon>
          <span class="collection-name">{{ collectionDetail.collection?.title || '加载中...' }}</span>
        </div>
        <div class="header-right" v-if="isOwnSpace">
          <el-button text :icon="Edit" @click="openEditCollectionDialog">
            编辑
          </el-button>
        </div>
      </div>

      <!-- 合集信息区 -->
      <div class="collection-info-section" v-if="collectionDetail.collection">
        <div class="info-container">
          <!-- 封面 -->
          <div class="collection-cover-large">
            <img loading="lazy" decoding="async"
              :src="collectionDetail.collection.coverUrl || getDefaultCover()"
              :alt="collectionDetail.collection.title"
            />
            <div class="cover-overlay" @click="playCollectionAll">
              <el-button type="primary" size="large" :icon="VideoPlay">
                播放全部
              </el-button>
            </div>
            <div class="video-count-badge">
              <el-icon><VideoPlay /></el-icon>
              <span>{{ collectionDetail.pagination.total }} 个视频</span>
            </div>
          </div>

          <!-- 信息 -->
          <div class="collection-meta-detail">
            <h1 class="collection-title-large">{{ collectionDetail.collection.title }}</h1>
            <p class="collection-desc-large">{{ collectionDetail.collection.description || '暂无描述' }}</p>

            <!-- 统计信息 -->
            <div class="stats-info">
              <span class="stat-item">
                <el-icon><Clock /></el-icon>
                更新于 {{ formatDate(collectionDetail.collection.updatedAt) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 视频网格区 -->
      <div class="collection-videos-section">
        <!-- 头部操作栏 -->
        <div class="videos-header">
          <div class="sort-options">
            <span
              class="sort-item"
              :class="{ active: collectionDetail.sortBy === 'default' }"
              @click="handleCollectionDetailSortChange('default')"
            >
              默认排序
            </span>
            <span
              class="sort-item"
              :class="{ active: collectionDetail.sortBy === 'newest' }"
              @click="handleCollectionDetailSortChange('newest')"
            >
              升序排序
            </span>
          </div>
          <div class="header-actions" v-if="isOwnSpace">
            <el-button type="primary" :icon="Edit" @click="openEditCollectionDialog">
              编辑
            </el-button>
            <el-dropdown trigger="click" @command="handleCollectionCommand">
              <el-button :icon="MoreFilled" circle />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="delete">删除视频列表</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <!-- 加载状态 -->
        <div v-if="collectionDetail.loading" class="loading-state">
          <el-skeleton :rows="3" animated />
        </div>

        <!-- 空状态 -->
        <div v-else-if="collectionDetail.manuscripts.length === 0 && !isOwnSpace" class="empty-state">
          <el-empty description="暂无视频" />
        </div>

        <!-- 视频网格 -->
        <div v-else class="videos-grid">
          <!-- 添加视频卡片 -->
          <div v-if="isOwnSpace" class="video-card add-video-card" @click="openAddVideoToCollectionDialog">
            <div class="add-video-content">
              <el-icon :size="40"><Plus /></el-icon>
              <span>添加视频</span>
            </div>
          </div>

          <!-- 视频卡片 -->
          <div
            v-for="(manuscript, index) in collectionDetail.manuscripts"
            :key="manuscript.id"
            class="collection-video-card"
          >
            <!-- 封面区域 -->
            <div class="collection-video-cover" @click="playManuscript(manuscript)">
              <img loading="lazy" decoding="async"
                :src="manuscript.coverUrl || getDefaultCover()"
                :alt="manuscript.title"
                class="collection-video-cover-img"
              />
              <!-- 序号 -->
              <div class="collection-video-index">{{ index + 1 }}</div>
              <!-- 时长 -->
              <div class="collection-video-duration">{{ manuscript.duration || '00:00' }}</div>
              <!-- 更多操作 -->
              <div class="collection-video-actions" v-if="isOwnSpace" @click.stop>
                <el-dropdown trigger="click" @command="(cmd) => handleVideoCommand(cmd, manuscript)">
                  <el-button :icon="MoreFilled" circle size="small" />
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="remove">从列表中移除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>

            <!-- 信息区域 -->
            <div class="collection-video-info">
              <h3 class="collection-video-title" :title="manuscript.title">{{ manuscript.title }}</h3>
              <div class="collection-video-meta">
                <span class="collection-meta-item">
                  <el-icon><VideoPlay /></el-icon>
                  {{ formatNumber(manuscript.viewCount) }}
                </span>
                <span class="collection-meta-item">{{ formatDate(manuscript.uploadTime) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 合集列表视图 -->
    <template v-else>
      <!-- 顶部标题和新建按钮 -->
      <div class="collections-header">
        <div class="collections-header-left">
          <h2 class="collections-title">我的合集和视频列表</h2>
          <el-button
            v-if="isOwnSpace"
            type="primary"
            :icon="Plus"
            class="new-collection-btn"
            @click="openCreateCollectionDialog"
          >
            新建
          </el-button>
        </div>
        <!-- 视图切换按钮 -->
        <div class="view-toggle">
          <button
            class="view-toggle-btn grid-view-btn"
            :class="{ active: collections.viewType === 'grid' }"
            @click="collections.viewType = 'grid'"
          >
            <el-icon><Grid /></el-icon>
          </button>
          <button
            class="view-toggle-btn list-view-btn"
            :class="{ active: collections.viewType === 'horizontal' }"
            @click="collections.viewType = 'horizontal'"
          >
            <el-icon><List /></el-icon>
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="collections.loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>

      <!-- 空状态 -->
      <div v-else-if="collections.items.length === 0" class="empty-state">
        <el-empty description="暂无合集">
          <el-button
            v-if="isOwnSpace"
            type="primary"
            @click="openCreateCollectionDialog"
          >
            创建合集
          </el-button>
        </el-empty>
      </div>

      <!-- 宫格视图 -->
      <div v-else-if="collections.viewType === 'grid'" class="collections-grid">
        <!-- 新建合集卡片 -->
        <div
          v-if="isOwnSpace"
          class="collection-item new-collection"
          @click="openCreateCollectionDialog"
        >
          <div class="new-collection-content">
            <el-icon :size="32"><Plus /></el-icon>
            <span>新建合集</span>
          </div>
        </div>

        <!-- 合集项 -->
        <div
          v-for="collection in collections.items"
          :key="collection.id"
          class="collection-item"
          @click="goToCollectionDetail(collection.id)"
        >
          <div class="collection-cover">
            <img loading="lazy" decoding="async" :src="collection.coverUrl || getDefaultCover()" :alt="collection.title" class="collection-cover-img">
            <div class="collection-video-count">
              <el-icon><VideoPlay /></el-icon>
              <span>{{ collection.manuscriptCount || 0 }}</span>
            </div>
          </div>
          <div class="collection-info">
            <div class="collection-title">{{ collection.title }}</div>
            <div class="collection-date">{{ collection.manuscriptCount || 0 }}个视频</div>
          </div>
        </div>
      </div>

      <!-- 水平列表视图 -->
      <div v-else-if="collections.viewType === 'horizontal'" class="collections-horizontal">
        <!-- 合集项 -->
        <div
          v-for="collection in collections.items"
          :key="collection.id"
          class="collection-horizontal-item"
        >
          <div class="collection-horizontal-header">
            <h3 class="collection-horizontal-title">
              {{ collection.title }}
              <span class="collection-video-count-badge">{{ collection.manuscriptCount || 0 }}</span>
            </h3>
            <div class="collection-horizontal-actions">
              <el-button
                v-if="collection.videos && collection.videos.length > 0"
                class="action-btn play-all-btn"
                :icon="VideoPlay"
                @click="playCollectionAllFromList(collection)"
              >
                播放全部
              </el-button>
              <el-button
                class="action-btn more-btn"
                @click="goToCollectionDetail(collection.id)"
              >
                更多
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>

          <!-- 水平视频列表 -->
          <div class="collection-videos-horizontal">
            <!-- 添加视频按钮 -->
            <div
              v-if="isOwnSpace"
              class="add-video-card"
              @click="openAddVideoDialog(collection.id)"
            >
              <div class="add-video-content">
                <el-icon :size="24"><Plus /></el-icon>
                <span>添加视频</span>
              </div>
            </div>

            <!-- 视频项 -->
            <div
              v-for="video in collection.videos"
              :key="video.id"
              class="video-horizontal-item"
              @click="router.push(`/manuscript/${video.id}`)"
            >
              <div class="video-horizontal-cover">
                <img loading="lazy" decoding="async" :src="video.coverUrl || getDefaultCover()" :alt="video.title" class="video-cover-img">
                <div class="video-duration">{{ video.duration }}</div>
              </div>
              <div class="video-horizontal-info">
                <div class="video-title" :title="video.title">{{ video.title }}</div>
                <div class="video-horizontal-meta">
                  <span class="video-views">
                    <el-icon><View /></el-icon>
                    {{ video.viewCount || 0 }}
                  </span>
                  <span class="video-date">{{ video.date }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- 编辑合集对话框 -->
    <el-dialog
      v-model="editCollectionDialogVisible"
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
        <el-button @click="editCollectionDialogVisible = false">取消</el-button>
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
      v-model="createCollectionDialogVisible"
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
        <el-button @click="createCollectionDialogVisible = false">取消</el-button>
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
      v-model="createCollectionVideoDialogVisible"
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
        <el-button @click="createCollectionVideoDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="creatingCollection"
          @click="handleCreateCollection"
        >
          确定 ({{ createCollectionSelectedVideos.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 添加视频到合集对话框 -->
    <el-dialog
      v-model="addVideoDialogVisible"
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
        <el-button @click="addVideoDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleAddVideosToCollection"
        >
          更新
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
/* 合集和列表页面样式 */
.collections-section {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  border-radius: 8px;
}

/* 顶部标题 */
.collections-header {
  margin-bottom: 20px;
}

.collections-title {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

/* 合集网格布局 */
.collections-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 20px;
}

/* 合集项样式 */
.collection-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.collection-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 新建合集按钮样式 */
.new-collection {
  border: 1px dashed #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
  gap: 10px;
}

.new-collection-icon {
  font-size: 32px;
  color: #999;
}

.new-collection-text {
  font-size: 14px;
  color: #999;
}

/* 合集封面样式 */
.collection-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.collection-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.collection-item:hover .collection-cover-img {
  transform: scale(1.05);
}

/* 合集视频数量样式 */
.collection-video-count {
  position: absolute;
  top: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 合集信息样式 */
.collection-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.collection-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-date {
  font-size: 12px;
  color: #999;
}

/* 响应式设计 */
@media (max-width: 576px) {
  .collections-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* 视图切换按钮样式 */
.view-toggle {
  display: flex;
  gap: 10px;
  align-items: center;
}

.view-toggle-btn {
  padding: 6px 10px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.3s ease;
}

.view-toggle-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.view-toggle-btn.active {
  background-color: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

/* 水平列表视图样式 */
.collections-horizontal {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 水平列表项样式 */
.collection-horizontal-item {
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  padding: 16px;
  gap: 16px;
}

/* 水平列表头部样式 */
.collection-horizontal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-horizontal-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.collection-horizontal-title .collection-video-count {
  font-size: 14px;
  color: #999;
  font-weight: normal;
}

/* 水平列表操作按钮 */
.collection-horizontal-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

/* 水平视频列表样式 */
.collection-videos-horizontal {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  overflow-x: auto;
  padding-bottom: 8px;
}

.collection-videos-horizontal::-webkit-scrollbar {
  height: 4px;
}

.collection-videos-horizontal::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 2px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 2px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}

/* 添加视频按钮 */
.add-video-card {
  width: 180px;
  height: 100px;
  border: 1px dashed #e0e0e0;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background-color: #fafafa;
  cursor: pointer;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.add-video-card:hover {
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.add-video-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #999;
  font-size: 14px;
}

/* 水平视频项样式 */
.video-horizontal-item {
  width: 180px;
  flex-shrink: 0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.video-horizontal-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.video-horizontal-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f0f0f0;
  margin-bottom: 8px;
}

.video-horizontal-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-horizontal-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.video-horizontal-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #9499a0;
  align-items: center;
}

.video-horizontal-meta .video-views {
  margin-right: 0;
}

/* 顶部标题和视图切换的布局 */
.collections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

/* 更多按钮 */
.more-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  color: #666;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.more-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

@media (max-width: 992px) {
  .collections-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 768px) {
  .collections-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 576px) {
  .collections-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* ==================== 合集功能样式 ==================== */

/* 合集区域 */
.collections-section {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  border-radius: 8px;
}

/* 合集头部 */
.collections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.collections-header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collections-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.new-collection-btn {
  font-size: 14px;
}

/* 宫格视图 */
.collections-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

.collection-item {
  cursor: pointer;
  transition: all 0.3s ease;
}

.collection-item:hover {
  transform: translateY(-4px);
}

.collection-item.new-collection {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 150px;
  border: 2px dashed #e0e0e0;
  border-radius: 8px;
  background-color: #fafafa;
}

.collection-item.new-collection:hover {
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.new-collection-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #999;
  font-size: 14px;
}

.collection-item.new-collection:hover .new-collection-content {
  color: #00aeec;
}

.collection-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.collection-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-count {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  border-radius: 12px;
}

.collection-info {
  margin-top: 8px;
}

.collection-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-date {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* 水平列表视图 */
.collections-horizontal {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.collection-horizontal-item {
  border-bottom: 1px solid #f0f0f0;
  padding-bottom: 24px;
}

.collection-horizontal-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.collection-horizontal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-horizontal-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.collection-video-count-badge {
  font-size: 14px;
  color: #999;
  font-weight: normal;
}

.collection-horizontal-actions {
  display: flex;
  gap: 12px;
}

.collection-horizontal-actions .action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-horizontal-actions .play-all-btn {
  background-color: #fff;
  border: 1px solid #e0e0e0;
  color: #666;
}

.collection-horizontal-actions .play-all-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.collection-horizontal-actions .more-btn {
  background-color: #fff;
  border: 1px solid #e0e0e0;
  color: #666;
}

.collection-horizontal-actions .more-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

/* 水平视频列表 */
.collection-videos-horizontal {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 8px;
}

.collection-videos-horizontal::-webkit-scrollbar {
  height: 4px;
}

.collection-videos-horizontal::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 2px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 2px;
}

/* 添加视频卡片 */
.add-video-card {
  width: 180px;
  height: 100px;
  border: 2px dashed #e0e0e0;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
  cursor: pointer;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.add-video-card:hover {
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.add-video-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #999;
  font-size: 14px;
}

.add-video-card:hover .add-video-content {
  color: #00aeec;
}

/* 视频项 */
.video-horizontal-item {
  width: 180px;
  flex-shrink: 0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.video-horizontal-item:hover {
  transform: translateY(-2px);
}

.video-horizontal-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.video-horizontal-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-horizontal-info {
  margin-top: 8px;
}

.video-horizontal-info .video-title {
  font-size: 14px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.video-horizontal-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.video-horizontal-meta .video-views {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 新建合集对话框样式 */
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

/* 合集详情视图样式 */
.collection-detail-view {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.collection-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.collection-detail-header .header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #666;
}

.collection-detail-header .header-title {
  color: #333;
  font-weight: 500;
}

.collection-detail-header .collection-name {
  color: #333;
  font-weight: 600;
}

/* 合集信息区 */
.collection-info-section {
  margin-bottom: 20px;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
}

.collection-info-section .info-container {
  display: flex;
  gap: 20px;
}

.collection-cover-large {
  position: relative;
  width: 280px;
  height: 158px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
  cursor: pointer;
}

.collection-cover-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-cover-large .cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.4);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.collection-cover-large:hover .cover-overlay {
  opacity: 1;
}

.collection-cover-large .video-count-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 13px;
  border-radius: 4px;
}

/* 合集元信息 */
.collection-meta-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.collection-title-large {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0;
  line-height: 1.4;
}

.collection-desc-large {
  font-size: 14px;
  color: #666;
  margin: 0;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.collection-meta-detail .stats-info {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #9499a0;
}

.collection-meta-detail .stats-info .stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 视频网格区 */
.collection-videos-section {
  padding: 20px;
  background-color: #fff;
  border-radius: 8px;
}

.collection-videos-section .videos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
}

.collection-videos-section .sort-options {
  display: flex;
  gap: 24px;
}

.collection-videos-section .sort-item {
  font-size: 14px;
  color: #666;
  cursor: pointer;
  padding-bottom: 4px;
  border-bottom: 2px solid transparent;
  transition: all 0.3s ease;
}

.collection-videos-section .sort-item:hover {
  color: #00aeec;
}

.collection-videos-section .sort-item.active {
  color: #00aeec;
  border-bottom-color: #00aeec;
}

.collection-videos-section .header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 视频网格 */
.collection-videos-section .videos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
}

/* 合集详情页视频卡片 */
.collection-video-card {
  display: flex;
  flex-direction: column;
  background-color: #fff;
  border-radius: 8px;
  overflow: hidden;
  transition: transform 0.3s ease;
}

.collection-video-card:hover {
  transform: translateY(-4px);
}

.collection-video-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  cursor: pointer;
  background-color: #f0f0f0;
  border-radius: 8px 8px 0 0;
}

.collection-video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.collection-video-card:hover .collection-video-cover-img {
  transform: scale(1.05);
}

.collection-video-index {
  position: absolute;
  top: 8px;
  left: 8px;
  width: 24px;
  height: 24px;
  background-color: rgba(0, 0, 0, 0.6);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  z-index: 10;
}

.collection-video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  padding: 2px 6px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  border-radius: 2px;
  z-index: 1;
}

.collection-video-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 10;
}

.collection-video-card:hover .collection-video-actions {
  opacity: 1;
}

.collection-video-info {
  padding: 10px;
  flex: 1;
  min-height: 0;
}

.collection-video-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0 0 8px 0;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #9499a0;
  flex-wrap: wrap;
}

.collection-meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.video-card.add-video-card {
  aspect-ratio: 16 / 10;
  border: 2px dashed #dcdcdc;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-card.add-video-card:hover {
  border-color: #00aeec;
  background-color: #f0f9ff;
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

.collection-cover-uploader .cover-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
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

/* 响应式设计 */
@media (max-width: 1200px) {
  .collections-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 992px) {
  .collections-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .collection-info-section .info-container {
    flex-direction: column;
  }

  .collection-cover-large {
    width: 100%;
    height: auto;
    aspect-ratio: 16 / 9;
  }
}

@media (max-width: 768px) {
  .collections-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .collection-horizontal-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .collection-detail-header .header-left {
    flex-wrap: wrap;
  }
}

@media (max-width: 576px) {
  .collections-grid {
    grid-template-columns: 1fr;
  }
}
</style>