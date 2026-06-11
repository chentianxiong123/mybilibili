<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { recommendConfigApi } from '../api/recommendConfig'

const loading = ref(false)
const config = ref({})

const fetchConfig = async () => {
  loading.value = true
  try {
    const res = await recommendConfigApi.getConfig()
    if (res.data.code === 200) {
      config.value = res.data.data
    }
  } catch (e) {
    ElMessage.error('获取推荐配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  try {
    const res = await recommendConfigApi.updateConfig(config.value)
    if (res.data.code === 200) {
      config.value = res.data.data
      ElMessage.success('保存成功')
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

const resetConfig = async () => {
  try {
    await ElMessageBox.confirm('确定要重置为默认配置吗？', '确认重置', { type: 'warning' })
    const res = await recommendConfigApi.resetConfig()
    if (res.data.code === 200) {
      config.value = res.data.data
      ElMessage.success('已重置为默认值')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('重置失败')
  }
}

onMounted(fetchConfig)
</script>

<template>
  <div class="recommend-config-page" v-loading="loading">
    <div class="page-header">
      <h2>推荐算法配置</h2>
      <p class="page-desc">调整推荐系统的各项参数，修改后实时生效</p>
    </div>

    <div class="config-sections" v-if="config">
      <el-card class="config-card">
        <template #header><span class="card-title">用户画像参数</span></template>
        <el-form label-width="160px">
          <el-form-item label="画像衰减因子">
            <el-input-number v-model="config.profileDecayFactor" :min="0.5" :max="1.0" :step="0.01" :precision="2" />
            <span class="param-hint">每次行为后旧权重乘以此值（越小遗忘越快）</span>
          </el-form-item>
          <el-form-item label="观看权重">
            <el-input-number v-model="config.watchWeight" :min="0.1" :max="10" :step="0.5" :precision="1" />
          </el-form-item>
          <el-form-item label="点赞权重">
            <el-input-number v-model="config.likeWeight" :min="0.1" :max="20" :step="0.5" :precision="1" />
          </el-form-item>
          <el-form-item label="收藏权重">
            <el-input-number v-model="config.collectWeight" :min="0.1" :max="30" :step="0.5" :precision="1" />
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="config-card">
        <template #header><span class="card-title">推荐算法参数</span></template>
        <el-form label-width="160px">
          <el-form-item label="分类 Boost">
            <el-input-number v-model="config.categoryBoost" :min="0.5" :max="10" :step="0.5" :precision="1" />
            <span class="param-hint">分类匹配时的权重放大倍数</span>
          </el-form-item>
          <el-form-item label="标签 Boost">
            <el-input-number v-model="config.tagBoost" :min="0.5" :max="10" :step="0.5" :precision="1" />
            <span class="param-hint">标签匹配时的权重放大倍数</span>
          </el-form-item>
          <el-form-item label="Top 分类数">
            <el-input-number v-model="config.topCategoryCount" :min="1" :max="20" :step="1" />
            <span class="param-hint">取用户画像中权重最高的N个分类</span>
          </el-form-item>
          <el-form-item label="Top 标签数">
            <el-input-number v-model="config.topTagCount" :min="1" :max="30" :step="1" />
            <span class="param-hint">取用户画像中权重最高的N个标签</span>
          </el-form-item>
          <el-form-item label="热门补充最少数">
            <el-input-number v-model="config.hotFillMinCount" :min="0" :max="20" :step="1" />
            <span class="param-hint">个性化结果不足时，至少补充N条热门</span>
          </el-form-item>
          <el-form-item label="新鲜度 Boost">
            <el-input-number v-model="config.freshnessBoost" :min="1.0" :max="5.0" :step="0.1" :precision="1" />
            <span class="param-hint">近期发布内容的额外加权</span>
          </el-form-item>
          <el-form-item label="新鲜度窗口(天)">
            <el-input-number v-model="config.freshnessWindowDays" :min="1" :max="30" :step="1" />
          </el-form-item>
          <el-form-item label="候选池倍数">
            <el-input-number v-model="config.candidateMultiplier" :min="1" :max="10" :step="1" />
            <span class="param-hint">先召回 size x 倍数的候选，再做排序和随机扰动</span>
          </el-form-item>
          <el-form-item label="热门随机权重">
            <el-input-number v-model="config.hotRandomWeight" :min="0" :max="1" :step="0.01" :precision="2" />
            <span class="param-hint">0 为纯热度排序，建议 0.10-0.25</span>
          </el-form-item>
          <el-form-item label="个性化随机权重">
            <el-input-number v-model="config.personalizedRandomWeight" :min="0" :max="1" :step="0.01" :precision="2" />
            <span class="param-hint">0 为纯画像相关排序，建议 0.05-0.20</span>
          </el-form-item>
          <el-form-item label="洗牌窗口">
            <el-input-number v-model="config.shuffleWindowSize" :min="1" :max="50" :step="1" />
            <span class="param-hint">只在前 N 个候选内随机重排，避免低质量内容冲到前排</span>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="config-card">
        <template #header><span class="card-title">防刷与限流</span></template>
        <el-form label-width="160px">
          <el-form-item label="播放去重窗口(分钟)">
            <el-input-number v-model="config.viewDeduplicationMinutes" :min="5" :max="1440" :step="5" />
            <span class="param-hint">同一用户对同一视频的播放计数冷却时间</span>
          </el-form-item>
          <el-form-item label="视频接口 QPS">
            <el-input-number v-model="config.sentinelVideoQps" :min="10" :max="5000" :step="10" />
          </el-form-item>
          <el-form-item label="用户接口 QPS">
            <el-input-number v-model="config.sentinelUserQps" :min="10" :max="5000" :step="10" />
          </el-form-item>
          <el-form-item label="搜索接口 QPS">
            <el-input-number v-model="config.sentinelSearchQps" :min="10" :max="5000" :step="10" />
          </el-form-item>
          <el-form-item label="登录接口 QPS">
            <el-input-number v-model="config.sentinelAuthQps" :min="1" :max="100" :step="1" />
            <span class="param-hint">防暴力破解，建议保持较低值</span>
          </el-form-item>
        </el-form>
      </el-card>

      <div class="config-actions">
        <el-button type="primary" size="large" @click="saveConfig">保存配置</el-button>
        <el-button size="large" @click="resetConfig">重置为默认值</el-button>
        <span class="last-updated" v-if="config.updatedAt">
          最后更新: {{ config.updatedAt }} ({{ config.updatedBy }})
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.recommend-config-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0 0 8px;
  font-size: 20px;
  color: #303133;
}

.page-desc {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.config-sections {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.config-card {
  border-radius: 8px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.param-hint {
  margin-left: 12px;
  font-size: 12px;
  color: #909399;
}

.config-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 0;
}

.last-updated {
  margin-left: auto;
  font-size: 12px;
  color: #909399;
}
</style>
