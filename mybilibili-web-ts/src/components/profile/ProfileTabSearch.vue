<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, VideoPlay, Search } from '@element-plus/icons-vue'

const props = defineProps({
  videoSearch: {
    type: Object,
    required: true
  },
  allVideos: {
    type: Array,
    default: () => []
  },
  searchCategories: {
    type: Array,
    default: () => []
  }
})

const router = useRouter()

const normalizeVideoSearchText = (value) => (value || '').toString().trim().toLowerCase()

const filteredSearchVideos = computed(() => {
  const keyword = normalizeVideoSearchText(props.videoSearch.keyword)
  let videos = [...props.allVideos]

  if (props.videoSearch.activeSort === '最多播放') {
    videos.sort((a, b) => (b.viewCount || 0) - (a.viewCount || 0))
  } else if (props.videoSearch.activeSort === '最多收藏') {
    videos.sort((a, b) => (b.collectCount || 0) - (a.collectCount || 0))
  } else {
    videos.sort((a, b) => new Date(b.uploadTime) - new Date(a.uploadTime))
  }

  if (!keyword) {
    return videos
  }

  return videos.filter(video => {
    const fields = [
      video.title,
      video.description,
      video.introduction,
      video.tname,
      video.id
    ]
    return fields.some(field => normalizeVideoSearchText(field).includes(keyword))
  })
})

const filteredSearchTotalCount = computed(() => filteredSearchVideos.value.length)

watch(filteredSearchVideos, (videos) => {
  props.videoSearch.searchResults = videos
  props.videoSearch.totalCount = videos.length
})

// 处理搜索排序变化
const handleSearchSortChange = (option) => {
  props.videoSearch.activeSort = option
  props.videoSearch.searchResults = filteredSearchVideos.value
  props.videoSearch.totalCount = filteredSearchVideos.value.length
}

// 播放搜索视频
const playAllSearchVideos = () => {
  if (filteredSearchVideos.value.length > 0) {
    router.push(`/manuscript/${filteredSearchVideos.value[0].id}`)
  } else {
    ElMessage.info('暂无视频')
  }
}
</script>

