<template>
  <div class="collection-detail-view">
    <!-- 返回按钮和标题 -->
    <div class="collection-detail-header">
      <div class="header-left">
        <el-button text @click="emit('back')">
          <el-icon><ArrowRight style="transform: rotate(180deg)" /></el-icon>
          返回
        </el-button>
        <span class="header-title">我的合集和视频列表</span>
        <el-icon><ArrowRight /></el-icon>
        <span class="collection-name">{{ collection?.title || '加载中...' }}</span>
      </div>
      <div class="header-right">
        <el-button text @click="emit('edit')">
          <el-icon><Setting /></el-icon>
          编辑
        </el-button>
      </div>
    </div>

    <!-- 合集信息区 -->
    <div class="collection-info-section" v-if="collection">
      <div class="info-container">
        <!-- 封面 -->
        <div class="collection-cover-large">
          <img loading="lazy" decoding="async"
            :src="collection.coverUrl || getDefaultCover()"
            :alt="collection.title"
          />
          <div class="cover-overlay" @click="emit('play-all')">
            <el-button type="primary" size="large" :icon="VideoPlay">
              播放全部
            </el-button>
          </div>
          <div class="video-count-badge">
            <el-icon><VideoPlay /></el-icon>
            <span>{{ pagination.total }} 个视频</span>
          </div>
        </div>

        <!-- 信息 -->
        <div class="collection-meta-detail">
          <h1 class="collection-title-large">{{ collection.title }}</h1>
          <p class="collection-desc-large">{{ collection.description || '暂无描述' }}</p>

          <!-- 统计信息 -->
          <div class="stats-info">
            <span class="stat-item">
              <el-icon><Monitor /></el-icon>
              {{ collection.viewCount || 0 }} 次观看
            </span>
            <span class="stat-item">
              更新于 {{ formatDate(collection.updatedAt) }}
            </span>
            <span v-if="!collection.isPublic" class="private-tag">私密合集</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 视频网格区 -->
    <div class="collection-videos-section">
      <!-- 头部操作栏 -->
      <div class="videos-header">
        <div class="sort-options">
          <span
            class="sort-item"
            :class="{ active: sortBy === 'default' }"
            @click="emit('sort-change', 'default')"
          >
            默认排序
          </span>
          <span
            class="sort-item"
            :class="{ active: sortBy === 'newest' }"
            @click="emit('sort-change', 'newest')"
          >
            升序排序
          </span>
        </div>
        <div class="header-actions">
          <el-button type="primary" :icon="Setting" @click="emit('edit')">
            编辑
          </el-button>
          <el-dropdown trigger="click" @command="handleCollectionCommand">
            <el-button :icon="MoreFilled" circle />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="delete">删除视频列表</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>

      <!-- 空状态 -->
      <div v-else-if="manuscripts.length === 0" class="empty-state">
        <el-empty description="暂无视频" />
      </div>

      <!-- 视频网格 -->
      <div v-else class="videos-grid">
        <!-- 添加视频卡片 -->
        <div class="video-card add-video-card" @click="emit('add-video')">
          <div class="add-video-content">
            <el-icon :size="40"><Plus /></el-icon>
            <span>添加稿件</span>
          </div>
        </div>

        <!-- 视频卡片 -->
        <div
          v-for="(manuscript, index) in manuscripts"
          :key="manuscript.id"
          class="collection-video-card"
        >
          <!-- 封面区域 -->
          <div class="collection-video-cover" @click="emit('play-video', manuscript)">
            <img loading="lazy" decoding="async"
              :src="manuscript.coverUrl || getDefaultCover()"
              :alt="manuscript.title"
              class="collection-video-cover-img"
            />
            <!-- 序号 -->
            <div class="collection-video-index">{{ index + 1 }}</div>
            <!-- 时长 -->
            <div class="collection-video-duration">{{ manuscript.duration || '00:00' }}</div>
            <!-- 更多操作 -->
            <div class="collection-video-actions" @click.stop>
              <el-dropdown trigger="click" @command="(cmd) => handleVideoCommand(cmd, manuscript)">
                <el-button :icon="MoreFilled" circle size="small" />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="remove">从列表中移除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>

          <!-- 信息区域 -->
          <div class="collection-video-info">
            <h3 class="collection-video-title" :title="manuscript.title">{{ manuscript.title }}</h3>
            <div class="collection-video-meta">
              <span class="collection-meta-item">
                <el-icon><VideoPlay /></el-icon>
                {{ formatNumber(manuscript.viewCount) }}
              </span>
              <span class="collection-meta-item">{{ formatDate(manuscript.uploadTime) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  VideoPlay,
  Setting,
  Monitor,
  Plus,
  MoreFilled,
  ArrowRight
} from '@element-plus/icons-vue'

const props = defineProps({
  collection: {
    type: Object,
    default: null
  },
  manuscripts: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  sortBy: {
    type: String,
    default: 'default'
  },
  pagination: {
    type: Object,
    default: () => ({
      page: 1,
      size: 20,
      total: 0
    })
  }
})

const emit = defineEmits([
  'back',
  'edit',
  'play-all',
  'sort-change',
  'delete',
  'add-video',
  'play-video',
  'remove-video'
])

const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

const handleCollectionCommand = (command) => {
  if (command === 'delete') {
    emit('delete')
  }
}

const handleVideoCommand = (command, manuscript) => {
  if (command === 'remove') {
    emit('remove-video', manuscript)
  }
}
</script>

<style scoped>
.collection-detail-view {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
}

.collection-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e0e0e0;
}

