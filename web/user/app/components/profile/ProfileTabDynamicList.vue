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
  <div class="space-dynamic">
    <!-- 左侧筛选栏（B站 side-nav）：全部 / 视频 -->
    <aside class="space-dynamic__left">
      <div class="side-nav">
        <div
          v-for="opt in filterOptions"
          :key="opt.key"
          :class="['side-nav__item', { active: activeFilter === opt.key }]"
          @click="activeFilter = opt.key"
        >
          <span class="side-nav__item-text">{{ opt.label }}</span>
        </div>
      </div>
    </aside>

    <!-- 动态流 -->
    <div class="space-dynamic__content">
      <div v-if="loading.dynamics" class="loading-state">
        <p>加载中...</p>
      </div>
      <div v-else-if="filteredDynamics.length === 0" class="empty-state">
        <p>暂无动态</p>
      </div>
      <div v-else class="dynamic-list">
        <!-- 动态列表 -->
        <div v-for="dynamic in filteredDynamics" :key="dynamic.id" class="bili-dyn-item">
          <!-- 置顶标记 -->
          <div v-if="dynamic.isTop" class="top-badge">
            <el-icon><Top /></el-icon>
            <span>置顶</span>
          </div>

          <!-- 头部（B站：88px 头像槽绝对定位 + header 名称/时间 + more 右上） -->
          <div class="bili-dyn-item__avatar" @click="router.push(`/profile/${dynamic.user?.id}/home`)">
            <img loading="lazy" decoding="async" :src="dynamic.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" :alt="dynamic.user?.name" class="user-avatar">
          </div>
          <div class="bili-dyn-item__header">
            <div class="bili-dyn-title">
              <span class="bili-dyn-title__text">{{ dynamic.user?.name || '未知用户' }}</span>
            </div>
            <div class="bili-dyn-item__desc">
              <div class="bili-dyn-time">{{ formatHeaderLine(dynamic) }}</div>
            </div>
            <!-- 更多菜单 -->
            <el-dropdown trigger="click" @command="(cmd) => cmd === 'stick' ? handleStickDynamic(dynamic.id) : handleDeleteDynamic(dynamic.id)">
              <button class="bili-dyn-more" @click.stop>
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
          <div class="bili-dyn-item__body">
            <div class="bili-dyn-content">
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
          </div>

          <!-- 操作栏（B站 footer：50 高，转发/评论/点赞 三等分，无上边框） -->
          <div class="bili-dyn-item__footer">
            <div class="bili-dyn-item__action">
              <button class="bili-dyn-action forward" @click="ElMessage.info('转发功能开发中')">
                <el-icon><Share /></el-icon>
                <span>{{ dynamic.stats?.shareCount > 0 ? dynamic.stats.shareCount : '转发' }}</span>
              </button>
            </div>
            <div class="bili-dyn-item__action">
              <button class="bili-dyn-action comment" :class="{ 'is-active': expandedComments.has(dynamic.id) }" @click="toggleComments(dynamic)">
                <el-icon><ChatDotRound /></el-icon>
                <span>{{ dynamic.stats?.commentCount > 0 ? `${dynamic.stats.commentCount} 条评论` : '评论' }}</span>
              </button>
            </div>
            <div class="bili-dyn-item__action">
              <button class="bili-dyn-action like" :class="{ 'is-liked': dynamic.stats?.isLiked }" @click="handleLikeDynamic(dynamic)">
                <svg class="thumb-icon" viewBox="0 0 24 24" width="16" height="16"><path fill="currentColor" d="M2 20h2V9H2v11zm20-10c0-1.1-.9-2-2-2h-6.31l.95-4.57.03-.32c0-.41-.17-.79-.44-1.06L13.17 2 7.59 7.59C7.22 7.95 7 8.45 7 9v9c0 1.1.9 2 2 2h9c.83 0 1.54-.5 1.84-1.22l3.02-7.05c.09-.23.14-.47.14-.73v-1z"/></svg>
                <span>{{ dynamic.stats?.likeCount > 0 ? dynamic.stats.likeCount : '点赞' }}</span>
              </button>
            </div>
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
/* 整体布局（B站 .space-dynamic 1625@x140 flex，margin-top 30）：左侧栏 240 + 内容列 + 右栏 300 */
.space-dynamic {
  display: flex;
  align-items: flex-start;
  width: 100%;
  margin-top: 30px;
}

