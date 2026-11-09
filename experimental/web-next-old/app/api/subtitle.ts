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

  uploadSrt(videoId: number, file: File, language: string, languageName: string, isDefault: boolean) {
    const formData = new FormData()
    formData.append('video_id', String(videoId))
    formData.append('file', file)
    formData.append('language', language)
    formData.append('language_name', languageName)
    if (isDefault) formData.append('is_default', 'true')
    return api.post('/subtitle/upload-srt', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  deleteSubtitle(subtitleId: number) {
    return api.delete(`/subtitle/${subtitleId}`)
  },

  setDefaultSubtitle(videoId: number, subtitleId: string) {
    return api.post('/subtitle/set-default', { video_id: videoId, id: subtitleId })
  }
}

export default subtitleApi

export const getVideosWithSubtitleInfo = () => api.get('/subtitle/videos')
export const getVideoSubtitles = (videoId: number) => api.get(`/subtitle/video/${videoId}`)
export const uploadSubtitle = (videoId: number, file: File, language: string, isDefault: boolean) => {
  const formData = new FormData()
  formData.append('video_id', String(videoId))
  formData.append('file', file)
  formData.append('language', language)
  formData.append('language_name', language === 'zh-CN' ? '中文' : language)
  if (isDefault) formData.append('is_default', 'true')
  return api.post('/subtitle/upload-srt', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const importSrtToMongo = (videoId: number, srtFilePath: string, language: string, isDefault: boolean) =>
  api.post('/subtitle/import-srt', { video_id: videoId, srt: srtFilePath })
export const setDefaultSubtitle = (subtitleId: number, videoId: number) => api.post(`/subtitle/${subtitleId}/set-default?video_id=${videoId}`)
export const deleteSubtitle = (subtitleId: number) => api.delete(`/subtitle/${subtitleId}`)
export const getPendingSubtitles = () => api.get('/subtitle/pending')
export const approveSubtitle = (subtitleId: number) => api.post(`/subtitle/${subtitleId}/approve`)
export const rejectSubtitle = (subtitleId: number, reason: string) => api.post(`/subtitle/${subtitleId}/reject`, { reason })
export const previewSubtitle = (subtitleId: number) => api.get(`/subtitle/${subtitleId}/preview`)
export const scanSystemSubtitles = (videoId: number) => api.get(`/subtitle/scan/${videoId}`)
export const importSystemSubtitle = (videoId: number, srtContent: string) => api.post('/subtitle/import-system', { video_id: videoId, srt: srtContent })