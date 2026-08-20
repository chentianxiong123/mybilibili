<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, DataAnalysis, CircleCheck, CircleClose, Warning, Search, Clock, Files } from '@element-plus/icons-vue'
import { indexManagerApi } from '~/api/indexManager'

interface IndexStatus {
  engine: string
  config: string
  indexName: string
  totalCount: number
  publishedCount: number
  indexedCount: number
  nullCount: number
  coverage: number
  ginIndex: boolean
  trigger: boolean
  health: string
}

const indexStatus = ref<IndexStatus | null>(null)
const loading = ref(false)
const actionLoading = ref(false)
const validating = ref(false)
const lastValidatedAt = ref('')

const healthInfo = (health: string) => {
  switch (health) {
    case 'active': return { text: '正常', type: 'success' as const }
    case 'degraded': return { text: '部分缺失', type: 'warning' as const }
    default: return { text: '警告', type: 'danger' as const }
  }
}

const fetchIndexStatus = async () => {
  loading.value = true
  try {
    const res = await indexManagerApi.getStatus()
    if (res.code === 200) {
      indexStatus.value = res.data
    } else {
      ElMessage.error(res.message || '获取索引状态失败')
    }
  } catch (error) {
    console.error('获取索引状态失败:', error)
    ElMessage.error('获取索引状态失败')
  } finally {
    loading.value = false
  }
}

// 重建全文索引
const handleRebuildIndex = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要重建全文索引吗？将重新计算所有稿件的 search_vector（标题+描述，zhparser 中文分词）。',
      '重建全文索引',
      {
        confirmButtonText: '确定重建',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    actionLoading.value = true
    const res = await indexManagerApi.rebuild()
    if (res.code === 200) {
      ElMessage.success(res.data?.message || '全文索引重建完成')
      setTimeout(fetchIndexStatus, 1000)
    } else {
      ElMessage.error(res.message || '重建索引失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重建索引失败:', error)
      ElMessage.error('重建索引失败')
    }
  } finally {
    actionLoading.value = false
  }
}

// 校验索引
const handleValidate = async () => {
  validating.value = true
  try {
    const res = await indexManagerApi.validate()
    if (res.code === 200) {
      indexStatus.value = res.data
      lastValidatedAt.value = new Date().toLocaleTimeString()
      const h = healthInfo(res.data.health)
      if (h.type === 'success') {
        ElMessage.success('索引校验通过')
      } else {
        ElMessage.warning(`索引校验提示：${h.type === 'warning' ? '存在缺失要素' : '部分稿件未索引'}`)
      }
    } else {
      ElMessage.error(res.message || '校验失败')
    }
  } catch (error) {
    console.error('校验索引失败:', error)
    ElMessage.error('校验索引失败')
  } finally {
    validating.value = false
  }
}

onMounted(() => {
  fetchIndexStatus()
})
</script>

