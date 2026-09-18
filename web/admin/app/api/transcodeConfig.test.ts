import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { transcodeConfigApi } from './transcodeConfig'
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

describe('transcodeConfig api', () => {
  it('getConfig GET /admin/transcode-config', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { encoder: 'x264' } })
    const res = await transcodeConfigApi.getConfig()
    expect(requestMock.get).toHaveBeenCalledWith('/admin/transcode-config')
    expect(res.data.encoder).toBe('x264')
  })

  it('updateConfig PUT 透传 data', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await transcodeConfigApi.updateConfig({ encoder: 'x265' })
    expect(requestMock.put).toHaveBeenCalledWith('/admin/transcode-config', { encoder: 'x265' })
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(transcodeConfigApi.getConfig()).rejects.toThrow('500')
  })
})
