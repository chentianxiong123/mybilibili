import api from './client'
import { getAdminToken } from '../utils/auth'

// 转码流水线看板 API（对齐旧版 admin-web videoProcess.js）
export const getCurrentTask = () => api.get('/video/process/admin/current')

export const getQueueInfo = () => api.get('/video/process/admin/queue')

export const getStatistics = () => api.get('/video/process/admin/statistics')

// SSE 推流地址（EventSource 无法走 axios 实例，也就无法设置 Authorization 头，
// 只能把凭证放在 query 上，由后端 TokenFromRequest 的 access_token 分支解析）。
// 阶段 1 后台登录改发 HttpOnly Cookie 后，EventSource 会自动带 Cookie，
// 届时可以去掉这个参数。
export const getStreamUrl = () => {
  const token = getAdminToken()
  return `/api/v1/video/process/admin/stream${token ? `?access_token=${encodeURIComponent(token)}` : ''}`
}
