<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getPendingManuscripts,
  getProcessingManuscripts,
  getAllManuscripts,
  getManuscriptDetail,
  approveManuscript,
  approveWithProcess,
  rejectManuscript,
  publishManuscript,
  unpublishManuscript,
  getManuscriptVideos,
  getManuscriptStatistics,
  retryManuscript,
  manualTranscode,
  manualExtractAudio,
  manualGenerateSubtitle,
  manualAiSummary,
  manualProcessAll,
  resetVideoStatus,
  getVideoSourceUrl
} from '../api/manuscript'

const tableData = ref([])
const loading = ref(false)
const statistics = ref({})

const keyword = ref('')
const statusFilter = ref('')

const videosDialogVisible = ref(false)
const videoPlayerVisible = ref(false)
const detailDialogVisible = ref(false)
const currentManuscript = ref({})
const currentDetail = ref(null)
const manuscriptVideos = ref([])
const currentVideo = ref(null)
const videoPlayerUrl = ref('')

const loadStatistics = async () => {
  try {
    const res = await getManuscriptStatistics()
    if (res.code === 200 || res.success) {
      statistics.value = res.data || {}
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const loadManuscripts = async () => {
  loading.value = true
  try {
    let res
    switch (statusFilter.value) {
      case '0':
        res = await getPendingManuscripts()
        break
      case '1':
        res = await getProcessingManuscripts()
        break
      default:
        res = await getAllManuscripts()
    }

    if (res.code === 200 || res.success) {
      let list = res.data || []
      if (keyword.value) {
        list = list.filter(item =>
          item.title?.includes(keyword.value) ||
          item.id?.toString().includes(keyword.value)
        )
      }
      if (statusFilter.value && statusFilter.value !== '') {
        list = list.filter(item => item.status === parseInt(statusFilter.value))
      }
      tableData.value = list
    } else {
      ElMessage.error(res.message || '获取稿件列表失败')
    }
  } catch (error) {
    ElMessage.error('获取稿件列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadManuscripts()
}

const handleReset = () => {
  keyword.value = ''
  statusFilter.value = ''
  loadManuscripts()
}

const handleApprove = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定审核通过该稿件吗？审核通过后将自动处理视频。',
      '审核通过',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    const res = await approveWithProcess(row.id, true)
    if (res.code === 200 || res.success) {
      ElMessage.success('审核通过，已开始处理视频')
      loadManuscripts()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '审核通过失败')
    }
  } catch (action) {
    if (action !== 'cancel') {
      console.error('审核通过异常:', action)
    }
  }
}

const handleReject = async (row) => {
  try {
    const { value: reason } = await ElMessageBox.prompt('请输入拒绝原因', '审核拒绝', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '请输入拒绝原因'
    })
    const res = await rejectManuscript(row.id, 1, reason)
    if (res.code === 200 || res.success) {
      ElMessage.success('已拒绝')
      loadManuscripts()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '审核拒绝失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('审核拒绝异常:', error)
    }
  }
}

const handlePublish = async (row) => {
  try {
    await ElMessageBox.confirm('确定要上架该稿件吗？', '上架稿件', { type: 'warning' })
    const res = await publishManuscript(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('上架成功')
      loadManuscripts()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '上架失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('上架稿件异常:', error)
    }
  }
}

const handleUnpublish = async (row) => {
  try {
    await ElMessageBox.confirm('确定要下架该稿件吗？', '下架稿件', { type: 'warning' })
    const res = await unpublishManuscript(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已下架')
      loadManuscripts()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '下架失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('下架稿件异常:', error)
    }
  }
}

const handleRetry = async (row) => {
  try {
    await ElMessageBox.confirm('确定要重试处理该稿件吗？', '重试处理', { type: 'warning' })
    const res = await retryManuscript(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始重试处理')
      loadManuscripts()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '重试失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重试处理异常:', error)
    }
  }
}

// ==================== 视频列表弹窗 ====================

const handleViewVideos = async (row) => {
  currentManuscript.value = row
  videosDialogVisible.value = true
  try {
    const res = await getManuscriptVideos(row.id)
    if (res.code === 200 || res.success) {
      manuscriptVideos.value = res.data || []
    } else {
      ElMessage.error(res.message || '获取视频列表失败')
    }
  } catch (error) {
    ElMessage.error('获取视频列表失败: ' + (error.message || '未知错误'))
  }
}

