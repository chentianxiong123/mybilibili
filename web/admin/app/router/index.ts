import { createRouter, createWebHistory } from 'vue-router'
import { firstAllowedPathByPermissions } from './permissionRoutes'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('~/views/admin/LoginView.vue'),
      meta: { title: '登录 - 管理后台' }
    },
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('~/views/admin/DashboardView.vue'),
      meta: { title: '数据概览 - 管理后台', requiresAuth: true, permission: 'statistics:manage' }
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('~/views/admin/UsersView.vue'),
      meta: { title: '用户管理 - 管理后台', requiresAuth: true, permission: 'user:manage' }
    },
    {
      path: '/manuscripts',
      name: 'manuscripts',
      component: () => import('~/views/admin/ManuscriptsView.vue'),
      meta: { title: '稿件管理 - 管理后台', requiresAuth: true, permission: 'review:manage' }
    },
    {
      path: '/operation-tasks',
      name: 'operationTasks',
      component: () => import('~/views/admin/OperationTasksView.vue'),
      meta: { title: '任务中心 - 管理后台', requiresAuth: true, permission: 'operation:manage' }
    },
    {
      path: '/video-process',
      name: 'videoProcess',
      component: () => import('~/views/admin/VideoProcessView.vue'),
      meta: { title: '视频处理看板 - 管理后台', requiresAuth: true, permission: 'video:manage' }
    },
    {
      path: '/scheduled-tasks',
      name: 'scheduledTasks',
      component: () => import('~/views/admin/ScheduledTasksView.vue'),
      meta: { title: '定时任务 - 管理后台', requiresAuth: true, permission: 'operation:manage' }
    },
    {
      path: '/audit-logs',
      name: 'auditLogs',
      component: () => import('~/views/admin/AuditLogsView.vue'),
      meta: { title: '审计日志 - 管理后台', requiresAuth: true, permission: 'audit:manage' }
    },
    {
      path: '/prohibited-words',
      name: 'prohibitedWords',
      component: () => import('~/views/admin/ProhibitedWordsView.vue'),
      meta: { title: '违禁词与安全设置 - 管理后台', requiresAuth: true, permission: 'comment:manage' }
    },
    {
      path: '/content-review',
      name: 'contentReview',
      component: () => import('~/views/admin/ContentReviewView.vue'),
      meta: { title: '内容审核中心 - 管理后台', requiresAuth: true, permission: 'review:manage' }
    },
    {
      path: '/categories',
      name: 'categories',
      component: () => import('~/views/admin/CategoriesView.vue'),
      meta: { title: '分类管理 - 管理后台', requiresAuth: true, permission: 'category:manage' }
    },
    {
      path: '/recommend-config',
      name: 'recommendConfig',
      component: () => import('~/views/admin/RecommendConfigView.vue'),
      meta: { title: '推荐配置 - 管理后台', requiresAuth: true, permission: 'search:manage' }
    },
    {
      path: '/admins',
      name: 'admins',
      component: () => import('~/views/admin/AdminsView.vue'),
      meta: { title: '管理员与角色权限 - 管理后台', requiresAuth: true, superAdminOnly: true, permission: 'role:manage' }
    },
    {
      path: '/index-manager',
      name: 'indexManager',
      component: () => import('~/views/admin/IndexManagerView.vue'),
      meta: { title: '索引管理 - 管理后台', requiresAuth: true, permission: 'search:manage' }
    },
    {
      path: '/banner-images',
      name: 'bannerImages',
      component: () => import('~/views/admin/BannerImagesView.vue'),
      meta: { title: '图片管理 - 管理后台', requiresAuth: true, permission: 'banner:manage' }
    },
    {
      path: '/api-management',
      name: 'apiManagement',
      component: () => import('~/views/admin/ApiManagementView.vue'),
      meta: { title: 'AI 渠道管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/ai-usage',
      name: 'aiUsage',
      component: () => import('~/views/admin/AiUsageView.vue'),
      meta: { title: 'AI 用量统计 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/ai-skills',
      name: 'aiSkills',
      component: () => import('~/views/admin/AiSkillsView.vue'),
      meta: { title: 'AI 技能管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/support-tickets',
      name: 'supportTickets',
      component: () => import('~/views/admin/SupportTicketsView.vue'),
      meta: { title: '工单中心 - 管理后台', requiresAuth: true, permission: 'operation:manage' }
    },
    {
      path: '/live-rooms',
      name: 'liveRooms',
      component: () => import('~/views/admin/LiveRoomsView.vue'),
      meta: { title: '直播管理 - 管理后台', requiresAuth: true, permission: 'live:manage' }
    },
    {
      path: '/customer-chat',
      name: 'customerChat',
      component: () => import('~/views/admin/CustomerChatView.vue'),
      meta: { title: '客服会话 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/login-logs',
      name: 'loginLogs',
      component: () => import('~/views/admin/LoginLogsView.vue'),
      meta: { title: '登录日志 - 管理后台', requiresAuth: true, permission: 'security:manage' }
    },
    {
      path: '/transcode-config',
      name: 'transcodeConfig',
      component: () => import('~/views/admin/TranscodeConfigView.vue'),
      meta: { title: '转码配置 - 管理后台', requiresAuth: true, permission: 'video:manage' }
    },
    {
      path: '/no-permission',
      name: 'noPermission',
      component: () => import('~/views/admin/NoPermissionView.vue'),
      meta: { title: '暂无权限 - 管理后台', requiresAuth: true }
    }
  ]
})

const getAdminPermissions = (): string[] => {
  try {
    return JSON.parse(localStorage.getItem('admin_permissions') || '[]')
  } catch {
    return []
  }
}

const getAdminRole = (): string => {
  return localStorage.getItem('admin_role') || ''
}

const hasPermission = (permission?: string): boolean => {
  if (!permission) return true
  const role = getAdminRole()
  if (role === '超级管理员') return true
  return getAdminPermissions().includes(permission)
}

router.beforeEach((to) => {
  const token = localStorage.getItem('admin_token')
  const role = getAdminRole()

  let redirect: string | boolean = true
  if (to.meta.requiresAuth && !token) {
    redirect = '/login'
  } else if (to.path === '/login' && token) {
    redirect = firstAllowedPathByPermissions(role, getAdminPermissions())
  } else if (to.path === '/' && token) {
    redirect = firstAllowedPathByPermissions(role, getAdminPermissions())
  } else if (to.meta.superAdminOnly && role !== '超级管理员') {
    redirect = firstAllowedPathByPermissions(role, getAdminPermissions())
  } else if (to.meta.requiresAuth && !hasPermission(to.meta.permission as string | undefined)) {
    const fallback = firstAllowedPathByPermissions(role, getAdminPermissions())
    redirect = fallback !== to.path ? fallback : '/no-permission'
  }

  if (to.meta.title) {
    document.title = to.meta.title as string
  }

  return redirect
})

export default router