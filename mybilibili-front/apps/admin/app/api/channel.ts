import request from './client'

function toSnake(data: Record<string, any>): Record<string, any> {
  const out: Record<string, any> = {}
  for (const [k, v] of Object.entries(data)) {
    out[k.replace(/([A-Z])/g, '_$1').toLowerCase()] = v
  }
  return out
}

// ========== 渠道管理 ==========
export function getChannels() {
  return request({ url: '/ai/configs', method: 'get' })
}

export function getChannel(id) {
  return request({ url: `/ai/configs/${id}`, method: 'get' })
}

export function createChannel(data) {
  return request({ url: '/ai/configs', method: 'post', data: toSnake(data) })
}

export function updateChannel(id, data) {
  return request({ url: `/ai/configs/${id}`, method: 'put', data: toSnake(data) })
}

export function deleteChannel(id) {
  return request({ url: `/ai/configs/${id}`, method: 'delete' })
}

export function toggleChannel(id) {
  return request({ url: `/ai/configs/${id}/toggle`, method: 'put' })
}

// ========== 按类型查询 ==========
export function getChannelsByType(type) {
  return request({ url: `/ai/configs?type=${type}`, method: 'get' })
}

// ========== 功能绑定 ==========
export function getBindings() {
  return request({ url: '/ai/bindings', method: 'get' })
}

export function bindFeature(feature, configId) {
  return request({ url: `/ai/bindings/${feature}`, method: 'post', data: { configId } })
}

// ========== 测试连接 ==========
export function testConnection(data) {
  return request({ url: '/ai/config/test', method: 'post', data: toSnake(data) })
}

// ========== 可用类型和功能 ==========
export function getAvailableTypes() {
  return request({ url: '/ai/configs/types', method: 'get' })
}

export function getAvailableFeatures() {
  return request({ url: '/ai/configs/features', method: 'get' })
}