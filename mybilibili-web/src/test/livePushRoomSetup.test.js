import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useLivePushRoomSetup } from '../composables/useLivePushRoomSetup.js'
import { LIVE_ROOM_STATUS } from '../utils/liveMeetingStatus.js'

const createSetup = (options = {}) => {
  const api = options.api || {
    getMyRoom: vi.fn(async () => ({ code: 200, data: options.room || null })),
    createRoom: vi.fn(async () => ({ code: 200, data: options.createdRoom || null })),
    updateRoomStatus: vi.fn(async () => ({ code: 200 })),
    uploadCover: vi.fn(async () => ({ code: 200, data: options.coverUrl || '/cover.jpg' }))
  }
  const message = options.message || {
    success: vi.fn(),
    error: vi.fn()
  }
  const copyText = options.copyText || vi.fn(async () => true)
  const getCurrentUserFn = options.getCurrentUserFn || vi.fn(() => ({ id: 7, name: '主播A' }))
  const logger = options.logger || { error: vi.fn() }

  const setup = useLivePushRoomSetup({
    api,
    message,
    getHostname: () => options.hostname || 'live.example.com',
    getCurrentUserFn,
    copyText,
    now: () => options.now ?? 10 * 60 * 1000,
    logger
  })

  return {
    api,
    message,
    copyText,
    getCurrentUserFn,
    logger,
    ...setup
  }
}

describe('useLivePushRoomSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads an existing push room and derives push configuration', async () => {
    const room = {
      id: 1,
      roomName: '直播间',
      streamKey: 'stream-1',
      status: LIVE_ROOM_STATUS.OFFLINE
    }
    const setup = createSetup({ room })
    setup.loading.value = false

    const loadedRoom = await setup.loadRoom()

    expect(setup.api.getMyRoom).toHaveBeenCalled()
    expect(setup.api.createRoom).not.toHaveBeenCalled()
    expect(loadedRoom).toEqual(room)
    expect(setup.room.value).toEqual(room)
    expect(setup.streamKey.value).toBe('stream-1')
    expect(setup.rtmpUrl.value).toBe('rtmp://live.example.com/live')
    expect(setup.me).toEqual({ id: 7, name: '主播A' })
    expect(setup.getCurrentUserFn).toHaveBeenCalledWith({ fallbackName: '主播' })
    expect(setup.loading.value).toBe(false)
  })

  it('creates a room when the account has none', async () => {
    const createdRoom = {
      id: 2,
      roomName: '我的直播间',
      streamKey: 'stream-2',
      status: LIVE_ROOM_STATUS.OFFLINE
    }
    const setup = createSetup({ createdRoom })

    await setup.loadRoom()

    expect(setup.api.createRoom).toHaveBeenCalledWith('我的直播间')
    expect(setup.room.value).toEqual(createdRoom)
    expect(setup.streamKey.value).toBe('stream-2')
  })

  it('marks the room live after WebRTC publishes while tolerating API failures', async () => {
    const setup = createSetup({
      api: {
        getMyRoom: vi.fn(),
        createRoom: vi.fn(),
        updateRoomStatus: vi.fn(async () => {
          throw new Error('network')
        }),
        uploadCover: vi.fn()
      }
    })
    setup.syncRoomState({
      id: 3,
      streamKey: 'stream-3',
      status: LIVE_ROOM_STATUS.OFFLINE
    })

    await setup.handleWebRtcPublished()

    expect(setup.api.updateRoomStatus).toHaveBeenCalledWith(3, LIVE_ROOM_STATUS.LIVE)
    expect(setup.room.value.status).toBe(LIVE_ROOM_STATUS.LIVE)
    expect(setup.isLive.value).toBe(true)
  })

  it('copies push fields, formats viewer join time, validates covers, and uploads a cover', async () => {
    const setup = createSetup({ coverUrl: '/new-cover.jpg' })
    setup.syncRoomState({
      id: 4,
      streamKey: 'stream-4',
      status: LIVE_ROOM_STATUS.OFFLINE
    })

    await setup.copyField('rtmp://live.example.com/live')
    expect(setup.copyText).toHaveBeenCalledWith('rtmp://live.example.com/live', {
      successMessage: '已复制',
      message: setup.message
    })

    expect(setup.formatJoinTime(10 * 60 * 1000 - 5 * 1000)).toBe('5 秒前')
    expect(setup.formatJoinTime(10 * 60 * 1000 - 3 * 60 * 1000)).toBe('3 分钟前')
    expect(setup.formatJoinTime(10 * 60 * 1000 - 2 * 3600 * 1000)).toBe('2 小时前')

    expect(setup.beforeCoverUpload({ type: 'text/plain', size: 100 })).toBe(false)
    expect(setup.message.error).toHaveBeenCalledWith('只能上传图片文件')

    expect(setup.beforeCoverUpload({ type: 'image/png', size: 3 * 1024 * 1024 })).toBe(false)
    expect(setup.message.error).toHaveBeenCalledWith('图片大小不能超过 2MB')

    const file = { type: 'image/jpeg', size: 512 * 1024 }
    expect(setup.beforeCoverUpload(file)).toBe(true)
    expect(await setup.uploadCover({ file })).toBe(true)
    expect(setup.api.uploadCover).toHaveBeenCalledWith(4, file)
    expect(setup.room.value.coverUrl).toBe('/new-cover.jpg')
    expect(setup.message.success).toHaveBeenCalledWith('封面上传成功')
  })
})
