<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTicketList, processTicket, deleteTicket } from '../api/supportTicket'

const loading = ref(false)
const ticketList = ref([])
const statusFilter = ref('')
const expandedRows = ref([])
const processDialogVisible = ref(false)
const currentTicket = ref(null)
const processForm = ref({ id: null, adminReply: '' })

const statusMap = {
  PENDING: { label: '待处理', type: 'warning' },
  PROCESSING: { label: '处理中', type: 'primary' },
  PROCESSED: { label: '已处理', type: 'success' },
  CLOSED: { label: '已关闭', type: 'info' }
}

const sourceMap = {
  USER_FEEDBACK: '平台反馈',
  AI_CUSTOMER_SERVICE: 'AI客服',
  HUMAN_CUSTOMER_SERVICE: '人工客服',
  ADMIN: '后台创建'
}

const categoryMap = {
  ACCOUNT: '账号问题',
  CONTENT_REVIEW: '内容审核',
  VIDEO_PLAYBACK: '视频播放',
  UPLOAD: '投稿上传',
  AI_FEATURE: 'AI功能',
  COMPLAINT: '投诉申诉',
  GENERAL: '一般问题'
}

const priorityMap = {
  LOW: { label: '低', type: 'info' },
  NORMAL: { label: '普通', type: '' },
  HIGH: { label: '高', type: 'warning' },
  URGENT: { label: '紧急', type: 'danger' }
}

