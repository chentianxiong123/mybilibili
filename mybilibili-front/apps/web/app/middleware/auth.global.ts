import { hasAuthSession } from '~/utils/auth'

export default defineNuxtRouteMiddleware((to) => {
  // 客户端环境才可访问 localStorage
  if (import.meta.server) return

  // 设置标题
  if (to.meta.title) {
    document.title = to.meta.title as string
  }

  // Web 路由守卫
  if (to.meta.requiresAuth && !hasAuthSession()) {
    return navigateTo({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (to.path === '/login' && hasAuthSession()) {
    return navigateTo('/')
  }
})
