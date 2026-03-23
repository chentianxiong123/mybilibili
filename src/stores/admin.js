import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminLogin } from '../api/admin'

export const useAdminStore = defineStore('admin', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('admin_user')) || null)

  const login = async (loginData) => {
    try {
      const res = await adminLogin(loginData)
      if (res.code === 200 || res.success) {
        token.value = res.data.token || res.token
        userInfo.value = res.data.user || res.user || { username: loginData.username }
        localStorage.setItem('admin_token', token.value)
        localStorage.setItem('admin_user', JSON.stringify(userInfo.value))
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
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
  }

  return {
    token,
    userInfo,
    login,
    logout
  }
})