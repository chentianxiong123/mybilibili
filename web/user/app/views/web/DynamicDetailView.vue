<template>
  <div class="dynamic-detail-page">
    <div class="content">
      <div class="card">
        <!-- 动态内容卡（对齐 B 站 bili-dyn-item：左侧 88px 头像槽 + 右 20px） -->
        <div class="dynamic-card" v-if="dynamic">
          <div class="dyn-avatar-wrap">
            <img loading="lazy" decoding="async"
              :src="dynamic.user?.avatar || defaultAvatar"
              alt=""
              class="user-avatar"
              @click="goToUser(dynamic.userId)"
            >
          </div>

          <div class="dynamic-main">
            <!-- 头部：用户名 + 时间 + 更多 -->
            <div class="dynamic-header">
              <div class="dyn-title" @click="goToUser(dynamic.userId)">
                <span class="username">{{ dynamic.user?.username || '用户' }}</span>
              </div>
              <div class="dyn-desc">
                <span class="publish-time">{{ formatTime(dynamic.createdAt) }}</span>
              </div>
              <div class="dyn-more" @click="handleMore">
                <el-icon :size="18"><MoreFilled /></el-icon>
              </div>
            </div>

            <!-- 正文 -->
            <div class="dynamic-body">
              <div class="dynamic-text">{{ dynamic.content }}</div>

              <!-- 图片九宫格 -->
              <div v-if="dynamic.imageUrls && dynamic.imageUrls.length > 0" class="dynamic-images">
                <img loading="lazy" decoding="async"
                  v-for="(url, index) in dynamic.imageUrls"
                  :key="index"
                  :src="url"
                  alt=""
                  class="dynamic-image"
                  @click="previewImage(url)"
                >
              </div>

              <!-- 引用视频 -->
              <div v-if="dynamic.refManuscriptId" class="video-card" @click="goToVideo(dynamic.refManuscriptId)">
                <div class="video-info">
                  <el-icon :size="24"><VideoPlay /></el-icon>
                  <span>引用了视频</span>
                </div>
                <div class="video-hint">点击跳转到视频页面</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 评论 / 赞与转发 Tabs（对齐 B 站 bili-tabs.dyn-tabs） -->
        <div class="dyn-tabs">
          <div class="tabs-header">
            <div class="tabs-nav">
              <div
                class="tab-item"
                :class="{ 'is-active': activeTab === 'comment' }"
                @click="activeTab = 'comment'"
              >
                评论
              </div>
              <div
                class="tab-item"
                :class="{ 'is-active': activeTab === 'interact' }"
                @click="activeTab = 'interact'"
              >
                赞与转发
              </div>
              <div class="tab-active-bar" :style="activeBarStyle"></div>
            </div>
          </div>

          <div class="tabs-content">
            <!-- 评论 tab -->
            <div v-show="activeTab === 'comment'" class="tab-pane comment-pane">
              <CommentSystem
                ref="commentSystemRef"
                :target-type="actualTargetType"
                :target-id="actualTargetId"
                :placeholder="commentPlaceholder"
              />
            </div>

            <!-- 赞与转发 tab -->
            <div v-show="activeTab === 'interact'" class="tab-pane interact-pane">
              <div class="interact-actions">
                <button class="interact-btn" :class="{ active: dynamic?.isLiked }" @click="handleLike">
                  <el-icon :size="20"><Star /></el-icon>
                  <span>{{ dynamic?.likeCount || '点赞' }}</span>
                </button>
                <button class="interact-btn" @click="handleShare">
                  <el-icon :size="20"><Share /></el-icon>
                  <span>{{ dynamic?.shareCount || '分享' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { safeStorage } from '@/utils/safeStorage'
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Star, Share, VideoPlay, MoreFilled } from '@element-plus/icons-vue'
import { dynamicApi } from '@/api/dynamic.ts'
import CommentSystem from '@/components/CommentSystem.vue'

const route = useRoute()
const router = useRouter()

const dynamicId = computed(() => parseInt(route.params.id))
const dynamic = ref(null)
const loading = ref(false)
const activeTab = ref('comment')

const defaultAvatar = '/default-avatar.svg'
const commentSystemRef = ref(null)

// 是否使用视频评论区（当动态引用了视频时）
const useVideoComments = computed(() => {
  return !!dynamic.value?.refManuscriptId
})

// 实际使用的 targetType
const actualTargetType = computed(() => {
  return useVideoComments.value ? 'VIDEO' : 'DYNAMIC'
})

// 实际使用的 targetId
const actualTargetId = computed(() => {
  return useVideoComments.value ? dynamic.value.refManuscriptId : dynamicId.value
})

// 评论输入框占位符
const commentPlaceholder = computed(() => {
  return useVideoComments.value ? '对该视频发表看法...' : '发一条友善的评论吧~'
})

// tab 下划线游标位移（评论 tab 宽 36 + margin 30 → 赞与转发偏移 66）
const activeBarStyle = computed(() => {
  return activeTab.value === 'comment'
    ? { transform: 'translateX(0)' }
    : { transform: 'translateX(66px)' }
})

// 获取动态详情
const fetchDynamicDetail = async () => {
  loading.value = true
  try {
    const res = await dynamicApi.getDynamicById(dynamicId.value)
    if (res.code === 200) {
      dynamic.value = res.data
    } else {
      ElMessage.error(res.message || '获取动态详情失败')
    }
  } catch (error) {
    console.error('获取动态详情失败:', error)
    ElMessage.error('获取动态详情失败')
  } finally {
    loading.value = false
  }
}

// 点赞/取消点赞
const handleLike = async () => {
  if (!safeStorage.getItem('user')) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    if (dynamic.value.isLiked) {
      const res = await dynamicApi.unlikeDynamic(dynamicId.value)
      if (res.code === 200) {
        dynamic.value.isLiked = false
        dynamic.value.likeCount = Math.max(0, dynamic.value.likeCount - 1)
      }
    } else {
      const res = await dynamicApi.likeDynamic(dynamicId.value)
      if (res.code === 200) {
        dynamic.value.isLiked = true
        dynamic.value.likeCount = (dynamic.value.likeCount || 0) + 1
      }
    }
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  }
}

