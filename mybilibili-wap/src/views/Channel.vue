<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Header from '../components/Header.vue'
import TabBar from '../components/TabBar.vue'
import Drawer from '../components/Drawer.vue'
import VideoItem from '../components/VideoItem.vue'
import ScrollToTop from '../components/ScrollToTop.vue'
import { getRankingPartitions, getRankingRegion } from '../api/channel'

const route = useRoute()
const rId = parseInt(route.params.rId) || 0

const tabBarData = ref([])
const activeId = ref(rId)
const drawerRef = ref(null)
const loading = ref(true)
const recommendVideos = ref([])
const partitions = ref([])

onMounted(async () => {
  try {
    const res = await getRankingPartitions()
    if (res.code === '1') {
      tabBarData.value = res.data || []
      tabBarData.value.unshift({ id: 0, name: '首页', pids: [] })
      tabBarData.value.push({ id: -1, name: '直播' })
    }
    await loadPartitionVideos()
  } catch (e) {
    console.error('加载失败:', e)
  } finally {
    loading.value = false
  }
})

const loadPartitionVideos = async () => {
  loading.value = true
  try {
    const res = await getRankingRegion(rId)
    if (res.code === '1') {
      const data = res.data || {}
      recommendVideos.value = data.hotVideos?.slice(0, 4) || []
      partitions.value = data.partitions || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleTabClick = (tab) => {
  activeId.value = tab.id
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
</script>

<template>
  <div class="channel-page">
    <Header />
    <div class="partition-bar">
      <TabBar :data="tabBarData" :active-id="activeId" @click="handleTabClick" />
      <div class="switch-btn" @click="drawerRef?.show()">▼</div>
    </div>
    <Drawer :data="tabBarData" ref="drawerRef" @click="handleTabClick" />

    <div class="content" v-if="!loading">
      <!-- 热门推荐 -->
      <div v-if="recommendVideos.length" class="recommend-section">
        <div class="section-title">
          <router-link :to="`/m/ranking/${rId}`">热门推荐 &gt;</router-link>
        </div>
        <div class="video-grid">
          <VideoItem v-for="v in recommendVideos" :key="v.aId" :video="v" />
        </div>
      </div>

      <!-- 分区内容 -->
      <div v-for="partition in partitions" :key="partition.id" class="partition-block">
        <div class="partition-header">
          <span class="partition-name">{{ partition.name }}</span>
        </div>
        <div class="video-grid">
          <VideoItem v-for="v in (partition.videos || [])" :key="v.aId" :video="v" />
        </div>
      </div>
    </div>

    <div v-else class="loading">加载中...</div>
    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.channel-page { padding-bottom: 20px; }

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
  }
}

.content { padding: 12px; }

.recommend-section, .partition-block { margin-bottom: 16px; }

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
  a { color: $theme-pink; }
}

.partition-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid $border-color;
}
.partition-name { font-size: 16px; font-weight: 600; }

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