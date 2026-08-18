<script setup>
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const props = defineProps({
  submissions: {
    type: Object,
    required: true
  },
  loading: {
    type: Object,
    required: true
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['sort-change'])

const router = useRouter()

// 处理投稿页面排序变化
const handleSubmissionsSortChange = (option) => {
  console.log('【调试】handleSubmissionsSortChange 被调用，选项:', option)
  props.submissions.activeSort = option
  // 同步更新主页和搜索页的排序选项
  emit('sort-change', option)
}

// 播放投稿视频
const playAllSubmissions = () => {
  if (props.submissions.videos.length > 0) {
    router.push(`/manuscript/${props.submissions.videos[0].id}`)
  } else {
    ElMessage.info('暂无投稿视频')
  }
}
</script>

<template>
  <div class="submissions-section">
    <!-- 顶部容器，与封面图片同宽 -->
    <div class="submissions-top">
      <!-- 左侧分类导航 -->
      <div class="submissions-sidebar">
        <div
          v-for="category in submissions.categories"
          :key="category.name"
          :class="['category-item', { active: submissions.activeCategory === category.name }]"
          @click="submissions.activeCategory = category.name"
        >
          <span class="category-name">{{ category.name }}</span>
          <span class="category-count">{{ category.count }}</span>
        </div>
      </div>

      <!-- 右侧内容区域 -->
      <div class="submissions-content">
        <!-- 标题和排序 -->
        <div class="submissions-header">
          <div class="header-content">
            <div class="submissions-title">
              <h3>{{ submissions.activeCategory === '视频' ? '我的视频' : '我的动态' }}</h3>
            </div>
            <!-- 排序选项 -->
            <div class="sort-options" v-if="submissions.activeCategory === '视频'">
              <div
                v-for="option in submissions.sortOptions"
                :key="option"
                :class="['sort-item', { active: submissions.activeSort === option }]"
                @click="handleSubmissionsSortChange(option)"
              >
                {{ option }}
              </div>
            </div>
          </div>
          <!-- 操作按钮 -->
          <div class="action-buttons" v-if="submissions.activeCategory === '视频'">
            <button class="action-btn play-all-btn" @click="playAllSubmissions">播放全部</button>
            <button
              class="action-btn view-toggle-btn grid-view-btn"
              :class="{ active: submissions.viewType === 'grid' }"
              @click="submissions.viewType = 'grid'"
            >
              <span class="view-icon">🔲</span>
            </button>
            <button
              class="action-btn view-toggle-btn list-view-btn"
              :class="{ active: submissions.viewType === 'list' }"
              @click="submissions.viewType = 'list'"
            >
              <span class="view-icon">📋</span>
            </button>
          </div>
        </div>

        <!-- 视频列表 -->
        <div v-if="submissions.activeCategory === '视频'">
          <div v-if="loading.videos" class="loading-state">
            <p>加载中...</p>
          </div>
          <div v-else-if="submissions.videos.length === 0" class="empty-state">
            <p>暂无投稿</p>
          </div>
          <div v-else>
            <!-- 宫格视图 -->
            <div v-if="submissions.viewType === 'grid'" class="videos-grid">
              <div v-for="video in submissions.videos" :key="video.id" class="video-item" @click="router.push(`/manuscript/${video.id}`)">
                <div class="video-cover">
                  <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title" class="video-cover-img">
                  <div class="video-duration">{{ video.duration }}</div>
                </div>
                <div class="video-title">{{ video.title }}</div>
                <div class="video-meta">
                  <span class="video-views">{{ video.viewCount }}</span>
                  <span class="video-date">{{ video.date }}</span>
                </div>
              </div>
            </div>

            <!-- 列表视图 -->
            <div v-else-if="submissions.viewType === 'list'" class="videos-list">
              <div v-for="video in submissions.videos" :key="video.id" class="video-list-item" @click="router.push(`/manuscript/${video.id}`)">
                <div class="video-list-cover">
                  <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title" class="video-cover-img">
                  <div class="video-duration">{{ video.duration }}</div>
                </div>
                <div class="video-list-info">
                  <div class="video-title">{{ video.title }}</div>
                  <!-- 视频简介 -->
                  <div class="video-description" v-if="video.description">
                    {{ video.description }}
                  </div>
                  <div class="video-description" v-else>
                    <!-- 空白占位 -->
                  </div>
                  <!-- 元信息显示在简介下方 -->
                  <div class="video-list-meta">
                    <span class="video-views">{{ video.viewCount }}</span>
                    <span class="video-date">{{ video.date }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 分页控件 -->
            <div class="pagination">
              <div class="pagination-buttons">
                <button
                  v-for="page in submissions.pagination.totalPages"
                  :key="page"
                  :class="['page-btn', { active: submissions.pagination.currentPage === page }]"
                  @click="submissions.pagination.currentPage = page"
                >
                  {{ page }}
                </button>
                <button
                  class="page-btn"
                  :disabled="submissions.pagination.currentPage === submissions.pagination.totalPages"
                >
                  下一页
                </button>
              </div>
              <div class="pagination-info">
                共 {{ submissions.pagination.totalPages }} 页，跳至
                <input type="text" link class="page-input" :value="submissions.pagination.currentPage" />
                页
              </div>
            </div>
          </div>
        </div>

        <!-- 动态内容 -->
        <div v-else class="dynamic-content">
          <div class="empty-state">
            <p>暂无动态</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 投稿页面样式 */
.submissions-section {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 0;
  width: 100%;
  border-radius: 0;
}

/* 顶部容器，与封面图片同宽 */
.submissions-top {
  display: flex;
  gap: 0;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

/* 左侧分类导航 */
.submissions-sidebar {
  width: 120px;
  min-width: 120px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-right: 20px;
  border-right: 1px solid #f0f0f0;
}

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

/* 右侧内容区域 */
.submissions-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 标题和排序 */
.submissions-header {
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

.submissions-title h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
}

/* 视频网格 */
.videos-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

/* 视频项 */
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

.video-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
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

.video-date {
  font-size: 12px;
  color: #9499a0;
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

/* 操作按钮 */
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

.action-buttons {
  display: flex !important;
  gap: 10px !important;
  flex-direction: row !important;
  align-items: center !important;
  flex-wrap: nowrap !important;
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

/* 分页控件 */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
  flex-wrap: wrap;
  gap: 15px;
}

.pagination-info {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #666;
  order: 2;
}

.page-input {
  width: 50px;
  padding: 4px 8px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 14px;
  text-align: center;
}

.pagination-buttons {
  display: flex;
  gap: 5px;
  order: 1;
  align-items: center;
}

.page-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  color: #666;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 32px;
  text-align: center;
}

.page-btn:hover:not(:disabled) {
  color: #00aeec;
  border-color: #00aeec;
}

.page-btn.active {
  background-color: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 排序选项样式 */
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

/* 标题和排序内联样式 */
.submissions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 15px;
  border-bottom: 1px solid #f0f0f0;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .videos-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 992px) {
  .submissions-section {
    flex-direction: column;
  }

  .submissions-sidebar {
    width: 100%;
    min-width: auto;
    flex-direction: row;
    padding-right: 0;
    border-right: none;
    border-bottom: 1px solid #f0f0f0;
    padding-bottom: 15px;
  }

  .category-item.active {
    border-right: none;
    border-bottom: none;
    padding-right: 12px;
    padding-bottom: 8px;
    background-color: #00aeec;
    color: #fff;
  }

  .videos-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .videos-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .submissions-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }

  .submissions-controls {
    width: 100%;
    flex-wrap: wrap;
  }
}

@media (max-width: 576px) {
  .videos-grid {
    grid-template-columns: 1fr;
  }

  .submissions-sidebar {
    flex-direction: column;
    align-items: flex-start;
  }

  .category-item.active {
    border-right: none;
    border-bottom: none;
    padding-right: 12px;
    padding-bottom: 8px;
    background-color: #00aeec;
    color: #fff;
  }
}
</style>