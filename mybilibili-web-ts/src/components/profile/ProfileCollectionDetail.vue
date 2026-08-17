<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, ArrowRight, Edit, Plus, VideoPlay, MoreFilled, Clock } from '@element-plus/icons-vue'

const props = defineProps({
  collectionDetail: {
    type: Object,
    default: () => ({})
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['back', 'edit', 'add-video', 'sort-change', 'play-manuscript', 'play-all', 'remove-video', 'delete-collection'])

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 格式化数字
const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

// 获取默认封面
const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}

// 处理合集操作命令
const handleCollectionCommand = (command) => {
  if (command === 'delete') {
    ElMessageBox.confirm('确定要删除这个视频列表吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      emit('delete-collection', props.collectionDetail.collectionId)
    }).catch(() => {})
  }
}

// 处理视频操作命令
const handleVideoCommand = (command, manuscript) => {
  if (command === 'remove') {
    ElMessageBox.confirm('确定要从列表中移除这个视频吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      emit('remove-video', manuscript.id)
    }).catch(() => {})
  }
}
</script>

<template>
  <div class="collection-detail-view">
    <!-- 返回按钮和标题 -->
    <div class="collection-detail-header">
      <div class="header-left">
        <el-button text :icon="ArrowLeft" @click="emit('back')">
          返回
        </el-button>
        <span class="header-title">我的合集和视频列表</span>
        <el-icon><ArrowRight /></el-icon>
        <span class="collection-name">{{ collectionDetail.collection?.title || '加载中...' }}</span>
      </div>
      <div class="header-right" v-if="isOwnSpace">
        <el-button text :icon="Edit" @click="emit('edit')">
          编辑
        </el-button>
      </div>
    </div>

    <!-- 合集信息区 -->
    <div class="collection-info-section" v-if="collectionDetail.collection">
      <div class="info-container">
        <!-- 封面 -->
        <div class="collection-cover-large">
          <img loading="lazy" decoding="async"
            :src="collectionDetail.collection.coverUrl || getDefaultCover()"
            :alt="collectionDetail.collection.title"
          />
          <div class="cover-overlay" @click="emit('play-all')">
            <el-button type="primary" size="large" :icon="VideoPlay">
              播放全部
            </el-button>
          </div>
          <div class="video-count-badge">
            <el-icon><VideoPlay /></el-icon>
            <span>{{ collectionDetail.pagination.total }} 个视频</span>
          </div>
        </div>

        <!-- 信息 -->
        <div class="collection-meta-detail">
          <h1 class="collection-title-large">{{ collectionDetail.collection.title }}</h1>
          <p class="collection-desc-large">{{ collectionDetail.collection.description || '暂无描述' }}</p>

          <!-- 统计信息 -->
          <div class="stats-info">
            <span class="stat-item">
              <el-icon><Clock /></el-icon>
              更新于 {{ formatDate(collectionDetail.collection.updatedAt) }}
            </span>
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
            :class="{ active: collectionDetail.sortBy === 'default' }"
            @click="emit('sort-change', 'default')"
          >
            默认排序
          </span>
          <span
            class="sort-item"
            :class="{ active: collectionDetail.sortBy === 'newest' }"
            @click="emit('sort-change', 'newest')"
          >
            升序排序
          </span>
        </div>
        <div class="header-actions" v-if="isOwnSpace">
          <el-button type="primary" :icon="Edit" @click="emit('edit')">
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
      <div v-if="collectionDetail.loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>

      <!-- 空状态 -->
      <div v-else-if="collectionDetail.manuscripts.length === 0 && !isOwnSpace" class="empty-state">
        <el-empty description="暂无视频" />
      </div>

      <!-- 视频网格 -->
      <div v-else class="videos-grid">
        <!-- 添加视频卡片 -->
        <div v-if="isOwnSpace" class="video-card add-video-card" @click="emit('add-video')">
          <div class="add-video-content">
            <el-icon :size="40"><Plus /></el-icon>
            <span>添加视频</span>
          </div>
        </div>

        <!-- 视频卡片 -->
        <div
          v-for="(manuscript, index) in collectionDetail.manuscripts"
          :key="manuscript.id"
          class="collection-video-card"
        >
          <!-- 封面区域 -->
          <div class="collection-video-cover" @click="emit('play-manuscript', manuscript)">
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
            <div class="collection-video-actions" v-if="isOwnSpace" @click.stop>
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

