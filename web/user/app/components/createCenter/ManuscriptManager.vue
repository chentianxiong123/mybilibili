<template>
  <div class="manuscript-manager">
    <!-- 主要选择栏 -->
    <div class="content-tabs">
      <el-tabs v-model="mainTab" type="card">
        <el-tab-pane label="视频管理" name="video"></el-tab-pane>
        <el-tab-pane label="合集管理" name="collection"></el-tab-pane>
      </el-tabs>
    </div>

    <div v-show="mainTab === 'video'">
      <!-- 工具栏：对齐B站 article-header（一层32px） -->
      <div class="video-toolbar">
        <!-- 左：全部稿件 / 草稿 + 进行中/已通过/未通过/已下架 -->
        <div class="toolbar-left">
          <div class="main-tabs">
            <span
              class="status-tab"
              :class="{ active: statusFilter === 'all' }"
              @click="setStatus('all')"
            >
              全部稿件 {{ totalCount }}
            </span>
            <span
              class="status-tab"
              @click="goDrafts"
            >
              草稿
            </span>
          </div>
          <div class="second-tabs">
            <span
              class="second-tab"
              :class="{ active: statusFilter === 'processing' }"
              @click="setStatus('processing')"
            >
              进行中 {{ processingCount }}
            </span>
            <span
              class="second-tab"
              :class="{ active: statusFilter === 'published' }"
              @click="setStatus('published')"
            >
              已通过 {{ approvedCount }}
            </span>
            <span
              class="second-tab"
              :class="{ active: statusFilter === 'rejected' }"
              @click="setStatus('rejected')"
            >
              未通过 {{ rejectedCount }}
            </span>
            <span
              class="second-tab"
              :class="{ active: statusFilter === 'unpublished' }"
              @click="setStatus('unpublished')"
            >
              已下架 {{ unpublishedCount }}
            </span>
          </div>
        </div>

        <!-- 右：搜索 + 排序 -->
        <div class="toolbar-right">
          <div class="search-box">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索稿件"
              clearable
              class="search-input"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </div>
          <el-select v-model="sortBy" class="sort-select">
            <el-option label="投稿时间排序" value="latest"></el-option>
            <el-option label="最早投稿" value="oldest"></el-option>
            <el-option label="播放量排序" value="views"></el-option>
          </el-select>
        </div>
      </div>

      <!-- 视频稿件列表 -->
      <div class="article-list" v-loading="articlesLoading">
        <div v-if="filteredArticles.length" class="manuscript-cards">
          <div
            v-for="article in filteredArticles"
            :key="article.id"
            class="manuscript-card"
          >
            <!-- 封面 -->
            <a class="card-cover" @click="goWatch(article)">
              <img
                v-if="article.coverUrl"
                :src="article.coverUrl"
                :alt="article.title"
                class="cover-img"
              />
              <div v-else class="cover-fallback">{{ getDefaultCoverText() }}</div>
              <span v-if="article.status === 1" class="cover-tag">仅审核中</span>
              <span v-if="article.duration || article.durationSeconds" class="cover-duration">
                {{ formatDuration(article.duration, article.durationSeconds) }}
              </span>
              <div class="cover-play">
                <el-icon><VideoPlay /></el-icon>
              </div>
            </a>

            <!-- 信息 -->
            <div class="card-info">
              <div class="info-top">
                <span class="card-title" @click="goWatch(article)">{{ article.title }}</span>
                <div class="card-ops">
                  <span class="status-pill" :class="statusPillClass(article.status)">
                    {{ getArticleStatusText(article.status) }}
                  </span>
                </div>
              </div>

              <div class="info-meta">
                <span class="meta-date">{{ formatDate(article.uploadTime || article.createdAt) }}</span>
                <span v-if="hasProcessingInfo(article)" class="meta-process">
                  <template v-for="(video, vi) in getArticleVideos(article)" :key="video.id || video.videoOrder">
                    <span class="process-chip" :class="processChipClass(video, article.status)">
                      {{ getVideoPartTitle(video) }}
                      <span v-if="article.status === 1" class="process-percent">
                        {{ getVideoProcessProgress(video) }}%
                      </span>
                    </span>
                  </template>
                </span>
                <span class="meta-spacer"></span>
                <div class="card-actions">
                  <button v-if="article.status !== 1" class="bili-btn" @click="openEditDialog(article.id)">
                    <el-icon><Edit /></el-icon>编辑
                  </button>
                  <el-dropdown
                    v-if="hasCardActions(article.status)"
                    trigger="click"
                    @command="(cmd) => handleCardCommand(cmd, article)"
                  >
                    <button class="bili-more-btn">
                      <el-icon><MoreFilled /></el-icon>
                    </button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item v-if="article.status === 3" command="unpublish">
                          下架
                        </el-dropdown-item>
                        <el-dropdown-item v-if="article.status === -1" command="publish">
                          上架
                        </el-dropdown-item>
                        <el-dropdown-item v-if="article.status === 0 || article.status === 1" disabled>
                          {{ article.status === 0 ? '审核中...' : '处理中...' }}
                        </el-dropdown-item>
                      <el-dropdown-item command="delete" divided>
                        删除
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
              </div>

              <div class="meta-foot">
                <div class="info-stats">
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.viewCount) }}</span>
                    <em>播放</em>
                  </div>
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.danmakuCount) }}</span>
                    <em>弹幕</em>
                  </div>
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.commentCount) }}</span>
                    <em>评论</em>
                  </div>
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.coinCount) }}</span>
                    <em>投币</em>
                  </div>
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.collectCount) }}</span>
                    <em>收藏</em>
                  </div>
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.likeCount) }}</span>
                    <em>点赞</em>
                  </div>
                  <div class="stat-item">
                    <span class="stat-num">{{ formatCount(article.shareCount) }}</span>
                    <em>分享</em>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <el-empty
          v-else-if="!articlesLoading"
          description="暂无稿件"
        />
      </div>

      <!-- 分页导航栏 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          layout="prev, pager, next, jumper, total"
          :total="totalArticles"
          @current-change="handleCurrentChange"
        ></el-pagination>
      </div>
    </div>

    <!-- 合集管理内容 -->
    <CollectionManager
      v-show="mainTab === 'collection'"
      :active="propsActive && mainTab === 'collection'"
    />

    <!-- 编辑稿件对话框 -->
    <ManuscriptEditDialog
      v-model="editDialogVisible"
      :manuscript-id="editManuscriptId"
      @saved="handleEditSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  VideoPlay,
  Edit,
  MoreFilled
} from '@element-plus/icons-vue'
import { manuscriptApi } from '@/api/creator'
import CollectionManager from './CollectionManager.vue'
import ManuscriptEditDialog from './ManuscriptEditDialog.vue'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const propsActive = computed(() => props.active)

