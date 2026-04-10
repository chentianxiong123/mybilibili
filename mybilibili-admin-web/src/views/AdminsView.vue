<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminRegister, getAdminList } from '../api/admin'

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).replace(/\//g, '-')
}

// 表格数据
const tableData = ref([])
const loading = ref(false)

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('添加管理员')
const dialogForm = ref({
  username: '',
  password: ''
})
const dialogFormRef = ref(null)

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3-20个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6个字符', trigger: 'blur' }
  ]
}

// 加载管理员列表
const loadAdmins = async () => {
  loading.value = true
  try {
    const res = await getAdminList()
    if (res.code === 200 || res.success) {
      tableData.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('获取管理员列表失败')
  } finally {
    loading.value = false
  }
}

// 打开添加对话框
const handleAdd = () => {
  dialogTitle.value = '添加管理员'
  dialogForm.value = { username: '', password: '' }
  dialogVisible.value = true
}

// 保存
const handleSave = async () => {
  if (!dialogFormRef.value) return

  await dialogFormRef.value.validate(async (valid) => {
    if (!valid) return

    try {
      const res = await adminRegister(dialogForm.value)
      if (res.code === 200 || res.success) {
        ElMessage.success('添加成功')
        dialogVisible.value = false
        loadAdmins()
      } else {
        ElMessage.error(res.message || '添加失败')
      }
    } catch (error) {
      ElMessage.error('添加失败')
    }
  })
}

onMounted(() => {
  loadAdmins()
})
</script>

<template>
  <div class="admins-page">
    <h2 class="page-title">管理员管理</h2>

    <!-- 操作区域 -->
    <div class="action-bar">
      <el-button type="success" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加管理员
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" min-width="200" />
      <el-table-column prop="updatedAt" label="最后登录时间" width="180">
        <template #default="{ row }">
          {{ row.updatedAt ? formatDate(row.updatedAt) : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ row.createdAt ? formatDate(row.createdAt) : '-' }}
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加管理员对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="450px"
    >
      <el-form
        ref="dialogFormRef"
        :model="dialogForm"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="dialogForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="dialogForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
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
.admins-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.action-bar {
  margin-bottom: 20px;
}
</style>
