<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getAiConfig, updateAiConfig, testAiConnection } from '../api/apiManagement'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const testing = ref(false)
const config = reactive({
  apiKey: '',
  apiUrl: '',
  model: '',
  maxTokens: 2000,
  temperature: 0.7,
  hasApiKey: false
})

const newApiKey = ref('')
const testText = ref('你好，请回复"API测试成功"')
const testResult = ref(null)
const showTestResult = ref(false)
const showApiKey = ref(false)

async function loadConfig() {
  loading.value = true
  try {
    const res = await getAiConfig()
    if (res.code === 200 && res.data) {
      config.apiKey = res.data.apiKey || ''
      config.apiUrl = res.data.apiUrl || ''
      config.model = res.data.model || ''
      config.maxTokens = res.data.maxTokens || 2000
      config.temperature = res.data.temperature || 0.7
      config.hasApiKey = res.data.hasApiKey || false
    }
  } catch (e) {
    ElMessage.error('获取配置失败')
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  const updates = {}
  if (newApiKey.value) {
    updates.apiKey = newApiKey.value
  }
  if (config.apiUrl) updates.apiUrl = config.apiUrl
  if (config.model) updates.model = config.model
  updates.maxTokens = config.maxTokens
  updates.temperature = config.temperature

  try {
    await updateAiConfig(updates)
    ElMessage.success('配置已保存')
    newApiKey.value = ''
    await loadConfig()
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

async function handleTest() {
  testing.value = true
  testResult.value = null
  showTestResult.value = true
  try {
    const res = await testAiConnection(testText.value || '你好，请回复"API测试成功"')
    if (res.code === 200 && res.data) {
      testResult.value = res.data
    } else {
      testResult.value = { success: false, message: res.msg || '测试失败' }
    }
  } catch (e) {
    testResult.value = { success: false, message: '请求异常: ' + e.message }
  } finally {
    testing.value = false
  }
}

function resetApiKey() {
  newApiKey.value = ''
}

onMounted(() => {
  loadConfig()
})
</script>

<template>
  <div class="api-management">
    <div class="page-header">
      <h2>API 管理</h2>
      <p class="desc">管理 AI 服务 API 密钥和配置，支持在线测试连接</p>
    </div>

    <el-row :gutter="20">
      <!-- 配置表单 -->
      <el-col :span="16">
        <el-card shadow="never" class="config-card">
          <template #header>
            <div class="card-header">
              <span>AI 服务配置</span>
              <el-tag :type="config.hasApiKey ? 'success' : 'danger'" size="small">
                {{ config.hasApiKey ? '已配置' : '未配置' }}
              </el-tag>
            </div>
          </template>

          <el-form label-width="120px" v-loading="loading">
            <el-form-item label="API 密钥">
              <div class="api-key-row">
                <el-input
                  v-model="newApiKey"
                  :type="showApiKey ? 'text' : 'password'"
                  :placeholder="config.hasApiKey ? '留空则保持当前密钥不变' : '请输入 API 密钥'"
                  clearable
                >
                  <template #append>
                    <el-button @click="showApiKey = !showApiKey">
                      <el-icon><View v-if="!showApiKey" /><Hide v-else /></el-icon>
                    </el-button>
                  </template>
                </el-input>
                <span class="current-key-hint" v-if="config.apiKey && !newApiKey">
                  当前: {{ config.apiKey }}
                </span>
              </div>
            </el-form-item>

            <el-form-item label="API 地址">
              <el-input v-model="config.apiUrl" placeholder="https://api.deepseek.com/chat/completions" />
            </el-form-item>

            <el-form-item label="模型">
              <el-input v-model="config.model" placeholder="deepseek-chat" />
            </el-form-item>

            <el-form-item label="最大 Token">
              <el-input-number v-model="config.maxTokens" :min="100" :max="8000" :step="100" />
            </el-form-item>

            <el-form-item label="温度">
              <el-slider
                v-model="config.temperature"
                :min="0"
                :max="2"
                :step="0.1"
                show-input
                style="width: 300px"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleSave" :loading="loading">
                保存配置
              </el-button>
              <el-button @click="loadConfig" :loading="loading">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 测试区域 -->
      <el-col :span="8">
        <el-card shadow="never" class="test-card">
          <template #header>
            <span>连接测试</span>
          </template>

          <el-form label-width="0">
            <el-form-item label="">
              <el-input
                v-model="testText"
                type="textarea"
                :rows="3"
                placeholder="输入测试文本"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="success"
                @click="handleTest"
                :loading="testing"
                :disabled="!config.hasApiKey && !newApiKey"
              >
                {{ testing ? '测试中...' : '测试连接' }}
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 测试结果 -->
          <div v-if="showTestResult" class="test-result">
            <div v-if="testing" class="testing">
              <el-icon class="is-loading" :size="24"><Loading /></el-icon>
              <span>正在测试 API 连接...</span>
            </div>

            <template v-else-if="testResult">
              <div :class="['result-banner', testResult.success ? 'success' : 'error']">
                <el-icon :size="20">
                  <SuccessFilled v-if="testResult.success" />
                  <WarningFilled v-else />
                </el-icon>
                <span>{{ testResult.message }}</span>
              </div>

              <div v-if="testResult.responseTime" class="result-item">
                <label>响应时间</label>
                <span>{{ testResult.responseTime }}</span>
              </div>

              <div v-if="testResult.response" class="result-item response">
                <label>响应内容</label>
                <el-input
                  type="textarea"
                  :rows="4"
                  :model-value="testResult.response"
                  readonly
                />
              </div>
            </template>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.api-management {
  padding: 0;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 22px;
  color: #303133;
}

.page-header .desc {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.config-card,
.config-card,
.test-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.api-key-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.current-key-hint {
  font-size: 12px;
  color: #909399;
}

.test-result {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.testing {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-size: 14px;
}

.result-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 14px;
  margin-bottom: 12px;
}

.result-banner.success {
  background-color: #f0f9eb;
  color: #67c23a;
}

.result-banner.error {
  background-color: #fef0f0;
  color: #f56c6c;
}

.result-item {
  margin-bottom: 10px;
}

.result-item label {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.result-item.response {
  margin-bottom: 0;
}
</style>