const handleViewDetail = async (row) => {
  currentManuscript.value = row
  detailDialogVisible.value = true
  currentDetail.value = null
  try {
    const res = await getManuscriptDetail(row.id)
    if (res.code === 200 || res.success) {
      currentDetail.value = res.data || null
    } else {
      ElMessage.error(res.message || '获取稿件详情失败')
    }
  } catch (error) {
    ElMessage.error('获取稿件详情失败: ' + (error.message || '未知错误'))
  }
}

// ==================== 视频播放器 ====================

const handlePlayVideo = async (video) => {
  try {
    const res = await getVideoSourceUrl(video.id)
    if (res.code === 200 || res.success) {
      currentVideo.value = video
      videoPlayerUrl.value = res.data?.sourceUrl || ''
      videoPlayerVisible.value = true
    } else {
      ElMessage.error('获取视频地址失败')
    }
  } catch (error) {
    console.error('播放视频失败:', error)
    ElMessage.error('播放失败')
  }
}

// ==================== 视频处理操作 ====================

const handleTranscode = async (video) => {
  try {
    const res = await manualTranscode(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始转码')
      refreshVideos()
    } else {
      ElMessage.error(res.message || '转码失败')
    }
  } catch (error) {
    ElMessage.error('转码异常: ' + (error.message || '未知错误'))
  }
}

const handleExtractAudio = async (video) => {
  try {
    const res = await manualExtractAudio(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始提取音频')
      refreshVideos()
    } else {
      ElMessage.error(res.message || '提取音频失败')
    }
  } catch (error) {
    ElMessage.error('提取音频异常: ' + (error.message || '未知错误'))
  }
}

const handleGenerateSubtitle = async (video) => {
  try {
    const res = await manualGenerateSubtitle(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始生成字幕')
      refreshVideos()
    } else {
      ElMessage.error(res.message || '生成字幕失败')
    }
  } catch (error) {
    ElMessage.error('生成字幕异常: ' + (error.message || '未知错误'))
  }
}

const handleAiSummary = async (video) => {
  try {
    const res = await manualAiSummary(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始AI总结')
      refreshVideos()
    } else {
      ElMessage.error(res.message || 'AI总结失败')
    }
  } catch (error) {
    ElMessage.error('AI总结异常: ' + (error.message || '未知错误'))
  }
}