const route = useRoute()
const router = useRouter()

// 主要选择栏
const mainTab = ref('video')

// 状态筛选
const statusFilter = ref('all')
const validArticleStatusFilters = new Set(['all', 'published', 'processing', 'rejected', 'unpublished'])

// 搜索与排序
const searchKeyword = ref('')
const sortBy = ref('latest')

watch(
  () => [route.path, route.query.status],
  () => {
    if (route.path !== '/platform/content-articles') {
      return
    }
    const queryStatus = Array.isArray(route.query.status) ? route.query.status[0] : route.query.status
    if (validArticleStatusFilters.has(queryStatus) && statusFilter.value !== queryStatus) {
      statusFilter.value = queryStatus
    }
  },
  { immediate: true }
)

// 稿件列表数据
const articles = ref([])
const articlesLoading = ref(false)
const totalArticles = ref(0)
let articleProcessPollingTimer = null

// 稿件统计数据
const approvedCount = ref(0)
const rejectedCount = ref(0)
const processingCount = ref(0)
const unpublishedCount = ref(0)

const totalCount = computed(() =>
  approvedCount.value + processingCount.value + rejectedCount.value + unpublishedCount.value
)

// 分页相关状态
const currentPage = ref(1)
const pageSize = ref(10)

// 编辑稿件对话框状态
const editDialogVisible = ref(false)
const editManuscriptId = ref(null)

