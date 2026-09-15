<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { getAdminComments, deleteAdminComment, restoreAdminComment } from '@/api/adminContent'

const tableData = ref([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const targetType = ref('comment')
const statusFilter = ref('')

const typeOptions = [
  { label: '评论', value: 'comment' },
  { label: '回复', value: 'reply' }
]

const statusOptions = [
  { label: '正常', value: '0' },
  { label: '已删除', value: '1' }
]

const replyStatusOptions = [
  { label: '正常', value: 'NORMAL' },
  { label: '已删除', value: 'REMOVED' }
]

const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const isRemoved = (row) => {
  return row.type === 'reply' ? row.status === 'REMOVED' : row.status === '1'
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getAdminComments({
      page: currentPage.value,
      size: pageSize.value,
      type: targetType.value,
      keyword: keyword.value || undefined,
      status: statusFilter.value || undefined
    })
    if (res.code === 200 || res.success) {
      tableData.value = (res.data?.list || []).map(item => ({ ...item, type: targetType.value }))
      total.value = Number(res.data?.total || 0)
    } else {
      ElMessage.error(res.message || '获取列表失败')
    }
  } catch (error) {
    console.error('获取列表失败:', error)
    ElMessage.error('获取列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadData()
}

const handleTypeChange = () => {
  statusFilter.value = ''
  currentPage.value = 1
  loadData()
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadData()
}

const handleSizeChange = () => {
  currentPage.value = 1
  loadData()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要下架该${row.type === 'reply' ? '回复' : '评论'}吗？下架后用户端将不可见。`,
      '提示',
      { type: 'warning' }
    )
    const res = await deleteAdminComment(row.type, row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('下架成功')
      loadData()
    } else {
      ElMessage.error(res.message || '下架失败')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('下架失败')
  }
}

const handleRestore = async (row) => {
  try {
    const res = await restoreAdminComment(row.type, row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('恢复成功')
      loadData()
    } else {
      ElMessage.error(res.message || '恢复失败')
    }
  } catch (error) {
    ElMessage.error('恢复失败')
  }
}

onMounted(loadData)
</script>

<template>
  <div class="comment-panel">
    <div class="filter-bar">
      <el-select v-model="targetType" style="width: 120px" @change="handleTypeChange">
        <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </el-select>
      <el-select v-model="statusFilter" placeholder="状态" clearable style="width: 130px" @change="handleSearch">
        <el-option
          v-for="opt in (targetType === 'reply' ? replyStatusOptions : statusOptions)"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-input
        v-model="keyword"
        placeholder="搜索内容关键词"
        clearable
        style="width: 240px"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button :icon="Refresh" @click="keyword = ''; statusFilter = ''; handleSearch()">重置</el-button>
    </div>

    <el-table v-loading="loading" :data="tableData" style="width: 100%">
      <el-table-column label="类型" width="80">
        <template #default="{ row }">
          <el-tag :type="row.type === 'reply' ? 'info' : 'primary'" size="small">
            {{ row.type === 'reply' ? '回复' : '评论' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="用户" width="150">
        <template #default="{ row }">
          <div class="user-info">
            <el-avatar :size="28" :src="row.userAvatar" />
            <span>{{ row.userName || '未知用户' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="内容" min-width="220">
        <template #default="{ row }">
          <div class="content-text">{{ row.content }}</div>
        </template>
      </el-table-column>
      <el-table-column label="回复对象" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.parentContent" class="parent-content">↳ {{ row.parentContent }}</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="manuscriptTitle" label="所属稿件" min-width="180" show-overflow-tooltip />
      <el-table-column prop="likeCount" label="点赞" width="70" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="isRemoved(row) ? 'danger' : 'success'" size="small">
            {{ isRemoved(row) ? '已删除' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时间" width="160">
        <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="130">
        <template #default="{ row }">
          <el-button v-if="!isRemoved(row)" link type="danger" size="small" @click="handleDelete(row)">下架</el-button>
          <el-button v-else link type="success" size="small" @click="handleRestore(row)">恢复</el-button>
        </template>
      </el-table-column>
      <template #empty>暂无数据</template>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.content-text {
  max-height: 48px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.parent-content {
  color: #909399;
  font-size: 12px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
