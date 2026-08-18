import { hasAuthSession, getAdminToken, getAdminRole, getAdminPermissions } from '~/utils/auth'

export default defineNuxtRouteMiddleware((to) => {
  // 客户端环境才可访问 localStorage
  if (import.meta.server) return

  // 设置标题
  if (to.meta.title) {
    document.title = to.meta.title as string
  }

  // Admin 路由守卫
  if (to.path.startsWith('/admin') && to.name !== 'admin-login') {
    const adminToken = getAdminToken()
    const role = getAdminRole()

    if (to.meta.requiresAuth && !adminToken) {
      return navigateTo('/admin/login')
    }
    if (to.path === '/admin/login' && adminToken) {
      return navigateTo(getFirstAllowedAdminPath())
    }
    if (to.path === '/admin' && adminToken) {
      return navigateTo(getFirstAllowedAdminPath())
    }
    if (to.meta.superAdminOnly && role !== '超级管理员') {
      return navigateTo(getFirstAllowedAdminPath())
    }
    if (to.meta.requiresAuth && !hasPermission(to.meta.permission as string)) {
      const fallback = getFirstAllowedAdminPath()
      return navigateTo(fallback !== to.path ? fallback : '/admin/no-permission')
    }
  }

  // Web 路由守卫
  if (to.meta.requiresAuth && !hasAuthSession()) {
    return navigateTo({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (to.path === '/login' && hasAuthSession()) {
    return navigateTo('/')
  }
})

function hasPermission(permission?: string): boolean {
  if (!permission) return true
  const role = getAdminRole()
  if (role === '超级管理员') return true
  return getAdminPermissions().includes(permission)
}

function getFirstAllowedAdminPath(): string {
  const role = getAdminRole()
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