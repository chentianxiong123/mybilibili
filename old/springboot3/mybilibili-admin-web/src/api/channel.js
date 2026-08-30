import request from './request'

// ========== 渠道管理 ==========
export function getChannels() {
  return request({ url: '/ai/admin/channels', method: 'get' })
}

export function getChannel(id) {
  return request({ url: `/ai/admin/channels/${id}`, method: 'get' })
}

export function createChannel(data) {
  return request({ url: '/ai/admin/channels', method: 'post', data })
}

export function updateChannel(id, data) {
  return request({ url: `/ai/admin/channels/${id}`, method: 'put', data })
}

export function deleteChannel(id) {
  return request({ url: `/ai/admin/channels/${id}`, method: 'delete' })
}

export function toggleChannel(id) {
  return request({ url: `/ai/admin/channels/${id}/toggle`, method: 'put' })
}

// ========== 按类型查询 ==========
export function getChannelsByType(type) {
  return request({ url: `/ai/admin/channels/type/${type}`, method: 'get' })
}

// ========== 功能绑定 ==========
export function getBindings() {
  return request({ url: '/ai/admin/channels/bindings', method: 'get' })
}

export function bindFeature(feature, configId) {
  return request({ url: `/ai/admin/channels/bindings/${feature}`, method: 'post', data: { configId } })
}

// ========== 测试连接 ==========
export function testConnection(data) {
  return request({ url: '/ai/admin/config/test', method: 'post', data })
}

// ========== 可用类型和功能 ==========
export function getAvailableTypes() {
  return request({ url: '/ai/admin/channels/types', method: 'get' })
}

export function getAvailableFeatures() {
  return request({ url: '/ai/admin/channels/features', method: 'get' })
}