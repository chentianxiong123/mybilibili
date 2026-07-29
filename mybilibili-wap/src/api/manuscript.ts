import api from './client'

const unwrapData = (res: any) => res?.data ?? res

export const manuscriptApi = {
  getMyManuscripts(params: Record<string, any> = {}) {
    return api.get('/manuscript/me/list', { params })
  },

  getMyStats() {
    return api.get('/manuscript/me/stats')
  },

  publish(id: number) {
    return api.post(`/manuscript/${id}/publish`)
  },

  unpublish(id: number) {
    return api.post(`/manuscript/${id}/unpublish`)
  },

  remove(id: number) {
    return api.delete(`/manuscript/${id}`)
  }
}

export const normalizeManuscriptList = (res: any) => {
  const data = unwrapData(res)
  const list = Array.isArray(data) ? data : (data?.list || [])
  return {
    list,
    total: Number(data?.total ?? list.length),
    page: Number(data?.page ?? 1),
    totalPages: Number(data?.totalPages ?? 1)
  }
}
