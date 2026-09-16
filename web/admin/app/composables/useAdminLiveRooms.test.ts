import { describe, it, expect, vi, beforeEach } from 'vitest'

import { useAdminLiveRooms } from './useAdminLiveRooms'
import { LIVE_ROOM_STATUS } from '@/utils/liveMeetingStatus'

function makeFetchRooms(impl: (page: number, size: number, status: string) => Promise<any>) {
  return vi.fn(impl)
}

function makeFetchStats(impl: () => Promise<any>) {
  return vi.fn(impl)
}

function makeUpdateStatus(impl: (id: any, status: string) => Promise<any>) {
  return vi.fn(impl)
}

const noopRunAction = vi.fn(async () => true)

describe('useAdminLiveRooms - initial state', () => {
  it('exposes default reactive state', () => {
    const r = useAdminLiveRooms()
    expect(r.loading.value).toBe(false)
    expect(r.rooms.value).toEqual([])
    expect(r.total.value).toBe(0)
    expect(r.page.value).toBe(1)
    expect(r.pageSize.value).toBe(10)
    expect(r.statusFilter.value).toBe('')
    expect(r.stats.value).toEqual({
      totalRooms: 0,
      onlineRooms: 0,
      totalViewers: 0
    })
  })

  it('exposes status options for the dropdown', () => {
    const r = useAdminLiveRooms()
    expect(r.statusOptions).toEqual([
      { label: '全部', value: '' },
      { label: '直播中', value: LIVE_ROOM_STATUS.LIVE },
      { label: '离线', value: LIVE_ROOM_STATUS.OFFLINE }
    ])
  })
})

describe('useAdminLiveRooms.loadRooms', () => {
  let fetchRooms: any
  let normalizePage: any
  let logger: any

  beforeEach(() => {
    fetchRooms = makeFetchRooms(async () => ({
      code: 200,
      data: { records: [{ id: 1, title: 'r1' }], total: 1 }
    }))
    normalizePage = vi.fn((d: any) => ({ records: d.records, total: d.total }))
    logger = { error: vi.fn(), warn: vi.fn(), log: vi.fn(), info: vi.fn(), debug: vi.fn() }
  })

  it('loads rooms successfully and normalizes the response', async () => {
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    const ok = await r.loadRooms()
    expect(ok).toBe(true)
    expect(fetchRooms).toHaveBeenCalledWith(1, 10, '')
    expect(normalizePage).toHaveBeenCalledWith({ records: [{ id: 1, title: 'r1' }], total: 1 })
    expect(r.total.value).toBe(1)
    expect(r.rooms.value).toEqual([
      {
        id: 1,
        roomName: 'r1',
        userId: null,
        status: undefined,
        viewerCount: 0,
        category: '',
        createdAt: ''
      }
    ])
    expect(r.loading.value).toBe(false)
  })

  it('normalizes snake_case room data via normalizeRoom mapping', async () => {
    fetchRooms.mockResolvedValueOnce({
      code: 200,
      data: {
        records: [{
          id: 9,
          title: 'snake',
          user_id: 7,
          status: 'live',
          viewer_count: 42,
          category: 'tech',
          created_at: '2024-01-01T00:00:00'
        }],
        total: 100
      }
    })
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    await r.loadRooms()
    expect(r.rooms.value).toEqual([{
      id: 9,
      roomName: 'snake',
      userId: 7,
      status: 'live',
      viewerCount: 42,
      category: 'tech',
      createdAt: '2024-01-01T00:00:00'
    }])
    expect(r.total.value).toBe(100)
  })

  it('falls back to roomName when title is missing', async () => {
    fetchRooms.mockResolvedValueOnce({
      code: 200,
      data: { records: [{ id: 1, roomName: 'fallback' }], total: 1 }
    })
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    await r.loadRooms()
    expect(r.rooms.value[0].roomName).toBe('fallback')
  })

  it('returns false when API responds with non-200 code', async () => {
    fetchRooms.mockResolvedValueOnce({ code: 500, msg: 'server error', data: null })
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    const ok = await r.loadRooms()
    expect(ok).toBe(false)
    expect(r.rooms.value).toEqual([])
    expect(r.total.value).toBe(0)
  })

  it('catches exceptions, logs and returns false', async () => {
    const err = new Error('network')
    fetchRooms.mockRejectedValueOnce(err)
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    const ok = await r.loadRooms()
    expect(ok).toBe(false)
    expect(logger.error).toHaveBeenCalledWith(err)
    expect(r.loading.value).toBe(false)
  })

  it('toggles loading flag during execution', async () => {
    let resolve: (v: any) => void
    fetchRooms.mockImplementationOnce(() => new Promise((res) => { resolve = res }))
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    const p = r.loadRooms()
    expect(r.loading.value).toBe(true)
    resolve!({ code: 200, data: { records: [], total: 0 } })
    await p
    expect(r.loading.value).toBe(false)
  })

  it('passes current page/size/statusFilter to fetchRooms', async () => {
    const r = useAdminLiveRooms({ fetchRooms, normalizePage, logger, runAction: noopRunAction })
    r.page.value = 3
    r.pageSize.value = 25
    r.statusFilter.value = 'live'
    await r.loadRooms()
    expect(fetchRooms).toHaveBeenCalledWith(3, 25, 'live')
  })
})

