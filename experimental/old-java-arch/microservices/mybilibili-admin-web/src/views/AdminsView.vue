<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminRegister, getAdminList, updateAdmin, getAdminRoles, setAdminRoles } from '../api/admin'
import {
  getRoleList,
  getAllPermissions,
  getRolePermissions,
  setRolePermissions,
  addRole,
  updateRole,
  deleteRole,
  getRoleTemplates,
  applyRoleTemplate
} from '../api/role'
import { useAdminStore } from '../stores/admin'

const adminStore = useAdminStore()
const isSuperAdmin = computed(() => adminStore.role === '超级管理员')

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/\//g, '-')
}

// ==================== 管理员 ====================
const admins = ref([])
const adminLoading = ref(false)
const addDialogVisible = ref(false)
const addForm = ref({ username: '', password: '' })
const addFormRef = ref(null)
const editDialogVisible = ref(false)
const editForm = ref({ id: null, username: '', newPassword: '' })
const editFormRef = ref(null)

const addRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }, { min: 3, max: 20, message: '3-20个字符', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6个字符', trigger: 'blur' }]
}
const editRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }, { min: 3, max: 20, message: '3-20个字符', trigger: 'blur' }]
}

// ==================== 角色 ====================
const roles = ref([])
const roleLoading = ref(false)
const roleDialogVisible = ref(false)
const roleDialogTitle = ref('添加角色')
const roleForm = ref({ id: null, name: '', description: '' })
const roleFormRef = ref(null)
const roleRules = { name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }] }

// ==================== 角色分配 ====================
const assignDialogVisible = ref(false)
const assignAdminId = ref(null)
const assignAdminName = ref('')
const allRoles = ref([])
const selectedRoleIds = ref([])

// ==================== 权限设置 ====================
const permDialogVisible = ref(false)
const permRoleId = ref(null)
const permRoleName = ref('')
const allPermissions = ref([])
const selectedPermIds = ref([])
const roleTemplates = ref([])
const selectedTemplateCode = ref('')
const permLoading = ref(false)

const permissionGroupRules = [
  { title: '运营', codes: ['operation:manage', 'search:manage', 'statistics:manage'] },
  { title: '内容审核', codes: ['review:manage', 'comment:manage'] },
  { title: 'AI', codes: ['ai:manage'] },
  { title: '媒体', codes: ['video:manage', 'category:manage', 'banner:manage', 'live:manage', 'meeting:manage'] },
  { title: '系统', codes: ['admin:manage', 'role:manage', 'security:manage', 'audit:manage'] },
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

// ==================== 加载数据 ====================
const loadAdmins = async () => {
  adminLoading.value = true
  try {
    const res = await getAdminList()
    if (res.code === 200 || res.success) admins.value = res.data || []
  } catch { ElMessage.error('获取管理员列表失败') }
  finally { adminLoading.value = false }
}

const loadRoles = async () => {
  roleLoading.value = true
  try {
    const res = await getRoleList()
    if (res.code === 200 || res.success) roles.value = res.data || []
  } catch { ElMessage.error('获取角色列表失败') }
  finally { roleLoading.value = false }
}

onMounted(() => { loadAdmins(); loadRoles() })

// ==================== 管理员操作 ====================
const handleAddAdmin = () => { addForm.value = { username: '', password: '' }; addDialogVisible.value = true }

const handleAddAdminSave = async () => {
  if (!addFormRef.value) return
  await addFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const res = await adminRegister(addForm.value)
      if (res.code === 200 || res.success) { ElMessage.success('添加成功'); addDialogVisible.value = false; loadAdmins() }
      else ElMessage.error(res.message || '添加失败')
    } catch (e) { ElMessage.error(e.response?.data?.message || '添加失败') }
  })
}

const handleEditAdmin = (row) => {
  editForm.value = { id: row.id, username: row.username, newPassword: '' }
  editDialogVisible.value = true
}

const handleEditAdminSave = async () => {
  if (!editFormRef.value) return
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const data = { username: editForm.value.username }
      if (editForm.value.newPassword) data.newPassword = editForm.value.newPassword
      const res = await updateAdmin(editForm.value.id, data)
      if (res.code === 200 || res.success) { ElMessage.success('修改成功'); editDialogVisible.value = false; loadAdmins() }
      else ElMessage.error(res.message || '修改失败')
    } catch (e) { ElMessage.error(e.response?.data?.message || '修改失败') }
  })
}

