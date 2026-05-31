import axios from 'axios'
import { ElMessage } from 'element-plus'

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
    // 从localStorage获取token
    const token = localStorage.getItem('token')
    // 对图片请求不添加Authorization头
    const isImageRequest = config.url.includes('/covers/') || config.url.includes('/images/') || config.url.match(/\.(jpg|jpeg|png|gif|mp4)$/i)
    if (token && !isImageRequest) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // 添加用户ID到请求头
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

// Token 刷新状态
let isRefreshing = false
let failedQueue = []

const processQueue = (error, token = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    const originalRequest = error.config

    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      const refreshToken = localStorage.getItem('refreshToken')

      if (!refreshToken || originalRequest.url === '/user/token/refresh') {
        localStorage.removeItem('token')
        localStorage.removeItem('refreshToken')
        localStorage.removeItem('user')
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
          .then(res => {
            if (res.code === 200 && res.data) {
              const { token, refreshToken: newRefreshToken } = res.data
              localStorage.setItem('token', token)
              localStorage.setItem('refreshToken', newRefreshToken)
              originalRequest.headers.Authorization = `Bearer ${token}`
              processQueue(null, token)
              resolve(api(originalRequest))
            } else {
              localStorage.removeItem('token')
              localStorage.removeItem('refreshToken')
              localStorage.removeItem('user')
              processQueue(new Error('refresh failed'), null)
              reject(error)
            }
          })
          .catch(err => {
            localStorage.removeItem('token')
            localStorage.removeItem('refreshToken')
            localStorage.removeItem('user')
            processQueue(err, null)
            reject(err)
          })
          .finally(() => {
            isRefreshing = false
          })
      })
    }

    if (error.response) {
      switch (error.response.status) {
        case 401:
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
          ElMessage.error(error.response.data.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

// 验证码相关API
export const captchaApi = {
  // 生成新验证码，返回 captchaId 和算术题
  newCaptcha: () => api.post('/captcha/new'),
  // 校验验证码
  verifyCaptcha: (captchaId, answer) => api.post('/captcha/verify', { captchaId, answer })
}

// 邮箱验证码API
export const emailCodeApi = {
  // 发送邮箱验证码
  sendCode: (email) => api.post('/user/email/code', { email }),
  // 校验邮箱验证码
  verifyCode: (email, code) => api.post('/user/email/verify', { email, code })
}

// 用户相关API
export const userApi = {
  // 注册
  register: (userData) => api.post('/user/register', userData),
  // 登录
  login: (username, password, loginType, email, emailCode) => {
    const data = {}
    if (loginType === 'email_code') {
      data.loginType = 'email_code'
      data.email = email
      data.emailCode = emailCode
    } else {
      data.username = username
      data.password = password
    }
    // 获取客户端IP
    data.loginIp = localStorage.getItem('clientIp') || ''
    return api.post('/user/login', data)
  },
  // 忘记密码
  forgotPassword: (email, code, newPassword) => api.post('/user/password/forgot', { email, code, newPassword }),
  // 获取登录日志
  getLoginLogs: (page, size) => api.get('/user/login-logs', { params: { page, size } }),
  // 获取用户信息
  getUserById: (id) => api.get(`/user/${id}`),
  // 更新用户信息
  updateUser: (id, userData) => api.put(`/user/${id}`, userData),
  // 上传用户头像
  uploadAvatar: (id, formData) => api.post(`/user/${id}/avatar`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  }),
  // 关注/取消关注用户
  follow: (userId, follow) => follow ? api.post(`/follow/${userId}`) : api.delete(`/follow/${userId}`),
  // 检查是否关注用户
  checkFollow: (userId) => api.get(`/follow/check/${userId}`),
  // 获取用户关注列表
  getFollowingList: (userId) => api.get(`/user/${userId}/following`),
  // 获取用户粉丝列表
  getFollowerList: (userId) => api.get(`/user/${userId}/followers`),
  // 获取置顶视频
  getPinnedVideo: (userId) => api.get(`/user/${userId}/pinned-video`),
  // 设置置顶视频
  setPinnedVideo: (videoId) => api.post(`/user/pinned-video`, { videoId }),
  // 取消置顶视频
  removePinnedVideo: () => api.delete(`/user/pinned-video`)
}

// 管理员登录日志API
export const adminLoginLogApi = {
  getLoginLogs: (params) => api.get('/admin/login-logs/list', { params }),
  getUserLoginLogs: (userId, page, size) => api.get(`/admin/login-logs/user/${userId}`, { params: { page, size } })
}

// 视频相关API
export const videoApi = {
  // 获取推荐视频
  getRecommendedVideos: () => api.get('/manuscript/recommended'),
  // 获取热门视频
  getHotVideos: () => api.get('/manuscript/hot'),
  // 获取稿件详情（旧接口，兼容）
  getVideoById: (id) => api.get(`/manuscript/${id}`),
  // 根据稿件ID获取视频（新接口，用于稿件详情页）
  getVideoByManuscriptId: (manuscriptId, params) => api.get(`/manuscript/${manuscriptId}`, { params }),
  // 获取分类视频
  getVideosByCategoryId: (id) => api.get(`/manuscript/category/${id}`),
  // 获取用户视频
  getVideosByUserId: (id, sort, status) => {
    let url = `/manuscript/user/${id}`
    const params = []
    if (sort) params.push(`sort=${sort}`)
    if (status !== undefined) params.push(`status=${status}`)
    if (params.length > 0) url += `?${params.join('&')}`
    return api.get(url)
  },
  // 搜索用户视频
  searchUserVideos: (userId, keyword, sort) => api.get(`/manuscript/user/${userId}/search?keyword=${encodeURIComponent(keyword)}${sort ? `&sort=${sort}` : ''}`),
  // 获取视频列表
  getVideoList: (page, size) => api.get(`/manuscript/list?page=${page}&size=${size}`),
  // 上传视频
  uploadVideo: (formData, onProgress) => api.post('/video/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress: onProgress ? (progressEvent) => {
      const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
      if (onProgress) {
        onProgress(percentCompleted)
      }
      return percentCompleted
    } : undefined
  })
}

