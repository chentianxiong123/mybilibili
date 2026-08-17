<template>
  <div class="collection-management">
    <!-- 合集详情视图 -->
    <div v-if="collectionDetail.visible" class="collection-detail-view">
      <!-- 返回按钮和标题 -->
      <div class="collection-detail-header">
        <div class="header-left">
          <el-button text @click="backToCollectionsList">
            <el-icon><ArrowRight style="transform: rotate(180deg)" /></el-icon>
            返回
          </el-button>
          <span class="header-title">我的合集和视频列表</span>
          <el-icon><ArrowRight /></el-icon>
          <span class="collection-name">{{ collectionDetail.collection?.title || '加载中...' }}</span>
        </div>
        <div class="header-right">
          <el-button text @click="openEditCollectionDialog">
            <el-icon><Setting /></el-icon>
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
                <el-icon><Monitor /></el-icon>
                {{ collectionDetail.collection.viewCount || 0 }} 次观看
              </span>
              <span class="stat-item">
                更新于 {{ formatDate(collectionDetail.collection.updatedAt) }}
              </span>
              <span v-if="!collectionDetail.collection.isPublic" class="private-tag">私密合集</span>
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
          <div class="header-actions">
            <el-button type="primary" :icon="Setting" @click="openEditCollectionDialog">
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
        <div v-else-if="collectionDetail.manuscripts.length === 0" class="empty-state">
          <el-empty description="暂无视频" />
        </div>

        <!-- 视频网格 -->
        <div v-else class="videos-grid">
          <!-- 添加视频卡片 -->
          <div class="video-card add-video-card" @click="openAddVideoToCollectionDialog">
            <div class="add-video-content">
              <el-icon :size="40"><Plus /></el-icon>
              <span>添加稿件</span>
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
              <div class="collection-video-actions" @click.stop>
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
            :class="{ active: myCollections.viewType === 'grid' }"
            @click="myCollections.viewType = 'grid'"
          >
            <el-icon><Menu /></el-icon>
          </button>
          <button
            class="view-toggle-btn list-view-btn"
            :class="{ active: myCollections.viewType === 'horizontal' }"
            @click="myCollections.viewType = 'horizontal'"
          >
            <el-icon><Document /></el-icon>
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="myCollections.loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>

      <!-- 空状态 -->
      <div v-else-if="myCollections.items.length === 0" class="empty-state">
        <el-empty description="暂无合集">
          <el-button
            type="primary"
            @click="openCreateCollectionDialog"
          >
            创建合集
          </el-button>
        </el-empty>
      </div>

      <!-- 宫格视图 -->
      <div v-else-if="myCollections.viewType === 'grid'" class="collections-grid">
        <!-- 新建合集卡片 -->
        <div
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
          v-for="collection in myCollections.items"
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
      <div v-else-if="myCollections.viewType === 'horizontal'" class="collections-horizontal">
        <!-- 合集项 -->
        <div
          v-for="collection in myCollections.items"
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
              class="add-video-card"
              @click="openAddVideoDialog(collection.id)"
            >
              <div class="add-video-content">
                <el-icon :size="24"><Plus /></el-icon>
                <span>添加稿件</span>
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
                    <el-icon><Monitor /></el-icon>
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

    <!-- 新建合集对话框 -->
    <el-dialog
      v-model="createCollectionDialogVisible"
      title="新建合集"
      width="600px"
    >
      <el-form :model="createCollectionForm" label-width="80px">
        <el-form-item label="合集名称" required>
          <el-input v-model="createCollectionForm.name" placeholder="请输入合集名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="合集描述">
          <el-input
            v-model="createCollectionForm.description"
            type="textarea"
            placeholder="请输入合集描述"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="公开">
          <el-switch v-model="createCollectionForm.isPublic" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createCollectionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCollection" :loading="creatingCollection">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑合集对话框 -->
    <el-dialog
      v-model="editCollectionDialogVisible"
      title="编辑合集"
      width="600px"
    >
      <el-form :model="editCollectionForm" label-width="80px">
        <el-form-item label="合集名称" required>
          <el-input v-model="editCollectionForm.name" placeholder="请输入合集名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="合集描述">
          <el-input
            v-model="editCollectionForm.description"
            type="textarea"
            placeholder="请输入合集描述"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="公开">
          <el-switch v-model="editCollectionForm.isPublic" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editCollectionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateCollection" :loading="updatingCollection">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加视频对话框 -->
    <el-dialog
      v-model="addVideoDialogVisible"
      title="管理合集稿件"
      width="700px"
      destroy-on-close
    >
      <div class="video-selection-header">
        <span class="selection-tip">勾选稿件添加到合集，取消勾选从合集移除</span>
      </div>

      <div v-if="addingVideo" class="dialog-loading">
        <el-skeleton :rows="5" animated />
      </div>

      <div v-else-if="availableVideos.length === 0" class="dialog-empty">
        <el-empty description="暂无可添加的稿件" />
      </div>

      <div v-else class="video-select-list">
        <el-checkbox-group v-model="selectedVideos">
          <div v-for="video in availableVideos" :key="video.id" class="video-select-item">
            <el-checkbox :label="video.id">
              <div class="video-select-content">
                <img loading="lazy" decoding="async" :src="video.coverUrl || getDefaultCover()" class="video-select-cover" />
                <div class="video-select-info">
                  <div class="video-select-title">{{ video.title }}</div>
                  <div class="video-select-meta">{{ formatNumber(video.viewCount) }} 播放</div>
                </div>
              </div>
            </el-checkbox>
            <span v-if="selectedVideos.includes(video.id)" class="in-collection-badge">已加入</span>
          </div>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="addVideoDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddVideoToCollection" :loading="addingVideo">更新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { manuscriptApi, collectionApi } from '@/api/creator'
