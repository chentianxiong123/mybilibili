<script setup>
import { computed, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRoleList,
  addRole,
  updateRole,
  deleteRole,
  getAllPermissions,
  getRolePermissions,
  setRolePermissions,
  getRoleTemplates,
  applyRoleTemplate
} from '../api/role'

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
const dialogTitle = ref('添加角色')
const dialogForm = ref({
  id: null,
  name: '',
  description: ''
})
const dialogFormRef = ref(null)

// 权限对话框
const permissionDialogVisible = ref(false)
const currentRoleId = ref(null)
const allPermissions = ref([])
const rolePermissions = ref([])
const roleTemplates = ref([])
const currentRoleName = ref('')
const selectedTemplateCode = ref('')
const permissionLoading = ref(false)

const permissionGroupRules = [
  { title: '运营', codes: ['operation:manage', 'search:manage', 'audit:manage', 'statistics:manage'] },
  { title: '内容审核', codes: ['review:manage', 'comment:manage'] },
  { title: 'AI', codes: ['ai:manage'] },
  { title: '媒体', codes: ['video:manage', 'category:manage', 'banner:manage', 'live:manage', 'meeting:manage'] },
  { title: '系统', codes: ['admin:manage', 'role:manage', 'security:manage'] },
  { title: '其他', codes: [] }
]

const groupedPermissions = computed(() => {
  const usedCodes = new Set()
  const groups = permissionGroupRules.map(group => {
    const permissions = group.codes.length
      ? group.codes
          .map(code => allPermissions.value.find(permission => permission.code === code))
          .filter(Boolean)
      : []
    permissions.forEach(permission => usedCodes.add(permission.code))
    return { ...group, permissions }
  })
  const otherGroup = groups.find(group => group.title === '其他')
  otherGroup.permissions = allPermissions.value.filter(permission => !usedCodes.has(permission.code))
  return groups.filter(group => group.permissions.length > 0)
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' }
  ]
}

// 加载角色列表
const loadRoles = async () => {
  loading.value = true
  try {
    const res = await getRoleList()
    if (res.code === 200 || res.success) {
      tableData.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// 打开添加对话框
const handleAdd = () => {
  dialogTitle.value = '添加角色'
  dialogForm.value = {
    id: null,
    name: '',
    description: ''
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (row) => {
  dialogTitle.value = '编辑角色'
  dialogForm.value = {
    id: row.id,
    name: row.name,
    description: row.description || ''
  }
  dialogVisible.value = true
}

// 保存角色
const handleSave = async () => {
  if (!dialogFormRef.value) return

  await dialogFormRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      let res
      if (dialogForm.value.id) {
        res = await updateRole(dialogForm.value.id, dialogForm.value)
      } else {
        res = await addRole(dialogForm.value)
      }

      if (res.code === 200 || res.success) {
        ElMessage.success(dialogForm.value.id ? '更新成功' : '添加成功')
        dialogVisible.value = false
        loadRoles()
      }
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      loading.value = false
    }
  })
}

// 删除角色
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该角色吗？此操作不可恢复！',
      '警告',
      { type: 'error' }
    )
    const res = await deleteRole(row.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadRoles()
    }
  } catch {}
}

// 打开权限设置对话框
const handlePermissions = async (row) => {
  currentRoleId.value = row.id
  currentRoleName.value = row.name
  selectedTemplateCode.value = ''
  permissionLoading.value = true
  permissionDialogVisible.value = true

  try {
    const [allRes, roleRes, templateRes] = await Promise.all([
      getAllPermissions(),
      getRolePermissions(row.id),
      getRoleTemplates()
    ])

    if (allRes.code === 200 || allRes.success) {
      allPermissions.value = allRes.data || []
    }
    if (templateRes.code === 200 || templateRes.success) {
      roleTemplates.value = templateRes.data || []
    }
    if (roleRes.code === 200 || roleRes.success) {
      const permissions = roleRes.data || []
      rolePermissions.value = permissions.map(p => p.id)
    }
  } catch (error) {
    ElMessage.error('获取权限失败')
  } finally {
    permissionLoading.value = false
  }
}

