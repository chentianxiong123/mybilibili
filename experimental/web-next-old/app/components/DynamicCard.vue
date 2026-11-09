<script setup>
import { MoreFilled, Share, ChatDotRound, Star, VideoPlay, Clock, View } from '@element-plus/icons-vue'
import CommentSystem from '@/components/CommentSystem.vue'

const props = defineProps({
  item: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['like', 'forward', 'toggle-comment', 'go-to-user', 'go-to-manuscript'])

const formatTime = (dateStr) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 30) return `${days}天前`
  return date.toLocaleDateString()
}

const formatNumber = (num) => {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}
</script>

<template>
  <div class="dynamic-card">
    <div class="dynamic-header">
      <img loading="lazy" decoding="async" :src="item.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" alt="" class="dynamic-avatar" @click="emit('go-to-user', item.userId)">
      <div class="dynamic-user-info">
        <div class="dynamic-username" @click="emit('go-to-user', item.userId)">
          {{ item.user?.username || '用户' }}
        </div>
        <div class="dynamic-time">{{ formatTime(item.createdAt) }}</div>
      </div>
      <button class="more-btn">
        <el-icon><MoreFilled /></el-icon>
      </button>
    </div>

    <div class="dynamic-body">
      <div class="dynamic-text">{{ item.content }}</div>

      <div v-if="item.imageUrls && item.imageUrls.length > 0" class="dynamic-images">
        <img loading="lazy" decoding="async" v-for="(url, index) in item.imageUrls" :key="index" :src="url" alt="" class="dynamic-image" @click="() => {}">
      </div>

      <div v-if="item.refManuscriptId" class="video-card" @click="emit('go-to-manuscript', item.refManuscriptId)">
        <img loading="lazy" decoding="async" v-if="item.refVideo?.cover" :src="item.refVideo.cover" alt="稿件封面" class="video-cover">
        <div v-else class="video-cover-placeholder">
          <el-icon><VideoPlay /></el-icon>
        </div>
        <div class="video-info">
          <div class="video-title">{{ item.refVideo?.title || '引用稿件 #' + item.refManuscriptId }}</div>
          <div class="video-meta" v-if="item.refVideo">
            <span v-if="item.refVideo.duration" class="video-duration">
              <el-icon><Clock /></el-icon>
              {{ item.refVideo.duration }}
            </span>
            <span v-if="item.refVideo.viewCount" class="video-views">
              <el-icon><View /></el-icon>
              {{ formatNumber(item.refVideo.viewCount) }}次观看
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="dynamic-footer">
      <button class="action-btn" @click="emit('forward', item)">
        <el-icon><Share /></el-icon>
        <span>{{ item.shareCount || '转发' }}</span>
      </button>
      <button class="action-btn" :class="{ active: item.showComments }" @click="emit('toggle-comment', item)">
        <el-icon><ChatDotRound /></el-icon>
        <span>{{ item.commentCount || '评论' }}</span>
      </button>
      <button class="action-btn" :class="{ liked: item.stats?.isLiked }" @click="emit('like', item)">
        <el-icon><Star /></el-icon>
        <span>{{ item.stats?.likeCount > 0 ? item.stats.likeCount : '点赞' }}</span>
      </button>
    </div>

    <!-- 评论区 - 展开栏形式 -->
    <el-collapse-transition>
      <div v-show="item.showComments" class="comment-section">
        <div class="comment-section-content">
          <CommentSystem
            :target-type="'DYNAMIC'"
            :target-id="item.id"
            :placeholder="'发一条友善的评论吧~'"
            :total-count="item.commentCount"
            @update:totalCount="val => item.commentCount = val"
          />
        </div>
      </div>
    </el-collapse-transition>
  </div>
</template>

<style scoped>
.dynamic-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.dynamic-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.dynamic-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  cursor: pointer;
}

.dynamic-user-info {
  flex: 1;
}

.dynamic-username {
  font-size: 14px;
  font-weight: 500;
  color: #fb7299;
  cursor: pointer;
}

.dynamic-time {
  font-size: 12px;
  color: #9499a0;
  margin-top: 4px;
}

.more-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9499a0;
}

.dynamic-body {
  margin-bottom: 16px;
}

.dynamic-text {
  font-size: 14px;
  color: #18191c;
  line-height: 1.6;
  white-space: pre-line;
}

.dynamic-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.dynamic-image {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 6px;
  cursor: pointer;
}

.video-card {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  padding: 12px;
  background: #f6f7f8;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.3s;
}

.video-card:hover {
  background: #e8e9ea;
}

.video-cover {
  width: 120px;
  height: 75px;
  object-fit: cover;
  border-radius: 6px;
  flex-shrink: 0;
}

.video-cover-placeholder {
  width: 120px;
  height: 75px;
  background: #e3e5e7;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9499a0;
  font-size: 24px;
  flex-shrink: 0;
}

.video-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.video-title {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-all;
}

.video-meta {
  display: flex;
  gap: 16px;
  margin-top: 8px;
  font-size: 12px;
  color: #9499a0;
}

.video-duration,
.video-views {
  display: flex;
  align-items: center;
  gap: 4px;
}

.dynamic-footer {
  display: flex;
  border-top: 1px solid #e3e5e7;
  padding-top: 12px;
}

.dynamic-footer .action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px;
  border: none;
  background: transparent;
  color: #61666d;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s;
  border-radius: 6px;
}

.dynamic-footer .action-btn:hover {
  background: #f1f2f3;
  color: #00aeec;
}

.dynamic-footer .action-btn.liked {
  color: #00a1d6;
}

.dynamic-footer .action-btn.active {
  color: #00aeec;
}

/* 评论区 - 展开栏形式 */
.comment-section {
  margin-top: 12px;
  border-top: 1px solid #e3e5e7;
  background: #f6f7f8;
  border-radius: 0 0 8px 8px;
  overflow: hidden;
}

.comment-section-content {
  padding: 16px;
  background: #fff;
}
</style>