// 分享
const handleShare = async () => {
  if (!safeStorage.getItem('user')) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    const res = await dynamicApi.shareDynamic(dynamicId.value)
    if (res.code === 200) {
      dynamic.value.shareCount = (dynamic.value.shareCount || 0) + 1
      ElMessage.success('分享成功')
    }
  } catch (error) {
    console.error('分享失败:', error)
    ElMessage.error('分享失败')
  }
}

// 更多操作（占位）
const handleMore = () => {}

// 跳转到用户主页
const goToUser = (userId) => {
  if (userId) {
    router.push(`/user/${userId}`)
  }
}

// 跳转到视频
const goToVideo = (manuscriptId) => {
  router.push(`/manuscript/${manuscriptId}`)
}

// 预览图片
const previewImage = (url) => {
  window.open(url, '_blank')
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
  if (diff < 604800000) return Math.floor(diff / 86400000) + '天前'
  return date.toLocaleDateString('zh-CN')
}

onMounted(() => {
  fetchDynamicDetail()
})
</script>

<style scoped>
/* 页面：对齐 B 站 t.bilibili.com 动态详情 —— 通栏 #d3e9e8 背景，内容起点 y72 */
.dynamic-detail-page {
  min-height: 100vh;
  background: #d3e9e8;
  padding-top: 72px;
  box-sizing: border-box;
}

/* 632px 居中单列（B 站 .content 632 @x644 居中） */
.content {
  width: 632px;
  margin: 0 auto;
}

/* 白卡：顶角 6px 圆角、底部留白 100px（B 站 .card） */
.card {
  background: #fff;
  border-radius: 6px 6px 0 0;
  padding-bottom: 100px;
}

/* 动态项：左侧 88px 头像槽 + 右 20px（B 站 bili-dyn-item__main padding 0 20 0 88） */
.dynamic-card {
  position: relative;
  background: #fff;
  border-radius: 6px;
  padding: 0 20px 0 88px;
}