const handleProcessAll = async (video) => {
  try {
    await ElMessageBox.confirm('确定要一键处理该视频吗？', '一键处理', { type: 'warning' })
    const res = await manualProcessAll(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始全流程处理')
      refreshVideos()
    } else {
      ElMessage.error(res.message || '处理失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('一键处理异常:', error)
    }
  }
}

const handleResetVideo = async (video) => {
  try {
    await ElMessageBox.confirm('确定要重置该视频状态吗？', '重置状态', { type: 'warning' })
    const res = await resetVideoStatus(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('视频状态已重置')
      refreshVideos()
    } else {
      ElMessage.error(res.message || '重置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重置异常:', error)
    }
  }
}

const refreshVideos = async () => {
  if (currentManuscript.value?.id) {
    try {
      const res = await getManuscriptVideos(currentManuscript.value.id)
      if (res.code === 200 || res.success) {
        manuscriptVideos.value = res.data || []
      }
    } catch (error) {
      console.error('刷新视频列表失败:', error)
    }
  }
}

// ==================== 状态显示 ====================

const getStatusText = (status) => {
  const statusMap = {
    0: '待审核',
    1: '处理中',
    2: '待上架',
    3: '已上架',
    4: '审核拒绝',
    5: '处理失败',
    [-1]: '已下架'
  }
  return statusMap[status] || '未知'
}

const getStatusType = (status) => {
  const typeMap = {
    0: 'warning',
    1: 'info',
    2: 'primary',
    3: 'success',
    4: 'danger',
    5: 'danger',
    [-1]: 'info'
  }
  return typeMap[status] || ''
}

const getProcessStatusText = (processStatus) => {
  const statusMap = {
    0: '待处理',
    1: '视频转码中',
    10: '转码失败',
    11: '转码成功',
    2: '音频提取中',
    20: '音频提取失败',
    21: '音频提取成功',
    3: '字幕生成中',
    30: '字幕生成失败',
    31: '字幕生成成功',
    4: 'AI总结中',
    40: 'AI总结失败',
    41: 'AI总结成功',
    5: '处理完成'
  }
  return statusMap[processStatus] || '未知(' + processStatus + ')'
}

const getProcessStatusType = (processStatus) => {
  // 5-全部完成, 11/21/31/41-各阶段成功 -> 绿色
  if (processStatus === 5 || processStatus === 11 || processStatus === 21 || processStatus === 31 || processStatus === 41) return 'success'
  // 10/20/30/40-各阶段失败 -> 红色
  if (processStatus === 10 || processStatus === 20 || processStatus === 30 || processStatus === 40) return 'danger'
  // 1/2/3/4-处理中 -> 蓝色
  if (processStatus === 1 || processStatus === 2 || processStatus === 3 || processStatus === 4) return 'primary'
  // 0-待处理 -> 灰色
  return 'info'
}

// 判断稿件的转码状态
const getTranscodeStatus = (row) => {
  // 如果没有视频列表数据，返回未知
  if (!row.videos || row.videos.length === 0) {
    return { text: '无视频', type: 'info' }
  }

  // 已完成：processStatus = 5 表示全部处理完成
  // 未完成：processStatus < 5 表示还有步骤未完成
  const allCompleted = row.videos.every(video => {
    const status = video.processStatus
    return status === 5
  })

  if (allCompleted) {
    return { text: '已完成', type: 'success' }
  } else {
    return { text: '未完成', type: 'warning' }
  }
}

// 格式化播放量
const formatViewCount = (count) => {
  if (!count) return '0'
  if (count >= 10000) {
    return (count / 10000).toFixed(1) + '万'
  }
  return count.toString()
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 格式化时长（秒 -> HH:MM:SS 或 MM:SS）
const formatDuration = (seconds) => {
  if (!seconds && seconds !== 0) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  if (hours > 0) {
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const getEditFieldLabel = (field) => {
  const map = {
    title: '标题',
    description: '简介',
    category: '分区',
    cover: '封面',
    tags: '标签',
    videos: '分P'
  }
  return map[field] || field
}

const formatSnapshotValue = (value) => {
  if (Array.isArray(value)) {
    return value.length > 0 ? value.join('，') : '无'
  }
  if (value === null || value === undefined || value === '') {
    return '无'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

onMounted(() => {
  loadManuscripts()
  loadStatistics()
})
</script>

<template>
  <div class="manuscripts-page">
    <h2 class="page-title">稿件管理</h2>

    <div class="statistics-bar">
      <div class="stat-item">
        <span class="stat-label">全部</span>
        <span class="stat-value">{{ statistics.total || 0 }}</span>
      </div>
      <div class="stat-item warning">
        <span class="stat-label">待审核</span>
        <span class="stat-value">{{ statistics.pending || 0 }}</span>
      </div>
      <div class="stat-item info">
        <span class="stat-label">处理中</span>
        <span class="stat-value">{{ statistics.processing || 0 }}</span>
      </div>
      <div class="stat-item primary">
        <span class="stat-label">待上架</span>
        <span class="stat-value">{{ statistics.ready || 0 }}</span>
      </div>
      <div class="stat-item success">
        <span class="stat-label">已上架</span>
        <span class="stat-value">{{ statistics.published || 0 }}</span>
      </div>
      <div class="stat-item danger">
        <span class="stat-label">处理失败</span>
        <span class="stat-value">{{ statistics.failed || 0 }}</span>
      </div>
    </div>

    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索稿件标题、ID"
        clearable
        style="width: 250px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select
        v-model="statusFilter"
        placeholder="状态筛选"
        clearable
        style="width: 150px"
        @change="handleSearch"
      >
        <el-option label="全部" value="" />
        <el-option label="待审核" value="0" />
        <el-option label="处理中" value="1" />
        <el-option label="待上架" value="2" />
        <el-option label="已上架" value="3" />
        <el-option label="审核拒绝" value="4" />
        <el-option label="处理失败" value="5" />
        <el-option label="已下架" value="-1" />
      </el-select>
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
      <el-button type="primary" @click="loadManuscripts">刷新</el-button>
    </div>

    <!-- 稿件列表 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="250" show-overflow-tooltip />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <div>
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
            <div v-if="row.status === 1 && row.processProgress" style="font-size: 12px; color: #999; margin-top: 4px;">
              {{ row.processProgress }}%
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="处理状态" width="120">
        <template #default="{ row }">
          <el-tag :type="getTranscodeStatus(row).type" size="small">
            {{ getTranscodeStatus(row).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="播放量" width="100">
        <template #default="{ row }">
          <span>{{ formatViewCount(row.viewCount) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="总时长" width="100">
        <template #default="{ row }">
          <span>{{ formatDuration(row.durationSeconds) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="上传时间" width="160">
        <template #default="{ row }">
          <span>{{ formatTime(row.uploadTime) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="300">
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            size="small"
            @click="handleViewVideos(row)"
          >
            视频列表
          </el-button>
          <el-button
            link
            type="info"
            size="small"
            @click="handleViewDetail(row)"
          >
            审核对比
          </el-button>

          <el-dropdown v-if="row.status === 0" trigger="click">
            <el-button link type="primary" size="small">
              审核 <el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleApprove(row)">通过</el-dropdown-item>
                <el-dropdown-item @click="handleReject(row)" divided>拒绝</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-button
            v-if="row.status === 2"
            link
            type="success"
            size="small"
            @click="handlePublish(row)"
          >
            上架
          </el-button>

          <el-button
            v-if="row.status === 3"
            link
            type="warning"
            size="small"
            @click="handleUnpublish(row)"
          >
            下架
          </el-button>

          <el-button
            v-if="row.status === 5"
            link
            type="danger"
            size="small"
            @click="handleRetry(row)"
          >
            重试
          </el-button>

          <el-button
            v-if="row.status === -1"
            link
            type="success"
            size="small"
            @click="handlePublish(row)"
          >
            上架
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 审核对比弹窗 -->
    <el-dialog v-model="detailDialogVisible" title="稿件审核对比" width="960px">
      <div v-if="currentDetail" class="review-detail">
        <el-descriptions :column="3" border size="small">
          <el-descriptions-item label="稿件ID">{{ currentDetail.manuscript?.id || currentManuscript.id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ currentDetail.manuscript?.title || currentManuscript.title }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentDetail.manuscript?.status)">
              {{ getStatusText(currentDetail.manuscript?.status) }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <template v-if="currentDetail.editVersion">
          <div class="edit-version-header">
            <div>
              <span class="section-title">本次编辑记录</span>
              <el-tag size="small" type="warning">{{ currentDetail.editVersion.status }}</el-tag>
            </div>
            <span class="edit-time">{{ formatTime(currentDetail.editVersion.createdAt) }}</span>
          </div>

          <div class="changed-fields">
            <span class="changed-label">变更字段</span>
            <el-tag
              v-for="field in currentDetail.editVersion.changedFields"
              :key="field"
              size="small"
              effect="plain"
            >
              {{ getEditFieldLabel(field) }}
            </el-tag>
            <span v-if="!currentDetail.editVersion.changedFields?.length" class="muted-text">无字段变化</span>
          </div>

          <el-table
            :data="[
              { field: '标题', before: currentDetail.editVersion.beforeSnapshot?.title, after: currentDetail.editVersion.afterSnapshot?.title },
              { field: '简介', before: currentDetail.editVersion.beforeSnapshot?.description, after: currentDetail.editVersion.afterSnapshot?.description },
              { field: '分区', before: currentDetail.editVersion.beforeSnapshot?.categoryName || currentDetail.editVersion.beforeSnapshot?.categoryId, after: currentDetail.editVersion.afterSnapshot?.categoryName || currentDetail.editVersion.afterSnapshot?.categoryId },
              { field: '标签', before: currentDetail.editVersion.beforeSnapshot?.tags, after: currentDetail.editVersion.afterSnapshot?.tags },
              { field: '封面', before: currentDetail.editVersion.beforeSnapshot?.coverUrl ? '已设置' : '无', after: currentDetail.editVersion.afterSnapshot?.coverUrl ? '已设置' : '无' }
            ]"
            size="small"
            border
            class="diff-table"
          >
            <el-table-column prop="field" label="字段" width="90" />
            <el-table-column label="编辑前">
              <template #default="{ row }">{{ formatSnapshotValue(row.before) }}</template>
            </el-table-column>
            <el-table-column label="编辑后">
              <template #default="{ row }">{{ formatSnapshotValue(row.after) }}</template>
            </el-table-column>
          </el-table>

          <div class="section-title video-diff-title">分P对比</div>
          <el-table
            :data="currentDetail.editVersion.afterSnapshot?.videos || []"
            size="small"
            border
          >
            <el-table-column label="分P" width="70">
              <template #default="{ row, $index }">P{{ (row.videoOrder ?? $index) + 1 }}</template>
            </el-table-column>
            <el-table-column label="视频ID" prop="id" width="90" />
            <el-table-column label="标题">
              <template #default="{ row }">{{ row.title || '-' }}</template>
            </el-table-column>
            <el-table-column label="源视频" width="120">
              <template #default="{ row }">
                <el-tag
                  size="small"
                  :type="(currentDetail.editVersion.beforeSnapshot?.videos || []).find(v => v.id === row.id)?.sourceVideoUrl === row.sourceVideoUrl ? 'info' : 'warning'"
                >
                  {{ (currentDetail.editVersion.beforeSnapshot?.videos || []).find(v => v.id === row.id)?.sourceVideoUrl === row.sourceVideoUrl ? '未替换' : '已替换' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </template>

        <el-empty v-else description="暂无编辑版本记录" :image-size="90" />
      </div>
      <el-skeleton v-else :rows="6" animated />
    </el-dialog>

    <!-- 视频列表弹窗 -->
    <el-dialog v-model="videosDialogVisible" title="视频列表" width="900px">
      <div v-if="currentManuscript">
        <h4>稿件: {{ currentManuscript.title }} (ID: {{ currentManuscript.id }})</h4>
        <el-table :data="manuscriptVideos" style="width: 100%; margin-top: 20px;">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="title" label="视频标题" min-width="180" show-overflow-tooltip />
          <el-table-column prop="videoOrder" label="分P" width="60">
            <template #default="{ row }">
              <span>P{{ row.videoOrder + 1 }}</span>
            </template>
          </el-table-column>
          <el-table-column label="时长" width="80">
            <template #default="{ row }">
              <span>{{ formatDuration(row.durationSeconds) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="处理状态" width="130">
            <template #default="{ row }">
              <el-tag :type="getProcessStatusType(row.processStatus)" size="small">
                {{ getProcessStatusText(row.processStatus) }}
              </el-tag>
              <div v-if="row.processError" style="font-size: 11px; color: #f56c6c; margin-top: 4px; max-width: 120px; overflow: hidden; text-overflow: ellipsis;">
                {{ row.processError }}
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <!-- 播放按钮 -->
              <el-button
                link
                type="primary"
                size="small"
                @click="handlePlayVideo(row)"
              >
                <el-icon><VideoPlay /></el-icon>
                播放
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <!-- 视频播放器弹窗 -->
    <el-dialog v-model="videoPlayerVisible" title="视频预览" width="800px" :close-on-click-modal="false">
      <div v-if="currentVideo" class="video-player-container">
        <h4>{{ currentVideo.title }}</h4>
        <p class="video-info">P{{ currentVideo.videoOrder }} | 时长: {{ currentVideo.duration }}</p>
        <div class="player-wrapper">
          <video
            v-if="videoPlayerUrl"
            :src="videoPlayerUrl"
            controls
            style="width: 100%; max-height: 450px;"
          />
          <div v-else class="no-video">
            <el-icon :size="48" color="#999"><VideoPlay /></el-icon>
            <p>暂无视频源</p>
          </div>
        </div>
        <div class="video-meta">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="视频ID">{{ currentVideo.id }}</el-descriptions-item>
            <el-descriptions-item label="处理状态">
              <el-tag :type="getProcessStatusType(currentVideo.processStatus)" size="small">
                {{ getProcessStatusText(currentVideo.processStatus) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="字幕">{{ currentVideo.hasSubtitle ? '有' : '无' }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.manuscripts-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
}

.statistics-bar {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-item {
  flex: 1;
  text-align: center;
  padding: 10px;
  background: #fff;
  border-radius: 6px;
}

.stat-item.warning .stat-value { color: #e6a23c; }
.stat-item.info .stat-value { color: #409eff; }
.stat-item.primary .stat-value { color: #409eff; }
.stat-item.success .stat-value { color: #67c23a; }
.stat-item.danger .stat-value { color: #f56c6c; }

.stat-label {
  display: block;
  font-size: 14px;
  color: #666;
  margin-bottom: 5px;
}

.stat-value {
  display: block;
  font-size: 24px;
  font-weight: bold;
  color: #333;
}

.search-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.review-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.edit-version-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding-top: 4px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-right: 10px;
}

.edit-time,
.muted-text {
  color: #909399;
  font-size: 13px;
}

.changed-fields {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.changed-label {
  color: #606266;
  font-size: 13px;
}

.diff-table {
  margin-top: 2px;
}

.video-diff-title {
  margin-top: 4px;
}

/* 视频播放器样式 */
.video-player-container {
  padding: 10px;
}

.video-player-container h4 {
  margin: 0 0 8px;
  font-size: 16px;
}

.video-info {
  color: #666;
  font-size: 14px;
  margin: 0 0 16px;
}

.player-wrapper {
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.no-video {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  color: #999;
}

.video-meta {
  margin-top: 16px;
}
</style>
