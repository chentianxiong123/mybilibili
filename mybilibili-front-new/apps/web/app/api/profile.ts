import api from './client'
import { safeStorage } from '../utils/safeStorage'

export const profileApi = {
  getMyProfile() {
    const userStr = safeStorage.getItem('user')
    if (!userStr) return Promise.reject(new Error('未登录'))
    const user = JSON.parse(userStr)
    return api.get(`/profile/${user.id}`)
  },
  getProfile(userId: number) {
    return api.get(`/profile/${userId}`)
  }
}