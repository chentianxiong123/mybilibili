import api from './client'

export const dynamicApi = {
  getDynamicList(page = 1, size = 10) {
    return api.get('/dynamic/list', { params: { page, size } })
  },

  getFollowingDynamics(page = 1, size = 10, userId?: number | null) {
    const params: any = { page, size }
    if (userId) {
      params.userId = userId
    }
    return api.get('/dynamic/following', { params })
  },

  getUserDynamics(userId: number, page = 1, limit = 10) {
    return api.get(`/dynamic/user/${userId}`, { params: { page, limit } })
  },

  getDynamicById(id: number) {
    return api.get(`/dynamic/${id}`)
  },

  publishDynamic(formData: FormData) {
    return api.post('/dynamic/publish', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  deleteDynamic(id: number) {
    return api.delete(`/dynamic/${id}`)
  },

  likeDynamic(id: number) {
    return api.post(`/dynamic/like/${id}`)
  },

  unlikeDynamic(id: number) {
    return api.delete(`/dynamic/like/${id}`)
  },

  checkLikeStatus(id: number) {
    return api.get(`/dynamic/like/status/${id}`)
  },

  shareDynamic(id: number) {
    return api.post(`/dynamic/share/${id}`)
  },

  getComments(dynamicId: number, page = 1, limit = 10) {
    return api.get(`/dynamic/${dynamicId}/comments`, { params: { page, limit } })
  },

  addComment(dynamicId: number, content: string, parentId?: number | null, replyUserId?: number | null) {
    const params: any = { content }
    if (parentId) params.parentId = parentId
    if (replyUserId) params.replyUserId = replyUserId
    return api.post(`/dynamic/${dynamicId}/comment`, null, { params })
  },

  deleteComment(commentId: number) {
    return api.delete(`/dynamic/comment/${commentId}`)
  },

  likeComment(commentId: number) {
    return api.post(`/dynamic/comment/like/${commentId}`)
  },

  unlikeComment(commentId: number) {
    return api.delete(`/dynamic/comment/like/${commentId}`)
  }
}

export default dynamicApi
