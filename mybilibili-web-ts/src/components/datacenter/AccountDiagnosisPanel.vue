<template>
  <div class="account-diagnosis">
    <div class="diagnosis-summary">
      <h3 class="section-title">表现总结</h3>
      <div class="summary-tip">
        <span>以下基于您近期的表现进行分析，关注提升空间。</span>
      </div>
      <div class="summary-content" v-if="overview.totalManuscripts > 0">
        <p v-for="(item, i) in summaryList" :key="i">• {{ item }}</p>
      </div>
      <div class="summary-content" v-else>
        <p>• 暂无已发布稿件，发布视频后即可查看诊断数据</p>
      </div>
    </div>

    <div class="diagnosis-ranking">
      <div class="ranking-box">
        <h4 class="ranking-title">总播放量</h4>
        <div class="ranking-value">{{ formatNumber(overview.totalViews) }}</div>
        <div class="ranking-change">近7日 {{ formatIncrease(overview.viewsIncrease) }}</div>
      </div>
      <div class="ranking-box">
        <h4 class="ranking-title">总互动数</h4>
        <div class="ranking-value">{{ formatNumber(totalInteractions) }}</div>
        <div class="ranking-change">近7日 {{ formatIncrease(interactionsIncrease) }}</div>
      </div>
    </div>

    <div class="diagnosis-metrics">
      <h3 class="section-title">核心指标</h3>
      <div class="metrics-grid">
        <div class="metric-item">
          <span class="metric-label">内容质量分</span>
          <span class="metric-value">{{ contentQualityScore }}%</span>
          <span class="metric-hint">(点赞+投币×2+收藏×2) ÷ 播放量</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">互动率</span>
          <span class="metric-value">{{ interactionRate }}%</span>
          <span class="metric-hint">点赞/评论/投币等 ÷ 播放量</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">粉丝转化率</span>
          <span class="metric-value">{{ fanConversionRate }}%</span>
          <span class="metric-hint">近7日新增粉丝 ÷ 总播放</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">活跃度</span>
          <span class="metric-value">{{ activityScore }}</span>
          <span class="metric-hint">近7日互动增长 ÷ 稿件数</span>
        </div>
      </div>
    </div>

    <div class="diagnosis-suggestions">
      <h3 class="section-title">优化建议</h3>
      <div class="suggestions-list">
        <div class="suggestion-item" v-for="(item, i) in suggestions" :key="i">
          <span class="suggestion-icon">{{ item.icon }}</span>
          <span class="suggestion-text">{{ item.text }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useCreatorStats } from '@/composables/useCreatorStats'

const { overview, loadOverview } = useCreatorStats()
loadOverview()

const totalInteractions = computed(() =>
  overview.value.totalLikes + overview.value.totalComments +
  overview.value.totalDanmaku + overview.value.totalCoins +
  overview.value.totalCollections + overview.value.totalShares
)

const interactionsIncrease = computed(() =>
  overview.value.likesIncrease + overview.value.commentsIncrease +
  overview.value.danmakuIncrease + overview.value.coinsIncrease +
  overview.value.collectionsIncrease + overview.value.sharesIncrease
)

const interactionRate = computed(() => {
  if (!overview.value.totalViews) return '0.00'
  return (totalInteractions.value / overview.value.totalViews * 100).toFixed(2)
})

const fanConversionRate = computed(() => {
  if (!overview.value.totalViews) return '0.00'
  return (overview.value.followersIncrease / overview.value.totalViews * 100).toFixed(2)
})

// 内容质量分：(点赞 + 投币×2 + 收藏×2) ÷ 播放量 × 100
const contentQualityScore = computed(() => {
  const o = overview.value
  if (!o.totalViews) return '0.00'
  const score = (o.totalLikes + o.totalCoins * 2 + o.totalCollections * 2) / o.totalViews * 100
  return Math.min(score, 100).toFixed(2)
})

