import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true
})

api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    const userStr = localStorage.getItem('user')
    if (userStr) {
      try {
        const user = JSON.parse(userStr)
        if (user && user.id) {
          config.headers['X-User-Id'] = user.id
        }
      } catch (e) {
        console.error('解析用户信息失败:', e)
      }
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          break
        case 403:
          ElMessage.error('没有权限访问该资源')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(error.response.data?.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export const creatorApi = {
  getMyFollowers: () => api.get(`/follow/me/followers`),
  
  getMyFollowing: () => api.get(`/follow/me/following`),
  
  getFans: (userId) => api.get(`/user/${userId}/followers`),
  
  getFollowing: (userId) => api.get(`/user/${userId}/following`),
  
  checkFollow: (userId) => api.get(`/follow/check/${userId}`),
  
  getComments: (params) => api.get('/creator/comments', { params }),
  
  deleteComment: (commentId) => api.delete(`/creator/comments/${commentId}`),

  deleteReply: (replyId) => api.delete(`/creator/comments/reply/${replyId}`),

  replyComment: (commentId, content, replyToUserId) => api.post(`/creator/comments/${commentId}/reply`, null, { params: { content, replyToUserId } }),

  getDanmakuList: (params) => api.get('/creator/danmaku/list', { params }),

  deleteDanmaku: (danmakuId) => api.delete(`/creator/danmaku/${danmakuId}`)
}

export const manuscriptApi = {
  getMyManuscripts: (params) => api.get(`/video/me/list`, { params }),
  
  getMyStats: () => api.get(`/video/me/stats`),
  
  getUserManuscripts: (userId, params) => api.get(`/video/user/${userId}`, { params }),
  
  getManuscriptById: (id) => api.get(`/video/manuscript/${id}`),
  
  updateManuscript: (id, data) => api.put(`/video/manuscript/${id}`, data),
  
  deleteManuscript: (id) => api.delete(`/video/manuscript/${id}`),
  
  unpublishManuscript: (id) => api.post(`/video/manuscript/${id}/unpublish`),
  
  publishManuscript: (id) => api.post(`/video/manuscript/${id}/publish`)
}

export const collectionApi = {
  getUserCollections: (userId, page = 1, size = 100) => api.get(`/collection/user/${userId}?page=${page}&size=${size}`),

  getCollectionById: (collectionId) => api.get(`/collection/${collectionId}`),

  getCollectionManuscripts: (collectionId, page = 1, size = 20) => api.get(`/collection/${collectionId}/manuscripts?page=${page}&size=${size}`),

  createCollection: (data) => {
    const formData = new FormData()
    formData.append('name', data.name)
    if (data.description) formData.append('description', data.description)
    if (data.cover) formData.append('cover', data.cover)
    formData.append('isPublic', data.isPublic !== false)
    return api.post('/collection', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  updateCollection: (id, data) => {
    const formData = new FormData()
    if (data.name) formData.append('name', data.name)
    if (data.description !== undefined) formData.append('description', data.description)
    if (data.cover) formData.append('cover', data.cover)
    if (data.isPublic !== undefined) formData.append('isPublic', data.isPublic)
    return api.put(`/collection/${id}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  deleteCollection: (id) => api.delete(`/collection/${id}`),

  addManuscriptToCollection: (collectionId, manuscriptId, order = 0) =>
    api.post(`/collection/${collectionId}/manuscript/${manuscriptId}?order=${order}`),

  removeManuscriptFromCollection: (collectionId, manuscriptId) =>
    api.delete(`/collection/${collectionId}/manuscript/${manuscriptId}`)
}

export const followApi = {
  follow: (userId) => api.post(`/follow/${userId}`),
  
  unfollow: (userId) => api.delete(`/follow/${userId}`),
  
  checkFollow: (userId) => api.get(`/follow/check/${userId}`)
}

export const statsApi = {
  getOverview: () => api.get('/creator/stats/overview'),

  getTrend: (days = 7) => api.get(`/creator/stats/trend?days=${days}`),

  getRanking: (sortBy = 'views', limit = 10) => api.get(`/creator/stats/ranking?sortBy=${sortBy}&limit=${limit}`),

  getLatestComments: (limit = 5) => api.get(`/creator/stats/latest-comments?limit=${limit}`),

  getFansRanking: (type = 'view', limit = 10) => api.get(`/creator/stats/fans-ranking?type=${type}&limit=${limit}`),

  getFansTrend: (days = 30) => api.get(`/creator/stats/fans-trend?days=${days}`),

  getManuscriptTrend: () => api.get('/creator/stats/manuscript-trend')
}

export default api