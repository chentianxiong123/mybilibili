<template>
  <!-- 顶部标题和新建按钮 -->
  <div class="collections-header">
    <div class="collections-header-left">
      <h2 class="collections-title">我的合集和视频列表</h2>
      <el-button
        type="primary"
        :icon="Plus"
        class="new-collection-btn"
        @click="emit('open-create')"
      >
        新建
      </el-button>
    </div>
    <!-- 视图切换按钮 -->
    <div class="view-toggle">
      <button
        class="view-toggle-btn grid-view-btn"
        :class="{ active: viewType === 'grid' }"
        @click="emit('update:viewType', 'grid')"
      >
        <el-icon><Menu /></el-icon>
      </button>
      <button
        class="view-toggle-btn list-view-btn"
        :class="{ active: viewType === 'horizontal' }"
        @click="emit('update:viewType', 'horizontal')"
      >
        <el-icon><Document /></el-icon>
      </button>
    </div>
  </div>

  <!-- 加载状态 -->
  <div v-if="loading" class="loading-state">
    <el-skeleton :rows="3" animated />
  </div>

  <!-- 空状态 -->
  <div v-else-if="items.length === 0" class="empty-state">
    <el-empty description="暂无合集">
      <el-button
        type="primary"
        @click="emit('open-create')"
      >
        创建合集
      </el-button>
    </el-empty>
  </div>

  <!-- 宫格视图 -->
  <div v-else-if="viewType === 'grid'" class="collections-grid">
    <!-- 新建合集卡片 -->
    <div
      class="collection-item new-collection"
      @click="emit('open-create')"
    >
      <div class="new-collection-content">
        <el-icon :size="32"><Plus /></el-icon>
        <span>新建合集</span>
      </div>
    </div>

    <!-- 合集项 -->
    <div
      v-for="collection in items"
      :key="collection.id"
      class="collection-item"
      @click="emit('go-to-detail', collection.id)"
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
  <div v-else-if="viewType === 'horizontal'" class="collections-horizontal">
    <!-- 合集项 -->
    <div
      v-for="collection in items"
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
            @click="emit('go-to-detail', collection.id)"
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
          class="add-video-card"
          @click="emit('add-video', collection.id)"
        >
          <div class="add-video-content">
            <el-icon :size="24"><Plus /></el-icon>
            <span>添加稿件</span>
          </div>
        </div>

        <!-- 视频项 -->
        <div
          v-for="video in collection.videos"
          :key="video.id"
          class="video-horizontal-item"
          @click="emit('play-video', video)"
        >
          <div class="video-horizontal-cover">
            <img loading="lazy" decoding="async" :src="video.coverUrl || getDefaultCover()" :alt="video.title" class="video-cover-img">
            <div class="video-duration">{{ video.duration }}</div>
          </div>
          <div class="video-horizontal-info">
            <div class="video-title" :title="video.title">{{ video.title }}</div>
            <div class="video-horizontal-meta">
              <span class="video-views">
                <el-icon><Monitor /></el-icon>
                {{ video.viewCount || 0 }}
              </span>
              <span class="video-date">{{ video.date }}</span>
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
  Menu,
  Document,
  Monitor,
  Plus,
  ArrowRight
} from '@element-plus/icons-vue'

const props = defineProps({
  items: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  viewType: {
    type: String,
    default: 'horizontal'
  }
})

const emit = defineEmits([
  'open-create',
  'go-to-detail',
  'add-video',
  'play-all',
  'play-video',
  'update:viewType'
])

const getDefaultCover = () => {
  return 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="400" height="225" viewBox="0 0 400 225"><rect fill="#e5e9ef" width="400" height="225"/><text fill="#9499a0" font-family="sans-serif" font-size="16" x="50%" y="50%" text-anchor="middle" dy=".3em">暂无封面</text></svg>')
}
</script>

<style scoped>
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
  color: #303133;
  margin: 0;
}

.view-toggle {
  display: flex;
  gap: 4px;
}

.view-toggle-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #dcdfe6;
  background: #fff;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.view-toggle-btn.active {
  background: #00a1d6;
  border-color: #00a1d6;
  color: #fff;
}

.collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 280px));
  gap: 16px;
  justify-content: start;
}

.collection-item {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  max-width: 280px;
  width: 100%;
}

.collection-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.collection-item.new-collection {
  border: 2px dashed #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.collection-item.new-collection:hover {
  border-color: #00a1d6;
  background: #f0f9ff;
}

.new-collection-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.collection-item.new-collection:hover .new-collection-content {
  color: #00a1d6;
}

.collection-cover {
  position: relative;
  aspect-ratio: 16/9;
  width: 100%;
  overflow: hidden;
}

.collection-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-count {
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

.collection-info {
  padding: 12px;
}

.collection-info .collection-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-info .collection-date {
  font-size: 12px;
  color: #909399;
}

.collections-horizontal {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.collection-horizontal-item {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.collection-horizontal-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
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
  color: #303133;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-count-badge {
  background: #f4f4f5;
  color: #909399;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: normal;
  flex-shrink: 0;
}

.collection-horizontal-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.collection-horizontal-actions .action-btn {
  padding: 6px 12px;
  font-size: 13px;
}

.collection-videos-horizontal {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
  scrollbar-width: thin;
  scrollbar-color: #e0e0e0 #f5f7fa;
}

.collection-videos-horizontal::-webkit-scrollbar {
  height: 6px;
}

.collection-videos-horizontal::-webkit-scrollbar-track {
  background: #f5f7fa;
  border-radius: 3px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 3px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb:hover {
  background: #d0d0d0;
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

.collection-videos-horizontal .add-video-card {
  min-width: 160px;
  height: 100px;
  flex-shrink: 0;
  max-width: 160px;
}

.video-horizontal-item {
  min-width: 160px;
  max-width: 160px;
  flex-shrink: 0;
  cursor: pointer;
  transition: transform 0.3s ease;
}

.video-horizontal-item:hover {
  transform: translateY(-2px);
}

.video-horizontal-cover {
  position: relative;
  aspect-ratio: 16/9;
  border-radius: 6px;
  overflow: hidden;
  background: #f0f2f5;
}

.video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.video-horizontal-item:hover .video-cover-img {
  transform: scale(1.05);
}

.video-horizontal-cover .video-duration {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 1px 4px;
  border-radius: 2px;
  font-size: 11px;
  z-index: 1;
}

.video-horizontal-info {
  padding: 8px 0;
}

.video-horizontal-info .video-title {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
  line-height: 1.3;
}

.video-horizontal-meta {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #909399;
  align-items: center;
}

.video-horizontal-meta .video-views {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-horizontal-meta .video-date {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
  width: 80px;
  text-align: right;
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