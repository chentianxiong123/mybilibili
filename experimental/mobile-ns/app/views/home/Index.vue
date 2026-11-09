<template>
  <Page class="page">
    <ActionBar title="哔哩哔哩" />
    <ScrollView>
      <StackLayout>
        <GridLayout class="partition-bar" height="44" @tap="onDrawerToggle">
          <ScrollView orientation="horizontal" scrollBarIndicatorVisible="false">
            <StackLayout orientation="horizontal" class="tabs-scroll">
              <Label v-for="tab in tabBarData.slice(0, 6)" :key="tab.id" :text="tab.name"
                class="tab-item" :class="{ active: activeTabId === tab.id }"
                @tap="onTabClick(tab)" />
            </StackLayout>
          </ScrollView>
          <Label text="▼" class="switch-btn" />
        </GridLayout>

        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else>
          <StackLayout class="banner-slider" v-if="activeTabId === 0 && banners.length > 0">
            <Image :src="banners[currentSlide]?.pic || ''" class="banner-img" />
            <Label :text="getBannerTitle(banners[currentSlide])" class="banner-caption" textWrap="true" />
          </StackLayout>

          <StackLayout class="video-section">
            <GridLayout columns="*, *" class="video-grid" v-if="displayVideos.length > 0">
              <VideoItem v-for="(v, i) in displayVideos" :key="v.aId || i" :video="v" :col="i % 2" />
            </GridLayout>
            <Label v-else text="暂无内容" class="loading" />
          </StackLayout>
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from 'nativescript-vue'
import { $navigateTo, $showModal } from 'nativescript-vue'
import VideoItem from '../../components/VideoItem.vue'
import { getHomeContent, getBanners, getVideosByCategory } from '../../api/index'
import api from '../../api/client'
import { getHotwords } from '../../api/search'
import { getLiveIndexData } from '../../api/live'

const banners = ref<any[]>([])
const partitions = ref<any[]>([])
const rankingVideos = ref<any[]>([])
const additionalVideos = ref<any[]>([])
const liveList = ref<any[]>([])
const loadMoreVideos = ref<any[]>([])
const categoryVideos = ref<Record<number, any[]>>({})
const loading = ref(true)
const activeTabId = ref(0)
const currentSlide = ref(0)
const searchPlaceholder = ref('搜索...')

let swiperTimer: any = null

onMounted(async () => {
  try {
    const [contentRes, bannerRes, hotRes, liveRes] = await Promise.all([
      getHomeContent(),
      getBanners(),
      getHotwords(),
      getLiveIndexData().catch(() => ({ code: '0', data: null }))
    ])
    if (contentRes.code === '1') {
      partitions.value = contentRes.data?.oneLevelPartitions || []
      additionalVideos.value = contentRes.data?.additionalVideos || []
      rankingVideos.value = contentRes.data?.rankingVideos || []
    }
    if (bannerRes.code === '1') {
      banners.value = bannerRes.data || []
    }
    if (hotRes.code === '1' && hotRes.data?.length) {
      const rand = hotRes.data[Math.floor(Math.random() * hotRes.data.length)]
      searchPlaceholder.value = rand.keyword || '搜索...'
    }
    if (liveRes && liveRes.code === '1') {
      liveList.value = liveRes.data?.itemList?.[0]?.lives || []
    }
    initSwiper()
  } catch (e) {
    console.error('首页加载失败:', e)
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (swiperTimer) clearInterval(swiperTimer)
})

function initSwiper() {
  if (swiperTimer) clearInterval(swiperTimer)
  if (banners.value.length > 1) {
    swiperTimer = setInterval(() => {
      currentSlide.value = (currentSlide.value + 1) % banners.value.length
    }, 3000)
  }
}

function getBannerTitle(banner: any) {
  return banner?.name || banner?.title || ''
}

const tabBarData = computed(() => {
  return [
    { id: 0, name: '推荐' },
    ...partitions.value.map((p: any) => ({ id: p.id, name: p.name }))
  ]
})

const displayVideos = computed(() => {
  if (activeTabId.value > 0) {
    return categoryVideos.value[activeTabId.value] || []
  }
  return additionalVideos.value || []
})

async function onTabClick(tab: any) {
  activeTabId.value = tab.id
  currentSlide.value = 0
  if (tab.id > 0 && !categoryVideos.value[tab.id]) {
    try {
      const res = await getVideosByCategory(tab.id)
      if (res.code === '1') {
        categoryVideos.value[tab.id] = res.data || []
      }
    } catch (e) {
      console.error('加载分类视频失败:', e)
    }
  }
}

function onDrawerToggle() {
  $showModal(require('../channel/Channel.vue').default, {
    fullscreen: true,
    props: { partitions: partitions.value }
  })
}

function onSearch() {
  $showModal(require('../../views/search/Search.vue').default, { fullscreen: true })
}
</script>

<style scoped>
.partition-bar {
  background-color: white;
  padding-right: 12;
  border-bottom-width: 1;
  border-bottom-color: #f4f4f4;
}

.tabs-scroll {
  padding-left: 8;
}

.tab-item {
  padding: 8 14;
  font-size: 13;
  color: #61666d;
}

.tab-item.active {
  color: #fb7299;
  font-weight: bold;
}

.switch-btn {
  font-size: 12;
  color: #999;
  padding: 10 8;
}

.banner-slider {
  padding: 6;
}

.banner-img {
  width: 100%;
  height: 180;
  border-radius: 5;
}

.banner-caption {
  padding: 8 12;
  font-size: 14;
  color: #18191c;
  font-weight: 600;
}

.video-section {
  padding: 4;
}

.video-grid {
  padding: 2;
}
</style>