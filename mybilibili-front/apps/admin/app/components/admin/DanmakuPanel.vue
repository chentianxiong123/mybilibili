<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { getAdminDanmaku, deleteAdminDanmaku } from '@/api/adminContent'

const tableData = ref([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const keyword = ref('')

const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const formatTime = (t) => {
  const sec = Math.floor(t || 0)
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getAdminDanmaku({
      page: currentPage.value,
      size: pageSize.value,
      keyword: keyword.value || undefined
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || []
      total.value = Number(res.data?.total || 0)
    } else {
      ElMessage.error(res.message || '获取弹幕列表失败')
    }
  } catch (error) {
    console.error('获取弹幕列表失败:', error)
    ElMessage.error('获取弹幕列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
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
      `确定要删除该弹幕吗？删除后不可恢复。`,
      '提示',
      { type: 'warning' }
    )
    const res = await deleteAdminDanmaku(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadData()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(loadData)
</script>

<template>
  <div class="danmaku-panel">
    <div class="filter-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索弹幕内容"
        clearable
        style="width: 240px"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button :icon="Refresh" @click="keyword = ''; handleSearch()">重置</el-button>
    </div>

    <el-table v-loading="loading" :data="tableData" style="width: 100%">
      <el-table-column label="弹幕内容" min-width="240">
        <template #default="{ row }">
          <span class="content-text">{{ row.content }}</span>
        </template>
      </el-table-column>
      <el-table-column label="发送者" width="160">
        <template #default="{ row }">
          <div class="user-info">
            <el-avatar :size="28" :src="row.userAvatar" />
            <span>{{ row.userName || '未知用户' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="所属稿件" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">{{ row.manuscriptTitle }}</template>
      </el-table-column>
      <el-table-column label="视频时间" width="90">
        <template #default="{ row }">{{ formatTime(row.time) }}</template>
      </el-table-column>
      <el-table-column label="发送时间" width="160">
        <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="90">
        <template #default="{ row }">
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>暂无弹幕数据</template>
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

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
