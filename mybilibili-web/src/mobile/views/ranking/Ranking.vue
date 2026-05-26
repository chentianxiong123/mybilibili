<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Header from '../../components/Header.vue'
import TabBar from '../../components/TabBar.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getRankings, getRankingArchive } from '../../api/ranking'
import { formatTenThousand } from '../../utils/format'

const route = useRoute()
const rId = parseInt(route.params.rId) || 0

const partitions = ref([])
const activeId = ref(rId)
const videos = ref([])
const loading = ref(true)
const page = ref(1)
const hasMore = ref(true)

onMounted(async () => {
  await loadPartitions()
  await loadVideos()
})

const loadPartitions = async () => {
  try {
    const res = await getRankings(0)
    if (res.code === '1') {
      partitions.value = res.data || []
      partitions.value.unshift({ id: 0, name: '全站' })
    }
  } catch (e) {
    console.error(e)
  }
}

const loadVideos = async () => {
  loading.value = true
  try {
    const res = await getRankingArchive({ rId: activeId.value, p: page.value, size: 20 })
    if (res.code === '1') {
      if (page.value === 1) {
        videos.value = res.data || []
      } else {
        videos.value.push(...(res.data || []))
      }
      hasMore.value = (res.data || []).length >= 20
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleTabClick = (tab) => {
  activeId.value = tab.id
  page.value = 1
  videos.value = []
  loadVideos()
}

const handleScroll = (e) => {
  const { scrollTop, scrollHeight, clientHeight } = e.target
  if (scrollTop + clientHeight >= scrollHeight - 100 && hasMore.value && !loading.value) {
    page.value++
    loadVideos()
  }
}
</script>

<template>
  <div class="ranking-page">
    <div class="ranking-header">
      <router-link to="/m/index" class="back-btn">‹</router-link>
      <span class="title">排行榜</span>
    </div>

    <div class="tab-bar-wrapper">
      <TabBar :data="partitions" :active-id="activeId" @click="handleTabClick" />
    </div>

    <div class="video-list" @scroll="handleScroll">
      <div v-if="loading && page === 1" class="loading">加载中...</div>
      <div v-else>
        <div
          v-for="(video, index) in videos"
          :key="video.aId"
          class="rank-item"
          @click="$router.push(`/m/video/${video.aId}`)"
        >
          <div class="rank-badge">
            <img
              v-if="index < 3"
              :src="`/images/rank${index + 1}.png`"
              class="rank-img"
            />
            <span v-else class="rank-num">{{ index + 1 }}</span>
          </div>
          <div class="rank-video">
            <img :src="video.pic" :alt="video.title" loading="lazy" class="rank-cover" />
            <div class="rank-info">
              <p class="rank-title">{{ video.title }}</p>
              <p class="rank-up">{{ video.author }}</p>
              <p class="rank-stats">
                {{ formatTenThousand(video.play) }}播放 · {{ formatTenThousand(video.videoReview) }}弹幕
              </p>
            </div>
          </div>
        </div>
        <div v-if="loading && page > 1" class="loading-more">加载中...</div>
        <div v-if="!hasMore && videos.length" class="no-more">— 加载完毕 —</div>
      </div>
    </div>

    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.ranking-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: $bg-color;
}

.ranking-header {
  display: flex;
  align-items: center;
  height: $header-height;
  padding: 0 12px;
  background: $theme-pink;
  color: #fff;

  .back-btn {
    font-size: 28px;
    width: 40px;
    text-align: center;
  }

  .title { font-size: 16px; font-weight: 600; }
}

.tab-bar-wrapper {
  background: $bg-white;
  border-bottom: 1px solid $border-color;
}

.video-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.rank-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  cursor: pointer;
}

.rank-badge {
  width: 36px;
  text-align: center;
  flex-shrink: 0;

  .rank-img { width: 24px; height: auto; }
  .rank-num { font-size: 14px; color: $text-secondary; font-weight: 600; }
}

.rank-video {
  flex: 1;
  display: flex;
  gap: 10px;
}

.rank-cover {
  width: 100px;
  height: 56px;
  object-fit: cover;
  border-radius: 4px;
  background: #e3e5e7;
}

.rank-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.rank-title {
  font-size: 13px;
  color: $text-primary;
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rank-up {
  font-size: 12px;
  color: $text-secondary;
}

.rank-stats {
  font-size: 11px;
  color: $text-third;
}

.loading, .loading-more {
  text-align: center;
  color: $text-secondary;
  padding: 20px;
}

.no-more {
  text-align: center;
  color: $text-third;
  padding: 16px;
  font-size: 12px;
}
</style>