<template>
  <div class="trend-chart-wrapper" ref="wrapperRef">
    <div v-if="loading" class="chart-loading">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else-if="isEmpty" class="chart-empty">
      <svg viewBox="0 0 120 100" width="80" height="66">
        <rect x="10" y="60" width="15" height="30" rx="2" fill="#e8e8e8" />
        <rect x="30" y="40" width="15" height="50" rx="2" fill="#e8e8e8" />
        <rect x="50" y="20" width="15" height="70" rx="2" fill="#e8e8e8" />
        <rect x="70" y="35" width="15" height="55" rx="2" fill="#e8e8e8" />
        <rect x="90" y="10" width="15" height="80" rx="2" fill="#e8e8e8" />
      </svg>
      <span class="empty-text">暂无数据</span>
    </div>
    <div v-else ref="chartRef" class="trend-chart" :style="{ height: height + 'px' }"></div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  xData: { type: Array, default: () => [] },
  series: { type: Array, default: () => [] },
  height: { type: Number, default: 350 },
  loading: { type: Boolean, default: false },
  darkMode: { type: Boolean, default: false },
  showLegend: { type: Boolean, default: true },
  horizontal: { type: Boolean, default: false },
  tooltipFormatter: { type: Function, default: null }
})

const wrapperRef = ref(null)
const chartRef = ref(null)
let chartInstance = null
let resizeObserver = null

const isEmpty = ref(false)

const defaultColorPalette = ['#00aeec', '#f3a832', '#52c41a', '#ff6b81', '#a855f7', '#06b6d4']

const getTextColor = () => props.darkMode ? '#ccc' : '#666'
const getAxisLineColor = () => props.darkMode ? '#444' : '#e8e8e8'
const getSplitLineColor = () => props.darkMode ? '#333' : '#f0f0f0'

function initChart() {
  if (!chartRef.value) return

  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }

  const validSeries = (props.series || []).filter(s => s && s.data && s.data.length > 0)
  if (validSeries.length === 0 || !props.xData || props.xData.length === 0) {
    isEmpty.value = true
    return
  }
  isEmpty.value = false

  chartInstance = echarts.init(chartRef.value)

  const option = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255,255,255,0.95)',
      borderColor: '#e8e8e8',
      textStyle: { color: '#333', fontSize: 12 },
      formatter: props.tooltipFormatter || undefined
    },
    legend: props.showLegend ? {
      bottom: 0,
      textStyle: { color: getTextColor(), fontSize: 12 },
      itemWidth: 16,
      itemHeight: 8,
      itemGap: 20
    } : undefined,
    grid: {
      left: '3%',
      right: '8%',
      top: '10%',
      bottom: props.showLegend ? '12%' : '3%',
      containLabel: true
    },
    xAxis: props.horizontal ? {
      type: 'value',
      splitLine: { lineStyle: { color: getSplitLineColor(), type: 'dashed' } },
      axisLabel: { color: getTextColor(), fontSize: 11, formatter: formatAxisValue }
    } : {
      type: 'category',
      boundaryGap: false,
      data: props.xData,
      axisLine: { lineStyle: { color: getAxisLineColor() } },
      axisTick: { show: false },
      axisLabel: { color: getTextColor(), fontSize: 11, formatter: formatCategoryLabel }
    },
    yAxis: props.horizontal ? {
      type: 'category',
      data: props.xData,
      axisLine: { lineStyle: { color: getAxisLineColor() } },
      axisTick: { show: false },
      axisLabel: { color: getTextColor(), fontSize: 11 }
    } : {
      type: 'value',
      splitLine: { lineStyle: { color: getSplitLineColor(), type: 'dashed' } },
      axisLabel: { color: getTextColor(), fontSize: 11, formatter: formatAxisValue }
    },
    series: validSeries.map((s, i) => ({
      name: s.name || '',
      type: s.type || 'line',
      smooth: !props.horizontal && (s.type || 'line') === 'line',
      symbol: !props.horizontal ? 'circle' : 'none',
      symbolSize: 6,
      lineStyle: {
        width: 2.5,
        color: s.color || defaultColorPalette[i % defaultColorPalette.length]
      },
      areaStyle: !props.horizontal && s.showArea !== false ? {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: hexToRgba(s.color || defaultColorPalette[i % defaultColorPalette.length], 0.2) },
          { offset: 1, color: hexToRgba(s.color || defaultColorPalette[i % defaultColorPalette.length], 0.02) }
        ])
      } : undefined,
      barWidth: props.horizontal ? '50%' : undefined,
      itemStyle: {
        color: s.color || defaultColorPalette[i % defaultColorPalette.length],
        borderRadius: props.horizontal ? [0, 4, 4, 0] : 0
      },
      data: s.data
    }))
  }

  chartInstance.setOption(option)
}

function formatCategoryLabel(value) {
  if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value)) {
    return value.slice(5)
  }
  return value
}

function formatAxisValue(value) {
  if (value >= 10000) return (value / 10000).toFixed(1) + 'w'
  if (value >= 1000) return (value / 1000).toFixed(1) + 'k'
  return value
}

function hexToRgba(hex, alpha) {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return `rgba(${r},${g},${b},${alpha})`
}

function handleResize() {
  chartInstance?.resize()
}

watch(() => [props.xData, props.series, props.darkMode], () => {
  nextTick(() => initChart())
}, { deep: true })

onMounted(() => {
  nextTick(() => {
    initChart()
    if (wrapperRef.value) {
      resizeObserver = new ResizeObserver(handleResize)
      resizeObserver.observe(wrapperRef.value)
    }
  })
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  chartInstance?.dispose()
  chartInstance = null
})
</script>

<style scoped>
.trend-chart-wrapper {
  position: relative;
  width: 100%;
  min-height: 200px;
}
.trend-chart {
  width: 100%;
}
.chart-loading {
  padding: 40px;
}
.chart-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 12px;
}
.empty-text {
  color: #999;
  font-size: 14px;
}
</style>
