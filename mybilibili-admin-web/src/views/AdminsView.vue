<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminRegister, getAdminList, getAdminRoles, setAdminRoles } from '../api/admin'
import { getRoleList } from '../api/role'

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

// 角色对话框
const roleDialogVisible = ref(false)
const currentAdminId = ref(null)
const allRoles = ref([])
const adminRoles = ref([])
const roleLoading = ref(false)

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
      const admins = res.data || []
      console.log('管理员列表原始数据:', admins)
      // 确保每个管理员都有roleNames字段
      tableData.value = admins.map(admin => {
        console.log('单个管理员数据:', admin)
        return {
          ...admin,
          roleNames: admin.roleNames || (admin.roles ? admin.roles.map(r => r.name).join(', ') : '无角色')
        }
      })
      console.log('处理后表格数据:', tableData.value)
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
  dialogForm.value = {
    username: '',
    password: ''
  }
  dialogVisible.value = true
}

// 保存管理员
const handleSave = async () => {
  if (!dialogFormRef.value) return

  await dialogFormRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await adminRegister(dialogForm.value)
      if (res.code === 200 || res.success) {
        ElMessage.success('添加成功')
        dialogVisible.value = false
        loadAdmins()
      }
    } catch (error) {
      ElMessage.error('添加失败')
    } finally {
      loading.value = false
    }
  })
}

// 打开角色设置对话框
const handleRoles = async (row) => {
  currentAdminId.value = row.id
  roleLoading.value = true
  roleDialogVisible.value = true

  try {
    // 同时获取所有角色和管理员已有角色
    const [allRolesRes, adminRolesRes] = await Promise.all([
      getRoleList(),
      getAdminRoles(row.id)
    ])

    if (allRolesRes.code === 200 || allRolesRes.success) {
      allRoles.value = allRolesRes.data || []
    }
    if (adminRolesRes.code === 200 || adminRolesRes.success) {
      // 管理员角色可能是对象数组，需要提取ID
      const roles = adminRolesRes.data || []
      adminRoles.value = roles.map(role => role.id)
    }
  } catch (error) {
    console.error('获取角色失败:', error)
    ElMessage.error('获取角色失败')
  } finally {
    roleLoading.value = false
  }
}

// 保存角色设置
const handleSaveRoles = async () => {
  try {
    console.log('设置管理员角色参数:', {
      adminId: currentAdminId.value,
      roleIds: adminRoles.value
    })
    const res = await setAdminRoles(currentAdminId.value, adminRoles.value)
    console.log('设置管理员角色响应:', res)
    if (res.code === 200 || res.success) {
      ElMessage.success('角色设置成功')
      roleDialogVisible.value = false
      // 重新加载管理员列表以更新角色显示
      loadAdmins()
    } else {
      ElMessage.error(`角色设置失败: ${res.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('设置管理员角色错误:', error)
    ElMessage.error(`角色设置失败: ${error.response?.data?.message || error.message || '未知错误'}`)
  }
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
      <el-table-column prop="username" label="用户名" width="200" />
      <el-table-column prop="roleNames" label="角色" min-width="200" />
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
      <el-table-column label="操作" fixed="right" width="150">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleRoles(row)">
            设置角色
          </el-button>
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

    <!-- 角色设置对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      title="设置角色"
      width="500px"
    >
      <div v-loading="roleLoading">
        <el-checkbox-group v-model="adminRoles">
          <el-checkbox
            v-for="role in allRoles"
            :key="role.id"
            :label="role.id"
            :value="role.id"
            style="display: block; margin: 8px 0"
          >
            {{ role.name }}
          </el-checkbox>
        </el-checkbox-group>
        <el-empty v-if="allRoles.length === 0" description="暂无角色数据" />
      </div>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRoles">保存</el-button>
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