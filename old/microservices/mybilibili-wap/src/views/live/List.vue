<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import LiveInfo from './LiveInfo.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getLiveListData } from '../../api/live'

const route = useRoute()
const parentAreaId = parseInt(route.query.parent_area_id) || 0
const areaId = parseInt(route.query.areaId) || 0

const lives = ref([])
const loading = ref(true)
const page = ref(1)
const hasMore = ref(true)
const areaName = ref('直播列表')

onMounted(async () => {
  await loadData()
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await getLiveListData({ parentAreaId, areaId, page: page.value, pageSize: 20 })
    if (res.code === '1') {
      const data = res.data || {}
      areaName.value = data.areaName || '直播列表'
      if (page.value === 1) {
        lives.value = data.list || []
      } else {
        lives.value.push(...(data.list || []))
      }
      hasMore.value = (data.list || []).length >= 20
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  page.value++
  loadData()
}
</script>

<template>
  <div class="live-list-page">
    <div class="list-header">
      <router-link to="/m/live" class="back-btn">‹</router-link>
      <span class="title">{{ areaName }}</span>
    </div>

    <div class="live-grid-wrap">
      <div v-if="loading && page === 1" class="loading">加载中...</div>
      <div v-else-if="lives.length" class="live-grid">
        <LiveInfo v-for="live in lives" :key="live.roomId" :live="live" />
      </div>
      <div v-else class="empty">暂无直播</div>
      <div v-if="hasMore && !loading" class="load-more" @click="loadMore">
        请给我更多！
      </div>
      <div v-if="!hasMore && lives.length" class="no-more">没有更多了！</div>
    </div>
    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.live-list-page { min-height: 100vh; background: $bg-color; }

.list-header {
  display: flex;
  align-items: center;
  height: $header-height;
  padding: 0 12px;
  background: $theme-pink;
  color: #fff;

  .back-btn { font-size: 28px; width: 40px; }
  .title { font-size: 16px; font-weight: 600; }
}

.live-grid-wrap { padding: 12px; }

.live-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.loading, .empty {
  text-align: center;
  padding: 40px;
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
}
</style>
