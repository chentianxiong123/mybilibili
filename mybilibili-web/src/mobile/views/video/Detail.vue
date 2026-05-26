<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Header from '../../components/Header.vue'
import VideoPlayer from './VideoPlayer.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getVideoInfo, getRecommendVides, getComments } from '../../api/video'

const route = useRoute()
const aId = parseInt(route.params.aId) || 0

const video = ref<any>(null)
const recommendVides = ref([])
const comments = ref([])
const loading = ref(true)
const commentPage = ref(1)
const hasMoreComments = ref(true)
const showDesc = ref(false)

onMounted(async () => {
  try {
    const res = await getVideoInfo(aId)
    if (res.code === '1') {
      video.value = res.data
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
</script>

<template>
  <div class="video-detail-page">
    <Header />
    <VideoPlayer :video="video" />

    <div class="video-content" v-if="video">
      <div class="video-info">
        <h1 class="title">{{ video.title }}</h1>
        <div class="meta">
          <span class="play">{{ video.play }}播放</span>
          <span class="danmaku">{{ video.videoReview }}弹幕</span>
        </div>
        <div class="up-info">
          <span class="up-name" @click="$router.push(`/m/space/${video.mid}`)">
            {{ video.author }}
          </span>
        </div>
        <div class="desc-toggle" @click="showDesc = !showDesc">
          <span>简介</span>
          <span>{{ showDesc ? '▲' : '▼' }}</span>
        </div>
        <div v-if="showDesc" class="desc">
          {{ video.description || '暂无简介' }}
        </div>
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
        <div class="section-title">评论</div>
        <div class="comment-list">
          <div v-for="c in comments" :key="c.rpid" class="comment-item">
            <img :src="c.user?.face || '/images/noface.gif'" class="comment-avatar" />
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
          加载更多
        </div>
        <div v-else-if="comments.length" class="no-more">— 没有更多了 —</div>
      </div>
    </div>

    <div v-else-if="loading" class="loading">加载中...</div>
    <div v-else class="empty">视频不存在</div>
    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.video-detail-page { padding-bottom: 20px; }

.video-info {
  padding: 12px;
  background: $bg-white;
  margin-bottom: 8px;
}

.title {
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
  margin-bottom: 8px;
}

.meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: $text-secondary;
  margin-bottom: 8px;
}

.up-info {
  margin-bottom: 8px;
}

.up-name {
  font-size: 14px;
  color: #00a1d6;
  cursor: pointer;
}

.desc-toggle {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-top: 1px solid $border-color;
  font-size: 13px;
  color: $text-secondary;
  cursor: pointer;
}

.desc {
  font-size: 13px;
  color: $text-secondary;
  padding: 8px 0;
  line-height: 1.5;
}

.recommend-section {
  padding: 12px;
  background: $bg-white;
  margin-bottom: 8px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 10px;
}

.recommend-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.comment-section {
  padding: 12px;
  background: $bg-white;
}

.comment-list { margin-bottom: 12px; }

.comment-item {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.comment-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #e3e5e7;
}

.comment-body { flex: 1; }

.comment-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.comment-name { font-size: 13px; color: #00a1d6; }
.comment-date { font-size: 11px; color: $text-third; }

.comment-content {
  font-size: 14px;
  line-height: 1.5;
  color: $text-primary;
}

.load-more {
  text-align: center;
  padding: 12px;
  color: $theme-pink;
  font-size: 14px;
  cursor: pointer;
}

.no-more {
  text-align: center;
  padding: 12px;
  color: $text-third;
  font-size: 12px;
}

.loading, .empty {
  text-align: center;
  padding: 40px;
  color: $text-secondary;
}
</style>