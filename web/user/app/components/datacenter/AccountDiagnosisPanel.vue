<template>
  <div class="account-diagnosis">
    <!-- 表现总结 -->
    <div class="performance-summary">
      <div class="header">
        <div class="title">
          <span class="name">表现总结</span>
          <el-tooltip placement="top" :show-after="100">
            <template #content>
              <div>1.为你提供近30天同类UP主的数据表现情况，帮助你更好地自我定位并找到努力方向，继续加油!</div>
              <div>2. 每日中午12点(北京时间)更新昨日数据。</div>
            </template>
            <el-icon class="tip-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </div>
        <div class="sub-title">
          <span class="sub-text">{{ summaryText }}</span>
        </div>
      </div>

      <div class="card">
        <div class="card-container">
          <div class="radar-card">
            <div class="left" ref="radarRef"></div>
            <div class="right">
              <div class="right-item" v-for="item in metricItems" :key="item.name">
                <div class="item-title">{{ item.name }}：</div>
                <div class="item-detail">
                  <div class="detail-text">
                    <span>你的{{ item.name }}为</span>
                    <span class="text-num">{{ formatNum(item.mine) }}</span>
                    <span>{{ item.compareLabel }}</span>
                    <span class="view-tip">查看提升建议</span>
                    <span class="text-tips">{{ item.tip }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 播放分布 -->
    <div class="play-dist">
      <div class="section-title-container">
        <h3 class="section-title">播放分布</h3>
        <el-tooltip placement="top" :show-after="100">
          <template #content>
            <div>1. 帮助您了解近期流量来源渠道，为您的视频营销策略调整提供参考</div>
            <div>2. 分终端占比</div>
            <div>PC：网页及pc版客户端</div>
            <div>移动端：iOS、安卓及Pad</div>
            <div>h5及小程序：站外分享链接及小程序</div>
            <div>云视听小电视：TV版应用</div>
            <div>其他：上述未统计到的来源</div>
          </template>
          <el-icon class="tip-icon"><QuestionFilled /></el-icon>
        </el-tooltip>
      </div>
      <div class="text2 summary">
        <span class="b-icon">统计</span>
        观众中：<span class="highlight-text">{{ formatPercent(person.guestPercent) }}</span> 是游客，主要使用
        <span class="highlight-text">{{ mainTerminal }}</span> 设备观看。最新数据中，
        <span class="highlight-text">《{{ topManuscript }}》</span>播放量最高
      </div>

      <div class="card-contianer-wrap">
        <div class="card-contianer-wrap-left">
          <!-- 人群分布 -->
          <div class="card-container">
            <div class="text1 title">
              <span class="card-title">人群分布</span>
              <span class="title-tip">近30天</span>
            </div>
            <div class="person-content">
              <div class="donut-wrap" ref="personDonutRef"></div>
              <div class="person-bars">
                <div class="person-row">
                  <span class="row-label">游客</span>
                  <div class="progress-bar">
                    <div class="progress-bar-item" :style="{ width: pct(person.guestPercent), background: '#00a1d6' }"></div>
                  </div>
                  <span class="row-value">{{ formatPercent(person.guestPercent) }}</span>
                </div>
                <div class="person-row">
                  <span class="row-label">粉丝</span>
                  <div class="progress-bar">
                    <div class="progress-bar-item" :style="{ width: pct(person.fanPercent), background: '#c9ccd0' }"></div>
                  </div>
                  <span class="row-value">{{ formatPercent(person.fanPercent) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 观看途径 -->
          <div class="card-container">
            <div class="text1 title">
              <span class="card-title">观看途径</span>
              <span class="title-tip">近30天</span>
            </div>
            <div class="terminal-bars">
              <div class="progress" v-for="t in terminals" :key="t.label">
                <div class="bar-title">{{ t.label }} : {{ formatPercent(t.percent) }}</div>
                <div class="progress-bar">
                  <div
                    class="progress-bar-item"
                    :class="{ highlight: t.percent === maxTerminalPercent }"
                    :style="{ width: pct(t.percent) }"
                  ></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 来源稿件占比 -->
        <div class="card-container card-contianer-wrap-right">
          <div class="text1 title">
            <span class="card-title">来源稿件占比</span>
            <span class="title-tip">昨日</span>
            <span class="source-select">播放</span>
          </div>
          <div class="archive-list">
            <div class="progress" v-for="(s, i) in sources" :key="s.manuscriptId">
              <div class="title-row">
                <span class="archive-title">{{ s.title }}</span>
              </div>
              <div class="progress-bar">
                <div
                  class="progress-bar-item"
                  :class="{ highlight: i === 0 }"
                  :style="{ width: pct(s.percent) }"
                ></div>
              </div>
              <div class="tips-value">{{ formatPercent(s.percent) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { QuestionFilled } from '@element-plus/icons-vue'
import echarts from '@/utils/echartsCore'
import { statsApi } from '@/api/creator'

const radarRef = ref(null)
const personDonutRef = ref(null)

const diag = ref({
  metrics: {
    uploads: 0, views: 0, likes: 0, newFollowers: 0,
    peerUploads: 0, peerViews: 0, peerLikes: 0, peerNewFollowers: 0
  },
  personDistribution: { guestPercent: 1, fanPercent: 0 },
  terminals: [],
  sources: [],
  sourceDate: ''
})

const load = async () => {
  try {
    const res = await statsApi.getAccountDiagnosis()
    if (res?.code === 200 && res.data) {
      diag.value = res.data
      await nextTick()
      initRadar()
      initPersonDonut()
    }
  } catch (e) {
    console.error('账号诊断加载失败:', e)
  }
}

// —— 表现总结 ——
const person = computed(() => diag.value.personDistribution || { guestPercent: 1, fanPercent: 0 })
const terminals = computed(() => diag.value.terminals || [])
const sources = computed(() => diag.value.sources || [])

const metricItems = computed(() => {
  const m = diag.value.metrics || {}
  const rows = [
    { name: '投稿', mine: m.uploads, peer: m.peerUploads, tip: '持续发布作品，才会吸引更多粉丝哟~' },
    { name: '播放', mine: m.views, peer: m.peerViews, tip: '作品内容和选题是提升播放量的制胜法宝，封面和标题是视频的"门面"，继续加油~' },
    { name: '点赞', mine: m.likes, peer: m.peerLikes, tip: '把握视频节奏，及时展示精彩瞬间，添加互动弹幕，都可能让更多小伙伴为你点赞！' },
    { name: '新增关注', mine: m.newFollowers, peer: m.peerNewFollowers, tip: '提升作品的播放量吸引更多人观看、积极和粉丝互动， 都可以增加粉丝哦~' }
  ]
  return rows.map((r) => {
    const above = r.mine >= r.peer
    return {
      ...r,
      compareLabel: above ? '高于同类UP主' : '低于同类UP主',
      above
    }
  })
})

const summaryText = computed(() => {
  const low = metricItems.value.find((x) => !x.above)
  if (low) {
    return `你的${low.name}活跃低于同类UP主，坚持创作，能帮助你快速进步与成长哦~`
  }
  return '你的各项表现高于同类UP主，保持节奏，继续创作更优质的内容吧！'
})

// —— 播放分布 ——
const mainTerminal = computed(() => {
  const t = terminals.value.slice()
  t.sort((a, b) => b.percent - a.percent)
  return t.length ? t[0].label : 'PC'
})
const maxTerminalPercent = computed(() => {
  const t = terminals.value.slice()
  t.sort((a, b) => b.percent - a.percent)
  return t.length ? t[0].percent : 0
})
const topManuscript = computed(() => {
  const s = sources.value.slice()
  s.sort((a, b) => b.percent - a.percent)
  return s.length ? s[0].title : '暂无'
})

// —— 图表 ——
// —— 图表 ——
const initRadar = () => {
  if (!radarRef.value) return
  const m = diag.value.metrics || {}
  const rows = metricItems.value
  const mineArr = rows.map((r) => Number(r.mine) || 0)
  const peerArr = rows.map((r) => Number(r.peer) || 0)
  const maxArr = rows.map((_, i) => Math.max(mineArr[i], peerArr[i], 1))
  const mineData = mineArr.map((v, i) => (v / maxArr[i]) * 100)
  const peerData = peerArr.map((v, i) => (v / maxArr[i]) * 100)
  const chart = echarts.init(radarRef.value)
  chart.setOption({
    radar: {
      indicator: rows.map((r) => ({ name: r.name, max: 100 })),
      center: ['50%', '48%'],
      radius: '62%',
      splitNumber: 4,
      axisName: { color: '#61666d', fontSize: 12 },
      splitLine: { lineStyle: { color: '#e8e8e8' } },
      splitArea: { areaStyle: { color: ['#fff', '#fafafa'] } },
      axisLine: { lineStyle: { color: '#e8e8e8' } }
    },
    series: [{
      type: 'radar',
      data: [
        { name: '我的指标', value: mineData, lineStyle: { color: '#00a1d6', width: 2 }, itemStyle: { color: '#00a1d6' }, areaStyle: { color: 'rgba(0,161,214,0.15)' } },
        { name: '同类UP主', value: peerData, lineStyle: { color: 'rgba(4,4,21,0.4)', width: 2 }, itemStyle: { color: 'rgba(4,4,21,0.4)' }, areaStyle: { color: 'rgba(4,4,21,0.05)' } }
      ]
    }],
    legend: {
      bottom: 0,
      itemWidth: 8,
      itemHeight: 8,
      icon: 'circle',
      textStyle: { color: '#61666d', fontSize: 12 },
      data: ['我的指标', '同类UP主']
    }
  })
  return chart
}

const initPersonDonut = () => {
  if (!personDonutRef.value) return
  const p = person.value
  const chart = echarts.init(personDonutRef.value)
  chart.setOption({
    series: [{
      type: 'pie',
      radius: ['60%', '80%'],
      center: ['50%', '50%'],
      label: { show: false },
      data: [
        { value: p.guestPercent, name: '游客', itemStyle: { color: '#00a1d6' } },
        { value: p.fanPercent, name: '粉丝', itemStyle: { color: '#c9ccd0' } }
      ]
    }],
    legend: { show: false }
  })
  return chart
}

// —— 格式化 ——
const pct = (v) => `${Math.max(0, Math.min(1, Number(v) || 0)) * 100}%`
const formatPercent = (v) => {
  const n = (Number(v) || 0) * 100
  if (n === 0) return '0.0%'
  if (n >= 99.95) return '100%'
  return n.toFixed(1) + '%'
}
const formatNum = (n) => {
  const v = Number(n) || 0
  if (v >= 10000) return (v / 10000).toFixed(1) + '万'
  if (v >= 1000) return v.toLocaleString('en-US')
  return String(v)
}

onMounted(load)
</script>

<style scoped>
.account-diagnosis {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* —— 表现总结 —— */
.performance-summary {
  width: 100%;
}
.header {
  margin-bottom: 12px;
}
.title {
  display: flex;
  align-items: center;
  gap: 6px;
}
.title .name {
  font-size: 16px;
  font-weight: 400;
  color: #000;
}
.tip-icon {
  font-size: 14px;
  color: #c9ccd0;
  cursor: pointer;
}
.sub-title {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
}
.sub-text {
  font-size: 14px;
  color: #61666d;
}
.card {
  background: #fff;
  border-radius: 20px;
  padding: 38px 55px 40px;
  margin-top: 14px;
}
.radar-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.left {
  width: 460px;
  height: 340px;
  flex-shrink: 0;
}
.right {
  flex: 1;
  padding-left: 25px;
}
.right-item {
  margin-bottom: 22px;
}
.item-title {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 6px;
}
.item-detail {
  font-size: 14px;
  color: #61666d;
  line-height: 1.6;
}
.detail-text {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0;
}
.text-num {
  font-size: 16px;
  font-weight: 700;
  color: #18191c;
  margin: 0 4px;
}
.text-tips {
  color: #c9ccd0;
  font-size: 12px;
  margin-left: 6px;
}
.view-tip {
  color: #00a1d6;
  font-size: 13px;
  cursor: pointer;
  margin-left: 6px;
}

/* —— 播放分布 —— */
.play-dist .section-title-container {
  display: flex;
  align-items: center;
  gap: 8px;
}
.section-title {
  font-size: 18px;
  font-weight: 500;
  color: #18191c;
  margin: 0;
}
.text2.summary {
  font-size: 14px;
  color: #61666d;
  margin: 12px 0 16px;
  line-height: 1.6;
}
.highlight-text {
  color: #00a1d6;
  font-weight: 500;
}
.b-icon {
  margin-right: 4px;
}
.card-contianer-wrap {
  display: flex;
  gap: 17px;
}
.card-contianer-wrap-left {
  display: flex;
  flex-direction: column;
  gap: 16px;
  flex: 1;
  min-width: 0;
}
.card-contianer-wrap .card-container {
  background: #fff;
  border-radius: 20px;
  padding: 24px 32px;
}
.card-contianer-wrap-right {
  width: 32%;
  flex-shrink: 0;
}
.text1.title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
}
.title-tip {
  font-size: 13px;
  color: #949aa0;
}
.source-select {
  margin-left: auto;
  font-size: 13px;
  color: #61666d;
  cursor: pointer;
}

/* 人群分布 */
.person-content {
  display: flex;
  align-items: center;
  gap: 28px;
}
.donut-wrap {
  width: 130px;
  height: 130px;
  flex-shrink: 0;
}
.person-bars {
  flex: 1;
  min-width: 0;
}
.person-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.row-label {
  width: 40px;
  font-size: 14px;
  color: #18191c;
}
.row-value {
  width: 55px;
  text-align: right;
  font-size: 14px;
  color: #18191c;
}

/* 进度条 */
.progress-bar {
  flex: 1;
  height: 8px;
  background: #f0f1f4;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}
.progress-bar-item {
  height: 100%;
  border-radius: 4px;
  background: #c9ccd0;
  transition: width .3s;
}

/* 观看途径 */
.terminal-bars .progress {
  margin-bottom: 14px;
}
.bar-title {
  font-size: 14px;
  color: #61666d;
  margin-bottom: 8px;
}

/* 来源稿件 */
.archive-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-height: 460px;
  overflow-y: auto;
}
.archive-list .progress {
  position: relative;
}
.title-row {
  margin-bottom: 8px;
}
.archive-title {
  display: block;
  font-size: 14px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tips-value {
  margin-top: 6px;
  font-size: 13px;
  color: #949aa0;
  text-align: right;
}
.progress-bar-item.highlight {
  background: #00a1d6;
}
</style>