<template>
  <div class="video-process-page">
    <div class="page-header">
      <h2>任务看板</h2>
      <div class="header-actions">
        <el-tag :type="wsConnected ? 'success' : 'danger'" size="small">
          {{ wsConnected ? '实时连接' : '已断开' }}
        </el-tag>
        <el-switch v-model="autoRefresh" active-text="自动刷新" @change="toggleAutoRefresh" />
        <el-button type="primary" :icon="Refresh" @click="loadAllData" :loading="loading">刷新</el-button>
      </div>
    </div>
    <p class="page-desc">视频处理任务串行执行，实时展示当前任务、队列与失败状态</p>

    <!-- 当前任务卡片 -->
    <el-card class="current-task-card" v-if="currentTask.processing">
      <template #header>
        <div class="card-header">
          <span><el-icon class="processing-icon"><Loading /></el-icon> 当前处理任务</span>
        </div>
      </template>
      <div class="current-task-content">
        <div class="task-info">
          <span class="task-label">视频ID:</span>
          <span class="task-value">{{ currentTask.videoId }}</span>
          <span class="task-label" style="margin-left: 20px;">标题:</span>
          <span class="task-value">{{ currentTask.videoTitle }}</span>
        </div>
        <div class="task-progress">
          <span class="stage-text">{{ currentTask.stageText }}</span>
          <el-progress 
            :percentage="currentTask.progress" 
            :status="currentTask.progress === 100 ? 'success' : ''"
            :stroke-width="20"
            :text-inside="true"
          />
        </div>
        <div class="task-meta">
          <span>状态: {{ currentTask.statusText }}</span>
        </div>
      </div>
    </el-card>

    <!-- 空闲状态卡片 -->
    <el-card class="idle-card" v-else>
      <div class="idle-content">
        <el-icon :size="40" color="#909399"><CircleCheck /></el-icon>
        <span class="idle-text">当前无处理任务</span>
        <div class="queue-detail" v-if="queueInfo.queueSize > 0">
          <el-divider>队列中任务</el-divider>
          <el-row :gutter="10" justify="center">
            <el-col :span="6" v-if="queueInfo.waitingTranscode > 0">
              <el-statistic title="等待转码" :value="queueInfo.waitingTranscode" />
            </el-col>
            <el-col :span="6" v-if="queueInfo.waitingAudio > 0">
              <el-statistic title="等待音频" :value="queueInfo.waitingAudio" />
            </el-col>
            <el-col :span="6" v-if="queueInfo.waitingSubtitle > 0">
              <el-statistic title="等待字幕" :value="queueInfo.waitingSubtitle" />
            </el-col>
            <el-col :span="6" v-if="queueInfo.waitingAi > 0">
              <el-statistic title="等待AI" :value="queueInfo.waitingAi" />
            </el-col>
          </el-row>
        </div>
      </div>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="statistics-row">
      <el-col :span="4">
        <el-card class="stat-card" :class="{ highlight: statistics.pending > 0 }">
          <div class="stat-value">{{ statistics.pending || 0 }}</div>
          <div class="stat-label">待处理</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" :class="{ highlight: statistics.transcoding > 0 }">
          <div class="stat-value processing">{{ statistics.transcoding || 0 }}</div>
          <div class="stat-label">转码阶段</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" :class="{ highlight: statistics.audioExtracting > 0 }">
          <div class="stat-value processing">{{ statistics.audioExtracting || 0 }}</div>
          <div class="stat-label">音频阶段</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" :class="{ highlight: statistics.subtitleGenerating > 0 }">
          <div class="stat-value processing">{{ statistics.subtitleGenerating || 0 }}</div>
          <div class="stat-label">字幕阶段</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" :class="{ highlight: statistics.aiSummarizing > 0 }">
          <div class="stat-value processing">{{ statistics.aiSummarizing || 0 }}</div>
          <div class="stat-label">AI阶段</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value success">{{ statistics.completed || 0 }}</div>
          <div class="stat-label">全部完成</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="处理状态">
          <el-select v-model="filterForm.processStatus" placeholder="全部状态" clearable @change="handleFilterChange">
            <el-option label="待处理" :value="0" />
            <el-option label="视频转码中" :value="1" />
            <el-option label="音频提取中" :value="2" />
            <el-option label="字幕生成中" :value="3" />
            <el-option label="AI总结中" :value="4" />
            <el-option label="处理完成" :value="5" />
            <el-option label="转码失败" :value="10" />
            <el-option label="音频提取失败" :value="20" />
            <el-option label="字幕生成失败" :value="30" />
            <el-option label="AI总结失败" :value="40" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filterForm.keyword" placeholder="搜索视频标题" clearable @keyup.enter="handleFilterChange" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleFilterChange">搜索</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 视频列表 -->
    <el-card class="table-card">
      <el-table :data="paginatedData" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="视频ID" width="80" />
        <el-table-column prop="manuscriptId" label="稿件ID" width="80" />
        <el-table-column prop="title" label="视频标题" min-width="250" show-overflow-tooltip />
        <el-table-column label="处理状态" width="140">
          <template #default="{ row }">
            <el-tag :type="getProcessStatusType(row.processStatus)" size="small">
              {{ getProcessStatusText(row.processStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="流水线进度" width="240">
          <template #default="{ row }">
            <div class="pipeline-progress">
              <span title="转码" style="cursor: pointer;">
                <div
                  class="pipeline-step"
                  :class="getStepClass(row.processStatus, 1)"
                  @click="handleStepClick(row, 1)"
                >
                  <el-icon><VideoCamera /></el-icon>
                </div>
              </span>
              <div class="pipeline-line" :class="{ active: isLineActive(row.processStatus, 1) }"></div>
              <span title="音频" style="cursor: pointer;">
                <div
                  class="pipeline-step"
                  :class="getStepClass(row.processStatus, 2)"
                  @click="handleStepClick(row, 2)"
                >
                  <el-icon><Microphone /></el-icon>
                </div>
              </span>
              <div class="pipeline-line" :class="{ active: isLineActive(row.processStatus, 2) }"></div>
              <span title="字幕" style="cursor: pointer;">
                <div
                  class="pipeline-step"
                  :class="getStepClass(row.processStatus, 3)"
                  @click="handleStepClick(row, 3)"
                >
                  <el-icon><ChatDotRound /></el-icon>
                </div>
              </span>
              <div class="pipeline-line" :class="{ active: isLineActive(row.processStatus, 3) }"></div>
              <span title="AI" style="cursor: pointer;">
                <div
                  class="pipeline-step"
                  :class="getStepClass(row.processStatus, 4)"
                  @click="handleStepClick(row, 4)"
                >
                  <el-icon><Cpu /></el-icon>
                </div>
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="danger"
              size="small"
              @click="handleReset(row)"
            >
              <el-icon><RefreshLeft /></el-icon>
              重置
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Loading, CircleCheck, VideoCamera, Microphone, ChatDotRound, Cpu, RefreshLeft } from '@element-plus/icons-vue'
import {
  getAllManuscripts,
  getManuscriptVideos,
  manualTranscode,
  manualExtractAudio,
  manualGenerateSubtitle,
  manualAiSummary,
  resetVideoStatus
} from '../api/manuscript'
import { getCurrentTask, getQueueInfo, getStatistics } from '../api/videoProcess'

const loading = ref(false)
const tableData = ref([])
const statistics = ref({})
const currentTask = ref({ processing: false })
const queueInfo = ref({ queueSize: 0 })
const autoRefresh = ref(true)
const refreshInterval = ref(null)

const sseSource = ref(null)
const wsConnected = ref(false)

const filterForm = reactive({
  processStatus: '',
  keyword: ''
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

const paginatedData = computed(() => {
  const start = (pagination.page - 1) * pagination.size
  const end = start + pagination.size
  return tableData.value.slice(start, end)
})

const connectSSE = () => {
  const protocol = window.location.protocol === 'https:' ? 'https:' : 'http:'
  const eventSource = new EventSource(`${protocol}//${window.location.host}/api/ai/admin/process/stream`)

  eventSource.onopen = () => {
    wsConnected.value = true
    loadAllData()
  }

  eventSource.addEventListener('snapshot', (event) => {
    const data = JSON.parse(event.data)
    if (data.current) {
      currentTask.value = data.current
    }
    if (data.statistics) {
      statistics.value = data.statistics
    }
    loadVideos()
    loadQueueInfo()
  })

  eventSource.addEventListener('progress', (event) => {
    handleSseMessage(JSON.parse(event.data))
  })

  eventSource.addEventListener('complete', (event) => {
    handleSseMessage(JSON.parse(event.data))
    ElMessage.success('视频处理完成')
    loadVideos()
    loadStatistics()
    loadQueueInfo()
  })

  eventSource.addEventListener('error', (event) => {
    if (event.data) {
      handleSseMessage(JSON.parse(event.data))
    }
  })

  eventSource.onerror = () => {
    wsConnected.value = false
  }

  sseSource.value = eventSource
}

const handleSseMessage = (data) => {
  if (!data) return

  if (data.videoId) {
    const index = tableData.value.findIndex(v => v.id === data.videoId)
    if (index !== -1) {
      const oldData = tableData.value[index]
      tableData.value.splice(index, 1, {
        ...oldData,
        processStatus: data.status ?? oldData.processStatus,
        processProgress: data.progress ?? oldData.processProgress,
        processStage: data.stage ?? oldData.processStage,
        processError: data.error ?? oldData.processError
      })
    }
  }

  if (data.type === 'complete') {
    currentTask.value = { processing: false }
  } else if (data.type === 'error') {
    currentTask.value = { processing: false, error: data.error }
    ElMessage.error(`视频处理失败: ${data.error || '未知错误'}`)
  } else {
    currentTask.value = {
      processing: true,
      videoId: data.videoId,
      manuscriptId: data.manuscriptId,
      videoTitle: data.videoTitle || data.title,
      stage: data.stage,
      stageText: data.stageText || data.statusText,
      progress: data.progress || 0,
      status: data.status,
      statusText: data.statusText || getProcessStatusText(data.status)
    }
  }

  loadStatistics()
  loadQueueInfo()
}

onMounted(() => {
  connectSSE()
  loadAllData()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
  if (sseSource.value) {
    sseSource.value.close()
    sseSource.value = null
  }
})

const loadAllData = async () => {
  loading.value = true
  try {
    await Promise.all([
      loadCurrentTask(),
      loadQueueInfo(),
      loadStatistics(),
      loadVideos()
    ])
  } finally {
    loading.value = false
  }
}

const loadCurrentTask = async () => {
  try {
    const res = await getCurrentTask()
    if (res.code === 200 || res.success) {
      currentTask.value = res.data || { processing: false }
    }
  } catch (error) {
    console.error('加载当前任务失败:', error)
  }
}

const loadQueueInfo = async () => {
  try {
    const res = await getQueueInfo()
    if (res.code === 200 || res.success) {
      queueInfo.value = res.data || { queueSize: 0 }
    }
  } catch (error) {
    console.error('加载队列信息失败:', error)
  }
}

const loadStatistics = async () => {
  try {
    const res = await getStatistics()
    if (res.code === 200 || res.success) {
      statistics.value = res.data || {}
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

const loadVideos = async () => {
  try {
    const res = await getAllManuscripts()
    if (res.code === 200 || res.success) {
      const manuscripts = res.data || []
      let allVideos = []
      
      for (const manuscript of manuscripts) {
        try {
          const videoRes = await getManuscriptVideos(manuscript.id)
          if (videoRes.code === 200 || videoRes.success) {
            const videos = videoRes.data || []
            videos.forEach(video => {
              allVideos.push({
                ...video,
                manuscriptId: manuscript.id,
                manuscriptTitle: manuscript.title,
                manuscriptCoverUrl: manuscript.coverUrl
              })
            })
          }
        } catch (error) {
          console.error(`加载稿件 ${manuscript.id} 的视频失败:`, error)
        }
      }
      
      if (filterForm.processStatus !== '') {
        allVideos = allVideos.filter(v => v.processStatus === filterForm.processStatus)
      }
      
      if (filterForm.keyword) {
        const keyword = filterForm.keyword.toLowerCase()
        allVideos = allVideos.filter(v => 
          v.title?.toLowerCase().includes(keyword) ||
          v.manuscriptTitle?.toLowerCase().includes(keyword)
        )
      }
      
      tableData.value = allVideos
      pagination.total = allVideos.length
    }
  } catch (error) {
    console.error('加载视频列表失败:', error)
    ElMessage.error('加载失败')
  }
}

const startAutoRefresh = () => {
  if (refreshInterval.value) return
  refreshInterval.value = setInterval(() => {
    loadCurrentTask()
    loadQueueInfo()
    loadStatistics()
  }, 10000)
}

const stopAutoRefresh = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
}

const toggleAutoRefresh = (val) => {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

const handleFilterChange = () => {
  pagination.page = 1
  loadVideos()
}

const resetFilter = () => {
  filterForm.processStatus = ''
  filterForm.keyword = ''
  handleFilterChange()
}

const handleSizeChange = () => {
  pagination.page = 1
}

const handlePageChange = () => {}

const isStepDone = (status, stepType) => {
  switch (stepType) {
    case 'transcode': return [11, 2, 20, 21, 3, 30, 31, 4, 40, 41, 5].includes(status)
    case 'audio': return [21, 3, 30, 31, 4, 40, 41, 5].includes(status)
    case 'subtitle': return [31, 4, 40, 41, 5].includes(status)
    case 'ai': return [41, 5].includes(status)
    default: return false
  }
}

const isLineActive = (status, step) => {
  if (step === 1) return [2, 20, 21, 3, 30, 31, 4, 40, 41, 5].includes(status)
  if (step === 2) return [3, 30, 31, 4, 40, 41, 5].includes(status)
  if (step === 3) return [4, 40, 41, 5].includes(status)
  return false
}

const canTranscode = (status) => {
  return status === 0 || status === 10 || status >= 11
}

const canExtractAudio = (status) => {
  return status === 11 || status === 20 || status >= 21
}

const canGenerateSubtitle = (status) => {
  return status === 21 || status === 30 || status >= 31
}

const canAiSummary = (status) => {
  return status === 31 || status === 40 || status >= 41 || status === 5
}

const isStepError = (status, step) => {
  const errorMap = { 1: 10, 2: 20, 3: 30, 4: 40 }
  return status === errorMap[step]
}

const getStepClass = (processStatus, step) => {
  const stepRanges = {
    1: { success: [11, 2, 20, 21, 3, 30, 31, 4, 40, 41, 5], fail: 10, processing: 1 },
    2: { success: [21, 3, 30, 31, 4, 40, 41, 5], fail: 20, processing: 2 },
    3: { success: [31, 4, 40, 41, 5], fail: 30, processing: 3 },
    4: { success: [41, 5], fail: 40, processing: 4 }
  }

  const range = stepRanges[step]

  if (processStatus === range.processing) return 'current'
  if (processStatus === range.fail) return 'error'
  if (range.success.includes(processStatus)) return 'completed'

  return 'pending'
}

const handleStepClick = (video, step) => {
  const status = video.processStatus
  const statusToStep = { 0: 1, 1: 1, 10: 1, 11: 2, 2: 2, 20: 2, 21: 3, 3: 3, 30: 3, 31: 4, 4: 4, 40: 4, 41: 4, 5: 4 }
  const expectedStep = statusToStep[status]

  if (step > expectedStep && !isStepError(status, step)) {
    return
  }

  const stepNames = { 1: '转码', 2: '音频提取', 3: '字幕生成', 4: 'AI总结' }
  const stepName = stepNames[step]

  let confirmMsg = `确定要执行 ${stepName} 吗？`
  if (isStepDone(status, ['transcode', 'audio', 'subtitle', 'ai'][step - 1])) {
    confirmMsg = `该视频已经执行过${stepName}，是否确认重新执行？`
  } else if (isStepError(status, step)) {
    confirmMsg = `${stepName}失败，是否确认重试？`
  }

  ElMessageBox.confirm(confirmMsg, '确认操作', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: isStepError(status, step) ? 'error' : 'warning'
  }).then(async () => {
    switch (step) {
      case 1: await handleTranscode(video); break
      case 2: await handleExtractAudio(video); break
      case 3: await handleGenerateSubtitle(video); break
      case 4: await handleAiSummary(video); break
    }
  }).catch(() => {})
}

const showConfirmDialog = async (stepName) => {
  try {
    await ElMessageBox.confirm(
      `该视频已经执行过${stepName}，是否确认重新执行？`,
      '确认重新执行',
      { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' }
    )
    return true
  } catch {
    return false
  }
}

const updateVideoStatus = (videoId, newStatus) => {
  const index = tableData.value.findIndex(v => v.id === videoId)
  if (index !== -1) {
    const oldData = tableData.value[index]
    const newData = { ...oldData, processStatus: newStatus }
    tableData.value.splice(index, 1, newData)
    
    const currentId = currentTask.value.videoId
    if (currentId === videoId && newStatus !== null) {
      currentTask.value = { ...currentTask.value, status: newStatus }
    }
  }
}

const handleTranscode = async (video) => {
  updateVideoStatus(video.id, 1)
  currentTask.value = {
    processing: true,
    videoId: video.id,
    videoTitle: video.title,
    stage: 'TRANSCODING',
    stageText: '视频转码中',
    progress: 0,
    status: 1,
    statusText: getProcessStatusText(1)
  }
  try {
    const res = await manualTranscode(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('转码完成')
    } else {
      ElMessage.error(res.message || '转码失败')
    }
  } catch (error) {
    // timeout or network error — backend may still succeed, refresh to check
  }
  currentTask.value = { processing: false }
  loadVideos()
  loadStatistics()
  loadQueueInfo()
}

const handleExtractAudio = async (video) => {
  updateVideoStatus(video.id, 2)
  currentTask.value = {
    processing: true,
    videoId: video.id,
    videoTitle: video.title,
    stage: 'AUDIO_EXTRACTING',
    stageText: '音频提取中',
    progress: 0,
    status: 2,
    statusText: getProcessStatusText(2)
  }
  try {
    const res = await manualExtractAudio(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('音频提取完成')
    } else {
      ElMessage.error(res.message || '提取音频失败')
    }
  } catch (error) {
    // timeout or network error — backend may still succeed, refresh to check
  }
  currentTask.value = { processing: false }
  loadVideos()
  loadStatistics()
  loadQueueInfo()
}

const handleGenerateSubtitle = async (video) => {
  updateVideoStatus(video.id, 3)
  currentTask.value = {
    processing: true,
    videoId: video.id,
    videoTitle: video.title,
    stage: 'SUBTITLE_GENERATING',
    stageText: '字幕生成中',
    progress: 0,
    status: 3,
    statusText: getProcessStatusText(3)
  }
  try {
    const res = await manualGenerateSubtitle(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('字幕生成完成')
    } else {
      ElMessage.error(res.message || '生成字幕失败')
    }
  } catch (error) {
    // timeout or network error — backend may still succeed, refresh to check
  }
  currentTask.value = { processing: false }
  loadVideos()
  loadStatistics()
  loadQueueInfo()
}

const handleAiSummary = async (video) => {
  updateVideoStatus(video.id, 4)
  currentTask.value = {
    processing: true,
    videoId: video.id,
    videoTitle: video.title,
    stage: 'AI_SUMMARIZING',
    stageText: 'AI总结中',
    progress: 0,
    status: 4,
    statusText: getProcessStatusText(4)
  }
  try {
    const res = await manualAiSummary(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('AI总结完成')
    } else {
      ElMessage.error(res.message || 'AI总结失败')
    }
  } catch (error) {
    // timeout or network error — backend may still succeed, refresh to check
  }
  currentTask.value = { processing: false }
  loadVideos()
  loadStatistics()
  loadQueueInfo()
}

const handleReset = async (video) => {
  try {
    await ElMessageBox.confirm(
      `确定要将视频 "${video.title}" 重置为未处理状态吗？`,
      '确认重置',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await resetVideoStatus(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('重置成功')
      updateVideoStatus(video.id, 0)
    } else {
      ElMessage.error(res.message || '重置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重置异常: ' + (error.message || '未知错误'))
    }
  }
}

const getProcessStatusType = (processStatus) => {
  if (processStatus === 5 || processStatus === 11 || processStatus === 21 || processStatus === 31 || processStatus === 41) return 'success'
  if (processStatus === 10 || processStatus === 20 || processStatus === 30 || processStatus === 40) return 'danger'
  if (processStatus === 1 || processStatus === 2 || processStatus === 3 || processStatus === 4) return 'primary'
  return 'info'
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
</script>

<style scoped>
.video-process-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.page-header h2 {
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 15px;
}

.page-desc {
  color: #666;
  margin-bottom: 20px;
}

.current-task-card {
  margin-bottom: 20px;
  border: 2px solid #409eff;
}

.card-header {
  display: flex;
  align-items: center;
  font-weight: bold;
  color: #409eff;
}

.processing-icon {
  margin-right: 8px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.current-task-content {
  padding: 10px 0;
}

.task-info {
  margin-bottom: 15px;
}

.task-label {
  color: #909399;
  margin-right: 8px;
}

.task-value {
  font-weight: 500;
}

.task-progress {
  margin-bottom: 15px;
}

.stage-text {
  display: block;
  margin-bottom: 10px;
  font-size: 16px;
  font-weight: 500;
  color: #409eff;
}

.task-meta {
  color: #909399;
  font-size: 14px;
}

.idle-card {
  margin-bottom: 20px;
}

.idle-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30px;
}

.idle-text {
  margin-top: 15px;
  font-size: 16px;
  color: #909399;
}

.queue-detail {
  margin-top: 20px;
  width: 100%;
}

.statistics-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
  transition: all 0.3s;
}

.stat-card.highlight {
  border-color: #409eff;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.stat-value.processing {
  color: #e6a23c;
  animation: pulse 2s infinite;
}

.stat-value.success {
  color: #67c23a;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 8px;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.filter-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.video-detail {
  padding: 20px;
  background: #f5f7fa;
}

.pipeline-progress {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 200px;
}

.pipeline-step {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 14px;
  position: relative;
  z-index: 1;
}

.pipeline-step.pending {
  background: #e4e7ed;
  color: #909399;
}

.pipeline-step.current {
  background: #409eff;
  color: #fff;
  animation: pulse-step 2s infinite;
}

.pipeline-step.completed {
  background: #67c23a;
  color: #fff;
}

.pipeline-step.error {
  background: #f56c6c;
  color: #fff;
}

.pipeline-step:hover {
  transform: scale(1.1);
}

@keyframes pulse-step {
  0%, 100% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4); }
  50% { box-shadow: 0 0 0 8px rgba(64, 158, 255, 0); }
}

.pipeline-line {
  width: 20px;
  height: 2px;
  background: #e4e7ed;
  transition: all 0.3s;
}

.pipeline-line.active {
  background: #67c23a;
}
</style>