// 评论相关API
export const commentApi = {
  // 新接口：获取评论列表（支持视频和动态）
  getComments: (targetType, targetId, page = 1, size = 20, sort = 'time') =>
    targetType === 'DYNAMIC'
      ? api.get(`/dynamic/comment/list?dynamicId=${targetId}&page=${page}&size=${size}&sort=${sort}`)
      : api.get(`/comment/list?manuscriptId=${targetId}&page=${page}&size=${size}&sort=${sort}`),

  // 发表评论（支持视频和动态）
  postComment: (targetType, targetId, content) => {
    const encodedContent = encodeURIComponent(content)
    if (targetType === 'DYNAMIC') {
      return api.post(`/dynamic/comment/add?dynamicId=${targetId}&content=${encodedContent}`, null, {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      })
    } else {
      return api.post(`/comment/add`, `manuscriptId=${targetId}&content=${encodedContent}`, {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      })
    }
  },

  // 向后兼容：获取视频评论
  getCommentsByVideoId: (videoId, page, size, sort = 'new') => api.get(`/comment/video/${videoId}?page=${page}&size=${size}&sort=${sort}`),

  // 向后兼容：发表视频评论
  postVideoComment: (videoId, content) => api.post('/comment/add', `manuscriptId=${videoId}&content=${encodeURIComponent(content)}`, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  }),

  // 回复评论（视频）
  replyComment: (commentId, content, replyToUserId) => api.post('/comment/reply', `commentId=${commentId}&content=${encodeURIComponent(content)}${replyToUserId ? `&replyToUserId=${replyToUserId}` : ''}`, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  }),

  // 回复评论（动态）
  replyDynamicComment: (dynamicId, commentId, content, replyToUserId) => api.post('/dynamic/comment/add', `dynamicId=${dynamicId}&content=${encodeURIComponent(content)}&parentId=${commentId}${replyToUserId ? `&replyUserId=${replyToUserId}` : ''}`, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  }),

  // 点赞评论（视频）
  likeComment: (commentId) => api.post(`/comment/${commentId}/like`),

  // 点赞评论（动态）
  likeDynamicComment: (commentId) => api.post(`/dynamic/comment/like/${commentId}`),

  // 取消点赞评论（视频）
  unlikeComment: (commentId) => api.delete(`/comment/${commentId}/like`),

  // 取消点赞评论（动态）
  unlikeDynamicComment: (commentId) => api.delete(`/dynamic/comment/like/${commentId}`),

  // 获取评论回复（视频）
  getRepliesByCommentId: (commentId, page, size) => api.get(`/comment/${commentId}/replies?page=${page}&size=${size}`),

  // 获取评论回复（动态）
  getDynamicRepliesByCommentId: (commentId) => api.get(`/dynamic/comment/replies?commentId=${commentId}`),

  // 点赞回复（视频）
  likeReply: (replyId) => api.post(`/comment/reply/${replyId}/like`),

  // 点赞回复（动态）
  likeDynamicReply: (replyId) => api.post(`/dynamic/comment/like/${replyId}`),

  // 取消点赞回复（视频）
  unlikeReply: (replyId) => api.delete(`/comment/reply/${replyId}/like`),

  // 取消点赞回复（动态）
  unlikeDynamicReply: (replyId) => api.delete(`/dynamic/comment/like/${replyId}`)
}

