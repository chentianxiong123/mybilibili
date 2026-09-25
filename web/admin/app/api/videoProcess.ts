import api from './client'

// 转码流水线看板 API（对齐旧版 admin-web videoProcess.js）
export const getCurrentTask = () => api.get('/video/process/admin/current')

export const getQueueInfo = () => api.get('/video/process/admin/queue')

export const getStatistics = () => api.get('/video/process/admin/statistics')

// SSE 推流地址。EventSource 虽然设不了请求头，但同源请求会自动带 HttpOnly cookie，
// 所以这里**不再也不能**把令牌拼进 URL——拼进 URL 就意味着令牌会进访问日志。
export const getStreamUrl = () => '/api/v1/video/process/admin/stream'
