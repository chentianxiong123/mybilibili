import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getLiveRooms, getLiveRoom, updateLiveRoomStatus, getLiveStats, liveApi } from './live'
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

describe('live api - admin 直播管理', () => {
  it('getLiveRooms 默认参数', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getLiveRooms()
    expect(requestMock.get).toHaveBeenCalledWith('/live/admin/rooms', {
      params: { page: 1, size: 10, status: '' },
    })
  })

  it('getLiveRooms 自定义参数', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getLiveRooms(2, 20, 'live')
    expect(requestMock.get).toHaveBeenCalledWith('/live/admin/rooms', {
      params: { page: 2, size: 20, status: 'live' },
    })
  })

  it('getLiveRoom GET /live/admin/rooms/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { id: 5 } })
    await getLiveRoom(5)
    expect(requestMock.get).toHaveBeenCalledWith('/live/admin/rooms/5')
  })

  it('updateLiveRoomStatus PUT 带 status', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await updateLiveRoomStatus(5, 'closed')
    expect(requestMock.put).toHaveBeenCalledWith('/live/admin/rooms/5/status', { status: 'closed' })
  })

  it('getLiveStats GET /live/admin/stats', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: {} })
    await getLiveStats()
    expect(requestMock.get).toHaveBeenCalledWith('/live/admin/stats')
  })
})

describe('live api - 用户直播', () => {
  it('createRoom POST /live/room/create', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: { id: 1 } })
    await liveApi.createRoom('我的直播间')
    expect(requestMock.post).toHaveBeenCalledWith('/live/room/create', { roomName: '我的直播间' })
  })

  it('getMyRoom GET /live/room/my', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await liveApi.getMyRoom()
    expect(requestMock.get).toHaveBeenCalledWith('/live/room/my')
  })

  it('getLiveList GET /live/room/list', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await liveApi.getLiveList()
    expect(requestMock.get).toHaveBeenCalledWith('/live/room/list')
  })

  it('getRoom GET /live/room/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await liveApi.getRoom(3)
    expect(requestMock.get).toHaveBeenCalledWith('/live/room/3')
  })

  it('updateRoomStatus PUT /live/room/:id/status', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await liveApi.updateRoomStatus(3, 'live')
    expect(requestMock.put).toHaveBeenCalledWith('/live/room/3/status', { status: 'live' })
  })

  it('updateRoom PUT /live/room/:id 透传 data', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await liveApi.updateRoom(3, { title: '新' })
    expect(requestMock.put).toHaveBeenCalledWith('/live/room/3', { title: '新' })
  })

  it('uploadCover POST multipart 带 FormData 和 roomId', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: { url: 'x' } })
    const file = new File(['x'], 'a.jpg', { type: 'image/jpeg' })
    await liveApi.uploadCover(3, file)
    expect(requestMock.post).toHaveBeenCalledWith(
      '/live/room/cover',
      expect.any(FormData),
      expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } }),
    )
  })

  it('scheduleRoom PUT /live/room/:id/schedule 透传 scheduledAt', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await liveApi.scheduleRoom(3, '2025-01-01T10:00:00Z')
    expect(requestMock.put).toHaveBeenCalledWith('/live/room/3/schedule', {
      scheduledAt: '2025-01-01T10:00:00Z',
    })
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('boom'))
    await expect(getLiveRooms()).rejects.toThrow('boom')
  })
})
