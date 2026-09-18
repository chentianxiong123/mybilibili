import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import {
  getAiSkills, getAiSkillsByType, getAiSkill, createAiSkill,
  updateAiSkill, deleteAiSkill, toggleAiSkill,
  initializeCustomerServiceSkills, testCustomerServiceSkillRoute
} from './aiSkill'
import request from '@/api/client'

const requestMock = request as unknown as ReturnType<typeof vi.fn> & {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  requestMock.mockReset()
  requestMock.get.mockReset()
  requestMock.post.mockReset()
  requestMock.put.mockReset()
  requestMock.delete.mockReset()
})

describe('aiSkill api', () => {
  it('getAiSkills GET /ai/skills', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAiSkills()
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/skills', method: 'get' })
  })

  it('getAiSkillsByType GET /ai/admin/skills/type/:type', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAiSkillsByType('customer_service')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/skills/type/customer_service',
      method: 'get',
    })
  })

  it('getAiSkill GET /ai/admin/skills/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 3 } })
    await getAiSkill(3)
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/admin/skills/3', method: 'get' })
  })

  it('createAiSkill POST /ai/skills 字段 camelCase 转 snake_case', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 1 } })
    await createAiSkill({ skillName: 'x', isActive: true, maxTokens: 100 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/skills',
      method: 'post',
      data: { skill_name: 'x', is_active: true, max_tokens: 100 },
    })
  })

  it('updateAiSkill PUT /ai/admin/skills/:id 字段转 snake_case', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateAiSkill(4, { skillName: 'y' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/skills/4',
      method: 'put',
      data: { skill_name: 'y' },
    })
  })

  it('deleteAiSkill DELETE /ai/admin/skills/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteAiSkill(8)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/skills/8',
      method: 'delete',
    })
  })

  it('toggleAiSkill PUT /ai/admin/skills/:id/toggle', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await toggleAiSkill(2)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/admin/skills/2/toggle',
      method: 'put',
    })
  })

  it('initializeCustomerServiceSkills POST 默认技能', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await initializeCustomerServiceSkills()
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/skills/customer-service/defaults',
      method: 'post',
    })
  })

  it('testCustomerServiceSkillRoute POST 透传 question', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { route: 'a' } })
    await testCustomerServiceSkillRoute('帮我退款')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/skills/customer-service/route-test',
      method: 'post',
      data: { question: '帮我退款' },
    })
  })
})
