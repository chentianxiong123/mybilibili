<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowDown,
  Connection,
  Cpu,
  DataAnalysis,
  DataBoard,
  DataLine,
  Document,
  DocumentChecked,
  Expand,
  Fold,
  Folder,
  Headset,
  List,
  Lock,
  Message,
  Monitor,
  Operation,
  Picture,
  Setting,
  SwitchButton,
  Tickets,
  User,
  UserFilled,
  VideoPlay,
  Warning
} from '@element-plus/icons-vue'
import { useAdminStore } from './stores/admin'
import AdminAiFloatingButton from './components/AdminAiFloatingButton.vue'
import AdminAiChatPanel from './components/AdminAiChatPanel.vue'

const route = useRoute()
const router = useRouter()
const adminStore = useAdminStore()
const isSuperAdmin = computed(() => adminStore.role === '超级管理员')
const canUseAdminAssistant = computed(() => adminStore.hasPermission('ai:manage'))
const isCollapse = ref(false)
const showAiAssistant = ref(false)

const isLoginPage = computed(() => route.path === '/login')

const iconMap = {
  Connection,
  Cpu,
  DataAnalysis,
  DataBoard,
  DataLine,
  Document,
  DocumentChecked,
  Folder,
  Headset,
  List,
  Lock,
  Message,
  Monitor,
  Operation,
  Picture,
  Setting,
  Tickets,
  User,
  VideoPlay,
  Warning
}

const allMenuItems = [
  {
    type: 'group',
    icon: 'Operation',
    title: '运营板块',
    children: [
      { path: '/operation-tasks', icon: 'List', title: '任务中心', permission: 'operation:manage' },
      { path: '/support-tickets', icon: 'Message', title: '工单中心', permission: 'operation:manage' },
      { path: '/index-manager', icon: 'DataLine', title: '索引管理', permission: 'search:manage' },
      { path: '/recommend-config', icon: 'DataAnalysis', title: '推荐配置', permission: 'search:manage' }
    ]
  },
  {
    type: 'group',
    icon: 'DocumentChecked',
    title: '审核治理',
    children: [
      { path: '/manuscripts', icon: 'Document', title: '稿件管理', permission: 'review:manage' },
      { path: '/content-review', icon: 'DocumentChecked', title: '内容审核中心', permission: 'review:manage' },
      { path: '/prohibited-words', icon: 'Warning', title: '违禁词与安全设置', permission: 'comment:manage' }
    ]
  },
  {
    type: 'group',
    icon: 'Cpu',
    title: 'AI 管理',
    children: [
      { path: '/ai-usage', icon: 'DataAnalysis', title: 'AI 用量统计', permission: 'ai:manage' },
      { path: '/ai-skills', icon: 'Cpu', title: 'AI 技能管理', permission: 'ai:manage' },
      { path: '/api-management', icon: 'Setting', title: 'AI 渠道管理', permission: 'ai:manage' },
      { path: '/customer-chat', icon: 'Headset', title: '客服会话', permission: 'ai:manage' }
    ]
  },
  {
    type: 'group',
    icon: 'VideoPlay',
    title: '媒体管理',
    children: [
      { path: '/categories', icon: 'Folder', title: '分类管理', permission: 'category:manage' },
      { path: '/banner-images', icon: 'Picture', title: '图片管理', permission: 'banner:manage' },
      { path: '/subtitles', icon: 'Tickets', title: '字幕管理', permission: 'video:manage' },
      { path: '/live-rooms', icon: 'Connection', title: '直播管理', permission: 'live:manage' },
      { path: '/meeting-admin', icon: 'Monitor', title: '会议管理', permission: 'meeting:manage' }
    ]
  },
  {
    type: 'group',
    icon: 'Lock',
    title: '系统管理',
    children: [
      { path: '/dashboard', icon: 'DataBoard', title: '数据概览', permission: 'statistics:manage' },
      { path: '/users', icon: 'User', title: '用户管理', permission: 'user:manage' },
      { path: '/login-logs', icon: 'List', title: '登录日志', permission: 'security:manage' },
      { path: '/audit-logs', icon: 'Tickets', title: '审计日志', permission: 'audit:manage' },
      { path: '/admins', icon: 'Lock', title: '管理员与角色权限', permission: 'role:manage', superAdminOnly: true }
    ]
  }
]

const filterMenuNode = (node) => {
  if (node.type === 'group') {
    const children = (node.children || [])
      .filter(child => (!child.superAdminOnly || isSuperAdmin.value) && adminStore.hasPermission(child.permission))
    return children.length ? { ...node, children } : null
  }
  return (!node.superAdminOnly || isSuperAdmin.value) && adminStore.hasPermission(node.permission) ? node : null
}

