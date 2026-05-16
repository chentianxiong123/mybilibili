<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getPendingList,
  getAllContent,
  restoreContent,
  deleteContent,
  batchProcess
} from '../api/contentReview'
import {
  getReportList,
  processReport
} from '../api/report'

// 表格数据
const tableData = ref([])
const loading = ref(false)
const total = ref(0)

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

// 标签页
const activeTab = ref('all')

// 内容类型筛选
const contentType = ref('')

// 多选
const selectedItems = ref([])

// 举报相关
const reportStatus = ref('')
const reportType = ref('')
const reportProcessDialog = ref(false)
const currentReport = ref(null)
const reportAction = ref('')
const reportAdminRemark = ref('')

// 内容类型选项
const contentTypeOptions = [
  { label: '全部', value: '' },
  { label: '评论', value: 'COMMENT' },
  { label: '回复', value: 'REPLY' },
  { label: '动态评论', value: 'DYNAMIC_COMMENT' }
]

// 加载所有内容
const loadAllContent = async () => {
  loading.value = true
  try {
    const status = activeTab.value === 'removed' ? 'REMOVED' : (activeTab.value === 'normal' ? 'NORMAL' : '')
    const res = await getAllContent({
      page: currentPage.value,
      size: pageSize.value,
      contentType: contentType.value || undefined,
      status: status || undefined
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || res.data || []
      total.value = res.data?.total || res.total || 0
    }
  } catch (error) {
    ElMessage.error('获取内容列表失败')
  } finally {
    loading.value = false
  }
}

// 加载举报列表
const loadReportData = async () => {
  loading.value = true
  try {
    const res = await getReportList({
      page: currentPage.value,
      size: pageSize.value,
      status: reportStatus.value || undefined,
      targetType: reportType.value || undefined
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
    }
  } catch (error) {
    ElMessage.error('获取举报列表失败')
  } finally {
    loading.value = false
  }
}

// 加载数据
const loadData = () => {
  if (activeTab.value === 'reports') {
    loadReportData()
  } else {
    loadAllContent()
  }
}

// 标签页切换
const handleTabChange = () => {
  currentPage.value = 1
  selectedItems.value = []
  loadData()
}

// 内容类型切换
const handleTypeChange = () => {
  currentPage.value = 1
  loadData()
}

// 举报状态切换
const handleReportStatusChange = () => {
  currentPage.value = 1
  loadReportData()
}

// 举报类型切换
const handleReportTypeChange = () => {
  currentPage.value = 1
  loadReportData()
}

// 分页改变
const handlePageChange = (page) => {
  currentPage.value = page
  loadData()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedItems.value = selection
}

// 恢复内容
const handleRestore = async (row) => {
  try {
    const res = await restoreContent(row.targetType, row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('恢复成功')
      loadData()
    }
  } catch (error) {
    ElMessage.error('恢复失败')
  }
}

// 下架内容
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要下架该内容吗？',
      '提示',
      { type: 'warning' }
    )
    const res = await deleteContent(row.targetType, row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('下架成功')
      loadData()
    }
  } catch {}
}

// 批量恢复
const handleBatchRestore = async () => {
  if (selectedItems.value.length === 0) {
    ElMessage.warning('请选择要恢复的内容')
    return
  }
  try {
    const items = selectedItems.value.map(item => ({
      type: item.targetType,
      id: item.id
    }))
    const res = await batchProcess({ action: 'restore', items })
    if (res.code === 200 || res.success) {
      ElMessage.success(`批量恢复完成：成功 ${res.data?.successCount || 0} 条`)
      loadData()
      selectedItems.value = []
    }
  } catch (error) {
    ElMessage.error('批量恢复失败')
  }
}

