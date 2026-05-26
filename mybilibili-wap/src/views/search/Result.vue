<script setup>
import { ref, watch } from 'vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getSearchResult } from '../../api/search'
import { formatTenThousand } from '../../utils/format'

const props = defineProps({
  keyword: { type: String, required: true }
})

const searchType = ref('all') // all | upuser
const orderType = ref('totalrank') // totalrank | click | pubdate | dm
const videos = ref([])
const upUsers = ref([])
const loading = ref(false)
const page = ref(1)
const hasMore = ref(true)

watch(() => props.keyword, (kw) => {
  page.value = 1
  videos.value = []
  upUsers.value = []
  loadResults()
}, { immediate: true })

const loadResults = async () => {
  loading.value = true
  try {
    const res = await getSearchResult({
      keyword: props.keyword,
      page: page.value,
      size: 20,
      searchType: searchType.value,
      order: orderType.value
    })
    if (res.code === '1') {
      if (searchType.value === 'upuser') {
        if (page.value === 1) upUsers.value = res.data || []
        else upUsers.value.push(...(res.data || []))
      } else {
        if (page.value === 1) videos.value = res.data || []
        else videos.value.push(...(res.data || []))
      }
      hasMore.value = (res.data || []).length >= 20
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const switchTab = (type) => {
  searchType.value = type
  page.value = 1
  videos.value = []
  upUsers.value = []
  loadResults()
}

const switchOrder = (type) => {
  orderType.value = type
  page.value = 1
  videos.value = []
  loadResults()
}

const loadMore = () => {
  page.value++
  loadResults()
}
</script>

<template>
  <div class="result-page">
    <div class="tabs">
      <span :class="['tab', { active: searchType === 'all' }]" @click="switchTab('all')">综合</span>
      <span :class="['tab', { active: searchType === 'upuser' }]" @click="switchTab('upuser')">UP主</span>
    </div>

    <div v-if="searchType === 'all' && page === 1" class="order-tabs">
      <span :class="['order-tab', { active: orderType === 'totalrank' }]" @click="switchOrder('totalrank')">默认</span>
      <span :class="['order-tab', { active: orderType === 'click' }]" @click="switchOrder('click')">播放多</span>
      <span :class="['order-tab', { active: orderType === 'pubdate' }]" @click="switchOrder('pubdate')">新发布</span>
      <span :class="['order-tab', { active: orderType === 'dm' }]" @click="switchOrder('dm')">弹幕多</span>
    </div>

    <div class="result-content">
      <div v-if="loading && page === 1" class="loading">加载中...</div>
      <template v-else>
        <!-- 视频结果 -->
        <div v-if="searchType === 'all'" class="video-list">
          <VideoItem v-for="v in videos" :key="v.aId" :video="v" />
          <div v-if="hasMore" class="load-more" @click="loadMore">加载更多</div>
          <div v-else-if="videos.length" class="no-more">— 到底了 —</div>
          <div v-if="!videos.length" class="empty">未找到结果</div>
        </div>

        <!-- UP主结果 -->
        <div v-else class="up-list">
          <div v-for="up in upUsers" :key="up.mid" class="up-item" @click="$router.push(`/m/space/${up.mid}`)">
            <img :src="up.face || '../../../assets/noface.gif'" class="up-avatar" />
            <div class="up-info">
              <p class="up-name">{{ up.name }}</p>
              <p class="up-stats">
                {{ formatTenThousand(up.fans || up.follower) }}粉丝 · {{ up.videos || 0 }}视频
              </p>
              <p v-if="up.sign" class="up-sign">{{ up.sign }}</p>
            </div>
          </div>
          <div v-if="hasMore" class="load-more" @click="loadMore">加载更多</div>
          <div v-else-if="upUsers.length" class="no-more">— 到底了 —</div>
          <div v-if="!upUsers.length" class="empty">未找到结果</div>
        </div>
      </template>
    </div>
    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.result-page { padding-bottom: 20px; }

.tabs {
  display: flex;
  background: $bg-white;
  border-bottom: 1px solid $border-color;
}

.tab {
  flex: 1;
  text-align: center;
  padding: 12px;
  font-size: 15px;
  color: $text-secondary;
  cursor: pointer;

  &.active {
    color: $theme-pink;
    font-weight: 600;
  }
}

.order-tabs {
  display: flex;
  padding: 8px 12px;
  background: $bg-white;
  gap: 8px;
  border-bottom: 1px solid $border-color;
}

.order-tab {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  color: $text-secondary;
  background: $bg-color;
  cursor: pointer;

  &.active {
    background: $theme-pink;
    color: #fff;
  }
}

.video-list { padding: 12px; }
.video-list .video-item { margin-bottom: 10px; }

.up-list { padding: 12px; }

.up-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid $border-color;
  cursor: pointer;
}

.up-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #e3e5e7;
}

.up-info { flex: 1; }
.up-name { font-size: 14px; font-weight: 600; margin-bottom: 4px; }
.up-stats { font-size: 12px; color: $text-secondary; margin-bottom: 4px; }
.up-sign { font-size: 12px; color: $text-third; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.load-more {
  text-align: center;
  padding: 12px;
  color: $theme-pink;
  cursor: pointer;
  font-size: 14px;
}

.no-more, .empty {
  text-align: center;
  padding: 20px;
  color: $text-third;
  font-size: 13px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: $text-secondary;
}
</style>