describe('useAdminLiveRooms.loadStats', () => {
  it('normalizes snake_case stats and merges with defaults', async () => {
    const fetchStats = makeFetchStats(async () => ({
      code: 200,
      data: { total_rooms: 10, live_count: 4, total_viewers: 999 }
    }))
    const logger = { error: vi.fn() }
    const r = useAdminLiveRooms({ fetchStats, runAction: noopRunAction, logger })
    const ok = await r.loadStats()
    expect(ok).toBe(true)
    expect(r.stats.value).toEqual({
      totalRooms: 10,
      onlineRooms: 4,
      totalViewers: 999
    })
  })

  it('normalizes camelCase stats fields', async () => {
    const fetchStats = makeFetchStats(async () => ({
      code: 200,
      data: { totalRooms: 5, onlineRooms: 1, totalViewers: 33 }
    }))
    const r = useAdminLiveRooms({ fetchStats, runAction: noopRunAction, logger: { error: vi.fn() } })
    await r.loadStats()
    expect(r.stats.value).toEqual({
      totalRooms: 5,
      onlineRooms: 1,
      totalViewers: 33
    })
  })

  it('falls back to defaults when data is null/empty', async () => {
    const fetchStats = makeFetchStats(async () => ({ code: 200, data: null }))
    const r = useAdminLiveRooms({ fetchStats, runAction: noopRunAction, logger: { error: vi.fn() } })
    await r.loadStats()
    expect(r.stats.value).toEqual({
      totalRooms: 0,
      onlineRooms: 0,
      totalViewers: 0
    })
  })

  it('returns false when API code is not 200', async () => {
    const fetchStats = makeFetchStats(async () => ({ code: 401, data: {} }))
    const r = useAdminLiveRooms({ fetchStats, runAction: noopRunAction, logger: { error: vi.fn() } })
    expect(await r.loadStats()).toBe(false)
  })

  it('catches exceptions and returns false', async () => {
    const err = new Error('boom')
    const fetchStats = makeFetchStats(async () => { throw err })
    const logger = { error: vi.fn() }
    const r = useAdminLiveRooms({ fetchStats, runAction: noopRunAction, logger })
    expect(await r.loadStats()).toBe(false)
    expect(logger.error).toHaveBeenCalledWith(err)
  })
})

describe('useAdminLiveRooms pagination & filter handlers', () => {
  let fetchRooms: any

  beforeEach(() => {
    fetchRooms = makeFetchRooms(async () => ({
      code: 200,
      data: { records: [], total: 0 }
    }))
  })

  it('handlePageChange updates page and reloads', async () => {
    const r = useAdminLiveRooms({ fetchRooms, runAction: noopRunAction, logger: { error: vi.fn() } })
    r.pageSize.value = 10
    r.page.value = 1
    await r.handlePageChange(5)
    expect(r.page.value).toBe(5)
    expect(fetchRooms).toHaveBeenCalledWith(5, 10, '')
  })

  it('handleStatusFilterChange resets page to 1 and reloads', async () => {
    const r = useAdminLiveRooms({ fetchRooms, runAction: noopRunAction, logger: { error: vi.fn() } })
    r.page.value = 4
    r.statusFilter.value = 'offline'
    await r.handleStatusFilterChange()
    expect(r.page.value).toBe(1)
    expect(fetchRooms).toHaveBeenLastCalledWith(1, 10, 'offline')
  })
})

