<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAiSkills,
  getAiSkillsByType,
  createAiSkill,
  updateAiSkill,
  deleteAiSkill,
  toggleAiSkill
} from '../api/aiSkill'

const loading = ref(false)
const tableData = ref([])
const activeTypeFilter = ref('CUSTOMER_SERVICE') // 默认显示客服技能

// 弹窗
const dialogVisible = ref(false)
const dialogTitle = ref('添加技能')
const dialogForm = ref({
  id: null,
  name: '',
  description: '',
  type: 'CUSTOMER_SERVICE',
  enabled: true,
  systemPrompt: '',
  fewShotExamples: ''
})
const dialogFormRef = ref(null)

// 预览相关
const previewVisible = ref(false)
const previewQuestion = ref('')
const previewResult = ref('')

const rules = {
  name: [{ required: true, message: '请输入技能名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

// 类型选项
const typeOptions = [
  { label: '客户服务', value: 'CUSTOMER_SERVICE' },
  { label: '内容审核', value: 'REVIEW' },
  { label: '管理后台', value: 'ADMIN' }
]

// 筛选Tab配置
const filterTabs = [
  { label: '全部', value: 'ALL' },
  { label: '客服', value: 'CUSTOMER_SERVICE' },
  { label: '审核', value: 'REVIEW' },
  { label: '管理', value: 'ADMIN' }
]

const typeColors = {
  CUSTOMER_SERVICE: 'primary',
  REVIEW: 'success',
  ADMIN: 'warning'
}

function getTypeLabel(type) {
  const opt = typeOptions.find(o => o.value === type)
  return opt ? opt.label : type
}

// 过滤后的数据
const filteredData = computed(() => {
  if (activeTypeFilter.value === 'ALL') {
    return tableData.value
  }
  return tableData.value.filter(item => item.type === activeTypeFilter.value)
})

async function loadData() {
  loading.value = true
  try {
    const res = await getAiSkills()
    if (res.code === 200 || res.success) {
      tableData.value = res.data || []
    }
  } catch (e) {
    ElMessage.error('获取技能列表失败')
  } finally {
    loading.value = false
  }
}

// 切换筛选类型
async function handleTypeFilterChange(type) {
  activeTypeFilter.value = type
  if (type !== 'ALL') {
    loading.value = true
    try {
      const res = await getAiSkillsByType(type)
      if (res.code === 200 || res.success) {
        tableData.value = res.data || []
      }
    } catch (e) {
      ElMessage.error('获取技能列表失败')
    } finally {
      loading.value = false
    }
  } else {
    await loadData()
  }
}

function handleAdd() {
  dialogTitle.value = '添加技能'
  dialogForm.value = {
    id: null,
    name: '',
    description: '',
    type: 'CUSTOMER_SERVICE',
    enabled: true,
    systemPrompt: '',
    fewShotExamples: ''
  }
  dialogVisible.value = true
}

function handleEdit(row) {
  dialogTitle.value = '编辑技能'
  dialogForm.value = {
    id: row.id,
    name: row.name,
    description: row.description || '',
    type: row.type,
    enabled: row.enabled,
    systemPrompt: row.systemPrompt || '',
    fewShotExamples: row.fewShotExamples || ''
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (dialogFormRef.value) await dialogFormRef.value.validate()
  loading.value = true
  try {
    let res
    if (dialogForm.value.id) {
      res = await updateAiSkill(dialogForm.value.id, dialogForm.value)
    } else {
      res = await createAiSkill(dialogForm.value)
    }
    if (res.code === 200 || res.success) {
      ElMessage.success(dialogForm.value.id ? '更新成功' : '添加成功')
      dialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '操作失败')
  } finally {
    loading.value = false
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除技能「${row.name}」？`, '确认', { type: 'warning' })
    const res = await deleteAiSkill(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('已删除')
      loadData()
    }
  } catch {}
}

async function handleToggle(row) {
  try {
    await toggleAiSkill(row.id)
    ElMessage.success(row.enabled ? '已禁用' : '已启用')
    loadData()
  } catch {
    ElMessage.error('操作失败')
  }
}

// 预览技能匹配
function handlePreview() {
  if (!previewQuestion.value.trim()) {
    ElMessage.warning('请输入要测试的问题')
    return
  }
  const question = previewQuestion.value.toLowerCase()
  const matched = filteredData.value.filter(skill => {
    if (!skill.enabled) return false
    const desc = (skill.description || '').toLowerCase()
    const name = (skill.name || '').toLowerCase()
    // 简单关键词匹配
    return desc.includes(question) || name.includes(question) ||
      question.includes(desc) || question.includes(name)
  })
  if (matched.length > 0) {
    previewResult.value = matched.map(s => `✓ ${s.name} (${getTypeLabel(s.type)})\n   描述: ${s.description}`).join('\n\n')
  } else {
    previewResult.value = '未找到匹配的技能'
  }
}

function formatDateTime(dateStr) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

onMounted(loadData)
</script>

<template>
  <div class="ai-skills-page">
    <div class="page-header">
      <h2>AI 技能管理</h2>
      <p class="desc">配置 AI 技能的系统指令和对话示例。技能描述用于 AI 判断何时调用此技能</p>
    </div>

    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加技能
      </el-button>
      <el-button @click="previewVisible = true">
        <el-icon><Search /></el-icon>
        测试匹配
      </el-button>
    </div>

    <!-- 类型筛选 Tab -->
    <el-tabs v-model="activeTypeFilter" @tab-change="handleTypeFilterChange" class="type-filter-tabs">
      <el-tab-pane
        v-for="tab in filterTabs"
        :key="tab.value"
        :label="tab.label"
        :name="tab.value"
      />
    </el-tabs>

    <el-table v-loading="loading" :data="filteredData" style="width: 100%">
      <el-table-column prop="name" label="名称" width="150" show-overflow-tooltip />
      <el-table-column label="类型" width="120">
        <template #default="{ row }">
          <el-tag :type="typeColors[row.type] || ''">
            {{ getTypeLabel(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述（路由提示）" min-width="250" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="description-cell">{{ row.description || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="启用" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">
            {{ row.enabled ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="180">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button link :type="row.enabled ? 'warning' : 'success'" size="small" @click="handleToggle(row)">
            {{ row.enabled ? '禁用' : '启用' }}
          </el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="dialogFormRef" :model="dialogForm" :rules="rules" label-width="100px">
        <el-form-item label="技能名称" prop="name">
          <el-input v-model="dialogForm.name" placeholder="如 哔哩客服助手" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="dialogForm.type" style="width: 100%">
            <el-option
              v-for="opt in typeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="dialogForm.description" placeholder="简短描述此技能的业务范围" />
          <div class="form-tip">简短描述此技能的业务范围，AI会根据此描述判断何时使用此技能</div>
        </el-form-item>
        <el-form-item label="系统指令">
          <el-input
            v-model="dialogForm.systemPrompt"
            type="textarea"
            :rows="10"
            placeholder="此技能的业务规则文档内容。AI回答问题前会加载此内容。支持多段规则，每段用 --- 分隔"
          />
          <div class="form-tip">此技能的业务规则文档内容。AI回答问题前会加载此内容。支持多段规则，每段用 --- 分隔</div>
        </el-form-item>
        <el-form-item label="对话示例">
          <el-input
            v-model="dialogForm.fewShotExamples"
            type="textarea"
            :rows="4"
            placeholder="示例对话，JSON格式或纯文本均可"
          />
          <div class="form-tip">示例对话，JSON格式或纯文本均可</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 预览匹配弹窗 -->
    <el-dialog v-model="previewVisible" title="测试技能匹配" width="500px">
      <p class="preview-tip">输入一个问题，查看哪个技能会被匹配（基于描述关键词）</p>
      <el-input
        v-model="previewQuestion"
        type="textarea"
        :rows="3"
        placeholder="如：我的账号被盗了怎么办？"
      />
      <el-button type="primary" style="margin-top: 12px" @click="handlePreview">测试</el-button>
      <el-input
        v-model="previewResult"
        type="textarea"
        :rows="6"
        readonly
        placeholder="匹配结果将显示在这里"
        style="margin-top: 12px"
      />
    </el-dialog>
  </div>
</template>

<style scoped>
.ai-skills-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0 0 8px;
  font-size: 22px;
  color: #303133;
}

.page-header .desc {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.toolbar {
  margin-bottom: 16px;
  display: flex;
  gap: 10px;
}

.type-filter-tabs {
  margin-bottom: 16px;
}

.description-cell {
  color: #606266;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.preview-tip {
  margin: 0 0 12px;
  color: #909399;
  font-size: 14px;
}
</style>
