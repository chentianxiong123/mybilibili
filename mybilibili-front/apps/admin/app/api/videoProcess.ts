import api from './client'

// 转码流水线看板 API（对齐旧版 admin-web videoProcess.js）
export const getCurrentTask = () => api.get('/video/process/admin/current')

export const getQueueInfo = () => api.get('/video/process/admin/queue')

export const getStatistics = () => api.get('/video/process/admin/statistics')

// SSE 推流地址（EventSource 无法走 axios 实例，用相对路径同源直连）
export const getStreamUrl = () => '/api/v1/video/process/admin/stream'
