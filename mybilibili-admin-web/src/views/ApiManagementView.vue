<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getChannels, getChannel, createChannel, updateChannel, deleteChannel, toggleChannel,
  getBindings, bindFeature, testConnection, getAvailableTypes, getAvailableFeatures
} from '../api/channel'

const loading = ref(false)
const channels = ref([])
const bindings = ref({})
const availableTypes = ref(['LLM', 'ASR', 'TTS', 'IMAGE'])
const availableFeatures = ref([])
const activeType = ref('LLM')

// 抽屉
const drawerVisible = ref(false)
const drawerMode = ref('view')
const form = ref({
  id: null, name: '', type: 'LLM', baseUrl: '', apiKey: '', model: '', maxTokens: 2000, temperature: 0.7
})
const formRef = ref(null)
const rules = {
  name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  baseUrl: [{ required: true, message: '请输入 API 地址', trigger: 'blur' }],
  apiKey: [{ required: true, message: '请输入 API 密钥', trigger: 'blur' }],
  model: [{ required: true, message: '请输入模型名称', trigger: 'blur' }]
}

const testingId = ref(null)
const testResult = ref({})

const typeLabels = { LLM: '大模型', ASR: '语音识别', TTS: '语音合成', IMAGE: '图像生成', MODERATION: '内容审核' }
const typeColors = { LLM: '', ASR: 'success', TTS: 'warning', IMAGE: 'danger', MODERATION: 'warning' }

const filteredChannels = computed(() =>
  channels.value.filter(ch => ch.type === activeType.value)
)

const filteredFeatures = computed(() =>
  availableFeatures.value.filter(f => f.type === activeType.value)
)

