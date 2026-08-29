import axios from 'axios'
import { safeStorage } from '../utils/safeStorage'
import { ElMessage } from 'element-plus'
import {
  clearAuthSession,
  getRefreshToken,
  getToken,
  setAuthSession,
  getAdminToken,
  clearAdminSession
} from '../utils/auth'

// CSR(浏览器): 相对路径 → 走 traefik 80 (统一入口)
// SSR(服务器, 容器内): 直接访问 core 容器 (docker 与 k8s 服务名通用)
const api = axios.create({
  baseURL: process.server ? 'http://core:8080/api/v1' : '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true
}) as any

// ====== 短期 GET 缓存（仅公开只读接口，TTL 30s） ======
const cacheStore = new Map<string, { data: any; expires: number }>()
const CACHE_TTL = 30 * 1000
const CACHE_PATHS = ['/category', '/manuscript/recommended', '/manuscript/hot']

const cacheEnabled = (url: string) => CACHE_PATHS.some(p => url.includes(p))
const cacheKeyFor = (method: string, url: string, params: any) =>
  `${method}:${url}:${params ? JSON.stringify(params) : ''}`

const clearCacheFor = (url: string) => {
  if (!cacheEnabled(url)) return
  ;[...cacheStore.keys()].forEach(key => {
    if (key.includes(url)) cacheStore.delete(key)
  })
}

api.interceptors.request.use(
  config => {
    const url = config.url || ''
    const isImageRequest = url.includes('/covers/') || url.includes('/images/') || url.match(/\.(jpg|jpeg|png|gif|mp4)$/i)
    if (isImageRequest) return config

    const adminToken = getAdminToken()
    const userToken = getToken()
    const token = adminToken || userToken
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    if ((config.method || 'get').toLowerCase() === 'get' && cacheEnabled(url)) {
      const key = cacheKeyFor('get', url, config.params)
      const hit = cacheStore.get(key)
      if (hit && hit.expires > Date.now()) {
        config.adapter = async (opts: any) => ({ data: hit.data, status: 200, statusText: 'OK', headers: {}, config: opts, request: {} })
      }
    }
    return config
  },
  error => Promise.reject(error)
)

let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (error: any) => void }> = []

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) prom.reject(error)
    else prom.resolve(token!)
  })
  failedQueue = []
}

