<script setup>
import { ref, onMounted } from 'vue'
import { getAiUsageOverview, getAiUsageFeatures, getAiUsageDaily } from '../api/aiUsage'
import echarts from '../utils/echartsCore'

const loading = ref(false)
const overview = ref({ totalCount: 0, totalTokens: 0, avgDuration: 0, successCount: 0 })
const features = ref([])
const daily = ref([])
const chartDays = ref(7)

let dailyChart = null
let featureChart = null

function formatDuration(ms) {
  if (!ms) return '0ms'
  if (ms < 1000) return ms.toFixed(0) + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}

const featureLabels = { CHAT: 'AI客服', REVIEW: '内容审核', SUMMARY: '视频摘要', TEST: '连接测试' }

async function loadData() {
  loading.value = true
  try {
    const [ovRes, featRes, dailyRes] = await Promise.all([
      getAiUsageOverview(),
      getAiUsageFeatures(),
      getAiUsageDaily(chartDays.value)
    ])
    if (ovRes.code === 200) overview.value = ovRes.data
    if (featRes.code === 200) features.value = featRes.data
    if (dailyRes.code === 200) daily.value = dailyRes.data
    renderCharts()
  } catch (e) {
    console.error('加载AI用量数据失败', e)
  } finally {
    loading.value = false
  }
}

function renderCharts() {
  // 每日调用趋势
  const dailyEl = document.getElementById('dailyChart')
  if (dailyEl) {
    if (dailyChart) dailyChart.dispose()
    dailyChart = echarts.init(dailyEl)
    dailyChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: daily.value.map(d => d.date) },
      yAxis: [
        { type: 'value', name: '调用次数' },
        { type: 'value', name: 'Token数' }
      ],
      series: [
        { name: '调用次数', type: 'bar', data: daily.value.map(d => d.count || 0) },
        { name: '成功数', type: 'line', data: daily.value.map(d => d.successCount || 0) },
        { name: 'Token数', type: 'line', yAxisIndex: 1,
          data: daily.value.map(d => d.totalTokens || 0) }
      ]
    })
  }

  // 功能分布
  const featEl = document.getElementById('featureChart')
  if (featEl) {
    if (featureChart) featureChart.dispose()
    featureChart = echarts.init(featEl)
    featureChart.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c}次 ({d}%)' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        data: features.value.map(f => ({
          name: featureLabels[f.feature] || f.feature,
          value: f.count
        }))
      }]
    })
  }
}

onMounted(loadData)
</script>

<template>
  <div class="ai-usage">
    <h2 class="page-title">AI 用量统计</h2>

    <!-- 概览卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon call">
          <el-icon :size="28"><Connection /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ overview.totalCount || 0 }}</div>
          <div class="stat-label">总调用次数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon token">
          <el-icon :size="28"><DataAnalysis /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ (overview.totalTokens || 0).toLocaleString() }}</div>
          <div class="stat-label">总 Token 消耗</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon duration">
          <el-icon :size="28"><Timer /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatDuration(overview.avgDuration) }}</div>
          <div class="stat-label">平均响应时间</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon rate">
          <el-icon :size="28"><SuccessFilled /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">
            {{ overview.totalCount ? ((overview.successCount / overview.totalCount) * 100).toFixed(1) : 0 }}%
          </div>
          <div class="stat-label">成功率</div>
        </div>
      </div>
    </div>

    <div class="charts-row">
      <div class="chart-panel">
        <div class="panel-header">
          <h3>每日调用趋势</h3>
          <el-select v-model="chartDays" size="small" @change="loadData" style="width: 120px">
            <el-option :value="7" label="近7天" />
            <el-option :value="14" label="近14天" />
            <el-option :value="30" label="近30天" />
          </el-select>
        </div>
        <div id="dailyChart" style="height: 300px"></div>
      </div>
      <div class="chart-panel">
        <div class="panel-header">
          <h3>功能分布</h3>
        </div>
        <div id="featureChart" style="height: 300px"></div>
      </div>
    </div>

    <!-- 功能明细 -->
    <div class="panel">
      <h3>功能明细</h3>
      <el-table :data="features" border stripe v-loading="loading" style="margin-top: 12px">
        <el-table-column label="功能" prop="feature" width="160">
          <template #default="{ row }">
            <el-tag>{{ featureLabels[row.feature] || row.feature }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="调用次数" prop="count" width="120" />
        <el-table-column label="成功数" prop="successCount" width="120" />
        <el-table-column label="Token 消耗" prop="totalTokens" width="160" />
        <el-table-column label="平均耗时">
          <template #default="{ row }">{{ formatDuration(row.avgDuration) }}</template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.ai-usage { padding: 20px; }
.page-title { margin: 0 0 24px; font-size: 24px; font-weight: 600; color: #333; }
.stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-bottom: 24px; }
.stat-card { display: flex; align-items: center; padding: 20px; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,.1); }
.stat-icon { display: flex; align-items: center; justify-content: center; width: 52px; height: 52px; border-radius: 8px; }
.stat-icon.call { background: rgba(64,158,255,.1); color: #409eff; }
.stat-icon.token { background: rgba(103,194,58,.1); color: #67c23a; }
.stat-icon.duration { background: rgba(230,162,60,.1); color: #e6a23c; }
.stat-icon.rate { background: rgba(245,108,108,.1); color: #f56c6c; }
.stat-content { margin-left: 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #333; }
.stat-label { font-size: 14px; color: #666; margin-top: 4px; }
.charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 24px; }
.panel { padding: 20px; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,.1); }
.panel-header { display: flex; justify-content: space-between; align-items: center; }
.panel-header h3, .panel h3 { margin: 0; font-size: 15px; font-weight: 600; color: #333; }
</style>
