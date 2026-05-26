<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Header from '../../components/Header.vue'
import VideoPlayer from './VideoPlayer.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getVideoInfo, getRecommendVides, getComments } from '../../api/video'
import storage from '../../utils/storage'

const route = useRoute()
const aId = parseInt(route.params.aId) || 0

const video = ref<any>(null)
const recommendVides = ref([])
const comments = ref([])
const loading = ref(true)
const commentPage = ref(1)
const commentCount = ref(0)
const hasMoreComments = ref(true)
const showFullDesc = ref(false)
const infoExpand = ref(false)

onMounted(async () => {
  try {
    const res = await getVideoInfo(aId)
    if (res.code === '1') {
      video.value = res.data
      // 记录观看历史（翻译自原项目 Detail.tsx）
      storage.setViewHistory({
        aId: video.value.aId,
        title: video.value.title,
        pic: video.value.pic,
        viewAt: new Date().getTime()
      })
    }
    const recRes = await getRecommendVides(aId)
    if (recRes.code === '1') {
      recommendVides.value = recRes.data || []
    }
    await loadComments()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})

const loadComments = async () => {
  try {
    const res = await getComments(aId, commentPage.value)
    if (res.code === '1') {
      if (commentPage.value === 1) {
        comments.value = res.data || []
      } else {
        comments.value.push(...(res.data || []))
      }
      commentCount.value = res.data?.length || 0
      hasMoreComments.value = (res.data || []).length >= 20
    }
  } catch (e) {
    console.error(e)
  }
}

const loadMoreComments = () => {
  commentPage.value++
  loadComments()
}

const toggleInfo = () => {
  infoExpand.value = !infoExpand.value
}

// 发布时间格式化（翻译自原项目 Detail.tsx getPubdate）
const formatPubdate = (timestamp) => {
  if (!timestamp) return ''
  const publicDate = new Date(timestamp * 1000)
  const date = new Date()
  if (publicDate.getFullYear() === date.getFullYear()) {
    if (publicDate.getMonth() === date.getMonth()) {
      const diffDate = date.getDate() - publicDate.getDate()
      switch (diffDate) {
        case 0:
          if (date.getHours() - publicDate.getHours() === 0) {
            return date.getMinutes() - publicDate.getMinutes() + '分钟前'
          }
          return date.getHours() - publicDate.getHours() + '小时前'
        case 1: return '昨天'
        case 2: return '前天'
        default: return (publicDate.getMonth() + 1) + '-' + publicDate.getDate()
      }
    }
    return (publicDate.getMonth() + 1) + '-' + publicDate.getDate()
  }
  return publicDate.getFullYear() + '-' + (publicDate.getMonth() + 1) + '-' + publicDate.getDate()
}
</script>

<template>
  <div class="video-detail-page">
    <Header />
    <VideoPlayer :video="video" />

    <div class="video-content" v-if="video">
      <!-- 视频信息（翻译自原项目样式） -->
      <div class="video-info" :class="{ expanded: infoExpand }">
        <div class="info-container" :class="{ show: infoExpand }">
          <div class="info-wrapper">
            <div class="title" :class="{ wrap: infoExpand }">{{ video.title }}</div>
            <div class="meta">
              <router-link :to="`/m/space/${video.mid}`" class="up-name">{{ video.author }}</router-link>
              <span class="play-count"><i class="icon-play-count" /> {{ video.play }}</span>
              <span class="barrage-count"><i class="icon-barrage-count" /> {{ video.videoReview }}</span>
              <span class="pubdate">{{ formatPubdate(video.ctime) }}</span>
            </div>
            <div class="desc" v-if="video.description">{{ video.description }}</div>
          </div>
        </div>
        <i
          :class="['icon-arrow-down', 'toggle-btn', { rotate: infoExpand }]"
          @click="toggleInfo"
        />
      </div>

      <!-- 推荐视频 -->
      <div v-if="recommendVides.length" class="recommend-section">
        <div class="section-title">相关推荐</div>
        <div class="recommend-grid">
          <VideoItem v-for="v in recommendVides" :key="v.aId" :video="v" />
        </div>
      </div>

      <!-- 评论区 -->
      <div class="comment-section">
        <div class="section-title">评论 <span v-if="commentCount">({{ commentCount }})</span></div>
        <div class="comment-list">
          <div v-for="c in comments" :key="c.rpid" class="comment-item">
            <img :src="c.user?.face || '../../../assets/noface.gif'" class="comment-avatar" />
            <div class="comment-body">
              <div class="comment-header">
                <span class="comment-name">{{ c.user?.name }}</span>
                <span class="comment-date">{{ c.ctime }}</span>
              </div>
              <p class="comment-content">{{ c.content }}</p>
            </div>
          </div>
        </div>
        <div v-if="hasMoreComments" class="load-more" @click="loadMoreComments">
          点击加载更多评论
        </div>
        <div v-else-if="comments.length" class="no-more">没有更多了 ~</div>
      </div>
    </div>

    <div v-else-if="loading" class="loading">加载中...</div>
    <div v-else class="empty">视频不存在</div>
    <ScrollToTop />
  </div>
</template>