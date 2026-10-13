<template>
  <div class="appeal-manager">
    <!-- 状态筛选：对齐B站申诉页 全部/进行中/已完成 -->
    <div class="appeal-tabs">
      <span
        class="appeal-tab"
        :class="{ active: statusFilter === 'all' }"
        @click="setStatus('all')"
      >
        全部({{ totalCount }})
      </span>
      <span
        class="appeal-tab"
        :class="{ active: statusFilter === 'processing' }"
        @click="setStatus('processing')"
      >
        进行中({{ processingCount }})
      </span>
      <span
        class="appeal-tab"
        :class="{ active: statusFilter === 'done' }"
        @click="setStatus('done')"
      >
        已完成({{ doneCount }})
      </span>
    </div>

    <!-- 申诉列表 -->
    <div class="appeal-list" v-loading="loading">
      <div v-if="appeals.length" class="appeal-cards">
        <div
          v-for="item in appeals"
          :key="item.id"
          class="appeal-card"
          @click="openDetail(item.id)"
        >
          <div class="card-main">
            <p class="card-title">{{ item.title || '（已删除稿件）' }}</p>
            <p class="card-meta">稿件状态： {{ item.manuscriptStatusText }}</p>
          </div>
          <div class="card-sub">
            <div class="sub-info">
              <p class="card-meta">申诉时间：{{ item.appealTime }}</p>
              <p class="card-meta">申诉编号：{{ item.appealNo }}</p>
              <p class="card-meta">受理状态：{{ item.statusText }}</p>
            </div>
            <span class="detail-link">申诉详情</span>
          </div>
        </div>
      </div>
      <el-empty
        v-else-if="!loading"
        description="暂无申诉"
      />
    </div>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        layout="prev, pager, next, jumper, total"
        :total="totalAppeals"
        @current-change="handleCurrentChange"
      ></el-pagination>
    </div>

    <!-- 申诉详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="申诉详情"
      width="520px"
      class="appeal-detail-dialog"
    >
      <div v-if="detail" class="detail-body">
        <div class="detail-row">
          <span class="detail-label">申诉编号</span>
          <span class="detail-value">{{ detail.appealNo }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">稿件标题</span>
          <span class="detail-value">{{ detail.title || '（已删除稿件）' }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">稿件状态</span>
          <span class="detail-value">{{ detail.manuscriptStatusText }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">申诉原因</span>
          <span class="detail-value">{{ detail.reason }}</span>
        </div>
        <div class="detail-row" v-if="detail.description">
          <span class="detail-label">申诉说明</span>
          <span class="detail-value">{{ detail.description }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">申诉时间</span>
          <span class="detail-value">{{ detail.appealTime }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">受理状态</span>
          <span class="detail-value">{{ detail.statusText }}</span>
        </div>
        <div class="detail-row" v-if="detail.adminRemark">
          <span class="detail-label">处理备注</span>
          <span class="detail-value">{{ detail.adminRemark }}</span>
        </div>
        <div class="detail-row" v-if="detail.processedTime">
          <span class="detail-label">处理时间</span>
          <span class="detail-value">{{ detail.processedTime }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { manuscriptApi } from '@/api/creator'

const statusFilter = ref('all')

const appeals = ref([])
const loading = ref(false)
const totalAppeals = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

const totalCount = ref(0)
const processingCount = ref(0)
const doneCount = ref(0)

const detailVisible = ref(false)
const detail = ref(null)

const setStatus = (status) => {
  if (statusFilter.value === status) {
    return
  }
  statusFilter.value = status
  currentPage.value = 1
  fetchAppeals()
}

const fetchAppeals = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      size: pageSize.value
    }
    if (statusFilter.value && statusFilter.value !== 'all') {
      params.status = statusFilter.value
    }
    const response = await manuscriptApi.getMyAppeals(params)
    if (response.code === 200) {
      appeals.value = response.data.list || []
      totalAppeals.value = response.data.total || 0
    }
  } catch (error) {
    console.error('获取申诉列表失败:', error)
    ElMessage.error('获取申诉列表失败')
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await manuscriptApi.getMyAppealStats()
    if (response.code === 200) {
      totalCount.value = response.data.total || 0
      processingCount.value = response.data.processing || 0
      doneCount.value = response.data.done || 0
    }
  } catch (error) {
    console.error('获取申诉统计失败:', error)
  }
}

const openDetail = async (id) => {
  try {
    const response = await manuscriptApi.getAppealById(id)
    if (response.code === 200) {
      detail.value = response.data
      detailVisible.value = true
    }
  } catch (error) {
    console.error('获取申诉详情失败:', error)
    ElMessage.error('获取申诉详情失败')
  }
}

const handleCurrentChange = (current) => {
  currentPage.value = current
}

watch(currentPage, () => {
  fetchAppeals()
})

onMounted(() => {
  fetchAppeals()
  fetchStats()
})
</script>

<style scoped>
.appeal-manager {
  width: 100%;
}

/* 状态tab：对齐B站 全部/进行中/已完成（32px间距，选中蓝色加粗） */
.appeal-tabs {
  display: flex;
  align-items: center;
  gap: 32px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.appeal-tab {
  font-size: 14px;
  color: #505050;
  cursor: pointer;
  transition: color .2s;
  white-space: nowrap;
}

.appeal-tab:hover {
  color: #ff6699;
}

.appeal-tab.active {
  color: #ff6699;
  font-weight: 500;
}

/* 申诉卡片：对齐B站 padding 20px 0 通栏 */
.appeal-list {
  margin-bottom: 20px;
}

.appeal-cards {
  display: flex;
  flex-direction: column;
}

.appeal-card {
  padding: 20px 0;
  border-bottom: 1px solid #f0f0f0;
  background: #fff;
  cursor: pointer;
  transition: background .2s;
}

.appeal-card:hover {
  background: #fafafa;
}

.appeal-card p {
  margin: 0;
}

.card-main {
  margin-bottom: 12px;
}

.card-title {
  font-size: 14px;
  color: #4fbdea;
  margin-bottom: 8px !important;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-meta {
  font-size: 12px;
  color: #99a2aa;
  line-height: 1.8;
}

.card-sub {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
}

.detail-link {
  font-size: 12px;
  color: #18191c;
  white-space: nowrap;
  flex-shrink: 0;
}

.appeal-card:hover .detail-link {
  color: #ff6699;
}

/* 详情弹窗 */
.detail-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: flex;
  gap: 12px;
  font-size: 14px;
  line-height: 1.6;
}

.detail-label {
  flex-shrink: 0;
  width: 72px;
  color: #99a2aa;
}

.detail-value {
  flex: 1;
  color: #18191c;
  word-break: break-all;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 20px;
  padding: 15px 0;
  border-top: 1px solid #e0e0e0;
}
</style>
