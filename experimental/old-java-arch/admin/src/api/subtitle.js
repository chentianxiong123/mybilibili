import request from './request'

// 获取视频列表（带字幕信息）
export const getVideosWithSubtitleInfo = () => {
  return request.get('/subtitle/videos')
}

// 获取视频的字幕列表
export const getVideoSubtitles = (videoId) => {
  return request.get(`/subtitle/video/${videoId}`)
}

// 上传字幕
export const uploadSubtitle = (videoId, file, language, isDefault) => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('language', language)
  formData.append('isDefault', isDefault)
  return request.post(`/subtitle/upload?videoId=${videoId}`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 导入SRT到MongoDB
export const importSrtToMongo = (videoId, srtFilePath, language, isDefault) => {
  return request.post('/subtitle/import-srt', {
    videoId,
    srtFilePath,
    language,
    isDefault
  })
}

// 设置默认字幕
export const setDefaultSubtitle = (subtitleId) => {
  return request.post(`/subtitle/${subtitleId}/set-default`)
}

// 删除字幕
export const deleteSubtitle = (subtitleId) => {
  return request.delete(`/subtitle/${subtitleId}`)
}

// 获取待审核字幕列表
export const getPendingSubtitles = () => {
  return request.get('/subtitle/pending')
}

// 审核通过字幕
export const approveSubtitle = (subtitleId) => {
  return request.post(`/subtitle/${subtitleId}/approve`)
}

// 审核拒绝字幕
export const rejectSubtitle = (subtitleId, reason) => {
  return request.post(`/subtitle/${subtitleId}/reject`, { reason })
}

// 预览字幕
export const previewSubtitle = (subtitleId) => {
  return request.get(`/subtitle/${subtitleId}/preview`)
}

// 扫描系统字幕文件
export const scanSystemSubtitles = (videoId) => {
  return request.get(`/subtitle/scan/${videoId}`)
}

// 系统字幕入库
export const importSystemSubtitle = (videoId, language) => {
  return request.post('/subtitle/import-system', {
    videoId,
    language
  })
}
