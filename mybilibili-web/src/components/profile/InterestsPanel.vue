<script setup>
import { ref, computed, onMounted } from 'vue'
import { profileApi } from '../../api/profile.js'
import { ElMessage } from 'element-plus'

const loading = ref(true)
const profile = ref(null)

const topCategories = computed(() => {
  if (!profile.value?.categoryWeights) return []
  return Object.entries(profile.value.categoryWeights)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 8)
})

const topTags = computed(() => {
  if (!profile.value?.tagWeights) return []
  return Object.entries(profile.value.tagWeights)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 20)
})

const maxTagWeight = computed(() => {
  if (topTags.value.length === 0) return 1
  return topTags.value[0][1]
})

const maxCategoryWeight = computed(() => {
  if (topCategories.value.length === 0) return 1
  return topCategories.value[0][1]
})

const activeHoursData = computed(() => {
  if (!profile.value?.activeHours) return []
  const hours = []
  for (let i = 0; i < 24; i++) {
    hours.push({
      hour: i,
      label: `${i}:00`,
      count: profile.value.activeHours[String(i)] || 0
    })
  }
  return hours
})

const maxHourCount = computed(() => {
  return Math.max(...activeHoursData.value.map(h => h.count), 1)
})

const durationLabel = computed(() => {
  const map = { short: '短视频 (<5分钟)', medium: '中等时长 (5-20分钟)', long: '长视频 (>20分钟)' }
  return map[profile.value?.preferredDuration] || '暂无数据'
})

const fetchProfile = async () => {
  loading.value = true
  try {
    const res = await profileApi.getMyProfile()
    if (res.code === 200) {
      profile.value = res.data
    }
  } catch (e) {
    console.error('获取画像失败:', e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchProfile)
</script>

<template>
  <div class="interests-panel">
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="6" animated />
    </div>

    <div v-else-if="!profile || (!topCategories.length && !topTags.length)" class="empty-state">
      <el-empty description="暂无兴趣数据，多看看视频就有啦~" />
    </div>

    <div v-else class="profile-content">
      <div class="section">
        <h3 class="section-title">兴趣标签</h3>
        <div class="tag-cloud">
          <span
            v-for="[tag, weight] in topTags"
            :key="tag"
            class="interest-tag"
            :style="{
              fontSize: (12 + (weight / maxTagWeight) * 14) + 'px',
              opacity: 0.5 + (weight / maxTagWeight) * 0.5
            }"
          >{{ tag }}</span>
        </div>
      </div>

      <div class="section" v-if="topCategories.length">
        <h3 class="section-title">分类偏好</h3>
        <div class="category-bars">
          <div v-for="[catId, weight] in topCategories" :key="catId" class="bar-item">
            <span class="bar-label">分类 {{ catId }}</span>
            <div class="bar-track">
              <div class="bar-fill" :style="{ width: (weight / maxCategoryWeight * 100) + '%' }"></div>
            </div>
            <span class="bar-value">{{ weight.toFixed(1) }}</span>
          </div>
        </div>
      </div>

      <div class="section">
        <h3 class="section-title">观看时段分布</h3>
        <div class="hours-chart">
          <div v-for="h in activeHoursData" :key="h.hour" class="hour-bar">
            <div class="hour-fill" :style="{ height: (h.count / maxHourCount * 100) + '%' }"></div>
            <span class="hour-label" v-if="h.hour % 3 === 0">{{ h.hour }}</span>
          </div>
        </div>
      </div>

      <div class="section stats-row">
        <div class="stat-card">
          <div class="stat-value">{{ profile.totalWatchCount || 0 }}</div>
          <div class="stat-label">观看次数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ profile.totalLikeCount || 0 }}</div>
          <div class="stat-label">点赞次数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ profile.totalCollectCount || 0 }}</div>
          <div class="stat-label">收藏次数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ durationLabel }}</div>
          <div class="stat-label">偏好时长</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.interests-panel {
  padding: 20px;
}

.loading-state, .empty-state {
  padding: 40px 0;
}

.section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 16px;
  padding-left: 10px;
  border-left: 3px solid #00a1d6;
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.interest-tag {
  color: #00a1d6;
  background: rgba(0, 161, 214, 0.08);
  padding: 4px 12px;
  border-radius: 16px;
  cursor: default;
  transition: all 0.2s;
}

.interest-tag:hover {
  background: rgba(0, 161, 214, 0.15);
}

.category-bars {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bar-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bar-label {
  width: 80px;
  font-size: 13px;
  color: #61666d;
  text-align: right;
  flex-shrink: 0;
}

.bar-track {
  flex: 1;
  height: 20px;
  background: #f1f2f3;
  border-radius: 10px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #00a1d6, #00b5e5);
  border-radius: 10px;
  transition: width 0.6s ease;
}

.bar-value {
  width: 40px;
  font-size: 12px;
  color: #9499a0;
}

.hours-chart {
  display: flex;
  align-items: flex-end;
  height: 100px;
  gap: 2px;
  padding: 0 4px;
  border-bottom: 1px solid #e3e5e7;
}

.hour-bar {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  justify-content: flex-end;
}

.hour-fill {
  width: 100%;
  background: linear-gradient(180deg, #00a1d6, #00b5e5);
  border-radius: 2px 2px 0 0;
  min-height: 2px;
  transition: height 0.4s ease;
}

.hour-label {
  font-size: 10px;
  color: #9499a0;
  margin-top: 4px;
}

.stats-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.stat-card {
  flex: 1;
  min-width: 120px;
  background: #f6f7f8;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #9499a0;
}
</style>