api.interceptors.response.use(
  response => {
    const config = response.config
    const url = config?.url || ''
    const method = (config?.method || 'get').toLowerCase()
    if (method === 'get' && cacheEnabled(url) && !(config as any).fromCache) {
      cacheStore.set(cacheKeyFor('get', url, config.params), { data: response.data, expires: Date.now() + CACHE_TTL })
    }
    if (method !== 'get') {
      clearCacheFor(url)
    }
    return response.data
  },
  error => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      const refreshToken = getRefreshToken()
      const adminToken = getAdminToken()

      if (adminToken) {
        clearAdminSession()
        window.location.href = '/admin/login'
        return Promise.reject(error)
      }

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
              const { token, refresh_token: newRefreshToken } = res.data
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

    if (error.response) {
      switch (error.response.status) {
        case 401:
          if (getToken() && getRefreshToken()) {
            clearAuthSession()
          }
          if (import.meta.client) {
            return Promise.resolve({ code: 401, data: error.response.data?.data || [], message: '请先登录' })
          }
          break
        case 403: if (import.meta.client) ElMessage.error('没有权限访问该资源'); break
        case 404: if (import.meta.client) ElMessage.error('请求的资源不存在'); break
        case 500: if (import.meta.client) ElMessage.error('服务器内部错误'); break
        default: if (import.meta.client) ElMessage.error(error.response.data?.message || '请求失败')
      }
    } else {
      if (import.meta.client) ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

function decodeJwtPayload(token: string) {
  try {
    const part = token.split('.')[1]
    if (!part) return null
    const b64 = part.replace(/-/g, '+').replace(/_/g, '/')
    const pad = b64.length % 4
    const padded = pad ? b64 + '='.repeat(4 - pad) : b64
    return JSON.parse(atob(padded))
  } catch (e) {
    return null
  }
}

function getAccessTokenRemainingMs() {
  const token = getToken()
  if (!token) return -1
  const payload = decodeJwtPayload(token)
  if (!payload || !payload.exp) return -1
  return payload.exp * 1000 - Date.now()
}

async function silentRefreshOnce() {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    stopSilentRefresh()
    return false
  }
  try {
    const res = await api.post('/user/token/refresh', { refreshToken })
    if (res && res.code === 200 && res.data && res.data.token) {
      setAuthSession({
        token: res.data.token,
        refreshToken: res.data.refresh_token || refreshToken
      })
      return true
    }
    clearAuthSession()
    stopSilentRefresh()
    return false
  } catch (e) {
    return false
  }
}

let silentRefreshTimer: ReturnType<typeof setInterval> | null = null

export function startSilentRefresh() {
  if (silentRefreshTimer) return
  silentRefreshTimer = setInterval(async () => {
    if (!getRefreshToken()) {
      stopSilentRefresh()
      return
    }
    const remaining = getAccessTokenRemainingMs()
    if (remaining < 10 * 60 * 1000) {
      await silentRefreshOnce()
    }
  }, 60 * 1000)
}

export function stopSilentRefresh() {
  if (silentRefreshTimer) {
    clearInterval(silentRefreshTimer)
    silentRefreshTimer = null
  }
}

export const captchaApi = {
  newCaptcha: () => api.post('/captcha/new'),
  verifyCaptcha: (captchaId: string, answer: string) => api.post('/captcha/verify', { captchaId, answer })
}

export const emailCodeApi = {
  sendCode: (email: string) => api.post('/user/email/code', { email }),
  verifyCode: (email: string, code: string) => api.post('/user/email/verify', { email, code })
}

export const userApi = {
  register: (userData: any) => api.post('/user/register', userData),
  login: (username: string, password: string, loginType: string, email: string, emailCode: string) => {
    const data: Record<string, any> = {}
    if (loginType === 'email_code') {
      data.loginType = 'email_code'
      data.email = email
      data.emailCode = emailCode
    } else {
      data.username = username
      data.password = password
    }
    data.loginIp = safeStorage.getItem('clientIp') || ''
    return api.post('/user/login', data)
  },
  forgotPassword: (email: string, code: string, newPassword: string) => api.post('/user/password/forgot', { email, code, newPassword }),
  getLoginLogs: (page: number, size: number) => api.get('/user/login-logs', { params: { page, size } }),
  getUserById: (id: number) => api.get(`/user/${id}`),
  updateUser: (id: number, userData: any) => api.put(`/user/${id}`, userData),
  uploadAvatar: (id: number, formData: FormData) => api.post(`/user/${id}/avatar`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  follow: (userId: number, follow: boolean) => follow ? api.post(`/follow/${userId}`) : api.delete(`/follow/${userId}`),
  checkFollow: (userId: number) => api.get(`/follow/check/${userId}`),
  getFollowingList: (userId: number) => api.get(`/follow/user/${userId}/following`),
  getFollowerList: (userId: number) => api.get(`/follow/user/${userId}/followers`),
  getPinnedVideo: () => api.get('/user/pinned-video'),
  setPinnedVideo: (videoId: number) => api.post(`/user/pinned-video`, { videoId }),
  removePinnedVideo: () => api.delete(`/user/pinned-video`)
}

export const adminLoginLogApi = {
  getLoginLogs: (params: any) => api.get('/admin/login-logs/list', { params }),
  getUserLoginLogs: (userId: number, page: number, size: number) => api.get(`/admin/login-logs/user/${userId}`, { params: { page, size } })
}

export const videoApi = {
  getRecommendedVideos: () => api.get('/manuscript/recommended'),
  getHotVideos: () => api.get('/manuscript/hot'),
  getVideoById: (id: number) => api.get(`/manuscript/${id}`),
  getVideoByManuscriptId: (manuscriptId: number, params: any) => api.get(`/manuscript/${manuscriptId}`, { params }),
  getVideosByCategoryId: (id: number) => api.get(`/manuscript/category/${id}`),
  getVideosByUserId: (id: number, sort: string, status: number) => {
    let url = `/manuscript/user/${id}`
    const params = []
    if (sort) params.push(`sort=${sort}`)
    if (status !== undefined) params.push(`status=${status}`)
    if (params.length > 0) url += `?${params.join('&')}`
    return api.get(url)
  },
  searchUserVideos: (userId: number, keyword: string, sort: string) => api.get(`/manuscript/user/${userId}/search?keyword=${encodeURIComponent(keyword)}${sort ? `&sort=${sort}` : ''}`),
  getVideoList: (page: number, size: number) => api.get(`/manuscript/list?page=${page}&size=${size}`),
  uploadVideo: (formData: FormData, onProgress?: (pct: number) => void) => api.post('/manuscript/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress ? (progressEvent: any) => {
      const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
      onProgress(percentCompleted)
    } : undefined
  })
}