describe('useAdminLiveRooms.toggleRoomStatus', () => {
  it('passes the inverted next status to runAction for a live room', async () => {
    const updateStatus = makeUpdateStatus(async () => ({ code: 200 }))
    const runAction = vi.fn(async () => true)
    const r = useAdminLiveRooms({
      updateStatus,
      runAction,
      runConfirmedAction: runAction,
      logger: { error: vi.fn() }
    })
    const row = { id: 1, roomName: '频道A', status: LIVE_ROOM_STATUS.LIVE }
    await r.toggleRoomStatus(row)
    expect(runAction).toHaveBeenCalledWith(expect.objectContaining({
      message: expect.stringContaining('下播'),
      successMessage: '下播成功',
      action: expect.any(Function)
    }))
    const cfg = runAction.mock.calls[0][0]
    await cfg.action()
    expect(updateStatus).toHaveBeenCalledWith(1, LIVE_ROOM_STATUS.OFFLINE)
  })

  it('uses 开播 when room is currently offline', async () => {
    const updateStatus = makeUpdateStatus(async () => ({ code: 200 }))
    const runAction = vi.fn(async () => true)
    const r = useAdminLiveRooms({
      updateStatus,
      runAction,
      logger: { error: vi.fn() }
    })
    const row = { id: 2, roomName: '频道B', status: LIVE_ROOM_STATUS.OFFLINE }
    await r.toggleRoomStatus(row)
    const cfg = runAction.mock.calls[0][0]
    expect(cfg.message).toContain('开播')
    expect(cfg.successMessage).toBe('开播成功')
    await cfg.action()
    expect(updateStatus).toHaveBeenCalledWith(2, LIVE_ROOM_STATUS.LIVE)
  })

  it('onSuccess reloads both rooms and stats', async () => {
    const updateStatus = makeUpdateStatus(async () => ({ code: 200 }))
    const fetchRooms = makeFetchRooms(async () => ({ code: 200, data: { records: [], total: 0 } }))
    const fetchStats = makeFetchStats(async () => ({ code: 200, data: { total_rooms: 1, live_count: 0, total_viewers: 0 } }))
    let capturedOnSuccess: any
    const runAction = vi.fn(async ({ onSuccess }: any) => { capturedOnSuccess = onSuccess; return true })
    const r = useAdminLiveRooms({
      updateStatus, fetchRooms, fetchStats,
      runAction,
      logger: { error: vi.fn() }
    })
    const row = { id: 1, roomName: 'r', status: LIVE_ROOM_STATUS.LIVE }
    await r.toggleRoomStatus(row)
    expect(capturedOnSuccess).toBeDefined()
    await capturedOnSuccess()
    expect(fetchRooms).toHaveBeenCalled()
    expect(fetchStats).toHaveBeenCalled()
  })
})

describe('useAdminLiveRooms.initialize', () => {
  it('loads both rooms and stats', async () => {
    const fetchRooms = makeFetchRooms(async () => ({ code: 200, data: { records: [{ id: 1, title: 'r' }], total: 1 } }))
    const fetchStats = makeFetchStats(async () => ({ code: 200, data: { total_rooms: 5 } }))
    const r = useAdminLiveRooms({ fetchRooms, fetchStats, runAction: noopRunAction, logger: { error: vi.fn() } })
    await r.initialize()
    expect(fetchRooms).toHaveBeenCalledTimes(1)
    expect(fetchStats).toHaveBeenCalledTimes(1)
    expect(r.total.value).toBe(1)
    expect(r.stats.value.totalRooms).toBe(5)
  })
})