// 分配角色
const handleAssignRole = async (row) => {
  assignAdminId.value = row.id
  assignAdminName.value = row.username
  selectedRoleIds.value = (row.roles || []).map(r => r.id)
  allRoles.value = roles.value
  assignDialogVisible.value = true
}

const handleAssignRoleSave = async () => {
  try {
    const res = await setAdminRoles(assignAdminId.value, selectedRoleIds.value)
    if (res.code === 200 || res.success) { ElMessage.success('角色分配成功'); assignDialogVisible.value = false; loadAdmins() }
    else ElMessage.error(res.message || '分配失败')
  } catch (e) { ElMessage.error('分配失败') }
}

// ==================== 角色操作 ====================
const handleAddRole = () => { roleForm.value = { id: null, name: '', description: '' }; roleDialogTitle.value = '添加角色'; roleDialogVisible.value = true }

const handleEditRole = (row) => { roleForm.value = { id: row.id, name: row.name, description: row.description || '' }; roleDialogTitle.value = '编辑角色'; roleDialogVisible.value = true }

const handleSaveRole = async () => {
  if (!roleFormRef.value) return
  await roleFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const res = roleForm.value.id ? await updateRole(roleForm.value.id, roleForm.value) : await addRole(roleForm.value)
      if (res.code === 200 || res.success) { ElMessage.success(roleForm.value.id ? '更新成功' : '添加成功'); roleDialogVisible.value = false; loadRoles() }
    } catch { ElMessage.error('操作失败') }
  })
}

const handleDeleteRole = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该角色？', '警告', { type: 'error' })
    const res = await deleteRole(row.id)
    if (res.code === 200 || res.success) { ElMessage.success('删除成功'); loadRoles() }
  } catch {}
}

// 权限设置
const handlePermissions = async (row) => {
  permRoleId.value = row.id
  permRoleName.value = row.name
  selectedTemplateCode.value = ''
  permLoading.value = true
  permDialogVisible.value = true
  try {
    const [allRes, roleRes, templateRes] = await Promise.all([getAllPermissions(), getRolePermissions(row.id), getRoleTemplates()])
    if (allRes.code === 200 || allRes.success) allPermissions.value = allRes.data || []
    if (roleRes.code === 200 || roleRes.success) selectedPermIds.value = (roleRes.data || []).map(p => p.id)
    if (templateRes.code === 200 || templateRes.success) roleTemplates.value = templateRes.data || []
  } catch { ElMessage.error('获取权限失败') }
  finally { permLoading.value = false }
}

const handleSavePermissions = async () => {
  try {
    const res = await setRolePermissions(permRoleId.value, selectedPermIds.value)
    if (res.code === 200 || res.success) { ElMessage.success('权限设置成功'); permDialogVisible.value = false }
    else ElMessage.error(res.message || '设置失败')
  } catch (e) { ElMessage.error('设置失败') }
}

const handleApplyTemplate = async () => {
  if (!selectedTemplateCode.value || !permRoleId.value) return
  const template = roleTemplates.value.find(item => item.code === selectedTemplateCode.value)
  try {
    await ElMessageBox.confirm(
      `确认将“${template?.name || selectedTemplateCode.value}”模板套用到“${permRoleName.value}”？`,
      '套用岗位模板',
      { type: 'warning' }
    )
    permLoading.value = true
    const res = await applyRoleTemplate(permRoleId.value, selectedTemplateCode.value)
    if (res.code === 200 || res.success) {
      const roleRes = await getRolePermissions(permRoleId.value)
      if (roleRes.code === 200 || roleRes.success) selectedPermIds.value = (roleRes.data || []).map(p => p.id)
      ElMessage.success('岗位模板已套用')
    } else {
      ElMessage.error(res.message || '套用失败')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.response?.data?.message || e.message || '套用失败')
  } finally {
    permLoading.value = false
  }
}
</script>

