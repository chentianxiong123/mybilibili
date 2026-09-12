import api from './client'

export const dynamicApi = {
  getDynamicList: (page = 1, size = 10) => {
    return api.get('/dynamic/list', { params: { page, size } })
  },

  getFollowingDynamics: (page = 1, size = 10, userId: number | null = null) => {
    const params: Record<string, any> = { page, size }
    if (userId) {
      params.userId = userId
    }
    return api.get('/dynamic/following', { params })
  },

  getUserDynamics: (userId: number, page = 1, limit = 10) => {
    return api.get(`/dynamic/user/${userId}`, { params: { page, limit } })
  },

  getDynamicById: (id: number) => {
    return api.get(`/dynamic/${id}`)
  },

  publishDynamic: (formData: FormData) => {
    return api.post('/dynamic/publish', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  deleteDynamic: (id: number) => {
    return api.delete(`/dynamic/${id}`)
  },

  likeDynamic: (id: number) => {
    return api.post(`/dynamic/like/${id}`)
  },

  unlikeDynamic: (id: number) => {
    return api.delete(`/dynamic/like/${id}`)
  },

  shareDynamic: (id: number) => {
    return api.post(`/dynamic/share/${id}`)
  }
}

export default dynamicApi