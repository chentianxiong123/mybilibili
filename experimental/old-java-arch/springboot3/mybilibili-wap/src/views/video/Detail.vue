<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Header from '../../components/Header.vue'
import VideoPlayer from './VideoPlayer.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getVideoInfo, getRecommendVides, getComments } from '../../api/video'
import { followUser, checkFollow, likeManuscript, coinManuscript, collectManuscript, shareManuscript, getInteractionStatus } from '../../api/interaction'
import { postComment, replyComment, likeComment } from '../../api/comment'
import storage from '../../utils/storage'

const route = useRoute()
const router = useRouter()
const aId = parseInt(route.params.aId) || 0

const video = ref(null)
const recommendVides = ref([])
const comments = ref([])
const loading = ref(true)
const commentPage = ref(1)
const commentCount = ref(0)
const hasMoreComments = ref(true)

// 当前激活的二级大标签（'desc'-简介，'comment'-评论）
const activeSection = ref('desc')

// 赞/踩/币/藏/转等互动逻辑状态
const isLiked = ref(false)
const isDisliked = ref(false)
const isCoined = ref(false)
const isStarred = ref(false)
const isFollowing = ref(false)

const likeCount = ref(0)
const coinCount = ref(0)
const starCount = ref(0)
const shareCount = ref(0)
const followerCount = ref(0)

// 单选多分P分集列表
const activePartIndex = ref(0)
const videoParts = ref([])

// 评论输入
const newComment = ref('')
const commentInput = ref('')

// 视频播放器引用
const videoPlayerRef = ref(null)

onMounted(async () => {
  await loadData()
  await loadInteractionState()
  await loadFollowState()
})

watch(() => route.params.aId, async () => {
  await loadData()
  await loadInteractionState()
  await loadFollowState()
})

const loadInteractionState = async () => {
  try {
    const token = localStorage.getItem('token')
    if (!token) return
    const res = await getInteractionStatus(aId)
    if (res.code === '1' && res.data) {
      const d = res.data
      // Jackson序列化boolean isXxx() → xxx，所以后端isLiked变成liked
      isLiked.value = !!d.liked || !!d.isLiked
      isCoined.value = (d.coinCount || 0) > 0 || !!d.coined
      isStarred.value = !!d.collected || !!d.isCollected
    }
  } catch (e) {}
}

