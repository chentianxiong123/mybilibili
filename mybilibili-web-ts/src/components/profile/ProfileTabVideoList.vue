<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check, View, ChatDotRound } from '@element-plus/icons-vue'
import { userApi } from '@/api/client'

const props = defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  allVideos: {
    type: Array,
    default: () => []
  },
  videoSortOption: {
    type: String,
    default: '最新发布'
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

const sortOptions = ['最新发布', '最多播放', '最多收藏']

// 置顶视频相关数据
const pinnedVideo = ref(null)
const showPinnedVideoDialog = ref(false)
const pinnedVideoSelection = ref(null)

// 加载置顶视频
const loadPinnedVideo = async () => {
  if (!props.userId) return
  try {
    const response = await userApi.getPinnedVideo(props.userId)
    if (response.code === 200 && response.data) {
      pinnedVideo.value = response.data
    } else {
      pinnedVideo.value = null
    }
  } catch (error) {
    console.warn('获取置顶视频失败（非关键功能）:', error)
    pinnedVideo.value = null
  }
}

// 打开置顶视频选择对话框
const openPinnedVideoDialog = () => {
  if (props.allVideos.length === 0) {
    ElMessage.warning('暂无可选的视频')
    return
  }
  pinnedVideoSelection.value = pinnedVideo.value ? { ...pinnedVideo.value } : null
  showPinnedVideoDialog.value = true
}

// 保存置顶视频
const savePinnedVideo = async () => {
  if (!pinnedVideoSelection.value) {
    ElMessage.warning('请选择一个视频作为置顶视频')
    return
  }
  try {
    const response = await userApi.setPinnedVideo(pinnedVideoSelection.value.id)
    if (response.code === 200) {
      pinnedVideo.value = pinnedVideoSelection.value
      showPinnedVideoDialog.value = false
      ElMessage.success('置顶视频设置成功')
    } else {
      ElMessage.error(response.message || '设置置顶视频失败')
    }
  } catch (error) {
    console.error('设置置顶视频失败:', error)
    ElMessage.error('设置置顶视频失败，请稍后重试')
  }
}

// 取消置顶视频
const removePinnedVideo = async () => {
  try {
    const response = await userApi.removePinnedVideo()
    if (response.code === 200) {
      pinnedVideo.value = null
      showPinnedVideoDialog.value = false
      ElMessage.success('已取消置顶视频')
    } else {
      ElMessage.error(response.message || '取消置顶失败')
    }
  } catch (error) {
    console.error('取消置顶视频失败:', error)
    ElMessage.error('取消置顶失败，请稍后重试')
  }
}

// 处理排序变化
const handleSortChange = (option) => {
  console.log('【调试】handleSortChange 被调用，选项:', option)
  emit('sort-change', option)
}

// 播放TA的视频
const playAllVideos = () => {
  if (props.allVideos.length > 0) {
    router.push(`/manuscript/${props.allVideos[0].id}`)
  } else {
    ElMessage.info('暂无视频')
  }
}

// 跳转到投稿页面
const goToSubmissions = () => {
  router.push(`/profile/${props.userId}/submissions`)
}

onMounted(() => {
  loadPinnedVideo()
})

watch(() => props.userId, () => {
  loadPinnedVideo()
})
</script>

<template>
  <div>
    <!-- 置顶视频 -->
    <div class="content-section">
      <div class="section-controls-wrapper">
        <div class="left-controls">
          <div class="section-header">
            <h3 class="section-title">置顶视频</h3>
          </div>
        </div>
        <div class="action-buttons" v-if="isOwnSpace">
          <button class="action-btn" @click="openPinnedVideoDialog">设置置顶</button>
        </div>
      </div>
      <div v-if="!pinnedVideo" class="empty-state">
        <p>暂无置顶视频</p>
      </div>
      <div v-else class="pinned-video-container">
        <div class="pinned-video-card" @click="router.push(`/manuscript/${pinnedVideo.id}`)">
          <div class="pinned-video-left">
            <img loading="lazy" decoding="async" class="pinned-video-img" :src="pinnedVideo.coverUrl" :alt="pinnedVideo.title">
            <span class="pinned-video-time">{{ pinnedVideo.duration }}</span>
            <div class="pinned-badge">置顶</div>
          </div>
          <div class="pinned-video-right">
            <p class="pinned-video-title">{{ pinnedVideo.title }}</p>
            <div class="pinned-video-stats">
              <span class="pinned-stat-item">
                <el-icon><View /></el-icon>
                {{ pinnedVideo.viewCount }}
              </span>
              <span class="pinned-stat-item">
                <el-icon><ChatDotRound /></el-icon>
                {{ pinnedVideo.commentCount || 0 }}
              </span>
              <span class="pinned-stat-item">{{ pinnedVideo.date }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TA的视频 -->
    <div class="content-section">
      <div class="section-controls-wrapper">
        <div class="left-controls">
          <div class="section-header">
            <h3 class="section-title">TA的视频</h3>
            <div class="section-count">{{ allVideos.length }}</div>
          </div>
          <div class="sort-options">
            <button
              v-for="option in sortOptions"
              :key="option"
              :class="['sort-btn', { active: videoSortOption === option }]"
              @click="handleSortChange(option)"
            >
              {{ option }}
            </button>
          </div>
        </div>
        <div class="action-buttons">
          <button class="action-btn play-all-btn" @click="playAllVideos">播放全部</button>
          <button class="action-btn more-btn" @click="goToSubmissions">更多</button>
        </div>
      </div>
      <div v-if="loading.videos" class="loading-state">
        <p>加载中...</p>
      </div>
      <div v-else-if="allVideos.length === 0" class="empty-state">
        <p>暂无视频</p>
      </div>
      <div v-else class="videos-grid">
        <div v-for="video in allVideos.slice(0, 6)" :key="video.id" class="video-item" @click="router.push(`/manuscript/${video.id}`)">
          <div class="video-cover">
            <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title">
            <div class="video-duration">{{ video.duration }}</div>
          </div>
          <div class="video-title">{{ video.title }}</div>
          <div class="video-meta">
            <span class="video-views">{{ video.viewCount }}</span>
            <span class="video-date">{{ video.date }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 置顶视频选择对话框 -->
    <el-dialog
      v-model="showPinnedVideoDialog"
      title="设置置顶视频"
      width="700px"
    >
      <div class="pinned-video-selection">
        <div v-if="allVideos.length === 0" class="empty-state">
          <p>暂无可选的视频</p>
        </div>
        <div v-else class="video-selection-grid">
          <div
            v-for="video in allVideos"
            :key="video.id"
            :class="['video-selection-card', { 'is-selected': pinnedVideoSelection?.id === video.id }]"
            @click="pinnedVideoSelection = video"
          >
            <div class="video-selection-cover">
              <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title">
              <span class="video-selection-duration">{{ video.duration }}</span>
              <div v-if="pinnedVideoSelection?.id === video.id" class="selected-badge">
                <el-icon><Check /></el-icon>
              </div>
            </div>
            <div class="video-selection-title">{{ video.title }}</div>
            <div class="video-selection-stats">
              <span>{{ video.viewCount }}播放</span>
              <span>{{ video.date }}</span>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button v-if="pinnedVideo" @click="removePinnedVideo">取消置顶</el-button>
          <el-button @click="showPinnedVideoDialog = false">取消</el-button>
          <el-button type="primary" @click="savePinnedVideo">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
/* 内容区域通用样式 */
.content-section {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
}

/* 控制区域样式 */
.section-controls-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.left-controls {
  display: flex;
  align-items: center;
  gap: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 0;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.section-count {
  font-size: 14px;
  color: #9499a0;
}

.sort-options {
  display: flex;
  gap: 10px;
  align-items: center;
}

.sort-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  color: #666;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.sort-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.sort-btn.active {
  background-color: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

.action-buttons {
  display: flex;
  gap: 10px;
  align-items: center;
}

.action-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  color: #666;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.play-all-btn {
  background-color: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

.play-all-btn:hover {
  background-color: #0095d9;
  border-color: #0095d9;
  color: #fff;
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

/* 置顶视频区域样式 */
.pinned-video-container {
  margin-top: 16px;
}

.pinned-video-card {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  min-height: 135px;
  cursor: pointer;
  transition: transform 0.2s;
}

.pinned-video-card:hover {
  transform: translateY(-2px);
}

.pinned-video-left {
  position: relative;
  width: 240px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
  aspect-ratio: 16 / 9;
}

.pinned-video-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.pinned-video-time {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.pinned-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  background: #00a1d6;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.pinned-video-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 0;
}

.pinned-video-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.pinned-video-stats {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: #999;
}

.pinned-stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 置顶视频选择对话框样式 */
.pinned-video-selection {
  max-height: 500px;
  overflow-y: auto;
}

.video-selection-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.video-selection-card {
  cursor: pointer;
  border: 2px solid transparent;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s;
}

.video-selection-card:hover {
  border-color: #00a1d6;
}

.video-selection-card.is-selected {
  border-color: #00a1d6;
  background-color: #f0f9ff;
}

.video-selection-cover {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
}

.video-selection-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-selection-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.selected-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  background: #00a1d6;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-selection-title {
  padding: 8px;
  font-size: 14px;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.video-selection-stats {
  padding: 0 8px 8px;
  font-size: 12px;
  color: #999;
  display: flex;
  justify-content: space-between;
}

/* 视频网格样式 */
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
}

.video-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #9499a0;
  flex-wrap: wrap;
  white-space: nowrap;
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

  .section-controls-wrapper {
    flex-direction: column;
    align-items: flex-start;
  }

  .left-controls {
    width: 100%;
    flex-wrap: wrap;
  }
}

@media (max-width: 576px) {
  .videos-grid {
    grid-template-columns: 1fr;
  }

  .sort-options {
    width: 100%;
    flex-wrap: wrap;
  }

  .sort-btn {
    flex: 1;
    min-width: 80px;
  }
}
</style>