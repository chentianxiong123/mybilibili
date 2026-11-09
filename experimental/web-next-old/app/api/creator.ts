import api from './client'

export const creatorApi = {
  getMyFollowers: () => api.get(`/follow/me/followers`),
  getMyFollowing: () => api.get(`/follow/me/following`),
  getFans: (userId: number) => api.get(`/user/${userId}/followers`),
  getFollowing: (userId: number) => api.get(`/user/${userId}/following`),
  checkFollow: (userId: number) => api.get(`/follow/check/${userId}`),
  getComments: (params: any) => api.get('/creator/comments', { params }),
  deleteComment: (commentId: number) => api.delete(`/creator/comments/${commentId}`),
  deleteReply: (replyId: number) => api.delete(`/creator/comments/reply/${replyId}`),
  replyComment: (commentId: number, content: string, replyToUserId: number) => api.post(`/creator/comments/${commentId}/reply`, null, { params: { content, replyToUserId } }),
  getDanmakuList: (params: any) => api.get('/creator/danmaku/list', { params }),
  deleteDanmaku: (danmakuId: number) => api.delete(`/creator/danmaku/${danmakuId}`)
}

export const manuscriptApi = {
  getMyManuscripts: (params: any) => api.get(`/manuscript/me/list`, { params }),
  getMyStats: () => api.get(`/manuscript/me/stats`),
  getUserManuscripts: (userId: number, params: any) => api.get(`/manuscript/user/${userId}`, { params }),
  getManuscriptById: (id: number) => api.get(`/manuscript/${id}`),
  getMyManuscriptById: (id: number) => api.get(`/manuscript/${id}`),
  updateManuscript: (id: number, data: any) => {
    const formData = new FormData()
    formData.append('title', data.title)
    formData.append('description', data.description || '')
    formData.append('categoryId', data.categoryId)
    if (data.cover) {
      formData.append('cover', data.cover)
    }
    if (data.tags && data.tags.length > 0) {
      data.tags.forEach((tag: string) => formData.append('tags', tag))
    }
    if (data.videos && data.videos.length > 0) {
      data.videos.forEach((video: any, index: number) => {
        formData.append(`videos[${index}].id`, video.id)
        formData.append(`videos[${index}].title`, video.title || `P${index + 1}`)
        formData.append(`videos[${index}].videoOrder`, video.videoOrder ?? index)
        formData.append(`videos[${index}].durationSeconds`, video.durationSeconds || 0)
        if (video.file) {
          formData.append(`videos[${index}].file`, video.file)
        }
      })
    }
    return api.put(`/manuscript/${id}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  deleteManuscript: (id: number) => api.delete(`/manuscript/${id}`),
  unpublishManuscript: (id: number) => api.post(`/manuscript/${id}/unpublish`),
  publishManuscript: (id: number) => api.post(`/manuscript/${id}/publish`)
}

export const collectionApi = {
  getUserCollections: (userId: number, page = 1, size = 100) => api.get(`/collection/user/${userId}?page=${page}&size=${size}`),
  getCollectionById: (collectionId: number) => api.get(`/collection/${collectionId}`),
  getCollectionManuscripts: (collectionId: number, page = 1, size = 20) => api.get(`/collection/${collectionId}/manuscripts?page=${page}&size=${size}`),
  createCollection: (data: any) => {
    const formData = new FormData()
    formData.append('name', data.name)
    if (data.description) formData.append('description', data.description)
    if (data.cover) formData.append('cover', data.cover)
    formData.append('isPublic', String(data.isPublic !== false))
    return api.post('/collection', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  updateCollection: (id: number, data: any) => {
    const formData = new FormData()
    if (data.name) formData.append('name', data.name)
    if (data.description !== undefined) formData.append('description', data.description)
    if (data.cover) formData.append('cover', data.cover)
    if (data.isPublic !== undefined) formData.append('isPublic', data.isPublic)
    return api.put(`/collection/${id}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  deleteCollection: (id: number) => api.delete(`/collection/${id}`),
  addManuscriptToCollection: (collectionId: number, manuscriptId: number, order = 0) =>
    api.post(`/collection/${collectionId}/manuscript/${manuscriptId}?order=${order}`),
  removeManuscriptFromCollection: (collectionId: number, manuscriptId: number) =>
    api.delete(`/collection/${collectionId}/manuscript/${manuscriptId}`)
}

export const followApi = {
  follow: (userId: number) => api.post(`/follow/${userId}`),
  unfollow: (userId: number) => api.delete(`/follow/${userId}`),
  checkFollow: (userId: number) => api.get(`/follow/check/${userId}`)
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