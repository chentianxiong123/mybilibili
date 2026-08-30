<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminLoginLogApi } from '../api/securitySettings.js'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 搜索条件
const searchForm = reactive({
  ip: '',
  userId: null,
  status: null,
  startTime: '',
  endTime: ''
})

// 加载数据
const loadData = () => {
  loading.value = true
  const params = {
    page: page.value,
    size: size.value,
    ...searchForm
  }
  adminLoginLogApi.getLoginLogs(params).then(res => {
    if (res.code === 200) {
      tableData.value = res.data.list
      total.value = res.data.total
    }
  }).finally(() => { loading.value = false })
}

loadData()

// 搜索
const handleSearch = () => {
  page.value = 1
  loadData()
}

// 重置
const handleReset = () => {
  searchForm.ip = ''
  searchForm.userId = null
  searchForm.status = null
  searchForm.startTime = ''
  searchForm.endTime = ''
  page.value = 1
  loadData()
}

// 分页
const handlePageChange = (p) => {
  page.value = p
  loadData()
}

const handleSizeChange = (s) => {
  size.value = s
  loadData()
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

// 格式化状态
const formatStatus = (status) => {
  return status === 1 ? '成功' : '失败'
}
</script>

<template>
  <div class="login-logs-container">
    <div class="page-header">
      <h2>登录日志</h2>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-input v-model="searchForm.ip" placeholder="IP模糊搜索" style="width: 200px; margin-right: 10px;" clearable />
      <el-input v-model="searchForm.userId" placeholder="用户ID" style="width: 120px; margin-right: 10px;" type="number" clearable />
      <el-select v-model="searchForm.status" placeholder="状态" style="width: 100px; margin-right: 10px;" clearable>
        <el-option label="成功" :value="1" />
        <el-option label="失败" :value="0" />
      </el-select>
      <el-date-picker v-model="searchForm.startTime" type="datetime" placeholder="开始时间" style="margin-right: 10px;" value-format="YYYY-MM-DD HH:mm:ss" />
      <el-date-picker v-model="searchForm.endTime" type="datetime" placeholder="结束时间" style="margin-right: 10px;" value-format="YYYY-MM-DD HH:mm:ss" />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="userId" label="用户ID" width="100" />
      <el-table-column prop="ip" label="登录IP" width="150" />
      <el-table-column prop="location" label="登录地点" width="150" />
      <el-table-column prop="loginTime" label="登录时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.loginTime) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ formatStatus(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="userAgent" label="User-Agent" show-overflow-tooltip />
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<style scoped>
.login-logs-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.search-bar {
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>