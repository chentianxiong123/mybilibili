// API 客户端 - 与 mybilibili-web/src/api/index.js 保持一致
import axios from 'axios'

// 简单 toast 提示（原项目不用 Element Plus）
function showToast(msg: string) {
  const el = document.createElement('div')
  el.className = 'wap-toast'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
}

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true
})

// 请求拦截器：添加token + 用户ID
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    const isImageRequest = config.url.includes('/covers/') || config.url.includes('/images/')
    if (token && !isImageRequest) {
      config.headers.Authorization = `Bearer ${token}`
    }
    config.headers['X-Client-Platform'] = 'wap'
    const userStr = localStorage.getItem('user')
    if (userStr) {
      try {
        const user = JSON.parse(userStr)
        if (user && user.id) {
          config.headers['X-User-Id'] = user.id
        }
      } catch (e) {}
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          break
        case 403:
          showToast('没有权限访问该资源')
          break
        case 404:
          showToast('请求的资源不存在')
          break
        case 500:
          showToast('服务器内部错误')
          break
        default:
          showToast(error.response.data?.message || '请求失败')
      }
    } else {
      showToast('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export default api