.collection-detail-header .header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.collection-detail-header .header-title {
  font-size: 14px;
  color: #606266;
}

.collection-detail-header .collection-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.collection-info-section {
  margin-bottom: 24px;
}

.collection-info-section .info-container {
  display: flex;
  gap: 24px;
}

.collection-cover-large {
  width: 240px;
  height: 135px;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
}

.collection-cover-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-cover-large .cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
  cursor: pointer;
}

.collection-cover-large:hover .cover-overlay {
  opacity: 1;
}

.collection-cover-large .video-count-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-meta-detail {
  flex: 1;
}

.collection-title-large {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.collection-desc-large {
  font-size: 14px;
  color: #606266;
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.collection-meta-detail .stats-info {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #909399;
}

.collection-meta-detail .stats-info .stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-meta-detail .private-tag {
  background: #f4f4f5;
  color: #909399;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.collection-videos-section {
  margin-top: 24px;
}

.collection-videos-section .videos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-videos-section .sort-options {
  display: flex;
  gap: 16px;
}

.collection-videos-section .sort-item {
  font-size: 14px;
  color: #909399;
  cursor: pointer;
  padding: 4px 0;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.collection-videos-section .sort-item.active {
  color: #00a1d6;
  border-bottom-color: #00a1d6;
}

.collection-videos-section .sort-item:hover {
  color: #00a1d6;
}

.collection-videos-section .header-actions {
  display: flex;
  gap: 8px;
}

.collection-videos-section .videos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.collection-video-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}

.collection-video-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.collection-video-cover {
  position: relative;
  aspect-ratio: 16/9;
  cursor: pointer;
}

.collection-video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-index {
  position: absolute;
  top: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.collection-video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.collection-video-actions {
  position: absolute;
  top: 8px;
  right: 8px;
}

.collection-video-info {
  padding: 12px;
}

.collection-video-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
}

.collection-meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.add-video-card {
  aspect-ratio: 16/9;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
}

.add-video-card:hover {
  border-color: #00a1d6;
  background: #f0f9ff;
}

.add-video-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.add-video-card:hover .add-video-content {
  color: #00a1d6;
}

.loading-state {
  padding: 40px;
  text-align: center;
  color: #909399;
}

.empty-state {
  padding: 40px;
  text-align: center;
}
</style>