<template>
  <div class="admin-role-page">
    <h2 class="page-title">管理员与角色权限</h2>

    <!-- 管理员列表 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">管理员列表</span>
          <el-button v-if="isSuperAdmin" type="primary" size="small" @click="handleAddAdmin">
            <el-icon><Plus /></el-icon> 添加管理员
          </el-button>
        </div>
      </template>

      <el-table v-loading="adminLoading" :data="admins" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column label="角色" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="r in (row.roles || [])" :key="r.id" size="small" style="margin-right:4px">{{ r.name }}</el-tag>
            <span v-if="!row.roles || row.roles.length === 0" style="color:#999;font-size:12px">未分配</span>
          </template>
        </el-table-column>
        <el-table-column label="级别" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.id === 1 || row.adminLevel === 2" type="danger" size="small">超级管理员</el-tag>
            <el-tag v-else size="small">普通管理员</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column v-if="isSuperAdmin" label="操作" fixed="right" width="220">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleAssignRole(row)">分配角色</el-button>
            <el-button link type="primary" size="small" @click="handleEditAdmin(row)">编辑</el-button>
            <span v-if="row.id === 1" style="color:#999;font-size:12px;margin-left:8px">不可操作</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 角色列表 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">角色管理</span>
          <el-button v-if="isSuperAdmin" type="success" size="small" @click="handleAddRole">
            <el-icon><Plus /></el-icon> 添加角色
          </el-button>
        </div>
      </template>

      <el-table v-loading="roleLoading" :data="roles" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="角色名称" width="160" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="createTime" label="创建时间" width="170">
          <template #default="{ row }">{{ formatDate(row.createTime) }}</template>
        </el-table-column>
        <el-table-column v-if="isSuperAdmin" label="操作" fixed="right" width="220">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handlePermissions(row)">权限设置</el-button>
            <el-button link type="primary" size="small" @click="handleEditRole(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDeleteRole(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加管理员对话框 -->
    <el-dialog v-model="addDialogVisible" title="添加管理员" width="450px">
      <el-form ref="addFormRef" :model="addForm" :rules="addRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="addForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="addForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddAdminSave">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑管理员对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑管理员" width="450px">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="editForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="editForm.newPassword" type="password" placeholder="留空则不修改" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditAdminSave">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色对话框 -->
    <el-dialog v-model="assignDialogVisible" :title="`为 ${assignAdminName} 分配角色`" width="500px">
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox v-for="r in allRoles" :key="r.id" :label="r.id" :value="r.id" style="display:block;margin:8px 0">
          <span style="font-weight:500">{{ r.name }}</span>
          <span v-if="r.description" style="color:#999;font-size:12px;margin-left:8px">{{ r.description }}</span>
        </el-checkbox>
      </el-checkbox-group>
      <el-empty v-if="allRoles.length === 0" description="暂无角色，请先添加角色" />
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAssignRoleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 角色编辑对话框 -->
    <el-dialog v-model="roleDialogVisible" :title="roleDialogTitle" width="500px">
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="roleForm.description" type="textarea" :rows="3" placeholder="请输入角色描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRole">确定</el-button>
      </template>
    </el-dialog>

    <!-- 权限设置对话框 -->
    <el-dialog v-model="permDialogVisible" :title="`权限设置 - ${permRoleName}`" width="760px">
      <div v-loading="permLoading">
        <div class="template-bar">
          <el-select v-model="selectedTemplateCode" clearable placeholder="岗位模板" class="template-select">
            <el-option
              v-for="template in roleTemplates"
              :key="template.code"
              :label="template.name"
              :value="template.code"
            />
          </el-select>
          <el-button type="primary" :disabled="!selectedTemplateCode" @click="handleApplyTemplate">套用模板</el-button>
        </div>
        <el-checkbox-group v-model="selectedPermIds">
          <div v-for="group in groupedPermissions" :key="group.title" class="permission-group">
            <div class="group-title">{{ group.title }}</div>
            <div class="permission-list">
              <el-checkbox
                v-for="p in group.permissions"
                :key="p.id"
                :label="p.id"
                :value="p.id"
                class="permission-item"
              >
                <span>{{ p.name }}</span>
                <span v-if="p.code" class="permission-code">{{ p.code }}</span>
              </el-checkbox>
            </div>
          </div>
        </el-checkbox-group>
        <el-empty v-if="allPermissions.length === 0" description="暂无权限数据" />
      </div>
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePermissions">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.admin-role-page { padding: 20px; }
.page-title { margin: 0 0 20px; font-size: 24px; font-weight: 600; color: #333; }
.section-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.card-title { font-size: 16px; font-weight: 600; }
.template-bar { display: flex; gap: 12px; margin-bottom: 18px; }
.template-select { width: 220px; }
.permission-group { margin-bottom: 18px; }
.group-title { margin-bottom: 8px; font-size: 14px; font-weight: 600; color: #303133; }
.permission-list { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px 16px; }
.permission-item { height: 28px; margin: 0; }
.permission-code { margin-left: 8px; font-size: 12px; color: #909399; }
</style>
