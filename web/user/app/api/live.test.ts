import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
  apiPut: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: mocks.apiPut,
    delete: vi.fn(),
  }),
}))

import {
  getLiveRooms,
  getLiveRoom,
  updateLiveRoomStatus,
  getLiveStats,
  liveApi,
} from './live'

describe('live admin functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getLiveRooms 默认参数', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await getLiveRooms()
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/admin/rooms', { params: { page: 1, size: 10, status: '' } })
  })

  it('getLiveRooms 自定义参数', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await getLiveRooms(2, 5, 'LIVE')
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/admin/rooms', { params: { page: 2, size: 5, status: 'LIVE' } })
  })

  it('getLiveRoom', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { id: 1 } })
    await getLiveRoom(1)
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/admin/rooms/1')
  })

  it('updateLiveRoomStatus PUT body', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await updateLiveRoomStatus(1, 'CLOSED')
    expect(mocks.apiPut).toHaveBeenCalledWith('/live/admin/rooms/1/status', { status: 'CLOSED' })
  })

  it('getLiveStats', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await getLiveStats()
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/admin/stats')
  })
})

describe('liveApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('createRoom POST body', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200, data: { id: 1 } })
    await liveApi.createRoom('my room')
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/room/create', { roomName: 'my room' })
  })

  it('getMyRoom', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { id: 1 } })
    await liveApi.getMyRoom()
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/room/my')
  })

  it('getLiveList', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await liveApi.getLiveList()
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/room/list')
  })

  it('getRoom', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await liveApi.getRoom(2)
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/room/2')
  })

  it('updateRoomStatus', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await liveApi.updateRoomStatus(3, 'PAUSED')
    expect(mocks.apiPut).toHaveBeenCalledWith('/live/room/3/status', { status: 'PAUSED' })
  })

  it('updateRoom', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await liveApi.updateRoom(3, { title: 'x' })
    expect(mocks.apiPut).toHaveBeenCalledWith('/live/room/3', { title: 'x' })
  })

  it('uploadCover 用 multipart', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 'c.jpg')
    await liveApi.uploadCover(4, file)
    expect(mocks.apiPost).toHaveBeenCalled()
    const [url, form, cfg] = mocks.apiPost.mock.calls[0]
    expect(url).toBe('/live/room/cover')
    expect(form.get('file')).toBe(file)
    expect(form.get('roomId')).toBe('4')
    expect(cfg).toEqual({ headers: { 'Content-Type': 'multipart/form-data' } })
  })

  it('scheduleRoom PUT', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await liveApi.scheduleRoom(5, '2024-01-01T00:00:00Z')
    expect(mocks.apiPut).toHaveBeenCalledWith('/live/room/5/schedule', { scheduledAt: '2024-01-01T00:00:00Z' })
  })
})
