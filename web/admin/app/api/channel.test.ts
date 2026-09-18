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
  getChannels, getChannel, createChannel, updateChannel, deleteChannel,
  toggleChannel, getChannelsByType, getBindings, bindFeature,
  testConnection, getAvailableTypes, getAvailableFeatures
} from './channel'
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

describe('channel api - 渠道管理', () => {
  it('getChannels GET /ai/configs', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getChannels()
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs', method: 'get' })
  })

  it('getChannel GET /ai/configs/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 1 } })
    await getChannel(1)
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs/1', method: 'get' })
  })

  it('createChannel POST 字段转 snake_case', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await createChannel({ channelName: 'openai', apiKey: 'sk' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/configs',
      method: 'post',
      data: { channel_name: 'openai', api_key: 'sk' },
    })
  })

  it('updateChannel PUT 字段转 snake_case', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateChannel(2, { channelName: 'y' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/configs/2',
      method: 'put',
      data: { channel_name: 'y' },
    })
  })

  it('deleteChannel DELETE /ai/configs/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteChannel(3)
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs/3', method: 'delete' })
  })

  it('toggleChannel PUT /ai/configs/:id/toggle', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await toggleChannel(4)
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs/4/toggle', method: 'put' })
  })
})

describe('channel api - 按类型查询', () => {
  it('getChannelsByType query string 拼接 type', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getChannelsByType('chat')
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs?type=chat', method: 'get' })
  })
})

describe('channel api - 功能绑定', () => {
  it('getBindings GET /ai/bindings', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getBindings()
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/bindings', method: 'get' })
  })

  it('bindFeature POST /ai/bindings/:feature 带 configId', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await bindFeature('chat', 1)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/bindings/chat',
      method: 'post',
      data: { configId: 1 },
    })
  })
})

describe('channel api - 测试连接', () => {
  it('testConnection POST 字段转 snake_case', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { ok: true } })
    await testConnection({ apiKey: 'k', baseUrl: 'https://a' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/ai/config/test',
      method: 'post',
      data: { api_key: 'k', base_url: 'https://a' },
    })
  })
})

describe('channel api - 可用类型和功能', () => {
  it('getAvailableTypes GET /ai/configs/types', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAvailableTypes()
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs/types', method: 'get' })
  })

  it('getAvailableFeatures GET /ai/configs/features', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getAvailableFeatures()
    expect(requestMock).toHaveBeenCalledWith({ url: '/ai/configs/features', method: 'get' })
  })
})