const menuItems = computed(() => allMenuItems.map(filterMenuNode).filter(Boolean))

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/users')) return '/users'
  if (path.startsWith('/manuscripts')) return '/manuscripts'
  if (path.startsWith('/comments')) return '/comments'
  if (path.startsWith('/operation-tasks')) return '/operation-tasks'
  if (path.startsWith('/audit-logs')) return '/audit-logs'
  if (path.startsWith('/prohibited-words')) return '/prohibited-words'
  if (path.startsWith('/content-review')) return '/content-review'
  if (path.startsWith('/categories')) return '/categories'
  if (path.startsWith('/banner-images')) return '/banner-images'
  if (path.startsWith('/index-manager')) return '/index-manager'
  if (path.startsWith('/recommend-config')) return '/recommend-config'
  if (path.startsWith('/admins')) return '/admins'
  if (path.startsWith('/api-management')) return '/api-management'
  if (path.startsWith('/ai-skills')) return '/ai-skills'
  if (path.startsWith('/ai-usage')) return '/ai-usage'
  if (path.startsWith('/support-tickets')) return '/support-tickets'
  if (path.startsWith('/live-rooms')) return '/live-rooms'
  if (path.startsWith('/meeting-admin')) return '/meeting-admin'
  if (path.startsWith('/login-logs')) return '/login-logs'
  if (path.startsWith('/customer-chat')) return '/customer-chat'
  return path
})

const handleCommand = (command) => {
  if (command === 'logout') {
    adminStore.logout()
    router.push('/login')
  }
}
</script>

<template>
  <div class="app-container" v-if="!isLoginPage">
    <el-container class="layout-container">
      <!-- 侧边栏 -->
      <el-aside :width="isCollapse ? '64px' : '220px'" class="aside">
        <div class="logo">
          <el-icon :size="28" color="#00aeec"><Monitor /></el-icon>
          <span v-if="!isCollapse" class="logo-text">管理后台</span>
        </div>
        <el-scrollbar class="menu-scrollbar">
          <el-menu
            :default-active="activeMenu"
            :collapse="isCollapse"
            router
            class="el-menu-vertical"
            background-color="#304156"
            text-color="#bfcbd9"
            active-text-color="#409eff"
          >
            <template v-for="item in menuItems" :key="item.type === 'group' ? item.title : item.path">
              <el-sub-menu v-if="item.type === 'group'" :index="item.title">
                <template #title>
                  <el-icon><component :is="iconMap[item.icon]" /></el-icon>
                  <span>{{ item.title }}</span>
                </template>
                <el-menu-item
                  v-for="child in item.children"
                  :key="child.path"
                  :index="child.path"
                >
                  <el-icon><component :is="iconMap[child.icon]" /></el-icon>
                  <template #title>{{ child.title }}</template>
                </el-menu-item>
              </el-sub-menu>
              <el-menu-item v-else :index="item.path">
                <el-icon><component :is="iconMap[item.icon]" /></el-icon>
                <template #title>{{ item.title }}</template>
              </el-menu-item>
            </template>
          </el-menu>
        </el-scrollbar>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <!-- 顶部导航栏 -->
        <el-header class="header">
          <div class="header-left">
            <el-icon :size="20" class="breadcrumb-icon" @click="isCollapse = !isCollapse">
              <Fold v-if="!isCollapse" /><Expand v-else />
            </el-icon>
          </div>
          <div class="header-right">
            <el-dropdown trigger="click" @command="handleCommand">
              <div class="user-info">
                <el-avatar :size="32" :icon="UserFilled" />
                <span class="username">{{ adminStore.userInfo?.username || '管理员' }}</span>
                <el-tag v-if="adminStore.role" size="small" :type="isSuperAdmin ? 'danger' : 'info'" style="margin-left:4px">{{ adminStore.role }}</el-tag>
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

    <AdminAiFloatingButton v-if="canUseAdminAssistant" v-model:visible="showAiAssistant" />
    <AdminAiChatPanel v-if="canUseAdminAssistant" v-model:visible="showAiAssistant" />
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
  transition: width 0.3s;
}

.menu-scrollbar {
  height: calc(100vh - 60px);
}

.menu-scrollbar :deep(.el-scrollbar__bar.is-vertical) {
  width: 4px;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60px;
  background-color: #2b3a4a;
  gap: 8px;
  white-space: nowrap;
  overflow: hidden;
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

.el-menu-vertical:not(.el-menu--collapse) {
  width: 220px;
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
