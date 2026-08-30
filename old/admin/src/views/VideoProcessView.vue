<template>
  <div class="video-process-page">
    <h2>视频处理管理</h2>
    <p class="page-desc">以视频为单位进行全流程处理控制，每个步骤可独立操作</p>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="statistics-row">
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ statistics.pending || 0 }}</div>
          <div class="stat-label">待处理</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ statistics.transcoding || 0 }}</div>
          <div class="stat-label">转码中</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ statistics.audioExtracting || 0 }}</div>
          <div class="stat-label">音频提取中</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ statistics.subtitleGenerating || 0 }}</div>
          <div class="stat-label">字幕生成中</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ statistics.aiSummarizing || 0 }}</div>
          <div class="stat-label">AI总结中</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ statistics.completed || 0 }}</div>
          <div class="stat-label">处理完成</div>
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
            <el-option label="转码失败" :value="6" />
            <el-option label="音频提取失败" :value="7" />
            <el-option label="字幕生成失败" :value="8" />
            <el-option label="AI总结失败" :value="9" />
          </el-select>
        </el-form-item>
        <el-form-item label="稿件状态">
          <el-select v-model="filterForm.manuscriptStatus" placeholder="全部" clearable @change="handleFilterChange">
            <el-option label="处理中" :value="1" />
            <el-option label="待上架" :value="2" />
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
      <el-table :data="tableData" v-loading="loading" style="width: 100%">
        <el-table-column type="expand" width="50">
          <template #default="{ row }">
            <div class="video-detail">
              <el-descriptions :column="3" border>
                <el-descriptions-item label="视频ID">{{ row.id }}</el-descriptions-item>
                <el-descriptions-item label="稿件ID">{{ row.manuscriptId }}</el-descriptions-item>
                <el-descriptions-item label="时长">{{ row.duration }}</el-descriptions-item>
                <el-descriptions-item label="字幕">{{ row.hasSubtitle ? '有' : '无' }}</el-descriptions-item>
                <el-descriptions-item label="AI摘要">{{ row.hasSummary ? '有' : '无' }}</el-descriptions-item>
                <el-descriptions-item v-if="row.processError" label="错误信息" :span="3">
                  <span style="color: #f56c6c;">{{ row.processError }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </template>
        </el-table-column>
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
        <el-table-column label="流水线进度" width="200">
          <template #default="{ row }">
            <div class="pipeline-progress">
              <el-tooltip content="转码" placement="top">
                <div 
                  class="pipeline-step" 
                  :class="getStepClass(row.processStatus, 1)"
                  @click="handleStepClick(row, 1)"
                >
                  <el-icon><VideoCamera /></el-icon>
                </div>
              </el-tooltip>
              <div class="pipeline-line" :class="{ active: row.processStatus >= 2 }"></div>
              <el-tooltip content="音频" placement="top">
                <div 
                  class="pipeline-step" 
                  :class="getStepClass(row.processStatus, 2)"
                  @click="handleStepClick(row, 2)"
                >
                  <el-icon><Microphone /></el-icon>
                </div>
              </el-tooltip>
              <div class="pipeline-line" :class="{ active: row.processStatus >= 3 }"></div>
              <el-tooltip content="字幕" placement="top">
                <div 
                  class="pipeline-step" 
                  :class="getStepClass(row.processStatus, 3)"
                  @click="handleStepClick(row, 3)"
                >
                  <el-icon><ChatDotRound /></el-icon>
                </div>
              </el-tooltip>
              <div class="pipeline-line" :class="{ active: row.processStatus >= 4 }"></div>
              <el-tooltip content="AI" placement="top">
                <div 
                  class="pipeline-step" 
                  :class="getStepClass(row.processStatus, 4)"
                  @click="handleStepClick(row, 4)"
                >
                  <el-icon><Cpu /></el-icon>
                </div>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <!-- 转码按钮 -->
            <el-button 
              v-if="canTranscode(row.processStatus)"
              type="primary" 
              size="small"
              @click="handleTranscode(row)"
            >
              <el-icon><VideoCamera /></el-icon>
              转码
            </el-button>
            
            <!-- 音频按钮 -->
            <el-button 
              v-if="canExtractAudio(row.processStatus)"
              type="success" 
              size="small"
              @click="handleExtractAudio(row)"
            >
              <el-icon><Microphone /></el-icon>
              音频
            </el-button>
            
            <!-- 字幕按钮 -->
            <el-button 
              v-if="canGenerateSubtitle(row.processStatus)"
              type="warning" 
              size="small"
              @click="handleGenerateSubtitle(row)"
            >
              <el-icon><ChatDotRound /></el-icon>
              字幕
            </el-button>
            
            <!-- AI按钮 -->
            <el-button
              v-if="canAiSummary(row.processStatus)"
              type="info"
              size="small"
              @click="handleAiSummary(row)"
            >
              <el-icon><Cpu /></el-icon>
              AI
            </el-button>

            <!-- 重置按钮 -->
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

      <!-- 分页 -->
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

    <!-- AI测试对话框 -->
    <el-dialog
      v-model="testDialog.visible"
      title="AI摘要测试"
      width="700px"
    >
      <el-tabs v-model="testDialog.activeTab">
        <el-tab-pane label="API连接测试" name="api">
          <div class="test-section">
            <p class="test-desc">测试DeepSeek API连接是否正常</p>
            <el-input
              v-model="testDialog.testText"
              type="textarea"
              :rows="3"
              placeholder="输入测试文本（可选，默认发送简单问候）"
            />
            <el-button type="primary" @click="handleTestApi" :loading="testDialog.loading">
              测试API连接
            </el-button>
          </div>
          <div v-if="testDialog.apiResult" class="test-result">
            <el-alert
              :title="testDialog.apiResult.success ? '测试成功' : '测试失败'"
              :type="testDialog.apiResult.success ? 'success' : 'error'"
              :description="testDialog.apiResult.message"
              show-icon
            />
            <div v-if="testDialog.apiResult.response" class="response-content">
              <h4>API响应：</h4>
              <pre>{{ testDialog.apiResult.response }}</pre>
              <p class="response-time">响应时间: {{ testDialog.apiResult.responseTime }}</p>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="字幕摘要测试" name="summary">
          <div class="test-section">
            <p class="test-desc">使用当前视频的字幕内容测试AI摘要生成</p>
            <div class="video-info" v-if="testDialog.currentVideo">
              <p><strong>视频ID:</strong> {{ testDialog.currentVideo.id }}</p>
              <p><strong>视频标题:</strong> {{ testDialog.currentVideo.title }}</p>
            </div>
            <el-button type="primary" @click="handleTestSummary" :loading="testDialog.loading">
              生成摘要
            </el-button>
          </div>
          <div v-if="testDialog.summaryResult" class="test-result">
            <el-alert
              :title="testDialog.summaryResult.success ? '生成成功' : '生成失败'"
              :type="testDialog.summaryResult.success ? 'success' : 'error'"
            />
            <div v-if="testDialog.summaryResult.data" class="summary-stats">
              <p>字幕长度: {{ testDialog.summaryResult.data.subtitleLength }} 字符</p>
              <p>估算Token: {{ testDialog.summaryResult.data.subtitleTokenEstimate }}</p>
              <p>摘要长度: {{ testDialog.summaryResult.data.summaryLength }} 字符</p>
              <p>响应时间: {{ testDialog.summaryResult.data.responseTime }}</p>
            </div>
            <div v-if="testDialog.summaryResult.data && testDialog.summaryResult.data.summary" class="summary-content">
              <h4>生成的摘要：</h4>
              <el-input
                v-model="testDialog.summaryResult.data.summary"
                type="textarea"
                :rows="10"
                readonly
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAllManuscripts,
  getManuscriptVideos,
  manualTranscode,
  manualExtractAudio,
  manualGenerateSubtitle,
  manualAiSummary,
  resetVideoStatus,
  testAiApi,
  testAiSummary
} from '../api/manuscript'

