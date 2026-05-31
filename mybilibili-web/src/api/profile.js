import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  const userStr = localStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      if (user && user.id) config.headers['X-User-Id'] = user.id
    } catch (e) {}
  }
  return config
})

api.interceptors.response.use(res => res.data, err => Promise.reject(err))

export const profileApi = {
  getMyProfile() {
    const userStr = localStorage.getItem('user')
    if (!userStr) return Promise.reject(new Error('未登录'))
    const user = JSON.parse(userStr)
    return api.get(`/profile/${user.id}`)
  },
  getProfile(userId) {
    return api.get(`/profile/${userId}`)
  }
}
