<script setup lang="ts">
import { safeStorage } from '@/utils/safeStorage'
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Avatar, Brush, Clock, House, User, UserFilled } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

// 当前登录用户信息
const currentUser = ref(JSON.parse(safeStorage.getItem('user')) || null)

// 导航选项
const navItems = [
  { name: '首页', icon: 'Home', path: '/personal-center/home' },
  { name: '我的信息', icon: 'User', path: '/personal-center/info' },
  { name: '我的头像', icon: 'Avatar', path: '/personal-center/avatar' },
  { name: '登录记录', icon: 'Clock', path: '/personal-center/login-logs' },
  { name: '个人空间', icon: 'UserFilled', path: 'space', special: true },
  { name: '创作中心', icon: 'Brush', path: '/platform/home', special: true }
]

const iconMap = {
  Avatar,
  Brush,
  Clock,
  Home: House,
  User,
  UserFilled
}

// 当前活跃的导航选项
const activeNav = computed(() => {
  return navItems.find(item => route.path === item.path)?.name || '首页'
})

// 处理导航点击
const handleNavClick = (item) => {
  if (item.special) {
    // 特殊处理：个人空间
    if (item.name === '个人空间') {
      if (currentUser.value && currentUser.value.id) {
        // 打开新标签页到用户主页
        const url = `${window.location.origin}/profile/${currentUser.value.id}/home`
        window.open(url, '_blank')
      } else {
        // 未登录：回首页触发登录弹窗
        window.location.href = `${window.location.origin}/`
      }
    } 
    // 特殊处理：创作中心
    else if (item.name === '创作中心') {
      // 打开新标签页到创作中心
      const url = `${window.location.origin}/platform/home`
      window.open(url, '_blank')
    }
    // 其他特殊项
    else {
      router.push(item.path)
    }
  } else {
    router.push(item.path)
  }
}
</script>

<template>
  <div class="personal-center">
    <!-- 主内容区域 -->
    <div class="content-container">
      <!-- 左侧侧部导航栏 -->
      <div class="side-nav">
        <div class="nav-header">
          <h3>个人中心</h3>
        </div>
        <div class="nav-list">
          <div
            v-for="item in navItems"
            :key="item.name"
            :class="['nav-item', { active: activeNav === item.name }]"
            @click="handleNavClick(item)"
          >
            <el-icon class="nav-icon"><component :is="iconMap[item.icon]" /></el-icon>
            <span class="nav-text">{{ item.name }}</span>
          </div>
        </div>
      </div>

      <!-- 右侧主内容 -->
      <div class="main-content">
        <div class="content-header">
          <h2>{{ activeNav }}</h2>
        </div>
        <div class="content-body">
          <NuxtPage />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 个人中心整体样式 */
.personal-center {
  width: 100%;
  min-height: 100vh;
  background-color: #fff;
  display: flex;
  flex-direction: column;
}

/* 主内容容器（对齐 account.bilibili.com：980px 居中 + 左侧 150px 侧栏 + 右侧内容，无间隙） */
.content-container {
  flex: 1;
  width: 980px;
  margin: 10px auto 100px;
  padding: 0;
  display: flex;
  gap: 0;
  box-sizing: content-box;
  background-color: #fafafa;
  border: 1px solid #e1e2e5;
  border-radius: 4px;
  box-shadow: rgba(0, 0, 0, 0.14) 0px 2px 4px 0px;
  overflow: hidden;
}

/* 左侧导航栏（对齐 B 站 security-left：150px 宽、无卡片阴影） */
.side-nav {
  flex: 0 0 auto;
  width: 150px;
  min-width: 150px;
  background-color: transparent;
  border-radius: 0;
  box-shadow: none;
  padding: 0;
}

.nav-header {
  padding: 0;
  border-bottom: 0;
  margin-bottom: 0;
  height: 50px;
}

.nav-header h3 {
  height: 50px;
  font-size: 16px;
  font-weight: 400;
  line-height: 50px;
  color: #99a2aa;
  text-align: center;
  margin: 0;
}

.nav-list {
  display: flex;
  flex-direction: column;
}

.nav-item {
  display: block;
  height: 48px;
  padding: 0;
  overflow: hidden;
  cursor: pointer;
  transition: background-color 0.3s ease;
  color: #717171;
  font-size: 14px;
  line-height: 48px;
}

.nav-item:hover {
  background-color: #00a1d7;
  color: #fff;
}

.nav-item.active {
  background-color: #00a1d7;
  color: #fff;
  border-left: 0;
}

.nav-icon {
  display: inline-block;
  width: 18px;
  height: 36px;
  margin-left: 16px;
  font-size: 12px;
  line-height: 48px;
  vertical-align: middle;
}

.nav-text {
  display: inline;
  margin-left: 12px;
  font-size: 14px;
  line-height: 48px;
  vertical-align: middle;
}

/* 右侧主内容（对齐 B 站：无圆角无阴影白底面板） */
.main-content {
  flex: 1;
  min-width: 0;
  background-color: #fff;
  border-radius: 0;
  box-shadow: none;
  padding: 20px;
}

.content-header {
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.content-header h2 {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.content-body {
  min-height: 400px;
  width: 100%;
}

/* 响应式设计 */

</style>