// 活跃度：近7日互动增长数 ÷ 稿件数
const activityScore = computed(() => {
  const o = overview.value
  if (!o.totalManuscripts) return '0'
  const increase = o.viewsIncrease + o.likesIncrease + o.commentsIncrease +
    o.danmakuIncrease + o.coinsIncrease + o.collectionsIncrease + o.sharesIncrease + o.followersIncrease
  return Math.round(increase / o.totalManuscripts)
})

const summaryList = computed(() => {
  const list = []
  const o = overview.value

  if (o.viewsIncrease > 0) {
    list.push('播放量处于上升趋势，内容曝光良好')
  } else if (o.totalViews > 0) {
    list.push('播放量保持稳定，建议优化标题和封面提升曝光')
  }

  if (o.likesIncrease > 0) {
    list.push('点赞量稳定提升，内容质量获得认可')
  }

  if (o.collectionsIncrease > 0) {
    list.push('收藏数增长较快，内容具有长期价值')
  }

  if (o.totalManuscripts > 0 && o.totalViews / o.totalManuscripts > 100) {
    list.push('单稿平均播放量可观，保持现有产出节奏')
  }

  if (list.length === 0) {
    list.push('数据积累中，持续发布优质内容即可看到诊断分析')
  }
  return list
})

const suggestions = computed(() => {
  const list = []
  const o = overview.value

  if (o.totalManuscripts < 5) {
    list.push({ icon: '💡', text: '建议增加视频发布频率，积累更多内容数据' })
  }

  const quality = parseFloat(contentQualityScore.value)
  if (quality > 5) {
    list.push({ icon: '🌟', text: '内容质量分优秀，用户对你的内容认可度很高' })
  } else if (quality < 1 && o.totalViews > 0) {
    list.push({ icon: '📈', text: '内容质量分偏低，建议优化内容质量引导投币和收藏' })
  }

  const rate = totalInteractions.value / (o.totalViews || 1)
  if (rate < 0.03) {
    list.push({ icon: '📈', text: '互动率偏低，可在视频中增加引导点赞/投币的话术' })
  }

  if (o.followersIncrease <= 0 && o.totalViews > 0) {
    list.push({ icon: '🎯', text: '近期粉丝增长放缓，尝试在热门时段发布视频' })
  }

  if (o.totalComments < o.totalLikes * 0.1 && o.totalLikes > 0) {
    list.push({ icon: '💬', text: '评论相对较少，可在视频结尾设置互动话题引导评论' })
  }

  if (list.length === 0) {
    list.push({ icon: '✨', text: '数据表现不错，继续保持当前的创作节奏' })
  }
  return list
})

function formatNumber(n) {
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n ?? 0)
}

function formatIncrease(n) {
  if (!n) return '无变化'
  return n > 0 ? `+${n}` : String(n)
}
</script>

<style scoped>
.account-diagnosis {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0 0 12px 0;
}
.diagnosis-summary {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.summary-tip {
  background: #fff7e6;
  border-radius: 6px;
  padding: 10px 14px;
  margin-bottom: 14px;
  color: #d48806;
  font-size: 13px;
}
.summary-content {
  color: #666;
  font-size: 14px;
  line-height: 1.8;
}
.diagnosis-ranking {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.ranking-box {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
  text-align: center;
}
.ranking-title {
  font-size: 14px;
  color: #999;
  margin: 0 0 10px 0;
}
.ranking-value {
  font-size: 28px;
  font-weight: 700;
  color: #00aeec;
}
.ranking-change {
  font-size: 12px;
  color: #999;
  margin-top: 6px;
}
.diagnosis-metrics {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.metric-item {
  text-align: center;
  padding: 16px;
  background: #f9f9f9;
  border-radius: 8px;
}
.metric-label {
  display: block;
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}
.metric-value {
  display: block;
  font-size: 20px;
  font-weight: 600;
  color: #333;
}
.metric-hint {
  display: block;
  font-size: 11px;
  color: #bbb;
  margin-top: 4px;
}
.diagnosis-suggestions {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.suggestions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.suggestion-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #f9f9f9;
  border-radius: 8px;
}
.suggestion-icon {
  font-size: 18px;
}
.suggestion-text {
  font-size: 14px;
  color: #666;
}
</style>
