<template>
  <div class="operation-tasks-page">
    <div class="page-header">
      <div>
        <h2>任务中心</h2>
        <p>统一查看上传、视频处理和存储迁移任务</p>
      </div>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
    </div>

    <div class="filter-bar">
      <el-select v-model="searchForm.taskType" clearable placeholder="任务类型" class="filter-item">
        <el-option v-for="item in taskTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-select v-model="searchForm.status" clearable placeholder="任务状态" class="filter-item">
        <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-input v-model="searchForm.keyword" clearable placeholder="任务名/Key/说明" class="filter-item" />
      <el-input v-model="searchForm.targetKeyword" clearable placeholder="对象类型/ID" class="filter-item" />
      <el-date-picker v-model="searchForm.startTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" placeholder="开始时间" class="date-item" />
      <el-date-picker v-model="searchForm.endTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" placeholder="结束时间" class="date-item" />
      <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <el-table :data="tableData" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="taskName" label="任务名称" min-width="180" show-overflow-tooltip />
      <el-table-column label="类型" width="130">
        <template #default="{ row }">
          <el-tag :type="taskTypeTag(row.taskType)" size="small">{{ formatTaskType(row.taskType) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ formatStatus(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="进度" width="160">
        <template #default="{ row }">
          <el-progress :percentage="normalizeProgress(row.progress)" :status="progressStatus(row.status)" />
        </template>
      </el-table-column>
      <el-table-column prop="stage" label="阶段" min-width="150" show-overflow-tooltip />
      <el-table-column label="对象" min-width="180">
        <template #default="{ row }">
          {{ row.targetType || '-' }}<span v-if="row.targetId"> / {{ row.targetId }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作人" width="140">
        <template #default="{ row }">
          <div class="operator-cell">
            <span>{{ row.operatorName || '-' }}</span>
            <span class="muted">#{{ row.operatorId || '-' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">{{ formatTime(row.updatedAt || row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="showDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>

    <el-drawer v-model="detailVisible" title="任务详情" size="520px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="currentDetail" :column="1" border>
          <el-descriptions-item label="任务ID">{{ currentDetail.id }}</el-descriptions-item>
          <el-descriptions-item label="任务Key">{{ currentDetail.taskKey || '-' }}</el-descriptions-item>
          <el-descriptions-item label="任务名称">{{ currentDetail.taskName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ formatTaskType(currentDetail.taskType) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentDetail.status)">{{ formatStatus(currentDetail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="进度">{{ normalizeProgress(currentDetail.progress) }}%</el-descriptions-item>
          <el-descriptions-item label="阶段">{{ currentDetail.stage || '-' }}</el-descriptions-item>
          <el-descriptions-item label="对象">{{ currentDetail.targetType || '-' }} / {{ currentDetail.targetId || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ currentDetail.operatorName || '-' }} #{{ currentDetail.operatorId || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatTime(currentDetail.startedAt) }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ formatTime(currentDetail.finishedAt) }}</el-descriptions-item>
          <el-descriptions-item label="消息">{{ currentDetail.message || '-' }}</el-descriptions-item>
          <el-descriptions-item label="错误">
            <pre class="detail-text">{{ currentDetail.errorMessage || '-' }}</pre>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getOperationTaskDetail, getOperationTasks } from '../api/operationTask'

const loading = ref(false)
const detailLoading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const size = ref(10)
const detailVisible = ref(false)
const currentDetail = ref(null)

const searchForm = reactive({
  taskType: '',
  status: '',
  keyword: '',
  targetKeyword: '',
  startTime: '',
  endTime: ''
})

const taskTypeOptions = [
  { label: '上传稿件', value: 'UPLOAD' },
  { label: '视频处理', value: 'VIDEO_PROCESS' },
  { label: '存储迁移', value: 'STORAGE_MIGRATION' },
  { label: '推荐刷新', value: 'RECOMMEND_REFRESH' }
]

const statusOptions = [
  { label: '待处理', value: 'PENDING' },
  { label: '处理中', value: 'RUNNING' },
  { label: '成功', value: 'SUCCESS' },
  { label: '失败', value: 'FAILED' },
  { label: '已取消', value: 'CANCELLED' }
]

const loadData = async () => {
  loading.value = true
  try {
    const res = await getOperationTasks({
      page: page.value,
      size: size.value,
      ...searchForm
    })
    if (res.code === 200) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
      return
    }
    ElMessage.error(res.message || '加载任务失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handleReset = () => {
  searchForm.taskType = ''
  searchForm.status = ''
  searchForm.keyword = ''
  searchForm.targetKeyword = ''
  searchForm.startTime = ''
  searchForm.endTime = ''
  page.value = 1
  loadData()
}

const handlePageChange = (p) => {
  page.value = p
  loadData()
}

const handleSizeChange = (s) => {
  size.value = s
  page.value = 1
  loadData()
}

const showDetail = async (row) => {
  detailVisible.value = true
  detailLoading.value = true
  currentDetail.value = null
  try {
    const res = await getOperationTaskDetail(row.id)
    if (res.code === 200) {
      currentDetail.value = res.data
      return
    }
    ElMessage.error(res.message || '加载任务详情失败')
  } finally {
    detailLoading.value = false
  }
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

const normalizeProgress = (value) => {
  if (value == null || Number.isNaN(Number(value))) return 0
  return Math.max(0, Math.min(100, Number(value)))
}

const formatTaskType = (type) => ({
  UPLOAD: '上传稿件',
  VIDEO_PROCESS: '视频处理',
  AI_PIPELINE: 'AI任务',
  STORAGE_MIGRATION: '存储迁移',
  RECOMMEND_REFRESH: '推荐刷新'
}[type] || type || '-')

const formatStatus = (status) => ({
  PENDING: '待处理',
  RUNNING: '处理中',
  SUCCESS: '成功',
  FAILED: '失败',
  CANCELLED: '已取消'
}[status] || status || '-')

const statusTag = (status) => ({
  PENDING: 'info',
  RUNNING: 'warning',
  SUCCESS: 'success',
  FAILED: 'danger',
  CANCELLED: 'info'
}[status] || 'info')

const taskTypeTag = (type) => ({
  UPLOAD: 'primary',
  VIDEO_PROCESS: 'success',
  AI_PIPELINE: 'warning',
  STORAGE_MIGRATION: 'danger',
  RECOMMEND_REFRESH: 'info'
}[type] || 'info')

const progressStatus = (status) => {
  if (status === 'FAILED') return 'exception'
  if (status === 'SUCCESS') return 'success'
  return ''
}

loadData()
</script>

<style scoped>
.operation-tasks-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0;
  font-size: 18px;
}

.page-header p {
  margin: 6px 0 0;
  color: #606266;
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-bottom: 16px;
}

.filter-item {
  width: 170px;
}

.date-item {
  width: 190px;
}

.operator-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.muted {
  color: #909399;
  font-size: 12px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.detail-text {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  color: #606266;
}
</style>