<template>
  <div class="search-section">
    <!-- 顶部容器，与投稿页面同宽 -->
    <div class="search-top">
      <!-- 左侧分类导航 -->
      <div class="search-sidebar">
        <div
          v-for="category in searchCategories"
          :key="category.name"
          :class="['category-item', { active: videoSearch.activeCategory === category.name }]"
          @click="videoSearch.activeCategory = category.name"
        >
          <span class="category-name">{{ category.name }}</span>
          <span class="category-count">{{ category.count }}</span>
        </div>
      </div>

      <!-- 右侧内容区域 -->
      <div class="search-content">
        <!-- 标题和排序 -->
        <div class="search-header">
          <div class="header-content">
            <div class="search-title">
              <h3>{{ videoSearch.activeCategory === '视频' ? '我的视频' : '我的动态' }}</h3>
            </div>
            <!-- 排序选项 -->
            <div class="sort-options" v-if="videoSearch.activeCategory === '视频'">
              <div
                v-for="option in videoSearch.sortOptions"
                :key="option"
                :class="['sort-item', { active: videoSearch.activeSort === option }]"
                @click="handleSearchSortChange(option)"
              >
                {{ option }}
              </div>
            </div>
          </div>
          <!-- 操作按钮 -->
          <div class="action-buttons" v-if="videoSearch.activeCategory === '视频'">
            <button class="action-btn play-all-btn" @click="playAllSearchVideos">播放全部</button>
            <button
              class="action-btn view-toggle-btn grid-view-btn"
              :class="{ active: videoSearch.viewType === 'grid' }"
              @click="videoSearch.viewType = 'grid'"
            >
              <span class="view-icon">🔲</span>
            </button>
            <button
              class="action-btn view-toggle-btn list-view-btn"
              :class="{ active: videoSearch.viewType === 'list' }"
              @click="videoSearch.viewType = 'list'"
            >
              <span class="view-icon">📋</span>
            </button>
          </div>
        </div>

        <!-- 搜索结果统计 -->
        <div class="search-result-info" v-if="videoSearch.keyword && videoSearch.activeCategory === '视频'">
          共找到关于"<span class="keyword">{{ videoSearch.keyword }}</span>"的 {{ filteredSearchTotalCount }} 个视频
        </div>

        <!-- 视频列表 -->
        <div v-if="videoSearch.activeCategory === '视频'">
          <div v-if="videoSearch.loading" class="loading-state">
            <p>搜索中...</p>
          </div>
          <div v-else-if="filteredSearchVideos.length === 0 && videoSearch.keyword" class="empty-state">
            <p>未找到相关视频</p>
          </div>
          <div v-else-if="filteredSearchVideos.length > 0">
            <!-- 宫格视图 -->
            <div v-if="videoSearch.viewType === 'grid'" class="videos-grid">
              <div v-for="video in filteredSearchVideos" :key="video.id" class="video-item" @click="router.push(`/manuscript/${video.id}`)">
                <div class="video-cover">
                  <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title" class="video-cover-img">
                  <div class="video-duration">{{ video.duration }}</div>
                  <!-- 仅自己可见标签 -->
                  <div class="video-visibility" v-if="video.status === 'private'">
                    <el-icon><Lock /></el-icon>
                    仅自己可见
                  </div>
                </div>
                <div class="video-title">{{ video.title }}</div>
                <div class="video-meta">
                  <span class="video-views">{{ video.viewCount || 0 }}</span>
                  <span class="video-date">{{ video.date }}</span>
                </div>
              </div>
            </div>

            <!-- 列表视图 -->
            <div v-else-if="videoSearch.viewType === 'list'" class="videos-list">
              <div v-for="video in filteredSearchVideos" :key="video.id" class="video-list-item" @click="router.push(`/manuscript/${video.id}`)">
                <div class="video-list-cover">
                  <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title" class="video-cover-img">
                  <div class="video-duration">{{ video.duration }}</div>
                </div>
                <div class="video-list-info">
                  <div class="video-title">{{ video.title }}</div>
                  <div class="video-description" v-if="video.description">
                    {{ video.description }}
                  </div>
                  <div class="video-description" v-else></div>
                  <div class="video-list-meta">
                    <span class="video-views">{{ video.viewCount || 0 }}</span>
                    <span class="video-date">{{ video.date }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <p>请输入关键词搜索视频</p>
          </div>
        </div>

        <!-- 动态搜索内容 -->
        <div v-else class="dynamic-search-content">
          <div class="empty-state">
            <p>动态搜索功能开发中...</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 搜索页面样式 - 与投稿页面一致 */
.search-section {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 0;
  width: 100%;
  border-radius: 0;
}

/* 顶部容器，与封面图片同宽 */
.search-top {
  display: flex;
  gap: 0;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

/* 左侧分类导航 */
.search-sidebar {
  width: 120px;
  min-width: 120px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-right: 20px;
  border-right: 1px solid #f0f0f0;
}

/* 右侧内容区域 */
.search-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 搜索结果统计 */
.search-result-info {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
  padding: 5px 0;
}

.search-result-info .keyword {
  color: #00aeec;
  font-weight: 500;
}

/* 视频可见性标签 */
.video-visibility {
  position: absolute;
  bottom: 8px;
  left: 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.video-visibility .el-icon {
  font-size: 12px;
}

/* 动态搜索内容 */
.dynamic-search-content {
  padding: 40px 0;
}

/* 搜索页面响应式设计 */
@media (max-width: 992px) {
  .search-section {
    flex-direction: column;
  }

  .search-sidebar {
    width: 100%;
    min-width: auto;
    flex-direction: row;
    padding-right: 0;
    border-right: none;
    border-bottom: 1px solid #f0f0f0;
    padding-bottom: 15px;
  }

  .search-content {
    padding-top: 15px;
  }
}

/* 分类导航项 */
.category-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.3s ease;
  border-radius: 4px;
  margin-right: 10px;
}

.category-item:hover {
  color: #00aeec;
  background-color: rgba(0, 174, 236, 0.1);
}

.category-item.active {
  color: #fff;
  font-weight: 500;
  background-color: #00aeec;
  border-right: none;
  padding-right: 12px;
}

.category-count {
  font-size: 12px;
  color: #9499a0;
  transition: color 0.3s ease;
}

.category-item.active .category-count {
  color: #fff;
}

/* 标题和排序 */
.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 15px;
  border-bottom: 1px solid #f0f0f0;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.search-title h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
}

/* 排序选项 */
.sort-options {
  display: flex;
  gap: 20px;
  align-items: center;
}

.sort-item {
  padding: 8px 16px;
  font-size: 16px;
  color: #666;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 6px;
  background-color: transparent;
  border: none;
  outline: none;
}

.sort-item:hover {
  color: #00aeec;
  background-color: rgba(0, 174, 236, 0.1);
}

.sort-item.active {
  color: #00aeec;
  background-color: rgba(0, 174, 236, 0.1);
  font-weight: 500;
  border-bottom: 2px solid #00aeec;
}

/* 操作按钮 */
.action-buttons {
  display: flex !important;
  gap: 10px !important;
  flex-direction: row !important;
  align-items: center !important;
  flex-wrap: nowrap !important;
}

.action-btn {
  padding: 6px 16px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.3s ease;
  display: inline-block !important;
  flex-direction: row !important;
  align-items: center !important;
  justify-content: center !important;
  white-space: nowrap !important;
  width: auto !important;
  height: auto !important;
}

.action-btn:hover {
  background-color: #f5f5f5;
  color: #333;
}

.play-all-btn {
  background-color: #fff;
  color: #666;
  border-color: #e0e0e0;
}

.play-all-btn:hover {
  background-color: #f5f5f5;
  border-color: #e0e0e0;
  color: #333;
}

/* 视图切换按钮样式 */
.view-toggle-btn {
  padding: 6px 10px;
  border: none;
  background-color: transparent;
  color: #666;
  border-radius: 4px;
}

.view-toggle-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: #333;
}

