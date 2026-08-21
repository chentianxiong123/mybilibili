<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { transcodeConfigApi } from '@/api/transcodeConfig'

const loading = ref(false)
const saving = ref(false)
const encoder = ref('auto')
const vaapi = ref(false)
const vaapiDev = ref('')
const options = ref(['auto', 'vaapi', 'x264'])

const vaapiStatusText = computed(() => {
  if (encoder.value === 'vaapi') return '硬编已启用'
  if (encoder.value === 'x264') return '硬编未启用（软件编码）'
  return vaapi.value ? '自动探测到可用，将优先硬编' : '未探测到，将回退软件编码'
})

const vaapiTagType = computed(() => {
  if (encoder.value === 'x264') return 'info'
  return vaapi.value ? 'success' : 'warning'
})

const encoderDesc = {
  auto: '自动探测：有 VAAPI 硬编则优先硬件编码，否则回退软件编码（推荐）',
  vaapi: '强制硬件编码：速度快、CPU 占用低，体积约为软件编码的 2 倍',
  x264: '强制软件编码：体积最省、画质最优，但转码时 CPU 占用高',
}

const fetchConfig = async () => {
  loading.value = true
  try {
    const res = await transcodeConfigApi.getConfig()
    const data = res.data?.code === 200 ? res.data.data : res
    encoder.value = data.encoder || 'auto'
    vaapi.value = !!data.vaapi
    vaapiDev.value = data.vaapiDev || ''
    options.value = data.options || ['auto', 'vaapi', 'x264']
  } catch (e) {
    ElMessage.error('获取转码配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    const res = await transcodeConfigApi.updateConfig({ encoder: encoder.value })
    if (res.data?.code === 200 || res.status === 'ok' || res.encoder) {
      ElMessage.success('配置已保存，重启转码服务后生效')
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchConfig)
</script>

<template>
  <div class="transcode-config-page" v-loading="loading">
    <div class="page-header">
      <h2>转码配置</h2>
      <p class="page-desc">设置视频转码编码方式，切换后需重启 work-service 生效</p>
    </div>

    <div class="config-sections">
      <el-card class="config-card">
        <template #header><span class="card-title">编码方式</span></template>
        <el-form label-width="140px">
          <el-form-item label="转码编码器">
            <el-select v-model="encoder" style="width: 300px">
              <el-option v-for="opt in options" :key="opt" :label="opt" :value="opt" />
            </el-select>
            <span class="param-hint">{{ encoderDesc[encoder] }}</span>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="config-card">
        <template #header><span class="card-title">本机硬件状态</span></template>
        <div class="hw-status">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="VAAPI 硬编可用">
              <el-tag :type="vaapi ? 'success' : 'info'">{{ vaapi ? '可用' : '不可用' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="设备节点">
              {{ vaapiDev || '/dev/dri/renderD128' }}
            </el-descriptions-item>
            <el-descriptions-item label="当前生效">
              <el-tag :type="vaapiTagType">{{ vaapiStatusText }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="提示">
              <span class="hint-text">无 GPU 环境自动回退软件编码，不会导致转码失败</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-card>

      <div class="config-actions">
        <el-button type="primary" size="large" :loading="saving" @click="saveConfig">保存配置</el-button>
        <span class="last-updated">保存后需重启 work-service 生效</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.transcode-config-page {
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
  max-width: 960px;
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

.hint-text {
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