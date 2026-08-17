import api from './client'

export const manuscriptApi = {
  uploadManuscript: (manuscriptData: any, onProgress?: (pct: number) => void) => {
    const formData = new FormData()

    formData.append('title', manuscriptData.title)
    formData.append('description', manuscriptData.description || '')
    formData.append('cover', manuscriptData.cover)
    formData.append('categoryId', manuscriptData.categoryId)

    if (manuscriptData.tags && manuscriptData.tags.length > 0) {
      manuscriptData.tags.forEach((tag: string) => {
        formData.append('tags', tag)
      })
    }

    if (manuscriptData.videos && manuscriptData.videos.length > 0) {
      manuscriptData.videos.forEach((video: any, index: number) => {
        formData.append(`videos[${index}].file`, video.file)
        formData.append(`videos[${index}].title`, video.title || `P${index + 1}`)
        formData.append(`videos[${index}].videoOrder`, video.sortOrder || index)
        formData.append(`videos[${index}].durationSeconds`, video.durationSeconds || 0)
      })
    }

    return api.post('/manuscript/upload', formData, {
      timeout: 300000,
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: onProgress ? (progressEvent: any) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(percentCompleted)
      } : undefined
    })
  },

  createUploadSession: (payload: any) => {
    return api.post('/manuscript/upload-session', payload, {
      headers: { 'Content-Type': 'application/json' }
    })
  },

  getUploadSessionStatus: (uploadId: string) => {
    return api.get(`/manuscript/upload-session/${uploadId}`)
  },

  uploadChunk: (chunkData: any, onProgress?: (p: any) => void) => {
    const formData = new FormData()
    formData.append('uploadId', chunkData.uploadId)
    formData.append('partIndex', chunkData.partIndex)
    formData.append('chunkIndex', chunkData.chunkIndex)
    formData.append('totalChunks', chunkData.totalChunks)
    formData.append('file', chunkData.file)
    return api.post('/manuscript/upload-chunk', formData, {
      timeout: 120000,
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: onProgress ? (progressEvent: any) => {
        const total = progressEvent.total || chunkData.file.size || 0
        const loaded = Math.min(progressEvent.loaded || 0, total)
        const percentCompleted = total > 0 ? Math.round((loaded * 100) / total) : 0
        onProgress({ percent: percentCompleted, loaded, total })
      } : undefined
    })
  },

  completeUploadSession: (uploadId: string, cover: File, onProgress?: (pct: number) => void) => {
    const formData = new FormData()
    formData.append('uploadId', uploadId)
    formData.append('cover', cover)
    return api.post('/manuscript/upload-complete', formData, {
      timeout: 600000,
      onUploadProgress: onProgress ? (progressEvent: any) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(percentCompleted)
      } : undefined
    })
  },

  cancelUploadSession: (uploadId: string) => {
    return api.delete(`/manuscript/upload-session/${uploadId}`)
  },

  getManuscriptList: (page = 1, size = 10, status: string | null = null) => {
    let url = `/manuscript/list?page=${page}&size=${size}`
    if (status) {
      url += `&status=${status}`
    }
    return api.get(url)
  },

  getManuscriptById: (id: number) => {
    return api.get(`/manuscript/${id}`)
  },

  getRecommendedManuscripts: () => {
    return api.get('/manuscript/recommended')
  },

  updateManuscript: (id: number, manuscriptData: any) => {
    const formData = new FormData()

    if (manuscriptData.title) {
      formData.append('title', manuscriptData.title)
    }
    if (manuscriptData.description !== undefined) {
      formData.append('description', manuscriptData.description)
    }
    if (manuscriptData.cover) {
      formData.append('cover', manuscriptData.cover)
    }
    if (manuscriptData.categoryId) {
      formData.append('categoryId', manuscriptData.categoryId)
    }
    if (manuscriptData.tags) {
      manuscriptData.tags.forEach((tag: string) => {
        formData.append('tags', tag)
      })
    }

    return api.put(`/manuscript/${id}`, formData)
  },

  deleteManuscript: (id: number) => {
    return api.delete(`/manuscript/${id}`)
  },

  getUserManuscripts: (userId: number, page = 1, size = 10, status: string | null = null) => {
    let url = `/manuscript/user/${userId}?page=${page}&size=${size}`
    if (status !== null) {
      url += `&status=${status}`
    }
    return api.get(url)
  },

  getManuscriptStats: (userId: number) => {
    return api.get(`/manuscript/user/${userId}/stats`)
  }
}

export default manuscriptApi

export const getPendingManuscripts = () => api.get('/manuscript/admin/pending')
export const getProcessingManuscripts = () => api.get('/manuscript/admin/processing')
export const getAllManuscripts = () => api.get('/manuscript/admin/all')
export const getManuscriptDetail = (manuscriptId: number) => api.get(`/manuscript/admin/${manuscriptId}`)
export const approveManuscript = (manuscriptId: number, reviewerId: number, reason: string) => api.post(`/manuscript/admin/approve/${manuscriptId}`, { reviewerId, reason })
export const approveWithProcess = (manuscriptId: number, autoProcess = false) => api.post(`/manuscript/admin/${manuscriptId}/approve-with-process`, { autoProcess })
export const rejectManuscript = (manuscriptId: number, reviewerId: number, reason: string) => api.post(`/manuscript/admin/reject/${manuscriptId}`, { reviewerId, reason })
export const publishManuscript = (manuscriptId: number) => api.post(`/manuscript/admin/publish/${manuscriptId}`)
export const unpublishManuscript = (manuscriptId: number) => api.post(`/manuscript/admin/unpublish/${manuscriptId}`)
export const getManuscriptVideos = (manuscriptId: number) => api.get(`/manuscript/admin/${manuscriptId}/videos`)
export const getManuscriptStatistics = () => api.get('/manuscript/admin/statistics')
export const retryManuscript = (manuscriptId: number) => api.post(`/manuscript/admin/retry/${manuscriptId}`)
export const manualTranscode = (videoId: number) => api.post(`/manuscript/admin/transcode/${videoId}`)
export const manualExtractAudio = (videoId: number) => api.post(`/manuscript/admin/extract-audio/${videoId}`)
export const manualGenerateSubtitle = (videoId: number) => api.post(`/manuscript/admin/generate-subtitle/${videoId}`)
export const manualAiSummary = (videoId: number) => api.post(`/manuscript/admin/ai-summary/${videoId}`)
export const manualProcessAll = (videoId: number) => api.post(`/manuscript/admin/process-all/${videoId}`)
export const resetVideoStatus = (videoId: number) => api.post(`/manuscript/admin/reset/${videoId}`)
export const getVideoSourceUrl = (videoId: number) => api.get(`/manuscript/admin/video-source/${videoId}`)