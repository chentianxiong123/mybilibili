<script setup>
import { ref, onMounted } from 'vue'
import Header from '../../components/Header.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getLiveIndexData } from '../../api/live'
import LiveInfo from './LiveInfo.vue'

const bannerList = ref([])
const itemList = ref([]) // { areaName, lives: Live[] }
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getLiveIndexData()
    if (res.code === '1') {
      bannerList.value = res.data?.bannerList || []
      itemList.value = res.data?.itemList || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
  initSwiper()
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
</script>

<template>
  <div class="live-index-page">
    <Header />

    <div v-if="bannerList.length" class="banner-slider">
      <div class="swiper-container">
        <div v-for="b in bannerList" :key="b.id" class="swiper-slide">
          <a :href="b.url">
            <img :src="b.pic" width="100%" alt="" />
          </a>
        </div>
      </div>
      <div class="pagination-dots">
        <span v-for="(_, i) in bannerList" :key="i" :class="['dot', { active: i === 0 }]" />
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <template v-else>
      <div v-for="group in itemList" :key="group.areaName" class="live-group">
        <div class="group-header">
          <span class="group-name">{{ group.areaName }}</span>
          <router-link :to="`/m/live/list?areaId=${group.areaId || 0}`" class="more-link">
            进去看看 &gt;
          </router-link>
        </div>
        <div class="live-grid">
          <LiveInfo v-for="live in group.lives" :key="live.roomId" :live="live" />
        </div>
      </div>

      <div class="bottom-nav">
        <router-link to="/m/live/list" class="nav-btn">全部直播</router-link>
        <router-link to="/m/live/areas" class="nav-btn">全部分类</router-link>
      </div>
    </template>

    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.live-index-page { padding-bottom: 60px; }

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
    &.active { background: $theme-pink; }
  }
}

.live-group {
  padding: 12px;
  background: $bg-white;
  margin-bottom: 8px;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.group-name { font-size: 15px; font-weight: 600; }

.more-link { font-size: 12px; color: $text-secondary; }

.live-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.bottom-nav {
  display: flex;
  gap: 12px;
  padding: 12px;
  justify-content: center;
}

.nav-btn {
  padding: 8px 20px;
  border-radius: 16px;
  background: $theme-pink;
  color: #fff;
  font-size: 14px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: $text-secondary;
}
</style>
