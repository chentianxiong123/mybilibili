import axios from 'axios'
import { ElMessage } from 'element-plus'
import {
  clearAuthSession,
  getRefreshToken,
  getToken,
  setAuthSession,
  getAdminToken,
  clearAdminSession
} from '../utils/auth'

// ====== Axios 实例 ======
const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true
})

// ====== 请求拦截器 ======
api.interceptors.request.use(
  config => {
    const url = config.url || ''
    const isImageRequest = url.includes('/covers/') || url.includes('/images/') || url.match(/\.(jpg|jpeg|png|gif|mp4)$/i)
    if (isImageRequest) return config

    // 优先用 admin_token，其次用 user token
    const adminToken = getAdminToken()
    const userToken = getToken()
    const token = adminToken || userToken
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// ====== Token 刷新 ======
let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (error: any) => void }> = []

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) prom.reject(error)
    else prom.resolve(token!)
  })
  failedQueue = []
}

// ====== 响应拦截器 ======
api.interceptors.response.use(
  response => response.data,
  error => {
    const originalRequest = error.config

    // 401 处理 - token 刷新（仅 user token）
    if (error.response?.status === 401 && !originalRequest._retry) {
      const refreshToken = getRefreshToken()
      const adminToken = getAdminToken()

      // admin 401 直接跳登录
      if (adminToken) {
        clearAdminSession()
        window.location.href = '/admin/login'
        return Promise.reject(error)
      }

      // user token 刷新
      if (!refreshToken || originalRequest.url === '/user/token/refresh') {
        clearAuthSession()
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return api(originalRequest)
        }).catch(err => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      return new Promise((resolve, reject) => {
        api.post('/user/token/refresh', { refreshToken })
          .then((res: any) => {
            if (res.code === 200 && res.data) {
              const { token, refreshToken: newRefreshToken } = res.data
              setAuthSession({ token, refreshToken: newRefreshToken || refreshToken })
              originalRequest.headers.Authorization = `Bearer ${token}`
              processQueue(null, token)
              resolve(api(originalRequest))
            } else {
              clearAuthSession()
              processQueue(new Error('refresh failed'))
              reject(error)
            }
          })
          .catch(err => {
            clearAuthSession()
            processQueue(err)
            reject(err)
          })
          .finally(() => { isRefreshing = false })
      })
    }

    // 通用错误提示
    if (error.response) {
      switch (error.response.status) {
        case 401: clearAuthSession(); break
        case 403: ElMessage.error('没有权限访问该资源'); break
        case 404: ElMessage.error('请求的资源不存在'); break
        case 500: ElMessage.error('服务器内部错误'); break
        default: ElMessage.error(error.response.data?.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export default api as {
  (config: any): Promise<any>
  (url: string, config?: any): Promise<any>
  get: (url: string, config?: any) => Promise<any>
  post: (url: string, data?: any, config?: any) => Promise<any>
  put: (url: string, data?: any, config?: any) => Promise<any>
  delete: (url: string, config?: any) => Promise<any>
  request: (config: any) => Promise<any>
  interceptors: typeof api.interceptors
}
