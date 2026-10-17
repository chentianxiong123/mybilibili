<script setup lang="ts">
import { ref, onMounted, nextTick, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, ArrowRight, View, Star } from '@element-plus/icons-vue'
import { videoApi, categoryApi } from '@/api/client'
import { getCategoryBanners } from '@/api/banner.ts'
import { formatDuration, formatMonthDay, normalizeVideoCard } from '@/utils/videoCard.ts'
import { useTabsVideoAlign } from '@/composables/useTabsVideoAlign'
import VideoCard from '@/components/VideoCard.vue'

const route = useRoute()
const categoryId = ref(route.params.id)
const categoryName = ref('')

// 轮播图数据
const bannerList = ref([])

// 轮播图组件引用
const carouselRef = ref(null)

// 轮播图当前索引
const currentBannerIndex = ref(0)

// 轮播图控制函数
const handleBannerChange = (index) => {
  currentBannerIndex.value = index
}

const changeBanner = (index) => {
  currentBannerIndex.value = index
  if (carouselRef.value) {
    carouselRef.value.setActiveItem(index)
  }
}

const prevBanner = () => {
  currentBannerIndex.value = (currentBannerIndex.value - 1 + bannerList.value.length) % bannerList.value.length
  if (carouselRef.value) {
    carouselRef.value.prev()
  }
}

const nextBanner = () => {
  currentBannerIndex.value = (currentBannerIndex.value + 1) % bannerList.value.length
  if (carouselRef.value) {
    carouselRef.value.next()
  }
}

// 分类视频列表
const videoList = ref([])
// 标记是否已经加载过视频列表
const hasLoadedVideos = ref(false)
// 加载状态
const loading = ref(false)

const { apply: applyTabsVideoAlign } = useTabsVideoAlign()

// 动态设置轮播图高度，使其与第一行视频格子的下边对齐
const adjustBannerHeight = () => {
  if (import.meta.server) return
  nextTick(() => {
    const videoItems = document.querySelectorAll('.video-item')
    const bannerSection = document.querySelector('.banner-section')
    if (!bannerSection) return
    if (videoItems.length > 0) {
      const firstItem = videoItems[0]
      const videoGrid = document.querySelector('.video-grid')
      if (firstItem && videoGrid) {
        const gridRect = videoGrid.getBoundingClientRect()
        const itemRect = firstItem.getBoundingClientRect()
        const bannerHeight = itemRect.bottom - gridRect.top
        bannerSection.style.height = `${bannerHeight}px`
      }
    } else {
      bannerSection.style.height = '200px'
    }
  })
}

// 从后端API获取分类视频列表数据
const fetchVideoList = async () => {
  // 防止重复加载
  if (hasLoadedVideos.value) {
    return
  }
  
  loading.value = true
  try {
    let response
    // 根据分类ID选择不同的API
    if (categoryId.value === '0') {
      // 热门标签，获取热门视频
      response = await videoApi.getHotVideos()
    } else {
      // 普通分类，获取分类视频
      response = await videoApi.getVideosByCategoryId(categoryId.value)
    }
    if (response.code === 200) {
      const videos = response.data?.list || response.data || []
      
      // 清空现有视频列表
      videoList.value = []
      
      // 添加去重逻辑
      const uniqueVideos = []
      const videoIds = new Set()
      
      if (videos && Array.isArray(videos)) {
        for (const video of videos) {
          const videoCard = normalizeVideoCard(video)
          if (videoCard) {
            if (!videoIds.has(videoCard.manuscriptId)) {
              videoIds.add(videoCard.manuscriptId)
              uniqueVideos.push(videoCard)
            }
          }
        }
      }
      
      videoList.value = uniqueVideos
      
      // 标记视频列表已经加载过
      hasLoadedVideos.value = true
    }
  } catch (error) {
    console.error('获取视频列表失败:', error)
  } finally {
    loading.value = false
    // 视频列表加载完成后，调整轮播图高度
    adjustBannerHeight()
  }
}

