<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import Header from '../../components/Header.vue'
import TabBar from '../../components/TabBar.vue'
import Drawer from '../../components/Drawer.vue'
import VideoItem from '../../components/VideoItem.vue'
import LiveInfo from '../live/LiveInfo.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getHomeContent, getBanners, getVideosByCategory } from '../../api/index'
import api from '../../api/client'
import { getHotwords } from '../../api/search'
import { getLiveIndexData } from '../../api/live'

const banners = ref([])
const categoryBanners = ref({})
const partitions = ref([])
const rankingVideos = ref([])
const additionalVideos = ref([])
const liveBanners = ref([])
const liveList = ref([])
const categoryVideos = ref({})
const partitionLoading = ref(false)
const activeTabId = ref(0)
const drawerRef = ref(null)
const loading = ref(true)
const currentSlide = ref(0)
const currentLiveSlide = ref(0)
const searchPlaceholder = ref('搜索...')
const route = useRoute()
let swiperTimer = null

onMounted(async () => {
  try {
    if (route.query.tab === 'hot') {
      activeTabId.value = -2
    } else if (route.query.tab === 'live') {
      activeTabId.value = -1
    } else {
      activeTabId.value = 0
    }
    const [contentRes, bannerRes, hotRes, liveRes] = await Promise.all([
      getHomeContent(),
      getBanners(),
      getHotwords(),
      getLiveIndexData().catch(() => ({ code: '0', data: null }))
    ])
    if (contentRes.code === '1') {
      partitions.value = contentRes.data.oneLevelPartitions || []
      additionalVideos.value = contentRes.data.additionalVideos || []
      rankingVideos.value = contentRes.data.rankingVideos || []
      console.log('Home content loaded. Recommended:', additionalVideos.value, 'Hot:', rankingVideos.value)
    } else {
      console.error('Home content load failed:', contentRes)
    }
    if (bannerRes.code === '1') {
      banners.value = bannerRes.data || []
      console.log('Banners loaded:', banners.value)
    } else {
      console.error('Banners load failed:', bannerRes)
    }
    if (hotRes.code === '1' && hotRes.data?.length) {
      const rand = hotRes.data[Math.floor(Math.random() * hotRes.data.length)]
      searchPlaceholder.value = rand.keyword || '搜索...'
    }
    if (liveRes && liveRes.code === '1') {
      liveList.value = liveRes.data?.itemList?.[0]?.lives || []
      if (liveRes.data?.bannerList?.length) {
        liveBanners.value = liveRes.data.bannerList
      } else {
        liveBanners.value = liveList.value.slice(0, 3).map((l, index) => ({
          id: l.roomId || index,
          pic: l.cover || 'https://picsum.photos/600/340?random=' + (l.roomId || index),
          url: `/m/live/${l.roomId}`
        }))
      }
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
  if (swiperTimer) clearInterval(swiperTimer)
  if (!needsAutoSwiper()) return
  swiperTimer = setInterval(() => {
    const current = getCurrentBanners()
    if (current.length > 1) {
      currentSlide.value = (currentSlide.value + 1) % current.length
    }
    if (liveBanners.value.length > 1) {
      currentLiveSlide.value = (currentLiveSlide.value + 1) % liveBanners.value.length
    }
  }, 3000)
}

const goToSlide = (i) => {
  currentSlide.value = i
  if (swiperTimer) clearInterval(swiperTimer)
  initSwiper()
}

const goToLiveSlide = (i) => {
  currentLiveSlide.value = i
  if (swiperTimer) clearInterval(swiperTimer)
  initSwiper()
}

// 热门在数据库里 category_id=0，特殊处理
function getCategoryIdForBanner(tabId) {
  if (tabId === -2) return 0
  return tabId
}

// 获取当前 tab 对应的轮播图列表
function getCurrentBanners() {
  if (activeTabId.value === 0) return banners.value
  if (activeTabId.value === -1) return liveBanners.value
  if (activeTabId.value === -2 || activeTabId.value > 0) return activeCategoryBanners.value
  return []
}

// 是否需要自动轮播（多张图才轮播）
function needsAutoSwiper() {
  const current = getCurrentBanners()
  return current.length > 1
}

// 分类轮播图 - 直接更新到响应式状态
function fetchCategoryBanners(categoryId) {
  const dbCategoryId = getCategoryIdForBanner(categoryId)
  console.log('[fetchCategoryBanners] requesting categoryId:', categoryId, '-> dbCategoryId:', dbCategoryId)
  api.get(`/banner-images/category/${dbCategoryId}`).then(res => {
    console.log('[fetchCategoryBanners] raw response for', categoryId, ':', JSON.stringify(res))
    const list = Array.isArray(res) ? res : (res?.data || [])
    console.log('[fetchCategoryBanners] parsed list for', categoryId, ':', list)
    if (list.length > 0) {
      console.log('[fetchCategoryBanners] first banner keys:', Object.keys(list[0]))
    }
    categoryBanners.value[categoryId] = list.map(b => ({
      id: b.id,
      pic: b.imageUrl || b.image_url || b.pic || '',
      url: b.linkUrl || b.link_url || '#',
      name: b.title || ''
    }))
    console.log('[fetchCategoryBanners] stored for', categoryId, ':', categoryBanners.value[categoryId])
  }).catch(e => {
    console.error('[fetchCategoryBanners] failed:', categoryId, e)
    categoryBanners.value[categoryId] = []
  })
}

const loadCategoryVideos = async (categoryId) => {
  console.log('Start loadCategoryVideos for category:', categoryId)
  if (categoryVideos.value[categoryId]) {
    console.log('Return cached videos for category:', categoryId, categoryVideos.value[categoryId])
    return
  }
  partitionLoading.value = true
  // 同时加载分类轮播图和视频
  const dbCatId = getCategoryIdForBanner(categoryId)
  const [catBannerRaw, videoRes] = await Promise.all([
    api.get(`/banner-images/category/${dbCatId}`).catch(e => {
      console.error('分类轮播图请求失败:', e)
      return null
    }),
    getVideosByCategory(categoryId)
  ])
  console.log('分类轮播图原始响应:', categoryId, catBannerRaw)
  const catBanners = Array.isArray(catBannerRaw) ? catBannerRaw
    : (catBannerRaw?.data || [])
  if (catBanners.length > 0) {
    console.log('[loadCategoryVideos] first banner keys:', categoryId, Object.keys(catBanners[0]))
  }
  // 无论是否有数据，都存储（用于判断条件）
  categoryBanners.value[categoryId] = catBanners.map(b => ({
    id: b.id,
    pic: b.imageUrl || b.image_url || b.pic || '',
    url: b.linkUrl || b.link_url || '#',
    name: b.title || ''
  }))
  console.log('分类轮播图解析后:', categoryBanners.value[categoryId])
  try {
    const res = videoRes
    console.log('loadCategoryVideos API response for category:', categoryId, res)
    if (res.code === '1') {
      categoryVideos.value[categoryId] = res.data || []
      console.log('Successfully set categoryVideos for category:', categoryId, categoryVideos.value[categoryId])
    }
  } catch (e) {
    console.error('加载分类视频失败:', e)
  } finally {
    partitionLoading.value = false
  }
}

const handleTabClick = async (tab) => {
  // 切换 tab 时重置轮播图索引
  currentSlide.value = 0
  if (tab.id === -1) {
    activeTabId.value = -1
  } else if (tab.id === 0) {
    activeTabId.value = 0
  } else if (tab.id === -2) {
    // 热门，直接使用 rankingVideos
    activeTabId.value = -2
    // 热门也有分类轮播图
    fetchCategoryBanners(-2)
  } else {
    activeTabId.value = tab.id
    await loadCategoryVideos(tab.id)
  }
  console.log('[handleTabClick] switched to tab:', tab.id, tab.name, 'activeTabId now:', activeTabId.value)
}

const tabBarData = computed(() => {
  return [
    { id: -1, name: '直播' },
    { id: 0, name: '推荐' },
    { id: -2, name: '热门' },
    ...partitions.value
  ]
})

const displayVideos = computed(() => {
  let result
  if (activeTabId.value === -2) {
    result = rankingVideos.value || []
  } else if (activeTabId.value > 0) {
    result = categoryVideos.value[activeTabId.value] || []
  } else {
    result = additionalVideos.value || []
  }
  console.log('[displayVideos] activeTabId:', activeTabId.value, 'count:', result.length, 'items:', result)
  return result
})

// 当前分类的轮播图（热门 id=-2 也算作分类）
const activeCategoryBanners = computed(() => {
  const banners = categoryBanners.value[activeTabId.value] || []
  console.log('[activeCategoryBanners] activeTabId:', activeTabId.value, 'banners:', banners, 'raw:', categoryBanners.value)
  return banners
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
      <!-- 直播轮播图 -->
      <div v-if="activeTabId === -1 && liveBanners.length" class="banner-slider">
        <div class="swiper-container">
          <div v-for="(b, i) in liveBanners" :key="b.id" class="swiper-slide" v-show="currentLiveSlide === i">
            <router-link :to="b.url">
              <img :src="b.pic" width="100%" height="100%" alt="" />
            </router-link>
          </div>
          <div class="pagination-dots">
            <span v-for="(_, i) in liveBanners" :key="i" :class="['dot', { active: currentLiveSlide === i }]" @click="goToLiveSlide(i)" />
          </div>
        </div>
      </div>
      <!-- 推荐轮播图 -->
      <div v-else-if="activeTabId === 0 && banners.length" class="banner-slider">
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
      <!-- 分类轮播图（只占一行高度，热门 id=-2 也算分类） -->
      <div v-else-if="(activeTabId > 0 || activeTabId === -2) && activeCategoryBanners.length" class="category-banner-slider">
        <div class="swiper-container">
          <div v-for="(b, i) in activeCategoryBanners" :key="b.id" class="swiper-slide" v-show="currentSlide === i">
            <a :href="b.url">
              <img :src="b.pic" width="100%" height="100%" alt="" />
            </a>
          </div>
          <div class="pagination-dots" v-if="activeCategoryBanners.length > 1">
            <span v-for="(_, i) in activeCategoryBanners" :key="i" :class="['dot', { active: currentSlide === i }]" @click="goToSlide(i)" />
          </div>
        </div>
      </div>

      <div class="video-section">
        <!-- 直播列表 -->
        <div v-if="activeTabId === -1" class="live-grid">
          <LiveInfo
            v-for="live in liveList"
            :key="live.roomId"
            :live="live"
          />
        </div>
        <!-- 分类视频加载中 -->
        <div v-else-if="partitionLoading" class="loading">加载中...</div>
        <!-- 视频列表 -->
        <div v-else class="video-grid">
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

.category-banner-slider {
  padding: 4px 4px 0;

  .swiper-container {
    position: relative;
    overflow: hidden;
    border-radius: 6px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    // 高度与一行视频格子对齐（约 56.25% 16:9 比例 + 标题高度）
    aspect-ratio: 16 / 9;
  }

  .swiper-slide {
    width: 100%;
    height: 100%;
    img { width: 100%; height: 100%; display: block; object-fit: cover; }
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

.video-grid,
.live-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: $text-secondary;
}
</style>