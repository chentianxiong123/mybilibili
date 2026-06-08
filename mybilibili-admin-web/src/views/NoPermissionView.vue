<template>
  <div class="no-permission-page">
    <el-empty description="当前账号暂无可访问的后台功能">
      <el-button type="primary" @click="handleLogout">退出登录</el-button>
    </el-empty>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAdminStore } from '../stores/admin'

const router = useRouter()
const adminStore = useAdminStore()

const handleLogout = () => {
  adminStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.no-permission-page {
  min-height: calc(100vh - 100px);
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
