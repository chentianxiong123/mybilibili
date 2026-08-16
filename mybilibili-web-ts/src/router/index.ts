import { createRouter, createWebHistory } from 'vue-router'
import { hasAuthSession } from '../utils/auth'

// ====== Web 路由（用户侧） ======
const webRoutes = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/web/HomeView.vue'),
    meta: { title: '首页 - 哔哩哔哩', layout: 'home', section: 'web' }
  },
  {
    path: '/search',
    name: 'search',
    component: () => import('../views/web/SearchView.vue'),
    meta: { title: '搜索 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/category/:id',
    name: 'category',
    component: () => import('../views/web/CategoryView.vue'),
    meta: { title: '分类 - 哔哩哔哩', layout: 'home', section: 'web' }
  },
  {
    path: '/manuscript/:id',
    name: 'manuscript',
    component: () => import('../views/web/VideoView.vue'),
    props: (route: any) => ({
      manuscriptId: route.params.id,
      p: parseInt(route.query.p as string) || 1
    }),
    meta: { title: '视频播放 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/user/:id',
    name: 'user',
    component: () => import('../views/web/UserView.vue'),
    meta: { title: '用户主页 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('../views/web/UserProfileView.vue'),
    meta: { title: '个人主页 - 哔哩哔哩', layout: 'simple', section: 'web' },
    children: [
      { path: '', name: 'profile-redirect', redirect: '/profile/home' },
      { path: 'home', name: 'profile-home', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'dynamic', name: 'profile-dynamic', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'submissions', name: 'profile-submissions', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'collections', name: 'profile-collections', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'favorites', name: 'profile-favorites', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'settings', name: 'profile-settings', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'following', name: 'profile-following', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'followers', name: 'profile-followers', component: () => import('../views/web/UserProfileView.vue') }
    ]
  },
  {
    path: '/profile/:id',
    name: 'user-profile',
    component: () => import('../views/web/UserProfileView.vue'),
    meta: { title: '用户主页 - 哔哩哔哩', layout: 'simple', section: 'web' },
    children: [
      { path: '', name: 'user-profile-redirect', redirect: (to: any) => `/profile/${to.params.id}/home` },
      { path: 'home', name: 'user-profile-home', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'dynamic', name: 'user-profile-dynamic', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'submissions', name: 'user-profile-submissions', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'collections', name: 'user-profile-collections', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'favorites', name: 'user-profile-favorites', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'following', name: 'user-profile-following', component: () => import('../views/web/UserProfileView.vue') },
      { path: 'followers', name: 'user-profile-followers', component: () => import('../views/web/UserProfileView.vue') }
    ]
  },
  {
    path: '/create-center',
    name: 'create-center',
    component: () => import('../views/web/CreateCenterView.vue'),
    meta: { title: '创作中心 - 哔哩哔哩', layout: 'none', section: 'web' },
    children: [
      { path: '', name: 'create-center-redirect', redirect: '/create-center/home' },
      { path: 'home', name: 'create-center-home', meta: { title: '创作中心首页 - 哔哩哔哩', layout: 'none' } },
      { path: 'upload', name: 'create-center-upload', meta: { title: '投稿 - 哔哩哔哩', layout: 'none' } },
      { path: 'content', name: 'create-center-content', meta: { title: '内容管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'content-articles', name: 'create-center-content-articles', meta: { title: '稿件管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'content-appeal', name: 'create-center-content-appeal', meta: { title: '申述管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'content-subtitle', name: 'create-center-content-subtitle', meta: { title: '字幕管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'data', name: 'create-center-data', meta: { title: '数据中心 - 哔哩哔哩', layout: 'none' } },
      { path: 'fans', name: 'create-center-fans', meta: { title: '粉丝管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'interaction', name: 'create-center-interaction', meta: { title: '互动管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'interaction-comment', name: 'create-center-interaction-comment', meta: { title: '评论管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'interaction-danmu', name: 'create-center-interaction-danmu', meta: { title: '弹幕管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'revenue', name: 'create-center-revenue', meta: { title: '收益管理 - 哔哩哔哩', layout: 'none' } },
      { path: 'settings', name: 'create-center-settings', meta: { title: '创作设置 - 哔哩哔哩', layout: 'none' } }
    ]
  },
  {
    path: '/dynamic',
    name: 'dynamic',
    component: () => import('../views/web/DynamicView.vue'),
    meta: { title: '动态 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/dynamic/:id',
    name: 'dynamic-detail',
    component: () => import('../views/web/DynamicDetailView.vue'),
    meta: { title: '动态详情 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/history',
    name: 'history',
    component: () => import('../views/web/HistoryView.vue'),
    meta: { title: '历史记录 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/collections',
    name: 'collections',
    component: () => import('../views/web/CollectionListView.vue'),
    meta: { title: '我的合集 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/collection/:id',
    name: 'collection-detail',
    component: () => import('../views/web/CollectionDetailView.vue'),
    meta: { title: '合集详情 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/collection/:id/edit',
    name: 'collection-edit',
    component: () => import('../views/web/CollectionEditView.vue'),
    meta: { title: '编辑合集 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/collection/create',
    name: 'collection-create',
    component: () => import('../views/web/CollectionEditView.vue'),
    meta: { title: '创建合集 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/personal-center',
    name: 'personal-center',
    component: () => import('../views/web/PersonalCenterView.vue'),
    meta: { title: '个人中心 - 哔哩哔哩', layout: 'simple', section: 'web' },
    children: [
      { path: '', name: 'personal-center-redirect', redirect: '/personal-center/home' },
      { path: 'home', name: 'personal-center-home', component: () => import('../views/web/personal/HomeView.vue') },
      { path: 'info', name: 'personal-center-info', component: () => import('../views/web/personal/InfoView.vue') },
      { path: 'avatar', name: 'personal-center-avatar', component: () => import('../views/web/AvatarView.vue') },
      { path: 'login-logs', name: 'personal-center-login-logs', component: () => import('../views/web/personal/LoginLogsView.vue') }
    ]
  },
  {
    path: '/message',
    name: 'message',
    component: () => import('../views/web/message/MessageView.vue'),
    meta: { title: '消息中心 - 哔哩哔哩', layout: 'simple', requiresAuth: true, section: 'web' }
  },
  {
    path: '/message/:type',
    name: 'message-type',
    component: () => import('../views/web/message/MessageView.vue'),
    meta: { title: '消息中心 - 哔哩哔哩', layout: 'simple', requiresAuth: true, section: 'web' }
  },
  {
    path: '/live',
    name: 'live',
    component: () => import('../views/web/live/LiveListView.vue'),
    meta: { title: '直播 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/live/:roomId',
    name: 'live-room',
    component: () => import('../views/web/live/LiveRoomView.vue'),
    meta: { title: '直播间 - 哔哩哔哩', layout: 'simple', section: 'web' }
  },
  {
    path: '/live/push',
    name: 'live-push',
    component: () => import('../views/web/live/LivePushView.vue'),
    meta: { title: '直播推流 - 哔哩哔哩', layout: 'simple', requiresAuth: true, section: 'web' }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/web/LoginView.vue'),
    meta: { title: '登录 - 哔哩哔哩', layout: 'none', section: 'web' }
  },
  {
    path: '/register',
    name: 'register',
    redirect: (to: any) => ({ path: '/login', query: { ...to.query, mode: 'register' } })
  }
]

// ====== Admin 路由（管理后台） ======
const adminRoutes = [
  {
    path: '/admin',
    name: 'admin',
    meta: { title: '管理后台 - 哔哩哔哩', layout: 'none', section: 'admin' },
    children: [
      { path: '', name: 'admin-redirect', redirect: '/admin/dashboard' },
      {
        path: 'login',
        name: 'admin-login',
        component: () => import('../views/admin/LoginView.vue'),
        meta: { title: '管理员登录 - 管理后台', layout: 'none' }
      },
      {
        path: 'dashboard',
        name: 'admin-dashboard',
        component: () => import('../views/admin/DashboardView.vue'),
        meta: { title: '数据概览 - 管理后台', requiresAuth: true, permission: 'statistics:manage' }
      },
      {
        path: 'users',
        name: 'admin-users',
        component: () => import('../views/admin/UsersView.vue'),
        meta: { title: '用户管理 - 管理后台', requiresAuth: true, permission: 'user:manage' }
      },
      {
        path: 'admins',
        name: 'admin-admins',
        component: () => import('../views/admin/AdminsView.vue'),
        meta: { title: '管理员与角色权限 - 管理后台', requiresAuth: true, superAdminOnly: true, permission: 'role:manage' }
      },
      {
        path: 'manuscripts',
        name: 'admin-manuscripts',
        component: () => import('../views/admin/ManuscriptsView.vue'),
        meta: { title: '稿件管理 - 管理后台', requiresAuth: true, permission: 'review:manage' }
      },
      {
        path: 'content-review',
        name: 'admin-content-review',
        component: () => import('../views/admin/ContentReviewView.vue'),
        meta: { title: '内容审核中心 - 管理后台', requiresAuth: true, permission: 'review:manage' }
      },
      {
        path: 'categories',
        name: 'admin-categories',
        component: () => import('../views/admin/CategoriesView.vue'),
        meta: { title: '分类管理 - 管理后台', requiresAuth: true, permission: 'category:manage' }
      },
      {
        path: 'banner-images',
        name: 'admin-banner-images',
        component: () => import('../views/admin/BannerImagesView.vue'),
        meta: { title: '图片管理 - 管理后台', requiresAuth: true, permission: 'banner:manage' }
      },
      {
        path: 'live-rooms',
        name: 'admin-live-rooms',
        component: () => import('../views/admin/LiveRoomsView.vue'),
        meta: { title: '直播管理 - 管理后台', requiresAuth: true, permission: 'live:manage' }
      },
      {
        path: 'subtitles',
        name: 'admin-subtitles',
        component: () => import('../views/admin/SubtitleManagementView.vue'),
        meta: { title: '字幕管理 - 管理后台', requiresAuth: true, permission: 'video:manage' }
      },
      {
        path: 'prohibited-words',
        name: 'admin-prohibited-words',
        component: () => import('../views/admin/ProhibitedWordsView.vue'),
        meta: { title: '违禁词与安全设置 - 管理后台', requiresAuth: true, permission: 'comment:manage' }
      },
      {
        path: 'recommend-config',
        name: 'admin-recommend-config',
        component: () => import('../views/admin/RecommendConfigView.vue'),
        meta: { title: '推荐配置 - 管理后台', requiresAuth: true, permission: 'search:manage' }
      },
      {
        path: 'index-manager',
        name: 'admin-index-manager',
        component: () => import('../views/admin/IndexManagerView.vue'),
        meta: { title: '索引管理 - 管理后台', requiresAuth: true, permission: 'search:manage' }
      },
      {
        path: 'ai-usage',
        name: 'admin-ai-usage',
        component: () => import('../views/admin/AiUsageView.vue'),
        meta: { title: 'AI 用量统计 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
      },
      {
        path: 'ai-skills',
        name: 'admin-ai-skills',
        component: () => import('../views/admin/AiSkillsView.vue'),
        meta: { title: 'AI 技能管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
      },
      {
        path: 'api-management',
        name: 'admin-api-management',
        component: () => import('../views/admin/ApiManagementView.vue'),
        meta: { title: 'AI 渠道管理 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
      },
      {
        path: 'customer-chat',
        name: 'admin-customer-chat',
        component: () => import('../views/admin/CustomerChatView.vue'),
        meta: { title: '客服会话 - 管理后台', requiresAuth: true, permission: 'ai:manage' }
      },
      {
        path: 'support-tickets',
        name: 'admin-support-tickets',
        component: () => import('../views/admin/SupportTicketsView.vue'),
        meta: { title: '工单中心 - 管理后台', requiresAuth: true, permission: 'operation:manage' }
      },
      {
        path: 'operation-tasks',
        name: 'admin-operation-tasks',
        component: () => import('../views/admin/OperationTasksView.vue'),
        meta: { title: '任务中心 - 管理后台', requiresAuth: true, permission: 'operation:manage' }
      },
      {
        path: 'audit-logs',
        name: 'admin-audit-logs',
        component: () => import('../views/admin/AuditLogsView.vue'),
        meta: { title: '审计日志 - 管理后台', requiresAuth: true, permission: 'audit:manage' }
      },
      {
        path: 'login-logs',
        name: 'admin-login-logs',
        component: () => import('../views/admin/LoginLogsView.vue'),
        meta: { title: '登录日志 - 管理后台', requiresAuth: true, permission: 'security:manage' }
      },
      {
        path: 'no-permission',
        name: 'admin-no-permission',
        component: () => import('../views/admin/NoPermissionView.vue'),
        meta: { title: '暂无权限 - 管理后台', requiresAuth: true }
      }
    ]
  }
]

// ====== 合并路由 ======
const routes: any = [
  ...webRoutes,
  ...adminRoutes,
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/web/NotFoundView.vue'),
    meta: { title: '页面未找到 - 哔哩哔哩', layout: 'simple' }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, savedPosition) {
    return savedPosition || { top: 0 }
  }
})

// ====== Admin 权限检查 ======
function getAdminPermissions(): string[] {
  try {
    return JSON.parse(localStorage.getItem('admin_permissions') || '[]')
  } catch {
    return []
  }
}

function hasPermission(permission?: string): boolean {
  if (!permission) return true
  const role = localStorage.getItem('admin_role')
  if (role === '超级管理员') return true
  return getAdminPermissions().includes(permission)
}

function getFirstAllowedAdminPath(): string {
  const role = localStorage.getItem('admin_role')
  const permissions = getAdminPermissions()
  if (role === '超级管理员') return '/admin/dashboard'
  const adminRouteOrder = [
    { path: '/admin/dashboard', permission: 'statistics:manage' },
    { path: '/admin/manuscripts', permission: 'review:manage' },
    { path: '/admin/users', permission: 'user:manage' }
  ]
  const allowed = adminRouteOrder.find(item => permissions.includes(item.permission))
  return allowed?.path || '/admin/no-permission'
}

// ====== 路由守卫 ======
router.beforeEach((to, from, next) => {
  // 设置标题
  if (to.meta.title) {
    document.title = to.meta.title as string
  }

  // Admin 路由守卫
  if (to.path.startsWith('/admin') && to.name !== 'admin-login') {
    const adminToken = localStorage.getItem('admin_token')
    const role = localStorage.getItem('admin_role')

    if (to.meta.requiresAuth && !adminToken) {
      next('/admin/login')
      return
    }
    if (to.path === '/admin/login' && adminToken) {
      next(getFirstAllowedAdminPath())
      return
    }
    if (to.path === '/admin' && adminToken) {
      next(getFirstAllowedAdminPath())
      return
    }
    if (to.meta.superAdminOnly && role !== '超级管理员') {
      next(getFirstAllowedAdminPath())
      return
    }
    if (to.meta.requiresAuth && !hasPermission(to.meta.permission as string)) {
      const fallback = getFirstAllowedAdminPath()
      next(fallback !== to.path ? fallback : '/admin/no-permission')
      return
    }
  }

  // Web 路由守卫
  if (to.meta.requiresAuth && !hasAuthSession()) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }
  if (to.path === '/login' && hasAuthSession()) {
    next('/')
    return
  }

  next()
})

export default router
