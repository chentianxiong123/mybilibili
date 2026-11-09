<template>
  <div class="data-center-view">
    <el-tabs v-model="activeTab" class="data-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="数据概览" name="overview">
        <DataOverviewPanel />
      </el-tab-pane>
      <el-tab-pane label="账号诊断" name="diagnosis">
        <AccountDiagnosisPanel />
      </el-tab-pane>
      <el-tab-pane label="稿件分析" name="manuscript">
        <ManuscriptAnalysisPanel />
      </el-tab-pane>
      <el-tab-pane label="粉丝分析" name="fans">
        <FanAnalysisPanel :key="'fans-' + activeTab + '-' + refreshKey" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DataOverviewPanel from '@/components/datacenter/DataOverviewPanel.vue'
import AccountDiagnosisPanel from '@/components/datacenter/AccountDiagnosisPanel.vue'
import ManuscriptAnalysisPanel from '@/components/datacenter/ManuscriptAnalysisPanel.vue'
import FanAnalysisPanel from '@/components/datacenter/FanAnalysisPanel.vue'

const activeTab = ref('overview')
const refreshKey = ref(0)

function handleTabChange(tab) {
  if (tab === 'fans') {
    refreshKey.value++
  }
}
</script>

<style scoped>
.data-center-view {
  padding: 0;
}
.data-tabs {
  background: transparent;
}
.data-tabs :deep(.el-tabs__header) {
  margin-bottom: 20px;
  background: #fff;
  border-radius: 10px;
  padding: 0 16px;
  border: 1px solid #f0f0f0;
}
.data-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none;
}
.data-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  font-weight: 500;
  padding: 0 20px;
  height: 44px;
  line-height: 44px;
}
.placeholder-tab {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
}
.placeholder-tab p {
  color: #999;
  font-size: 15px;
}
</style>
