import request from './request'

export function getAiConfig() {
  return request({
    url: '/ai/admin/config',
    method: 'get'
  })
}

export function updateAiConfig(data) {
  return request({
    url: '/ai/admin/config',
    method: 'put',
    data
  })
}

export function testAiConnection(text) {
  return request({
    url: '/ai/admin/config/test',
    method: 'post',
    data: { text }
  })
}