<template>
  <div class="overview-panel">
    <!-- 核心数据概览 -->
    <div class="overview-section">
      <div class="section-header">
        <div class="header-left">
          <h4>核心数据概览</h4>
          <el-tooltip content="统计所选时间范围内的核心数据，增量为所选天数内的新增量" placement="top">
            <el-icon class="help-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
          <span class="update-time">更新至{{ updateDateText }}</span>
        </div>
        <div class="header-right">
          <span class="range-label">时间选择</span>
          <el-select v-model="rangeDays" class="range-select" @change="switchRange">
            <el-option label="近7天" :value="7"></el-option>
            <el-option label="近30天" :value="30"></el-option>
          </el-select>
          <el-button link class="export-btn" @click="exportCsv">
            <el-icon><Download /></el-icon>导出数据
          </el-button>
        </div>
      </div>

      <div class="metric-grid">
        <div
          v-for="m in metricCards"
          :key="m.key"
          class="metric-card"
          :class="{ selected: selectedMetric === m.trendKey }"
          @click="selectMetric(m.trendKey)"
        >
          <div class="metric-top">
            <span class="metric-label">{{ m.label }}</span>
            <span v-if="m.increase !== null && m.increase !== undefined" class="metric-delta">▼{{ m.increase }}</span>
          </div>
          <div class="metric-value">{{ formatCount(m.value) }}</div>
        </div>
      </div>

      <div class="trend-block">
        <div class="trend-title">近{{ rangeDays }}天{{ selectedLabel }}</div>
        <TrendChart
          :xData="trendXData"
          :series="trendSeries"
          :height="280"
          :loading="loading.trend"
        />
      </div>
    </div>

    <!-- 近期稿件对比 -->
    <div class="overview-section">
      <div class="section-header">
        <div class="header-left">
          <h4>近期稿件对比</h4>
          <el-tooltip content="按播放量取TOP10稿件，对比各项历史累积数据" placement="top">
            <el-icon class="help-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </div>
      </div>
      <div class="compare-sub">历史累积数据</div>
      <el-table
        :data="compareRows"
        v-loading="compareLoading"
        style="width: 100%"
        stripe
      >
        <el-table-column label="稿件" min-width="250">
          <template #default="scope">
            <div class="compare-manuscript">
              <div class="compare-cover">
                <img v-if="scope.row.coverUrl" :src="scope.row.coverUrl" :alt="scope.row.title" />
                <span v-else class="cover-fallback">暂无封面</span>
                <span v-if="scope.row.duration || scope.row.durationSeconds" class="cover-duration">
                  {{ formatDuration(scope.row.duration, scope.row.durationSeconds) }}
                </span>
              </div>
              <div class="compare-info">
                <div class="compare-title">{{ scope.row.title }}</div>
                <div class="compare-date">{{ formatDate(scope.row.uploadTime || scope.row.createdAt) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="viewCount" label="播放量" width="90" sortable></el-table-column>
        <el-table-column prop="likeCount" label="点赞量" width="90" sortable></el-table-column>
        <el-table-column prop="commentCount" label="评论量" width="90" sortable></el-table-column>
        <el-table-column prop="danmakuCount" label="弹幕量" width="90" sortable></el-table-column>
        <el-table-column prop="collectCount" label="收藏量" width="90" sortable></el-table-column>
        <el-table-column prop="coinCount" label="投币量" width="90" sortable></el-table-column>
        <el-table-column prop="shareCount" label="转发量" width="90" sortable></el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { QuestionFilled, Download } from '@element-plus/icons-vue'
import TrendChart from './TrendChart.vue'
import { useCreatorStats } from '@/composables/useCreatorStats'
import { manuscriptApi } from '@/api/creator'

const { loading, overview, trendData, loadOverview, loadTrend } = useCreatorStats()

// 时间范围（B站：近7天 / 近30天）
const rangeDays = ref(7)
// 当前选中的趋势指标（点击指标卡切换，B站同款首卡高亮）
const selectedMetric = ref('views')

const metricCards = computed(() => [
  { key: 'views', trendKey: 'views', label: '播放量', value: overview.value.totalViews, increase: overview.value.viewsIncrease },
  { key: 'followers', trendKey: 'followers', label: '粉丝总数', value: overview.value.totalFollowers, increase: overview.value.followersIncrease },
  { key: 'followersNew', trendKey: 'followers', label: '净增粉丝', value: overview.value.followersIncrease, increase: null },
  { key: 'likes', trendKey: 'likes', label: '点赞', value: overview.value.totalLikes, increase: overview.value.likesIncrease },
  { key: 'collects', trendKey: 'collects', label: '收藏', value: overview.value.totalCollections, increase: overview.value.collectionsIncrease },
  { key: 'coins', trendKey: 'coins', label: '硬币', value: overview.value.totalCoins, increase: overview.value.coinsIncrease },
  { key: 'comments', trendKey: 'comments', label: '评论', value: overview.value.totalComments, increase: overview.value.commentsIncrease },
  { key: 'danmaku', trendKey: 'danmaku', label: '弹幕', value: overview.value.totalDanmaku, increase: overview.value.danmakuIncrease },
  { key: 'shares', trendKey: 'shares', label: '分享', value: overview.value.totalShares, increase: overview.value.sharesIncrease }
])

const metricLabelMap = {
  views: '播放量',
  followers: '播放量',
  likes: '点赞量',
  collects: '收藏量',
  coins: '投币量',
  comments: '评论量',
  danmaku: '弹幕量',
  shares: '转发量'
}

const selectedLabel = computed(() => metricLabelMap[selectedMetric.value] || '播放量')

const updateDateText = computed(() => {
  const t = overview.value.updateTime
  if (!t) return ''
  const parts = String(t).split('-')
  if (parts.length < 3) return t
  return `${Number(parts[1])}月${Number(parts[2].slice(0, 2))}日`
})

const trendXData = computed(() => {
  const dates = trendData.value.dates || []
  return dates.map((d) => {
    const parts = String(d).split('-')
    if (parts.length < 3) return d
    return `${parts[1]}-${parts[2]}`
  })
})

const trendSeries = computed(() => [{
  name: `总${selectedLabel.value}`,
  type: 'line',
  color: '#ff6699',
  data: trendData.value[selectedMetric.value] || []
}])

const selectMetric = (key) => {
  selectedMetric.value = key
}

const switchRange = async () => {
  await Promise.all([
    loadOverview(rangeDays.value),
    loadTrend(rangeDays.value)
  ])
}

// 近期稿件对比（按播放量TOP10）
const compareRows = ref([])
const compareLoading = ref(false)

const loadCompare = async () => {
  compareLoading.value = true
  try {
    const response = await manuscriptApi.getMyManuscripts({ page: 1, size: 50, status: 'published' })
    if (response.code === 200) {
      const list = response.data.list || []
      compareRows.value = [...list]
        .sort((a, b) => (b.viewCount || 0) - (a.viewCount || 0))
        .slice(0, 10)
    }
  } catch (error) {
    console.error('获取稿件对比数据失败:', error)
  } finally {
    compareLoading.value = false
  }
}

const exportCsv = () => {
  try {
    const rows = [['指标', '累计值', `近${rangeDays.value}天增量`]]
    metricCards.value.forEach((m) => {
      rows.push([m.label, String(m.value ?? 0), m.increase === null || m.increase === undefined ? '' : String(m.increase)])
    })
    const csv = '﻿' + rows.map((r) => r.join(',')).join('\n')
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `核心数据概览_近${rangeDays.value}天.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(a.href)
  } catch (error) {
    console.error('导出数据失败:', error)
    ElMessage.error('导出数据失败')
  }
}

const formatCount = (n) => {
  const num = Number(n) || 0
  if (num >= 100000000) return (num / 100000000).toFixed(1) + '亿'
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  return String(num)
}

const formatDuration = (duration, seconds) => {
  let total = 0
  if (typeof duration === 'string' && /^\d+:\d+:\d+$/.test(duration)) {
    const [h, m, s] = duration.split(':').map(Number)
    total = h * 3600 + m * 60 + s
  } else if (typeof duration === 'string' && /^\d+:\d+$/.test(duration)) {
    const [m, s] = duration.split(':').map(Number)
    total = m * 60 + s
  } else {
    total = Number(seconds) || Number(duration) || 0
  }
  if (total <= 0) return ''
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) {
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}年${month}月${day}日`
}

onMounted(async () => {
  await Promise.all([
    loadOverview(rangeDays.value),
    loadTrend(rangeDays.value),
    loadCompare()
  ])
})
</script>

<style scoped>
.overview-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.overview-section {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-left h4 {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
  margin: 0;
}

.help-icon {
  font-size: 14px;
  color: #c9ccd0;
  cursor: pointer;
}

.update-time {
  font-size: 12px;
  color: #99a2aa;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.range-label {
  font-size: 12px;
  color: #99a2aa;
}

.range-select {
  width: 110px;
}

.export-btn {
  font-size: 13px;
}

/* 指标卡：5列网格，选中粉色高亮（B站同款） */
.metric-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}

.metric-card {
  background: #f7f8fa;
  border-radius: 6px;
  padding: 12px 14px;
  cursor: pointer;
  transition: all .2s;
  min-width: 0;
}

.metric-card:hover {
  background: #f0f1f4;
}

.metric-card.selected {
  background: linear-gradient(135deg, #ff7aa8, #ff6699);
}

.metric-card.selected .metric-label,
.metric-card.selected .metric-value,
.metric-card.selected .metric-delta {
  color: #fff;
}

.metric-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-bottom: 8px;
}

.metric-label {
  font-size: 12px;
  color: #61666d;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.metric-delta {
  font-size: 12px;
  color: #52c41a;
  white-space: nowrap;
  flex-shrink: 0;
}

.metric-value {
  font-size: 20px;
  font-weight: 600;
  color: #18191c;
  line-height: 1.2;
}

.trend-block {
  border-top: 1px solid #f0f0f0;
  padding-top: 16px;
}

.trend-title {
  font-size: 14px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 12px;
}

/* 稿件对比表 */
.compare-sub {
  font-size: 12px;
  color: #99a2aa;
  margin-bottom: 12px;
}

.compare-manuscript {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 4px 0;
}

.compare-cover {
  position: relative;
  width: 80px;
  height: 50px;
  flex-shrink: 0;
  border-radius: 4px;
  overflow: hidden;
  background: #e5e9ef;
}

.compare-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  font-size: 10px;
  color: #99a2aa;
}

.cover-duration {
  position: absolute;
  right: 4px;
  bottom: 4px;
  padding: 1px 5px;
  font-size: 11px;
  color: #fff;
  background: rgba(0, 0, 0, .6);
  border-radius: 3px;
}

.compare-info {
  flex: 1;
  min-width: 0;
}

.compare-title {
  font-size: 13px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 6px;
}

.compare-date {
  font-size: 12px;
  color: #99a2aa;
}
</style>
