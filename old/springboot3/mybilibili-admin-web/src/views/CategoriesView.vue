<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCategoryList,
  addCategory,
  updateCategory,
  deleteCategory
} from '../api/category'

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

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
const dialogTitle = ref('添加分区')
const dialogForm = ref({
  id: null,
  name: ''
})
const dialogFormRef = ref(null)

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入分区名称', trigger: 'blur' }
  ]
}

// 加载分区列表
const loadCategories = async () => {
  loading.value = true
  try {
    const res = await getCategoryList({
      page: currentPage.value,
      size: pageSize.value,
      keyword: keyword.value
    })
    if (res.code === 200 || res.success) {
      tableData.value = res.data?.list || res.data || []
      total.value = res.data?.total || res.total || 0
    }
  } catch (error) {
    ElMessage.error('获取分区列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadCategories()
}

// 重置搜索
const handleReset = () => {
  keyword.value = ''
  currentPage.value = 1
  loadCategories()
}

// 分页改变
const handlePageChange = (page) => {
  currentPage.value = page
  loadCategories()
}

// 打开添加对话框
const handleAdd = () => {
  dialogTitle.value = '添加分区'
  dialogForm.value = {
    id: null,
    name: ''
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (row) => {
  dialogTitle.value = '编辑分区'
  dialogForm.value = {
    id: row.id,
    name: row.name
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
        // 编辑
        res = await updateCategory(dialogForm.value.id, {
          name: dialogForm.value.name
        })
      } else {
        // 添加
        res = await addCategory({
          name: dialogForm.value.name
        })
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

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该分区吗？此操作不可恢复！',
      '警告',
      { type: 'error' }
    )
    const res = await deleteCategory(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadCategories()
    }
  } catch {}
}

onMounted(() => {
  loadCategories()
})
</script>

<template>
  <div class="categories-page">
    <h2 class="page-title">分区管理</h2>

    <!-- 搜索和操作区域 -->
    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索分区名称"
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
        添加分区
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="分区名称" width="200" />
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
        @size-change="loadCategories"
      />
    </div>

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
        label-width="80px"
      >
        <el-form-item label="分区名称" prop="name">
          <el-input v-model="dialogForm.name" placeholder="请输入分区名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

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
  margin-bottom: 20px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>