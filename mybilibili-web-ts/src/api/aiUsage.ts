import request from './client'

export function getAiUsageOverview() {
  return request({ url: '/ai/admin/usage/overview', method: 'get' })
}

export function getAiUsageFeatures() {
  return request({ url: '/ai/admin/usage/features', method: 'get' })
}

export function getAiUsageDaily(days = 7) {
  return request({ url: '/ai/admin/usage/daily', params: { days }, method: 'get' })
}
