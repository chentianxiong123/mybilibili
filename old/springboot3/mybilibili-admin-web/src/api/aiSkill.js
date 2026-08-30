import request from './request'

export function getAiSkills() {
  return request({
    url: '/ai/admin/skills',
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
    url: '/ai/admin/skills',
    method: 'post',
    data
  })
}

export function updateAiSkill(id, data) {
  return request({
    url: `/ai/admin/skills/${id}`,
    method: 'put',
    data
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