const openEditDialog = (id) => {
  editManuscriptId.value = id
  editDialogVisible.value = true
}

// 跳转到播放页（已通过稿件才可看）
const goWatch = (article) => {
  if (article.status !== 3) {
    return
  }
  const id = article.firstVideoId || article.id
  if (id) {
    window.location.href = `/video/${id}`
  }
}

// 当前页数据（前端搜索过滤 + 排序）
const filteredArticles = computed(() => {
  let list = articles.value
  const kw = searchKeyword.value.trim()
  if (kw) {
    list = list.filter((a) => (a.title || '').toLowerCase().includes(kw.toLowerCase()))
  }
  const sorted = [...list]
  if (sortBy.value === 'latest') {
    sorted.sort((a, b) => timestampOf(b) - timestampOf(a))
  } else if (sortBy.value === 'oldest') {
    sorted.sort((a, b) => timestampOf(a) - timestampOf(b))
  } else if (sortBy.value === 'views') {
    sorted.sort((a, b) => (b.viewCount || 0) - (a.viewCount || 0))
  }
  return sorted
})

const timestampOf = (a) => {
  const t = a.uploadTime || a.createdAt
  const n = new Date(t).getTime()
  return Number.isFinite(n) ? n : 0
}

// 获取稿件列表
const fetchArticles = async () => {
  articlesLoading.value = true
  await loadArticles(false)
}

const refreshArticlesSilently = async () => {
  await loadArticles(true)
}

const loadArticles = async (silent = false) => {
  if (!silent) {
    articlesLoading.value = true
  }
  try {
    const params = {
      page: currentPage.value,
      size: pageSize.value
    }
    if (statusFilter.value && statusFilter.value !== 'all') {
      params.status = statusFilter.value
    }

    const response = await manuscriptApi.getMyManuscripts(params)

    if (response.code === 200) {
      articles.value = response.data.list || []
      totalArticles.value = response.data.total || 0
      syncArticleProcessPolling()
    }
  } catch (error) {
    console.error('获取稿件列表失败:', error)
    if (!silent) {
      ElMessage.error('获取稿件列表失败')
    }
  } finally {
    if (!silent) {
      articlesLoading.value = false
    }
  }
}

const startArticleProcessPolling = () => {
  if (articleProcessPollingTimer) {
    return
  }
  articleProcessPollingTimer = window.setInterval(() => {
    if (props.active && mainTab.value === 'video') {
      refreshArticlesSilently()
    }
  }, 5000)
}

const stopArticleProcessPolling = () => {
  if (!articleProcessPollingTimer) {
    return
  }
  window.clearInterval(articleProcessPollingTimer)
  articleProcessPollingTimer = null
}

const syncArticleProcessPolling = () => {
  const shouldPoll = props.active
    && mainTab.value === 'video'
    && articles.value.some(article => article.status === 0 || article.status === 1)
  if (shouldPoll) {
    startArticleProcessPolling()
  } else {
    stopArticleProcessPolling()
  }
}

// 获取稿件统计
const fetchManuscriptStats = async () => {
  try {
    const response = await manuscriptApi.getMyStats()

    if (response.code === 200) {
      const stats = response.data
      processingCount.value = stats.processing || 0
      approvedCount.value = stats.published || 0
      rejectedCount.value = stats.rejected || 0
      unpublishedCount.value = stats.unpublished || 0
    }
  } catch (error) {
    console.error('获取稿件统计失败:', error)
  }
}

const setStatus = (status) => {
  if (statusFilter.value === status) {
    return
  }
  statusFilter.value = status
  currentPage.value = 1
  fetchArticles()
}

// 草稿走独立草稿箱页（对齐B站草稿tab）
const goDrafts = () => {
  router.push('/platform/drafts')
}