const loading = ref(false)
const tableData = ref([])
const statistics = ref({})

const filterForm = reactive({
  processStatus: '',
  manuscriptStatus: '',
  keyword: ''
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

onMounted(() => {
  loadData()
})

const loadData = async () => {
  loading.value = true
  try {
    await loadVideos()
    calculateStatistics()
  } finally {
    loading.value = false
  }
}

const loadVideos = async () => {
  try {
    // 获取所有稿件
    const res = await getAllManuscripts()
    if (res.code === 200 || res.success) {
      const manuscripts = res.data || []
      let allVideos = []
      
      // 获取每个稿件的视频
      for (const manuscript of manuscripts) {
        if (filterForm.manuscriptStatus && manuscript.status !== filterForm.manuscriptStatus) {
          continue
        }
        
        try {
          const videoRes = await getManuscriptVideos(manuscript.id)
          if (videoRes.code === 200 || videoRes.success) {
            const videos = videoRes.data || []
            videos.forEach(video => {
              allVideos.push({
                ...video,
                manuscriptTitle: manuscript.title,
                manuscriptCoverUrl: manuscript.coverUrl,
                manuscriptStatus: manuscript.status
              })
            })
          }
        } catch (error) {
          console.error(`加载稿件 ${manuscript.id} 的视频失败:`, error)
        }
      }
      
      // 过滤
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

const calculateStatistics = () => {
  const stats = {
    pending: 0,
    transcoding: 0,
    audioExtracting: 0,
    subtitleGenerating: 0,
    aiSummarizing: 0,
    completed: 0
  }
  
  tableData.value.forEach(video => {
    switch (video.processStatus) {
      case 0: stats.pending++; break
      case 1: stats.transcoding++; break
      case 2: stats.audioExtracting++; break
      case 3: stats.subtitleGenerating++; break
      case 4: stats.aiSummarizing++; break
      case 5: stats.completed++; break
    }
  })
  
  statistics.value = stats
}

const handleFilterChange = () => {
  pagination.page = 1
  loadData()
}

const resetFilter = () => {
  filterForm.processStatus = ''
  filterForm.manuscriptStatus = ''
  filterForm.keyword = ''
  handleFilterChange()
}

const handleSizeChange = (size) => {
  pagination.size = size
  loadVideos()
}

const handlePageChange = (page) => {
  pagination.page = page
  loadVideos()
}

// 判断是否可以执行某步骤（根据新的状态码规则）
// 0-待处理, 1-转码中, 10-转码失败, 11-转码成功
// 2-音频提取中, 20-音频失败, 21-音频成功
// 3-字幕生成中, 30-字幕失败, 31-字幕成功
// 4-AI总结中, 40-AI失败, 41-AI成功
// 5-全部完成

// 检查步骤是否已经做过（用于确认弹窗）
const isStepDone = (status, stepType) => {
  switch (stepType) {
    case 'transcode':
      return status >= 11 // 转码成功或之后的状态
    case 'audio':
      return status >= 21 // 音频提取成功或之后的状态
    case 'subtitle':
      return status >= 31 // 字幕生成成功或之后的状态
    case 'ai':
      return status >= 41 || status === 5 // AI总结成功或全部完成
    default:
      return false
  }
}

const canTranscode = (status) => {
  // 待处理、转码失败、音频失败、字幕失败、AI失败、以及已完成状态下可以转码（重复执行）
  return status === 0 || status === 10 || status === 20 || status === 30 || status === 40 || status >= 11
}

const canExtractAudio = (status) => {
  // 转码成功、音频失败、以及已完成状态下可以提取音频（重复执行）
  return status === 11 || status === 20 || status >= 21
}

const canGenerateSubtitle = (status) => {
  // 音频提取成功、字幕失败、以及已完成状态下可以生成字幕（重复执行）
  return status === 21 || status === 30 || status >= 31
}

const canAiSummary = (status) => {
  // 字幕生成成功、AI失败、以及已完成状态下可以进行AI总结（重复执行）
  return status === 31 || status === 40 || status >= 41 || status === 5
}

// 流水线步骤样式
// step: 1-转码, 2-音频, 3-字幕, 4-AI
const getStepClass = (processStatus, step) => {
  // 定义每个步骤的状态范围
  const stepRanges = {
    1: { min: 1, max: 11, success: 11, fail: 10, processing: 1 },  // 转码
    2: { min: 2, max: 21, success: 21, fail: 20, processing: 2 },  // 音频
    3: { min: 3, max: 31, success: 31, fail: 30, processing: 3 },  // 字幕
    4: { min: 4, max: 41, success: 41, fail: 40, processing: 4 }   // AI
  }
  
  const range = stepRanges[step]
  
  // 当前正在处理这个步骤
  if (processStatus === range.processing) return 'current'
  
  // 这个步骤失败了
  if (processStatus === range.fail) return 'error'
  
  // 这个步骤已完成（状态大于等于成功状态，或者全部完成）
  if (processStatus >= range.success || processStatus === 5) return 'completed'
  
  // 待处理
  return 'pending'
}

const handleStepClick = (video, step) => {
  // 根据步骤执行相应操作
  switch (step) {
    case 1:
      if (canTranscode(video.processStatus)) handleTranscode(video)
      break
    case 2:
      if (canExtractAudio(video.processStatus)) handleExtractAudio(video)
      break
    case 3:
      if (canGenerateSubtitle(video.processStatus)) handleGenerateSubtitle(video)
      break
    case 4:
      if (canAiSummary(video.processStatus)) handleAiSummary(video)
      break
  }
}

// 显示重复执行确认弹窗
const showConfirmDialog = async (stepName, stepType) => {
  try {
    await ElMessageBox.confirm(
      `该视频已经执行过${stepName}，是否确认重新执行？`,
      '确认重新执行',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    return true
  } catch {
    return false
  }
}

// 更新单个视频状态（本地更新，不重新加载整个列表）
const updateVideoStatus = (videoId, newStatus) => {
  const index = tableData.value.findIndex(v => v.id === videoId)
  if (index !== -1) {
    // 使用解构创建新对象，确保Vue检测到变化
    tableData.value[index] = {
      ...tableData.value[index],
      processStatus: newStatus
    }
    calculateStatistics()
  }
}

// 视频处理操作
const handleTranscode = async (video) => {
  // 如果已经做过，先询问确认
  if (isStepDone(video.processStatus, 'transcode')) {
    const confirmed = await showConfirmDialog('转码', 'transcode')
    if (!confirmed) return
  }
  try {
    const res = await manualTranscode(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始转码')
      updateVideoStatus(video.id, 1) // 1-转码中
    } else {
      ElMessage.error(res.message || '转码失败')
    }
  } catch (error) {
    ElMessage.error('转码异常: ' + (error.message || '未知错误'))
  }
}

const handleExtractAudio = async (video) => {
  // 如果已经做过，先询问确认
  if (isStepDone(video.processStatus, 'audio')) {
    const confirmed = await showConfirmDialog('音频提取', 'audio')
    if (!confirmed) return
  }
  try {
    const res = await manualExtractAudio(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始提取音频')
      updateVideoStatus(video.id, 2) // 2-音频提取中
    } else {
      ElMessage.error(res.message || '提取音频失败')
    }
  } catch (error) {
    ElMessage.error('提取音频异常: ' + (error.message || '未知错误'))
  }
}

const handleGenerateSubtitle = async (video) => {
  // 如果已经做过，先询问确认
  if (isStepDone(video.processStatus, 'subtitle')) {
    const confirmed = await showConfirmDialog('字幕生成', 'subtitle')
    if (!confirmed) return
  }
  try {
    const res = await manualGenerateSubtitle(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始生成字幕')
      updateVideoStatus(video.id, 3) // 3-字幕生成中
    } else {
      ElMessage.error(res.message || '生成字幕失败')
    }
  } catch (error) {
    ElMessage.error('生成字幕异常: ' + (error.message || '未知错误'))
  }
}

const handleAiSummary = async (video) => {
  // 如果已经做过，先询问确认
  if (isStepDone(video.processStatus, 'ai')) {
    const confirmed = await showConfirmDialog('AI总结', 'ai')
    if (!confirmed) return
  }
  try {
    const res = await manualAiSummary(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已开始AI总结')
      updateVideoStatus(video.id, 4) // 4-AI总结中
    } else {
      ElMessage.error(res.message || 'AI总结失败')
    }
  } catch (error) {
    ElMessage.error('AI总结异常: ' + (error.message || '未知错误'))
  }
}

// 重置视频处理状态
const handleReset = async (video) => {
  try {
    await ElMessageBox.confirm(
      `确定要将视频 "${video.title}" 重置为未处理状态吗？`,
      '确认重置',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const res = await resetVideoStatus(video.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('重置成功')
      updateVideoStatus(video.id, 0) // 0-待处理
    } else {
      ElMessage.error(res.message || '重置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重置异常: ' + (error.message || '未知错误'))
    }
  }
}

// ==================== AI测试功能 ====================

const testDialog = reactive({
  visible: false,
  activeTab: 'api',
  testText: '',
  loading: false,
  currentVideo: null,
  apiResult: null,
  summaryResult: null
})

const openTestDialog = (video) => {
  testDialog.currentVideo = video
  testDialog.visible = true
  testDialog.activeTab = 'api'
  testDialog.testText = ''
  testDialog.apiResult = null
  testDialog.summaryResult = null
}

const handleTestApi = async () => {
  testDialog.loading = true
  testDialog.apiResult = null
  try {
    const res = await testAiApi(testDialog.testText)
    testDialog.apiResult = {
      success: res.code === 200 || res.success,
      message: res.message || (res.data && res.data.message) || '未知',
      response: res.data && res.data.response,
      responseTime: res.data && res.data.responseTime
    }
    if (testDialog.apiResult.success) {
      ElMessage.success('API连接测试成功')
    } else {
      ElMessage.error('API连接测试失败')
    }
  } catch (error) {
    testDialog.apiResult = {
      success: false,
      message: error.message || '请求异常'
    }
    ElMessage.error('测试异常: ' + error.message)
  } finally {
    testDialog.loading = false
  }
}

const handleTestSummary = async () => {
  if (!testDialog.currentVideo) {
    ElMessage.error('未选择视频')
    return
  }
  testDialog.loading = true
  testDialog.summaryResult = null
  try {
    const res = await testAiSummary(testDialog.currentVideo.id)
    testDialog.summaryResult = {
      success: res.code === 200 || res.success,
      data: res.data
    }
    if (testDialog.summaryResult.success) {
      ElMessage.success('摘要生成成功')
    } else {
      ElMessage.error(res.message || '摘要生成失败')
    }
  } catch (error) {
    testDialog.summaryResult = {
      success: false,
      message: error.message || '请求异常'
    }
    ElMessage.error('测试异常: ' + error.message)
  } finally {
    testDialog.loading = false
  }
}

// 状态显示（根据新的状态码规则）
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

.page-desc {
  color: #666;
  margin-bottom: 20px;
}

.statistics-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 8px;
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

/* 流水线进度样式 */
.pipeline-progress {
  display: flex;
  align-items: center;
  gap: 4px;
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
}

.pipeline-step.pending {
  background: #e4e7ed;
  color: #909399;
}

.pipeline-step.current {
  background: #409eff;
  color: #fff;
  animation: pulse 2s infinite;
}

.pipeline-step.completed {
  background: #67c23a;
  color: #fff;
}

.pipeline-step.error {
  background: #f56c6c;
  color: #fff;
  animation: shake 0.5s;
}

.pipeline-step:hover {
  transform: scale(1.1);
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-3px); }
  75% { transform: translateX(3px); }
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

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(64, 158, 255, 0);
  }
}

/* AI测试对话框样式 */
.test-section {
  padding: 20px 0;
}

.test-desc {
  color: #666;
  margin-bottom: 15px;
}

.test-section .el-input {
  margin-bottom: 15px;
}

.test-result {
  margin-top: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}

.response-content {
  margin-top: 15px;
}

.response-content h4 {
  margin-bottom: 10px;
  color: #333;
}

.response-content pre {
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 200px;
  overflow-y: auto;
}

.response-time {
  margin-top: 10px;
  color: #909399;
  font-size: 12px;
}

.video-info {
  background: #f5f7fa;
  padding: 10px 15px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.video-info p {
  margin: 5px 0;
}

.summary-stats {
  margin: 15px 0;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.summary-stats p {
  margin: 5px 0;
  color: #606266;
}

.summary-content {
  margin-top: 15px;
}

.summary-content h4 {
  margin-bottom: 10px;
  color: #333;
}
</style>
