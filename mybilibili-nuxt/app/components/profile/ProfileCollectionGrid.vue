<script setup>
import { useRouter } from 'vue-router'
import { Plus, VideoPlay, Grid, List, View, ArrowRight } from '@element-plus/icons-vue'
import { ElButton } from 'element-plus'

const props = defineProps({
  collections: {
    type: Object,
    default: () => ({
      viewType: 'horizontal',
      items: [],
      loading: false
    })
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['create-collection', 'view-detail', 'add-video', 'play-all', 'view-change'])

const router = useRouter()

// 获取默认封面
const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}
</script>

<template>
  <div>
    <!-- 顶部标题和新建按钮 -->
    <div class="collections-header">
      <div class="collections-header-left">
        <h2 class="collections-title">我的合集和视频列表</h2>
        <el-button
          v-if="isOwnSpace"
          type="primary"
          :icon="Plus"
          class="new-collection-btn"
          @click="emit('create-collection')"
        >
          新建
        </el-button>
      </div>
      <!-- 视图切换按钮 -->
      <div class="view-toggle">
        <button
          class="view-toggle-btn grid-view-btn"
          :class="{ active: collections.viewType === 'grid' }"
          @click="emit('view-change', 'grid')"
        >
          <el-icon><Grid /></el-icon>
        </button>
        <button
          class="view-toggle-btn list-view-btn"
          :class="{ active: collections.viewType === 'horizontal' }"
          @click="emit('view-change', 'horizontal')"
        >
          <el-icon><List /></el-icon>
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="collections.loading" class="loading-state">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 空状态 -->
    <div v-else-if="collections.items.length === 0" class="empty-state">
      <el-empty description="暂无合集">
        <el-button
          v-if="isOwnSpace"
          type="primary"
          @click="emit('create-collection')"
        >
          创建合集
        </el-button>
      </el-empty>
    </div>

    <!-- 宫格视图 -->
    <div v-else-if="collections.viewType === 'grid'" class="collections-grid">
      <!-- 新建合集卡片 -->
      <div
        v-if="isOwnSpace"
        class="collection-item new-collection"
        @click="emit('create-collection')"
      >
        <div class="new-collection-content">
          <el-icon :size="32"><Plus /></el-icon>
          <span>新建合集</span>
        </div>
      </div>

      <!-- 合集项 -->
      <div
        v-for="collection in collections.items"
        :key="collection.id"
        class="collection-item"
        @click="emit('view-detail', collection.id)"
      >
        <div class="collection-cover">
          <img loading="lazy" decoding="async" :src="collection.coverUrl || getDefaultCover()" :alt="collection.title" class="collection-cover-img">
          <div class="collection-video-count">
            <el-icon><VideoPlay /></el-icon>
            <span>{{ collection.manuscriptCount || 0 }}</span>
          </div>
        </div>
        <div class="collection-info">
          <div class="collection-title">{{ collection.title }}</div>
          <div class="collection-date">{{ collection.manuscriptCount || 0 }}个视频</div>
        </div>
      </div>
    </div>

    <!-- 水平列表视图 -->
    <div v-else-if="collections.viewType === 'horizontal'" class="collections-horizontal">
      <!-- 合集项 -->
      <div
        v-for="collection in collections.items"
        :key="collection.id"
        class="collection-horizontal-item"
      >
        <div class="collection-horizontal-header">
          <h3 class="collection-horizontal-title">
            {{ collection.title }}
            <span class="collection-video-count-badge">{{ collection.manuscriptCount || 0 }}</span>
          </h3>
          <div class="collection-horizontal-actions">
            <el-button
              v-if="collection.videos && collection.videos.length > 0"
              class="action-btn play-all-btn"
              :icon="VideoPlay"
              @click="emit('play-all', collection)"
            >
              播放全部
            </el-button>
            <el-button
              class="action-btn more-btn"
              @click="emit('view-detail', collection.id)"
            >
              更多
              <el-icon><ArrowRight /></el-icon>
            </el-button>
          </div>
        </div>

        <!-- 水平视频列表 -->
        <div class="collection-videos-horizontal">
          <!-- 添加视频按钮 -->
          <div
            v-if="isOwnSpace"
            class="add-video-card"
            @click="emit('add-video', collection.id)"
          >
            <div class="add-video-content">
              <el-icon :size="24"><Plus /></el-icon>
              <span>添加视频</span>
            </div>
          </div>

          <!-- 视频项 -->
          <div
            v-for="video in collection.videos"
            :key="video.id"
            class="video-horizontal-item"
            @click="router.push(`/manuscript/${video.id}`)"
          >
            <div class="video-horizontal-cover">
              <img loading="lazy" decoding="async" :src="video.coverUrl || getDefaultCover()" :alt="video.title" class="video-cover-img">
              <div class="video-duration">{{ video.duration }}</div>
            </div>
            <div class="video-horizontal-info">
              <div class="video-title" :title="video.title">{{ video.title }}</div>
              <div class="video-horizontal-meta">
                <span class="video-views">
                  <el-icon><View /></el-icon>
                  {{ video.viewCount || 0 }}
                </span>
                <span class="video-date">{{ video.date }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 顶部标题 */
.collections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.collections-header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collections-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.new-collection-btn {
  font-size: 14px;
}

/* 视图切换按钮样式 */
.view-toggle {
  display: flex;
  gap: 10px;
  align-items: center;
}

.view-toggle-btn {
  padding: 6px 10px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.3s ease;
}

.view-toggle-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.view-toggle-btn.active {
  background-color: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

/* 合集网格布局 */
.collections-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

/* 合集项样式 */
.collection-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.collection-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 新建合集按钮样式 */
.collection-item.new-collection {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 150px;
  border: 2px dashed #e0e0e0;
  border-radius: 8px;
  background-color: #fafafa;
}

.collection-item.new-collection:hover {
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.new-collection-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #999;
  font-size: 14px;
}

.collection-item.new-collection:hover .new-collection-content {
  color: #00aeec;
}

/* 合集封面样式 */
.collection-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.collection-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.collection-item:hover .collection-cover-img {
  transform: scale(1.05);
}

/* 合集视频数量样式 */
.collection-video-count {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  border-radius: 12px;
}

/* 合集信息样式 */
.collection-info {
  margin-top: 8px;
}

.collection-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-date {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* 水平列表视图样式 */
.collections-horizontal {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.collection-horizontal-item {
  border-bottom: 1px solid #f0f0f0;
  padding-bottom: 24px;
}

.collection-horizontal-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.collection-horizontal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-horizontal-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.collection-video-count-badge {
  font-size: 14px;
  color: #999;
  font-weight: normal;
}

.collection-horizontal-actions {
  display: flex;
  gap: 12px;
}

.collection-horizontal-actions .action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-horizontal-actions .play-all-btn {
  background-color: #fff;
  border: 1px solid #e0e0e0;
  color: #666;
}

.collection-horizontal-actions .play-all-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.collection-horizontal-actions .more-btn {
  background-color: #fff;
  border: 1px solid #e0e0e0;
  color: #666;
}

.collection-horizontal-actions .more-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

/* 水平视频列表样式 */
.collection-videos-horizontal {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 8px;
}

.collection-videos-horizontal::-webkit-scrollbar {
  height: 4px;
}

.collection-videos-horizontal::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 2px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 2px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}

/* 添加视频按钮 */
.add-video-card {
  width: 180px;
  height: 100px;
  border: 2px dashed #e0e0e0;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
  cursor: pointer;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.add-video-card:hover {
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.add-video-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #999;
  font-size: 14px;
}

.add-video-card:hover .add-video-content {
  color: #00aeec;
}

/* 水平视频项样式 */
.video-horizontal-item {
  width: 180px;
  flex-shrink: 0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.video-horizontal-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.video-horizontal-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.video-horizontal-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-horizontal-info {
  margin-top: 8px;
}

.video-horizontal-info .video-title {
  font-size: 14px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.video-horizontal-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.video-horizontal-meta .video-views {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 更多按钮 */
.more-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  color: #666;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
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
@media (max-width: 1200px) {
  .collections-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 992px) {
  .collections-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .collections-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .collection-horizontal-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}

@media (max-width: 576px) {
  .collections-grid {
    grid-template-columns: 1fr;
  }
}
</style>