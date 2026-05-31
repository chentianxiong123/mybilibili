import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { title: '登录 - 管理后台' }
    },
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { title: '数据概览 - 管理后台', requiresAuth: true, permission: 'statistics:manage' }
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue'),
      meta: { title: '用户管理 - 管理后台', requiresAuth: true, permission: 'user:manage' }
    },
    {
      path: '/manuscripts',
      name: 'manuscripts',
      component: () => import('../views/ManuscriptsView.vue'),
      meta: { title: '稿件管理 - 管理后台', requiresAuth: true, permission: 'review:manage' }
    },
    {
      path: '/video-process',
      name: 'videoProcess',
      component: () => import('../views/VideoProcessView.vue'),
      meta: { title: '视频处理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/prohibited-words',
      name: 'prohibitedWords',
      component: () => import('../views/ProhibitedWordsView.vue'),
      meta: { title: '违禁词与安全设置 - 管理后台', requiresAuth: true, permission: 'comment:manage' }
    },
    {
      path: '/content-review',
      name: 'contentReview',
      component: () => import('../views/ContentReviewView.vue'),
      meta: { title: '内容审核中心 - 管理后台', requiresAuth: true, permission: 'review:manage' }
    },
    {
      path: '/categories',
      name: 'categories',
      component: () => import('../views/CategoriesView.vue'),
      meta: { title: '分类管理 - 管理后台', requiresAuth: true, permission: 'category:manage' }
    },
    {
      path: '/recommend-config',
      name: 'recommendConfig',
      component: () => import('../views/RecommendConfigView.vue'),
      meta: { title: '推荐配置 - 管理后台', requiresAuth: true, permission: 'search:manage' }
    },
    {
      path: '/roles',
      redirect: '/admins'
    },
    {
      path: '/admins',
      name: 'admins',
      component: () => import('../views/AdminsView.vue'),
      meta: { title: '管理员与角色权限 - 管理后台', requiresAuth: true, superAdminOnly: true, permission: 'role:manage' }
    },
    {
      path: '/subtitles',
      name: 'subtitles',
      component: () => import('../views/SubtitleManagementView.vue'),
      meta: { title: '字幕管理 - 管理后台', requiresAuth: true, permission: 'video:manage' }
    },
    {
      path: '/index-manager',
      name: 'indexManager',
      component: () => import('../views/IndexManagerView.vue'),
      meta: { title: '索引管理 - 管理后台', requiresAuth: true, permission: 'search:manage' }
    },
    {
      path: '/banner-images',
      name: 'bannerImages',
      component: () => import('../views/BannerImagesView.vue'),
      meta: { title: '图片管理 - 管理后台', requiresAuth: true, permission: 'banner:manage' }
    },
    {
      path: '/api-management',
      name: 'apiManagement',
      component: () => import('../views/ApiManagementView.vue'),
      meta: { title: 'AI 渠道管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/ai-usage',
      name: 'aiUsage',
      component: () => import('../views/AiUsageView.vue'),
      meta: { title: 'AI 用量统计 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/ai-skills',
      name: 'aiSkills',
      component: () => import('../views/AiSkillsView.vue'),
      meta: { title: 'AI 技能管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/ai-feedback',
      name: 'aiFeedback',
      component: () => import('../views/AiFeedbackView.vue'),
      meta: { title: 'AI 反馈管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/live-rooms',
      name: 'liveRooms',
      component: () => import('../views/LiveRoomsView.vue'),
      meta: { title: '直播管理 - 管理后台', requiresAuth: true, permission: 'live:manage' }
    },
    {
      path: '/meeting-admin',
      name: 'meetingAdmin',
      component: () => import('../views/MeetingView.vue'),
      meta: { title: '会议管理 - 管理后台', requiresAuth: true, permission: 'meeting:manage' }
    },
    {
      path: '/customer-chat',
      name: 'customerChat',
      component: () => import('../views/CustomerChatView.vue'),
      meta: { title: '客服会话 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
    },
    {
      path: '/login-logs',
      name: 'loginLogs',
      component: () => import('../views/LoginLogsView.vue'),
      meta: { title: '登录日志 - 管理后台', requiresAuth: true, permission: 'security:manage' }
    }
  ]
})

const getAdminPermissions = () => {
  try {
    return JSON.parse(localStorage.getItem('admin_permissions') || '[]')
  } catch {
    return []
  }
}

const hasPermission = (permission) => {
  if (!permission) return true
  const role = localStorage.getItem('admin_role')
  if (role === '超级管理员') return true
  return getAdminPermissions().includes(permission)
}

const clearAdminSession = () => {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_user')
  localStorage.removeItem('admin_role')
  localStorage.removeItem('admin_permissions')
  localStorage.removeItem('admin_id')
}

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('admin_token')
  const role = localStorage.getItem('admin_role')

  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else if (to.meta.superAdminOnly && role !== '超级管理员') {
    next('/dashboard')
  } else if (to.meta.requiresAuth && !hasPermission(to.meta.permission)) {
    if (getAdminPermissions().length === 0) {
      clearAdminSession()
      next('/login')
      return
    }
    const fallback = router.getRoutes()
      .find(route => route.meta.requiresAuth && !route.meta.superAdminOnly && hasPermission(route.meta.permission))
    next(fallback && fallback.path !== to.path ? fallback.path : '/login')
  } else {
    next()
  }

  // 设置页面标题
  if (to.meta.title) {
    document.title = to.meta.title
  }
})

export default router
