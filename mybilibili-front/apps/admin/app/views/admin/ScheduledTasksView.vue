<template>
  <div class="scheduled-tasks-page">
    <div class="page-header">
      <div>
        <h2>定时任务管理</h2>
        <p>统一管理后台定时任务</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">新建任务</el-button>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
    </div>

    <el-table :data="tableData" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="taskName" label="任务名称" min-width="150" />
      <el-table-column prop="taskKey" label="任务Key" width="160" />
      <el-table-column prop="taskType" label="类型" width="150">
        <template #default="{ row }">
          <el-tag :type="row.taskType === 'hot_search_cleanup' ? 'warning' : 'info'" size="small">{{ row.taskType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="cronExpr" label="Cron表达式" width="140" />
      <el-table-column prop="lastRunResult" label="上次结果" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.lastRunResult === 'success'" type="success" size="small">成功</el-tag>
          <el-tag v-else-if="row.lastRunResult === 'failed'" type="danger" size="small">失败</el-tag>
          <el-tag v-else-if="row.lastRunResult === 'skipped'" type="info" size="small">跳过</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="lastRunMessage" label="上次消息" min-width="200" show-overflow-tooltip />
      <el-table-column prop="runCount" label="运行次数" width="90" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-switch :model-value="row.enabled === 1" :loading="togglingId === row.id" @change="handleToggle(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleTrigger(row)">执行</el-button>
          <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreateDialog" :title="editingId ? '编辑任务' : '新建任务'" width="600px">
      <el-form :model="form" label-width="120px">
        <el-form-item label="任务Key" v-if="!editingId">
          <el-input v-model="form.taskKey" placeholder="唯一标识，如 hot_search_cleanup" />
        </el-form-item>
        <el-form-item label="任务名称">
          <el-input v-model="form.taskName" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="Cron表达式">
          <el-input v-model="form.cronExpr" placeholder="如 0 0 3 * * ?" />
        </el-form-item>
        <el-form-item label="任务类型">
          <el-select v-model="form.taskType">
            <el-option label="热搜清理" value="hot_search_cleanup" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务配置">
          <el-input v-model="form.taskConfig" type="textarea" :rows="2" placeholder="JSON 格式配置" />
        </el-form-item>
        <el-form-item label="超时时间">
          <el-input-number v-model="form.timeoutSeconds" :min="10" :max="3600" /> 秒
        </el-form-item>
        <el-form-item label="最大重试">
          <el-input-number v-model="form.maxRetries" :min="0" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { getScheduledTasks, createScheduledTask, updateScheduledTask, toggleScheduledTask, triggerScheduledTask, deleteScheduledTask } from '@/api/scheduledTask'

const loading = ref(false)
const tableData = ref<any[]>([])
const showCreateDialog = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const togglingId = ref<number | null>(null)

const form = ref<any>({
  taskKey: '',
  taskName: '',
  description: '',
  cronExpr: '',
  taskType: 'hot_search_cleanup',
  taskConfig: '',
  timeoutSeconds: 30,
  maxRetries: 3
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await getScheduledTasks()
    if (res.code === 200) {
      tableData.value = res.data?.list || []
    }
  } catch (e: any) {
    ElMessage.error('获取任务列表失败')
  } finally {
    loading.value = false
  }
}

const handleToggle = async (row: any) => {
  togglingId.value = row.id
  try {
    const res = await toggleScheduledTask(row.id, row.enabled === 1 ? 0 : 1)
    if (res.code === 200) {
      row.enabled = row.enabled === 1 ? 0 : 1
    }
  } catch {
    ElMessage.error('操作失败')
  } finally {
    togglingId.value = null
  }
}

const handleTrigger = async (row: any) => {
  try {
    const res = await triggerScheduledTask(row.taskKey)
    if (res.code === 200) {
      ElMessage.success('任务已触发')
      setTimeout(loadData, 2000)
    }
  } catch {
    ElMessage.error('触发失败')
  }
}

const handleEdit = (row: any) => {
  editingId.value = row.id
  form.value = {
    taskKey: row.taskKey,
    taskName: row.taskName,
    description: row.description,
    cronExpr: row.cronExpr,
    taskType: row.taskType,
    taskConfig: row.taskConfig,
    timeoutSeconds: row.timeoutSeconds,
    maxRetries: row.maxRetries
  }
  showCreateDialog.value = true
}

const handleDelete = async (id: number) => {
  try {
    const res = await deleteScheduledTask(id)
    if (res.code === 200) {
      ElMessage.success('已删除')
      loadData()
    }
  } catch {
    ElMessage.error('删除失败')
  }
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    let res
    if (editingId.value) {
      res = await updateScheduledTask({ id: editingId.value, ...form.value })
    } else {
      res = await createScheduledTask(form.value)
    }
    if (res.code === 200) {
      ElMessage.success(editingId.value ? '已更新' : '已创建')
      showCreateDialog.value = false
      editingId.value = null
      form.value = { taskKey: '', taskName: '', description: '', cronExpr: '', taskType: 'hot_search_cleanup', taskConfig: '', timeoutSeconds: 30, maxRetries: 3 }
      loadData()
    }
  } catch {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

onMounted(loadData)
</script>