<style scoped>
/* 合集详情视图样式 */
.collection-detail-view {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.collection-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.collection-detail-header .header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #666;
}

.collection-detail-header .header-title {
  color: #333;
  font-weight: 500;
}

.collection-detail-header .collection-name {
  color: #333;
  font-weight: 600;
}

/* 合集信息区 */
.collection-info-section {
  margin-bottom: 20px;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
}

.collection-info-section .info-container {
  display: flex;
  gap: 20px;
}

.collection-cover-large {
  position: relative;
  width: 280px;
  height: 158px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
  cursor: pointer;
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
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.4);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.collection-cover-large:hover .cover-overlay {
  opacity: 1;
}

.collection-cover-large .video-count-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 13px;
  border-radius: 4px;
}

/* 合集元信息 */
.collection-meta-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.collection-title-large {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0;
  line-height: 1.4;
}

.collection-desc-large {
  font-size: 14px;
  color: #666;
  margin: 0;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.collection-meta-detail .stats-info {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #9499a0;
}

.collection-meta-detail .stats-info .stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 视频网格区 */
.collection-videos-section {
  padding: 20px;
  background-color: #fff;
  border-radius: 8px;
}

.collection-videos-section .videos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
}

.collection-videos-section .sort-options {
  display: flex;
  gap: 24px;
}

.collection-videos-section .sort-item {
  font-size: 14px;
  color: #666;
  cursor: pointer;
  padding-bottom: 4px;
  border-bottom: 2px solid transparent;
  transition: all 0.3s ease;
}

.collection-videos-section .sort-item:hover {
  color: #00aeec;
}

.collection-videos-section .sort-item.active {
  color: #00aeec;
  border-bottom-color: #00aeec;
}

.collection-videos-section .header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 视频网格 */
.collection-videos-section .videos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
}

/* 合集详情页视频卡片 */
.collection-video-card {
  display: flex;
  flex-direction: column;
  background-color: #fff;
  border-radius: 8px;
  overflow: hidden;
  transition: transform 0.3s ease;
}

.collection-video-card:hover {
  transform: translateY(-4px);
}

.collection-video-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  cursor: pointer;
  background-color: #f0f0f0;
  border-radius: 8px 8px 0 0;
}

.collection-video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.collection-video-card:hover .collection-video-cover-img {
  transform: scale(1.05);
}

.collection-video-index {
  position: absolute;
  top: 8px;
  left: 8px;
  width: 24px;
  height: 24px;
  background-color: rgba(0, 0, 0, 0.6);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  z-index: 10;
}

.collection-video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  padding: 2px 6px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  border-radius: 2px;
  z-index: 1;
}

.collection-video-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 10;
}

.collection-video-card:hover .collection-video-actions {
  opacity: 1;
}

.collection-video-info {
  padding: 10px;
  flex: 1;
  min-height: 0;
}

.collection-video-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0 0 8px 0;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #9499a0;
  flex-wrap: wrap;
}

.collection-meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.video-card.add-video-card {
  aspect-ratio: 16 / 10;
  border: 2px dashed #dcdcdc;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-card.add-video-card:hover {
  border-color: #00aeec;
  background-color: #f0f9ff;
}

/* 加载状态和空状态 */
.loading-state,
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
  color: #9499a0;
  font-size: 14px;
}

/* 响应式设计 */
@media (max-width: 992px) {
  .collection-info-section .info-container {
    flex-direction: column;
  }

  .collection-cover-large {
    width: 100%;
    height: auto;
    aspect-ratio: 16 / 9;
  }
}

@media (max-width: 768px) {
  .collection-detail-header .header-left {
    flex-wrap: wrap;
  }
}
</style>