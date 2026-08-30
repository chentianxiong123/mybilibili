import { beforeEach, describe, expect, it, vi } from 'vitest'
import { LIVE_ROOM_STATUS } from '../utils/liveMeetingStatus.js'
import { useAdminLiveRooms } from '../composables/useAdminLiveRooms.js'

const createHarness = (overrides = {}) => {
  const fetchRooms = vi.fn(async () => ({
    code: 200,
    data: {
      records: [{ id: 1, roomName: '直播间A', status: LIVE_ROOM_STATUS.OFFLINE }],
      total: 1
    }
  }))
  const fetchStats = vi.fn(async () => ({
    code: 200,
    data: {
      totalRooms: 2,
      onlineRooms: 1,
      totalViewers: 9
    }
  }))
  const updateStatus = vi.fn(async () => ({ code: 200 }))
  const normalizePage = vi.fn((data) => ({
    records: data.records ?? [],
    total: data.total ?? 0
  }))
  const runAction = vi.fn(async ({ action, onSuccess }) => {
    await action()
    if (onSuccess) await onSuccess()
    return true
  })
  const logger = { error: vi.fn() }

  return {
    ...useAdminLiveRooms({
      fetchRooms,
      fetchStats,
      updateStatus,
      normalizePage,
      runAction,
      logger,
      ...overrides
    }),
    fetchRooms,
    fetchStats,
    updateStatus,
    normalizePage,
    runAction,
    logger
  }
}

describe('useAdminLiveRooms', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads rooms and stats', async () => {
    const harness = createHarness()

    await harness.initialize()

    expect(harness.fetchRooms).toHaveBeenCalledWith(1, 10, '')
    expect(harness.fetchStats).toHaveBeenCalled()
    expect(harness.rooms.value).toEqual([{ id: 1, roomName: '直播间A', status: LIVE_ROOM_STATUS.OFFLINE }])
    expect(harness.total.value).toBe(1)
    expect(harness.stats.value).toEqual({
      totalRooms: 2,
      onlineRooms: 1,
      totalViewers: 9
    })
  })

  it('changes pages, filters rooms, and toggles status', async () => {
    const harness = createHarness()

    await harness.handlePageChange(3)
    expect(harness.page.value).toBe(3)
    expect(harness.fetchRooms).toHaveBeenLastCalledWith(3, 10, '')

    harness.statusFilter.value = LIVE_ROOM_STATUS.LIVE
    await harness.handleStatusFilterChange()
    expect(harness.page.value).toBe(1)
    expect(harness.fetchRooms).toHaveBeenLastCalledWith(1, 10, LIVE_ROOM_STATUS.LIVE)

    const row = { id: 7, roomName: '主直播间', status: LIVE_ROOM_STATUS.LIVE }
    await expect(harness.toggleRoomStatus(row)).resolves.toBe(true)

    expect(harness.runAction).toHaveBeenCalled()
    expect(harness.updateStatus).toHaveBeenCalledWith(7, LIVE_ROOM_STATUS.OFFLINE)
    expect(harness.fetchRooms).toHaveBeenCalledTimes(3)
    expect(harness.fetchStats).toHaveBeenCalledTimes(1)
  })
})
