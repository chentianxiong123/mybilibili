<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

// 判断是否是登录页面
const isLoginPage = computed(() => route.path === '/login')

// 菜单项
const menuItems = [
  { path: '/dashboard', icon: 'DataBoard', title: '数据概览' },
  { path: '/users', icon: 'User', title: '用户管理' },
  { path: '/manuscripts', icon: 'Document', title: '稿件管理' },
  { path: '/video-process', icon: 'VideoPlay', title: '视频处理' },
  { path: '/prohibited-words', icon: 'Warning', title: '违禁词管理' },
  { path: '/content-review', icon: 'DocumentChecked', title: '内容审核中心' },
  { path: '/categories', icon: 'Folder', title: '分类管理' },
  { path: '/banner-images', icon: 'Picture', title: '图片管理' },
  { path: '/index-manager', icon: 'DataLine', title: '索引管理' },
  { path: '/admins', icon: 'Lock', title: '管理员管理' }
]

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/users')) return '/users'
  if (path.startsWith('/manuscripts')) return '/manuscripts'
  if (path.startsWith('/comments')) return '/comments'
  if (path.startsWith('/video-process')) return '/video-process'
  if (path.startsWith('/prohibited-words')) return '/prohibited-words'
  if (path.startsWith('/content-review')) return '/content-review'
  if (path.startsWith('/categories')) return '/categories'
  if (path.startsWith('/banner-images')) return '/banner-images'
  if (path.startsWith('/index-manager')) return '/index-manager'
  if (path.startsWith('/admins')) return '/admins'
  return path
})

const handleCommand = (command) => {
  const token = localStorage.getItem('admin_token')
  if (command === 'logout') {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    router.push('/login')
  }
}
</script>

<template>
  <div class="app-container" v-if="!isLoginPage">
    <el-container class="layout-container">
      <!-- 侧边栏 -->
      <el-aside width="220px" class="aside">
        <div class="logo">
          <el-icon :size="28" color="#00aeec"><Monitor /></el-icon>
          <span class="logo-text">管理后台</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          class="el-menu-vertical"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409eff"
        >
          <el-menu-item
            v-for="item in menuItems"
            :key="item.path"
            :index="item.path"
          >
            <el-icon><component :is="ElementPlusIconsVue[item.icon]" /></el-icon>
            <span>{{ item.title }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <!-- 顶部导航栏 -->
        <el-header class="header">
          <div class="header-left">
            <el-icon :size="20" class="breadcrumb-icon"><Fold /></el-icon>
          </div>
          <div class="header-right">
            <el-dropdown trigger="click" @command="handleCommand">
              <div class="user-info">
                <el-avatar :size="32" :icon="ElementPlusIconsVue.UserFilled" />
                <span class="username">管理员</span>
                <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="logout">
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 主内容 -->
        <el-main class="main">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>

  <!-- 登录页面 -->
  <div v-else class="login-page">
    <router-view />
  </div>
</template>

<style scoped>
.app-container {
  height: 100vh;
}

.layout-container {
  height: 100%;
}

.aside {
  background-color: #304156;
  overflow: hidden;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60px;
  background-color: #2b3a4a;
}

.logo-text {
  margin-left: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.el-menu-vertical {
  border-right: none;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
}

.breadcrumb-icon {
  cursor: pointer;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #333;
}

.dropdown-icon {
  font-size: 12px;
}

.main {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}

.login-page {
  height: 100vh;
}
</style>