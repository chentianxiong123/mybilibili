<script setup>
import { ref, computed, watch } from 'vue'
import { profileApi } from '../../api/profile.js'
import { categoryApi } from '../../api/index.js'

const props = defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const loading = ref(true)
const profile = ref(null)
const categoryNames = ref({})

const numberValue = (value) => Number(value || 0)

const topCategories = computed(() => {
  if (!profile.value?.categoryWeights) return []
  return Object.entries(profile.value.categoryWeights)
    .map(([id, weight]) => ({
      id,
      name: categoryNames.value[String(id)] || `分区 ${id}`,
      weight: numberValue(weight)
    }))
    .sort((a, b) => b.weight - a.weight)
    .slice(0, 8)
})

const topTags = computed(() => {
  if (!profile.value?.tagWeights) return []
  return Object.entries(profile.value.tagWeights)
    .map(([name, weight]) => ({ name, weight: numberValue(weight) }))
    .sort((a, b) => b.weight - a.weight)
    .slice(0, 24)
})

const maxTagWeight = computed(() => topTags.value[0]?.weight || 1)
const maxCategoryWeight = computed(() => topCategories.value[0]?.weight || 1)

const activeHoursData = computed(() => {
  const source = profile.value?.activeHours || {}
  return Array.from({ length: 24 }, (_, hour) => ({
    hour,
    label: `${String(hour).padStart(2, '0')}:00`,
    count: numberValue(source[String(hour)])
  }))
})

const maxHourCount = computed(() => Math.max(...activeHoursData.value.map(item => item.count), 1))

const peakHours = computed(() => {
  return activeHoursData.value
    .filter(item => item.count > 0)
    .sort((a, b) => b.count - a.count)
    .slice(0, 3)
})

const durationLabel = computed(() => {
  const map = {
    short: '短视频',
    medium: '中等时长',
    long: '长视频'
  }
  return map[profile.value?.preferredDuration] || '暂无数据'
})

const durationRange = computed(() => {
  const map = {
    short: '< 5 分钟',
    medium: '5-20 分钟',
    long: '> 20 分钟'
  }
  return map[profile.value?.preferredDuration] || '--'
})

const totalActions = computed(() => {
  return numberValue(profile.value?.totalWatchCount)
    + numberValue(profile.value?.totalLikeCount)
    + numberValue(profile.value?.totalCollectCount)
})

const stats = computed(() => [
  { label: '观看', value: numberValue(profile.value?.totalWatchCount), unit: '次' },
  { label: '点赞', value: numberValue(profile.value?.totalLikeCount), unit: '次' },
  { label: '收藏', value: numberValue(profile.value?.totalCollectCount), unit: '次' },
  { label: '累计行为', value: totalActions.value, unit: '次' }
])

const hasProfileData = computed(() => {
  return topCategories.value.length > 0
    || topTags.value.length > 0
    || activeHoursData.value.some(item => item.count > 0)
    || totalActions.value > 0
})

const loadCategoryNames = async () => {
  try {
    const res = await categoryApi.getCategoryList()
    if (res.code === 200 && Array.isArray(res.data)) {
      categoryNames.value = res.data.reduce((map, item) => {
        map[String(item.id)] = item.name
        return map
      }, {})
    }
  } catch (error) {
    console.error('获取分类名称失败:', error)
  }
}

const fetchProfile = async () => {
  if (!props.userId) {
    loading.value = false
    profile.value = null
    return
  }

  loading.value = true
  try {
    const res = await profileApi.getProfile(props.userId)
    profile.value = res.code === 200 ? res.data : null
  } catch (error) {
    console.error('获取画像失败:', error)
    profile.value = null
  } finally {
    loading.value = false
  }
}

watch(
  () => props.userId,
  async () => {
    await Promise.all([loadCategoryNames(), fetchProfile()])
  },
  { immediate: true }
)
</script>