// 监听筛选条件变化，重新获取数据
watch(statusFilter, () => {
  currentPage.value = 1
  fetchArticles()
})

// 监听分页变化
watch(currentPage, () => {
  fetchArticles()
})

// 监听当前激活菜单变化，加载稿件数据
watch(
  () => props.active,
  (newVal) => {
    if (newVal) {
      fetchArticles()
      fetchManuscriptStats()
    }
    syncArticleProcessPolling()
  },
  { immediate: true }
)

// 监听 mainTab 变化，同步轮询
watch(mainTab, () => {
  syncArticleProcessPolling()
})

// 获取稿件状态类型
const getArticleStatusType = (status) => {
  const statusTypeMap = {
    0: 'info',      // 待审核
    1: 'warning',   // 进行中
    3: 'success',   // 已发布
    4: 'danger',    // 已拒绝
    '-1': 'warning' // 已下架
  }
  return statusTypeMap[status] || 'info'
}

// 获取稿件状态文本
const getArticleStatusText = (status) => {
  const statusTextMap = {
    0: '待审核',
    1: '处理中',
    2: '待发布',
    3: '已通过',
    4: '未通过',
    '-1': '已下架'
  }
  return statusTextMap[status] || '未知'
}

// 状态 pill 徽章：近真页 article auto-check tag（成功绿色 / 进行中蓝 / 待审核灰 / 异常橘红）
const statusPillClass = (status) => {
  const pillMap = {
    0: 'pill-info',     // 待审核
    1: 'pill-loading',  // 处理中
    2: 'pill-info',     // 待发布
    3: 'pill-success',  // 已通过
    4: 'pill-danger',   // 未通过
    '-1': 'pill-warn'   // 已下架
  }
  return pillMap[status] || 'pill-info'
}

const getArticleVideos = (article) => {
  return Array.isArray(article?.videos)
    ? [...article.videos].sort((a, b) => (a.videoOrder ?? 0) - (b.videoOrder ?? 0))
    : []
}

const getVideoPartTitle = (video) => {
  const order = (video.videoOrder ?? 0) + 1
  return `P${order} ${video.title || '未命名视频'}`
}

const getVideoProcessProgress = (video) => {
  const value = Number(video?.processProgress)
  if (!Number.isFinite(value)) {
    return isVideoProcessDone(video) ? 100 : 0
  }
  return Math.min(100, Math.max(0, value))
}

const isVideoProcessDone = (video) => {
  return video?.processStatus === 5
}

const isVideoProcessFailed = (video) => {
  return [10, 20, 30, 40].includes(video?.processStatus) || !!video?.processError
}

const hasProcessingInfo = (article) => {
  return article.status === 0 || article.status === 1 || getArticleVideos(article).some(
    (v) => v.processStatus != null || v.processError
  )
}

const processChipClass = (video, articleStatus) => {
  if (isVideoProcessFailed(video)) return 'chip-danger'
  if (isVideoProcessDone(video)) return 'chip-done'
  if (articleStatus === 1 || [1, 2, 3, 4, 11, 21, 31, 41].includes(video?.processStatus)) return 'chip-working'
  return 'chip-idle'
}

const hasCardActions = (status) => {
  return [3, -1, 0, 1].includes(status)
}

const handleCardCommand = async (command, article) => {
  if (command === 'delete') {
    await deleteArticle(article.id)
  } else if (command === 'unpublish') {
    await unpublishArticle(article.id)
  } else if (command === 'publish') {
    await publishArticle(article.id)
  }
}

