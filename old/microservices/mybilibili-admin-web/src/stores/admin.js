import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminLogin } from '../api/admin'

export const useAdminStore = defineStore('admin', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('admin_user')) || null)
  const role = ref(localStorage.getItem('admin_role') || '')
  const permissions = ref(JSON.parse(localStorage.getItem('admin_permissions') || '[]'))

  const login = async (loginData) => {
    try {
      const res = await adminLogin(loginData)
      if (res.code === 200 || res.success) {
        token.value = res.data.token || res.token
        userInfo.value = res.data.adminUser || res.data.user || res.user || { username: loginData.username }
        role.value = res.data.role || '管理员'
        permissions.value = res.data.permissions || []
        const adminId = res.data.adminUser?.id || res.data.user?.id || null
        localStorage.setItem('admin_token', token.value)
        localStorage.setItem('admin_user', JSON.stringify(userInfo.value))
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
    token.value = ''
    userInfo.value = null
    role.value = ''
    permissions.value = []
    localStorage.removeItem('admin_token')
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
    token,
    userInfo,
    role,
    permissions,
    hasPermission,
    login,
    logout
  }
})
