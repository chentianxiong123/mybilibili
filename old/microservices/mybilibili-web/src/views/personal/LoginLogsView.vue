<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { userApi } from '../../api/index.js'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const size = ref(10)

const loadData = () => {
  loading.value = true
  userApi.getLoginLogs(page.value, size.value).then(res => {
    if (res.code === 200) {
      tableData.value = res.data.list
      total.value = res.data.total
    } else {
      ElMessage.error(res.message || '获取登录记录失败')
    }
  }).finally(() => { loading.value = false })
}

onMounted(() => {
  loadData()
})

const handlePageChange = (p) => {
  page.value = p
  loadData()
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}
</script>

<template>
  <div class="login-logs">
    <div class="page-title">登录记录</div>

    <el-table :data="tableData" v-loading="loading" stripe style="width: 100%">
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
            {{ row.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="userAgent" label="设备信息" show-overflow-tooltip />
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="size"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<style scoped>
.login-logs {
  padding: 20px;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>