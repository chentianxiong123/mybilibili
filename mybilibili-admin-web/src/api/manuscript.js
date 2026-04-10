import request from './request'

export const getPendingManuscripts = () => {
  return request({
    url: '/manuscript/admin/pending',
    method: 'get'
  })
}

export const getProcessingManuscripts = () => {
  return request({
    url: '/manuscript/admin/processing',
    method: 'get'
  })
}

export const getReadyManuscripts = () => {
  return request({
    url: '/manuscript/admin/ready',
    method: 'get'
  })
}

export const getAllManuscripts = () => {
  return request({
    url: '/manuscript/admin/all',
    method: 'get'
  })
}

export const getManuscriptDetail = (manuscriptId) => {
  return request({
    url: `/manuscript/admin/${manuscriptId}`,
    method: 'get'
  })
}

export const approveManuscript = (manuscriptId, reviewerId, reason) => {
  return request({
    url: `/manuscript/admin/approve/${manuscriptId}`,
    method: 'post',
    params: { reviewerId, reason }
  })
}

export const rejectManuscript = (manuscriptId, reviewerId, reason) => {
  return request({
    url: `/manuscript/admin/reject/${manuscriptId}`,
    method: 'post',
    params: { reviewerId, reason }
  })
}

export const publishManuscript = (manuscriptId) => {
  return request({
    url: `/manuscript/admin/publish/${manuscriptId}`,
    method: 'post'
  })
}

export const unpublishManuscript = (manuscriptId) => {
  return request({
    url: `/manuscript/admin/unpublish/${manuscriptId}`,
    method: 'post'
  })
}

export const getManuscriptVideos = (manuscriptId) => {
  return request({
    url: `/manuscript/admin/${manuscriptId}/videos`,
    method: 'get'
  })
}

export const getManuscriptStatistics = () => {
  return request({
    url: '/manuscript/admin/statistics',
    method: 'get'
  })
}

export const retryManuscript = (manuscriptId) => {
  return request({
    url: `/manuscript/admin/retry/${manuscriptId}`,
    method: 'post'
  })
}

export const manualTranscode = (videoId) => {
  return request({
    url: `/manuscript/admin/transcode/${videoId}`,
    method: 'post'
  })
}

export const manualExtractAudio = (videoId) => {
  return request({
    url: `/manuscript/admin/extract-audio/${videoId}`,
    method: 'post'
  })
}

export const manualGenerateSubtitle = (videoId) => {
  return request({
    url: `/manuscript/admin/generate-subtitle/${videoId}`,
    method: 'post'
  })
}

export const manualAiSummary = (videoId) => {
  return request({
    url: `/manuscript/admin/ai-summary/${videoId}`,
    method: 'post',
    timeout: 120000
  })
}

export const manualProcessAll = (videoId) => {
  return request({
    url: `/manuscript/admin/process-all/${videoId}`,
    method: 'post'
  })
}

export const resetVideoStatus = (videoId) => {
  return request({
    url: `/manuscript/admin/reset/${videoId}`,
    method: 'post'
  })
}

export const getVideoProcessStatus = (videoId) => {
  return request({
    url: `/manuscript/admin/video-status/${videoId}`,
    method: 'get'
  })
}

export const getVideoSourceUrl = (videoId) => {
  return request({
    url: `/manuscript/admin/video-source/${videoId}`,
    method: 'get'
  })
}

export const testAiApi = (text) => {
  return request({
    url: '/manuscript/admin/test-ai-api',
    method: 'post',
    data: { text }
  })
}

export const testAiSummary = (videoId) => {
  return request({
    url: `/manuscript/admin/test-ai-summary/${videoId}`,
    method: 'post'
  })
}
