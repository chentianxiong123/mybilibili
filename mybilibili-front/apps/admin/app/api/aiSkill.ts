import request from './client'

function toSnake(data: Record<string, any>): Record<string, any> {
  const out: Record<string, any> = {}
  for (const [k, v] of Object.entries(data)) {
    out[k.replace(/([A-Z])/g, '_$1').toLowerCase()] = v
  }
  return out
}

export function getAiSkills() {
  return request({
    url: '/ai/skills',
    method: 'get'
  })
}

export function getAiSkillsByType(type) {
  return request({
    url: `/ai/admin/skills/type/${type}`,
    method: 'get'
  })
}

export function getAiSkill(id) {
  return request({
    url: `/ai/admin/skills/${id}`,
    method: 'get'
  })
}

export function createAiSkill(data) {
  return request({
    url: '/ai/skills',
    method: 'post',
    data: toSnake(data)
  })
}

export function updateAiSkill(id, data) {
  return request({
    url: `/ai/admin/skills/${id}`,
    method: 'put',
    data: toSnake(data)
  })
}

export function deleteAiSkill(id) {
  return request({
    url: `/ai/admin/skills/${id}`,
    method: 'delete'
  })
}

export function toggleAiSkill(id) {
  return request({
    url: `/ai/admin/skills/${id}/toggle`,
    method: 'put'
  })
}

export function initializeCustomerServiceSkills() {
  return request({
    url: '/ai/skills/customer-service/defaults',
    method: 'post'
  })
}

export function testCustomerServiceSkillRoute(question) {
  return request({
    url: '/ai/skills/customer-service/route-test',
    method: 'post',
    data: { question }
  })
}
