import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminLogin } from '@/api/admin'
import { clearServerSession } from '@/api/session'
import { scrubCredentials } from '@/utils/auth'

export const useAdminStore = defineStore('admin', () => {
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

  const logout = () => {
    // 先让服务端作废 HttpOnly 的 admin_token / admin_refresh，否则 cookie 会留到过期
    void clearServerSession()
    userInfo.value = null
    role.value = ''
    permissions.value = []
    localStorage.removeItem('admin_token') // 清掉升级前的遗留副本
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_role')
    localStorage.removeItem('admin_permissions')
    localStorage.removeItem('admin_id')
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