export const commentApi = {
  getComments: (targetType: string, targetId: number, page = 1, size = 20, sort = 'time') =>
    targetType === 'DYNAMIC'
      ? api.get(`/dynamic/comment/list?dynamicId=${targetId}&page=${page}&size=${size}&sort=${sort}`)
      : api.get(`/comment/list?manuscriptId=${targetId}&page=${page}&size=${size}&sort=${sort}`),

  postComment: (targetType: string, targetId: number, content: string) => {
    const encodedContent = encodeURIComponent(content)
    if (targetType === 'DYNAMIC') {
      return api.post(`/dynamic/comment/add?dynamicId=${targetId}&content=${encodedContent}`, null, {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
      }).then((res: any) => {
        if (res && res.code === 200 && Array.isArray(res.data)) res.data = res.data[0] || null
        return res
      })
    } else {
      return api.post(`/comment/add`, `manuscriptId=${targetId}&content=${encodedContent}`, {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
      })
    }
  },

  // ===== 统一评论接口：视频/动态共用一套契约，仅 targetType 分流 =====

  // 回复（顶层评论或楼中楼），动态侧复用 add 接口携带 parentId
  replyTo: (targetType: string, targetId: number, commentId: number, content: string, replyToUserId?: number) => {
    const enc = encodeURIComponent(content)
    if (targetType === 'DYNAMIC') {
      return api.post(`/dynamic/comment/add?dynamicId=${targetId}&content=${enc}&parentId=${commentId}${replyToUserId ? `&replyUserId=${replyToUserId}` : ''}`, null, {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
      }).then((res: any) => {
        if (res && res.code === 200 && Array.isArray(res.data)) res.data = res.data[0] || null
        return res
      })
    }
    return api.post('/comment/reply', `commentId=${commentId}&content=${enc}${replyToUserId ? `&replyToUserId=${replyToUserId}` : ''}`, {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    })
  },

  // 评论/回复点赞与取消（kind 区分顶层评论与回复）
  likeTarget: (targetType: string, kind: 'comment' | 'reply', id: number) => {
    if (targetType === 'DYNAMIC') return api.post(`/dynamic/comment/like/${id}`)
    return kind === 'reply' ? api.post(`/comment/reply/${id}/like`) : api.post(`/comment/${id}/like`)
  },
  unlikeTarget: (targetType: string, kind: 'comment' | 'reply', id: number) => {
    if (targetType === 'DYNAMIC') return api.delete(`/dynamic/comment/like/${id}`)
    return kind === 'reply' ? api.delete(`/comment/reply/${id}/like`) : api.delete(`/comment/${id}/like`)
  },

  // 加载某条评论的回复列表
  getReplies: (targetType: string, commentId: number, page = 1, size = 20) =>
    targetType === 'DYNAMIC'
      ? api.get(`/dynamic/comment/replies?commentId=${commentId}&page=${page}&size=${size}`)
      : api.get(`/comment/${commentId}/replies?page=${page}&size=${size}`),

  replyComment: (commentId: number, content: string, replyToUserId?: number) => api.post('/comment/reply', `commentId=${commentId}&content=${encodeURIComponent(content)}${replyToUserId ? `&replyToUserId=${replyToUserId}` : ''}`, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  }),

  replyDynamicComment: (dynamicId: number, commentId: number, content: string, replyToUserId?: number) => api.post('/dynamic/comment/add', `dynamicId=${dynamicId}&content=${encodeURIComponent(content)}&parentId=${commentId}${replyToUserId ? `&replyUserId=${replyToUserId}` : ''}`, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  }),

  likeComment: (commentId: number) => api.post(`/comment/${commentId}/like`),
  unlikeComment: (commentId: number) => api.delete(`/comment/${commentId}/like`),
  getRepliesByCommentId: (commentId: number, page: number, size: number) => api.get(`/comment/${commentId}/replies?page=${page}&size=${size}`),
  likeReply: (replyId: number) => api.post(`/comment/reply/${replyId}/like`),
  unlikeReply: (replyId: number) => api.delete(`/comment/reply/${replyId}/like`)
}

