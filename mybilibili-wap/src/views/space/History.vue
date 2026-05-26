<script setup lang="ts">
import { ref, onMounted } from 'vue'
import VideoItem from '../../components/VideoItem.vue'
import { historyApi } from '../../api/index'

const histories = ref<Array<{date: string; items: any[]}>>([])
const loading = ref(true)
const page = ref(1)
const hasMore = ref(true)

onMounted(async () => {
  await loadHistory()
})

const loadHistory = async () => {
  loading.value = true
  try {
    const res = await historyApi.getHistoryList(page.value, 20)
    const list = res?.data?.records || res?.data || []
    const groupMap: Record<string, any[]> = {}
    const now = new Date()
    list.forEach((item: any) => {
      const d = new Date(item.viewTime || item.createTime)
      const diff = Math.floor((now.getTime() - d.getTime()) / 86400000)
      let label: string
      if (diff === 0) label = '今天'
      else if (diff === 1) label = '昨天'
      else if (diff === 2) label = '前天'
      else label = '更早'
      if (!groupMap[label]) groupMap[label] = []
      groupMap[label].push({
        aId: item.manuscriptId || item.id,
        title: item.manuscriptTitle || item.title || '',
        pic: item.coverUrl || item.cover || '',
        timeStr: item.viewTime || item.createTime || ''
      })
    })
    histories.value = Object.entries(groupMap).map(([date, items]) => ({ date, items }))
  } catch (e) {
    console.error('加载历史��录失败:', e)
  } finally {
    loading.value = false
  }
}

const clearAll = async () => {
  try {
    await historyApi.clearHistory()
    histories.value = []
  } catch (e) {
    console.error(e)
  }
}
</script>

<template>
  <div class="history-page">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="histories.length" class="history-list">
      <div v-for="group in histories" :key="group.date" class="history-group">
        <div class="group-title">{{ group.date }}</div>
        <div class="history-items">
          <div v-for="item in group.items" :key="item.aId" class="history-item">
            <router-link :to="`/m/video/${item.aId}`" class="history-link">
              <img :src="item.pic" class="history-cover" />
              <div class="history-info">
                <p class="history-title">{{ item.title }}</p>
                <p class="history-time">{{ item.timeStr }}</p>
              </div>
            </router-link>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="empty">
      <p>你还没有历史记录</p>
    </div>
    <div v-if="histories.length" class="clear-btn-wrap">
      <span class="clear-btn" @click="clearAll">清空历史</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.history-page { padding: 12px; padding-bottom: 60px; }

.history-group { margin-bottom: 16px; }

.group-title {
  font-size: 13px;
  color: $text-secondary;
  margin-bottom: 8px;
  padding-left: 4px;
}

.history-items { background: $bg-white; border-radius: 8px; overflow: hidden; }

.history-item { border-bottom: 1px solid $border-color;
  &:last-child { border-bottom: none; }
}

.history-link {
  display: flex;
  gap: 10px;
  padding: 10px;
}

.history-cover {
  width: 100px;
  height: 56px;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
  background: #e3e5e7;
}

.history-info { flex: 1; }

.history-title {
  font-size: 13px;
  color: $text-primary;
  line-height: 1.4;
  margin-bottom: 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.history-time { font-size: 12px; color: $text-third; }

.loading, .empty {
  text-align: center;
  padding: 40px;
  color: $text-secondary;
}

.clear-btn-wrap {
  text-align: center;
  padding: 16px;
}

.clear-btn {
  color: $theme-pink;
  font-size: 14px;
  cursor: pointer;
}
</style>