/* 头像槽：绝对定位左侧 86×86，内部 48 头像居中（偏移 19,19） */
.dyn-avatar-wrap {
  position: absolute;
  left: 0;
  top: 0;
  width: 86px;
  height: 86px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  cursor: pointer;
}

.dynamic-main {
  min-width: 0;
}

/* 头部：padding-top 16（B 站 bili-dyn-item__header） */
.dynamic-header {
  position: relative;
  padding-top: 16px;
}

/* 用户名：17px/600/32 #18191c（B 站 bili-dyn-title__text fs-large） */
.dyn-title {
  display: flex;
  margin: 1px 0 2px;
  cursor: pointer;
}

.username {
  font-size: 17px;
  font-weight: 600;
  line-height: 32px;
  color: #18191c;
  transition: color 0.2s;
}

.username:hover {
  color: #00aeec;
}

/* 时间：13px/18 #9499a0（B 站 bili-dyn-time） */
.dyn-desc {
  display: flex;
}

.publish-time {
  font-size: 13px;
  line-height: 18px;
  color: #9499a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 更多按钮：右上 24×24（B 站 bili-dyn-item__more） */
.dyn-more {
  position: absolute;
  right: 0;
  top: 18px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9499a0;
  cursor: pointer;
}

.dyn-more:hover {
  color: #00aeec;
}

/* 正文：margin-top 8，15px/25 #18191c（B 站 bili-dyn-item__body + opus 正文） */
.dynamic-body {
  margin-top: 8px;
}

.dynamic-text {
  font-size: 15px;
  line-height: 25px;
  color: #18191c;
  white-space: pre-wrap;
  word-break: break-word;
}

/* 图片网格：3 列方形（B 站图片动态） */
.dynamic-images {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin-top: 8px;
}

.dynamic-image {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  border-radius: 6px;
  cursor: pointer;
}

/* 引用视频卡片 */
.video-card {
  margin-top: 8px;
  padding: 16px;
  background: #f1f2f3;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.3s;
}

.video-card:hover {
  background: #e3e5e7;
}

.video-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #00aeec;
  font-size: 14px;
}

.video-hint {
  font-size: 12px;
  color: #9499a0;
  margin-top: 4px;
}

/* Tabs（B 站 bili-tabs.dyn-tabs margin-top 20） */
.dyn-tabs {
  margin-top: 20px;
}

/* tab 头：白底，height 48，padding-left 26（B 站 bili-tabs__header） */
.tabs-header {
  position: relative;
  height: 48px;
  padding-left: 26px;
  background: #fff;
  border-bottom: 1px solid #e3e5e7;
}

.tabs-nav {
  position: relative;
  display: flex;
  align-items: center;
  height: 48px;
}

/* tab 项：18px/500/25，首项 margin-left 30（B 站 bili-tabs__nav__item） */
.tab-item {
  display: flex;
  align-items: center;
  margin-left: 30px;
  font-size: 18px;
  font-weight: 500;
  line-height: 25px;
  color: #61666d;
  cursor: pointer;
  transition: color 0.2s;
}

.tab-item.is-active {
  color: #00aede;
}

.tab-item:hover {
  color: #00aede;
}

/* 激活游标：36×2 #00aede（B 站 bili-tabs__active-bar，left 30 对齐首 tab） */
.tab-active-bar {
  position: absolute;
  left: 30px;
  bottom: 0;
  width: 36px;
  height: 2px;
  background: #00aede;
  transition: transform 0.3s cubic-bezier(0.645, 0.045, 0.355, 1);
}

/* tab 内容：左右 padding 24（584 内容宽 / 632 容器） */
.tabs-content {
  padding: 16px 24px 0;
}

.tab-pane {
  min-height: 80px;
}

/* 赞与转发操作区 */
.interact-pane {
  padding: 8px 0;
}

.interact-actions {
  display: flex;
  gap: 16px;
}

.interact-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 24px;
  background: #f1f2f3;
  border: none;
  border-radius: 6px;
  color: #61666d;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.3s, color 0.3s;
}

.interact-btn:hover,
.interact-btn.active {
  background: #e3e5e7;
  color: #00aeec;
}
</style>