/* 左侧筛选栏（B站 .space-dynamic__left 240@x140 > .side-nav） */
.space-dynamic__left {
  width: 240px;
  flex-shrink: 0;
  margin-right: 16px;
}

/* B站 .side-nav：透明无白卡，项 240x52 r6 pad14 16 */
.side-nav {
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.side-nav__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 240px;
  height: 52px;
  padding: 14px 16px;
  margin-top: 8px;
  font-size: 16px;
  font-weight: 400;
  color: #18191c;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  box-sizing: border-box;
}

.side-nav__item:first-child {
  margin-top: 0;
}

.side-nav__item:hover {
  background: #f6f7f8;
}

/* B站 .side-nav__item.active：实底 #00aeec 白字 */
.side-nav__item.active {
  background: #00aeec;
  color: #fff;
  font-weight: 400;
}

/* 内容列（B站 .space-dynamic__content 1053@x396，卡片 1001，右侧 300 rail @x1465） */
.space-dynamic__content {
  width: 1053px;
  max-width: 100%;
  flex-shrink: 0;
  margin-right: 16px;
  min-width: 0;
}

.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 动态卡片（B站 .bili-dyn-item 1001 白卡 r6 border#e3e5e7，__main pad 0 20 0 88 = 88px 头像槽） */
.bili-dyn-item {
  position: relative;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 6px;
  padding: 0 20px 0 88px;
  box-sizing: border-box;
}

/* 置顶标记 */
.top-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #00aeec;
  margin-bottom: 12px;
}

.top-badge .el-icon {
  font-size: 14px;
}

/* 头像槽（B站 .bili-dyn-item__avatar 86x86 绝对定位卡片左上角，内含 48 圆头像） */
.bili-dyn-item__avatar {
  position: absolute;
  top: 0;
  left: 0;
  width: 86px;
  height: 86px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.bili-dyn-item__avatar .user-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

/* 头部（B站 .bili-dyn-item__header pad-top16，名称 17/600/lh32 + 时间 13/18 #9499a0） */
.bili-dyn-item__header {
  position: relative;
  padding-top: 16px;
}

.bili-dyn-title {
  display: flex;
  align-items: center;
  margin: 1px 0 2px;
}

.bili-dyn-title__text {
  font-size: 17px;
  font-weight: 600;
  line-height: 32px;
  color: #18191c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bili-dyn-item__desc {
  display: flex;
  align-items: center;
}

.bili-dyn-time {
  font-size: 13px;
  font-weight: 400;
  line-height: 18px;
  color: #9499a0;
}

/* 更多按钮（B站 .bili-dyn-item__more 24x24 右上） */
.bili-dyn-more {
  position: absolute;
  top: 16px;
  right: 0;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: #9499a0;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

/* 操作栏（B站 .bili-dyn-item__footer 890x50 flex space-between，无上边框，pad-right 20） */
.bili-dyn-item__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 50px;
  margin-top: 12px;
  padding-right: 20px;
}

.bili-dyn-item__action {
  display: flex;
  align-items: center;
}

.bili-dyn-action {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 50px;
  gap: 6px;
  padding: 0;
  border: none;
  background: transparent;
  color: #61666d;
  font-size: 14px;
  cursor: pointer;
  border-radius: 4px;
  transition: color 0.2s;
}

.bili-dyn-action:hover {
  color: #00aeec;
}

.bili-dyn-action.is-liked {
  color: #00aeec;
}

.bili-dyn-action.is-active {
  color: #00aeec;
}

.bili-dyn-action .el-icon {
  font-size: 18px;
}

/* 正文文字（B站 bili-dyn-content__orig 16px #18191c） */
.original-text {
  font-size: 16px;
  color: #18191c;
  line-height: 25px;
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
  .space-dynamic {
    flex-direction: column;
  }

  .space-dynamic__left {
    width: 100%;
    margin-right: 0;
    margin-bottom: 16px;
  }

  .side-nav {
    display: flex;
    gap: 8px;
  }

  .side-nav__item {
    width: auto;
    margin-top: 0;
    padding: 8px 24px;
  }
}
</style>
