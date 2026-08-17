<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Top, MoreFilled, Delete, VideoPlay, Share, ChatDotRound, Star } from '@element-plus/icons-vue'
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
    dynamics.value = dynamics.value.filter(d => d.id !== dynamicId)
    ElMessage.success('删除成功')
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

// 格式化时间
const formatDynamicTime = (timeStr) => {
  const date = new Date(timeStr)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}月${day}日`
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
        type: item.dynamicType === 2 ? 'video' : (item.dynamicType === 1 ? 'original' : 'original'),
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
        video: item.refVideoId ? {
          id: item.refVideoId,
          title: '引用视频',
          cover: '',
          duration: '',
          views: 0
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
  <div class="dynamic-section">
    <div v-if="loading.dynamics" class="loading-state">
      <p>加载中...</p>
    </div>
    <div v-else-if="dynamics.length === 0" class="empty-state">
      <p>暂无动态</p>
    </div>
    <div v-else class="dynamic-list">
      <!-- 动态列表 -->
      <div v-for="dynamic in dynamics" :key="dynamic.id" :class="['dynamic-card', { 'is-top': dynamic.isTop }]">
        <!-- 置顶标记 -->
        <div v-if="dynamic.isTop" class="top-badge">
          <el-icon><Top /></el-icon>
          <span>置顶</span>
        </div>

        <!-- 头部 -->
        <div class="dynamic-header-new">
          <img loading="lazy" decoding="async" :src="dynamic.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" :alt="dynamic.user?.name" class="user-avatar">
          <div class="user-info">
            <div class="user-name">{{ dynamic.user?.name || '未知用户' }}</div>
            <div class="dynamic-time">{{ formatDynamicTime(dynamic.createTime) }}</div>
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

        <!-- 动态类型标签 -->
        <div v-if="dynamic.type === 'share'" class="dynamic-type-label">
          转发动态
        </div>

        <!-- 动态内容 -->
        <div class="dynamic-content-new">
          <!-- 转发动态 -->
          <div v-if="dynamic.type === 'share' && dynamic.shareSource" class="share-card">
            <div class="share-source">
              <img loading="lazy" decoding="async" :src="dynamic.shareSource?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" class="source-avatar">
              <span class="source-name">{{ dynamic.shareSource?.name || '未知来源' }}</span>
            </div>
            <div class="share-content-wrapper">
              <div class="share-text" :class="{ 'is-collapsed': !expandedDynamics.has(dynamic.id) && dynamic.shareSource?.hasMore }">
                {{ dynamic.shareSource?.content || '' }}
              </div>
              <button v-if="dynamic.shareSource?.hasMore" class="expand-btn" @click="toggleExpand(dynamic.id)">
                {{ expandedDynamics.has(dynamic.id) ? '收起' : '展开' }}
              </button>
            </div>
            <div v-if="dynamic.shareSource?.images && dynamic.shareSource.images.length" class="share-images">
              <img loading="lazy" decoding="async" v-for="(img, idx) in dynamic.shareSource.images" :key="idx" :src="img" class="share-image">
            </div>
          </div>

          <!-- 视频动态 -->
          <div v-else-if="dynamic.type === 'video' && dynamic.video" class="video-dynamic-new" @click="router.push(`/manuscript/${dynamic.video.id}`)">
            <div class="video-card">
              <div class="video-cover-wrapper">
                <img loading="lazy" decoding="async" :src="dynamic.video?.cover" class="video-cover">
                <div class="video-duration">{{ dynamic.video?.duration }}</div>
                <div class="video-play-icon">
                  <el-icon><VideoPlay /></el-icon>
                </div>
              </div>
              <div class="video-info-new">
                <div class="video-title-new">{{ dynamic.video?.title }}</div>
                <div class="video-views-new">{{ dynamic.video?.views }} 播放</div>
              </div>
            </div>
          </div>

          <!-- 原创内容 -->
          <div v-else class="original-content">
            <div class="original-text">{{ dynamic.content }}</div>
            <div v-if="dynamic.images && dynamic.images.length" class="original-images">
              <img loading="lazy" decoding="async" v-for="(img, idx) in dynamic.images" :key="idx" :src="img" class="original-image">
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
            <el-icon><Star /></el-icon>
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
</template>

<style scoped>
/* ==================== 动态页面新样式 ==================== */

/* 动态列表容器 */
.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 动态卡片 */
.dynamic-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  position: relative;
}

.dynamic-card.is-top {
  border: 1px solid #00aeec;
}

/* 置顶标记 */
.top-badge {
  position: absolute;
  top: 12px;
  right: 48px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #00aeec;
  background: rgba(0, 174, 236, 0.1);
  padding: 2px 8px;
  border-radius: 12px;
}

.top-badge .el-icon {
  font-size: 12px;
}

/* 动态头部 */
.dynamic-header-new {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.dynamic-header-new .user-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  margin-right: 12px;
  object-fit: cover;
}

.dynamic-header-new .user-info {
  flex: 1;
}

.dynamic-header-new .user-name {
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.dynamic-header-new .dynamic-time {
  font-size: 13px;
  color: #999;
  margin-top: 4px;
}

.dynamic-header-new .more-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: #999;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dynamic-header-new .more-btn:hover {
  background: #f5f5f5;
  color: #666;
}

/* 动态类型标签 */
.dynamic-type-label {
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
  margin-left: 60px;
}

/* 动态内容区域 */
.dynamic-content-new {
  margin-left: 60px;
}

/* 转发卡片 */
.share-card {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px;
  margin-top: 8px;
}

.share-source {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.source-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  margin-right: 8px;
  object-fit: cover;
}

.source-name {
  font-size: 14px;
  color: #00aeec;
  font-weight: 500;
}

.share-content-wrapper {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
}

.share-text {
  white-space: pre-line;
}

.share-text.is-collapsed {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.expand-btn {
  color: #00aeec;
  font-size: 13px;
  cursor: pointer;
  background: none;
  border: none;
  padding: 0;
  margin-top: 8px;
}

.expand-btn:hover {
  color: #0095d9;
}

/* 转发图片 */
.share-images {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.share-image {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
}

/* 视频动态新样式 */
.video-dynamic-new {
  margin-top: 8px;
}

.video-card {
  display: flex;
  background: #f5f7fa;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
}

.video-cover-wrapper {
  position: relative;
  width: 160px;
  height: 90px;
  flex-shrink: 0;
}

.video-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-duration {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 2px;
}

.video-play-icon {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 40px;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
}

.video-info-new {
  flex: 1;
  padding: 12px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.video-title-new {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 8px;
  line-height: 1.4;
}

.video-views-new {
  font-size: 13px;
  color: #999;
}

/* 原创内容 */
.original-content {
  margin-top: 8px;
}

.original-text {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
  white-space: pre-line;
}

.original-images {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.original-image {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
}

/* 操作栏新样式 */
.dynamic-actions-new {
  display: flex;
  justify-content: space-around;
  padding-top: 16px;
  margin-top: 16px;
  margin-left: 60px;
  border-top: 1px solid #f0f0f0;
}

.action-btn-new {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 24px;
  border: none;
  background: transparent;
  color: #666;
  font-size: 13px;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.3s;
}

.action-btn-new:hover {
  background: #f5f5f5;
  color: #00aeec;
}

.action-btn-new.is-liked {
  color: #00aeec;
}

.action-btn-new.is-active {
  color: #409eff;
}

.action-btn-new .el-icon {
  font-size: 16px;
}

/* 动态页面响应式设计 */
@media (max-width: 768px) {
  .dynamic-content-new {
    margin-left: 0;
  }

  .dynamic-type-label {
    margin-left: 0;
  }

  .dynamic-actions-new {
    margin-left: 0;
  }

  .share-image,
  .original-image {
    width: calc(33.333% - 6px);
    height: auto;
    aspect-ratio: 1;
  }

  .video-card {
    flex-direction: column;
  }

  .video-cover-wrapper {
    width: 100%;
    height: auto;
    aspect-ratio: 16/9;
  }
}

/* 动态页面样式 */
.dynamic-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

/* 评论展开区域样式 */
.comment-section {
  margin-top: 16px;
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 8px;
}
</style>