<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElDialog, ElDescriptions, ElDescriptionsItem, ElTag } from 'element-plus'
import { Search, Loading } from '@element-plus/icons-vue'
import { getCommentList, deleteComment, updateCommentStatus, getCommentById } from '../api/comment'

// 表格数据
const tableData = ref([])
const loading = ref(false)
const total = ref(0)

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

// 搜索
const keyword = ref('')
const videoIdFilter = ref('')

// 详情弹窗相关
const detailDialogVisible = ref(false)
const commentDetail = ref({})

// 加载评论列表
const loadComments = async () => {
  loading.value = true
  try {
    const res = await getCommentList({
      page: currentPage.value,
      size: pageSize.value,
      keyword: keyword.value,
      videoId: videoIdFilter.value ? parseInt(videoIdFilter.value) : undefined
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || res.data || []
      total.value = res.data?.total || res.total || 0
    }
  } catch (error) {
    ElMessage.error('获取评论列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadComments()
}

// 重置搜索
const handleReset = () => {
  keyword.value = ''
  videoIdFilter.value = ''
  currentPage.value = 1
  loadComments()
}

// 分页改变
const handlePageChange = (page) => {
  currentPage.value = page
  loadComments()
}

// 删除评论
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该评论吗？此操作不可恢复！',
      '警告',
      { type: 'error' }
    )
    const res = await deleteComment(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadComments()
    }
  } catch {}
}

// 更新评论状态
const handleStatusChange = async (row) => {
  // 保存原始状态，以防需要恢复
  const originalStatus = row.status
  const newStatus = originalStatus === 1 ? 0 : 1
  const statusText = newStatus === 1 ? '显示' : '隐藏'
  
  try {
    await ElMessageBox.confirm(
      `确定要将该评论设置为${statusText}状态吗？`,
      '提示',
      { type: 'warning' }
    )
    
    const res = await updateCommentStatus(row.id, newStatus)
    
    if (res.code === 200 || res.success) {
      ElMessage.success(`${statusText}成功`)
      // 更新本地状态
      row.status = newStatus
      // 重新加载列表以确保数据同步
      setTimeout(() => {
        loadComments()
      }, 100)
    } else {
      // 操作失败，恢复原来的状态
      row.status = originalStatus
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    // 用户取消操作或操作失败，恢复原来的状态
    row.status = originalStatus
    if (error !== 'cancel') {
      ElMessage.error('操作失败：' + (error.message || '未知错误'))
    }
  }
}

// 打开详情弹窗
const openDetailDialog = async (commentId) => {
  try {
    loading.value = true
    const res = await getCommentById(commentId)
    if (res.code === 200 || res.success) {
      commentDetail.value = res.data || {}
      detailDialogVisible.value = true
    } else {
      ElMessage.error('获取评论详情失败')
    }
  } catch (error) {
    ElMessage.error('获取评论详情失败')
  } finally {
    loading.value = false
  }
}

// 关闭详情弹窗
const closeDetailDialog = () => {
  detailDialogVisible.value = false
  commentDetail.value = {}
}

// 格式化状态显示
const formatStatus = (status) => {
  switch (status) {
    case 1: return '显示'
    case 0: return '隐藏'
    default: return '未知'
  }
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

onMounted(() => {
  loadComments()
})
</script>

<template>
  <div class="comments-page">
    <h2 class="page-title">评论管理</h2>

    <!-- 搜索区域 -->
    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索评论内容、用户名"
        clearable
        style="width: 250px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-input
        v-model="videoIdFilter"
        placeholder="视频ID"
        clearable
        style="width: 150px"
        @keyup.enter="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      stripe
    >
      <el-table-column prop="id" label="ID" width="70" align="center" />
      <el-table-column prop="content" label="评论内容" min-width="250" show-overflow-tooltip />
      <el-table-column label="用户信息" width="150">
        <template #default="{ row }">
          <div style="font-size: 13px">
            <div>ID: {{ row.userId }}</div>
            <div>{{ row.username || '未知用户' }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="视频信息" width="180">
        <template #default="{ row }">
          <div style="font-size: 13px">
            <div>ID: {{ row.videoId }}</div>
            <div>{{ row.videoTitle ? row.videoTitle.substring(0, 20) + '...' : '未知视频' }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="互动数据" width="120" align="center">
        <template #default="{ row }">
          <div style="font-size: 13px">
            <div>👍 {{ row.likeCount || 0 }}</div>
            <div>💬 {{ row.replyCount || 0 }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="0"
            :loading="loading"
            active-text="显示"
            inactive-text="隐藏"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="发布时间" width="150">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="180" align="center">
        <template #default="{ row }">
          <el-button
            type="primary"
            size="small"
            @click="openDetailDialog(row.id)"
          >
            详情
          </el-button>
          <el-button
            type="warning"
            size="small"
            @click="handleStatusChange(row)"
          >
            {{ row.status === 1 ? '隐藏' : '显示' }}
          </el-button>
          <el-button
            type="danger"
            size="small"
            @click="handleDelete(row)"
          >
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
        @size-change="loadComments"
      />
    </div>

    <!-- 评论详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="评论详细信息"
      width="700px"
      :before-close="closeDetailDialog"
    >
      <div v-if="commentDetail.id">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="评论ID">{{ commentDetail.id }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="commentDetail.status === 1 ? 'success' : 'info'">
              {{ formatStatus(commentDetail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ commentDetail.userId }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ commentDetail.username || '未知' }}</el-descriptions-item>
          <el-descriptions-item label="视频ID">{{ commentDetail.videoId }}</el-descriptions-item>
          <el-descriptions-item label="视频标题">{{ commentDetail.videoTitle || '未知' }}</el-descriptions-item>
          <el-descriptions-item label="点赞数">{{ commentDetail.likeCount || 0 }}</el-descriptions-item>
          <el-descriptions-item label="回复数">{{ commentDetail.replyCount || 0 }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(commentDetail.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(commentDetail.updatedAt) }}</el-descriptions-item>
        </el-descriptions>

        <div style="margin-top: 24px">
          <h4 style="margin-bottom: 12px">评论内容</h4>
          <div style="padding: 16px; background: #f5f7fa; border-radius: 4px; min-height: 100px">
            {{ commentDetail.content || '无内容' }}
          </div>
        </div>
      </div>
      <div v-else style="text-align: center; padding: 40px">
        <el-icon size="40" color="#909399"><Loading /></el-icon>
        <div style="margin-top: 16px; color: #909399">加载评论信息中...</div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="closeDetailDialog">关闭</el-button>
          <el-button
            type="warning"
            @click="handleStatusChange(commentDetail)"
          >
            {{ commentDetail.status === 1 ? '隐藏评论' : '显示评论' }}
          </el-button>
          <el-button
            type="danger"
            @click="handleDelete(commentDetail)"
          >
            删除评论
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.comments-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>