// 从API获取分类轮播图数据
const fetchCategoryBanners = async (catId) => {
  try {
    const res = await getCategoryBanners(catId)
    if (res.code === 200 && res.data && res.data.length > 0) {
      bannerList.value = res.data.map(banner => ({
        id: banner.id,
        img: banner.imageUrl,
        link: banner.linkUrl || '/',
        title: banner.title
      }))
    } else {
      bannerList.value = []
    }
  } catch (error) {
    console.error('获取分类轮播图失败:', error)
    bannerList.value = []
  }
}

// 跳转到稿件详情页
const goToVideo = (video) => {
  // 优先使用稿件ID
  const manuscriptId = video.manuscriptId || video.id
  window.location.href = `/manuscript/${manuscriptId}`
}

// 跳转到作者主页
const goToAuthor = (authorId) => {
  if (authorId) {
    window.open(`/profile/${authorId}/home`, '_blank')
  }
}

const categoryMap = ref({})

// 全部分区列表（原始，含 parentId）
const categoryListAll = ref([])
// 当前分区对象（含 parentId/子分区）
const currentCategory = ref(null)
// 当前一级分区下的子分区 tab（一级分区时展示）
const subCategories = ref([])
// 当前是否热门(0)标签页
const isHotTab = computed(() => categoryId.value === '0')
// 当前是否为二级分区（有父级）
const isSubCategory = computed(() => {
  const c = currentCategory.value
  return c && (c.parentId === 0 ? false : true)
})

const fetchCategoryName = async () => {
  try {
    const res = await categoryApi.getCategoryList()
    const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : (res?.list || []))
    const map = {}
    const all = []
    for (const cat of list) {
      map[cat.id] = cat.name
      all.push(cat)
    }
    map[0] = '热门'
    categoryMap.value = map
    categoryListAll.value = all
    updateCurrentCategory()
  } catch (error) {
    console.error('获取分类列表失败:', error)
  }
}

// 根据当前 categoryId 计算分类名、是否一级分区、子分区列表
const updateCurrentCategory = () => {
  const id = categoryId.value
  const numId = Number(id)
  if (!id || isHotTab.value) {
    currentCategory.value = null
    subCategories.value = []
    categoryName.value = id === '0' ? '热门' : (categoryMap.value[id] || '未知分类')
    return
  }
  const cat = categoryListAll.value.find(c => Number(c.id) === numId) || null
  currentCategory.value = cat
  categoryName.value = cat ? cat.name : (categoryMap.value[numId] || '未知分类')
  // 一级分区：列出直接子分区 tab
  if (cat && cat.parentId === 0) {
    subCategories.value = categoryListAll.value.filter(c => Number(c.parentId) === numId)
  } else {
    subCategories.value = []
  }
}

const goSubCategory = (cat) => {
  router.push(`/category/${cat.id}`)
}

// 二级分区的一级父分区（用于返回聚合页）
const parentCategory = computed(() => {
  const c = currentCategory.value
  if (!c || c.parentId === 0) return null
  return categoryListAll.value.find(x => Number(x.id) === Number(c.parentId)) || null
})
const parentCategoryName = computed(() => parentCategory.value?.name || '上级分区')

const goParentCategory = () => {
  if (parentCategory.value) router.push(`/category/${parentCategory.value.id}`)
}

const router = useRouter()

// 监听路由参数变化，当categoryId变化时重新加载数据
watch(() => route.params.id, (newId) => {
  if (!import.meta.client || !newId) return
  categoryId.value = newId
    hasLoadedVideos.value = false
    fetchCategoryBanners(newId)
    fetchVideoList()
    updateCurrentCategory()

    nextTick(() => {
      adjustBannerHeight()
      applyTabsVideoAlign()
    })
}, { immediate: true })

onMounted(async () => {
  await fetchCategoryName()
  adjustBannerHeight()
  applyTabsVideoAlign()
  
  window.addEventListener('resize', adjustBannerHeight)
})
</script>