async function loadData() {
  loading.value = true
  try {
    const params = statusFilter.value ? { status: statusFilter.value } : {}
    const res = await getTicketList(params)
    if (res.code === 200) {
      ticketList.value = res.data
    }
  } catch (e) {
    console.error('加载工单列表失败', e)
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  loadData()
}

function openProcessDialog(row) {
  currentTicket.value = row
  processForm.value = { id: row.id, adminReply: '' }
  processDialogVisible.value = true
}

async function handleProcess() {
  if (!processForm.value.adminReply.trim()) {
    return
  }
  try {
    const res = await processTicket(processForm.value.id, processForm.value.adminReply)
    if (res.code !== 200) {
      ElMessage.error(res.message || '处理工单失败')
      return
    }
    ElMessage.success('工单已处理')
    processDialogVisible.value = false
    currentTicket.value = null
    loadData()
  } catch (e) {
    console.error('处理工单失败', e)
    ElMessage.error('处理工单失败')
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除工单 ${row.ticketNo || row.id}？`, '删除工单', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteTicket(row.id)
    if (res.code !== 200) {
      ElMessage.error(res.message || '删除工单失败')
      return
    }
    ElMessage.success('工单已删除')
    loadData()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('删除工单失败', e)
      ElMessage.error('删除工单失败')
    }
  }
}

function getStatusType(status) {
  return statusMap[status]?.type || 'info'
}

function getStatusLabel(status) {
  return statusMap[status]?.label || status || '-'
}

function getSourceLabel(source) {
  return sourceMap[source] || source || '-'
}

function getCategoryLabel(category) {
  return categoryMap[category] || category || '-'
}

function getPriorityLabel(priority) {
  return priorityMap[priority]?.label || priority || '-'
}

function getPriorityType(priority) {
  return priorityMap[priority]?.type || 'info'
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString() : '-'
}

onMounted(loadData)
</script>

<template>
  <div class="support-tickets">
    <div class="page-header">
      <h2 class="page-title">工单中心</h2>
      <p class="page-desc">处理来自 AI 客服、人工客服和平台反馈入口的用户诉求。</p>
    </div>

    <div class="filter-bar">
      <el-radio-group v-model="statusFilter" @change="handleFilter">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="PENDING">待处理</el-radio-button>
        <el-radio-button label="PROCESSING">处理中</el-radio-button>
        <el-radio-button label="PROCESSED">已处理</el-radio-button>
        <el-radio-button label="CLOSED">已关闭</el-radio-button>
      </el-radio-group>
    </div>

    <el-table
      :data="ticketList"
      border
      stripe
      v-loading="loading"
      style="margin-top: 16px"
      :expand-row-keys="expandedRows"
      row-key="id"
      @expand-change="(row, rows) => expandedRows = rows"
    >
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="expand-content">
            <div class="expand-grid">
              <div>工单编号：{{ row.ticketNo || '-' }}</div>
              <div>来源：{{ getSourceLabel(row.source) }}</div>
              <div>分类：{{ getCategoryLabel(row.category) }}</div>
              <div>优先级：{{ getPriorityLabel(row.priority) }}</div>
              <div>用户ID：{{ row.userId || '-' }}</div>
              <div>客服会话：{{ row.sessionId ? `#${row.sessionId}` : '-' }}</div>
              <div>创建时间：{{ formatTime(row.createdAt) }}</div>
              <div>处理时间：{{ formatTime(row.processedAt) }}</div>
            </div>
            <div class="expand-item">
              <strong>工单标题：</strong>
              <p>{{ row.title || '-' }}</p>
            </div>
            <div class="expand-item">
              <strong>用户诉求：</strong>
              <p>{{ row.content || '-' }}</p>
            </div>
            <div class="expand-item" v-if="row.entryReply">
              <strong>入口回复：</strong>
              <p>{{ row.entryReply }}</p>
            </div>
            <div class="expand-item" v-if="row.adminReply">
              <strong>处理结果：</strong>
              <p>{{ row.adminReply }}</p>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="工单编号" prop="ticketNo" width="180">
        <template #default="{ row }">
          <span class="mono">{{ row.ticketNo || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="来源" width="110">
        <template #default="{ row }">
          <el-tag size="small" type="primary">{{ getSourceLabel(row.source) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="用户ID" prop="userId" width="100">
        <template #default="{ row }">
          {{ row.userId || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="分类" width="110">
        <template #default="{ row }">
          {{ getCategoryLabel(row.category) }}
        </template>
      </el-table-column>
      <el-table-column label="优先级" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="getPriorityType(row.priority)">
            {{ getPriorityLabel(row.priority) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="标题" prop="title" min-width="180">
        <template #default="{ row }">
          <span class="text-ellipsis">{{ row.title || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="用户诉求" prop="content" min-width="240">
        <template #default="{ row }">
          <span class="text-ellipsis">{{ row.content || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时间" prop="createdAt" width="180">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button
            type="primary"
            size="small"
            :disabled="row.status === 'PROCESSED'"
            @click="openProcessDialog(row)"
          >
            处理
          </el-button>
          <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="processDialogVisible" title="处理工单" width="500px">
      <div v-if="currentTicket" class="dialog-summary">
        <div>{{ currentTicket.ticketNo || `#${currentTicket.id}` }}</div>
        <strong>{{ currentTicket.title || '-' }}</strong>
        <p>{{ currentTicket.content || '-' }}</p>
      </div>
      <el-form>
        <el-form-item label="处理结果">
          <el-input
            v-model="processForm.adminReply"
            type="textarea"
            :rows="4"
            placeholder="请输入处理结论或后续动作"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleProcess">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.support-tickets { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-title { margin: 0 0 8px; font-size: 24px; font-weight: 600; color: #333; }
.page-desc { margin: 0; color: #909399; font-size: 14px; }
.filter-bar { margin-bottom: 8px; }
.expand-content { padding: 8px 16px; }
.expand-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 8px 16px;
  margin-bottom: 16px;
  color: #606266;
  font-size: 13px;
}
.expand-item { margin-bottom: 12px; }
.expand-item:last-child { margin-bottom: 0; }
.expand-item p { margin: 4px 0 0; color: #666; white-space: pre-wrap; }
.mono { font-family: Consolas, Monaco, monospace; }
.text-ellipsis {
  display: inline-block;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
.dialog-summary {
  padding: 12px;
  margin-bottom: 16px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #fafafa;
  color: #606266;
}
.dialog-summary strong {
  display: block;
  margin: 4px 0;
  color: #303133;
}
.dialog-summary p {
  margin: 0;
  white-space: pre-wrap;
}
</style>
