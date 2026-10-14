<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Top, MoreFilled, Delete, Share, ChatDotRound } from '@element-plus/icons-vue'
import { dynamicApi } from '@/api/dynamic.ts'
import CommentSystem from '@/components/CommentSystem.vue'

const props = defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  userInfo: {
    type: Object,
    required: true
  },
  loading: {
    type: Object,
    required: true
  }
})

const router = useRouter()

// 动态数据
const dynamics = ref([])

// 左侧筛选：全部 / 视频
const activeFilter = ref('all')
const filterOptions = [
  { key: 'all', label: '全部' },
  { key: 'video', label: '视频' }
]

const filteredDynamics = computed(() => {
  if (activeFilter.value === 'video') {
    return dynamics.value.filter(d => d.video)
  }
  return dynamics.value
})

// 评论展开状态管理
const expandedComments = ref(new Set())

// 展开状态管理
const expandedDynamics = ref(new Set())

// 切换展开状态
const toggleExpand = (dynamicId) => {
  if (expandedDynamics.value.has(dynamicId)) {
    expandedDynamics.value.delete(dynamicId)
  } else {
    expandedDynamics.value.add(dynamicId)
  }
}

// 处理置顶
const handleStickDynamic = async (dynamicId) => {
  try {
    const dynamic = dynamics.value.find(d => d.id === dynamicId)
    if (dynamic) {
      dynamic.isTop = !dynamic.isTop
      // 重新排序：置顶的排在前面
      dynamics.value.sort((a, b) => {
        if (a.isTop === b.isTop) {
          return new Date(b.createTime) - new Date(a.createTime)
        }
        return a.isTop ? -1 : 1
      })
      ElMessage.success(dynamic.isTop ? '置顶成功' : '取消置顶成功')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 处理删除
const handleDeleteDynamic = async (dynamicId) => {
  try {
    await ElMessageBox.confirm('确定要删除这条动态吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await dynamicApi.deleteDynamic(dynamicId)
    if (res.code === 200) {
      dynamics.value = dynamics.value.filter(d => d.id !== dynamicId)
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 处理点赞
const handleLikeDynamic = async (dynamic) => {
  try {
    if (dynamic.stats.isLiked) {
      // 取消点赞
      const res = await dynamicApi.unlikeDynamic(dynamic.id)
      if (res.code === 200) {
        // 使用后端返回的真实数据
        dynamic.stats.isLiked = res.data?.isLiked ?? false
        dynamic.stats.likeCount = res.data?.likeCount ?? 0
        ElMessage.success('取消点赞成功')
      } else {
        ElMessage.error(res.message || '取消点赞失败')
      }
    } else {
      // 点赞
      const res = await dynamicApi.likeDynamic(dynamic.id)
      if (res.code === 200) {
        // 使用后端返回的真实数据
        dynamic.stats.isLiked = res.data?.isLiked ?? true
        dynamic.stats.likeCount = res.data?.likeCount ?? 0
        ElMessage.success('点赞成功')
      } else {
        ElMessage.error(res.message || '点赞失败')
      }
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 切换评论展开状态
const toggleComments = (dynamic) => {
  const dynamicId = dynamic.id
  if (expandedComments.value.has(dynamicId)) {
    expandedComments.value.delete(dynamicId)
  } else {
    expandedComments.value.add(dynamicId)
  }
}

// 格式化时间：MM月DD日
const formatDynamicTime = (timeStr) => {
  const date = new Date(timeStr)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}月${day}日`
}

// 时间行：纯动态只有时间，带视频的加动作描述（与真站一致）
const formatHeaderLine = (dynamic) => {
  const time = formatDynamicTime(dynamic.createTime)
  if (dynamic.video) {
    return `${time} · 投稿了视频`
  }
  if (dynamic.images && dynamic.images.length) {
    return `${time} · 投稿了图文`
  }
  return time
}

// 格式化播放数（万）
const formatNumber = (num) => {
  if (!num) return 0
  if (num >= 10000) {
    return (num / 10000).toFixed(1).replace(/\.0$/, '') + '万'
  }
  return num
}

// 打开视频
const openVideo = (videoId) => {
  if (videoId) router.push(`/manuscript/${videoId}`)
}

// 打开图片预览（点击放大）
const previewImage = (url) => {
  window.open(url, '_blank')
}

// 加载用户动态
const loadUserDynamics = async () => {
  if (!props.userId) return

  props.loading.dynamics = true
  try {
    const res = await dynamicApi.getUserDynamics(props.userId, 1, 20)
    if (res.code === 200) {
      // 后端返回的是 DynamicVO，直接映射到前端格式
      dynamics.value = (res.data || []).map(item => ({
        id: item.id,
        dynamicType: item.dynamicType,
        user: item.user ? {
          id: item.user.id,
          name: item.user.nickname || item.user.username,
          avatar: item.user.avatar
        } : {
          id: props.userId,
          name: props.userInfo.username,
          avatar: props.userInfo.avatar
        },
        createTime: item.createdAt,
        content: item.content,
        isTop: false,
        images: item.imageUrls || [],
        // 引用稿件：可能为空，不一定每条动态都带视频
        video: item.refManuscriptId ? {
          id: item.refManuscriptId,
          title: item.refVideo?.title || '',
          cover: item.refVideo?.cover || '',
          duration: item.refVideo?.duration || '',
          views: item.refVideo?.viewCount || 0
        } : null,
        stats: {
          shareCount: item.shareCount || 0,
          commentCount: item.commentCount || 0,
          likeCount: item.likeCount || 0,
          isLiked: item.isLiked || false
        }
      }))
    }
  } catch (error) {
    console.error('加载用户动态失败:', error)
  } finally {
    props.loading.dynamics = false
  }
}

const loadData = () => {
  loadUserDynamics()
}

onMounted(() => {
  loadData()
})

watch(() => props.userId, () => {
  loadData()
})
</script>

<template>
  <div class="dynamic-layout">
    <!-- 左侧筛选栏：全部 / 视频 -->
    <aside class="filter-sidebar">
      <div class="filter-card">
        <div
          v-for="opt in filterOptions"
          :key="opt.key"
          :class="['filter-item', { active: activeFilter === opt.key }]"
          @click="activeFilter = opt.key"
        >
          {{ opt.label }}
        </div>
      </div>
    </aside>

    <!-- 动态流 -->
    <div class="dynamic-feed">
      <div v-if="loading.dynamics" class="loading-state">
        <p>加载中...</p>
      </div>
      <div v-else-if="filteredDynamics.length === 0" class="empty-state">
        <p>暂无动态</p>
      </div>
      <div v-else class="dynamic-list">
        <!-- 动态列表 -->
        <div v-for="dynamic in filteredDynamics" :key="dynamic.id" class="dynamic-card">
          <!-- 置顶标记 -->
          <div v-if="dynamic.isTop" class="top-badge">
            <el-icon><Top /></el-icon>
            <span>置顶</span>
          </div>

          <!-- 头部 -->
          <div class="dynamic-header-new">
            <img loading="lazy" decoding="async" :src="dynamic.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" :alt="dynamic.user?.name" class="user-avatar" @click="router.push(`/profile/${dynamic.user?.id}/home`)">
            <div class="user-info">
              <div class="user-name">{{ dynamic.user?.name || '未知用户' }}</div>
              <div class="dynamic-time">{{ formatHeaderLine(dynamic) }}</div>
            </div>
            <!-- 更多菜单 -->
            <el-dropdown trigger="click" @command="(cmd) => cmd === 'stick' ? handleStickDynamic(dynamic.id) : handleDeleteDynamic(dynamic.id)">
              <button class="more-btn" @click.stop>
                <el-icon><MoreFilled /></el-icon>
              </button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="'stick'">
                    <el-icon><Top /></el-icon>
                    <span>{{ dynamic.isTop ? '取消置顶' : '置顶' }}</span>
                  </el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>
                    <el-icon><Delete /></el-icon>
                    <span>删除</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>

          <!-- 正文 -->
          <div class="dynamic-content-new">
            <!-- 正文文字 -->
            <div v-if="dynamic.content" class="original-content">
              <div class="original-text">{{ dynamic.content }}</div>
            </div>

            <!-- 图文：图片网格 -->
            <div v-if="dynamic.images && dynamic.images.length" class="original-images">
              <img
                loading="lazy"
                decoding="async"
                v-for="(img, idx) in dynamic.images"
                :key="idx"
                :src="img"
                class="original-image"
                @click="previewImage(img)"
              >
            </div>

            <!-- 带视频：不一定每条动态都有，仅有引用稿件时渲染 -->
            <div v-if="dynamic.video" class="video-card" @click="openVideo(dynamic.video.id)">
              <div class="video-cover-wrapper">
                <img loading="lazy" decoding="async" v-if="dynamic.video.cover" :src="dynamic.video.cover" class="video-cover">
                <div v-else class="video-cover-placeholder"></div>
                <div v-if="dynamic.video.duration" class="video-duration">{{ dynamic.video.duration }}</div>
              </div>
              <div class="video-info-new">
                <div class="video-title-new">{{ dynamic.video.title }}</div>
                <div class="video-stats">
                  <span class="video-stat-item">
                    <svg class="stat-icon" viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M4 6h13v12H4zM19 8l3-2v12l-3-2z"/></svg>
                    {{ formatNumber(dynamic.video.views) }}
                  </span>
                  <span class="video-stat-item">
                    <el-icon><ChatDotRound /></el-icon>
                    {{ dynamic.stats?.commentCount || 0 }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 操作栏 -->
          <div class="dynamic-actions-new">
            <button class="action-btn-new" @click="ElMessage.info('转发功能开发中')">
              <el-icon><Share /></el-icon>
              <span>{{ dynamic.stats?.shareCount > 0 ? dynamic.stats.shareCount : '转发' }}</span>
            </button>
            <button class="action-btn-new" :class="{ 'is-active': expandedComments.has(dynamic.id) }" @click="toggleComments(dynamic)">
              <el-icon><ChatDotRound /></el-icon>
              <span>{{ dynamic.stats?.commentCount > 0 ? `${dynamic.stats.commentCount} 条评论` : '评论' }}</span>
            </button>
            <button class="action-btn-new" :class="{ 'is-liked': dynamic.stats?.isLiked }" @click="handleLikeDynamic(dynamic)">
              <svg class="thumb-icon" viewBox="0 0 24 24" width="16" height="16"><path fill="currentColor" d="M2 20h2V9H2v11zm20-10c0-1.1-.9-2-2-2h-6.31l.95-4.57.03-.32c0-.41-.17-.79-.44-1.06L13.17 2 7.59 7.59C7.22 7.95 7 8.45 7 9v9c0 1.1.9 2 2 2h9c.83 0 1.54-.5 1.84-1.22l3.02-7.05c.09-.23.14-.47.14-.73v-1z"/></svg>
              <span>{{ dynamic.stats?.likeCount > 0 ? dynamic.stats.likeCount : '点赞' }}</span>
            </button>
          </div>

          <!-- 评论展开区域 -->
          <div v-if="expandedComments.has(dynamic.id)" class="comment-section">
            <CommentSystem
              :target-type="'DYNAMIC'"
              :target-id="dynamic.id"
              placeholder="发一条友善的评论吧~"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 整体布局：左筛选 + 动态流 */
.dynamic-layout {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

/* 左侧筛选栏 */
.filter-sidebar {
  width: 148px;
  flex-shrink: 0;
  position: sticky;
  top: 84px;
}

.filter-card {
  background: #fff;
  border-radius: 6px;
  padding: 10px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.filter-item {
  padding: 12px 0;
  text-align: center;
  font-size: 15px;
  color: #61666d;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-item:hover {
  color: #00a1d6;
}

.filter-item.active {
  background: #00a1d6;
  color: #fff;
  font-weight: 500;
}

/* 动态流 */
.dynamic-feed {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 动态卡片 */
.dynamic-card {
  background: #fff;
  border-radius: 6px;
  padding: 20px;
  position: relative;
}

/* 置顶标记 */
.top-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #00a1d6;
  margin-bottom: 12px;
}

.top-badge .el-icon {
  font-size: 14px;
}

/* 动态头部 */
.dynamic-header-new {
  display: flex;
  align-items: center;
}

.dynamic-header-new .user-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  margin-right: 12px;
  object-fit: cover;
  cursor: pointer;
}

.dynamic-header-new .user-info {
  flex: 1;
}

.dynamic-header-new .user-name {
  font-size: 15px;
  font-weight: 500;
  color: #212121;
}

.dynamic-header-new .dynamic-time {
  font-size: 13px;
  color: #9499a0;
  margin-top: 4px;
}

.dynamic-header-new .more-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: #9499a0;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dynamic-header-new .more-btn:hover {
  background: #f1f2f3;
  color: #61666d;
}

/* 内容区域 */
.dynamic-content-new {
  margin-left: 60px;
  margin-top: 12px;
}

/* 正文 */
.original-content {
  margin-top: 0;
}

.original-text {
  font-size: 15px;
  color: #212121;
  line-height: 1.6;
  white-space: pre-line;
  word-break: break-word;
}

/* 图片网格 */
.original-images {
  display: grid;
  grid-template-columns: repeat(3, 150px);
  gap: 8px;
  margin-top: 12px;
}

.original-image {
  width: 150px;
  height: 150px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  background: #f1f2f3;
}

.original-image:hover {
  opacity: 0.9;
}

/* 视频卡：白底边框横排（不一定每条动态都有） */
.video-card {
  display: flex;
  margin-top: 12px;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  transition: background 0.2s;
}

.video-card:hover {
  background: #f7f8fa;
}

.video-cover-wrapper {
  position: relative;
  width: 240px;
  height: 135px;
  flex-shrink: 0;
}

.video-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.video-cover-placeholder {
  width: 100%;
  height: 100%;
  background: #e3e5e7;
}

.video-duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 13px;
  padding: 1px 6px;
  border-radius: 4px;
}

.video-info-new {
  flex: 1;
  min-width: 0;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
}

.video-title-new {
  font-size: 15px;
  font-weight: 500;
  color: #212121;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.video-stats {
  display: flex;
  gap: 24px;
  font-size: 13px;
  color: #9499a0;
}

.video-stat-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.stat-icon {
  display: inline-block;
}

/* 操作栏：三等分 */
.dynamic-actions-new {
  display: flex;
  margin-left: 60px;
  margin-top: 16px;
  border-top: 1px solid #e3e5e7;
}

.action-btn-new {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px 0;
  border: none;
  background: transparent;
  color: #61666d;
  font-size: 14px;
  cursor: pointer;
  border-radius: 4px;
  transition: color 0.2s;
}

.action-btn-new:hover {
  color: #00a1d6;
}

.action-btn-new.is-liked {
  color: #00a1d6;
}

.action-btn-new.is-active {
  color: #00a1d6;
}

.action-btn-new .el-icon {
  font-size: 18px;
}

.thumb-icon {
  display: inline-block;
}

/* 加载状态和空状态 */
.loading-state,
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 60px 0;
  color: #9499a0;
  font-size: 14px;
  background: #fff;
  border-radius: 6px;
}

/* 评论展开区域样式 */
.comment-section {
  margin-top: 16px;
  padding: 16px;
  background-color: #f6f7f8;
  border-radius: 6px;
}

/* 窄屏：筛选栏收起为顶部横条 */
@media (max-width: 900px) {
  .dynamic-layout {
    flex-direction: column;
  }

  .filter-sidebar {
    width: 100%;
    position: static;
  }

  .filter-card {
    display: flex;
    gap: 8px;
  }

  .filter-item {
    padding: 8px 24px;
  }
}
</style>
