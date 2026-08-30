import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    // 处理隐私设置接口404错误，返回默认值
    if (error.response?.status === 404 && error.config?.url?.includes('/user/privacy/')) {
      // 返回默认的隐私设置数据
      return {
        code: 200,
        message: '接口不存在，返回默认设置',
        data: {
          publicCollection: true,
          publicBirthdayTags: false,
          publicCoinVideos: false,
          publicLikeVideos: false,
          publicFollowingList: false,
          publicFollowersList: false,
          tags: []
        }
      }
    }
    return Promise.reject(error)
  }
)

// 用户隐私设置相关API
export const userPrivacyApi = {
  // 获取隐私设置
  getPrivacySettings: () => api.get('/user/privacy/settings'),
  // 更新隐私设置
  updatePrivacySettings: (data) => api.put('/user/privacy/settings', data),
  // 获取个人标签
  getUserTags: () => api.get('/user/privacy/tags'),
  // 添加个人标签
  addUserTag: (tagName) => api.post('/user/privacy/tags', null, { params: { tagName } }),
  // 删除个人标签
  removeUserTag: (tagName) => api.delete('/user/privacy/tags', { params: { tagName } })
}

export default userPrivacyApi