// 删除稿件
const deleteArticle = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个稿件吗？删除后无法恢复。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const response = await manuscriptApi.deleteManuscript(id)

    if (response.code === 200) {
      ElMessage.success('删除成功')
      fetchArticles()
      fetchManuscriptStats()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除稿件失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 下架稿件
const unpublishArticle = async (id) => {
  try {
    await ElMessageBox.confirm('确定要下架这个稿件吗？下架后观众将无法观看。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const response = await manuscriptApi.unpublishManuscript(id)

    if (response.code === 200) {
      ElMessage.success('下架成功')
      statusFilter.value = 'unpublished'
      fetchArticles()
      fetchManuscriptStats()
    } else {
      ElMessage.error(response.message || '下架失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('下架稿件失败:', error)
      ElMessage.error('下架失败')
    }
  }
}

// 上架稿件
const publishArticle = async (id) => {
  try {
    await ElMessageBox.confirm('确定要重新上架这个稿件吗？上架后观众将可以观看。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    })

    const response = await manuscriptApi.publishManuscript(id)

    if (response.code === 200) {
      ElMessage.success('上架成功')
      statusFilter.value = 'published'
      fetchArticles()
      fetchManuscriptStats()
    } else {
      ElMessage.error(response.message || '上架失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('上架稿件失败:', error)
      ElMessage.error('上架失败')
    }
  }
}

// 当前页变化
const handleCurrentChange = (current) => {
  currentPage.value = current
}

// 编辑保存后处理
const handleEditSaved = () => {
  statusFilter.value = 'processing'
  fetchArticles()
  fetchManuscriptStats()
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatDuration = (duration, seconds) => {
  let total = 0
  if (typeof duration === 'string' && /^\d+:\d+:\d+$/.test(duration)) {
    const [h, m, s] = duration.split(':').map(Number)
    total = h * 3600 + m * 60 + s
  } else {
    total = Number(seconds) || Number(duration) || 0
  }
  if (total <= 0) return ''
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) {
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

const formatCount = (n) => {
  const num = Number(n) || 0
  if (num >= 100000000) return (num / 100000000).toFixed(1) + '亿'
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  return String(num)
}

const getDefaultCoverText = () => '暂无封面'

onUnmounted(() => {
  stopArticleProcessPolling()
})
</script>

<style scoped>
.manuscript-manager {
  width: 100%;
  padding: 0 40px;
}

/* 主要选择栏 */
.content-tabs {
  margin-bottom: 15px;
}

/* 工具栏：对齐B站article-header（左右两段，8px 下留白 -> 列表） */
.video-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.toolbar-left {
  display: flex;
  align-items: flex-start;
  flex-direction: column;
  gap: 0;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.search-box {
  width: 230px;
  flex-shrink: 0;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 0 0 1px #e3e5e7 inset;
}

.search-input :deep(.el-input__wrapper.is-focus) {
  background: #fff;
  box-shadow: 0 0 0 1px #ff6699 inset;
}

/* 全部稿件 / 草稿（对齐B站.tabs：16px间距，选中不变粗） */
.main-tabs {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.status-tab {
  padding: 6px 0;
  font-size: 14px;
  color: #505050;
  cursor: pointer;
  transition: color .2s;
  white-space: nowrap;
}

.status-tab:hover {
  color: #ff6699;
}

.status-tab.active {
  color: #ff6699;
}

/* 状态细分（对齐B站secondTabs：32px大间距，弱化灰色，选中仅加粗） */
.second-tabs {
  display: flex;
  align-items: center;
  gap: 32px;
  flex-wrap: wrap;
}

.second-tab {
  font-size: 14px;
  color: rgba(80, 80, 80, .7);
  cursor: pointer;
  transition: color .2s;
  white-space: nowrap;
}

.second-tab:hover {
  color: #ff6699;
}

.second-tab.active {
  font-weight: 500;
}

.sort-select {
  width: 150px;
  flex-shrink: 0;
}

/* 卡片列表 */
.article-list {
  margin-bottom: 20px;
}

.manuscript-cards {
  display: flex;
  flex-direction: column;
}

/* 真页 article-card.v2：白底紧贴堆叠，无外边框，靠底部一条缝线分隔 */
.manuscript-card {
  display: flex;
  gap: 24px;
  padding: 24px 0;
  border-bottom: 1px solid #e7e7e7;
  background: #fff;
  transition: background .2s;
}

.manuscript-card:last-child {
  border-bottom: none;
}

.manuscript-card:hover {
  background: #fafbfc;
}

.card-cover {
  position: relative;
  width: 154px;
  height: 87px;
  flex-shrink: 0;
  border-radius: 4px;
  overflow: hidden;
  background: #f1f3f7;
  cursor: pointer;
  display: block;
}

/* 封面角标：真页 仅自己可见/时长 黑药丸，radius 0 4px 0 0 */
.cover-tag {
  position: absolute;
  left: 0;
  top: 0;
  padding: 2px 8px;
  font-size: 12px;
  color: #fff;
  background: rgba(0, 0, 0, .5);
  border-radius: 0 4px 0 0;
}

.cover-play {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, .25);
  opacity: 0;
  transition: opacity .2s;
}

.card-cover:hover .cover-play {
  opacity: 1;
}

.cover-play .el-icon {
  font-size: 40px;
  color: rgba(255, 255, 255, .95);
}

.cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #99a2aa;
  font-size: 13px;
  background: linear-gradient(135deg, #f4f6f8, #e9edf2);
}

.cover-duration {
  position: absolute;
  right: 0;
  bottom: 0;
  padding: 2px 8px;
  font-size: 12px;
  color: #fff;
  background: rgba(0, 0, 0, .5);
  border-radius: 4px 0 0 0;
}

.card-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.info-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #212121;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-title:hover {
  color: #ff6699;
}

.card-ops {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

/* 状态 pill 徽章（真页 auto-check tag：radius 12px 语义描边） */
.status-pill {
  display: inline-block;
  padding: 2px 10px;
  font-size: 12px;
  line-height: 18px;
  border-radius: 12px;
  white-space: nowrap;
}

.status-pill.pill-success {
  color: #00b96b;
  background: rgba(0, 185, 107, .1);
}

.status-pill.pill-loading {
  color: #00a1d6;
  background: rgba(0, 161, 214, .1);
}

.status-pill.pill-danger {
  color: #f56c6c;
  background: rgba(245, 108, 108, .1);
}

.status-pill.pill-warn {
  color: #f78a24;
  background: rgba(247, 138, 36, .1);
}

.status-pill.pill-info {
  color: #61666d;
  background: rgba(97, 102, 109, .1);
}

.info-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.meta-spacer {
  flex: 1;
}

.meta-date {
  font-size: 14px;
  color: #505050;
}

.meta-process {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.process-chip {
  padding: 1px 8px;
  font-size: 12px;
  border-radius: 10px;
  white-space: nowrap;
}

.process-chip.chip-working {
  color: #e6a23c;
  background: rgba(230, 162, 60, .12);
}

.process-chip.chip-done {
  color: #67c23a;
  background: rgba(103, 194, 58, .12);
}

.process-chip.chip-danger {
  color: #f56c6c;
  background: rgba(245, 108, 108, .12);
}

.process-chip.chip-idle {
  color: #909399;
  background: rgba(144, 147, 153, .12);
}

.process-percent {
  margin-left: 2px;
}

.meta-foot {
  margin-top: auto;
  padding-top: 14px;
  border-top: 1px solid #f0f1f4;
}

.info-stats {
  display: flex;
  align-items: center;
  gap: 24px;
}

.stat-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.stat-item .stat-num {
  font-size: 14px;
  color: #9499a0;
}

.stat-item em {
  font-style: normal;
  font-size: 12px;
  color: #999aa0;
}

/* 操作按钮：编辑 84x32 描边、more 32x32 描边（真页 bili-btn radius2） */
.card-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bili-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  font-size: 14px;
  color: #505050;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 2px;
  cursor: pointer;
  transition: color .2s, border-color .2s;
}

.bili-btn:hover {
  color: #ff6699;
  border-color: #ff6699;
}

.bili-more-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  font-size: 18px;
  color: #505050;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 2px;
  cursor: pointer;
  transition: color .2s, border-color .2s;
}

.bili-more-btn:hover {
  color: #ff6699;
  border-color: #ff6699;
}

/* 分页导航栏 */
.pagination {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 20px;
  padding: 15px 0;
  border-top: 1px solid #e0e0e0;
}
</style>
