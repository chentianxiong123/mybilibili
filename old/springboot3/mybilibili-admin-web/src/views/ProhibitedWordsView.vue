<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getProhibitedWordList,
  addProhibitedWord,
  updateProhibitedWord,
  deleteProhibitedWord,
  batchImportProhibitedWords
} from '../api/prohibitedWord'
import { getSecuritySettings, updateSecuritySettings } from '../api/securitySettings'

// 表格数据
const tableData = ref([])
const loading = ref(false)
const total = ref(0)

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

// 搜索
const keyword = ref('')

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('添加违禁词')
const dialogForm = ref({
  id: null,
  word: '',
  matchType: 'CONTAINS',
  category: '',
  isEnabled: 1
})
const dialogFormRef = ref(null)

// 批量导入对话框
const importDialogVisible = ref(false)
const importForm = ref({
  file: null,
  matchType: 'CONTAINS',
  category: ''
})
const fileList = ref([])

// 安全设置
const securityLoading = ref(false)
const securityForm = ref({
  commentMaxCount: 10,
  commentWindowSeconds: 60,
  replyMaxCount: 20,
  replyWindowSeconds: 60,
  cacheRefreshIntervalSeconds: 300
})

// 匹配类型选项
const matchTypeOptions = [
  { label: '包含匹配', value: 'CONTAINS' },
  { label: '精确匹配', value: 'EXACT' }
]

// 分类选项
const categoryOptions = [
  { label: '政治', value: 'POLITICS' },
  { label: '色情', value: 'PORN' },
  { label: '广告', value: 'AD' },
  { label: '暴力', value: 'VIOLENCE' },
  { label: '其他', value: 'OTHER' }
]

// 表单验证规则
const rules = {
  word: [
    { required: true, message: '请输入违禁词', trigger: 'blur' }
  ]
}

// 加载违禁词列表
const loadWords = async () => {
  loading.value = true
  try {
    const res = await getProhibitedWordList({
      page: currentPage.value,
      size: pageSize.value,
      keyword: keyword.value
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || res.data || []
      total.value = res.data?.total || res.total || 0
    }
  } catch (error) {
    ElMessage.error('获取违禁词列表失败')
  } finally {
    loading.value = false
  }
}

// 加载安全设置
const loadSecuritySettings = async () => {
  securityLoading.value = true
  try {
    const res = await getSecuritySettings()
    if (res.code === 200 || res.success) {
      const data = res.data || {}
      securityForm.value.commentMaxCount = data.commentMaxCount || 10
      securityForm.value.commentWindowSeconds = data.commentWindowSeconds || 60
      securityForm.value.replyMaxCount = data.replyMaxCount || 20
      securityForm.value.replyWindowSeconds = data.replyWindowSeconds || 60
      securityForm.value.cacheRefreshIntervalSeconds = data.cacheRefreshIntervalSeconds || 300
    }
  } catch (error) {
    ElMessage.error('获取安全设置失败')
  } finally {
    securityLoading.value = false
  }
}

// 保存安全设置
const handleSaveSecurity = async () => {
  securityLoading.value = true
  try {
    const res = await updateSecuritySettings(securityForm.value)
    if (res.code === 200 || res.success) {
      ElMessage.success('保存成功')
    }
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    securityLoading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadWords()
}

// 重置搜索
const handleReset = () => {
  keyword.value = ''
  currentPage.value = 1
  loadWords()
}

// 分页改变
const handlePageChange = (page) => {
  currentPage.value = page
  loadWords()
}

// 打开添加对话框
const handleAdd = () => {
  dialogTitle.value = '添加违禁词'
  dialogForm.value = {
    id: null,
    word: '',
    matchType: 'CONTAINS',
    category: '',
    isEnabled: 1
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (row) => {
  dialogTitle.value = '编辑违禁词'
  dialogForm.value = {
    id: row.id,
    word: row.word,
    matchType: row.matchType || 'CONTAINS',
    category: row.category || '',
    isEnabled: row.isEnabled
  }
  dialogVisible.value = true
}

// 保存
const handleSave = async () => {
  if (!dialogFormRef.value) return

  await dialogFormRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      let res
      if (dialogForm.value.id) {
        res = await updateProhibitedWord(dialogForm.value.id, dialogForm.value)
      } else {
        res = await addProhibitedWord(dialogForm.value)
      }

      if (res.code === 200 || res.success) {
        ElMessage.success(dialogForm.value.id ? '更新成功' : '添加成功')
        dialogVisible.value = false
        loadWords()
      }
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      loading.value = false
    }
  })
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该违禁词吗？此操作不可恢复！',
      '警告',
      { type: 'error' }
    )
    const res = await deleteProhibitedWord(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadWords()
    }
  } catch {}
}

// 打开批量导入对话框
const handleOpenImport = () => {
  importForm.value = {
    file: null,
    matchType: 'CONTAINS',
    category: ''
  }
  fileList.value = []
  importDialogVisible.value = true
}

// 文件选择变化
const handleFileChange = (file) => {
  importForm.value.file = file.raw
}

