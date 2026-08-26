<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowDown, Connection, Cpu, DataAnalysis, DataBoard, DataLine,
  Document, DocumentChecked, Expand, Fold, Folder, Headset, List,
  Lock, Message, Monitor, Operation, Picture, Setting, SwitchButton,
  Tickets, User, UserFilled, VideoCamera, VideoPlay, Warning
} from '@element-plus/icons-vue'
import { useAdminStore } from '~/stores/admin'
import AdminAiFloatingButton from '~/components/AdminAiFloatingButton.vue'
import AdminAiChatPanel from '~/components/AdminAiChatPanel.vue'

const route = useRoute()
const router = useRouter()
const adminStore = useAdminStore()
const isSuperAdmin = computed(() => adminStore.role === '超级管理员')
const canUseAdminAssistant = computed(() => adminStore.hasPermission('ai:manage'))
const isCollapse = ref(false)
const showAiAssistant = ref(false)

const isLoginPage = computed(() => route.path === '/login')

const iconMap: Record<string, any> = {
  Connection, Cpu, DataAnalysis, DataBoard, DataLine, Document, DocumentChecked,
  Folder, Headset, List, Lock, Message, Monitor, Operation, Picture, Setting, Tickets, User, VideoCamera, VideoPlay, Warning
}

const allMenuItems = [
  {
    type: 'group', icon: 'Operation', title: '运营板块',
    children: [
      { path: '/operation-tasks', icon: 'List', title: '任务中心', permission: 'operation:manage' },
      { path: '/video-process', icon: 'VideoCamera', title: '视频处理看板', permission: 'video:manage' },
      { path: '/support-tickets', icon: 'Message', title: '工单中心', permission: 'operation:manage' },
      { path: '/index-manager', icon: 'DataLine', title: '索引管理', permission: 'search:manage' },
      { path: '/recommend-config', icon: 'DataAnalysis', title: '推荐配置', permission: 'search:manage' }
    ]
  },
  {
    type: 'group', icon: 'DocumentChecked', title: '审核治理',
    children: [
      { path: '/manuscripts', icon: 'Document', title: '稿件管理', permission: 'review:manage' },
      { path: '/content-review', icon: 'DocumentChecked', title: '内容审核中心', permission: 'review:manage' },
      { path: '/prohibited-words', icon: 'Warning', title: '违禁词与安全设置', permission: 'comment:manage' }
    ]
  },
  {
    type: 'group', icon: 'Cpu', title: 'AI 管理',
    children: [
      { path: '/ai-usage', icon: 'DataAnalysis', title: 'AI 用量统计', permission: 'ai:manage' },
      { path: '/ai-skills', icon: 'Cpu', title: 'AI 技能管理', permission: 'ai:manage' },
      { path: '/api-management', icon: 'Setting', title: 'AI 渠道管理', permission: 'ai:manage' },
      { path: '/customer-chat', icon: 'Headset', title: '客服会话', permission: 'ai:manage' }
    ]
  },
  {
    type: 'group', icon: 'VideoPlay', title: '媒体管理',
    children: [
      { path: '/categories', icon: 'Folder', title: '分类管理', permission: 'category:manage' },
      { path: '/banner-images', icon: 'Picture', title: '图片管理', permission: 'banner:manage' },
      { path: '/subtitles', icon: 'Tickets', title: '字幕管理', permission: 'video:manage' },
      { path: '/live-rooms', icon: 'Connection', title: '直播管理', permission: 'live:manage' }
    ]
  },
  {
    type: 'group', icon: 'Lock', title: '系统管理',
    children: [
      { path: '/dashboard', icon: 'DataBoard', title: '数据概览', permission: 'statistics:manage' },
      { path: '/users', icon: 'User', title: '用户管理', permission: 'user:manage' },
      { path: '/login-logs', icon: 'List', title: '登录日志', permission: 'security:manage' },
      { path: '/audit-logs', icon: 'Tickets', title: '审计日志', permission: 'audit:manage' },
      { path: '/transcode-config', icon: 'Cpu', title: '转码配置', permission: 'video:manage' },
      { path: '/admins', icon: 'Lock', title: '管理员与角色权限', permission: 'role:manage', superAdminOnly: true }
    ]
  }
]

const filterMenuNode = (node: any) => {
  if (node.type === 'group') {
    const children = (node.children || []).filter(
      (child: any) => (!child.superAdminOnly || isSuperAdmin.value) && adminStore.hasPermission(child.permission)
    )
    return children.length ? { ...node, children } : null
  }
  return (!node.superAdminOnly || isSuperAdmin.value) && adminStore.hasPermission(node.permission) ? node : null
}

const menuItems = computed(() => allMenuItems.map(filterMenuNode).filter(Boolean))

const activeMenu = computed(() => {
  const path = route.path
  const prefixes = ['/users', '/manuscripts', '/operation-tasks', '/audit-logs', '/prohibited-words',
    '/content-review', '/categories', '/banner-images', '/index-manager', '/recommend-config',
    '/admins', '/api-management', '/ai-skills', '/ai-usage', '/support-tickets', '/live-rooms',
    '/login-logs', '/customer-chat', '/subtitles']
  for (const p of prefixes) {
    if (path.startsWith(p)) return p
  }
  return path
})

const handleCommand = (command: string) => {
  if (command === 'logout') {
    adminStore.logout()
    router.push('/login')
  }
}
</script>

<template>
  <div class="app-container" v-if="!isLoginPage">
    <el-container class="layout-container">
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
                <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
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

      <el-container>
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

        <el-main class="main">
          <RouterView />
        </el-main>
      </el-container>
    </el-container>

    <AdminAiFloatingButton v-if="canUseAdminAssistant" v-model:visible="showAiAssistant" />
    <AdminAiChatPanel v-if="canUseAdminAssistant" v-model:visible="showAiAssistant" />
  </div>

  <div v-else class="login-page">
    <RouterView />
  </div>
</template>

<style>
@import "@mybilibili/ui";

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  background-color: var(--bili-bg);
  color: var(--bili-text);
}

a {
  text-decoration: none;
  color: inherit;
}
</style>

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