<template>
  <div class="overview-panel">
    <h3 class="panel-title">数据概览</h3>

    <DataCards :cards="statCards" />

    <div class="chart-section">
      <div class="chart-header">
        <h4>数据趋势</h4>
        <div class="chart-actions">
          <el-radio-group v-model="trendDays" size="small" @change="switchTrendDays" v-if="trendMetric !== 'views'">
            <el-radio-button :value="7">7天</el-radio-button>
            <el-radio-button :value="30">30天</el-radio-button>
            <el-radio-button :value="90">90天</el-radio-button>
          </el-radio-group>
          <el-radio-group v-model="trendMetric" size="small" @change="switchTrendMetric">
            <el-radio-button value="views">播放量</el-radio-button>
            <el-radio-button value="likes">点赞</el-radio-button>
            <el-radio-button value="comments">评论</el-radio-button>
            <el-radio-button value="danmaku">弹幕</el-radio-button>
            <el-radio-button value="followers">粉丝</el-radio-button>
            <el-radio-button value="coins">硬币</el-radio-button>
            <el-radio-button value="collects">收藏</el-radio-button>
            <el-radio-button value="shares">分享</el-radio-button>
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
  manuscriptTrend,
  loadOverview,
  loadTrend,
  loadRanking,
  loadFansTrend,
  loadManuscriptTrend
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

const isManuscriptTrendMode = computed(() => trendMetric.value === 'views')

const visibleTrendData = computed(() => {
  if (isManuscriptTrendMode.value) {
    return {
      dates: manuscriptTrend.value.dates || [],
      views: manuscriptTrend.value.views || [],
      danmaku: manuscriptTrend.value.danmaku || []
    }
  }

  const dates = trendData.value.dates || []
  const likes = trendData.value.likes || []
  const comments = trendData.value.comments || []
  const followers = trendData.value.followers || []
  const danmaku = trendData.value.danmaku || []
  const coins = trendData.value.coins || []
  const collects = trendData.value.collects || []
  const shares = trendData.value.shares || []

  const filtered = dates.map((date, index) => ({
    date,
    likes: likes[index] || 0,
    comments: comments[index] || 0,
    followers: followers[index] || 0,
    danmaku: danmaku[index] || 0,
    coins: coins[index] || 0,
    collects: collects[index] || 0,
    shares: shares[index] || 0
  })).filter(item => item.likes > 0 || item.comments > 0 || item.followers > 0 || item.danmaku > 0 || item.coins > 0 || item.collects > 0 || item.shares > 0)

  return {
    dates: filtered.map(item => item.date),
    likes: filtered.map(item => item.likes),
    comments: filtered.map(item => item.comments),
    followers: filtered.map(item => item.followers),
    danmaku: filtered.map(item => item.danmaku),
    coins: filtered.map(item => item.coins),
    collects: filtered.map(item => item.collects),
    shares: filtered.map(item => item.shares)
  }
})

const trendSeries = computed(() => {
  if (isManuscriptTrendMode.value) {
    return [{
      name: '播放量',
      type: 'line',
      color: '#00aeec',
      data: visibleTrendData.value.views || []
    }, {
      name: '弹幕',
      type: 'line',
      color: '#52c41a',
      data: visibleTrendData.value.danmaku || []
    }]
  }

  const metricConfig = {
    likes: { name: '点赞', color: '#ff6b81' },
    comments: { name: '评论', color: '#f3a832' },
    followers: { name: '粉丝', color: '#52c41a' },
    danmaku: { name: '弹幕', color: '#00aeec' },
    coins: { name: '硬币', color: '#ff6b81' },
    collects: { name: '收藏', color: '#a855f7' },
    shares: { name: '分享', color: '#06b6d4' }
  }
  const cfg = metricConfig[trendMetric.value] || { name: '播放量', color: '#00aeec' }
  return [{
    name: cfg.name,
    type: 'line',
    color: cfg.color,
    data: visibleTrendData.value[trendMetric.value] || []
  }]
})

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
  if (isManuscriptTrendMode.value) {
    loadManuscriptTrend()
  } else {
    loadTrend(trendDays.value, true)
  }
}

function switchTrendDays() {
  loadTrend(trendDays.value, true)
}

onMounted(async () => {
  await Promise.all([
    loadOverview(),
    loadManuscriptTrend(),
    loadTrend(7),
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