// 批量导入
const handleImport = async () => {
  if (!importForm.value.file) {
    ElMessage.warning('请选择文件')
    return
  }

  loading.value = true
  try {
    const formData = new FormData()
    formData.append('file', importForm.value.file)
    formData.append('matchType', importForm.value.matchType)
    if (importForm.value.category) {
      formData.append('category', importForm.value.category)
    }

    const res = await batchImportProhibitedWords(formData)
    if (res.code === 200 || res.success) {
      ElMessage.success(`导入完成：成功 ${res.data?.successCount || 0} 条，失败 ${res.data?.failCount || 0} 条`)
      importDialogVisible.value = false
      loadWords()
    }
  } catch (error) {
    ElMessage.error('导入失败')
  } finally {
    loading.value = false
  }
}

// 获取匹配类型标签
const getMatchTypeLabel = (type) => {
  const option = matchTypeOptions.find(opt => opt.value === type)
  return option ? option.label : type
}

// 获取分类标签
const getCategoryLabel = (category) => {
  if (!category) return '-'
  const option = categoryOptions.find(opt => opt.value === category)
  return option ? option.label : category
}

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

onMounted(() => {
  loadWords()
  loadSecuritySettings()
})
</script>

<template>
  <div class="prohibited-words-page">
    <h2 class="page-title">违禁词与安全设置</h2>

    <!-- 搜索和操作区域 -->
    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索违禁词"
        clearable
        style="width: 250px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
      <el-button type="success" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加违禁词
      </el-button>
      <el-button type="warning" @click="handleOpenImport">
        <el-icon><Upload /></el-icon>
        批量导入
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="word" label="违禁词" width="200" />
      <el-table-column label="匹配类型" width="120">
        <template #default="{ row }">
          <el-tag :type="row.matchType === 'EXACT' ? 'warning' : 'info'">
            {{ getMatchTypeLabel(row.matchType) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="分类" width="120">
        <template #default="{ row }">
          {{ getCategoryLabel(row.category) }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.isEnabled === 1 ? 'success' : 'danger'">
            {{ row.isEnabled === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="150">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="loadWords"
      />
    </div>

    <!-- 安全设置区域 -->
    <el-card v-loading="securityLoading" class="security-card">
      <template #header>
        <div class="security-header">
          <span class="security-title">频率限制与缓存设置</span>
        </div>
      </template>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-divider content-position="left">评论频率限制</el-divider>
          <el-form-item label="最大评论次数">
            <el-input-number
              v-model="securityForm.commentMaxCount"
              :min="1"
              :max="1000"
              style="width: 100%"
            />
            <div class="form-tip">用户在时间窗口内最多可发送的评论数量</div>
          </el-form-item>
          <el-form-item label="时间窗口（秒）">
            <el-input-number
              v-model="securityForm.commentWindowSeconds"
              :min="1"
              :max="3600"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>

        <el-col :span="12">
          <el-divider content-position="left">回复频率限制</el-divider>
          <el-form-item label="最大回复次数">
            <el-input-number
              v-model="securityForm.replyMaxCount"
              :min="1"
              :max="1000"
              style="width: 100%"
            />
            <div class="form-tip">用户在时间窗口内最多可发送的回复数量</div>
          </el-form-item>
          <el-form-item label="时间窗口（秒）">
            <el-input-number
              v-model="securityForm.replyWindowSeconds"
              :min="1"
              :max="3600"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">缓存设置</el-divider>
      <el-form-item label="违禁词缓存刷新间隔（秒）">
        <el-input-number
          v-model="securityForm.cacheRefreshIntervalSeconds"
          :min="10"
          :max="3600"
          style="width: 100%"
        />
        <div class="form-tip">违禁词缓存在 Redis 中的自动刷新间隔（建议值：60-600）</div>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSaveSecurity">保存设置</el-button>
      </el-form-item>
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
    >
      <el-form
        ref="dialogFormRef"
        :model="dialogForm"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="违禁词" prop="word">
          <el-input v-model="dialogForm.word" placeholder="请输入违禁词" />
        </el-form-item>
        <el-form-item label="匹配类型">
          <el-radio-group v-model="dialogForm.matchType">
            <el-radio label="CONTAINS">包含匹配</el-radio>
            <el-radio label="EXACT">精确匹配</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="dialogForm.category" placeholder="请选择分类" clearable style="width: 100%">
            <el-option
              v-for="opt in categoryOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dialogForm.isEnabled">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <el-dialog
      v-model="importDialogVisible"
      title="批量导入违禁词"
      width="500px"
    >
      <el-form label-width="100px">
        <el-form-item label="选择文件" required>
          <el-upload
            v-model:file-list="fileList"
            :auto-upload="false"
            :limit="1"
            accept=".txt"
            @change="handleFileChange"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .txt 文件，每行一个违禁词
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="匹配类型">
          <el-radio-group v-model="importForm.matchType">
            <el-radio label="CONTAINS">包含匹配</el-radio>
            <el-radio label="EXACT">精确匹配</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="importForm.category" placeholder="请选择分类" clearable style="width: 100%">
            <el-option
              v-for="opt in categoryOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.prohibited-words-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.security-card {
  margin-top: 30px;
}

.security-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.security-title {
  font-weight: 600;
  font-size: 16px;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  line-height: 1.4;
}
</style>