const requireAuthResult = (data: any) => (
  getToken()
    ? null
    : Promise.resolve({ code: 401, message: '请先登录', data })
)

export const interactionApi = {
  likeManuscript: (manuscriptId: number, liked: boolean) => liked ? api.post(`/manuscript/${manuscriptId}/like`) : api.delete(`/manuscript/${manuscriptId}/like`),
  coinManuscript: (manuscriptId: number, coinCount: number) => api.post(`/manuscript/${manuscriptId}/coin?coinCount=${coinCount}`),
  collectManuscript: (manuscriptId: number, collected: boolean) => collected ? api.post(`/manuscript/${manuscriptId}/collect`) : api.delete(`/manuscript/${manuscriptId}/collect`),
  shareManuscript: (manuscriptId: number, channel?: string) => api.post(`/manuscript/${manuscriptId}/share`, { channel: channel || 'clipboard' }),
  getShareStatistics: (manuscriptId: number) => api.get(`/manuscript/${manuscriptId}/share/statistics`),
  sendDanmaku: (videoId: number, manuscriptId: number, content: string, time: number, color: string, mode: number) => api.post('/danmaku/send', {
    video_id: videoId, manuscript_id: manuscriptId, content, time, color: color || '#ffffff', mode: mode || 0
  }),
  getDanmakus: (videoId: number) => api.get(`/danmaku/video/${videoId}`),
  getInteractionStatus: (manuscriptId: number) => api.get(`/manuscript/${manuscriptId}/status`),
  getLikedVideos: () => requireAuthResult([]) || api.get('/manuscript/user/likes'),
  getCollectedVideos: () => requireAuthResult([]) || api.get('/manuscript/user/collections'),
  getFavoriteFolders: () => requireAuthResult([]) || api.get('/manuscript/favorite/folders'),
  createFavoriteFolder: (folderData: any) => api.post('/manuscript/favorite/folders', folderData),
  updateFavoriteFolder: (folderId: number, name: string) => api.put(`/manuscript/favorite/folders/${folderId}`, { name }),
  deleteFavoriteFolder: (folderId: number) => api.delete(`/manuscript/favorite/folders/${folderId}`),
  getFavoriteFolderVideos: (folderId: number, page = 1, size = 12, sortOrder = 'desc') =>
    api.get(`/manuscript/favorite/folders/${folderId}/videos`, { params: { page, size, sortOrder } }),
  addToFavoriteFolders: (manuscriptId: number, folderIds: number[]) => api.post(`/manuscript/${manuscriptId}/favorite`, { folderIds }),
  removeVideoFromFavoriteFolder: (folderId: number, manuscriptId: number) =>
    api.delete(`/manuscript/favorite/folders/${folderId}/videos/${manuscriptId}`),
  getVideoFavoriteFolders: (manuscriptId: number) => api.get(`/manuscript/${manuscriptId}/favorite/folders`),
  updateVideoFavoriteFolders: (manuscriptId: number, folderIds: number[]) => api.put(`/manuscript/${manuscriptId}/favorite/folders`, { folderIds })
}

export const historyApi = {
  getHistoryList: (page = 1, size = 20) => requireAuthResult([]) || api.get('/watch-history', { params: { page, size } }),
  clearHistory: () => api.delete('/watch-history'),
  deleteHistoryItem: (id: number) => api.delete(`/watch-history/${id}`)
}

export const categoryApi = {
  getCategoryList: () => api.get('/category')
}

export const reportApi = {
  submitReport: (data: any) => api.post('/report/submit', data)
}

export const feedbackApi = {
  // 意见反馈 → 客服工单 POST /api/v1/operation/tickets（后端 source=USER_FEEDBACK）
  submit: (data: { type: string; content: string; contact: string }) =>
    api.post('/operation/tickets', {
      title: data.type,
      content: data.contact
        ? `${data.content}\n\n联系方式：${data.contact}`
        : data.content
    })
}

export const getUserList = (params: Record<string, any>) => api.get('/user/admin/list', { params })
export const getUserById = (id: number) => api.get(`/user/admin/${id}`)
export const updateUserStatus = (id: number, status: string) => api.put(`/user/admin/${id}/status`, { status })
export const resetPassword = (id: number, newPassword: string) => api.put(`/user/admin/${id}/password`, { newPassword })

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