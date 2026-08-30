import request from './request'

export function getAiUsageOverview() {
  return request({ url: '/ai/usage/overview', method: 'get' })
}

export function getAiUsageFeatures() {
  return request({ url: '/ai/usage/features', method: 'get' })
}

export function getAiUsageDaily(days = 7) {
  return request({ url: '/ai/usage/daily', params: { days }, method: 'get' })
}
