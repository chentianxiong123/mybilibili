<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminRegister, getAdminList, updateAdmin } from '../api/admin'
import { useAdminStore } from '../stores/admin'

const adminStore = useAdminStore()

const isSuperAdmin = computed(() => {
  const user = adminStore.userInfo
  return user && (user.id === 1 || user.adminLevel === 0)
})

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

// 添加对话框
const addDialogVisible = ref(false)
const addForm = ref({ username: '', password: '' })
const addFormRef = ref(null)

// 编辑对话框
const editDialogVisible = ref(false)
const editForm = ref({ id: null, username: '', adminLevel: 1, newPassword: '' })
const editFormRef = ref(null)

// 表单验证规则
const addRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3-20个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6个字符', trigger: 'blur' }
  ]
}

const editRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3-20个字符', trigger: 'blur' }
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
  addForm.value = { username: '', password: '' }
  addDialogVisible.value = true
}

// 添加管理员
const handleAddSave = async () => {
  if (!addFormRef.value) return
  await addFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const res = await adminRegister(addForm.value)
      if (res.code === 200 || res.success) {
        ElMessage.success('添加成功')
        addDialogVisible.value = false
        loadAdmins()
      } else {
        ElMessage.error(res.message || '添加失败')
      }
    } catch (error) {
      ElMessage.error(error.response?.data?.message || '添加失败')
    }
  })
}

// 打开编辑对话框
const handleEdit = (row) => {
  editForm.value = {
    id: row.id,
    username: row.username,
    adminLevel: row.adminLevel || 1,
    newPassword: ''
  }
  editDialogVisible.value = true
}

// 保存编辑
const handleEditSave = async () => {
  if (!editFormRef.value) return
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const data = {
        username: editForm.value.username,
        adminLevel: editForm.value.adminLevel
      }
      if (editForm.value.newPassword) {
        data.newPassword = editForm.value.newPassword
      }
      const res = await updateAdmin(editForm.value.id, data)
      if (res.code === 200 || res.success) {
        ElMessage.success('修改成功')
        editDialogVisible.value = false
        loadAdmins()
      } else {
        ElMessage.error(res.message || '修改失败')
      }
    } catch (error) {
      ElMessage.error(error.response?.data?.message || '修改失败')
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
      <el-button v-if="isSuperAdmin" type="success" @click="handleAdd">
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
      <el-table-column prop="username" label="用户名" min-width="150" />
      <el-table-column label="级别" width="120">
        <template #default="{ row }">
          <el-tag v-if="row.id === 1" type="danger">超级管理员</el-tag>
          <el-tag v-else-if="row.adminLevel === 0" type="warning">超级管理员</el-tag>
          <el-tag v-else>普通管理员</el-tag>
        </template>
      </el-table-column>
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
      <el-table-column v-if="isSuperAdmin" label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.id !== 1" type="primary" link size="small" @click="handleEdit(row)">
            编辑
          </el-button>
          <span v-else style="color: #999; font-size: 12px;">不可操作</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加管理员对话框 -->
    <el-dialog
      v-model="addDialogVisible"
      title="添加管理员"
      width="450px"
    >
      <el-form
        ref="addFormRef"
        :model="addForm"
        :rules="addRules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="addForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="addForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddSave">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑管理员对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑管理员"
      width="450px"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="editForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input
            v-model="editForm.newPassword"
            type="password"
            placeholder="留空则不修改密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSave">确定</el-button>
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