const loadFollowState = async () => {
  try {
    const token = localStorage.getItem('token')
    if (!token || !video.value?.mid) return
    const res = await checkFollow(video.value.mid)
    if (res.code === '1') {
      // 后端返回的是 { following: true/false }，兼容直接布尔值
      isFollowing.value = typeof res.data === 'object' ? !!res.data.following : !!res.data
    }
  } catch (e) {}
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getVideoInfo(aId)
    if (res.code === '1' && res.data) {
      const d = res.data
      video.value = d

      likeCount.value = d.likeCount || 0
      coinCount.value = d.coinCount || 0
      starCount.value = d.collectCount || 0
      shareCount.value = d.shareCount || 0
      followerCount.value = d.uploader?.followerCount || 0

      // 多分P
      if (d.videos && d.videos.length > 0) {
        videoParts.value = d.videos.map((v, i) => ({
          id: v.id,
          title: v.title,
          active: i === 0
        }))
      }

      // 记录观看历史
      storage.setViewHistory({
        aId: d.aId,
        title: d.title,
        pic: d.pic,
        viewAt: new Date().getTime()
      })
    } else {
      video.value = null
    }

    const recRes = await getRecommendVides(aId)
    if (recRes.code === '1' && recRes.data?.length > 0) {
      recommendVides.value = recRes.data || []
    }

    await loadComments()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const loadComments = async () => {
  try {
    const res = await getComments(aId, commentPage.value)
    if (res.code === '1' && res.data?.length > 0) {
      if (commentPage.value === 1) {
        comments.value = res.data || []
      } else {
        comments.value.push(...(res.data || []))
      }
      commentCount.value = comments.value.length
      hasMoreComments.value = (res.data || []).length >= 20
    } else if (commentPage.value === 1) {
      comments.value = []
      commentCount.value = 0
      hasMoreComments.value = false
    }
  } catch (e) {
    console.error(e)
  }
}

const loadMoreComments = () => {
  commentPage.value++
  loadComments()
}

// 互动点击操作
const handleLike = async () => {
  const token = localStorage.getItem('token')
  if (!token) { router.push('/m/login'); return }
  try {
    const res = await likeManuscript(aId, !isLiked.value)
    if (res.code === '1') {
      isLiked.value = !isLiked.value
      likeCount.value += isLiked.value ? 1 : -1
      if (isDisliked.value && isLiked.value) isDisliked.value = false
    }
  } catch (e) {}
}

const handleDislike = () => {
  isDisliked.value = !isDisliked.value
  if (isDisliked.value && isLiked.value) {
    isLiked.value = false
    likeCount.value = Math.max(0, likeCount.value - 1)
  }
}

const handleCoin = async () => {
  const token = localStorage.getItem('token')
  if (!token) { router.push('/m/login'); return }
  if (isCoined.value) return
  try {
    const res = await coinManuscript(aId, 1)
    if (res.code === '1') {
      isCoined.value = true
      coinCount.value++
    }
  } catch (e) {}
}

const handleStar = async () => {
  const token = localStorage.getItem('token')
  if (!token) { router.push('/m/login'); return }
  try {
    const res = await collectManuscript(aId, !isStarred.value)
    if (res.code === '1') {
      isStarred.value = !isStarred.value
      starCount.value += isStarred.value ? 1 : -1
    }
  } catch (e) {}
}

const handleShare = async () => {
  const shareUrl = `${window.location.origin}/m/video/${aId}`
  navigator.clipboard.writeText(shareUrl)
  try {
    await shareManuscript(aId)
    shareCount.value++
  } catch (e) {}
}

const handleFollow = async () => {
  const token = localStorage.getItem('token')
  if (!token) { router.push('/m/login'); return }
  try {
    const res = await followUser(video.value.mid, !isFollowing.value)
    if (res.code === '1') {
      isFollowing.value = !isFollowing.value
      followerCount.value += isFollowing.value ? 1 : -1
    }
  } catch (e) {}
}

// 提交评论
const submitComment = async () => {
  const token = localStorage.getItem('token')
  if (!token) { router.push('/m/login'); return }
  if (!commentInput.value.trim()) return
  try {
    const res = await postComment(aId, commentInput.value)
    if (res.code === '1') {
      commentInput.value = ''
      // 重新加载评论列表
      commentPage.value = 1
      await loadComments()
    }
  } catch (e) {}
}

const selectPart = (index) => {
  activePartIndex.value = index
}

const formatTimeLabel = (timeStr) => {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  if (isNaN(d.getTime())) return timeStr
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const r = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  return `${y}年${m}月${r}日 ${h}:${min}`
}

const formatCount = (num) => {
  const n = parseInt(num)
  if (isNaN(n)) return 0
  if (n >= 10000) {
    return (Math.floor(n / 1000) / 10) + '万'
  }
  return n
}

const goBack = () => {
  router.back()
}
</script>

<template>
  <div class="video-detail-page">
    <!-- 顶部极简状态返回头 -->
    <div class="top-nav-bar">
      <div class="back-icon" @click="goBack">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="#fff" stroke-width="2">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </div>
      <div class="home-icon" @click="router.push('/m/index')">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="#fff" stroke-width="2">
          <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
          <polyline points="9 22 9 12 15 12 15 22" />
        </svg>
      </div>
      <div class="right-more">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="#fff" stroke-width="2">
          <circle cx="12" cy="5" r="1" />
          <circle cx="12" cy="12" r="1" />
          <circle cx="12" cy="19" r="1" />
        </svg>
      </div>
    </div>

    <!-- 视频播放器区 -->
    <VideoPlayer ref="videoPlayerRef" :video="video" danmaku-mount="#danmaku-emitter-mount" />

    <!-- 简介/评论/弹幕控制栏 -->
    <div class="tab-section-header">
      <div class="tabs-titles">
        <span
          :class="['tab-title', { active: activeSection === 'desc' }]"
          @click="activeSection = 'desc'"
        >
          简介
        </span>
        <span
          :class="['tab-title', { active: activeSection === 'comment' }]"
          @click="activeSection = 'comment'"
        >
          评论 <span class="comment-count-label">{{ commentCount }}</span>
        </span>
      </div>
      <!-- artplayer 弹幕插件控制区（由插件自动挂载） -->
      <div id="danmaku-emitter-mount" class="danmaku-mount-area"></div>
    </div>

    <!-- 1. 简介模式内容区块 -->
    <div v-if="activeSection === 'desc' && video" class="desc-content-wrapper">
      <!-- UP主信息卡片 -->
      <div class="up-info-row">
        <div class="up-left">
          <img :src="video.uploader?.face || ''" class="up-avatar" />
          <div class="up-meta">
            <div class="up-name">{{ video.author }}</div>
            <div class="up-fans">{{ followerCount }}粉丝</div>
          </div>
        </div>
        <div :class="['follow-up-btn', { following: isFollowing }]" @click="handleFollow">
          {{ isFollowing ? '已关注' : '+ 关注' }}
        </div>
      </div>

      <!-- 视频标题、播放量与时间 -->
      <div class="video-main-details">
        <h1 class="video-main-title">{{ video.title }}</h1>
        <div class="video-plays-meta">
          <span class="plays-icon">▶</span>
          <span class="num">{{ formatCount(video.play) }}</span>
          <span class="plays-icon m-left">💬</span>
          <span class="num">{{ video.videoReview || 0 }}</span>
          <span class="date">{{ formatTimeLabel(video.ctime) }}</span>
        </div>
        <div class="video-full-desc" v-if="video.description">
          {{ video.description }}
        </div>
      </div>

      <!-- 五大互动操作按钮 -->
      <div class="interactions-actions-row">
        <div :class="['action-btn', { active: isLiked }]" @click="handleLike">
          <div class="action-icon">👍</div>
          <div class="action-num">{{ likeCount }}</div>
        </div>
        <div :class="['action-btn', { active: isDisliked }]" @click="handleDislike">
          <div class="action-icon">👎</div>
          <div class="action-num">不喜欢</div>
        </div>
        <div :class="['action-btn', { active: isCoined }]" @click="handleCoin">
          <div class="action-icon">🪙</div>
          <div class="action-num">{{ coinCount }}</div>
        </div>
        <div :class="['action-btn', { active: isStarred }]" @click="handleStar">
          <div class="action-icon">⭐</div>
          <div class="action-num">{{ starCount }}</div>
        </div>
        <div :class="['action-btn', { active: false }]" @click="handleShare">
          <div class="action-icon">↪️</div>
          <div class="action-num">{{ shareCount }}</div>
        </div>
      </div>

      <!-- 分P选集列表 -->
      <div v-if="videoParts.length > 0" class="parts-slider-container">
        <div class="parts-list">
          <div
            v-for="(part, index) in videoParts"
            :key="part.id"
            :class="['part-card', { active: activePartIndex === index }]"
            @click="selectPart(index)"
          >
            <div class="part-title">{{ part.title }}</div>
          </div>
        </div>
      </div>

      <!-- 视频标签 -->
      <div v-if="video.tags && video.tags.length > 0" class="category-tag-row">
        <span v-for="tag in video.tags" :key="tag" class="tag-item">{{ tag }}</span>
      </div>

      <!-- 相关推荐视频列表 -->
      <div class="recommendations-list">
        <div
          v-for="rec in recommendVides"
          :key="rec.aId"
          class="recommend-card-row"
          @click="router.push(`/m/video/${rec.aId}`)"
        >
          <div class="rec-cover-wrapper">
            <img :src="rec.pic" class="rec-cover" />
            <span class="rec-duration">{{ rec.duration || '' }}</span>
          </div>
          <div class="rec-info">
            <div class="rec-title">{{ rec.title }}</div>
            <div class="rec-tags-row">
              <span class="rec-author">{{ rec.author }}</span>
            </div>
            <div class="rec-stats">
              <span>▶ {{ formatCount(rec.play) }}</span>
              <span class="m-left">💬 {{ rec.videoReview }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 2. 评论模式内容区块 -->
    <div v-else-if="activeSection === 'comment'" class="comments-content-wrapper">
      <div class="comments-sorting-header">
        <span class="comment-type-title">评论</span>
      </div>

      <div class="comment-list-block">
        <div v-for="c in comments" :key="c.rpid" class="comment-card-item">
          <div class="comment-card-top">
            <img :src="c.user?.face || ''" class="commenter-avatar" />
            <div class="commenter-main-wrap">
              <div class="commenter-meta">
                <span class="commenter-name">{{ c.user?.name }}</span>
                <span class="commenter-level" v-if="c.user?.level">LV{{ c.user.level }}</span>
              </div>

              <div class="comment-text-box">
                <p class="comment-text-content">
                  <span v-if="c.isTop" class="top-tag-badge">置顶</span>
                  {{ c.content }}
                </p>
              </div>

              <div class="comment-footer-actions">
                <span class="time-label">{{ c.ctime }}</span>
              </div>

              <!-- 子回复 -->
              <div v-if="c.replies && c.replies.length > 0" class="sub-replies-box">
                <div v-for="(sub, sIdx) in c.replies" :key="sIdx" class="sub-reply-row">
                  <span class="sub-name">{{ sub.name }}</span>
                  <span v-if="sub.isUp" class="sub-up-badge">UP</span>：
                  <span class="sub-text-content">{{ sub.content }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="comments.length === 0 && !loading" class="no-more-label">暂无评论</div>
      <div v-if="hasMoreComments" class="click-load-more" @click="loadMoreComments">
        点击加载更多评论
      </div>
      <div v-else-if="comments.length > 0" class="no-more-label">— 已经到底啦 —</div>

      <!-- 底部发表评论 -->
      <div class="bottom-comment-bar-wrap">
        <input
          v-model="commentInput"
          class="comment-input-area-placeholder"
          placeholder="发一条友善的评论吧..."
          @keyup.enter="submitComment"
        />
        <div class="comment-emoji-btn" @click="submitComment">发送</div>
      </div>
    </div>

    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.video-detail-page {
  min-height: 100vh;
  background: #f4f5f7;
  position: relative;
  padding-bottom: 60px;
}

/* 顶部导航 */
.top-nav-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 48px;
  display: flex;
  align-items: center;
  padding: 0 14px;
  background: linear-gradient(rgba(0,0,0,0.5), transparent);
  z-index: 999;
  gap: 16px;

  .back-icon, .home-icon {
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .right-more {
    margin-left: auto;
    cursor: pointer;
    display: flex;
    align-items: center;
  }
}

/* Tabs 头 */
.tab-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  height: 40px;
  padding: 0 16px;
  border-bottom: 1px solid #f1f2f3;

  .tabs-titles {
    display: flex;
    gap: 24px;

    .tab-title {
      font-size: 14px;
      color: #61666d;
      position: relative;
      cursor: pointer;
      height: 38px;
      line-height: 38px;

      .comment-count-label {
        font-size: 11px;
        color: #9499a0;
        margin-left: 2px;
      }

      &.active {
        color: #fb7299;
        font-weight: bold;
        &::after {
          content: '';
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          height: 2px;
          background: #fb7299;
          border-radius: 1px;
        }
      }
    }
  }

  .danmaku-mount-area {
    display: flex;
    align-items: center;
    gap: 8px;

    // artplayer-plugin-danmuku 挂载后的样式覆盖
    :deep(.apd-emitter) {
      display: flex;
      align-items: center;
      gap: 6px;
    }

    :deep(.apd-input) {
      width: 80px;
      height: 28px;
      border: none;
      outline: none;
      background: #f1f2f3;
      border-radius: 14px;
      padding: 0 10px;
      font-size: 12px;
      color: #18191c;
      &::placeholder { color: #9499a0; }
    }

    :deep(.apd-send) {
      font-size: 12px;
      color: #fb7299;
      cursor: pointer;
      font-weight: 500;
      padding: 2px 8px;
      white-space: nowrap;
    }

    :deep(.apd-style) {
      display: flex;
      align-items: center;
      cursor: pointer;
    }
  }
}

/* 简介模块 */
.desc-content-wrapper {
  background: #fff;
  padding-bottom: 20px;

  .up-info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;

    .up-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .up-avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        object-fit: cover;
        background: #f1f2f3;
      }

      .up-name {
        font-size: 13px;
        font-weight: bold;
        color: #18191c;
      }

      .up-fans {
        font-size: 11px;
        color: #9499a0;
        margin-top: 2px;
      }
    }

    .follow-up-btn {
      font-size: 12px;
      background: #fb7299;
      color: #fff;
      padding: 5px 16px;
      border-radius: 14px;
      font-weight: bold;
      cursor: pointer;

      &.following {
        background: #f1f2f3;
        color: #61666d;
      }
    }
  }

  .video-main-details {
    padding: 0 16px;

    .video-main-title {
      font-size: 15px;
      font-weight: bold;
      color: #18191c;
      line-height: 1.4;
      margin: 0 0 8px 0;
    }

    .video-plays-meta {
      font-size: 11px;
      color: #9499a0;
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 4px;
      margin-bottom: 12px;

      .plays-icon {
        font-size: 10px;
        &.m-left { margin-left: 8px; }
      }
    }

    .video-full-desc {
      font-size: 12px;
      color: #61666d;
      line-height: 1.5;
    }
  }

  .interactions-actions-row {
    display: flex;
    justify-content: space-around;
    padding: 16px 10px;
    border-bottom: 1px solid #f1f2f3;

    .action-btn {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 6px;
      cursor: pointer;

      .action-icon { font-size: 20px; color: #61666d; }
      .action-num { font-size: 11px; color: #9499a0; }

      &.active {
        .action-icon, .action-num {
          color: #fb7299;
          font-weight: bold;
        }
      }
    }
  }

  .parts-slider-container {
    padding: 14px 16px;
    border-bottom: 1px solid #f1f2f3;

    .parts-list {
      display: flex;
      gap: 10px;
      overflow-x: auto;
      scrollbar-width: none;
      &::-webkit-scrollbar { display: none; }

      .part-card {
        flex-shrink: 0;
        width: 140px;
        background: #f1f2f3;
        padding: 12px;
        border-radius: 6px;
        cursor: pointer;

        .part-title {
          font-size: 12px;
          color: #18191c;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
          line-height: 1.4;
        }

        &.active {
          background: #fff1f3;
          border: 1px solid #fb7299;
          .part-title { color: #fb7299; font-weight: bold; }
        }
      }
    }
  }

  .category-tag-row {
    padding: 12px 16px;
    display: flex;
    gap: 8px;
    flex-wrap: wrap;

    .tag-item {
      display: inline-block;
      font-size: 11px;
      color: #61666d;
      background: #f1f2f3;
      padding: 4px 10px;
      border-radius: 12px;
    }
  }

  .recommendations-list {
    padding: 8px 16px;

    .recommend-card-row {
      display: flex;
      gap: 12px;
      margin-bottom: 16px;
      cursor: pointer;

      .rec-cover-wrapper {
        position: relative;
        width: 120px;
        height: 75px;
        border-radius: 6px;
        overflow: hidden;
        background: #e3e5e7;
        flex-shrink: 0;

        .rec-cover { width: 100%; height: 100%; object-fit: cover; }

        .rec-duration {
          position: absolute;
          bottom: 4px;
          right: 4px;
          background: rgba(0,0,0,0.6);
          color: #fff;
          font-size: 9px;
          padding: 1px 4px;
          border-radius: 2px;
        }
      }

      .rec-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        justify-content: space-between;

        .rec-title {
          font-size: 13px;
          font-weight: 500;
          color: #18191c;
          line-height: 1.4;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }

        .rec-tags-row { margin-top: 4px; }
        .rec-author { font-size: 11px; color: #9499a0; }

        .rec-stats {
          font-size: 11px;
          color: #9499a0;
          margin-top: 4px;
          .m-left { margin-left: 8px; }
        }
      }
    }
  }
}

/* 评论模块 */
.comments-content-wrapper {
  background: #fff;
  padding: 14px 16px 80px;

  .comments-sorting-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 14px;

    .comment-type-title {
      font-size: 14px;
      font-weight: bold;
      color: #18191c;
    }
  }

  .comment-card-item {
    margin-bottom: 18px;
    padding-bottom: 12px;
    border-bottom: 1px solid #f1f2f3;

    &:last-child { border-bottom: none; }

    .comment-card-top {
      display: flex;
      gap: 12px;

      .commenter-avatar {
        width: 34px;
        height: 34px;
        border-radius: 50%;
        object-fit: cover;
        flex-shrink: 0;
        background: #f1f2f3;
      }

      .commenter-main-wrap {
        flex: 1;

        .commenter-meta {
          display: flex;
          align-items: center;
          gap: 6px;

          .commenter-name { font-size: 12px; font-weight: bold; color: #61666d; }

          .commenter-level {
            font-size: 8px;
            background: #ff6699;
            color: #fff;
            padding: 0 3px;
            border-radius: 2px;
            font-weight: bold;
          }
        }

        .comment-text-box {
          margin-top: 6px;
          .comment-text-content {
            font-size: 13px;
            color: #18191c;
            line-height: 1.5;
            margin: 0;
          }
        }

        .comment-footer-actions {
          margin-top: 8px;
          display: flex;
          align-items: center;
          font-size: 11px;
          color: #9499a0;
        }

        .sub-replies-box {
          margin-top: 10px;
          background: #f6f7f8;
          padding: 8px 12px;
          border-radius: 6px;

          .sub-reply-row {
            font-size: 12px;
            line-height: 1.6;
            margin-bottom: 6px;
            &:last-child { margin-bottom: 0; }

            .sub-name { color: #507ea6; font-weight: 500; }
            .sub-up-badge {
              font-size: 8px;
              background: #fb7299;
              color: #fff;
              padding: 0 2px;
              border-radius: 2px;
              font-weight: bold;
            }
            .sub-text-content { color: #18191c; }
          }
        }
      }
    }
  }

  .click-load-more {
    text-align: center;
    padding: 16px;
    font-size: 13px;
    color: #fb7299;
    cursor: pointer;
  }

  .no-more-label {
    text-align: center;
    padding: 24px;
    font-size: 11px;
    color: #9499a0;
  }
}

.bottom-comment-bar-wrap {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50px;
  background: #fff;
  border-top: 1px solid #e3e5e7;
  display: flex;
  align-items: center;
  padding: 0 16px;
  z-index: 999;
  gap: 12px;

  .comment-input-area-placeholder {
    flex: 1;
    height: 34px;
    line-height: 34px;
    background: #f1f2f3;
    border-radius: 17px;
    padding: 0 14px;
    font-size: 12px;
    color: #61666d;
    border: none;
    outline: none;
  }

  .comment-emoji-btn {
    font-size: 13px;
    color: #fb7299;
    cursor: pointer;
    font-weight: bold;
    white-space: nowrap;
  }
}

.loading {
  text-align: center;
  padding: 80px 0;
  color: #9499a0;
}
</style>