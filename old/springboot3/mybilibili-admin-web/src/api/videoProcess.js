import request from './request'

export const getCurrentTask = () => {
  return request.get('/ai/admin/process/current')
}

export const getQueueInfo = () => {
  return request.get('/ai/admin/process/queue')
}

export const getStatistics = () => {
  return request.get('/ai/admin/process/statistics')
}

export const getManuscriptProcessStatus = (manuscriptId) => {
  return request.get(`/ai/admin/process/manuscript/${manuscriptId}/status`)
}

export const triggerTranscode = (videoId) => {
  return request.post(`/ai/admin/process/transcode/${videoId}`)
}

export const triggerAudioExtract = (videoId) => {
  return request.post(`/ai/admin/process/audio/${videoId}`)
}

export const triggerSubtitleGenerate = (videoId) => {
  return request.post(`/ai/admin/process/subtitle/${videoId}`)
}

export const triggerAiSummary = (videoId) => {
  return request.post(`/ai/admin/process/ai-summary/${videoId}`)
}

export const resetVideoProcess = (videoId) => {
  return request.post(`/ai/admin/process/reset/${videoId}`)
}