.view-toggle-btn.active {
  background-color: rgba(0, 174, 236, 0.1);
  color: #00aeec;
}

.view-icon {
  font-size: 16px;
}

/* 视频网格 */
.videos-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

.video-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  position: relative;
  transition: opacity 0.2s;
}

.video-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.video-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.video-cover:hover img {
  transform: scale(1.05);
}

.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 2px;
}

.video-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  margin: 0;
  min-height: 39.2px;
}

.video-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #999;
}

.video-meta .video-views {
  margin-right: 10px;
}

.video-date {
  font-size: 12px;
  color: #999;
}

.video-views {
  font-size: 12px;
  color: #999;
  margin: 0;
}

/* 列表视图样式 */
.videos-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.video-list-item {
  display: flex;
  gap: 15px;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  align-items: flex-start;
}

.video-list-item:last-child {
  border-bottom: none;
}

.video-list-cover {
  position: relative;
  width: 160px;
  height: 90px;
  flex-shrink: 0;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.video-list-cover .video-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-list-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.video-list-info .video-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  margin: 0;
  min-height: 21px;
}

.video-description {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  margin: 0;
  min-height: 42px;
}

.video-list-meta {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #999;
  align-items: center;
  margin-top: 0;
}

.video-list-meta .video-views {
  margin-right: 0;
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
  .videos-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 992px) {
  .videos-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .videos-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .search-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
}

@media (max-width: 576px) {
  .videos-grid {
    grid-template-columns: 1fr;
  }
}
</style>