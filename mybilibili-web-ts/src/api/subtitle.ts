import api from './client'

export const subtitleApi = {
  getSubtitles(videoId: number) {
    return api.get(`/subtitle/video/${videoId}`)
  },

  getSubtitle(videoId: number, language: string) {
    return api.get(`/subtitle/video/${videoId}/${language}`)
  },

  uploadSubtitle(data: any) {
    return api.post('/subtitle/upload', data)
  },

  uploadSrt(videoId: number, srtContent: string, language: string, languageName: string, uploadedBy: string) {
    return api.post('/subtitle/upload-srt', { videoId, srtContent, language, languageName, uploadedBy })
  },

  deleteSubtitle(subtitleId: number) {
    return api.delete(`/subtitle/${subtitleId}`)
  },

  setDefaultSubtitle(videoId: number, language: string) {
    return api.post('/subtitle/set-default', { videoId, language })
  }
}

export default subtitleApi

export const getVideosWithSubtitleInfo = () => api.get('/subtitle/videos')
export const getVideoSubtitles = (videoId: number) => api.get(`/subtitle/video/${videoId}`)
export const uploadSubtitle = (videoId: number, file: File, language: string, isDefault: boolean) => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('language', language)
  formData.append('isDefault', String(isDefault))
  return api.post(`/subtitle/upload?videoId=${videoId}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const importSrtToMongo = (videoId: number, srtFilePath: string, language: string, isDefault: boolean) =>
  api.post('/subtitle/import-srt', { videoId, srtFilePath, language, isDefault })
export const setDefaultSubtitle = (subtitleId: number) => api.post(`/subtitle/${subtitleId}/set-default`)
export const deleteSubtitle = (subtitleId: number) => api.delete(`/subtitle/${subtitleId}`)
export const getPendingSubtitles = () => api.get('/subtitle/pending')
export const approveSubtitle = (subtitleId: number) => api.post(`/subtitle/${subtitleId}/approve`)
export const rejectSubtitle = (subtitleId: number, reason: string) => api.post(`/subtitle/${subtitleId}/reject`, { reason })
export const previewSubtitle = (subtitleId: number) => api.get(`/subtitle/${subtitleId}/preview`)
export const scanSystemSubtitles = (videoId: number) => api.get(`/subtitle/scan/${videoId}`)
export const importSystemSubtitle = (videoId: number, language: string) => api.post('/subtitle/import-system', { videoId, language })