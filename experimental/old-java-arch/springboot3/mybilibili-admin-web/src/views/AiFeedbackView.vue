<script setup>
import { ref, onMounted } from 'vue'
import { getFeedbackList, processFeedback, deleteFeedback } from '../api/aiFeedback'

const loading = ref(false)
const feedbackList = ref([])
const statusFilter = ref('')
const expandedRows = ref([])
const processDialogVisible = ref(false)
const processForm = ref({ id: null, adminReply: '' })

async function loadData() {
  loading.value = true
  try {
    const params = statusFilter.value ? { status: statusFilter.value } : {}
    const res = await getFeedbackList(params)
    if (res.code === 200) {
      feedbackList.value = res.data
    }
  } catch (e) {
    console.error('加载反馈列表失败', e)
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  loadData()
}

function openProcessDialog(row) {
  processForm.value = { id: row.id, adminReply: '' }
  processDialogVisible.value = true
}

async function handleProcess() {
  if (!processForm.value.adminReply.trim()) {
    return
  }
  try {
    await processFeedback(processForm.value.id, processForm.value.adminReply)
    processDialogVisible.value = false
    loadData()
  } catch (e) {
    console.error('处理反馈失败', e)
  }
}

async function handleDelete(id) {
  try {
    await deleteFeedback(id)
    loadData()
  } catch (e) {
    console.error('删除反馈失败', e)
  }
}

function getStatusType(status) {
  return status === 'PROCESSED' ? 'success' : 'warning'
}

function getStatusLabel(status) {
  return status === 'PROCESSED' ? '已处理' : '待处理'
}

onMounted(loadData)
</script>

<template>
  <div class="ai-feedback">
    <h2 class="page-title">AI 反馈管理</h2>

    <div class="filter-bar">
      <el-radio-group v-model="statusFilter" @change="handleFilter">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="PENDING">待处理</el-radio-button>
        <el-radio-button label="PROCESSED">已处理</el-radio-button>
      </el-radio-group>
    </div>

    <el-table
      :data="feedbackList"
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
            <div class="expand-item">
              <strong>问题：</strong>
              <p>{{ row.question }}</p>
            </div>
            <div class="expand-item">
              <strong>回答：</strong>
              <p>{{ row.answer }}</p>
            </div>
            <div class="expand-item" v-if="row.adminReply">
              <strong>管理员回复：</strong>
              <p>{{ row.adminReply }}</p>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="用户ID" prop="userId" width="100" />
      <el-table-column label="技能ID" prop="skillId" width="100" />
      <el-table-column label="评分" width="120">
        <template #default="{ row }">
          <el-rate :model-value="row.rating" disabled text-color="#ff9900" />
        </template>
      </el-table-column>
      <el-table-column label="反馈内容" prop="feedbackText" width="200">
        <template #default="{ row }">
          <span class="text-ellipsis">{{ row.feedbackText }}</span>
        </template>
      </el-table-column>
      <el-table-column label="管理员回复" prop="adminReply" width="200">
        <template #default="{ row }">
          <span class="text-ellipsis">{{ row.adminReply || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时间" prop="createdAt" width="180">
        <template #default="{ row }">
          {{ row.createdAt ? new Date(row.createdAt).toLocaleString() : '-' }}
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
          <el-button type="danger" size="small" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="processDialogVisible" title="处理反馈" width="500px">
      <el-form>
        <el-form-item label="管理员回复">
          <el-input
            v-model="processForm.adminReply"
            type="textarea"
            :rows="4"
            placeholder="请输入处理意见"
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
.ai-feedback { padding: 20px; }
.page-title { margin: 0 0 24px; font-size: 24px; font-weight: 600; color: #333; }
.filter-bar { margin-bottom: 8px; }
.expand-content { padding: 8px 16px; }
.expand-item { margin-bottom: 12px; }
.expand-item:last-child { margin-bottom: 0; }
.expand-item p { margin: 4px 0 0; color: #666; white-space: pre-wrap; }
.text-ellipsis {
  display: inline-block;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
</style>