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

// 内容类型选项
const contentTypeOptions = [
  { label: '全部', value: '' },
  { label: '评论', value: 'COMMENT' },
  { label: '回复', value: 'REPLY' },
  { label: '动态评论', value: 'DYNAMIC_COMMENT' }
]

// 加载待审核列表
const loadPendingList = async () => {
  loading.value = true
  try {
    const res = await getPendingList({
      page: currentPage.value,
      size: pageSize.value,
      contentType: contentType.value || undefined
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || res.data || []
      total.value = res.data?.total || res.total || 0
    }
  } catch (error) {
    ElMessage.error('获取待审核列表失败')
  } finally {
    loading.value = false
  }
}

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

// 加载数据
const loadData = () => {
  loadAllContent()
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

// 删除内容
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该内容吗？此操作不可恢复！',
      '警告',
      { type: 'error' }
    )
    const res = await deleteContent(row.targetType, row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
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

// 批量删除
const handleBatchDelete = async () => {
  if (selectedItems.value.length === 0) {
    ElMessage.warning('请选择要删除的内容')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedItems.value.length} 条内容吗？此操作不可恢复！`,
      '警告',
      { type: 'error' }
    )
    const items = selectedItems.value.map(item => ({
      type: item.targetType,
      id: item.id
    }))
    const res = await batchProcess({ action: 'delete', items })
    if (res.code === 200 || res.success) {
      ElMessage.success(`批量删除完成：成功 ${res.data?.successCount || 0} 条`)
      loadData()
      selectedItems.value = []
    }
  } catch {}
}

// 获取内容类型标签
const getContentTypeLabel = (type) => {
  const option = contentTypeOptions.find(opt => opt.value === type)
  return option ? option.label : type
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
    </el-tabs>

    <!-- 筛选和操作区域 -->
    <div class="filter-bar">
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
        批量删除
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table
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
          <el-button link type="danger" size="small" @click="handleDelete(row)">
            删除
          </el-button>
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
