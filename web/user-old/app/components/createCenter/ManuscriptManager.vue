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
      <!-- 视频管理次级选择栏 -->
      <div class="video-filters">
        <!-- 第一行：全部稿件 -->
        <div class="filter-row">
          <el-radio-group v-model="statusFilter" size="small">
            <el-radio-button value="published">已通过 ({{ approvedCount }})</el-radio-button>
            <el-radio-button value="processing">等待审核 / 处理中</el-radio-button>
            <el-radio-button value="rejected">未通过 ({{ rejectedCount }})</el-radio-button>
            <el-radio-button value="unpublished">已下架 ({{ unpublishedCount }})</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <!-- 视频稿件列表 -->
      <div class="article-list" v-loading="articlesLoading">
        <el-table :data="articles" stripe style="width: 100%">
          <el-table-column prop="id" label="稿件ID" width="120"></el-table-column>
          <el-table-column prop="title" label="标题" min-width="300"></el-table-column>
          <el-table-column prop="status" label="状态" width="180">
            <template #default="scope">
              <el-tag
                :type="getArticleStatusType(scope.row.status)"
                size="small"
              >
                {{ getArticleStatusText(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="处理进度" min-width="280">
            <template #default="scope">
              <div class="manuscript-process">
                <div
                  v-for="video in getArticleVideos(scope.row)"
                  :key="video.id || video.videoOrder"
                  class="process-video-row"
                >
                  <div class="process-video-header">
                    <span class="process-video-title">{{ getVideoPartTitle(video) }}</span>
                    <el-tag :type="getVideoProcessTagType(video)" size="small">
                      {{ getVideoProcessText(video, scope.row.status) }}
                    </el-tag>
                  </div>
                  <el-progress
                    v-if="shouldShowVideoProgress(video, scope.row.status)"
                    :percentage="getVideoProcessProgress(video)"
                    :status="getVideoProgressStatus(video)"
                    :stroke-width="6"
                    :show-text="false"
                  />
                  <div v-if="video.processError" class="process-error">
                    {{ video.processError }}
                  </div>
                </div>
                <span v-if="getArticleVideos(scope.row).length === 0" class="process-empty">
                  暂无视频处理信息
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="viewCount" label="播放量" width="100"></el-table-column>
          <el-table-column prop="commentCount" label="评论数" width="100"></el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.uploadTime) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="scope">
              <el-button
                v-if="scope.row.status !== 1"
                type="primary"
                size="small"
                @click="openEditDialog(scope.row.id)"
              >
                编辑
              </el-button>
              <!-- 待审核/处理中状态 -->
              <template v-if="scope.row.status === 0 || scope.row.status === 1">
                <el-button
                  type="info"
                  size="small"
                  disabled
                >
                  {{ scope.row.status === 0 ? '审核中...' : '处理中...' }}
                </el-button>
              </template>
              <!-- 已发布状态 -->
              <el-button v-if="scope.row.status === 3" type="warning" size="small" @click="unpublishArticle(scope.row.id)">下架</el-button>
              <!-- 已下架状态 -->
              <el-button v-if="scope.row.status === -1" type="success" size="small" @click="publishArticle(scope.row.id)">上架</el-button>
              <!-- 删除按钮 -->
              <el-button type="danger" size="small" @click="deleteArticle(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页导航栏 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="totalArticles"
          @size-change="handleSizeChange"
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

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
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

// 主要选择栏
const mainTab = ref('video')

// 状态筛选
const statusFilter = ref('published') // published: 已通过, processing: 进行中, rejected: 未通过
const validArticleStatusFilters = new Set(['published', 'processing', 'rejected', 'unpublished'])

watch(
  () => [route.path, route.query.status],
  () => {
    if (route.path !== '/create-center/content-articles') {
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
    if (statusFilter.value) {
      params.status = statusFilter.value
    }

    console.log('获取稿件参数:', params)
    const response = await manuscriptApi.getMyManuscripts(params)
    console.log('获取稿件响应:', response)

    if (response.code === 200) {
      articles.value = response.data.list || []
      totalArticles.value = response.data.total || 0
      console.log('稿件列表:', articles.value)
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

// 监听筛选条件变化，重新获取数据
watch(statusFilter, () => {
  currentPage.value = 1
  fetchArticles()
})

// 监听分页变化
watch([currentPage, pageSize], () => {
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

const getVideoProcessTagType = (video) => {
  if (isVideoProcessFailed(video)) return 'danger'
  if (isVideoProcessDone(video)) return 'success'
  if ([1, 2, 3, 4, 11, 21, 31, 41].includes(video?.processStatus)) return 'warning'
  return 'info'
}

const getVideoProcessText = (video, articleStatus) => {
  const statusMap = {
    0: '待处理',
    1: '转码中',
    2: '抽取音频中',
    3: '字幕生成中',
    4: '摘要生成中',
    5: '处理完成',
    10: '转码失败',
    11: '转码完成',
    20: '音频抽取失败',
    21: '音频抽取完成',
    30: '字幕生成失败',
    31: '字幕生成完成',
    40: '摘要生成失败',
    41: '摘要生成完成'
  }
  if (video?.processStatus != null) {
    return statusMap[video.processStatus] || `处理状态 ${video.processStatus}`
  }
  if (articleStatus === 0) return '等待审核'
  if (articleStatus === 1) return '等待处理'
  return '无处理任务'
}

const getVideoProgressStatus = (video) => {
  if (isVideoProcessFailed(video)) return 'exception'
  if (isVideoProcessDone(video)) return 'success'
  return undefined
}

const shouldShowVideoProgress = (video, articleStatus) => {
  return articleStatus === 1 || video?.processProgress != null || isVideoProcessDone(video) || isVideoProcessFailed(video)
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
    
    console.log('下架稿件ID:', id)
    const response = await manuscriptApi.unpublishManuscript(id)
    console.log('下架响应:', response)
    
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

// 分页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
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

onUnmounted(() => {
  stopArticleProcessPolling()
})
</script>

<style scoped>
.manuscript-manager {
  width: 100%;
}

/* 主要选择栏 */
.content-tabs {
  margin-bottom: 15px;
}

/* 视频管理选择栏 */
.video-filters {
  margin-bottom: 20px;
}

/* 过滤行样式 */
.filter-row {
  margin-bottom: 10px;
}

/* 过滤行之间的间距 */
.filter-row:not(:last-child) {
  margin-bottom: 15px;
}

/* 单个过滤组样式 */
.filter-row .el-radio-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 隐藏次级选择栏和状态选择栏 */
.sub-tabs,
.status-tabs {
  display: none;
}

/* 视频稿件列表 */
.article-list {
  margin-bottom: 20px;
}

/* 表格样式调整 */
.article-list .el-table {
  border-radius: 6px;
  overflow: hidden;
}

.manuscript-process {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 4px 0;
}

.process-video-row {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.process-video-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.process-video-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #606266;
  font-size: 13px;
}

.process-error {
  color: #f56c6c;
  font-size: 12px;
  line-height: 1.4;
}

.process-empty {
  color: #909399;
  font-size: 13px;
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