<template>
  <!-- 主内容区域：5列网格布局 -->
  <div class="main-section">
    <!-- 分类标题 -->
    <div class="category-header">
      <h1 class="category-title">{{ categoryName }}</h1>
      <div class="category-sub" v-if="isHotTab">
        <div class="sub-tabs">
          <div
            :class="['sub-tab', { 'active': categoryId === '0' }]"
            @click="router.push('/category/0')"
          >热门</div>
        </div>
      </div>
      <div class="category-sub" v-else-if="isSubCategory">
        <!-- 二级分区：显示其一级父分区名，回跳父分区聚合页 -->
        <div class="sub-tabs">
          <div class="sub-tab back-tab" @click="goParentCategory()">
            <el-icon><ArrowLeft /></el-icon>
            <span>{{ parentCategoryName }}</span>
          </div>
          <div class="sub-tab-sub current">{{ categoryName }}</div>
        </div>
      </div>
      <div class="category-sub" v-else-if="subCategories.length > 0">
        <div class="sub-tabs">
          <div
            :class="['sub-tab', { 'active': currentCategory && Number(categoryId) === Number(currentCategory.id) }]"
            @click="goSubCategory(currentCategory)"
          >全部</div>
          <div
            v-for="sub in subCategories"
            :key="sub.id"
            :class="['sub-tab', { 'active': Number(categoryId) === Number(sub.id) }]"
            @click="goSubCategory(sub)"
          >{{ sub.name }}</div>
        </div>
      </div>
    </div>
    
    <!-- 统一的视频网格：包含轮播图和视频项 -->
    <div class="video-grid">
      <!-- 左侧轮播图：占据2x1的位置（只占一行高度） -->
      <section class="banner-section" v-if="bannerList.length > 0">
        <div class="banner-wrapper">
          <el-carousel ref="carouselRef" :autoplay="true" :interval="3000" arrow="never" indicator-position="none" @change="handleBannerChange">
            <el-carousel-item v-for="(banner, index) in bannerList" :key="banner.id">
              <a :href="banner.link" class="banner-item">
                <img loading="lazy" decoding="async" :src="banner.img" alt="轮播图" class="banner-img">
              </a>
            </el-carousel-item>
          </el-carousel>
          <!-- 底部透明黑色区域：固定在底部，不随图片切换 -->
          <div class="banner-overlay">
            <div class="banner-content">
              <!-- 左边：标题和点位 -->
              <div class="banner-left">
                <h3 class="banner-title">{{ bannerList[currentBannerIndex]?.title }}</h3>
                <div class="banner-indicators">
                  <span
                    v-for="(item, i) in bannerList"
                    :key="i"
                    :class="['indicator', { active: i === currentBannerIndex }]"
                    @click="changeBanner(i)"
                  ></span>
                </div>
              </div>
              <!-- 右边：左右箭头 -->
              <div class="banner-right">
                <button class="arrow-btn arrow-prev" @click="prevBanner">
                  <el-icon><ArrowLeft /></el-icon>
                </button>
                <button class="arrow-btn arrow-next" @click="nextBanner">
                  <el-icon><ArrowRight /></el-icon>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
      
      <!-- 视频项：显示分类视频列表 -->
      <VideoCard
        v-for="video in videoList"
        :key="video.id"
        :video="video"
      />
    </div>
  </div>
</template>

<style scoped>
/* 主内容区域：5列网格布局 */
.main-section {
  width: 100%;
  max-width: 1980px;
  margin: 0 auto;
  padding: 0 20px;
  box-sizing: border-box;
  margin-bottom: 20px;
}

/* 分类标题 */
.category-header {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.category-title {
  font-size: 28px;
  font-weight: bold;
  color: #333;
  margin: 0;
}

/* 子分区导航条 */
.category-sub {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f1f2;
}

.sub-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.sub-tab {
  padding: 7px 18px;
  font-size: 14px;
  color: #333;
  background-color: #f5f5f5;
  border-radius: 18px;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}

.sub-tab:hover {
  background-color: #e0e0e0;
  color: #00a1d6;
}

.sub-tab.active {
  background-color: #00a1d6;
  color: #fff;
}

.sub-tab.back-tab {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background-color: #fff;
  border: 1px solid #e0e0e0;
}

.sub-tab.back-tab:hover {
  border-color: #00a1d6;
}

.sub-tab-sub {
  padding: 7px 18px;
  font-size: 14px;
  color: #00a1d6;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
}

/* 统一的视频网格：5列布局，自动行 */
.video-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  grid-auto-rows: auto;
  gap: 20px;
  width: 100%;
}