// 批量下架
const handleBatchDelete = async () => {
  if (selectedItems.value.length === 0) {
    ElMessage.warning('请选择要下架的内容')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要下架选中的 ${selectedItems.value.length} 条内容吗？`,
      '提示',
      { type: 'warning' }
    )
    const items = selectedItems.value.map(item => ({
      type: item.targetType,
      id: item.id
    }))
    const res = await batchProcess({ action: 'delete', items })
    if (res.code === 200 || res.success) {
      ElMessage.success(`批量下架完成：成功 ${res.data?.successCount || 0} 条`)
      loadData()
      selectedItems.value = []
    }
  } catch {}
}

// 举报处理
const openReportProcess = (row, action) => {
  currentReport.value = row
  reportAction.value = action
  reportAdminRemark.value = ''
  reportProcessDialog.value = true
}

const confirmReportProcess = async () => {
  if (!currentReport.value) return
  try {
    const res = await processReport(currentReport.value.id, {
      action: reportAction.value,
      adminRemark: reportAdminRemark.value
    })
    if (res.code === 200 || res.success) {
      ElMessage.success(reportAction.value === 'resolve' ? '已处理并下架内容' : '已驳回举报')
      reportProcessDialog.value = false
      loadReportData()
    }
  } catch (error) {
    ElMessage.error('处理失败')
  }
}

// 获取内容类型标签
const getContentTypeLabel = (type) => {
  const option = contentTypeOptions.find(opt => opt.value === type)
  return option ? option.label : type
}

// 获取举报类型标签
const getReportTypeLabel = (type) => {
  switch (type) {
    case 'COMMENT': return '评论'
    case 'REPLY': return '回复'
    case 'DYNAMIC_COMMENT': return '动态评论'
    case 'MANUSCRIPT': return '稿件'
    default: return type
  }
}

// 获取举报状态标签
const getReportStatusLabel = (status) => {
  switch (status) {
    case 'PENDING': return '待处理'
    case 'RESOLVED': return '已处理'
    case 'REJECTED': return '已驳回'
    default: return status
  }
}

const getReportStatusType = (status) => {
  switch (status) {
    case 'PENDING': return 'warning'
    case 'RESOLVED': return 'success'
    case 'REJECTED': return 'info'
    default: return ''
  }
}

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="content-review-page">
    <h2 class="page-title">内容审核中心</h2>

    <!-- 标签页 -->
    <el-tabs v-model="activeTab" @tab-change="handleTabChange" class="review-tabs">
      <el-tab-pane label="全部" name="all"></el-tab-pane>
      <el-tab-pane label="正常" name="normal"></el-tab-pane>
      <el-tab-pane label="已下架" name="removed"></el-tab-pane>
      <el-tab-pane label="举报管理" name="reports"></el-tab-pane>
    </el-tabs>

    <!-- 内容审核筛选 -->
    <div class="filter-bar" v-if="activeTab !== 'reports'">
      <el-select v-model="contentType" placeholder="内容类型" clearable @change="handleTypeChange" style="width: 150px">
        <el-option
          v-for="opt in contentTypeOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-button type="success" @click="handleBatchRestore" :disabled="selectedItems.length === 0">
        <el-icon><Check /></el-icon>
        批量恢复
      </el-button>
      <el-button type="danger" @click="handleBatchDelete" :disabled="selectedItems.length === 0">
        <el-icon><Delete /></el-icon>
        批量下架
      </el-button>
    </div>

    <!-- 举报筛选 -->
    <div class="filter-bar" v-if="activeTab === 'reports'">
      <el-select v-model="reportStatus" placeholder="处理状态" clearable @change="handleReportStatusChange" style="width: 150px">
        <el-option label="待处理" value="PENDING" />
        <el-option label="已处理" value="RESOLVED" />
        <el-option label="已驳回" value="REJECTED" />
      </el-select>
      <el-select v-model="reportType" placeholder="内容类型" clearable @change="handleReportTypeChange" style="width: 150px">
        <el-option label="评论" value="COMMENT" />
        <el-option label="回复" value="REPLY" />
        <el-option label="动态评论" value="DYNAMIC_COMMENT" />
      </el-select>
    </div>

    <!-- 内容审核表格 -->
    <el-table
      v-if="activeTab !== 'reports'"
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="row.targetType === 'COMMENT' ? 'primary' : 'info'">
            {{ getContentTypeLabel(row.targetType) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="用户" width="180">
        <template #default="{ row }">
          <div class="user-info">
            <el-avatar :size="32" :src="row.userAvatar">
              <el-icon><UserFilled /></el-icon>
            </el-avatar>
            <span class="username">{{ row.userName || '未知用户' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="内容" min-width="300">
        <template #default="{ row }">
          <div class="content-text">{{ row.content }}</div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'NORMAL' ? 'success' : 'danger'">
            {{ row.status === 'NORMAL' ? '正常' : '已下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="150">
        <template #default="{ row }">
          <el-button
            v-if="row.status === 'REMOVED'"
            link
            type="success"
            size="small"
            @click="handleRestore(row)"
          >
            恢复
          </el-button>
          <el-button v-if="row.status !== 'REMOVED'" link type="danger" size="small" @click="handleDelete(row)">
            下架
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 举报管理表格 -->
    <el-table
      v-if="activeTab === 'reports'"
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag type="info">{{ getReportTypeLabel(row.targetType) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="举报原因" width="120">
        <template #default="{ row }">
          <span>{{ row.reason }}</span>
        </template>
      </el-table-column>
      <el-table-column label="被举报内容" min-width="250">
        <template #default="{ row }">
          <div v-if="row.targetContent" class="content-text">{{ row.targetContent }}</div>
          <div v-else style="color: #909399;">(内容已删除或无法获取)</div>
          <div v-if="row.manuscriptId" style="font-size: 12px; color: #909399; margin-top: 4px;">
            稿件ID: {{ row.manuscriptId }}
            <el-button link type="primary" size="small" @click="window.open(`/video/${row.manuscriptId}`, '_blank')">查看稿件</el-button>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="补充说明" width="180">
        <template #default="{ row }">
          <span class="content-text">{{ row.description || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getReportStatusType(row.status)">
            {{ getReportStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="举报时间" width="170">
        <template #default="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="管理员备注" width="160">
        <template #default="{ row }">
          <span class="content-text">{{ row.adminRemark || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="160">
        <template #default="{ row }">
          <template v-if="row.status === 'PENDING'">
            <el-button link type="danger" size="small" @click="openReportProcess(row, 'resolve')">
              下架内容
            </el-button>
            <el-button link type="info" size="small" @click="openReportProcess(row, 'reject')">
              驳回
            </el-button>
          </template>
          <span v-else style="color: #909399; font-size: 12px;">已处理</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="loadData"
      />
    </div>

    <!-- 举报处理弹窗 -->
    <el-dialog
      v-model="reportProcessDialog"
      :title="reportAction === 'resolve' ? '确认下架' : '确认驳回'"
      width="420px"
    >
      <div style="padding: 10px 0;">
        <p v-if="reportAction === 'resolve'" style="margin-bottom: 16px; color: #606266;">
          确认下架该被举报内容？下架后举报人和内容作者都将收到系统通知。
        </p>
        <p v-else style="margin-bottom: 16px; color: #606266;">
          确认驳回该举报？驳回后举报人将收到系统通知。
        </p>
        <div style="font-size: 14px; color: #606266; margin-bottom: 8px;">管理员备注（选填）</div>
        <el-input
          v-model="reportAdminRemark"
          type="textarea"
          :rows="3"
          placeholder="输入备注信息..."
          maxlength="200"
          show-word-limit
        />
      </div>
      <template #footer>
        <el-button @click="reportProcessDialog = false">取消</el-button>
        <el-button :type="reportAction === 'resolve' ? 'danger' : 'primary'" @click="confirmReportProcess">
          {{ reportAction === 'resolve' ? '确认下架' : '确认驳回' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.content-review-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.review-tabs {
  margin-bottom: 20px;
}

.tab-badge {
  margin-left: 8px;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.username {
  font-size: 14px;
  color: #333;
}

.content-text {
  max-height: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  line-height: 1.5;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
