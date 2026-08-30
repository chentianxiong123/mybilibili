import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ElMessage } from 'element-plus'
import { useMeetingRoomCatalog } from '../composables/useMeetingRoomCatalog.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn()
  }
}))

const createApi = (overrides = {}) => ({
  getMyRooms: vi.fn(async () => ({ code: 200, data: [] })),
  createRoom: vi.fn(async () => ({ code: 200, data: { roomCode: 'ABC123' } })),
  reserveRoom: vi.fn(async () => ({ code: 200, data: { id: 1 } })),
  ...overrides
})

describe('useMeetingRoomCatalog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads current user meeting rooms', async () => {
    const rooms = [{ id: 1, roomCode: 'A1' }]
    const api = createApi({
      getMyRooms: vi.fn(async () => ({ code: 200, data: rooms }))
    })
    const catalog = useMeetingRoomCatalog({ api })

    expect(await catalog.loadMyRooms()).toBe(true)

    expect(api.getMyRooms).toHaveBeenCalled()
    expect(catalog.myRooms.value).toEqual(rooms)
  })

  it('logs room loading failures without mutating local state', async () => {
    const logger = { error: vi.fn() }
    const api = createApi({
      getMyRooms: vi.fn(async () => { throw new Error('network down') })
    })
    const catalog = useMeetingRoomCatalog({ api, logger })

    expect(await catalog.loadMyRooms()).toBe(false)

    expect(logger.error).toHaveBeenCalledWith('获取会议室失败:', expect.any(Error))
    expect(catalog.myRooms.value).toEqual([])
  })

  it('validates create input and joins the newly created room', async () => {
    const api = createApi({
      createRoom: vi.fn(async () => ({ code: 200, data: { roomCode: 'NEW001' } })),
      getMyRooms: vi.fn(async () => ({ code: 200, data: [{ roomCode: 'NEW001' }] }))
    })
    const onRoomCreated = vi.fn()
    const catalog = useMeetingRoomCatalog({ api, onRoomCreated })

    expect(await catalog.createRoom()).toBe(false)
    expect(ElMessage.warning).toHaveBeenCalledWith('请输入会议室名称')
    expect(api.createRoom).not.toHaveBeenCalled()

    catalog.newRoomName.value = '  产品例会  '
    expect(await catalog.createRoom()).toBe(true)

    expect(api.createRoom).toHaveBeenCalledWith('产品例会')
    expect(catalog.newRoomName.value).toBe('')
    expect(catalog.myRooms.value).toEqual([{ roomCode: 'NEW001' }])
    expect(onRoomCreated).toHaveBeenCalledWith({ roomCode: 'NEW001' })
    expect(ElMessage.success).toHaveBeenCalledWith('创建成功')
  })

  it('validates and submits meeting reservations', async () => {
    const api = createApi()
    const catalog = useMeetingRoomCatalog({ api })

    expect(await catalog.reserveRoom()).toBe(false)
    expect(ElMessage.warning).toHaveBeenCalledWith('请输入会议室名称')

    catalog.reserveForm.roomName = '排期会'
    expect(await catalog.reserveRoom()).toBe(false)
    expect(ElMessage.warning).toHaveBeenCalledWith('请选择预约时间段')

    catalog.reserveForm.roomName = ' 排期会 '
    catalog.reserveForm.timeRange = ['2026-05-31T10:00:00', '2026-05-31T11:00:00']
    catalog.reserveForm.reason = '  同步排期  '

    expect(await catalog.reserveRoom()).toBe(true)

    expect(api.reserveRoom).toHaveBeenCalledWith({
      roomName: '排期会',
      scheduledStart: '2026-05-31T10:00:00',
      scheduledEnd: '2026-05-31T11:00:00',
      reason: '同步排期'
    })
    expect(catalog.reserveForm.roomName).toBe('')
    expect(catalog.reserveForm.timeRange).toEqual([])
    expect(catalog.reserveForm.reason).toBe('')
    expect(ElMessage.success).toHaveBeenCalledWith('预约已提交')
  })

  it('keeps form state when reservation fails', async () => {
    const api = createApi({
      reserveRoom: vi.fn(async () => ({ code: 500, message: '时间冲突' }))
    })
    const catalog = useMeetingRoomCatalog({ api })
    catalog.reserveForm.roomName = '冲突会议'
    catalog.reserveForm.timeRange = ['2026-05-31T10:00:00', '2026-05-31T11:00:00']
    catalog.reserveForm.reason = '排查'

    expect(await catalog.reserveRoom()).toBe(false)

    expect(ElMessage.error).toHaveBeenCalledWith('时间冲突')
    expect(catalog.reserveForm.roomName).toBe('冲突会议')
    expect(catalog.reserveForm.timeRange).toEqual(['2026-05-31T10:00:00', '2026-05-31T11:00:00'])
    expect(catalog.reserveForm.reason).toBe('排查')
  })
})
