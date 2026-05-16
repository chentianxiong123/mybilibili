<template>
  <div class="overview-panel">
    <h3 class="panel-title">数据概览</h3>

    <DataCards :cards="statCards" />

    <div class="chart-section">
      <div class="chart-header">
        <h4>数据趋势</h4>
        <div class="chart-actions">
          <el-radio-group v-model="trendDays" size="small" @change="switchTrendDays">
            <el-radio-button :value="7">7天</el-radio-button>
            <el-radio-button :value="30">30天</el-radio-button>
            <el-radio-button :value="90">90天</el-radio-button>
          </el-radio-group>
          <el-radio-group v-model="trendMetric" size="small" @change="switchTrendMetric">
            <el-radio-button value="views">播放量</el-radio-button>
            <el-radio-button value="likes">点赞</el-radio-button>
            <el-radio-button value="comments">评论</el-radio-button>
            <el-radio-button value="followers">粉丝</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      <TrendChart
        :xData="visibleTrendData.dates"
        :series="trendSeries"
        :height="300"
        :loading="loading.trend"
      />
    </div>

    <div class="chart-section">
      <div class="chart-header">
        <h4>稿件TOP5排行</h4>
      </div>
      <TrendChart
        :xData="rankingXData"
        :series="rankingSeries"
        :height="260"
        :loading="loading.ranking"
        :showLegend="false"
        horizontal
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import DataCards from './DataCards.vue'
import TrendChart from './TrendChart.vue'
import { useCreatorStats } from '@/composables/useCreatorStats'
import { VideoPlay, Star, ChatDotRound, ChatLineSquare, User, Coin, FolderAdd, Share } from '@element-plus/icons-vue'

const {
  loading,
  overview,
  trendData,
  rankingList,
  loadOverview,
  loadTrend,
  loadRanking,
  loadFansTrend
} = useCreatorStats()

const trendMetric = ref('views')
const trendDays = ref(7)

const statCards = computed(() => [
  { key: 'views', label: '播放量', value: overview.value.totalViews, increase: overview.value.viewsIncrease, bgColor: '#e6f7ff', icon: VideoPlay, highlight: true },
  { key: 'likes', label: '点赞', value: overview.value.totalLikes, increase: overview.value.likesIncrease, bgColor: '#fff0f6', icon: Star },
  { key: 'comments', label: '评论', value: overview.value.totalComments, increase: overview.value.commentsIncrease, bgColor: '#fff7e6', icon: ChatDotRound },
  { key: 'danmaku', label: '弹幕', value: overview.value.totalDanmaku, increase: overview.value.danmakuIncrease, bgColor: '#f6ffed', icon: ChatLineSquare },
  { key: 'followers', label: '粉丝', value: overview.value.totalFollowers, increase: overview.value.followersIncrease, bgColor: '#f0f5ff', icon: User },
  { key: 'coins', label: '硬币', value: overview.value.totalCoins, increase: overview.value.coinsIncrease, bgColor: '#fff0e6', icon: Coin },
  { key: 'collections', label: '收藏', value: overview.value.totalCollections, increase: overview.value.collectionsIncrease, bgColor: '#f0fcf0', icon: FolderAdd },
  { key: 'shares', label: '分享', value: overview.value.totalShares, increase: overview.value.sharesIncrease, bgColor: '#e8faf8', icon: Share }
])

const visibleTrendData = computed(() => {
  const dates = trendData.value.dates || []
  const views = trendData.value.views || []
  const likes = trendData.value.likes || []
  const comments = trendData.value.comments || []
  const followers = trendData.value.followers || []

  const filtered = dates.map((date, index) => ({
    date,
    views: views[index] || 0,
    likes: likes[index] || 0,
    comments: comments[index] || 0,
    followers: followers[index] || 0
  })).filter(item => item.views > 0 || item.likes > 0 || item.comments > 0 || item.followers > 0)

  return {
    dates: filtered.map(item => item.date),
    views: filtered.map(item => item.views),
    likes: filtered.map(item => item.likes),
    comments: filtered.map(item => item.comments),
    followers: filtered.map(item => item.followers)
  }
})

const trendSeries = computed(() => [{
  name: trendMetric.value === 'views' ? '播放量'
    : trendMetric.value === 'likes' ? '点赞'
    : trendMetric.value === 'comments' ? '评论'
    : '粉丝',
  type: 'line',
  color: trendMetric.value === 'views' ? '#00aeec'
    : trendMetric.value === 'likes' ? '#ff6b81'
    : trendMetric.value === 'comments' ? '#f3a832'
    : '#52c41a',
  data: visibleTrendData.value[trendMetric.value] || []
}])

const rankingXData = computed(() => (rankingList.value || []).map(item => {
  const title = item.title || ''
  return title.length > 8 ? title.slice(0, 8) + '...' : title
}))

const rankingSeries = computed(() => [{
  name: '播放量',
  type: 'bar',
  color: '#00aeec',
  showArea: false,
  data: (rankingList.value || []).map(item => item.viewCount || 0)
}])

function switchTrendMetric() {
  loadTrend(trendDays.value, true)
}

function switchTrendDays() {
  loadTrend(trendDays.value, true)
}

onMounted(async () => {
  await Promise.all([
    loadOverview(),
    loadTrend(trendDays.value),
    loadRanking('views', 5),
    loadFansTrend(30)
  ])
})
</script>

<style scoped>
.overview-panel {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.panel-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}
.chart-section {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.chart-header h4 {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin: 0;
}
.chart-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
</style>