// 稿件互动相关API
export const interactionApi = {
  // 点赞稿件
  likeManuscript: (manuscriptId, liked) => liked ? api.post(`/manuscript/${manuscriptId}/like`) : api.delete(`/manuscript/${manuscriptId}/like`),
  // 投币支持
  coinManuscript: (manuscriptId, coinCount) => api.post(`/manuscript/${manuscriptId}/coin?coinCount=${coinCount}`),
  // 收藏稿件
  collectManuscript: (manuscriptId, collected) => collected ? api.post(`/manuscript/${manuscriptId}/collect`) : api.delete(`/manuscript/${manuscriptId}/collect`),
  // 分享稿件
  shareManuscript: (manuscriptId, channel) => api.post(`/manuscript/${manuscriptId}/share`, { channel }),
  // 获取分享统计
  getShareStatistics: (manuscriptId) => api.get(`/manuscript/${manuscriptId}/share/statistics`),
  // 发送弹幕
  sendDanmaku: (videoId, content, time, color, mode) => api.post(`/manuscript/${videoId}/danmaku`, {
    videoId,
    content,
    time,
    color: color || '#ffffff',
    mode: mode || 0
  }),
  // 获取视频弹幕
  getDanmakus: (videoId) => api.get(`/danmaku/video/${videoId}`),
  // 获取互动状态
  getInteractionStatus: (manuscriptId) => api.get(`/manuscript/${manuscriptId}/status`),
  // 获取用户点赞视频
  getLikedVideos: () => api.get('/manuscript/user/likes'),
  // 获取用户收藏视频
  getCollectedVideos: () => api.get('/manuscript/user/collections'),
  // 获取收藏夹列表
  getFavoriteFolders: () => api.get('/manuscript/favorite/folders'),
  // 创建收藏夹
  createFavoriteFolder: (folderData) => api.post('/manuscript/favorite/folders', folderData),
  // 更新收藏夹
  updateFavoriteFolder: (folderId, name) => api.put(`/manuscript/favorite/folders/${folderId}`, { name }),
  // 删除收藏夹
  deleteFavoriteFolder: (folderId) => api.delete(`/manuscript/favorite/folders/${folderId}`),
  // 获取收藏夹内的视频列表
  getFavoriteFolderVideos: (folderId, page = 1, size = 12, sortOrder = 'desc') =>
    api.get(`/manuscript/favorite/folders/${folderId}/videos`, { params: { page, size, sortOrder } }),
  // 添加视频到收藏夹
  addToFavoriteFolders: (manuscriptId, folderIds) => api.post(`/manuscript/${manuscriptId}/favorite`, { folderIds }),
  // 从收藏夹移除视频
  removeVideoFromFavoriteFolder: (folderId, manuscriptId) =>
    api.delete(`/manuscript/favorite/folders/${folderId}/videos/${manuscriptId}`),
  // 获取视频在哪些收藏夹中
  getVideoFavoriteFolders: (manuscriptId) => api.get(`/manuscript/${manuscriptId}/favorite/folders`),
  // 更新视频的收藏夹
  updateVideoFavoriteFolders: (manuscriptId, folderIds) => api.put(`/manuscript/${manuscriptId}/favorite/folders`, { folderIds })
}

// 历史记录相关API
export const historyApi = {
  // 获取历史记录列表
  getHistoryList: (page = 1, size = 20) => api.get('/watch-history', { params: { page, size } }),
  // 清空历史记录
  clearHistory: () => api.delete('/watch-history'),
  // 删除单条历史记录
  deleteHistoryItem: (id) => api.delete(`/watch-history/${id}`)
}

// 分类相关API
export const categoryApi = {
  // 获取分类列表
  getCategoryList: () => api.get('/category')
}

// 举报相关API
export const reportApi = {
  submitReport: (data) => api.post('/report/submit', data)
}

export default api