async function loadData() {
  loading.value = true
  try {
    const [chRes, bindRes, typeRes, featureRes] = await Promise.all([
      getChannels(),
      getBindings(),
      getAvailableTypes(),
      getAvailableFeatures()
    ])
    if (chRes.code === 200) channels.value = chRes.data || []
    if (bindRes.code === 200) bindings.value = bindRes.data || {}
    if (typeRes.code === 200) availableTypes.value = typeRes.data || ['LLM', 'ASR', 'TTS', 'IMAGE']
    if (featureRes.code === 200) availableFeatures.value = featureRes.data || []
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function getBindingFeature(channelId) {
  return Object.entries(bindings.value).find(([, id]) => id === channelId)
}

function handleCreate() {
  drawerMode.value = 'create'
  form.value = { id: null, name: '', type: activeType.value, baseUrl: '', apiKey: '', model: '', maxTokens: 2000, temperature: 0.7 }
  testResult.value = null
  drawerVisible.value = true
}

function handleEdit(channel) {
  drawerMode.value = 'edit'
  form.value = { ...channel, apiKey: '' }
  testResult.value = null
  drawerVisible.value = true
}

function handleView(channel) {
  drawerMode.value = 'view'
  form.value = { ...channel, apiKey: '••••••••' }
  testResult.value = null
  drawerVisible.value = true
}

async function handleSave() {
  if (formRef.value) await formRef.value.validate()
  try {
    if (drawerMode.value === 'create') {
      const res = await createChannel(form.value)
      if (res.code === 200) { ElMessage.success('创建成功'); drawerVisible.value = false; loadData() }
      else ElMessage.error(res.message || '创建失败')
    } else {
      const data = { ...form.value }
      if (!data.apiKey || data.apiKey === '••••••••') delete data.apiKey
      const res = await updateChannel(form.value.id, data)
      if (res.code === 200) { ElMessage.success('保存成功'); drawerVisible.value = false; loadData() }
      else ElMessage.error(res.message || '保存失败')
    }
  } catch (e) { ElMessage.error(e.response?.data?.message || '操作失败') }
}

async function handleDelete(channel) {
  try {
    await ElMessageBox.confirm(`确定删除渠道「${channel.name}」？`, '确认', { type: 'warning' })
    const res = await deleteChannel(channel.id)
    if (res.code === 200) { ElMessage.success('已删除'); loadData() }
  } catch {}
}

async function handleToggle(channel) {
  try {
    await toggleChannel(channel.id)
    ElMessage.success(channel.enabled ? '已启用' : '已禁用')
    loadData()
  } catch { ElMessage.error('操作失败') }
}

async function handleTest(channelId) {
  testingId.value = channelId
  testResult.value[channelId] = null
  try {
    const res = await testConnection({ channelId, text: '你好，请回复"API测试成功"' })
    if (res.code === 200 && res.data) testResult.value[channelId] = res.data
    else testResult.value[channelId] = { success: false, message: res.message || '测试失败' }
  } catch (e) { testResult.value[channelId] = { success: false, message: '请求异常: ' + (e.response?.data?.message || e.message) } }
  finally { testingId.value = null }
}

async function handleBindFeature(feature, configId) {
  try {
    await bindFeature(feature, configId)
    ElMessage.success('绑定成功')
    loadData()
  } catch { ElMessage.error('绑定失败') }
}

function maskKey(key) {
  if (!key || key.length < 8) return key || ''
  return key.substring(0, 6) + '****' + key.substring(key.length - 4)
}

onMounted(loadData)
</script>

<template>
  <div class="api-management">
    <div class="page-header">
      <h2>AI 渠道管理</h2>
      <p class="desc">按类型管理 AI API 渠道，为不同功能分别配置对接来源</p>
    </div>

    <!-- 类型切换 -->
    <div class="type-tabs">
      <el-radio-group v-model="activeType">
        <el-radio-button v-for="t in availableTypes" :key="t" :value="t">
          {{ typeLabels[t] || t }}
        </el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="handleCreate" style="margin-left: 16px">
        <el-icon><Plus /></el-icon>
        添加{{ typeLabels[activeType] || activeType }}渠道
      </el-button>
    </div>

    <!-- 渠道卡片网格 -->
    <div v-loading="loading" class="channel-grid">
      <div v-for="ch in filteredChannels" :key="ch.id" class="channel-card" :class="{ disabled: !ch.enabled }" @click="handleView(ch)">
        <div class="card-top">
          <div class="card-title">{{ ch.name }}</div>
          <el-tag :type="ch.enabled ? 'success' : 'info'" size="small">{{ ch.enabled ? '已启用' : '已禁用' }}</el-tag>
        </div>
        <div class="card-model">{{ ch.model }}</div>
        <div class="card-url">{{ ch.baseUrl }}</div>
        <div class="card-features" v-if="getBindingFeature(ch.id)">
          <el-tag v-for="f in getBindingFeature(ch.id)" :key="f" size="small" type="primary" style="margin-right:4px">
            {{ f }}
          </el-tag>
        </div>
        <div class="card-actions" @click.stop>
          <el-button size="small" type="primary" :loading="testingId === ch.id" @click="handleTest(ch.id)">测试</el-button>
          <el-button size="small" @click="handleEdit(ch)">编辑</el-button>
          <el-button size="small" :type="ch.enabled ? 'warning' : 'success'" @click="handleToggle(ch)">
            {{ ch.enabled ? '禁用' : '启用' }}
          </el-button>
          <el-button size="small" type="danger" @click="handleDelete(ch)">删除</el-button>
        </div>
        <div v-if="testResult[ch.id]" class="card-test-result" @click.stop>
          <el-alert
            :type="testResult[ch.id].success ? 'success' : 'error'"
            :title="testResult[ch.id].message"
            :closable="true"
            @close="delete testResult[ch.id]"
            show-icon
          />
        </div>
      </div>

      <div v-if="filteredChannels.length === 0 && !loading" class="channel-empty">
        <el-empty :description="`暂无${typeLabels[activeType] || activeType}渠道，点击上方添加`" />
      </div>
    </div>

    <!-- 功能绑定设置 -->
    <el-card shadow="never" class="binding-card">
      <template #header><span class="card-title">{{ typeLabels[activeType] || activeType }} 功能绑定</span></template>
      <div class="binding-row" v-for="f in filteredFeatures" :key="f.value">
        <span class="binding-label">{{ f.label }}</span>
        <el-select
          :model-value="bindings[f.value] || null"
          placeholder="选择渠道"
          clearable
          size="default"
          @change="(val) => handleBindFeature(f.value, val)"
          style="width: 240px"
        >
          <el-option
            v-for="ch in channels.filter(c => c.enabled && c.type === activeType)"
            :key="ch.id"
            :label="ch.name"
            :value="ch.id"
          />
        </el-select>
      </div>
      <el-empty v-if="filteredFeatures.length === 0" description="该类型暂无功能绑定" />
    </el-card>

    <!-- 详情/编辑 抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      :title="drawerMode === 'create' ? '添加渠道' : drawerMode === 'edit' ? '编辑渠道' : '渠道详情'"
      size="450px"
      direction="rtl"
    >
      <el-form ref="formRef" :model="form" :rules="drawerMode !== 'view' ? rules : {}" label-width="80px" :disabled="drawerMode === 'view'">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 OpenAI Compatible / GLM ASR" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" :disabled="drawerMode !== 'create'" style="width: 100%">
            <el-option v-for="t in availableTypes" :key="t" :label="typeLabels[t] || t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 地址" prop="baseUrl">
          <el-input v-model="form.baseUrl" placeholder="https://api.example.com" />
          <div class="form-tip">填根地址即可，系统自动拼接对应接口路径</div>
        </el-form-item>
        <el-form-item label="API 密钥" prop="apiKey">
          <el-input v-model="form.apiKey" type="password" show-password
            :placeholder="drawerMode === 'edit' ? '留空则不修改' : '请输入 API Key'" />
        </el-form-item>
        <el-form-item label="模型" prop="model">
          <el-input v-model="form.model" placeholder="如 gpt-4.1-mini / whisper-1" />
        </el-form-item>
        <el-form-item label="Max Tokens">
          <el-input-number v-model="form.maxTokens" :min="100" :max="8000" :step="100" />
        </el-form-item>
        <el-form-item label="温度">
          <el-slider v-model="form.temperature" :min="0" :max="2" :step="0.1" show-input style="width: 300px" />
        </el-form-item>
      </el-form>

      <template #footer v-if="drawerMode !== 'view'">
        <el-button @click="drawerVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.api-management { padding: 0; }
.page-header { margin-bottom: 20px; }
.page-header h2 { margin: 0 0 8px; font-size: 22px; color: #303133; }
.page-header .desc { margin: 0; color: #909399; font-size: 14px; }
.type-tabs { display: flex; align-items: center; margin-bottom: 20px; }
.channel-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; margin-bottom: 24px; }
.channel-card {
  padding: 20px; background: #fff; border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,.08); cursor: pointer;
  transition: box-shadow .2s; border: 2px solid transparent;
}
.channel-card:hover { box-shadow: 0 4px 16px rgba(0,0,0,.12); border-color: #409eff; }
.channel-card.disabled { opacity: .6; }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.card-title { font-size: 16px; font-weight: 600; color: #303133; }
.card-model { font-size: 13px; color: #67c23a; font-weight: 500; margin-bottom: 4px; }
.card-url { font-size: 12px; color: #909399; margin-bottom: 8px; word-break: break-all; }
.card-features { margin-bottom: 8px; }
.card-actions { margin-top: 8px; display: flex; gap: 6px; flex-wrap: wrap; }
.card-test-result { margin-top: 10px; }
.channel-empty { grid-column: 1 / -1; }
.binding-card { margin-bottom: 20px; }
.card-title { font-size: 15px; font-weight: 600; color: #303133; }
.binding-row { display: flex; align-items: center; gap: 16px; margin-bottom: 12px; }
.binding-label { width: 80px; font-size: 14px; color: #606266; font-weight: 500; }
.form-tip { font-size: 12px; color: #909399; margin-top: 4px; line-height: 1.4; }
</style>