<template>
  <div class="interests-panel">
    <div class="panel-head">
      <div>
        <h2 class="panel-title">兴趣画像</h2>
        <p class="panel-subtitle">
          {{ isOwnSpace ? '基于你的观看、点赞和收藏行为生成' : '基于该用户公开视频互动行为生成' }}
        </p>
      </div>
      <div class="duration-card">
        <span class="duration-label">时长偏好</span>
        <strong>{{ durationLabel }}</strong>
        <span>{{ durationRange }}</span>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="8" animated />
    </div>

    <div v-else-if="!profile || !hasProfileData" class="empty-state">
      <el-empty description="暂无兴趣数据，多看几个视频后会逐步生成画像" />
    </div>

    <div v-else class="profile-content">
      <div class="stats-grid">
        <div v-for="item in stats" :key="item.label" class="stat-card">
          <div class="stat-value">{{ item.value }}</div>
          <div class="stat-label">{{ item.label }}{{ item.unit }}</div>
        </div>
      </div>

      <div class="insight-grid">
        <section class="section">
          <div class="section-head">
            <h3>分区偏好</h3>
            <span>Top {{ topCategories.length }}</span>
          </div>
          <div class="category-bars">
            <div v-for="item in topCategories" :key="item.id" class="bar-item">
              <div class="bar-meta">
                <span class="bar-label">{{ item.name }}</span>
                <span class="bar-value">{{ item.weight.toFixed(1) }}</span>
              </div>
              <div class="bar-track">
                <div class="bar-fill" :style="{ width: `${item.weight / maxCategoryWeight * 100}%` }"></div>
              </div>
            </div>
          </div>
        </section>

        <section class="section">
          <div class="section-head">
            <h3>观看时段</h3>
            <span v-if="peakHours.length">高峰 {{ peakHours.map(item => item.label).join(' / ') }}</span>
          </div>
          <div class="hours-chart">
            <div
              v-for="item in activeHoursData"
              :key="item.hour"
              class="hour-cell"
              :title="`${item.label} ${item.count}次`"
            >
              <div
                class="hour-fill"
                :style="{
                  height: `${Math.max(item.count / maxHourCount * 100, item.count ? 8 : 0)}%`,
                  opacity: 0.25 + item.count / maxHourCount * 0.75
                }"
              ></div>
              <span v-if="item.hour % 6 === 0">{{ item.hour }}</span>
            </div>
          </div>
        </section>
      </div>

      <section class="section tag-section">
        <div class="section-head">
          <h3>兴趣标签</h3>
          <span>按权重排序</span>
        </div>
        <div class="tag-cloud">
          <span
            v-for="item in topTags"
            :key="item.name"
            class="interest-tag"
            :style="{
              '--tag-ratio': item.weight / maxTagWeight,
              fontSize: `${12 + item.weight / maxTagWeight * 8}px`
            }"
          >
            {{ item.name }}
            <em>{{ item.weight.toFixed(1) }}</em>
          </span>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.interests-panel {
  padding: 24px;
  background: #fff;
  border-radius: 6px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 22px;
}

.panel-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #18191c;
}

.panel-subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  color: #61666d;
}

.duration-card {
  min-width: 142px;
  padding: 12px 16px;
  border: 1px solid #e3e5e7;
  border-radius: 6px;
  background: #f6f7f8;
  display: grid;
  gap: 4px;
}

.duration-card strong {
  font-size: 18px;
  color: #18191c;
}

.duration-card span {
  font-size: 12px;
  color: #9499a0;
}

.duration-label {
  color: #61666d;
}

.loading-state,
.empty-state {
  padding: 44px 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.stat-card {
  padding: 16px;
  border: 1px solid #e3e5e7;
  border-radius: 6px;
  background: #fff;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #18191c;
  line-height: 1.2;
}

.stat-label {
  margin-top: 6px;
  font-size: 12px;
  color: #61666d;
}

.insight-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 0.9fr);
  gap: 18px;
}

.section {
  padding: 18px;
  border: 1px solid #e3e5e7;
  border-radius: 6px;
  background: #fff;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.section-head h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #18191c;
}

.section-head span {
  font-size: 12px;
  color: #9499a0;
  white-space: nowrap;
}

.category-bars {
  display: grid;
  gap: 12px;
}

.bar-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 6px;
}

.bar-label {
  font-size: 13px;
  color: #18191c;
}

.bar-value {
  font-size: 12px;
  color: #9499a0;
}

.bar-track {
  height: 8px;
  background: #f1f2f3;
  border-radius: 999px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #00a1d6, #73c9e5);
  border-radius: inherit;
  transition: width 0.4s ease;
}

.hours-chart {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  align-items: end;
  height: 168px;
  gap: 3px;
  padding-top: 10px;
  border-bottom: 1px solid #e3e5e7;
}

.hour-cell {
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  align-items: center;
  gap: 5px;
}

.hour-fill {
  width: 100%;
  min-height: 0;
  background: #00a1d6;
  border-radius: 3px 3px 0 0;
  transition: height 0.35s ease;
}

.hour-cell span {
  height: 14px;
  font-size: 10px;
  color: #9499a0;
}

.tag-section {
  margin-top: 18px;
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.interest-tag {
  --tag-ratio: 0.5;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 220px;
  padding: 6px 11px;
  border: 1px solid rgba(0, 161, 214, calc(0.18 + var(--tag-ratio) * 0.32));
  border-radius: 999px;
  color: #0077a3;
  background: rgba(0, 161, 214, calc(0.06 + var(--tag-ratio) * 0.12));
  line-height: 1.2;
}

.interest-tag em {
  font-style: normal;
  font-size: 11px;
  color: #9499a0;
}

@media (max-width: 900px) {
  .panel-head,
  .insight-grid {
    grid-template-columns: 1fr;
  }

  .panel-head {
    display: grid;
  }

  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .interests-panel {
    padding: 16px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .hours-chart {
    gap: 2px;
  }
}
</style>
