<script setup>
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCategoryList,
  addCategory,
  updateCategory,
  deleteCategory
} from '@/api/category'

// 分区扁平列表
const rawCategories = ref([])
const loading = ref(false)
const keyword = ref('')

// 构建树
const categoryTree = computed(() => {
  const flat = rawCategories.value
  const byId = {}
  flat.forEach(c => { byId[c.id] = { ...c, children: [] } })
  const roots = []
  flat.forEach(c => {
    const node = byId[c.id]
    if (c.parent_id && byId[c.parent_id]) {
      byId[c.parent_id].children.push(node)
    } else {
      roots.push(node)
    }
  })
  roots.sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || a.id - b.id)
  return roots
})

// 分组：一级/二级
const groups = computed(() => {
  const lv2 = rawCategories.value.filter(c => c.parent_id && c.parent_id !== 0)
  return {
    parents: rawCategories.value.filter(c => !c.parent_id || c.parent_id === 0),
    children: lv2
  }
})

const filteredTree = computed(() => {
  const kw = keyword.value.trim()
  if (!kw) return categoryTree.value
  const hitIds = new Set()
  rawCategories.value.forEach(c => {
    if ((c.name || '').includes(kw)) {
      hitIds.add(c.id)
      if (c.parent_id) hitIds.add(c.parent_id)
    }
  })
  const clone = node => {
    const kids = (node.children || []).map(clone).filter(k => k !== null)
    const keep = hitIds.has(node.id) || kids.length > 0
    return keep ? { ...node, children: kids } : null
  }
  return categoryTree.value.map(clone).filter(n => n !== null)
})

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('添加分区')
const dialogForm = ref({
  id: null,
  name: '',
  parent_id: 0,
  sort_order: 0
})
const dialogFormRef = ref(null)

const rules = {
  name: [
    { required: true, message: '请输入分区名称', trigger: 'blur' }
  ]
}

const loadCategories = async () => {
  loading.value = true
  try {
    const res = await getCategoryList({})
    if (res.code === 200 || res.success) {
      const list = res.data?.list || res.data || []
      rawCategories.value = Array.isArray(list) ? list.map(d => ({
        id: d.id,
        name: d.name || '',
        parent_id: d.parent_id ?? d.parentId ?? 0,
        sort_order: d.sort_order ?? d.sortOrder ?? 0
      })) : []
    }
  } catch (error) {
    ElMessage.error('获取分区列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {}

const handleReset = () => {
  keyword.value = ''
}

const handleAdd = () => {
  dialogTitle.value = '添加分区'
  dialogForm.value = {
    id: null,
    name: '',
    parent_id: 0,
    sort_order: groups.value.parents.length
  }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑分区'
  dialogForm.value = {
    id: row.id,
    name: row.name,
    parent_id: row.parent_id || 0,
    sort_order: row.sort_order || 0
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!dialogFormRef.value) return
  await dialogFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const payload = {
        name: dialogForm.value.name,
        parent_id: dialogForm.value.parent_id || 0,
        sort_order: dialogForm.value.sort_order || 0
      }
      let res
      if (dialogForm.value.id) {
        res = await updateCategory(dialogForm.value.id, payload)
      } else {
        res = await addCategory(payload)
      }
      if (res.code === 200 || res.success) {
        ElMessage.success(dialogForm.value.id ? '更新成功' : '添加成功')
        dialogVisible.value = false
        loadCategories()
      }
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      loading.value = false
    }
  })
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该分区吗？若为一级分区其子分区将一并删除，此操作不可恢复！',
      '警告',
      { type: 'error' }
    )
    const res = await deleteCategory(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadCategories()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error('删除失败')
    }
  }
}

const thenParentName = (row) => {
  if (!row.parent_id) return ''
  const p = rawCategories.value.find(c => c.id === row.parent_id)
  return p ? p.name : ''
}

onMounted(() => {
  loadCategories()
})
</script>

<template>
  <div class="categories-page">
    <h2 class="page-title">分区管理</h2>

    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索分区名称（含子分区匹配）"
        clearable
        style="width: 300px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button @click="handleReset">
        <el-icon><Refresh /></el-icon>
        重置
      </el-button>
      <el-button type="success" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加分区
      </el-button>
    </div>

    <div class="stat-bar">
      <el-tag type="info">一级分区：{{ groups.parents.length }} 个</el-tag>
      <el-tag type="success">二级分区：{{ groups.children.length }} 个</el-tag>
      <span class="stat-tip">一级分区出现在首页与顶栏频道，二级分区隶属于一级，用于稿件分类展示</span>
    </div>

    <el-table
      v-loading="loading"
      :data="filteredTree"
      row-key="id"
      :tree-props="{ children: 'children' }"
      default-expand-all
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="name" label="分区名称" min-width="240" />
      <el-table-column label="层级" width="100">
        <template #default="{ row }">
          <el-tag :type="row.parent_id ? 'success' : 'primary'" size="small">
            {{ row.parent_id ? '二级' : '一级' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="父分区" width="160">
        <template #default="{ row }">
          {{ row.parent_id ? thenParentName(row) : '—' }}
        </template>
      </el-table-column>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column label="操作" fixed="right" width="160">
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

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
    >
      <el-form
        ref="dialogFormRef"
        :model="dialogForm"
        :rules="rules"
        label-width="90px"
      >
        <el-form-item label="分区名称" prop="name">
          <el-input v-model="dialogForm.name" placeholder="请输入分区名称" />
        </el-form-item>
        <el-form-item label="父分区">
          <el-select v-model="dialogForm.parent_id" style="width: 100%">
            <el-option :value="0" label="无（一级分区）" />
            <el-option
              v-for="p in groups.parents"
              :key="p.id"
              :value="p.id"
              :label="`${p.name}（id=${p.id}）`"
              :disabled="dialogForm.id === p.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="dialogForm.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
export default {
  methods: {}
}
</script>

<style scoped>
.categories-page {
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
  margin-bottom: 16px;
}

.stat-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 10px 16px;
  background: #f5f7fa;
  border-radius: 6px;
}

.stat-tip {
  color: #909399;
  font-size: 12px;
}
</style>