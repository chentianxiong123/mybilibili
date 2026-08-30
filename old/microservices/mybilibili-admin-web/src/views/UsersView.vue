<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElDialog, ElDescriptions, ElDescriptionsItem, ElTag } from 'element-plus'
import { Search, Loading } from '@element-plus/icons-vue'
import { getUserList, updateUserStatus, resetPassword, getUserById } from '../api/user'

// 表格数据
const tableData = ref([])
const loading = ref(false)
const total = ref(0)

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

// 搜索
const keyword = ref('')

// 详情弹窗相关
const detailDialogVisible = ref(false)
const userDetail = ref({})

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getUserList({
      page: currentPage.value,
      size: pageSize.value,
      keyword: keyword.value
    })
    if (res.code === 200 || res.success) {
      const list = res.data?.list || res.data || []
      
      // 简单验证status字段是否存在
      const hasStatusField = list.length > 0 && list[0].status !== undefined
      console.log(`获取用户列表成功，共${list.length}条记录，status字段${hasStatusField ? '存在' : '不存在'}`)
      
      // 如果status字段不存在，添加默认值（兼容性处理）
      const processedList = list.map(user => ({
        ...user,
        status: user.status !== undefined && user.status !== null ? user.status : 0
      }))
      
      tableData.value = processedList
      total.value = res.data?.total || res.total || 0
    }
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadUsers()
}

// 重置搜索
const handleReset = () => {
  keyword.value = ''
  currentPage.value = 1
  loadUsers()
}

// 分页改变
const handlePageChange = (page) => {
  currentPage.value = page
  loadUsers()
}

// 更新用户状态
const handleStatusChange = async (row) => {
  const newStatus = row.status == 1 ? 0 : 1
  try {
    await ElMessageBox.confirm(
      `确定要${newStatus === 1 ? '启用' : '禁用'}该用户吗？`,
      '提示',
      { type: 'warning' }
    )
    
    const res = await updateUserStatus(row.id, newStatus)
    
    if (res.code === 200 || res.success) {
      ElMessage.success('操作成功')
      loadUsers()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    // 用户取消操作，不显示错误
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 重置密码
const handleResetPassword = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入新密码', '重置密码', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '密码不能为空'
    })
    const res = await resetPassword(row.id, value)
    if (res.code === 200 || res.success) {
      ElMessage.success('重置成功')
    }
  } catch {}
}

// 打开详情弹窗
const openDetailDialog = async (userId) => {
  try {
    loading.value = true
    const res = await getUserById(userId)
    if (res.code === 200 || res.success) {
      userDetail.value = res.data || {}
      detailDialogVisible.value = true
    } else {
      ElMessage.error('获取用户详情失败')
    }
  } catch (error) {
    ElMessage.error('获取用户详情失败')
  } finally {
    loading.value = false
  }
}

// 关闭详情弹窗
const closeDetailDialog = () => {
  detailDialogVisible.value = false
  userDetail.value = {}
}

// 格式化性别显示
const formatGender = (gender) => {
  switch (gender) {
    case 1: return '男'
    case 2: return '女'
    case 0: return '保密'
    default: return '未知'
  }
}

// 加密联系方式显示
const maskContact = (value, type = 'email') => {
  if (!value) return '未设置'
  
  if (type === 'email') {
    const [username, domain] = value.split('@')
    if (!domain) return value
    const maskedUsername = username.length > 2 
      ? username.substring(0, 2) + '***' 
      : username + '***'
    return `${maskedUsername}@${domain}`
  } else if (type === 'phone') {
    if (value.length >= 7) {
      return value.substring(0, 3) + '****' + value.substring(value.length - 4)
    }
    return '****'
  }
  return value
}

onMounted(() => {
  loadUsers()
})
</script>

