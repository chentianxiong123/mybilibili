<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh, View, Check, Close } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getArticleReviewList,
  getPendingArticles,
  getArticleReviewStatistics,
  reviewArticle,
  getArticleDetail
} from '@/api/articleReview'

const activeTab = ref('pending')
const keyword = ref('')
const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const size = ref(10)
const stats = ref({ total: 0, pending: 0, approved: 0, rejected: 0, draft: 0 })

const previewVisible = ref(false)
const previewLoading = ref(false)
const preview = ref(null)

const statusText = (row) => {
  if (row.status === 0) return '草稿'
  if (row.status === 2 || row.review_status === 2) return '审核拒绝'
  if (row.status === 1 && row.review_status === 0) return '待审核'
  if (row.status === 1 && row.review_status === 1) return '已发布'
  return '未知'
}

const statusType = (row) => {
  const text = statusText(row)
  if (text === '待审核') return 'warning'
  if (text === '已发布') return 'success'
  if (text === '审核拒绝') return 'danger'
  return 'info'
}

const formattedTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const loadStatistics = async () => {
  try {
    const res = await getArticleReviewStatistics()
    if (res.code === 200 || res.success) {
      stats.value = { ...stats.value, ...(res.data || {}) }
    }
  } catch {
    // 统计失败不影响列表
  }
}

const loadList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, size: size.value, keyword: keyword.value.trim() || undefined }
    const res = activeTab.value === 'pending'
      ? await getPendingArticles(params)
      : await getArticleReviewList({ ...params, status: '1' })
    if (res.code === 200 || res.success) {
      list.value = res.data?.list || []
      total.value = res.data?.total || 0
    } else {
      ElMessage.error(res.message || '加载失败')
    }
  } catch {
    ElMessage.error('加载专栏列表失败')
  } finally {
    loading.value = false
  }
}

const reload = () => {
  page.value = 1
  loadList()
  loadStatistics()
}

const handleTabChange = () => {
  page.value = 1
  loadList()
}

const openPreview = async (row) => {
  previewVisible.value = true
  previewLoading.value = true
  preview.value = { ...row, content: '' }
  try {
    const res = await getArticleDetail(row.id)
    if (res.code === 200 || res.success) {
      preview.value = { ...row, ...(res.data || {}) }
    }
  } catch {
    ElMessage.error('加载专栏正文失败')
  } finally {
    previewLoading.value = false
  }
}

const handleApprove = async (row) => {
  try {
    await ElMessageBox.confirm(`确定审核通过《${row.title}》吗？通过后将立即在专栏区展示。`, '审核通过', {
      type: 'warning',
      confirmButtonText: '通过',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const res = await reviewArticle(row.id, true)
    if (res.code === 200 || res.success) {
      ElMessage.success('审核通过')
      loadList()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleReject = async (row) => {
  let reason = ''
  try {
    const { value } = await ElMessageBox.prompt('请输入拒绝原因（作者可见）', '审核拒绝', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputPattern: /.+/,
      inputErrorMessage: '请填写拒绝原因'
    })
    reason = value
  } catch {
    return
  }
  try {
    const res = await reviewArticle(row.id, false, reason)
    if (res.code === 200 || res.success) {
      ElMessage.success('已拒绝')
      loadList()
      loadStatistics()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

const statCards = computed(() => [
  { label: '待审核', value: stats.value.pending, type: 'pending' },
  { label: '已通过', value: stats.value.approved, type: 'approved' },
  { label: '已拒绝', value: stats.value.rejected, type: 'rejected' },
  { label: '草稿', value: stats.value.draft, type: 'draft' }
])

onMounted(() => {
  loadList()
  loadStatistics()
})
</script>

<template>
  <div class="article-review">
    <h2 class="page-title">专栏审核</h2>

    <div class="stat-cards">
      <div v-for="card in statCards" :key="card.type" class="stat-card" :class="`stat-card--${card.type}`">
        <span class="stat-value">{{ card.value }}</span>
        <span class="stat-label">{{ card.label }}</span>
      </div>
    </div>

    <div class="toolbar">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="待审核" name="pending" />
        <el-tab-pane label="已发布" name="published" />
      </el-tabs>
      <div class="toolbar-right">
        <el-input
          v-model="keyword"
          placeholder="搜索专栏标题"
          clearable
          class="search-input"
          @keyup.enter="reload"
          @clear="reload"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button :icon="Refresh" @click="reload">刷新</el-button>
      </div>
    </div>

    <el-table :data="list" v-loading="loading" border stripe class="review-table">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="标题" min-width="220">
        <template #default="{ row }">
          <a class="title-link" @click="openPreview(row)">{{ row.title || '（无标题）' }}</a>
        </template>
      </el-table-column>
      <el-table-column prop="author_name" label="作者" width="120" />
      <el-table-column prop="word_count" label="字数" width="80" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row)" size="small">{{ statusText(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="150">
        <template #default="{ row }">{{ formattedTime(row.updated_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :icon="View" @click="openPreview(row)">预览</el-button>
          <template v-if="row.status === 1 && row.review_status === 0">
            <el-button link type="success" :icon="Check" @click="handleApprove(row)">通过</el-button>
            <el-button link type="danger" :icon="Close" @click="handleReject(row)">拒绝</el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        :page-size="size"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="loadList"
      />
    </div>

    <el-dialog v-model="previewVisible" title="专栏预览" width="820px">
      <div v-loading="previewLoading" class="preview">
        <h3 class="preview-title">{{ preview?.title }}</h3>
        <div class="preview-meta">
          <span>作者：{{ preview?.author_name || '-' }}</span>
          <span>字数：{{ preview?.word_count ?? '-' }}</span>
          <span>状态：{{ preview ? statusText(preview) : '-' }}</span>
        </div>
        <div v-if="preview?.review_reason" class="preview-reason">
          拒绝原因：{{ preview.review_reason }}
        </div>
        <div class="preview-content" v-html="preview?.content || '<p>（无正文）</p>'" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.article-review { padding: 20px; }
.page-title { margin: 0 0 16px; font-size: 20px; font-weight: 600; }
.stat-cards { display: flex; gap: 16px; margin-bottom: 16px; }
.stat-card {
  flex: 1; display: flex; flex-direction: column; gap: 4px;
  padding: 16px; border-radius: 8px; background: #fff; border: 1px solid #ebeef5;
}
.stat-value { font-size: 26px; font-weight: 700; color: #303133; }
.stat-label { font-size: 13px; color: #909399; }
.stat-card--pending .stat-value { color: #e6a23c; }
.stat-card--approved .stat-value { color: #67c23a; }
.stat-card--rejected .stat-value { color: #f56c6c; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.toolbar-right { display: flex; align-items: center; gap: 10px; }
.search-input { width: 240px; }
.title-link { color: #409eff; cursor: pointer; }
.title-link:hover { text-decoration: underline; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }
.preview-title { margin: 0 0 8px; font-size: 18px; }
.preview-meta { display: flex; gap: 16px; font-size: 13px; color: #909399; margin-bottom: 12px; }
.preview-reason { padding: 8px 12px; margin-bottom: 12px; border-radius: 6px; background: #fef0f0; color: #f56c6c; }
.preview-content { line-height: 1.8; color: #303133; max-height: 460px; overflow: auto; }
</style>
