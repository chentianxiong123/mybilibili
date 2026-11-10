<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Search } from '@element-plus/icons-vue'
import { collectionApi } from '@/api/collection.ts'
import { manuscriptApi } from '@/api/manuscript.ts'

const route = useRoute()
const router = useRouter()

// 可选 userId prop：嵌入空间页 tab 时由父组件传入空间主人的 id；
// 独立路由 /collections 不传，回落到 route.params.userId / 当前登录用户（保持原行为）。
const props = defineProps({
  userId: { type: [String, Number], default: null }
})

const { user: currentUser } = useAuth()
const userId = ref(props.userId || route.params.userId || currentUser.value?.id)
const currentUserId = computed(() => currentUser.value?.id)

const isOwnSpace = computed(() => {
  return currentUserId.value && userId.value && String(currentUserId.value) === String(userId.value)
})

const loading = ref(false)

// 合集分组（对齐 B站 space.bilibili.com/{uid}/lists 的 .video-list 分组模型）
const groups = ref([])

const createDialogVisible = ref(false)
const creating = ref(false)
const newCollection = ref({
  name: '',
  description: '',
  isPublic: true,
  manuscriptIds: []
})

const userManuscripts = ref([])
const manuscriptLoading = ref(false)
const searchKeyword = ref('')

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 格式化时长
const formatDuration = (seconds) => {
  if (!seconds) return '00:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

// 格式化播放量（万）
const formatView = (n) => {
  const v = Number(n) || 0
  if (v >= 10000) return (v / 10000).toFixed(1).replace(/\.0$/, '') + '万'
  return String(v)
}

const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e3e5e7" width="400" height="225"/></svg>')
}

// 合集 API 返回 snake_case，归一化到 camelCase
const normalizeCollection = (d) => ({
  id: d.id,
  name: d.title || d.name || '',
  description: d.description || '',
  coverUrl: d.cover_url || d.coverUrl || '',
  isPublic: d.status === 1 || d.isPublic === true,
  videoCount: d.manuscript_count || d.videoCount || 0,
  viewCount: d.view_count || d.viewCount || 0,
  createTime: d.created_at || d.createTime || '',
  updateTime: d.updated_at || d.updateTime || '',
  userId: d.user_id || d.userId || null,
  userName: d.user_name || d.userName || '',
  userAvatar: d.user_avatar || d.userAvatar || ''
})

