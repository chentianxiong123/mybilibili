<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import Header from '../../components/Header.vue'
import TabBar from '../../components/TabBar.vue'
import Drawer from '../../components/Drawer.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getHomeContent, getBanners } from '../../api/index'
import { getHotwords } from '../../api/search'

const banners = ref([])
const partitions = ref([])
const rankingVideos = ref([])
const additionalVideos = ref([])
const activeTabId = ref(0)
const drawerRef = ref(null)
const loading = ref(true)
const currentSlide = ref(0)
const searchPlaceholder = ref('搜索...')
let swiperTimer = null

onMounted(async () => {
  try {
    const [contentRes, bannerRes, hotRes] = await Promise.all([
      getHomeContent(),
      getBanners(),
      getHotwords()
    ])
    if (contentRes.code === '1') {
      partitions.value = contentRes.data.oneLevelPartitions || []
      additionalVideos.value = contentRes.data.additionalVideos || []
      rankingVideos.value = contentRes.data.rankingVideos || []
    }
    if (bannerRes.code === '1') {
      banners.value = bannerRes.data || []
    }
    if (hotRes.code === '1' && hotRes.data?.length) {
      const rand = hotRes.data[Math.floor(Math.random() * hotRes.data.length)]
      searchPlaceholder.value = rand.keyword || '搜索...'
    }
    initSwiper()
    await nextTick()
    alignAvatar()
  } catch (e) {
    console.error('首页加载失败:', e)
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (swiperTimer) clearInterval(swiperTimer)
})

const alignAvatar = () => {
  const avatar = document.querySelector('.mobile-header .user-avatar')
  const firstTab = document.querySelector('.tab-bar .tab-item')
  if (!avatar || !firstTab) return
  const range = document.createRange()
  range.selectNodeContents(firstTab)
  const textRect = range.getBoundingClientRect()
  const avatarRect = avatar.getBoundingClientRect()
  const offset = textRect.left - avatarRect.left
  if (offset !== 0) {
    avatar.style.marginLeft = offset + 'px'
  }
}

const initSwiper = () => {
  if (banners.value.length <= 1) return
  swiperTimer = setInterval(() => {
    currentSlide.value = (currentSlide.value + 1) % banners.value.length
  }, 3000)
}

const goToSlide = (i) => {
  currentSlide.value = i
  if (swiperTimer) clearInterval(swiperTimer)
  swiperTimer = setInterval(() => {
    currentSlide.value = (currentSlide.value + 1) % banners.value.length
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
  return [...(additionalVideos.value || []), ...(rankingVideos.value || [])]
    .filter((v, i, arr) => arr.findIndex(x => x.aId === v.aId) === i)
})
</script>

<template>
  <div class="mobile-index">
    <Header :placeholder="searchPlaceholder" />
    <div class="partition-bar">
      <TabBar :data="tabBarData.slice(0, 6)" :active-id="activeTabId" @click="handleTabClick" />
      <div class="switch-btn" @click="drawerRef?.show()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </div>
    </div>
    <Drawer :data="tabBarData" ref="drawerRef" @click="handleTabClick" />

    <div class="content" v-if="!loading">
      <div v-if="banners.length" class="banner-slider">
        <div class="swiper-container">
          <div v-for="(b, i) in banners" :key="b.id" class="swiper-slide" v-show="currentSlide === i">
            <a :href="b.url">
              <img :src="b.pic" width="100%" height="100%" alt="" />
            </a>
          </div>
          <div class="pagination-dots">
            <span v-for="(_, i) in banners" :key="i" :class="['dot', { active: currentSlide === i }]" @click="goToSlide(i)" />
          </div>
        </div>
      </div>

      <div class="video-section">
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
  padding-right: 12px;
  position: sticky;
  top: 44px;
  z-index: 10;
  border-bottom: 1px solid #f4f4f4;

  .tab-bar { flex: 1; }

  .switch-btn {
    width: 44px;
    height: 44px;
    display: flex;
    justify-content: center;
    align-items: center;
    color: #999;
    cursor: pointer;
    user-select: none;
    svg { width: 16px; height: 16px; }
  }
}

.banner-slider {
  padding: 4px 4px 0;

  .swiper-container {
    position: relative;
    overflow: hidden;
    border-radius: 6px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .swiper-slide {
    width: 100%;
    img { width: 100%; display: block; object-fit: contain; }
  }

  .pagination-dots {
    position: absolute;
    bottom: 6px;
    right: 8px;
    display: flex;
    gap: 6px;
    z-index: 1;

    .dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: rgba(255,255,255,0.5);
      cursor: pointer;

      &.active {
        background: #fff;
      }
    }
  }
}

.video-section {
  padding: 4px 4px 0;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin: 12px 8px 10px;
  color: $text-primary;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 4px;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: $text-secondary;
}
</style>