/* 轮播图：占据2x1的位置（只占一行高度） */
.banner-section {
  grid-column: 1 / 3;
  grid-row: 1 / 2;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  position: relative;
  height: 200px;
}

.video-item {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.video-item:hover {
  transform: none;
  box-shadow: none;
}

.video-cover-link {
  text-decoration: none;
  color: inherit;
  display: inline-block;
}

.video-cover-link img {
  display: block;
}

.video-title {
  font-size: 15px;
  font-weight: 500;
  color: #212121;
  margin: 10px 10px 4px 10px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  height: 42px; /* 固定两行高度：15px * 1.4 * 2 = 42px */
  flex: 0 0 auto;
  pointer-events: none; /* 外层不响应鼠标事件 */
}

.video-title-text {
  cursor: pointer;
  transition: color 0.3s;
  pointer-events: auto; /* 内层响应鼠标事件 */
}

.video-title-text:hover {
  color: #00aeec;
}

.video-meta {
  margin: 0 10px 6px 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 0 0 auto;
  min-width: 0;
}

.video-author {
  font-size: 11px;
  color: #9499a0;
  cursor: pointer;
  transition: color 0.3s;
  min-width: 0;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-author:hover {
  color: #00aeec;
}

.video-date {
  font-size: 11px;
  color: #9499a0;
  flex: 0 0 auto;
}

.video-separator {
  color: #9499a0;
  font-size: 11px;
  flex: 0 0 auto;
}

.video-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  border-radius: 6px;
  overflow: hidden;
  background-color: #f5f5f5;
}

.video-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.source-badge {
  position: absolute;
  top: 6px;
  left: 6px;
  z-index: 2;
  font-size: 11px;
  line-height: 1;
  padding: 3px 6px;
  border-radius: 3px;
  color: #fff;
  background-color: rgba(0, 0, 0, 0.55);
  border: 1px solid rgba(255, 255, 255, 0.35);
}

.video-cover:hover img {
  transform: scale(1.05);
  transition: transform 0.3s;
}

.video-stats-overlay {
  position: absolute;
  bottom: 8px;
  left: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  background-color: transparent;
  padding: 4px 8px;
  border-radius: 4px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #fff;
  font-size: 12px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
}

.stat-item .el-icon {
  font-size: 12px;
}

.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: transparent;
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
}

.banner-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.banner-section :deep(.el-carousel) {
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.banner-section :deep(.el-carousel__container) {
  height: 100%;
  flex: 1;
}

.banner-section :deep(.el-carousel__item) {
  height: 100%;
}

.banner-item {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
  flex: 1;
}

.banner-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.banner-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.8) 0%, rgba(0, 0, 0, 0.6) 50%, rgba(0, 0, 0, 0) 100%);
  padding: 20px;
  display: flex;
  align-items: flex-end;
  z-index: 10;
}

.banner-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  width: 100%;
}

.banner-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.banner-title {
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
}

.banner-indicators {
  display: flex;
  gap: 8px;
}

.indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: all 0.3s;
}

.indicator:hover {
  background-color: rgba(255, 255, 255, 0.8);
}

.indicator.active {
  background-color: #fff;
  transform: scale(1.2);
}

.banner-right {
  display: flex;
  gap: 10px;
}

.arrow-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.arrow-btn:hover {
  background-color: rgba(255, 255, 255, 0.4);
  transform: scale(1.1);
}

.arrow-btn:active {
  transform: scale(0.95);
}

.arrow-btn .el-icon {
  font-size: 18px;
}

/* 响应式：调整左右白边 */
@media (max-width: 2560px) {
  .main-section {
    padding-left: 120px;
    padding-right: 120px;
  }
}

@media (max-width: 2200px) {
  .main-section {
    padding-left: 100px;
    padding-right: 100px;
  }
}

@media (max-width: 1920px) {
  .main-section {
    padding-left: 80px;
    padding-right: 80px;
  }
}












</style>