<template>
  <div class="users-page">
    <h2 class="page-title">用户管理</h2>

    <!-- 搜索区域 -->
    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索用户名、昵称"
        clearable
        style="width: 300px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      stripe
    >
      <el-table-column prop="id" label="ID" width="70" align="center" />
      <el-table-column label="用户" width="180">
        <template #default="{ row }">
          <div style="display: flex; align-items: center; gap: 12px">
            <el-avatar
              :size="40"
              :src="row.avatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'"
            />
            <div>
              <div style="font-weight: 600">{{ row.nickname }}</div>
              <div style="font-size: 12px; color: #666">@{{ row.username }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="level" label="等级" width="80" align="center">
        <template #default="{ row }">
          <el-tag type="info" size="small">Lv.{{ row.level || 1 }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="粉丝/关注" width="130" align="center">
        <template #default="{ row }">
          <div style="font-size: 13px">
            <div>粉丝: {{ row.followerCount || 0 }}</div>
            <div>关注: {{ row.followingCount || 0 }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="稿件" width="100" align="center">
        <template #default="{ row }">
          <div style="font-size: 13px">
            {{ row.manuscriptCount || 0 }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="email" label="联系方式" width="180">
        <template #default="{ row }">
          <div style="font-size: 13px">
            <div>{{ maskContact(row.email, 'email') }}</div>
            <div>{{ maskContact(row.phone, 'phone') }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.status == 1" type="success" size="small">正常</el-tag>
          <el-tag v-else-if="row.status == 0" type="info" size="small">禁用</el-tag>
          <el-tag v-else type="warning" size="small">未知</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="180" align="center">
        <template #default="{ row }">
          <el-button
            type="primary"
            size="small"
            @click="openDetailDialog(row.id)"
          >
            详情
          </el-button>
          <el-button
            type="warning"
            size="small"
            @click="handleStatusChange(row)"
          >
            {{ row.status == 1 ? '禁用' : '启用' }}
          </el-button>
          <el-button
            type="danger"
            size="small"
            @click="handleResetPassword(row)"
          >
            重置密码
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
        @size-change="loadUsers"
      />
    </div>

    <!-- 用户详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="用户详细信息"
      width="800px"
      :before-close="closeDetailDialog"
    >
      <div v-if="userDetail.id">
        <div style="display: flex; align-items: flex-start; gap: 24px; margin-bottom: 24px">
          <el-avatar
            :size="80"
            :src="userDetail.avatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'"
          />
          <div style="flex: 1">
            <h3 style="margin: 0 0 8px">{{ userDetail.nickname }}</h3>
            <div style="color: #666; margin-bottom: 12px">@{{ userDetail.username }}</div>
            <div style="display: flex; gap: 16px; flex-wrap: wrap">
              <el-tag type="info">Lv.{{ userDetail.level || 1 }}</el-tag>
              <el-tag :type="userDetail.status == 1 ? 'success' : 'info'">
                {{ userDetail.status == 1 ? '正常' : '禁用' }}
              </el-tag>
              <el-tag>{{ formatGender(userDetail.gender) }}</el-tag>
            </div>
          </div>
        </div>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户ID">{{ userDetail.id }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ maskContact(userDetail.email, 'email') }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ maskContact(userDetail.phone, 'phone') }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ formatGender(userDetail.gender) }}</el-descriptions-item>
          <el-descriptions-item label="生日">{{ userDetail.birthdate || '未设置' }}</el-descriptions-item>
        </el-descriptions>

        <div style="margin-top: 24px">
          <h4 style="margin-bottom: 12px">社交数据</h4>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="粉丝数">{{ userDetail.followerCount || 0 }}</el-descriptions-item>
            <el-descriptions-item label="关注数">{{ userDetail.followingCount || 0 }}</el-descriptions-item>
            <el-descriptions-item label="视频数">{{ userDetail.manuscriptCount || 0 }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <div style="margin-top: 24px">
          <h4 style="margin-bottom: 12px">个人简介</h4>
          <div style="padding: 16px; background: #f5f7fa; border-radius: 4px">
            {{ userDetail.bio || '暂无个人简介' }}
          </div>
        </div>

        <div style="margin-top: 24px">
          <h4 style="margin-bottom: 12px">个性签名</h4>
          <div style="padding: 16px; background: #f5f7fa; border-radius: 4px">
            {{ userDetail.signature || '暂无签名' }}
          </div>
        </div>

        <div style="margin-top: 24px">
          <h4 style="margin-bottom: 12px">个人公告</h4>
          <div style="padding: 16px; background: #f5f7fa; border-radius: 4px">
            {{ userDetail.announcement || '暂无公告' }}
          </div>
        </div>


      </div>
      <div v-else style="text-align: center; padding: 40px">
        <el-icon size="40" color="#909399"><Loading /></el-icon>
        <div style="margin-top: 16px; color: #909399">加载用户信息中...</div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="closeDetailDialog">关闭</el-button>
          <el-button
            type="warning"
            @click="handleStatusChange(userDetail)"
          >
            {{ userDetail.status == 1 ? '禁用用户' : '启用用户' }}
          </el-button>
          <el-button
            type="danger"
            @click="handleResetPassword(userDetail)"
          >
            重置密码
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.users-page {
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