// 加载合集 + 每个合集的稿件，构建 B站 .video-list 分组
const loadGroups = async () => {
  if (!userId.value) return
  loading.value = true
  try {
    const response = await collectionApi.getUserCollections(userId.value, 1, 20)
    if (response.code === 200) {
      const collections = (response.data?.list || []).map(normalizeCollection)
      // 并发拉取每个合集的前 6 个视频（B站 每组最多展示 6 个，其余「查看更多」）
      const built = await Promise.all(collections.map(async (c) => {
        let videos = []
        try {
          const r = await collectionApi.getCollectionManuscripts(c.id, 1, 6)
          if (r.code === 200) videos = r.data?.list || []
        } catch (e) {
          console.error('获取合集视频失败:', e)
        }
        return { ...c, videos }
      }))
      groups.value = built
    }
  } catch (error) {
    console.error('获取合集列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 跳转到合集详情
const goToCollectionDetail = (collectionId) => {
  router.push(`/collection/${collectionId}`)
}

// 播放全部：进详情页（详情页有播放全部）
const playAll = (collection) => {
  const first = collection.videos && collection.videos[0]
  if (first && first.id) {
    router.push(`/manuscript/${first.id}`)
  } else {
    router.push(`/collection/${collection.id}`)
  }
}

// 打开创建对话框
const openCreateDialog = async () => {
  newCollection.value = { name: '', description: '', isPublic: true, manuscriptIds: [] }
  searchKeyword.value = ''
  userManuscripts.value = []
  createDialogVisible.value = true
  await loadUserManuscripts()
}

const loadUserManuscripts = async () => {
  if (!currentUserId.value) return
  manuscriptLoading.value = true
  try {
    const response = await manuscriptApi.getUserManuscripts(currentUserId.value, 1, 100)
    if (response.code === 200) {
      userManuscripts.value = response.data || []
    }
  } catch (error) {
    console.error('获取用户稿件失败:', error)
  } finally {
    manuscriptLoading.value = false
  }
}

const handleSearchManuscripts = () => {
  if (!searchKeyword.value.trim()) {
    loadUserManuscripts()
    return
  }
  const keyword = searchKeyword.value.toLowerCase()
  userManuscripts.value = userManuscripts.value.filter(m =>
    m.title.toLowerCase().includes(keyword)
  )
}

// 创建合集
const handleCreateCollection = async () => {
  if (!newCollection.value.name.trim()) {
    ElMessage.warning('请输入合集名称')
    return
  }
  creating.value = true
  try {
    const response = await collectionApi.createCollection({
      name: newCollection.value.name,
      description: newCollection.value.description,
      isPublic: newCollection.value.isPublic,
      manuscriptIds: newCollection.value.manuscriptIds
    })
    if (response.code === 200) {
      ElMessage.success('创建成功')
      createDialogVisible.value = false
      loadGroups()
      if (response.data?.id) {
        router.push(`/collection/${response.data.id}/edit`)
      }
    } else {
      ElMessage.error(response.message || '创建失败')
    }
  } catch (error) {
    console.error('创建合集失败:', error)
    ElMessage.error('创建合集失败')
  } finally {
    creating.value = false
  }
}

// 删除合集
const handleDeleteCollection = async (collection, event) => {
  event.stopPropagation()
  try {
    await ElMessageBox.confirm(
      `确定要删除合集 "${collection.name}" 吗？此操作不可恢复。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    const response = await collectionApi.deleteCollection(collection.id)
    if (response.code === 200) {
      ElMessage.success('删除成功')
      loadGroups()
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

onMounted(() => {
  loadGroups()
})
</script>

<template>
  <div class="space-lists">
    <!-- 头部：标题 + 操作（B站 .lists-header） -->
    <div class="lists-header">
      <div class="title">合集和系列</div>
      <div class="actions">
        <button v-if="isOwnSpace" class="vui_button action-btn" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          <span>新建合集</span>
        </button>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="lists-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 空状态 -->
      <div v-else-if="groups.length === 0" class="empty-state">
        <el-empty description="暂无合集">
          <button v-if="isOwnSpace" class="vui_button action-btn" @click="openCreateDialog">创建合集</button>
        </el-empty>
      </div>

      <!-- 合集分组（B站 .video-list） -->
      <div v-for="group in groups" :key="group.id" class="video-list">
        <div class="video-list__header">
          <div class="video-list__title" @click="goToCollectionDetail(group.id)">{{ group.name }}</div>
          <div class="video-list__desc">·{{ group.videoCount || group.videos.length }}</div>
          <div class="video-list__actions">
            <button class="vui_button action-btn playall-btn" @click="playAll(group)">
              <i class="iconfont icon-bofang"></i>
              <span>播放全部</span>
            </button>
            <button class="vui_button" @click="goToCollectionDetail(group.id)">查看更多</button>
            <button v-if="isOwnSpace" class="vui_button icon-btn" @click.stop="router.push(`/collection/${group.id}/edit`)">
              <el-icon><Edit /></el-icon>
            </button>
            <button v-if="isOwnSpace" class="vui_button icon-btn" @click.stop="handleDeleteCollection(group, $event)">
              <el-icon><Delete /></el-icon>
            </button>
          </div>
        </div>
        <div class="video-list__content">
          <div
            v-for="video in group.videos"
            :key="video.id"
            class="video-list__item"
            @click="router.push(`/manuscript/${video.id}`)"
          >
            <div class="bili-video-card">
              <div class="bili-video-card__wrap">
                <div class="bili-video-card__cover">
                  <a class="bili-cover-card">
                    <div class="bili-cover-card__thumbnail">
                      <img :src="video.coverUrl || getDefaultCover()" :alt="video.title" loading="lazy" decoding="async" />
                    </div>
                    <div class="bili-cover-card__stats">
                      <div class="bili-cover-card__stat">
                        <i class="iconfont icon-bofangfangkuai"></i>
                        <span>{{ formatView(video.viewCount) }}</span>
                      </div>
                      <div class="bili-cover-card__stat">
                        <span>{{ formatDuration(video.duration || video.durationSeconds) }}</span>
                      </div>
                    </div>
                  </a>
                  <div class="bili-card-watch-later">
                    <div class="bili-card-watch-later__btn">
                      <i class="iconfont icon-bofang"></i>
                    </div>
                    <span class="bili-card-watch-later__tip">稍后再看</span>
                  </div>
                </div>
                <div class="bili-video-card__details">
                  <div class="bili-video-card__title" :title="video.title">
                    <a>{{ video.title }}</a>
                  </div>
                  <div class="bili-video-card__subtitle">
                    <span>{{ formatDate(video.uploadTime || video.createdAt) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建合集对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="新建合集"
      width="700px"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item label="合集名称" required>
          <el-input v-model="newCollection.name" placeholder="请输入合集名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="合集描述">
          <el-input v-model="newCollection.description" type="textarea" :rows="3" placeholder="请输入合集描述（选填）" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="newCollection.isPublic">公开合集</el-checkbox>
        </el-form-item>
        <el-form-item label="选择视频（可选）">
          <div class="manuscript-selector">
            <div class="search-bar">
              <el-input v-model="searchKeyword" placeholder="搜索视频标题" :prefix-icon="Search" clearable @keyup.enter="handleSearchManuscripts" @clear="loadUserManuscripts" />
            </div>
            <div v-if="manuscriptLoading" class="manuscript-loading">
              <el-skeleton :rows="3" animated />
            </div>
            <div v-else-if="userManuscripts.length === 0" class="manuscript-empty">
              <el-empty description="暂无可添加的视频" :image-size="60" />
            </div>
            <div v-else class="manuscript-list">
              <el-checkbox-group v-model="newCollection.manuscriptIds">
                <div v-for="manuscript in userManuscripts" :key="manuscript.id" class="manuscript-item">
                  <el-checkbox :label="manuscript.id">
                    <div class="manuscript-content">
                      <img loading="lazy" decoding="async" :src="manuscript.coverUrl || getDefaultCover()" :alt="manuscript.title" class="manuscript-cover" />
                      <div class="manuscript-info">
                        <span class="manuscript-title">{{ manuscript.title }}</span>
                        <span class="manuscript-date">{{ formatDate(manuscript.uploadTime) }}</span>
                      </div>
                    </div>
                  </el-checkbox>
                </div>
              </el-checkbox-group>
            </div>
            <div v-if="newCollection.manuscriptIds.length > 0" class="selected-count">
              已选择 {{ newCollection.manuscriptIds.length }} 个视频
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateCollection">创建合集</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
/* ===== 对齐 space.bilibili.com/{uid}/lists 合集和系列概览页 =====
   参考值来自真实 B 站 getComputedStyle 抓取（1920×1080 / dpr1）：
   space-lists 1640 居中 / lists-header 1640x64(pad-b 30) / title 24-600-34 #18191c
   video-list 1640(pad-b 33 margin-b 30) / header 34(flex mb20 #61666d)
   title 20-400-20 #18191c / desc 16-400-16 #61666d margin 0 8 0 6
   actions 右 / playall-btn + 查看更多 100x34 fs14 border #e3e5e7 r8
   content grid col-gap 16 / card 260x212 / cover 260x146 r6 #e3e5e7 */
.space-lists {
  width: 1640px;
  margin: 0 auto;
  padding: 20px 0 60px;
}

/* 头部（1640x64：title 34 + padding-bottom 30） */
.lists-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 30px;
}

.lists-header .title {
  font-size: 24px;
  font-weight: 600;
  line-height: 34px;
  color: #18191c;
}

.lists-header .actions {
  display: flex;
  align-items: center;
}

/* B站 vui_button（100x34 fs14 border #e3e5e7 r8） */
.vui_button {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  padding: 0 12px;
  font-size: 14px;
  font-weight: 400;
  line-height: 14px;
  color: #18191c;
  background-color: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  cursor: pointer;
}

.vui_button:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.vui_button .iconfont,
.vui_button .el-icon {
  font-size: 14px;
  margin-right: 4px;
}

.vui_button.icon-btn {
  width: 34px;
  padding: 0;
  margin-left: 8px;
}

.vui_button.icon-btn .el-icon {
  margin-right: 0;
}

/* 内容区 */
.lists-content {
  min-height: 200px;
}

/* 合集分组（1640，padding-bottom 33 + margin-bottom 30 = 组间距 330） */
.video-list {
  padding-bottom: 33px;
  margin-bottom: 30px;
}

.video-list:last-child {
  margin-bottom: 0;
}

/* 分组头（1640x34 flex 居中 mb20） */
.video-list__header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  font-size: 16px;
  line-height: 16px;
  color: #61666d;
}

.video-list__title {
  font-size: 20px;
  font-weight: 400;
  line-height: 20px;
  color: #18191c;
  cursor: pointer;
}

.video-list__title:hover {
  color: #00aeec;
}

.video-list__desc {
  margin: 0 8px 0 6px;
  font-size: 16px;
  font-weight: 400;
  line-height: 16px;
  color: #61666d;
}

.video-list__actions {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.video-list__actions .vui_button {
  margin-left: 12px;
}

.video-list__actions .vui_button:first-child {
  margin-left: 0;
}

/* 视频网格（B站 .video-list__content grid col-gap 16，6 列，卡 260x212） */
.video-list__content {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
}

.video-list__item {
  cursor: pointer;
}

.bili-video-card {
  width: 100%;
}

.bili-video-card__wrap {
  width: 100%;
}

/* 封面（260x146 r6 底 #e3e5e7） */
.bili-video-card__cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 6px;
  overflow: hidden;
  background-color: #e3e5e7;
}

.bili-cover-card {
  display: block;
  width: 100%;
  height: 100%;
  text-decoration: none;
}

.bili-cover-card__thumbnail {
  width: 100%;
  height: 100%;
}

.bili-cover-card__thumbnail img {
  display: inline-block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 6px;
}

/* 封面底部统计（260x34 白字 12-12，圆角 0 0 6 6） */
.bili-cover-card__stats {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  height: 34px;
  padding: 14px 8px 4px 8px;
  border-radius: 0 0 6px 6px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.6) 100%);
  color: #fff;
}

.bili-cover-card__stat {
  display: flex;
  align-items: center;
  margin-right: 10px;
  font-size: 12px;
  font-weight: 400;
  line-height: 12px;
  color: #fff;
}

.bili-cover-card__stat .iconfont {
  font-size: 12px;
  margin-right: 2px;
}

.bili-cover-card__stat:last-child {
  margin-right: 0;
}

/* 稍后再看（28x28 圆角，hover 显现） */
.bili-card-watch-later {
  position: absolute;
  top: 6px;
  right: 6px;
  display: flex;
  align-items: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background-color: rgba(0, 0, 0, 0.6);
  color: #fff;
  opacity: 0;
  transition: opacity 0.3s;
}

.video-list__item:hover .bili-card-watch-later {
  opacity: 1;
}

.bili-card-watch-later__btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.bili-card-watch-later__btn .iconfont {
  font-size: 16px;
}

/* 卡片详情：标题 + 日期（B站 margin-top 10px → 卡片总高 212） */
.bili-video-card__details {
  margin-top: 10px;
}

.bili-video-card__title {
  display: flow-root;
  max-height: 36px;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bili-video-card__title a {
  color: #18191c;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  text-decoration: none;
}

.video-list__item:hover .bili-video-card__title a {
  color: #00aeec;
}

.bili-video-card__subtitle {
  display: flex;
  margin-top: 4px;
  font-size: 12px;
  font-weight: 400;
  line-height: 16px;
  color: #9499a0;
}

.bili-video-card__subtitle span {
  color: #9499a0;
}

/* 加载 / 空状态 */
.loading-state {
  padding: 20px;
}

.empty-state {
  padding: 60px 0;
}

/* 稿件选择器（创建对话框内，隐藏态不影响像素布局） */
.manuscript-selector {
  width: 100%;
}

.manuscript-selector .search-bar {
  margin-bottom: 12px;
}

.manuscript-loading,
.manuscript-empty {
  padding: 20px 0;
}

.manuscript-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

.manuscript-item {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.manuscript-item:last-child {
  border-bottom: none;
}

.manuscript-content {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: 8px;
}

.manuscript-cover {
  width: 80px;
  height: 45px;
  border-radius: 4px;
  object-fit: cover;
}

.manuscript-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.manuscript-title {
  font-size: 14px;
  color: #333;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.manuscript-date {
  font-size: 12px;
  color: #9499a0;
}

.selected-count {
  margin-top: 8px;
  padding: 8px 12px;
  background-color: #e6f7ff;
  border-radius: 4px;
  font-size: 13px;
  color: #1890ff;
}
</style>
