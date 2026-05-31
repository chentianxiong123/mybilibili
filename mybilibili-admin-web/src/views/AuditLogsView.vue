<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getAuditLogDetail, getAuditLogs } from '../api/audit'

const loading = ref(false)
const detailLoading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const size = ref(10)
const detailVisible = ref(false)
const currentDetail = ref(null)

const searchForm = reactive({
  operatorKeyword: '',
  module: '',
  action: '',
  result: null,
  targetKeyword: '',
  startTime: '',
  endTime: ''
})

const moduleOptions = [
  { label: '认证', value: 'auth' },
  { label: '管理员', value: 'admin' },
  { label: '角色权限', value: 'role' },
  { label: '用户', value: 'user' },
  { label: '稿件', value: 'manuscript' },
  { label: '任务', value: 'task' },
  { label: '存储', value: 'storage' },
  { label: '推荐', value: 'recommend' }
]

const actionLabels = {
  admin_login: '管理员登录',
  admin_register: '新增管理员',
  admin_update: '修改管理员',
  admin_role_assign: '分配管理员角色',
  role_create: '新增角色',
  role_update: '编辑角色',
  role_delete: '删除角色',
  role_permission_update: '设置角色权限',
  user_status_update: '修改用户状态',
  user_password_reset: '重置用户密码',
  manuscript_approve: '稿件审核通过',
  manuscript_approve_with_process: '审核通过并处理',
  manuscript_reject: '稿件审核拒绝',
  manuscript_publish: '发布稿件',
  manuscript_unpublish: '下架稿件',
  manuscript_retry_publish: '重试发布稿件',
  video_transcode_trigger: '触发视频转码',
  video_audio_trigger: '触发音频提取',
  video_subtitle_trigger: '触发字幕生成',
  video_ai_summary_trigger: '触发AI总结',
  video_process_all_trigger: '触发全流程处理',
  video_process_reset: '重置视频处理',
  video_transcode_execute: '执行视频转码',
  video_audio_execute: '执行音频提取',
  video_subtitle_execute: '执行字幕生成',
  video_ai_summary_execute: '执行AI总结',
  pipeline_submit: '提交流水线任务',
  pipeline_cancel: '取消流水线任务',
  pipeline_retry: '重试流水线任务',
  pipeline_clear: '清空流水线队列',
  storage_migration_trigger: '触发存储迁移',
  recommend_config_update: '更新推荐配置',
  recommend_config_reset: '重置推荐配置'
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getAuditLogs({
      page: page.value,
      size: size.value,
      ...searchForm
    })
    if (res.code === 200) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
      return
    }
    ElMessage.error(res.message || '加载审计日志失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handleReset = () => {
  searchForm.operatorKeyword = ''
  searchForm.module = ''
  searchForm.action = ''
  searchForm.result = null
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
    const res = await getAuditLogDetail(row.id)
    if (res.code === 200) {
      currentDetail.value = res.data
      return
    }
    ElMessage.error(res.message || '加载审计详情失败')
  } finally {
    detailLoading.value = false
  }
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

const formatAction = (action) => actionLabels[action] || action || '-'
const resultType = (result) => result === 1 ? 'success' : 'danger'
const resultText = (result) => result === 1 ? '成功' : '失败'

loadData()
</script>

<template>
  <div class="audit-logs-page">
    <div class="page-header">
      <h2>审计日志</h2>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
    </div>

    <div class="search-bar">
      <el-input v-model="searchForm.operatorKeyword" placeholder="操作者ID/用户名" clearable class="search-item" />
      <el-select v-model="searchForm.module" placeholder="模块" clearable class="search-item">
        <el-option v-for="item in moduleOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-input v-model="searchForm.action" placeholder="动作编码" clearable class="search-item" />
      <el-select v-model="searchForm.result" placeholder="结果" clearable class="small-item">
        <el-option label="成功" :value="1" />
        <el-option label="失败" :value="0" />
      </el-select>
      <el-input v-model="searchForm.targetKeyword" placeholder="对象类型/ID" clearable class="search-item" />
      <el-date-picker v-model="searchForm.startTime" type="datetime" placeholder="开始时间" value-format="YYYY-MM-DD HH:mm:ss" class="date-item" />
      <el-date-picker v-model="searchForm.endTime" type="datetime" placeholder="结束时间" value-format="YYYY-MM-DD HH:mm:ss" class="date-item" />
      <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <el-table :data="tableData" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="86" />
      <el-table-column label="操作者" width="170">
        <template #default="{ row }">
          <div class="operator-cell">
            <span>{{ row.operatorName || '-' }}</span>
            <span class="muted">#{{ row.operatorId || '-' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="module" label="模块" width="110" />
      <el-table-column label="动作" min-width="150">
        <template #default="{ row }">{{ formatAction(row.action) }}</template>
      </el-table-column>
      <el-table-column label="对象" min-width="150">
        <template #default="{ row }">
          {{ row.targetType || '-' }}<span v-if="row.targetId"> / {{ row.targetId }}</span>
        </template>
      </el-table-column>
      <el-table-column label="结果" width="90">
        <template #default="{ row }">
          <el-tag :type="resultType(row.result)">{{ resultText(row.result) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="结果说明" min-width="180" show-overflow-tooltip />
      <el-table-column label="时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
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

    <el-drawer v-model="detailVisible" title="审计详情" size="520px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="currentDetail" :column="1" border>
          <el-descriptions-item label="日志ID">{{ currentDetail.id }}</el-descriptions-item>
          <el-descriptions-item label="操作者">{{ currentDetail.operatorName || '-' }} #{{ currentDetail.operatorId || '-' }}</el-descriptions-item>
          <el-descriptions-item label="角色">{{ currentDetail.operatorRole || '-' }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentDetail.module }}</el-descriptions-item>
          <el-descriptions-item label="动作">{{ formatAction(currentDetail.action) }}</el-descriptions-item>
          <el-descriptions-item label="对象">{{ currentDetail.targetType || '-' }} / {{ currentDetail.targetId || '-' }}</el-descriptions-item>
          <el-descriptions-item label="请求">{{ currentDetail.requestMethod || '-' }} {{ currentDetail.requestUri || '' }}</el-descriptions-item>
          <el-descriptions-item label="来源IP">{{ currentDetail.clientIp || '-' }}</el-descriptions-item>
          <el-descriptions-item label="结果">
            <el-tag :type="resultType(currentDetail.result)">{{ resultText(currentDetail.result) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="结果说明">{{ currentDetail.message || '-' }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ formatTime(currentDetail.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="User-Agent">{{ currentDetail.userAgent || '-' }}</el-descriptions-item>
          <el-descriptions-item label="详情">
            <pre class="detail-text">{{ currentDetail.detail || '-' }}</pre>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.audit-logs-page {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.page-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.search-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-bottom: 20px;
}

.search-item {
  width: 180px;
}

.small-item {
  width: 110px;
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
