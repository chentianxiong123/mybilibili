import request from './request'

export const getPendingManuscripts = () => {
  return request({
    url: '/manuscript/pending',
    method: 'get'
  })
}

export const getProcessingManuscripts = () => {
  return request({
    url: '/manuscript/processing',
    method: 'get'
  })
}

export const getReadyManuscripts = () => {
  return request({
    url: '/manuscript/ready',
    method: 'get'
  })
}

export const getAllManuscripts = () => {
  return request({
    url: '/manuscript/all',
    method: 'get'
  })
}

export const getManuscriptDetail = (manuscriptId) => {
  return request({
    url: `/manuscript/${manuscriptId}`,
    method: 'get'
  })
}

export const approveManuscript = (manuscriptId, reviewerId, reason) => {
  return request({
    url: `/manuscript/approve/${manuscriptId}`,
    method: 'post',
    params: { reviewerId, reason }
  })
}

export const rejectManuscript = (manuscriptId, reviewerId, reason) => {
  return request({
    url: `/manuscript/reject/${manuscriptId}`,
    method: 'post',
    params: { reviewerId, reason }
  })
}

export const publishManuscript = (manuscriptId) => {
  return request({
    url: `/manuscript/publish/${manuscriptId}`,
    method: 'post'
  })
}

export const unpublishManuscript = (manuscriptId) => {
  return request({
    url: `/manuscript/unpublish/${manuscriptId}`,
    method: 'post'
  })
}

export const getManuscriptVideos = (manuscriptId) => {
  return request({
    url: `/manuscript/${manuscriptId}/videos`,
    method: 'get'
  })
}

export const getManuscriptStatistics = () => {
  return request({
    url: '/manuscript/statistics',
    method: 'get'
  })
}

export const retryManuscript = (manuscriptId) => {
  return request({
    url: `/manuscript/retry/${manuscriptId}`,
    method: 'post'
  })
}

// ==================== 手动处理流程 API ====================

export const manualTranscode = (videoId) => {
  return request({
    url: `/manuscript/transcode/${videoId}`,
    method: 'post'
  })
}

export const manualExtractAudio = (videoId) => {
  return request({
    url: `/manuscript/extract-audio/${videoId}`,
    method: 'post'
  })
}

export const manualGenerateSubtitle = (videoId) => {
  return request({
    url: `/manuscript/generate-subtitle/${videoId}`,
    method: 'post'
  })
}

export const manualAiSummary = (videoId) => {
  return request({
    url: `/manuscript/ai-summary/${videoId}`,
    method: 'post'
  })
}

export const manualProcessAll = (videoId) => {
  return request({
    url: `/manuscript/process-all/${videoId}`,
    method: 'post'
  })
}

export const resetVideoStatus = (videoId) => {
  return request({
    url: `/manuscript/reset/${videoId}`,
    method: 'post'
  })
}

// ==================== 视频处理查询 API ====================

export const getVideoProcessStatus = (videoId) => {
  return request({
    url: `/manuscript/video-status/${videoId}`,
    method: 'get'
  })
}

export const getVideoSourceUrl = (videoId) => {
  return request({
    url: `/manuscript/video-source/${videoId}`,
    method: 'get'
  })
}

// ==================== AI摘要测试 API ====================

export const testAiApi = (text) => {
  return request({
    url: '/manuscript/test-ai-api',
    method: 'post',
    data: { text }
  })
}

export const testAiSummary = (videoId) => {
  return request({
    url: `/manuscript/test-ai-summary/${videoId}`,
    method: 'post'
  })
}
