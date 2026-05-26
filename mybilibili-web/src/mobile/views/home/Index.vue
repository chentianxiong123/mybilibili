<script setup>
import { ref, computed, onMounted } from 'vue'
import Header from '../../components/Header.vue'
import TabBar from '../../components/TabBar.vue'
import Drawer from '../../components/Drawer.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getHomeContent, getBanners } from '../../api/index'

const banners = ref([])
const partitions = ref([])
const rankingVideos = ref([])
const additionalVideos = ref([])
const activeTabId = ref(0)
const drawerRef = ref(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const [contentRes, bannerRes] = await Promise.all([
      getHomeContent(),
      getBanners()
    ])
    if (contentRes.code === '1') {
      partitions.value = contentRes.data.oneLevelPartitions || []
      additionalVideos.value = contentRes.data.additionalVideos || []
      rankingVideos.value = contentRes.data.rankingVideos || []
    }
    if (bannerRes.code === '1') {
      banners.value = bannerRes.data || []
    }
    initSwiper()
  } catch (e) {
    console.error('首页加载失败:', e)
  } finally {
    loading.value = false
  }
})

const initSwiper = () => {
  let idx = 0
  const items = document.querySelectorAll('.swiper-slide')
  if (!items.length) return
  const showSlide = (i) => {
    items.forEach((el, index) => {
      el.style.display = index === i ? 'block' : 'none'
    })
  }
  showSlide(0)
  setInterval(() => {
    idx = (idx + 1) % items.length
    showSlide(idx)
  }, 3000)
}

const handleTabClick = (tab) => {
  if (tab.id === -1) {
    window.location.href = '/m/live'
    return
  }
  if (tab.id === 0) {
    window.location.href = '/m/index'
  } else {
    window.location.href = `/m/channel/${tab.id}`
  }
}

const tabBarData = computed(() => {
  return [
    { id: 0, name: '首页' },
    ...partitions.value,
    { id: -1, name: '直播' }
  ]
})

const displayVideos = computed(() => {
  let videos = [...rankingVideos.value]
  if (additionalVideos.value.length > 0) {
    const addIds = new Set(additionalVideos.value.map(v => v.aId))
    videos = videos.filter(v => !addIds.has(v.aId))
  }
  return videos
})
</script>

<template>
  <div class="mobile-index">
    <Header />
    <div class="partition-bar">
      <TabBar :data="tabBarData" :active-id="activeTabId" @click="handleTabClick" />
      <div class="switch-btn" @click="drawerRef?.show()">▼</div>
    </div>
    <Drawer :data="tabBarData" ref="drawerRef" @click="handleTabClick" />

    <div class="content" v-if="!loading">
      <div v-if="banners.length" class="banner-slider">
        <div class="swiper-container">
          <div v-for="b in banners" :key="b.id" class="swiper-slide">
            <a :href="b.url">
              <img :src="b.pic" width="100%" height="100%" alt="" />
            </a>
          </div>
        </div>
        <div class="pagination-dots">
          <span v-for="(_, i) in banners" :key="i" :class="['dot', { active: i === 0 }]" />
        </div>
      </div>

      <div class="video-section">
        <div v-if="additionalVideos.length" class="section-title">推荐</div>
        <div class="video-grid">
          <VideoItem
            v-for="v in additionalVideos"
            :key="v.aId"
            :video="v"
            :show-statistics="false"
          />
        </div>
      </div>

      <div class="video-section">
        <div class="section-title">排行榜</div>
        <div class="video-grid">
          <VideoItem
            v-for="v in displayVideos"
            :key="v.aId"
            :video="v"
            :show-statistics="true"
          />
        </div>
      </div>
    </div>

    <div v-else class="loading">加载中...</div>
    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.mobile-index {
  padding-bottom: 20px;
}

.partition-bar {
  display: flex;
  align-items: center;
  background: $bg-white;
  border-bottom: 1px solid $border-color;

  .tab-bar { flex: 1; }

  .switch-btn {
    width: 40px;
    text-align: center;
    color: $text-secondary;
    font-size: 16px;
    cursor: pointer;
    user-select: none;
  }
}

.banner-slider {
  .swiper-slide {
    width: 100%;
    img { width: 100%; display: block; }
    display: none;
    &:first-child { display: block; }
  }
}

.pagination-dots {
  display: flex;
  justify-content: center;
  gap: 4px;
  padding: 8px 0;

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: rgba(0,0,0,0.3);

    &.active {
      background: $theme-pink;
    }
  }
}

.video-section {
  padding: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
  color: $text-primary;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: $text-secondary;
}
</style>