import { useUserStore } from '@/stores/user'
import {
  VideoPlay,
  Setting,
  Menu,
  Document,
  Monitor,
  Plus,
  MoreFilled,
  ArrowRight
} from '@element-plus/icons-vue'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()
const userStore = useUserStore()

// 获取当前用户ID
const getCurrentUserId = () => {
  const userStr = localStorage.getItem('user')
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

// 合集管理相关状态
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

const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

const loadUserCollections = async () => {
  console.log('【调试】loadUserCollections 被调用')
  const userId = getCurrentUserId()
  console.log('【调试】currentUser:', userId)
  console.log('【调试】getCurrentUserId():', userId)
  if (!userId) {
    console.log('【调试】用户ID为空，直接返回')
    return
  }

  myCollections.value.loading = true
  try {
    console.log('【调试】开始获取合集列表，用户ID:', userId)
    const response = await collectionApi.getUserCollections(userId, 1, 100)
    console.log('【调试】合集列表响应:', response)
    if (response.code === 200) {
      const list = response.data || []
      for (const collection of list) {
        try {
          const videoResponse = await collectionApi.getCollectionManuscripts(collection.id, 1, 10)
          if (videoResponse.code === 200) {
            collection.videos = (videoResponse.data || []).map(video => ({
              ...video,
              date: formatDate(video.uploadTime)
            }))
            // 如果合集没有封面，用第一个视频的封面
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
      collectionDetail.value.collection = response.data
    }

    const videoResponse = await collectionApi.getCollectionManuscripts(collectionId, 1, 20)
    if (videoResponse.code === 200) {
      collectionDetail.value.manuscripts = videoResponse.data || []
      collectionDetail.value.pagination.total = videoResponse.data?.length || 0
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

const handleCreateCollectionCoverChange = (file) => {
  const isImage = file.raw.type.startsWith('image/')
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('封面图片只能是图片格式!')
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

const handleEditCollectionCoverChange = (file) => {
  const isImage = file.raw.type.startsWith('image/')
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('封面图片只能是图片格式!')
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

const handleCollectionCommand = async (command) => {
  if (command === 'delete') {
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
    console.log('开始获取数据，合集ID:', collectionId)

    const [manuscriptsRes, videosRes] = await Promise.all([
      collectionApi.getCollectionManuscripts(collectionId, 1, 100),
      manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    ])

    console.log('合集稿件响应:', manuscriptsRes)
    console.log('用户稿件响应:', videosRes)

    if (videosRes.code === 200) {
      const data = videosRes.data || {}
      console.log('data对象的keys:', Object.keys(data))
      console.log('data对象的完整结构:', data)

      const list = data.list || data.records || data.items || data.data || []
      console.log('处理前的稿件列表:', list)

      availableVideos.value = list.map(v => ({
        ...v,
        id: v.manuscriptId || v.id,
        coverUrl: v.coverUrl || v.cover_url,
        viewCount: v.viewCount || v.view_count || 0,
        title: v.title || '无标题'
      }))

      console.log('处理后的稿件列表:', availableVideos.value)
    } else {
      console.error('获取用户稿件失败:', videosRes.message)
      availableVideos.value = []
    }

    if (manuscriptsRes.code === 200) {
      const manuscripts = manuscriptsRes.data || []
      selectedVideos.value = manuscripts.map(m => m.manuscriptId || m.id)
      console.log('当前合集中的稿件IDs:', selectedVideos.value)
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

const handleVideoCommand = async (command, manuscript) => {
  if (command === 'remove') {
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

// 监听激活状态，激活时加载合集数据
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

.collection-detail-view {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
}

.collection-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e0e0e0;
}

.collection-detail-header .header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.collection-detail-header .header-title {
  font-size: 14px;
  color: #606266;
}

.collection-detail-header .collection-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.collection-info-section {
  margin-bottom: 24px;
}

.collection-info-section .info-container {
  display: flex;
  gap: 24px;
}

.collection-cover-large {
  width: 240px;
  height: 135px;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
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
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
  cursor: pointer;
}

.collection-cover-large:hover .cover-overlay {
  opacity: 1;
}

.collection-cover-large .video-count-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-meta-detail {
  flex: 1;
}

.collection-title-large {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.collection-desc-large {
  font-size: 14px;
  color: #606266;
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.collection-meta-detail .stats-info {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #909399;
}

.collection-meta-detail .stats-info .stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-meta-detail .private-tag {
  background: #f4f4f5;
  color: #909399;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.collection-videos-section {
  margin-top: 24px;
}

.collection-videos-section .videos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-videos-section .sort-options {
  display: flex;
  gap: 16px;
}

.collection-videos-section .sort-item {
  font-size: 14px;
  color: #909399;
  cursor: pointer;
  padding: 4px 0;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.collection-videos-section .sort-item.active {
  color: #00a1d6;
  border-bottom-color: #00a1d6;
}

.collection-videos-section .sort-item:hover {
  color: #00a1d6;
}

.collection-videos-section .header-actions {
  display: flex;
  gap: 8px;
}

.collection-videos-section .videos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.collection-video-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}

.collection-video-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.collection-video-cover {
  position: relative;
  aspect-ratio: 16/9;
  cursor: pointer;
}

.collection-video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-index {
  position: absolute;
  top: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.collection-video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.collection-video-actions {
  position: absolute;
  top: 8px;
  right: 8px;
}

.collection-video-info {
  padding: 12px;
}

.collection-video-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
}

.collection-meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.add-video-card {
  aspect-ratio: 16/9;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
}

.add-video-card:hover {
  border-color: #00a1d6;
  background: #f0f9ff;
}

.add-video-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.add-video-card:hover .add-video-content {
  color: #00a1d6;
}

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
  color: #303133;
  margin: 0;
}

.view-toggle {
  display: flex;
  gap: 4px;
}

.view-toggle-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #dcdfe6;
  background: #fff;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.view-toggle-btn.active {
  background: #00a1d6;
  border-color: #00a1d6;
  color: #fff;
}

.collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 280px));
  gap: 16px;
  justify-content: start;
}

.collection-item {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  max-width: 280px;
  width: 100%;
}

.collection-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.collection-item.new-collection {
  border: 2px dashed #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.collection-item.new-collection:hover {
  border-color: #00a1d6;
  background: #f0f9ff;
}

.new-collection-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.collection-item.new-collection:hover .new-collection-content {
  color: #00a1d6;
}

.collection-cover {
  position: relative;
  aspect-ratio: 16/9;
  width: 100%;
  overflow: hidden;
}

.collection-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-count {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-info {
  padding: 12px;
}

.collection-info .collection-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-info .collection-date {
  font-size: 12px;
  color: #909399;
}

.collections-horizontal {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.collection-horizontal-item {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.collection-horizontal-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
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
  color: #303133;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-count-badge {
  background: #f4f4f5;
  color: #909399;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: normal;
  flex-shrink: 0;
}

.collection-horizontal-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.collection-horizontal-actions .action-btn {
  padding: 6px 12px;
  font-size: 13px;
}

.collection-videos-horizontal {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
  scrollbar-width: thin;
  scrollbar-color: #e0e0e0 #f5f7fa;
}

.collection-videos-horizontal::-webkit-scrollbar {
  height: 6px;
}

.collection-videos-horizontal::-webkit-scrollbar-track {
  background: #f5f7fa;
  border-radius: 3px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 3px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb:hover {
  background: #d0d0d0;
}

.collection-videos-horizontal .add-video-card {
  min-width: 160px;
  height: 100px;
  flex-shrink: 0;
  max-width: 160px;
}

.video-horizontal-item {
  min-width: 160px;
  max-width: 160px;
  flex-shrink: 0;
  cursor: pointer;
  transition: transform 0.3s ease;
}

.video-horizontal-item:hover {
  transform: translateY(-2px);
}

.video-horizontal-cover {
  position: relative;
  aspect-ratio: 16/9;
  border-radius: 6px;
  overflow: hidden;
  background: #f0f2f5;
}

.video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.video-horizontal-item:hover .video-cover-img {
  transform: scale(1.05);
}

.video-horizontal-cover .video-duration {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 1px 4px;
  border-radius: 2px;
  font-size: 11px;
  z-index: 1;
}

.video-horizontal-info {
  padding: 8px 0;
}

.video-horizontal-info .video-title {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
  line-height: 1.3;
}

.video-horizontal-meta {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #909399;
  align-items: center;
}

.video-horizontal-meta .video-views {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-horizontal-meta .video-date {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
  width: 80px;
  text-align: right;
}

.loading-state {
  padding: 40px;
  text-align: center;
  color: #909399;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

.video-select-list {
  max-height: 400px;
  overflow-y: auto;
}

.video-selection-header {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.selection-tip {
  font-size: 13px;
  color: #606266;
}

.dialog-loading,
.dialog-empty {
  padding: 40px 0;
  text-align: center;
}

.video-select-item {
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.video-select-item:last-child {
  border-bottom: none;
}

.video-select-item .el-checkbox {
  flex: 1;
}

.video-select-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.video-select-cover {
  width: 80px;
  height: 45px;
  object-fit: cover;
  border-radius: 4px;
}

.in-collection-badge {
  background: linear-gradient(135deg, #005980, #1890ff);
  color: #ffffff;
  padding: 6px 14px;
  border-radius: 4px;
  font-size: 12px;
  flex-shrink: 0;
  margin-left: 8px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.video-select-info {
  flex: 1;
}

.video-select-title {
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
}

.video-select-meta {
  font-size: 12px;
  color: #909399;
}
</style>