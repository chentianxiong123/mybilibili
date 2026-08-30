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
      meta: { title: '数据概览 - 管理后台', requiresAuth: true }
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue'),
      meta: { title: '用户管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/manuscripts',
      name: 'manuscripts',
      component: () => import('../views/ManuscriptsView.vue'),
      meta: { title: '稿件管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/video-process',
      name: 'videoProcess',
      component: () => import('../views/VideoProcessView.vue'),
      meta: { title: '视频处理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/prohibited-words',
      name: 'prohibitedWords',
      component: () => import('../views/ProhibitedWordsView.vue'),
      meta: { title: '违禁词与安全设置 - 管理后台', requiresAuth: true }
    },
    {
      path: '/content-review',
      name: 'contentReview',
      component: () => import('../views/ContentReviewView.vue'),
      meta: { title: '内容审核中心 - 管理后台', requiresAuth: true }
    },
    {
      path: '/categories',
      name: 'categories',
      component: () => import('../views/CategoriesView.vue'),
      meta: { title: '分类管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/roles',
      redirect: '/admins'
    },
    {
      path: '/admins',
      name: 'admins',
      component: () => import('../views/AdminsView.vue'),
      meta: { title: '管理员与角色权限 - 管理后台', requiresAuth: true, superAdminOnly: true }
    },
    {
      path: '/subtitles',
      name: 'subtitles',
      component: () => import('../views/SubtitleManagementView.vue'),
      meta: { title: '字幕管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/index-manager',
      name: 'indexManager',
      component: () => import('../views/IndexManagerView.vue'),
      meta: { title: '索引管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/banner-images',
      name: 'bannerImages',
      component: () => import('../views/BannerImagesView.vue'),
      meta: { title: '图片管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/api-management',
      name: 'apiManagement',
      component: () => import('../views/ApiManagementView.vue'),
      meta: { title: 'AI 渠道管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/ai-usage',
      name: 'aiUsage',
      component: () => import('../views/AiUsageView.vue'),
      meta: { title: 'AI 用量统计 - 管理后台', requiresAuth: true }
    },
    {
      path: '/ai-skills',
      name: 'aiSkills',
      component: () => import('../views/AiSkillsView.vue'),
      meta: { title: 'AI 技能管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/ai-feedback',
      name: 'aiFeedback',
      component: () => import('../views/AiFeedbackView.vue'),
      meta: { title: 'AI 反馈管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/live-rooms',
      name: 'liveRooms',
      component: () => import('../views/LiveRoomsView.vue'),
      meta: { title: '直播管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/meeting-admin',
      name: 'meetingAdmin',
      component: () => import('../views/MeetingView.vue'),
      meta: { title: '会议管理 - 管理后台', requiresAuth: true }
    },
    {
      path: '/customer-chat',
      name: 'customerChat',
      component: () => import('../views/CustomerChatView.vue'),
      meta: { title: '客服会话 - 管理后台', requiresAuth: true }
    },
    {
      path: '/login-logs',
      name: 'loginLogs',
      component: () => import('../views/LoginLogsView.vue'),
      meta: { title: '登录日志 - 管理后台', requiresAuth: true }
    }
  ]
})

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
  } else {
    next()
  }

  // 设置页面标题
  if (to.meta.title) {
    document.title = to.meta.title
  }
})

export default router