// 保存权限设置
const handleSavePermissions = async () => {
  try {
    const res = await setRolePermissions(currentRoleId.value, rolePermissions.value)
    if (res.code === 200 || res.success) {
      ElMessage.success('权限设置成功')
      permissionDialogVisible.value = false
    } else {
      ElMessage.error(`权限设置失败: ${res.message || '未知错误'}`)
    }
  } catch (error) {
    ElMessage.error(`权限设置失败: ${error.response?.data?.message || error.message || '未知错误'}`)
  }
}

const handleApplyTemplate = async () => {
  if (!selectedTemplateCode.value || !currentRoleId.value) return
  const template = roleTemplates.value.find(item => item.code === selectedTemplateCode.value)
  try {
    await ElMessageBox.confirm(
      `确认将“${template?.name || selectedTemplateCode.value}”模板套用到“${currentRoleName.value}”？`,
      '套用岗位模板',
      { type: 'warning' }
    )
    permissionLoading.value = true
    const res = await applyRoleTemplate(currentRoleId.value, selectedTemplateCode.value)
    if (res.code === 200 || res.success) {
      const roleRes = await getRolePermissions(currentRoleId.value)
      if (roleRes.code === 200 || roleRes.success) {
        rolePermissions.value = (roleRes.data || []).map(p => p.id)
      }
      ElMessage.success('岗位模板已套用')
    } else {
      ElMessage.error(res.message || '套用失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || error.message || '套用失败')
    }
  } finally {
    permissionLoading.value = false
  }
}

onMounted(() => {
  loadRoles()
})
</script>

<template>
  <div class="roles-page">
    <h2 class="page-title">角色权限管理</h2>

    <!-- 操作区域 -->
    <div class="action-bar">
      <el-button type="success" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加角色
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="角色名称" width="200" />
      <el-table-column prop="description" label="描述" min-width="300" />
      <el-table-column prop="createTime" label="创建时间" width="180">
        <template #default="{ row }">
          {{ row.createTime ? formatDate(row.createTime) : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="200">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handlePermissions(row)">
            权限设置
          </el-button>
          <el-button link type="primary" size="small" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

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
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="dialogForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="dialogForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>

    <!-- 权限设置对话框 -->
    <el-dialog
      v-model="permissionDialogVisible"
      :title="`权限设置${currentRoleName ? ` - ${currentRoleName}` : ''}`"
      width="760px"
    >
      <div v-loading="permissionLoading">
        <div class="template-bar">
          <el-select v-model="selectedTemplateCode" clearable placeholder="岗位模板" class="template-select">
            <el-option
              v-for="template in roleTemplates"
              :key="template.code"
              :label="template.name"
              :value="template.code"
            />
          </el-select>
          <el-button type="primary" :disabled="!selectedTemplateCode" @click="handleApplyTemplate">
            套用模板
          </el-button>
        </div>
        <el-checkbox-group v-model="rolePermissions">
          <div
            v-for="group in groupedPermissions"
            :key="group.title"
            class="permission-group"
          >
            <div class="group-title">{{ group.title }}</div>
            <div class="permission-list">
              <el-checkbox
                v-for="permission in group.permissions"
                :key="permission.id"
                :label="permission.id"
                :value="permission.id"
                class="permission-item"
              >
                <span>{{ permission.name }}</span>
                <span class="permission-code">{{ permission.code }}</span>
              </el-checkbox>
            </div>
          </div>
        </el-checkbox-group>
        <el-empty v-if="allPermissions.length === 0" description="暂无权限数据" />
      </div>
      <template #footer>
        <el-button @click="permissionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePermissions">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.roles-page {
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

.template-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 18px;
}

.template-select {
  width: 220px;
}

.permission-group {
  margin-bottom: 18px;
}

.group-title {
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.permission-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 16px;
}

.permission-item {
  height: 28px;
  margin: 0;
}

.permission-code {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
