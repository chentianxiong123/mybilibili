<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getOverviewStatistics, getManuscriptStatusStatistics, getRecentManuscripts } from '../api/statistics'

const router = useRouter()

const statistics = ref({
  userCount: 0,
  manuscriptCount: 0,
  videoCount: 0,
  viewCount: 0,
  pendingManuscriptCount: 0
})

const manuscriptStatus = ref([])
const recentManuscripts = ref([])
const loading = ref(false)

const loadStatistics = async () => {
  loading.value = true
  try {
    const [overviewRes, statusRes, recentRes] = await Promise.all([
      getOverviewStatistics(),
      getManuscriptStatusStatistics(),
      getRecentManuscripts(5)
    ])
    
    if (overviewRes.code === 200 || overviewRes.success) {
      statistics.value = overviewRes.data
    }
    
    if (statusRes.code === 200 || statusRes.success) {
      manuscriptStatus.value = statusRes.data || []
    }
    
    if (recentRes.code === 200 || recentRes.success) {
      recentManuscripts.value = recentRes.data || []
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  } finally {
    loading.value = false
  }
}

const getStatusText = (status) => {
  const statusMap = {
    0: '待审核',
    1: '处理中',
    2: '待上架',
    3: '已上架',
    4: '审核拒绝',
    5: '处理失败',
    '-1': '已下架'
  }
  return statusMap[status] || '未知'
}

const getStatusType = (status) => {
  const typeMap = {
    0: 'warning',
    1: 'info',
    2: 'primary',
    3: 'success',
    4: 'danger',
    5: 'danger',
    '-1': 'info'
  }
  return typeMap[status] || ''
}

const goToManuscripts = () => {
  router.push('/manuscripts')
}

onMounted(() => {
  loadStatistics()
})
</script>

<template>
  <div class="dashboard">
    <h2 class="page-title">数据概览</h2>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon user">
          <el-icon :size="32"><User /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ statistics.userCount }}</div>
          <div class="stat-label">用户总数</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon manuscript">
          <el-icon :size="32"><Document /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ statistics.manuscriptCount }}</div>
          <div class="stat-label">稿件总数</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon video">
          <el-icon :size="32"><VideoCamera /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ statistics.videoCount }}</div>
          <div class="stat-label">视频总数</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon view">
          <el-icon :size="32"><View /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ statistics.viewCount?.toLocaleString() || 0 }}</div>
          <div class="stat-label">播放总数</div>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div class="panel">
        <h3 class="section-title">稿件状态分布</h3>
        <div class="status-list">
          <div 
            v-for="item in manuscriptStatus" 
            :key="item.status" 
            class="status-item"
          >
            <el-tag :type="getStatusType(item.status)" size="small">
              {{ getStatusText(item.status) }}
            </el-tag>
            <span class="status-count">{{ item.count }}</span>
          </div>
          <div v-if="manuscriptStatus.length === 0" class="empty-text">
            暂无数据
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h3 class="section-title">最近稿件</h3>
          <el-button link type="primary" @click="goToManuscripts">查看全部</el-button>
        </div>
        <div class="recent-list">
          <div 
            v-for="item in recentManuscripts" 
            :key="item.id" 
            class="recent-item"
          >
            <div class="recent-title">{{ item.title }}</div>
            <div class="recent-meta">
              <el-tag :type="getStatusType(item.status)" size="small">
                {{ getStatusText(item.status) }}
              </el-tag>
              <span class="recent-time">{{ item.uploadTime }}</span>
            </div>
          </div>
          <div v-if="recentManuscripts.length === 0" class="empty-text">
            暂无数据
          </div>
        </div>
      </div>
    </div>

    <div class="quick-actions">
      <h3 class="section-title">快捷操作</h3>
      <div class="action-grid">
        <div class="action-item" @click="$router.push('/manuscripts')">
          <el-icon :size="24" color="#409eff"><Document /></el-icon>
          <span>稿件管理</span>
          <el-badge v-if="statistics.pendingManuscriptCount > 0" 
            :value="statistics.pendingManuscriptCount" 
            class="action-badge" 
          />
        </div>
        <div class="action-item" @click="$router.push('/video-process')">
          <el-icon :size="24" color="#67c23a"><VideoCamera /></el-icon>
          <span>视频处理</span>
        </div>
        <div class="action-item" @click="$router.push('/comments')">
          <el-icon :size="24" color="#e6a23c"><ChatDotRound /></el-icon>
          <span>评论管理</span>
        </div>
        <div class="action-item" @click="$router.push('/categories')">
          <el-icon :size="24" color="#f56c6c"><Folder /></el-icon>
          <span>分类管理</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 20px;
}

.page-title {
  margin: 0 0 24px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 8px;
}

.stat-icon.user {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.stat-icon.manuscript {
  background: rgba(144, 147, 153, 0.1);
  color: #909399;
}

.stat-icon.video {
  background: rgba(103, 194, 58, 0.1);
  color: #67c23a;
}

.stat-icon.view {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.stat-content {
  margin-left: 16px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #333;
}

.stat-label {
  margin-top: 4px;
  font-size: 14px;
  color: #666;
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.panel {
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  margin: 0 0 16px;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.panel-header .section-title {
  margin: 0;
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.status-count {
  font-weight: 600;
  color: #333;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recent-item {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.recent-title {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recent-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.recent-time {
  font-size: 12px;
  color: #999;
}

.empty-text {
  text-align: center;
  color: #999;
  padding: 20px;
}

.quick-actions {
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.quick-actions .section-title {
  margin-bottom: 16px;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
}

.action-item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.action-item:hover {
  border-color: #409eff;
  background: rgba(64, 158, 255, 0.05);
}

.action-item span {
  margin-top: 12px;
  font-size: 14px;
  color: #333;
}

.action-badge {
  position: absolute;
  top: 10px;
  right: 10px;
}
</style>
