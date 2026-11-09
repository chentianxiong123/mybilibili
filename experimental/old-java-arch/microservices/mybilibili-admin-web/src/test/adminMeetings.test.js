import { beforeEach, describe, expect, it, vi } from 'vitest'
import { MEETING_STATUS } from '../utils/liveMeetingStatus.js'
import { useAdminMeetings } from '../composables/useAdminMeetings.js'

const createHarness = (overrides = {}) => {
  const fetchRooms = vi.fn(async () => ({
    code: 200,
    data: {
      records: [{ id: 1, roomName: '排期会', status: MEETING_STATUS.NOT_STARTED }],
      total: 1
    }
  }))
  const fetchPending = vi.fn(async () => ({
    code: 200,
    data: [{ id: 2, roomName: '待审批会议', status: MEETING_STATUS.PENDING_APPROVAL }]
  }))
  const approveReservation = vi.fn(async () => ({ code: 200 }))
  const rejectReservation = vi.fn(async () => ({ code: 200 }))
  const endMeeting = vi.fn(async () => ({ code: 200 }))
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
    ...useAdminMeetings({
      fetchRooms,
      fetchPending,
      approveReservation,
      rejectReservation,
      endMeeting,
      normalizePage,
      runAction,
      logger,
      ...overrides
    }),
    fetchRooms,
    fetchPending,
    approveReservation,
    rejectReservation,
    endMeeting,
    normalizePage,
    runAction,
    logger
  }
}

describe('useAdminMeetings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads meeting rooms and pending reservations', async () => {
    const harness = createHarness()

    await harness.initialize()

    expect(harness.fetchRooms).toHaveBeenCalledWith(1, 10, null)
    expect(harness.fetchPending).toHaveBeenCalled()
    expect(harness.rooms.value).toEqual([{ id: 1, roomName: '排期会', status: MEETING_STATUS.NOT_STARTED }])
    expect(harness.pendingReservations.value).toEqual([
      { id: 2, roomName: '待审批会议', status: MEETING_STATUS.PENDING_APPROVAL }
    ])
    expect(harness.total.value).toBe(1)
  })

  it('approves reservations and refreshes the admin lists', async () => {
    const harness = createHarness()

    await expect(harness.approve({ id: 9, roomName: '排期会' })).resolves.toBe(true)

    expect(harness.runAction).toHaveBeenCalledWith(expect.objectContaining({
      message: '确认通过「排期会」的预约？',
      title: '审批',
      successMessage: '已通过'
    }))
    expect(harness.approveReservation).toHaveBeenCalledWith(9)
    expect(harness.fetchPending).toHaveBeenCalledTimes(1)
    expect(harness.fetchRooms).toHaveBeenCalledTimes(1)
  })

  it('rejects reservations and refreshes the admin lists', async () => {
    const harness = createHarness()

    await expect(harness.reject({ id: 10, roomName: '排期会' })).resolves.toBe(true)

    expect(harness.runAction).toHaveBeenCalledWith(expect.objectContaining({
      message: '确认拒绝「排期会」的预约？',
      title: '审批',
      successMessage: '已拒绝'
    }))
    expect(harness.rejectReservation).toHaveBeenCalledWith(10)
    expect(harness.fetchPending).toHaveBeenCalledTimes(1)
    expect(harness.fetchRooms).toHaveBeenCalledTimes(1)
  })

  it('forces meeting end and opens detail dialogs', async () => {
    const harness = createHarness()
    const row = { id: 12, roomName: '回顾会' }

    await expect(harness.forceEnd(row)).resolves.toBe(true)

    expect(harness.endMeeting).toHaveBeenCalledWith(12)
    expect(harness.runAction).toHaveBeenCalledWith(expect.objectContaining({
      title: '警告',
      successMessage: '已强制结束'
    }))
    expect(harness.fetchRooms).toHaveBeenCalledTimes(1)
    expect(harness.fetchPending).not.toHaveBeenCalled()

    harness.viewDetail(row)
    expect(harness.currentRoom.value).toEqual(row)
    expect(harness.detailDialog.value).toBe(true)
  })
})
