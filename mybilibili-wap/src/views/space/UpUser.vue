<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getUserInfo, getUserVideos } from '../../api/up-user'
import { formatTenThousand } from '../../utils/format'

const route = useRoute()
const mId = parseInt(route.params.mId) || 0

const upUser = ref<any>(null)
const videos = ref([])
const loading = ref(true)
const showMore = ref(true)
const page = ref(1)

onMounted(async () => {
  try {
    const res = await getUserInfo(mId)
    if (res.code === '1') {
      upUser.value = res.data
    }
    await loadVideos()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})

const loadVideos = async () => {
  try {
    const res = await getUserVideos(mId, page.value, 20)
    if (res.code === '1') {
      if (page.value === 1) videos.value = res.data || []
      else videos.value.push(...(res.data || []))
      showMore.value = (res.data || []).length >= 20
    }
  } catch (e) {
    console.error(e)
  }
}

const loadMore = () => {
  page.value++
  loadVideos()
}
</script>

<template>
  <div class="up-user-page">
    <div v-if="upUser" class="user-header">
      <img :src="upUser.face || '../../../assets/noface.gif'" class="user-avatar" />
      <div class="user-info">
        <div class="user-name-row">
          <span class="user-name">{{ upUser.name }}</span>
          <img v-if="upUser.level" :src="`../../../assets/lv${upUser.level}.png`" class="user-level" />
        </div>
        <p class="user-id">UID: {{ upUser.mid }}</p>
        <div class="user-stats">
          <span>{{ formatTenThousand(upUser.following || 0) }} 关注</span>
          <span>{{ formatTenThousand(upUser.follower || 0) }} 粉丝</span>
        </div>
        <p v-if="upUser.sign" class="user-sign">{{ upUser.sign }}</p>
      </div>
    </div>

    <div class="videos-section">
      <div class="section-title">投稿视频</div>
      <div v-if="loading && page === 1" class="loading">加载中...</div>
      <div v-else-if="videos.length" class="video-grid">
        <VideoItem v-for="v in videos" :key="v.aId" :video="v" />
      </div>
      <div v-else class="empty">Ta还没有投过稿~</div>
      <div v-if="showMore && !loading" class="load-more" @click="loadMore">加载更多</div>
      <div v-if="!showMore && videos.length" class="no-more">— 加载完毕 —</div>
    </div>

    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.up-user-page { padding-bottom: 20px; }

.user-header {
  background: $bg-white;
  padding: 16px;
  display: flex;
  gap: 14px;
  margin-bottom: 8px;
}

.user-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #e3e5e7;
}

.user-info { flex: 1; }

.user-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.user-name { font-size: 16px; font-weight: 600; }

.user-level { width: 20px; height: auto; }

.user-id { font-size: 12px; color: $text-secondary; margin-bottom: 4px; }

.user-stats {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: $text-secondary;
  margin-bottom: 8px;
}

.user-sign { font-size: 12px; color: $text-third; }

.videos-section { padding: 12px; }

.section-title { font-size: 15px; font-weight: 600; margin-bottom: 10px; }

.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.loading, .empty {
  text-align: center;
  padding: 30px;
  color: $text-secondary;
}

.load-more {
  text-align: center;
  padding: 12px;
  color: $theme-pink;
  cursor: pointer;
}

.no-more {
  text-align: center;
  padding: 12px;
  color: $text-third;
  font-size: 12px;
}
</style>