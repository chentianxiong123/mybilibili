import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminLogin } from '@/api/admin'
import { clearServerSession } from '@/api/session'
import { scrubCredentials } from '@/utils/auth'

export const useAdminStore = defineStore('admin', () => {
  // 跳转责任收归 store：之前散落在 app.vue / NoPermissionView 的 router.push('/login')
  // 任何一处遗漏都会出现「logout 200 但页面停在 dashboard」的 UX bug。
  // store 里读不到 router 时（比如单测）降级为 no-op，不影响功能。
  let router: ReturnType<typeof useRouter> | null = null
  try { router = useRouter() } catch { /* 不在 router 上下文 */ }

  // 凭证 admin_token 是 HttpOnly cookie，store 里只留展示与授权信息
  const userInfo = ref(scrubCredentials(JSON.parse(localStorage.getItem('admin_user')) || null))
  const role = ref(localStorage.getItem('admin_role') || '')
  const permissions = ref(JSON.parse(localStorage.getItem('admin_permissions') || '[]'))

  const login = async (loginData) => {
    try {
      const res = await adminLogin(loginData)
      if (res.code === 200 || res.success) {
        // 后台登录响应给的是 admin_id/username/role，没有 adminUser 对象，
        // 这里拼出一份展示信息——它同时是前端判断登录态的唯一信号。
        const adminId = res.data.adminUser?.id || res.data.user?.id || res.data.admin_id || res.data.adminId || null
        userInfo.value = res.data.adminUser || res.data.user || res.user
          || { id: adminId, username: res.data.username || loginData.username }
        role.value = res.data.role || '管理员'
        permissions.value = res.data.permissions || []
        // admin_user 会被 persist，登录响应里的 token 不能跟着进去
        localStorage.setItem('admin_user', JSON.stringify(scrubCredentials(userInfo.value)))
        localStorage.setItem('admin_role', role.value)
        localStorage.setItem('admin_permissions', JSON.stringify(permissions.value))
        if (adminId) localStorage.setItem('admin_id', adminId)
        return { success: true }
      }
      return { success: false, message: res.message || '登录失败' }
    } catch (error) {
      return { success: false, message: error.message || '登录失败' }
    }
  }

  const logout = async () => {
    // 先让服务端作废 HttpOnly 的 admin_token / admin_refresh，否则 cookie 会留到过期
    await clearServerSession()
    userInfo.value = null
    role.value = ''
    permissions.value = []
    localStorage.removeItem('admin_token') // 清掉升级前的遗留副本
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_role')
    localStorage.removeItem('admin_permissions')
    localStorage.removeItem('admin_id')
    // 跳转由 store 兜底，调用方不用关心。已在 /login 上时 push 会触发导航守卫 false，吞掉
    if (router) {
      try { await router.push('/login') } catch { /* 已在 /login */ }
    }
  }

  const hasPermission = (permission) => {
    if (!permission) return true
    if (role.value === '超级管理员') return true
    return permissions.value.includes(permission)
  }

  return {
    userInfo,
    role,
    permissions,
    hasPermission,
    login,
    logout
  }
})