<template>
  <div class="index-manager">
    <div class="page-header">
      <h1 class="page-title">
        <el-icon><DataAnalysis /></el-icon>
        搜索索引管理
      </h1>
      <p class="page-desc">PostgreSQL 全文搜索索引（zh_cn / zhparser 中文分词）</p>
    </div>

    <!-- 状态卡片 -->
    <el-card class="summary-card" v-loading="loading">
      <div class="summary-row">
        <div class="summary-item">
          <div class="label">索引引擎</div>
          <div class="value">{{ indexStatus?.engine || '--' }}</div>
          <div class="sub">{{ indexStatus?.config || '' }}</div>
        </div>
        <div class="summary-item">
          <div class="label">索引名称</div>
          <div class="value">{{ indexStatus?.indexName || '--' }}</div>
        </div>
        <div class="summary-item">
          <div class="label">健康状态</div>
          <el-tag :type="healthInfo(indexStatus?.health || 'warning').type" effect="dark" round size="large">
            {{ healthInfo(indexStatus?.health || 'warning').text }}
          </el-tag>
        </div>
      </div>
      <el-divider />
      <div class="stat-grid">
        <div class="stat-item">
          <el-icon color="#00aeec"><Files /></el-icon>
          <div class="stat">
            <span class="stat-value">{{ indexStatus?.totalCount ?? '--' }}</span>
            <span class="stat-label">稿件总数</span>
          </div>
        </div>
        <div class="stat-item">
          <el-icon color="#67c23a"><CircleCheck /></el-icon>
          <div class="stat">
            <span class="stat-value">{{ indexStatus?.publishedCount ?? '--' }}</span>
            <span class="stat-label">已上架</span>
          </div>
        </div>
        <div class="stat-item">
          <el-icon color="#409eff"><DataAnalysis /></el-icon>
          <div class="stat">
            <span class="stat-value">{{ indexStatus?.indexedCount ?? '--' }}</span>
            <span class="stat-label">已索引</span>
          </div>
        </div>
        <div class="stat-item">
          <el-icon color="#e6a23c"><Clock /></el-icon>
          <div class="stat">
            <span class="stat-value">{{ indexStatus?.nullCount ?? '--' }}</span>
            <span class="stat-label">待索引</span>
          </div>
        </div>
        <div class="stat-item">
          <el-icon color="#00aeec"><DataAnalysis /></el-icon>
          <div class="stat">
            <span class="stat-value">{{ (indexStatus?.coverage ?? 0).toFixed(0) }}%</span>
            <span class="stat-label">索引覆盖率</span>
            <el-progress
              :percentage="Math.round(indexStatus?.coverage || 0)"
              :color="(indexStatus?.coverage || 0) >= 100 ? '#67c23a' : '#e6a23c'"
              :show-text="false"
              class="coverage-bar"
            />
          </div>
        </div>
      </div>
      <el-divider />
      <div class="feature-row">
        <div class="feature-item">
          <span class="feature-label">GIN 全文索引</span>
          <el-tag :type="indexStatus?.ginIndex ? 'success' : 'danger'" effect="light">
            <el-icon v-if="indexStatus?.ginIndex"><CircleCheck /></el-icon>
            <el-icon v-else><CircleClose /></el-icon>
            {{ indexStatus?.ginIndex ? '已启用' : '缺失' }}
          </el-tag>
        </div>
        <div class="feature-item">
          <span class="feature-label">自动维护触发器</span>
          <el-tag :type="indexStatus?.trigger ? 'success' : 'danger'" effect="light">
            <el-icon v-if="indexStatus?.trigger"><CircleCheck /></el-icon>
            <el-icon v-else><CircleClose /></el-icon>
            {{ indexStatus?.trigger ? '已启用' : '缺失' }}
          </el-tag>
        </div>
      </div>
    </el-card>

    <!-- 操作区域 -->
    <el-card class="action-card">
      <template #header>
        <div class="card-header">
          <span>索引操作</span>
          <el-button type="primary" :icon="Refresh" @click="fetchIndexStatus" :loading="loading">
            刷新状态
          </el-button>
        </div>
      </template>

      <div class="action-list">
        <div class="action-item">
          <div class="action-info">
            <h3><el-icon><Search /></el-icon> 校验索引</h3>
            <p>检查 search_vector 覆盖率、GIN 索引与自动维护触发器是否正常；统计待索引稿件数</p>
          </div>
          <el-button
            type="success"
            :icon="CircleCheck"
            @click="handleValidate"
            :loading="validating"
          >
            执行校验{{ lastValidatedAt ? `（上次 ${lastValidatedAt}）` : '' }}
          </el-button>
        </div>

        <el-divider />

        <div class="action-item">
          <div class="action-info">
            <h3><el-icon><Refresh /></el-icon> 重建全文索引</h3>
            <p>重新计算全部稿件的 search_vector（title + description，zh_cn 分词）；用于修复索引数据异常</p>
          </div>
          <el-button
            type="danger"
            :icon="Refresh"
            @click="handleRebuildIndex"
            :loading="actionLoading"
          >
            重建全文索引
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 说明区域 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span><el-icon><Warning /></el-icon> 使用说明</span>
        </div>
      </template>
      <div class="info-content">
        <h4>索引机制（自动维护）</h4>
        <ul>
          <li>搜索基于 <strong>PostgreSQL 全文搜索</strong>，使用 <strong>zh_cn</strong> 分词配置（zhparser 中文分词）。</li>
          <li>稿件标题/描述变更时，触发器 <code>trg_manuscript_search_vector</code> 自动重建 <code>search_vector</code>，<strong>无需手动批量索引</strong>。</li>
          <li>查询通过 <code>plainto_tsquery('zh_cn', ?)</code> 匹配，联想通过前缀 <code>to_tsquery('zh_cn', ?:*)</code>，热词独立存于 hot_search 表。</li>
        </ul>

        <h4>操作说明</h4>
        <ul>
          <li><strong>校验索引</strong>：查看索引完整性（覆盖率 / GIN 索引 / 触发器），代替旧版“批量/增量索引”。</li>
          <li><strong>重建全文索引</strong>：仅在 search_vector 数据异常或手动修复迁移数据后使用，刷新全部稿件索引。</li>
        </ul>

        <h4>注意事项</h4>
        <ul>
          <li>正常情况下索引由数据库触发器自动维护，<strong>无需执行任何索引操作</strong>。</li>
          <li>重建索引对当前数据量影响极小；数据量大时建议在低峰期执行。</li>
        </ul>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.index-manager {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}
.page-header {
  margin-bottom: 24px;
}
.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px;
}
.page-desc {
  color: #666;
  margin: 0;
}
.summary-card,
.action-card,
.info-card {
  margin-bottom: 24px;
}
.summary-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.summary-item {
  text-align: center;
  padding: 8px;
}
.summary-item .label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 6px;
}
.summary-item .value {
  font-size: 20px;
  font-weight: 600;
  color: #333;
}
.summary-item .sub {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
}
.stat-item {
  display: flex;
  align-items: center;
  gap: 12px;
}
.stat-item > .el-icon {
  font-size: 32px;
  flex-shrink: 0;
}
.stat {
  display: flex;
  flex-direction: column;
}
.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
}
.stat-label {
  font-size: 12px;
  color: #909399;
}
.coverage-bar {
  margin-top: 4px;
  width: 120px;
}
.feature-row {
  display: flex;
  gap: 32px;
}
.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
}
.feature-label {
  font-size: 14px;
  color: #606266;
}
.action-list {
  padding: 8px;
}
.action-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  gap: 24px;
}
.action-info {
  flex: 1;
}
.action-info h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 8px;
  font-size: 16px;
  color: #333;
}
.action-info p {
  margin: 0;
  font-size: 14px;
  color: #666;
}
.info-card {
  background-color: #f5f7fa;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.card-header > span {
  display: flex;
  align-items: center;
  gap: 8px;
}
.info-content h4 {
  margin: 16px 0 8px;
  font-size: 16px;
  color: #333;
}
.info-content h4:first-child {
  margin-top: 0;
}
.info-content ul {
  margin: 0;
  padding-left: 20px;
}
.info-content li {
  margin: 8px 0;
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}
.info-content code {
  background: #eef2f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SFMono-Regular', Consolas, monospace;
}
.info-content strong {
  color: #333;
}
</style>