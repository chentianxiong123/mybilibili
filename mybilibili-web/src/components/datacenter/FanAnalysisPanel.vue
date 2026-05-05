<template>
  <div class="fan-panel">
    <h3 class="panel-title">粉丝分析</h3>

    <div class="fan-overview-cards">
      <div class="fan-card">
        <div class="fan-card-label">当前粉丝</div>
        <div class="fan-card-value">{{ formatNum(fansTrend.currentFollowers) }}</div>
      </div>
      <div class="fan-card">
        <div class="fan-card-label">今日新增</div>
        <div class="fan-card-value up">+{{ fansTrend.newFollowersToday }}</div>
      </div>
      <div class="fan-card">
        <div class="fan-card-label">今日取关</div>
        <div class="fan-card-value down">-{{ fansTrend.unfollowsToday }}</div>
      </div>
      <div class="fan-card">
        <div class="fan-card-label">增长率</div>
        <div class="fan-card-value" :class="(fansTrend.growthRate || 0) >= 0 ? 'up' : 'down'">
          {{ (fansTrend.growthRate || 0).toFixed(2) }}%
        </div>
      </div>
    </div>

    <div class="chart-section">
      <div class="chart-header">
        <h4>粉丝增长趋势</h4>
        <el-radio-group v-model="fanTrendDays" size="small" @change="switchFanDays">
          <el-radio-button :value="7">7条</el-radio-button>
          <el-radio-button :value="30">30条</el-radio-button>
        </el-radio-group>
      </div>
      <TrendChart
        :xData="fansTrend.dates"
        :series="fanSeries"
        :height="320"
        :loading="loading.fansTrend"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import TrendChart from './TrendChart.vue'
import { useCreatorStats } from '@/composables/useCreatorStats'

const {
  loading,
  fansTrend,
  loadFansTrend
} = useCreatorStats()

const fanTrendDays = ref(30)

const fanSeries = computed(() => [
  {
    name: '新增粉丝',
    type: 'line',
    color: '#52c41a',
    data: fansTrend.value.newFollowers || []
  },
  {
    name: '取关',
    type: 'line',
    color: '#ff6b81',
    data: fansTrend.value.unfollows || []
  },
  {
    name: '累计粉丝',
    type: 'line',
    color: '#00aeec',
    data: fansTrend.value.totalFollowers || []
  }
])

function switchFanDays() {
  loadFansTrend(fanTrendDays.value, true)
}

function formatNum(num) {
  if (num == null) return '0'
  if (num >= 10000) return (num / 10000).toFixed(1) + 'w'
  return String(num)
}

onMounted(() => {
  loadFansTrend(30)
})
</script>

<style scoped>
.fan-panel {
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
.fan-overview-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}
.fan-card {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 16px;
  text-align: center;
}
.fan-card-label {
  font-size: 13px;
  color: #999;
  margin-bottom: 6px;
}
.fan-card-value {
  font-size: 22px;
  font-weight: 700;
  color: #333;
}
.fan-card-value.up {
  color: #52c41a;
}
.fan-card-value.down {
  color: #ff4d4f;
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
</style>
