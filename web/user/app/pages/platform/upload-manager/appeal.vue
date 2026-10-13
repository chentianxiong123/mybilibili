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
          <div class="card-top">
            <div class="card-title-block">
              <p class="card-title">{{ item.title || '（已删除稿件）' }}</p>
              <p class="card-meta">稿件状态： {{ item.manuscriptStatusText }}</p>
            </div>
            <div class="card-cover">
              <img
                v-if="item.coverUrl"
                :src="item.coverUrl"
                :alt="item.title"
              />
              <span v-else class="cover-fallback">暂无封面</span>
            </div>
          </div>
          <div class="card-bottom">
            <div class="bottom-info">
              <p class="card-meta">申诉时间：{{ item.appealTime }}</p>
              <p class="card-meta">申诉编号：{{ item.appealNo }}</p>
              <p class="card-meta">受理状态：{{ item.statusText }}</p>
            </div>
            <span class="detail-btn">申诉详情</span>
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

/* 申诉卡片：对齐B站 双列浮动卡片（460px + 右间距20px，阴影细边框圆角4px） */
.appeal-list {
  margin-bottom: 20px;
}

.appeal-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 0 20px;
}

.appeal-card {
  width: 460px;
  max-width: 100%;
  background: #fff;
  border-radius: 4px;
  box-shadow: #dfe5ed 0 0 1px 1px;
  padding: 20px 0;
  margin-bottom: 20px;
  cursor: pointer;
  transition: box-shadow .2s;
}

.appeal-card:hover {
  box-shadow: rgba(255, 102, 153, .4) 0 0 1px 1px;
}

.appeal-card p {
  margin: 0;
}

/* 上段：标题 + 封面 */
.card-top {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 0 20px;
}

.card-title-block {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 14px;
  color: #00a1d6;
  margin-bottom: 10px !important;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-meta {
  font-size: 12px;
  color: #99a2aa;
  line-height: 1.8;
}

.card-cover {
  width: 80px;
  height: 50px;
  flex-shrink: 0;
  border-radius: 2px;
  overflow: hidden;
  background: #e5e9ef;
}

.card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  font-size: 10px;
  color: #99a2aa;
}

/* 下段：时间/编号/状态 + 描边按钮 */
.card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid #e5e9ef;
  margin-top: 18px;
  padding: 18px 20px 0;
}

.detail-btn {
  flex-shrink: 0;
  height: 28px;
  line-height: 28px;
  padding: 0 10px;
  font-size: 12px;
  color: #18191c;
  background: #fff;
  border: 1px solid #ccd0d7;
  border-radius: 4px;
  white-space: nowrap;
  transition: all .2s;
}

.appeal-card:hover .detail-btn {
  color: #ff6699;
  border-color: #ff6699;
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
